package main

import (
	"context"
	"time"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func seedSmartcarUserID(ctx context.Context, logger *zerolog.Logger, pdb database.DbStore, cipher shared.Cipher) error {
	scInt, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(services.SmartCarVendor)).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	activeInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scInt.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNotNull(),
	).All(ctx, pdb.DBS().Reader.DB)
	if err != nil {
		return err
	}

	success := 0

	scClient := smartcar.NewClient()

	for _, integ := range activeInts {
		logger := logger.With().Str("userDeviceId", integ.UserDeviceID).Logger()

		if !integ.AccessExpiresAt.Valid {
			logger.Error().Msg("No access token expiry. Odd!")
			continue
		}

		if !integ.ExternalID.Valid {
			logger.Error().Msg("No external id. Odd!")
			continue
		}

		meta := new(services.UserDeviceAPIIntegrationsMetadata)
		if err := integ.Metadata.Unmarshal(meta); err != nil {
			logger.Err(err).Msg("Couldn't deserialize metadata.")
			continue
		}

		if meta.SmartcarUserID != nil && *meta.SmartcarUserID != "" {
			// Already treated.
			continue
		}

		if staleness := time.Now().Sub(integ.AccessExpiresAt.Time); staleness > 0 {
			logger.Info().Msgf("Access token expired by %s. Run this later.", staleness)
			continue
		}

		access, err := cipher.Decrypt(integ.AccessToken.String)
		if err != nil {
			logger.Err(err).Msg("Couldn't decrypt access token.")
			continue
		}

		vehIDs, err := scClient.GetVehicleIDs(ctx, &smartcar.VehicleIDsParams{Access: access})
		if err != nil {
			logger.Err(err).Msg("Couldn't get vehicle IDs.")
			continue
		}

		// This had better not be null!
		if l := len(*vehIDs); l != 1 {
			logger.Error().Msgf("Found %d vehicles, should only be one.", l)
			continue
		}

		vehID := (*vehIDs)[0]
		if vehID != integ.ExternalID.String {
			logger.Error().Msgf("We had %s stored, but the token says %s.", integ.ExternalID.String, vehID)
			continue
		}

		scUserID, err := scClient.GetUserID(ctx, &smartcar.UserIDParams{Access: access})
		if err != nil {
			logger.Err(err).Msg("Couldn't get user ID.")
			continue
		}

		meta.SmartcarUserID = scUserID
		if err := integ.Metadata.Marshal(meta); err != nil {
			logger.Err(err).Msg("Couldn't serialize modified metadata object.")
			continue
		}

		if _, err := integ.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			logger.Err(err).Msg("Couldn't update UDAI record.")
			continue
		}

		success++
	}

	logger.Info().Msgf("Restarted %d/%d Tesla jobs.", success, len(activeInts))

	return nil
}
