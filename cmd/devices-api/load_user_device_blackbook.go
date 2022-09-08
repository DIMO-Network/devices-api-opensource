package main

import (
	"context"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
)

// load user devices.
func loadUserDeviceBlackbook(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	// get all devices from DB.
	all, err := models.UserDevices(models.UserDeviceWhere.VinConfirmed.EQ(true)).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	logger.Info().Msgf("found %d user ud verified", len(all))

	deviceDefinitionSvc := services.NewDeviceDefinitionService(pdb.DBS, logger, nil, settings)

	for _, ud := range all {
		err = deviceDefinitionSvc.PullBlackbookData(ctx, ud.ID, ud.DeviceDefinitionID, ud.VinIdentifier.String)
		if err != nil {
			logger.Err(err).Msg("error pulling blackbook data")
			continue
		}
		logger.Info().Msgf("processed vin: %s", ud.VinIdentifier.String)
	}

	return nil
}
