package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func restartSmartcarTasks(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, scTaskSvc services.SmartcarTaskService) error {
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

		if err := scTaskSvc.StopPoll(integ); err != nil {
			logger.Err(err).Msg("Couldn't stop old job.")
			continue
		}

		integ.TaskID = null.StringFrom(ksuid.New().String())
		if _, err := integ.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			logger.Err(err).Msg("Couldn't update record.")
			continue
		}

		if err := scTaskSvc.StopPoll(integ); err != nil {
			logger.Err(err).Msg("Couldn't start new job.")
			continue
		}

		success++
	}

	logger.Info().Msgf("Restarted %d/%d Smartcar tasks.", success, len(activeInts))

	return nil
}
