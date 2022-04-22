package main

import (
	"context"

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

func migrateSmartcarTasks(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, taskService services.TaskService, smartcarTaskService services.SmartcarTaskService, cipher shared.Cipher) error {
	sc, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("SmartCar")).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	activeInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(sc.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNull(),
	).All(ctx, pdb.DBS().Reader.DB)
	if err != nil {
		return err
	}

	success := 0

	for _, integ := range activeInts {
		logger := logger.With().Str("userDeviceId", integ.UserDeviceID).Logger()

		rawAcces := integ.AccessToken.String
		var err error

		integ.TaskID = null.StringFrom(ksuid.New().String())

		encAccess, err := cipher.Encrypt(rawAcces)
		if err != nil {
			logger.Err(err).Msg("Couldn't encrypt access token.")
			continue
		}
		integ.AccessToken = null.StringFrom(encAccess)

		encRefresh, err := cipher.Encrypt(integ.RefreshToken.String)
		if err != nil {
			logger.Err(err).Msg("Couldn't encrypt refresh token.")
			continue
		}
		integ.RefreshToken = null.StringFrom(encRefresh)

		if err := taskService.StartSmartcarDeregistrationTasks(integ.UserDeviceID, sc.ID, integ.ExternalID.String, ""); err != nil {
			logger.Err(err).Msg("Failed to start webhook reregistration.")
			continue
		}

		_, err = integ.Update(ctx, pdb.DBS().Writer, boil.Infer())
		if err != nil {
			logger.Err(err).Msg("Couldn't update database record with task ID and encrypted credentials.")

		}

		if err := smartcarTaskService.StartPoll(integ); err != nil {
			logger.Err(err).Msg("Couldn't start new Kafka task.")
			continue
		}

		success++
	}

	logger.Info().Msgf("Migrated %d/%d Smartcar tasks.", success, len(activeInts))

	return nil
}
