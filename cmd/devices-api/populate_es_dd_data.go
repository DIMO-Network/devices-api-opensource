package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	es "github.com/DIMO-Network/devices-api/internal/elasticsearch"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func populateESDDData(ctx context.Context, settings *config.Settings, e es.ElasticSearch, pdb database.DbStore, logger *zerolog.Logger) error {
	db := pdb.DBS().Reader

	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.ExternalID.IsNotNull(),
		qm.Where("metadata ->> 'elasticDefinitionSynced' IS NULL OR metadata ->> 'elasticDefinitionSynced' = ?", false),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).All(ctx, db)

	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}

	for _, apiInt := range apiInts {
		makeRel := apiInt.R.UserDevice.R.DeviceDefinition.R.DeviceMake
		ddRel := apiInt.R.UserDevice.R.DeviceDefinition
		ddID := ddRel.ID

		md := services.UserDeviceMetadata{}
		if err = apiInt.R.UserDevice.Metadata.Unmarshal(&md); err != nil {
			logger.Error().Msgf("Could not unmarshal userdevice metadata for device: %s", apiInt.R.UserDevice.ID)
			continue
		}

		if !md.ElasticDefinitionSynced {
			dd := services.DeviceDefinitionDTO{
				DeviceDefinitionID: ddID,
				UserDeviceID:       apiInt.R.UserDevice.ID,
				Make:               makeRel.Name,
				Model:              ddRel.Model,
				Year:               int(ddRel.Year),
			}
			err = e.UpdateAutopiDevicesByQuery(dd, settings.ElasticDeviceStatusIndex)
			if err != nil {
				logger.Error().Msgf("error occurred during es update: %s", err)
				continue
			}

			md.ElasticDefinitionSynced = true
			err = apiInt.R.UserDevice.Metadata.Marshal(&md)
			if err != nil {
				logger.Error().Msgf("could not marshal userdevice metadata for device: %s", apiInt.R.UserDevice.ID)
				continue
			}

			if _, err := apiInt.R.UserDevice.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				logger.Err(err).Str("userDeviceId", apiInt.UserDeviceID).Msg("Could not update metadata for device.")
				continue
			}
		} else {
			logger.Debug().Msgf("device record has already been updated for user device %s", apiInt.R.UserDevice.ID)
		}
	}

	return nil
}
