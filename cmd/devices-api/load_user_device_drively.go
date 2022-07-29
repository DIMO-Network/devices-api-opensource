package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// load user devices.
func loadUserDeviceDrively(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	// get all devices from DB.
	all, err := models.UserDevices(models.UserDeviceWhere.VinConfirmed.EQ(true)).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	logger.Info().Msgf("found %d user ud verified", len(all))

	drivlyService := services.NewDrivlyAPIService(settings, pdb.DBS)

	deviceDefIDs := make([]string, len(all)) // preallocate for all, but likely won't hit max size
	for _, ud := range all {
		if contains(deviceDefIDs, ud.DeviceDefinitionID) {
			logger.Info().Msgf("DeviceDefinitionID %s already exists, skipping", ud.DeviceDefinitionID)
			continue
		}
		deviceDefIDs = append(deviceDefIDs, ud.DeviceDefinitionID)
		// should we do a VIN checksum before, a lot of these seem to be just failed vin checksum
		deviceDefinition, err := models.FindDeviceDefinition(ctx, pdb.DBS().Reader, ud.DeviceDefinitionID)
		if err != nil {
			return err
		}
		// skip if already pulled this data in last 2 weeks
		existingData, err := models.DrivlyData(
			models.DrivlyDatumWhere.DeviceDefinitionID.EQ(null.StringFrom(ud.DeviceDefinitionID)),
			qm.OrderBy("updated_at desc"), qm.Limit(1)).
			One(context.Background(), pdb.DBS().Writer)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if existingData != nil && existingData.UpdatedAt.Add(time.Hour*24*14).After(time.Now()) {
			log.Info().Msgf("skipping device definition %s %s %d since already pulled in last 2 weeks",
				deviceDefinition.ID, deviceDefinition.Model, deviceDefinition.Year)
			deviceDefIDs = append(deviceDefIDs, ud.DeviceDefinitionID)
			continue
		}

		vinInfo, err := drivlyService.GetVINInfo(ud.VinIdentifier.String)
		if err != nil {
			log.Err(err).Msgf("error getting VIN %s. skipping", ud.VinIdentifier.String)
			continue
		}

		logger.Info().Msgf("DeviceDefinitionID Year %d Model %s", deviceDefinition.Year, deviceDefinition.Model)

		metaData := new(services.DeviceVehicleInfo) // make as pointer
		if err := deviceDefinition.Metadata.Unmarshal(metaData); err == nil {
			if vinInfo["mpgCity"] != nil {
				metaData.MPGCity = fmt.Sprintf("%f", vinInfo["mpgCity"])
			}
			if vinInfo["mpgHighway"] != nil {
				metaData.MPGHighway = fmt.Sprintf("%f", vinInfo["mpgHighway"])
			}
			if vinInfo["fuelTankCapacityGal"] != nil {
				metaData.FuelTankCapacityGal = fmt.Sprintf("%f", vinInfo["fuelTankCapacityGal"])
			}
		}
		err = deviceDefinition.Metadata.Marshal(metaData)
		if err != nil {
			return err
		}
		// todo future: set the device_style_id based on the edmunds response, will need gjson probably.

		_, err = deviceDefinition.Update(ctx, pdb.DBS().Writer, boil.Infer())
		if err != nil {
			return err
		}
		// insert drivly raw json data
		drivlyData := &models.DrivlyDatum{
			ID:                 ksuid.New().String(),
			DeviceDefinitionID: null.StringFrom(deviceDefinition.ID),
			Vin:                ud.VinIdentifier.String,
			UserDeviceID:       null.StringFrom(ud.ID),
		}

		summary, err := drivlyService.GetSummaryByVIN(ud.VinIdentifier.String)
		if err != nil {
			logger.Err(err).Msg("error getting summary for vin for all sources") // just continue if problem here
		}
		// does martiallying nil object cause crash?
		_ = drivlyData.VinMetadata.Marshal(vinInfo)
		_ = drivlyData.BuildMetadata.Marshal(summary.Build)
		_ = drivlyData.AutocheckMetadata.Marshal(summary.AutoCheck)
		_ = drivlyData.CargurusMetadata.Marshal(summary.Cargurus)
		_ = drivlyData.CarmaxMetadata.Marshal(summary.Carmax)
		_ = drivlyData.KBBMetadata.Marshal(summary.KBB)
		_ = drivlyData.CarstoryMetadata.Marshal(summary.Carstory)
		_ = drivlyData.CarvanaMetadata.Marshal(summary.Carvana)
		_ = drivlyData.EdmundsMetadata.Marshal(summary.Edmunds)
		_ = drivlyData.OfferMetadata.Marshal(summary.Offers)
		_ = drivlyData.TMVMetadata.Marshal(summary.TMV)
		_ = drivlyData.VroomMetadata.Marshal(summary.VRoom)
		//_ = drivlyData.PricingMetadata.Marshal(summary.Pricing) todo

		err = drivlyData.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
		if err != nil {
			return err
		}

		// todo future: did MMY from vininfo match the device definition?
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
