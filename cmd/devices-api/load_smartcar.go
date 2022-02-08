package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func loadSmartCarData(ctx context.Context, logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	apiSvc := services.NewSmartCarService("https://api.smartcar.com/v2.0/", pdb.DBS, logger)

	err := apiSvc.SeedDeviceDefinitionsFromSmartCar(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("error seeding device defs from smart car")
	}
}

func smartCarForwardCompatibility(ctx context.Context, logger zerolog.Logger, pdb database.DbStore) error {
	// get all device integrations where smartcar ordered by date asc.
	// then group them by make and model, so we can iterate over the years.
	// if there is a gap in the years, insert device_integration
	scSvc := services.NewSmartCarService("https://api.smartcar.com/v2.0/", pdb.DBS, logger)
	integrationID, err := scSvc.GetOrCreateSmartCarIntegration(ctx)
	if err != nil {
		return err
	}

	deviceDefs, err := models.DeviceDefinitions(qm.InnerJoin("device_integrations di on di.device_definition_id = device_definitions.id"),
		qm.Where("di.integration_id = ?", integrationID), qm.OrderBy("year, make, model DESC")).
		All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	lastMM := ""
	lastYear := int16(0)
	for _, dd := range deviceDefs {
		if lastMM == "" {
			lastMM = dd.Make + dd.Model
		}
		if lastYear == 0 {
			lastYear = dd.Year
		}
		thisMM := dd.Make + dd.Model
		if lastMM == thisMM {
			// we care about year gaps
			yearDiff := dd.Year - lastYear
			if yearDiff > 1 {
				// we have a gap
				fmt.Printf("found a year gap...\n")
				// look for a DD with matching MM and Y? If exists, then insert device integration
				gapDd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Year.EQ(lastYear+1),
					models.DeviceDefinitionWhere.Model.EQ(dd.Model),
					models.DeviceDefinitionWhere.Make.EQ(dd.Make)).One(ctx, pdb.DBS().Reader)
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				if err != nil {
					return err
				}
				// found a record that needs to be attached to integration
				fmt.Printf("found device def for year gap %s, inserting device_integration\n", printMMY(gapDd, Green, true))
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
			}
		}

		lastYear = dd.Year
		lastMM = dd.Make + dd.Model
	}

	return nil
}
