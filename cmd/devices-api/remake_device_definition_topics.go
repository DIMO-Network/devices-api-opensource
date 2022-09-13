package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// remakeDeviceDefinitionTopics invokes [services.DeviceDefinitionRegistrar] for each user device
// with an integration.
func remakeDeviceDefinitionTopics(ctx context.Context, settings *config.Settings, pdb database.DbStore, producer sarama.SyncProducer, logger *zerolog.Logger) error {
	reg := services.NewDeviceDefinitionRegistrar(producer, settings)
	db := pdb.DBS().Reader

	// Find all integrations instances.
	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.ExternalID.IsNotNull(),
		qm.Load(
			qm.Rels(
				models.UserDeviceAPIIntegrationRels.UserDevice,
				models.UserDeviceRels.DeviceDefinition,
				models.DeviceDefinitionRels.DeviceMake,
			),
		),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve integration instances: %w", err)
	}

	failures := 0

	// For each of these, register the device's device definition with the data pipeline.
	for _, apiInt := range apiInts {
		userDeviceID := apiInt.UserDeviceID
		dd := apiInt.R.UserDevice.R.DeviceDefinition
		ddMake := apiInt.R.UserDevice.R.DeviceDefinition.R.DeviceMake.Name

		ddReg := services.DeviceDefinitionDTO{
			UserDeviceID:       userDeviceID,
			DeviceDefinitionID: dd.ID,
			IntegrationID:      apiInt.IntegrationID,
			Make:               ddMake,
			Model:              dd.Model,
			Year:               int(dd.Year),
		}

		err := reg.Register(ddReg)
		if err != nil {
			logger.Err(err).Str("userDeviceId", userDeviceID).Msg("Failed to register device's device definition.")
			failures++
		}
	}

	log.Info().Int("attempted", len(apiInts)).Int("failed", failures).Msg("Finished device definition registration.")

	return nil
}
