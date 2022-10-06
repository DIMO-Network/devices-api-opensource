package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/database"
	es "github.com/DIMO-Network/devices-api/internal/elasticsearch"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func populateESRegionData(ctx context.Context, settings *config.Settings, elastic es.ElasticSearch, pdb database.DbStore, logger *zerolog.Logger) error {
	db := pdb.DBS().Reader

	userDevices, err := models.UserDevices(
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all user devices: %w", err)
	}

	for _, ud := range userDevices {
		logger := logger.With().Str("command", "populate-es-region-data").Str("userDeviceId", ud.ID).Logger()

		if len(ud.R.UserDeviceAPIIntegrations) == 0 {
			continue
		}

		integrationID := ud.R.UserDeviceAPIIntegrations[0].IntegrationID

		if ud.CountryCode.IsZero() {
			logger.Error().Msg("Device has an integration but is missing country.")
			continue
		}

		md := services.UserDeviceMetadata{}
		if err = ud.Metadata.Unmarshal(&md); err != nil {
			logger.Err(err).Msg("Could not unmarshal user device metadata.")
			continue
		}

		if md.ElasticRegionSynced {
			logger.Debug().Msgf("Records have already been updated for this device.")
			continue
		}

		dd := services.DeviceDefinitionDTO{
			DeviceDefinitionID: ud.DeviceDefinitionID,
			IntegrationID:      "dimo/integration/" + integrationID,
			UserDeviceID:       ud.ID,
		}

		country := constants.FindCountry(ud.CountryCode.String)
		if country == nil || country.Region == "" {
			logger.Error().Str("country", ud.CountryCode.String).Msg("Could not get region from device's country.")
			continue
		}

		err = elastic.UpdateDeviceRegionsByQuery(dd, country.Region, settings.ElasticDeviceStatusIndex)
		if err != nil {
			logger.Err(err).Msgf("Error occurred during Elastic region update.")
			continue
		}

		md.ElasticRegionSynced = true
		err = ud.Metadata.Marshal(&md)
		if err != nil {
			logger.Error().Msgf("Could not marshal updated metadata.")
			continue
		}

		if _, err := ud.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			logger.Err(err).Msg("Error updating device metadata.")
			continue
		}

	}

	return nil
}
