package main

import (
	"context"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/rs/zerolog"
)

func loadSmartCarData(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	apiSvc := services.NewSmartCarService("https://api.smartcar.com/v2.0/", pdb.DBS, logger)

	err := apiSvc.SeedDeviceDefinitionsFromSmartCar(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("error seeding device defs from smart car")
	}
}
