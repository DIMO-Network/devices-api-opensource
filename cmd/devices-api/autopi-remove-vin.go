package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
)

func processRemoveVINFromAutopi(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	// instantiate
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)

	// iterate all autopi units
	all, err := models.AutopiUnits().All(ctx, pdb.DBS().Reader)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to query db")
	}
	logger.Info().Msgf("processing %d autopi units", len(all))

	for _, unit := range all {
		innerLogger := logger.With().Str("autopiUnitID", unit.AutopiUnitID).Logger()

		autoPiDevice, err := autoPiSvc.GetDeviceByUnitID(unit.AutopiUnitID)
		if err != nil {
			innerLogger.Err(err).Msg("failed to call autopi api to get autoPiDevice")
			continue
		}
		if autoPiDevice == nil c {
			innerLogger.Info().Msg("skipped due to nil")
			continue
		}
		// call api svc to update profile, setting vin = ""
		err = autoPiSvc.PatchVehicleProfile(autoPiDevice.Vehicle.ID, services.PatchVehicleProfile{
			Vin: "",
		})
		if err != nil {
			// uh oh spaghettie oh
			innerLogger.Err(err).Msg("failed to set VIN on autopi service")
		}
	}

	logger.Info().Msg("all done")
}
