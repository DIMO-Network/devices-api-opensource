package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/ahmetb/go-linq/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const edmundsSource = "edmunds"

// loadEdmundsDeviceDefinitions default assumes an initial migration has already been done. This will only insert where
// it doesn't find a matching styleId and MMY level id. mergeMMYMatch will lookup the definition by MMY and if find match attempt to merge instead of insert
func loadEdmundsDeviceDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, mergeMMYMatch bool) error {
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
		deviceDefExists := linq.From(allDefinitions).WhereT(func(d models.DeviceDefinition) bool {
			return d.ExternalID.String == strconv.Itoa(vehicle.ModelYearID) && d.Source == null.StringFrom(edmundsSource)
		}).Any()
		if deviceDefExists {
			// ignore matching style id definition
			continue
		}

		if mergeMMYMatch {
			// lookup ilike MMY, if match, update and continue loop
			matchingDD, err := models.DeviceDefinitions(
				qm.Where("year = ?", vehicle.Year),
				qm.And("make ilike ?", vehicle.Make),
				qm.And("model ilike ?", vehicle.Model)).One(ctx, pdb.DBS().Writer)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return errors.Wrap(err, "error querying for existing DD")
			}
			if err == nil && matchingDD != nil {
				// match, update it!
				fmt.Printf("Found exact match with: %s\n.", printMMY(matchingDD, Green))
				matchingDD.Make = vehicle.Make
				matchingDD.Model = vehicle.Model
				matchingDD.Source = null.StringFrom(edmundsSource)
				matchingDD.ExternalID = null.StringFrom(strconv.Itoa(vehicle.ModelYearID))
				matchingDD.Verified = true
				_, err = matchingDD.Update(ctx, pdb.DBS().Writer, boil.Infer())
				if err != nil {
					return errors.Wrap(err, "error updating device_definition with edmunds data")
				}
				// insert styles
				err = insertStyles(ctx, vehicle, matchingDD.ID, tx)
				if err != nil {
					return errors.Wrapf(err, "error inserting styles for device_definition_id: %s", matchingDD.ID)
				}

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
			ExternalID: null.StringFrom(strconv.Itoa(vehicle.ModelYearID)),
		}
		err = newDD.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		err = insertStyles(ctx, vehicle, newDD.ID, tx)
		if err != nil {
			return err
		}
	}
	logger.Info().Msg("Committing huge transaction. Please wait.")
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func insertStyles(ctx context.Context, vehicle services.FlatMMYDefinition, deviceDefinitionID string, tx *sql.Tx) error {
	for _, style := range vehicle.Styles {
		newDs := models.DeviceStyle{
			ID:                 ksuid.New().String(),
			DeviceDefinitionID: deviceDefinitionID,
			Name:               style.Name,
			SubModel:           style.Trim,
			ExternalStyleID:    strconv.Itoa(style.StyleID),
			Source:             edmundsSource,
		}
		err := newDs.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
	}
	return nil
}
