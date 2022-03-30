package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
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
func loadEdmundsDeviceDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	//nhtsaSvc := services.NewNHTSAService()
	//ddSvc := services.NewDeviceDefinitionService(settings, pdb.dbs, &logger, nhtsaSvc)
	edmundsSvc := services.NewEdmundsService(settings.TorProxyURL, logger)
	deviceDefSvc := services.NewDeviceDefinitionService(settings.TorProxyURL, pdb.DBS, logger, nil)

	latestEdmunds, err := edmundsSvc.GetFlattenedVehicles()
	if err != nil {
		return err
	}
	//prefilter edmunds sourced data
	existingDDsEdmundsSrc, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Source.EQ(null.StringFrom(edmundsSource)),
		qm.OrderBy("id")).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	fmt.Printf("Found %d existing edmunds Device Definitions\n", len(existingDDsEdmundsSrc))

	tx, err := pdb.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint

	for _, edmundVehicle := range *latestEdmunds {
		// check if the device definition for the edmunds models.years.id exists
		deviceDefExists := linq.From(existingDDsEdmundsSrc).WhereT(func(d *models.DeviceDefinition) bool {
			return d.ExternalID.String == strconv.Itoa(edmundVehicle.ModelYearID) && d.Source.String == edmundsSource
		}).Any()
		if deviceDefExists {
			// find the existing DD in our list
			existingDD := linq.From(existingDDsEdmundsSrc).WhereT(func(d *models.DeviceDefinition) bool {
				return d.ExternalID.String == strconv.Itoa(edmundVehicle.ModelYearID) && d.Source == null.StringFrom(edmundsSource)
			}).First().(*models.DeviceDefinition)
			// insert styles deduping
			err = insertStyles(ctx, logger, edmundVehicle, existingDD.ID, tx)
			if err != nil {
				return errors.Wrapf(err, "error inserting styles for device_definition_id: %s", existingDD.ID)
			}

			// back to loop since don't want to insert MMY
			continue
		}

		// lookup ilike MMY, if match, update, insert styles and continue loop, need to do this to avoid inserting duplicate MMY
		matchingDD, err := deviceDefSvc.FindDeviceDefinitionByMMY(ctx, tx, edmundVehicle.Make, edmundVehicle.Model, edmundVehicle.Year, false)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "error querying for existing DD")
		}
		dm, err := deviceDefSvc.GetOrCreateMake(ctx, tx, edmundVehicle.Make)
		if err != nil {
			return errors.Wrap(err, "error quering for make")
		}

		if err == nil && matchingDD != nil {
			// match, update it!
			fmt.Printf("Found MMY match with: %s. dd_id: %s\n.", printMMY(matchingDD, Green, true), matchingDD.ID)
			if matchingDD.Source.String == edmundsSource {
				fmt.Printf("weird, MMY match but external ID's dont. existing external id: %s, edmunds model year id: %d. updating to match\n",
					matchingDD.ExternalID.String, edmundVehicle.ModelYearID)
			}

			matchingDD.DeviceMakeID = dm.ID
			matchingDD.Model = edmundVehicle.Model
			matchingDD.Source = null.StringFrom(edmundsSource)
			matchingDD.ExternalID = null.StringFrom(strconv.Itoa(edmundVehicle.ModelYearID))
			matchingDD.Verified = true
			_, err = matchingDD.Update(ctx, tx, boil.Infer())
			if err != nil {
				return errors.Wrap(err, "error updating device_definition with edmunds data")
			}
			// insert styles
			err = insertStyles(ctx, logger, edmundVehicle, matchingDD.ID, tx)
			if err != nil {
				return errors.Wrapf(err, "error inserting styles for device_definition_id: %s", matchingDD.ID)
			}

			continue
		}

		// no matching style found. Insert new Device Definition
		newDD := models.DeviceDefinition{
			ID:           ksuid.New().String(),
			DeviceMakeID: dm.ID,
			Model:        edmundVehicle.Model,
			Year:         int16(edmundVehicle.Year),
			Source:       null.StringFrom(edmundsSource),
			Verified:     true,
			ExternalID:   null.StringFrom(strconv.Itoa(edmundVehicle.ModelYearID)),
		}
		err = newDD.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return err
		}
		err = insertStyles(ctx, logger, edmundVehicle, newDD.ID, tx)
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

func insertStyles(ctx context.Context, logger *zerolog.Logger, vehicle services.FlatMMYDefinition, deviceDefinitionID string, tx *sql.Tx) error {
	// get styles
	existingStyles, err := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(deviceDefinitionID)).All(ctx, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err, "error looking for styles for device_definition: %s", deviceDefinitionID)
	}
	// loop and compare
	for _, edmundsStyle := range vehicle.Styles {
		matchFound := false
		for _, existingStyle := range existingStyles {
			if existingStyle.ExternalStyleID == strconv.Itoa(edmundsStyle.StyleID) {
				matchFound = true
			}
		}
		if !matchFound {
			// insert edmundsStyle
			newStyle := models.DeviceStyle{
				ID:                 ksuid.New().String(),
				DeviceDefinitionID: deviceDefinitionID,
				Name:               edmundsStyle.Name,
				ExternalStyleID:    strconv.Itoa(edmundsStyle.StyleID),
				Source:             edmundsSource,
				SubModel:           edmundsStyle.Trim,
			}
			err = newStyle.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return errors.Wrapf(err, "error inserting new device style %+v", edmundsStyle)
			}
			logger.Info().Msgf("inserted new style: %s %s for existing dd: %s", newStyle.Name, newStyle.SubModel, deviceDefinitionID)
		}
	}
	return nil
}
