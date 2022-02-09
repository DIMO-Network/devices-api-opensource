package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func loadSmartCarData(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	apiSvc := services.NewSmartCarService("https://api.smartcar.com/v2.0/", pdb.DBS, logger)

	err := apiSvc.SeedDeviceDefinitionsFromSmartCar(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("error seeding device defs from smart car")
	}
}

func smartCarForwardCompatibility(ctx context.Context, logger zerolog.Logger, pdb database.DbStore) error {
	// get all device integrations where smartcar ordered by date asc.
	// then group them by make and model, so we can iterate over the years.
	// if there is a gap in the years, insert device_integration
	scSvc := services.NewSmartCarService("https://api.smartcar.com/v2.0/", pdb.DBS, logger)
	integrationID, err := scSvc.GetOrCreateSmartCarIntegration(ctx)
	deviceDefSvc := services.NewDeviceDefinitionService(&config.Settings{}, pdb.DBS, &logger, nil)

	if err != nil {
		return err
	}

	deviceDefs, err := models.DeviceDefinitions(qm.InnerJoin("device_integrations di on di.device_definition_id = device_definitions.id"),
		qm.Where("di.integration_id = ?", integrationID), qm.OrderBy("make, model, year DESC")).
		All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	fmt.Printf("found %d device definitions with smartcar integration\n", len(deviceDefs))

	lastMM := ""
	lastYear := int16(0)
	// meant to be used at the end of each loop to update the "last" values
	funcLastValues := func(dd *models.DeviceDefinition) {
		lastYear = dd.Year
		lastMM = dd.Make + dd.Model
	}
	// year will be descending
	for _, dd := range deviceDefs {
		thisMM := dd.Make + dd.Model
		if lastMM == thisMM {
			// we care about year gaps
			yearDiff := lastYear - dd.Year
			if yearDiff > 1 {
				// we have a gap
				fmt.Printf("found a year gap of %d...", yearDiff)
				// todo: need to loop for each yearDiff -1, eg. if found 3 years of gap, iterate over dd.Year +1 and +2
				gapDd, err := deviceDefSvc.FindDeviceDefinitionByMMY(ctx, pdb.DBS().Reader, dd.Make, dd.Model, int(dd.Year+1), true)
				if errors.Is(err, sql.ErrNoRows) {
					funcLastValues(dd)
					continue
				}
				if err != nil {
					return err
				}
				// found a record that needs to be attached to integration
				if len(gapDd.R.DeviceIntegrations) == 0 {
					fmt.Printf("\nfound device def for year gap %s, inserting device_integration\n", printMMY(gapDd, Green, true))
					diGap := models.DeviceIntegration{
						DeviceDefinitionID: gapDd.ID,
						IntegrationID:      integrationID,
						Country:            "USA",       // default
						Capabilities:       null.JSON{}, // we'd need to copy from previous dd?
					}
					err = diGap.Insert(ctx, pdb.DBS().Writer, boil.Infer())
					if err != nil {
						return errors.Wrap(err, "error inserting device_integration")
					}
				} else {
					fmt.Printf("but %s already had an integration set\n", printMMY(gapDd, Red, true))
				}
			}
		} else {
			// this should mean we are back at the start of a new make/model starting at highest year
			nextYearDd, err := deviceDefSvc.FindDeviceDefinitionByMMY(ctx, pdb.DBS().Writer, dd.Make, dd.Model, int(dd.Year+1), true)
			if errors.Is(err, sql.ErrNoRows) {
				funcLastValues(dd)
				continue
			}
			if err != nil {
				return err
			}
			// does it have any integrations?
			if len(nextYearDd.R.DeviceIntegrations) == 0 {
				// attach smartcar integration
				fmt.Printf("found device def for future year %s, that does not have any integrations. inserting device_integration\n", printMMY(nextYearDd, Green, true))
				diGap := models.DeviceIntegration{
					DeviceDefinitionID: nextYearDd.ID,
					IntegrationID:      integrationID,
					Country:            "USA",       // default
					Capabilities:       null.JSON{}, // we'd need to copy from previous dd?
				}
				err = diGap.Insert(ctx, pdb.DBS().Writer, boil.Infer())
				if err != nil {
					return errors.Wrap(err, "error inserting device_integration")
				}
			}
		}

		funcLastValues(dd)
	}

	return nil
}
