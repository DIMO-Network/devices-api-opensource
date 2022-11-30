package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
)

// loadUserDeviceDrively iterates over user_devices with vin verified and tries pulling data from drivly
func loadUserDeviceDrively(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, forceSetAll bool, pdb database.DbStore) error {
	// get all devices from DB.
	all, err := models.UserDevices(models.UserDeviceWhere.VinConfirmed.EQ(true)).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	logger.Info().Msgf("processing %d user_devices with verified VINs", len(all))

	deviceDefinitionSvc := services.NewDeviceDefinitionService(pdb.DBS, logger, nil, settings)
	statsAggr := map[services.DrivlyDataStatusEnum]int{}
	for _, ud := range all {
		status, err := deviceDefinitionSvc.PullDrivlyData(ctx, ud.ID, ud.DeviceDefinitionID, ud.VinIdentifier.String, forceSetAll)
		if err != nil {
			logger.Err(err).Str("vin", ud.VinIdentifier.String).Msg("error pulling drivly data")
		} else {
			logger.Info().Msgf("processed vin: %s", ud.VinIdentifier.String)
		}
		statsAggr[status]++
	}
	fmt.Println("-------------------RUN SUMMARY--------------------------")
	// colorize each result
	fmt.Printf("Total VINs processed: %d \n", len(all))
	fmt.Printf("New Drivly Pulls (vin + valuations): %d \n", statsAggr[services.PulledAllDrivlyStatus])
	fmt.Printf("Pulled New Pricing & Offers: %d \n", statsAggr[services.PulledValuationDrivlyStatus])
	fmt.Printf("SkippedDrivlyStatus due to biz logic: %d \n", statsAggr[services.SkippedDrivlyStatus])
	fmt.Printf("SkippedDrivlyStatus due to error: %d \n", statsAggr[""])
	fmt.Println("--------------------------------------------------------")
	return nil
}
