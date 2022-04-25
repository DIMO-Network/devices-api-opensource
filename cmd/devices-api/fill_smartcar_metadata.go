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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func fillSmartcarMetadata(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, scClient services.SmartcarClient, cipher shared.Cipher) error {
	scInteg, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("SmartCar")).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	activeInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scInteg.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNotNull(),
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

		decAccess, err := cipher.Decrypt(integ.AccessToken.String)
		if err != nil {
			logger.Err(err).Msg("Couldn't decrypt access token.")
			continue
		}

		perms, err := scClient.GetEndpoints(ctx, decAccess, integ.ExternalID.String)
		if err != nil {
			logger.Err(err).Msg("Token doesn't work. Did you refresh it?")
			continue
		}

		meta := services.UserDeviceAPIIntegrationsMetadata{SmartcarEndpoints: perms}
		b, err := json.Marshal(meta)
		if err != nil {
			logger.Err(err).Msg("Couldn't marshal endpoint data.")
			continue
		}

		integ.Metadata = null.JSONFrom(b)
		_, err = integ.Update(ctx, pdb.DBS().Writer, boil.Infer())
		if err != nil {
			logger.Err(err).Msg("Failed to update database row.")
			continue
		}

		success++
	}

	logger.Info().Msgf("Filled in %d/%d Metadata.", success, len(activeInts))

	return nil
}
