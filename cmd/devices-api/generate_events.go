package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func generateEvents(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, eventService services.EventService) {
	ctx := context.Background()
	tx, err := pdb.DBS().Reader.BeginTx(ctx, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create transaction")
	}
	defer tx.Rollback() //nolint
	devices, err := models.UserDevices(
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).All(ctx, tx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve all devices and definitions for event generation")
	}
	for _, device := range devices {
		err = eventService.Emit(
			&services.Event{
				Type:    controllers.UserDeviceCreationEventType,
				Subject: device.UserID,
				Source:  "devices-api",
				Data: controllers.UserDeviceEvent{
					Timestamp: device.CreatedAt,
					UserID:    device.UserID,
					Device: services.UserDeviceEventDevice{
						ID:    device.ID,
						Make:  device.R.DeviceDefinition.R.DeviceMake.Name,
						Model: device.R.DeviceDefinition.Model,
						Year:  int(device.R.DeviceDefinition.Year),
					},
				},
			},
		)
		if err != nil {
			logger.Err(err).Msgf("Failed to emit creation event for device %s", device.ID)
		}
	}

	scints, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		qm.Load(models.UserDeviceAPIIntegrationRels.Integration),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.DeviceDefinition)),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice, models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).All(ctx, tx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve all active integrations")
	}
	for _, scint := range scints {
		if !scint.R.UserDevice.VinIdentifier.Valid {
			logger.Warn().Msgf("Device %s has an active integration but no VIN", scint.UserDeviceID)
			continue
		}
		if !scint.R.UserDevice.VinConfirmed {
			logger.Warn().Msgf("Device %s has an active integration but the VIN %s is unconfirmed", scint.UserDeviceID, scint.R.UserDevice.VinIdentifier.String)
			continue
		}
		err = eventService.Emit(
			&services.Event{
				Type:    "com.dimo.zone.device.integration.create",
				Subject: scint.UserDeviceID,
				Source:  "devices-api",
				Data: services.UserDeviceIntegrationEvent{
					Timestamp: scint.CreatedAt,
					UserID:    scint.R.UserDevice.UserID,
					Device: services.UserDeviceEventDevice{
						ID:    scint.UserDeviceID,
						Make:  scint.R.UserDevice.R.DeviceDefinition.R.DeviceMake.Name,
						Model: scint.R.UserDevice.R.DeviceDefinition.Model,
						Year:  int(scint.R.UserDevice.R.DeviceDefinition.Year),
						VIN:   scint.R.UserDevice.VinIdentifier.String,
					},
					Integration: services.UserDeviceEventIntegration{
						ID:     scint.R.Integration.ID,
						Type:   scint.R.Integration.Type,
						Style:  scint.R.Integration.Style,
						Vendor: scint.R.Integration.Vendor,
					},
				},
			},
		)
		if err != nil {
			logger.Err(err).Msgf("Failed to emit integration creation event for device %s", scint.UserDeviceID)
		}
	}

	err = tx.Commit()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to commit (kinda dumb)")
	}
}
