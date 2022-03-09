package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func syncSmartCarCompatibility(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	smartCarSvc := services.NewSmartCarService(pdb.DBS, logger)

	smartCarVehicleData, err := services.GetSmartCarVehicleData()
	if err != nil {
		logger.Fatal().Err(err)
	}

	err = setIntegrationForMatchingMakeYears(ctx, &smartCarSvc, pdb, smartCarVehicleData)
	if err != nil {
		logger.Fatal().Err(err)
	}
}

// setIntegrationForMatchingMakeYears will set the smartcar device_integration for any matching Year and Make DD existing in our database
func setIntegrationForMatchingMakeYears(ctx context.Context, smartCarSvc *services.SmartCarService, pdb database.DbStore,
	data *services.SmartCarCompatibilityData) error {
	scIntegrationID, err := smartCarSvc.GetOrCreateSmartCarIntegration(ctx)
	if err != nil {
		return err
	}
	// get all of our devices, that do not have a smartcar integration set, years 2012+
	deviceDefs, err := models.DeviceDefinitions(qm.LeftOuterJoin("device_integrations di on di.device_definition_id = device_definitions.id"),
		qm.Where("di is null or di.integration_id != ?", scIntegrationID), qm.And("year >= 2012")).
		All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	fmt.Printf("found %d device definitions that are not attached to smartcar integration\n", len(deviceDefs))

	makeYears := make(map[string][]int)

	for _, usData := range data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData.US {
		vehicleMake := usData.Name
		if strings.Contains(vehicleMake, "Nissan") || strings.Contains(vehicleMake, "Hyundai") || strings.Contains(vehicleMake, "All makes") {
			continue // skip if nissan or hyundai b/c not really supported
		}

		for _, row := range usData.Rows {
			years := row[0].Subtext // eg. 2017+ or 2012-2017

			if years == nil {
				fmt.Println("Skipping row as years is nil")
				continue
			}

			yearRange, err := services.ParseSmartCarYears(years)
			if err != nil {
				return errors.Wrapf(err, "could not parse years: %s", *years)
			}
			//fmt.Printf("found year range %s for make %s\n", *years, vehicleMake) // for debugging

			for _, yr := range yearRange {
				// update the map with the years
				// check if yr exists in map first
				yrExists := false
				if makeYears[vehicleMake] != nil {
					for _, y := range makeYears[vehicleMake] {
						if y == yr {
							yrExists = true
							break
						}
					}
				}
				if !yrExists {
					makeYears[vehicleMake] = append(makeYears[vehicleMake], yr)
				}
			}
		}
	}

	fmt.Printf("built up a map with %d makes and year ranges\n", len(makeYears))

	for mk, years := range makeYears {
		fmt.Printf("%s- processing make %s -%s", Green, mk, Reset)
		for _, year := range years {
			// get list of all device defs that have this yr and make.
			filtered := filterDeviceDefs(deviceDefs, mk, year)
			fmt.Printf("found %d device defs for year %d and make %s that don't have smartcar\n", len(filtered), year, mk)
			// insert device_integration for each one
			for _, definition := range filtered {
				di := models.DeviceIntegration{
					DeviceDefinitionID: definition.ID,
					IntegrationID:      scIntegrationID,
					Country:            "USA",
				}
				err = di.Upsert(ctx, pdb.DBS().Writer, false,
					[]string{"device_definition_id", "integration_id", "country"}, boil.Infer(), boil.Infer())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func filterDeviceDefs(items models.DeviceDefinitionSlice, make string, year int) []models.DeviceDefinition {
	var filtered []models.DeviceDefinition
	for _, item := range items {
		if item.Year == int16(year) && strings.EqualFold(item.R.DeviceMake.Name, make) {
			filtered = append(filtered, *item)
		}
	}
	return filtered
}

func smartCarForwardCompatibility(ctx context.Context, logger zerolog.Logger, pdb database.DbStore) error {
	// get all device integrations where smartcar ordered by date asc.
	// then group them by make and model, so we can iterate over the years.
	// if there is a gap in the years, insert device_integration
	scSvc := services.NewSmartCarService(pdb.DBS, logger)
	integrationID, err := scSvc.GetOrCreateSmartCarIntegration(ctx)
	deviceDefSvc := services.NewDeviceDefinitionService("", pdb.DBS, &logger, nil)

	if err != nil {
		return err
	}

	deviceDefs, err := models.DeviceDefinitions(qm.InnerJoin("device_integrations di on di.device_definition_id = device_definitions.id"),
		qm.Where("di.integration_id = ?", integrationID), qm.OrderBy("make, model, year DESC"),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).
		All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	fmt.Printf("found %d device definitions with smartcar integration\n", len(deviceDefs))

	lastMM := ""
	lastYear := int16(0)
	// meant to be used at the end of each loop to update the "last" values
	funcLastValues := func(dd *models.DeviceDefinition) {
		lastYear = dd.Year
		lastMM = dd.R.DeviceMake.Name + dd.Model
	}
	// year will be descending
	for _, dd := range deviceDefs {
		thisMM := dd.R.DeviceMake.Name + dd.Model
		if lastMM == thisMM {
			// we care about year gaps
			yearDiff := lastYear - dd.Year
			if yearDiff > 1 {
				// we have a gap
				fmt.Printf("%s found a year gap of %d...\n", thisMM, yearDiff)
				for i := int16(1); i < yearDiff; i++ {
					gapDd, err := deviceDefSvc.FindDeviceDefinitionByMMY(ctx, pdb.DBS().Reader, dd.R.DeviceMake.Name, dd.Model, int(dd.Year+i), true)
					if errors.Is(err, sql.ErrNoRows) {
						continue // this continues internal loop, so funcLastValues will still get set at end of outer loop
					}
					if err != nil {
						return err
					}
					// found a record that needs to be attached to integration
					if len(gapDd.R.DeviceIntegrations) == 0 {
						fmt.Printf("found device def for year gap %s, inserting smartcar device_integration\n", printMMY(gapDd, Green, true))
						diGap := models.DeviceIntegration{
							DeviceDefinitionID: gapDd.ID,
							IntegrationID:      integrationID,
							Country:            "USA",       // default
							Capabilities:       null.JSON{}, // we'd need to copy from previous dd?
						}
						err = diGap.Insert(ctx, pdb.DBS().Writer, boil.Infer())
						if err != nil {
							return errors.Wrap(err, "error inserting device_integration")
						}
					} else {
						fmt.Printf("%s already had an integration set\n", printMMY(gapDd, Red, true))
					}
				}
			}
		} else {
			// this should mean we are back at the start of a new make/model starting at highest year
			nextYearDd, err := deviceDefSvc.FindDeviceDefinitionByMMY(ctx, pdb.DBS().Writer, dd.R.DeviceMake.Name, dd.Model, int(dd.Year+1), true)
			if errors.Is(err, sql.ErrNoRows) {
				funcLastValues(dd)
				continue
			}
			if err != nil {
				return err
			}
			// does it have any integrations?
			if len(nextYearDd.R.DeviceIntegrations) == 0 {
				// attach smartcar integration
				fmt.Printf("found device def for future year %s, that does not have any integrations. inserting device_integration\n", printMMY(nextYearDd, Green, true))
				diGap := models.DeviceIntegration{
					DeviceDefinitionID: nextYearDd.ID,
					IntegrationID:      integrationID,
					Country:            "USA",       // default
					Capabilities:       null.JSON{}, // we'd need to copy from previous dd?
				}
				err = diGap.Insert(ctx, pdb.DBS().Writer, boil.Infer())
				if err != nil {
					return errors.Wrap(err, "error inserting device_integration")
				}
			}
		}

		funcLastValues(dd)
	}

	return nil
}

/* Terminal colors */

var Red = "\033[31m"
var Reset = "\033[0m"
var Green = "\033[32m"
var Purple = "\033[35m"

func printMMY(definition *models.DeviceDefinition, color string, includeSource bool) string {
	mk := ""
	if definition.R != nil && definition.R.DeviceMake != nil {
		mk = definition.R.DeviceMake.Name
	}
	if !includeSource {
		return fmt.Sprintf("%s%d %s %s%s", color, definition.Year, mk, definition.Model, Reset)
	}
	return fmt.Sprintf("%s%d %s %s %s(source: %s)%s",
		color, definition.Year, mk, definition.Model, Purple, definition.Source.String, Reset)
}
