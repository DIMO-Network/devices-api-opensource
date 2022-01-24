package main

import (
	"context"
	"strconv"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/ahmetb/go-linq/v3"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const edmundsSource = "edmunds"

// loadEdmundsDeviceDefinitions default assumes an initial migration has already been done. This will only insert where
// it doesn't find a matching styleId and MMY level id. cleanDB assumes fresh db (eg. local) and want to insert everything
func loadEdmundsDeviceDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, cleanDB bool) error {
	//nhtsaSvc := services.NewNHTSAService()
	//ddSvc := services.NewDeviceDefinitionService(settings, pdb.DBS, &logger, nhtsaSvc)
	edmundsSvc := services.NewEdmundsService(settings.TorProxyURL, logger)

	vehicles, err := edmundsSvc.GetFlattenedVehicles()
	if err != nil {
		return err
	}
	//prefilter edmunds sourced data
	allDefinitions, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Source.EQ(null.StringFrom(edmundsSource))).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	tx, err := pdb.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint

	for _, vehicle := range *vehicles {
		// check if the device definition for the edmunds models.years.id exists
		if !cleanDB {
			deviceDefExists := linq.From(allDefinitions).WhereT(func(d models.DeviceDefinition) bool {
				return d.ExternalID.String == vehicle.ModelYearID && d.Source == null.StringFrom(edmundsSource)
			}).Any()
			if deviceDefExists {
				// ignore matching style id definition
				continue
			}
		}

		// no matching style found. Insert new Device Definition
		newDD := models.DeviceDefinition{
			ID:         ksuid.New().String(),
			Make:       vehicle.Make,
			Model:      vehicle.Model,
			Year:       int16(vehicle.Year),
			Source:     null.StringFrom(edmundsSource),
			Verified:   true,
			ExternalID: null.StringFrom(vehicle.ModelYearID),
		}
		err = newDD.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		for _, style := range vehicle.Styles {
			newDs := models.DeviceStyle{
				ID:                 ksuid.New().String(),
				DeviceDefinitionID: newDD.ID,
				Name:               style.Name,
				SubModel:           style.Trim,
				ExternalStyleID:    strconv.Itoa(style.StyleID),
				Source:             edmundsSource,
			}
			err = newDs.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
