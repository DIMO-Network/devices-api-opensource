package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// orchestration objective when doing a reload data.
// 1. get all engines and check if matching meta engine name exists, if not hold on for a minute.
// 2. Create a regular source engine (not meta type, no sources defined), with timestamp appended to name above
// 3. Add All the documents from DB
// 4. if an existing meta engine exists, add #2 to this one as source, otherwise create new meta engine with #2 as source
// 5. if an existing meta engine exists, remove the sources that are not #2
// 6. Delete the previous engine that we removed from meta engine in #5
func loadElasticDevices(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	esSvc, err := services.NewElasticSearchService(settings, *logger)
	if err != nil {
		return err
	}
	existingEngines, err := esSvc.GetEngines()
	if err != nil {
		return err
	}
	logger.Info().Msgf("found existing engines: %d", len(existingEngines.Results))

	// get all devices from DB.
	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles)).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	logger.Info().Msgf("found %d device definitions verified", len(all))
	if len(all) == 0 {
		return errors.New("0 items found to index, stopping")
	}

	docs := make([]services.DeviceDefinitionSearchDoc, len(all))
	for i, definition := range all {
		sd := fmt.Sprintf("%d %s %s", definition.Year, definition.Make, definition.Model)
		sm := services.SubModelsFromStylesDB(definition.R.DeviceStyles)
		for i2, s := range sm {
			sm[i2] = sd + " " + s
		}
		docs[i] = services.DeviceDefinitionSearchDoc{
			ID:            definition.ID,
			SearchDisplay: sd,
			Make:          definition.Make,
			Model:         definition.Model,
			Year:          int(definition.Year),
			SubModels:     sm,
			ImageURL:      definition.ImageURL.String,
		}
	}

	tempEngineName := fmt.Sprintf("%s-%s", esSvc.MetaEngineName, time.Now().Format("2006-01-02t15-04"))
	tempEngine, err := esSvc.CreateEngine(tempEngineName, nil)
	if err != nil {
		return err
	}
	logger.Info().Msgf("created engine %s", tempEngine.Name)
	err = esSvc.CreateDocumentsBatched(docs, tempEngine.Name)
	if err != nil {
		return err
	}
	logger.Info().Msgf("created documents in engine %s", tempEngine.Name)

	var metaEngine *services.EngineDetail
	var previousTempEngines []string
	// look for existing meta engine, and any previous core engines that should be removed.
	for _, result := range existingEngines.Results {
		if result.Name == esSvc.MetaEngineName && *result.Type == "meta" {
			metaEngine = &result
			logger.Info().Msgf("found existing meta engine: %+v", *metaEngine)
		}
		if strings.Contains(result.Name, esSvc.MetaEngineName+"-") && *result.Type == "default" {
			previousTempEngines = append(previousTempEngines, result.Name)
			logger.Info().Msgf("found previous device defs engine: %s. It will be removed", result.Name)
		}
	}
	if metaEngine == nil {
		_, err = esSvc.CreateEngine(esSvc.MetaEngineName, &tempEngineName)
		if err != nil {
			return err
		}
		logger.Info().Msg("created meta engine with temp engine assigned.")
	} else {
		_, err = esSvc.AddSourceEngineToMetaEngine(tempEngineName, esSvc.MetaEngineName)
		if err != nil {
			return err
		}
		logger.Info().Msgf("added source %s to meta engine %s", tempEngine.Name, esSvc.MetaEngineName)
		for _, prev := range previousTempEngines {
			// loop over all previous ones
			if services.Contains(metaEngine.SourceEngines, prev) {
				_, err = esSvc.RemoveSourceEngine(prev, esSvc.MetaEngineName)
				if err != nil {
					return err
				}
				logger.Info().Msgf("removed previous source engine %s from %s", prev, esSvc.MetaEngineName)
			}

			err = esSvc.DeleteEngine(prev)
			if err != nil {
				return err
			}
			logger.Info().Msgf("delete engine: %s", prev)
		}
	}
	err = esSvc.UpdateSearchSettingsForDeviceDefs(tempEngineName)
	if err != nil {
		return err
	}
	err = esSvc.UpdateSearchSettingsForDeviceDefs(esSvc.MetaEngineName)
	if err != nil {
		return err
	}
	logger.Info().Msg("completed load ok")

	// orchestrate the process using the device search service. try to refactor some methods out to it that could be standardish.
	return nil
}
