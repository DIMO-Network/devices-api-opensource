package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func loadEdmundsImages(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, overwrite bool) {
	nhtsaSvc := services.NewNHTSAService()
	ddSvc := services.NewDeviceDefinitionService(pdb.DBS, &logger, nhtsaSvc, settings)
	var all models.DeviceDefinitionSlice
	var err error

	if overwrite {
		all, err = models.DeviceDefinitions(qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, pdb.DBS().Writer)
	} else {
		all, err = models.DeviceDefinitions(qm.Load(models.DeviceDefinitionRels.DeviceMake),
			models.DeviceDefinitionWhere.ImageURL.IsNull()).All(ctx, pdb.DBS().Writer)
	}
	total := len(all)
	logger.Info().Msgf("Found %d device definitions to process", total)

	if err != nil {
		logger.Fatal().Err(err).Msg("could not query all")
	}
	for i, definition := range all {
		err = ddSvc.CheckAndSetImage(definition, overwrite)
		if err != nil {
			logger.Error().Err(err).Msgf("could not find image for vehicle %s %s %d", definition.R.DeviceMake.Name, definition.Model, definition.Year)
		}
		if definition.ImageURL.Ptr() != nil {
			logger.Info().Msgf("%d of %d: replacing image_url for %s %s %d", i, total, definition.R.DeviceMake.Name, definition.Model, definition.Year)
		}
		_, err = definition.Update(ctx, pdb.DBS().Writer, boil.Whitelist(models.DeviceDefinitionColumns.ImageURL, models.DeviceDefinitionColumns.UpdatedAt))
		if err != nil {
			logger.Fatal().Err(err).Msg("could not update device definition in DB")
		}
	}
}
