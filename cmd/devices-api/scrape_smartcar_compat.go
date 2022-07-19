package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const sunsetYearCutoff = 2017

func syncSmartCarCompatibility(ctx context.Context, logger zerolog.Logger, pdb database.DbStore) {
	smartCarSvc := services.NewSmartCarService(pdb.DBS, logger)

	smartCarVehicleData, err := services.GetSmartCarVehicleData()
	if err != nil {
		logger.Fatal().Err(err)
	}

	err = setIntegrationForMatchingMakeYears(ctx, logger, &smartCarSvc, pdb, smartCarVehicleData)
	if err != nil {
		logger.Fatal().Err(err)
	}
}

// setIntegrationForMatchingMakeYears will set the smartcar device_integration for any matching Year and Make DD existing in our database
func setIntegrationForMatchingMakeYears(ctx context.Context, logger zerolog.Logger, smartCarSvc *services.SmartCarService, pdb database.DbStore,
	data *services.SmartCarCompatibilityData) error {
	scIntegrationID, err := smartCarSvc.GetOrCreateSmartCarIntegration(ctx)
	if err != nil {
		return err
	}

	for shortRegion, data := range data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData {
		region := ""
		switch shortRegion {
		case "US":
			region = services.AmericasRegion.String()
		case "EU":
			region = services.EuropeRegion.String()
		default:
			continue
		}

		regionLogger := logger.With().Str("region", region).Logger()
		sunsetSkipCount := 0

		for _, datum := range data {
			if datum.Name == "All makes" {
				for i, row := range datum.Rows {
					mkName := row[0].Text
					if mkName == nil {
						logger.Error().Msgf("No make name at row %d", i)
						continue
					}

					mkLogger := regionLogger.With().Str("make", *mkName).Logger()
					rangeStr := row[0].Subtext
					if rangeStr == nil || *rangeStr == "" {
						mkLogger.Error().Msg("Empty year range string, skipping manufacturer")
						continue
					}
					// Currently this describes Hyundai and Nissan.
					if strings.HasSuffix(*rangeStr, " (contact us)") {
						continue
					}
					startYear, err := strconv.Atoi((*rangeStr)[:len(*rangeStr)-1])
					if err != nil {
						mkLogger.Err(err).Msg("Couldn't parse range string, skipping")
						continue
					}
					if startYear < 2000 {
						mkLogger.Error().Msgf("Start year %d is suspiciously low, skipping", startYear)
						continue
					}

					dbMk, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(*mkName)).One(ctx, pdb.DBS().Writer)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							mkLogger.Warn().Msg("No make with this name found in the database, skipping")
							continue
						}
						return fmt.Errorf("database failure: %w", err)
					}
					dds, err := dbMk.DeviceDefinitions(
						qm.LeftOuterJoin(models.TableNames.DeviceIntegrations+" ON "+models.DeviceIntegrationTableColumns.DeviceDefinitionID+" = "+models.DeviceDefinitionTableColumns.ID+" AND "+models.DeviceIntegrationTableColumns.Region+" = ?", region),
						qm.Where(models.TableNames.DeviceIntegrations+" IS NULL"),
						models.DeviceDefinitionWhere.Year.GTE(int16(startYear)),
					).All(ctx, pdb.DBS().Writer)
					if err != nil {
						return fmt.Errorf("database error: %w", err)
					}

					if len(dds) == 0 {
						continue
					}
					mkLogger.Info().Msgf("Planning to insert %d compatibility records from %d onward", len(dds), startYear)

					for _, dd := range dds {
						if dd.Year < sunsetYearCutoff {
							// skipping as likelihood of being a 3g sunset vehicle
							sunsetSkipCount++
							continue
						}
						if err := dd.AddDeviceIntegrations(ctx, pdb.DBS().Writer.DB, true, &models.DeviceIntegration{
							DeviceDefinitionID: dd.ID,
							IntegrationID:      scIntegrationID,
							Region:             region,
						}); err != nil {
							return fmt.Errorf("database failure: %w", err)
						}
					}
				}
			}
		}
		regionLogger.Info().Msgf("skipped %d device definitions before year %d due to likelihood of 3g sunset", sunsetSkipCount, sunsetYearCutoff)
	}
	return nil
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
							Region:             services.AmericasRegion.String(), // default
							Capabilities:       null.JSON{},                      // we'd need to copy from previous dd?
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
					Region:             services.AmericasRegion.String(), // default
					Capabilities:       null.JSON{},                      // we'd need to copy from previous dd?
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
