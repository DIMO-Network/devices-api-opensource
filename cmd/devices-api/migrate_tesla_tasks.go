package main

import (
	"context"
	"strconv"

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

func migrateTeslaTasks(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, teslaSvc services.TeslaService, teslaTask services.TeslaTaskService, cipher shared.Cipher) error {
	tesla, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("Tesla")).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	activeInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(tesla.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNull(),
	).All(ctx, pdb.DBS().Reader.DB)
	if err != nil {
		return err
	}

	success := 0

	for _, integ := range activeInts {
		access, err := cipher.Decrypt(integ.AccessToken.String)
		if err != nil {
			logger.Err(err).Str("userDevice", integ.UserDeviceID).Msg("Couldn't decrypt access token.")
			continue
		}
		intID, err := strconv.Atoi(integ.ExternalID.String)
		if err != nil {
			logger.Err(err).Str("userDevice", integ.UserDeviceID).Msgf("External ID %s wasn't an integer.", integ.ExternalID.String)
			continue
		}
		v, err := teslaSvc.GetVehicle(access, intID)
		if err != nil {
			logger.Err(err).Str("userDeviceId", integ.UserDeviceID).Msg("Couldn't get vehicle information from Tesla.")
			continue
		}
		if err := teslaTask.StopPoll(integ); err != nil {
			logger.Err(err).Str("userDeviceId", integ.UserDeviceID).Msg("Failed to stop old task.")
			continue
		}
		integ.TaskID = null.StringFrom(ksuid.New().String())
		if err := teslaTask.StartPoll(v, integ); err != nil {
			logger.Err(err).Str("userDeviceId", integ.UserDeviceID).Msg("Failed to start new task.")
			continue
		}
		if _, err = integ.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			logger.Err(err).Str("userDeviceId", integ.UserDeviceID).Msg("Failed to update integration record.")
		}

		success++
	}

	logger.Info().Msgf("Migrated %d/%d Teslas.", success, len(activeInts))

	return nil
}
