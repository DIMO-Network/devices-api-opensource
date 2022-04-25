package main

import (
	"context"
	"encoding/json"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func migrateSmartcarPoll(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, scClient services.SmartcarClient, scTaskSvc services.SmartcarTaskService, scHook *services.SmartcarWebhookClient, cipher shared.Cipher) error {
	scInteg, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("SmartCar")).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	activeInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scInteg.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNull(),
	).All(ctx, pdb.DBS().Reader.DB)
	if err != nil {
		return err
	}

	success := 0

	for _, integ := range activeInts {
		logger := logger.With().Str("userDeviceId", integ.UserDeviceID).Logger()

		if !integ.ExternalID.Valid {
			logger.Error().Msg("No externalId. This should never happen!")
			continue
		}

		logger = logger.With().Str("externalId", integ.ExternalID.String).Logger()

		perms, err := scClient.GetEndpoints(ctx, integ.AccessToken.String, integ.ExternalID.String)
		if err != nil {
			logger.Err(err).Msg("Token doesn't work.")
			continue
		}

		meta := services.UserDeviceAPIIntegrationsMetadata{SmartcarEndpoints: perms}
		b, err := json.Marshal(meta)
		if err != nil {
			logger.Err(err).Msg("Couldn't marshal endpoint data.")
			continue
		}

		encAccess, err := cipher.Encrypt(integ.AccessToken.String)
		if err != nil {
			logger.Err(err).Msg("Couldn't encrypt access token.")
			continue
		}

		encRefresh, err := cipher.Encrypt(integ.RefreshToken.String)
		if err != nil {
			logger.Err(err).Msg("Couldn't encrypt access token.")
			continue
		}

		// Happily, this does not generate an event.
		if err := scHook.Unsubscribe(integ.ExternalID.String); err != nil {
			logger.Err(err).Msg("Couldn't detach existing webhook.")
			continue
		}

		integ.AccessToken = null.StringFrom(encAccess)
		integ.RefreshToken = null.StringFrom(encRefresh)
		integ.TaskID = null.StringFrom(ksuid.New().String())
		integ.Metadata = null.JSONFrom(b)
		_, err = integ.Update(ctx, pdb.DBS().Writer, boil.Infer())
		if err != nil {
			logger.Err(err).Msg("Failed to update database row.")
			continue
		}

		if err := scTaskSvc.StartPoll(integ); err != nil {
			logger.Err(err).Msg("Failed to start new task.")
			continue
		}

		logger.Info().Msg("Successfully migrated.")

		success++
	}

	logger.Info().Msgf("Migrated %d/%d Smartcar jobs.", success, len(activeInts))

	return nil
}
