package main

import (
	"context"
	"fmt"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// remakeDeviceDefinitionTopics invokes [services.DeviceDefinitionRegistrar] for each user device
// with an integration.
func remakeDeviceDefinitionTopics(ctx context.Context, settings *config.Settings, pdb database.DbStore, producer sarama.SyncProducer, logger *zerolog.Logger, ddSvc services.DeviceDefinitionService) error {
	reg := services.NewDeviceDefinitionRegistrar(producer, settings)
	db := pdb.DBS().Reader

	// Find all integrations instances.
	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.ExternalID.IsNotNull(),
		qm.Load(
			qm.Rels(
				models.UserDeviceAPIIntegrationRels.UserDevice,
			),
		),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve integration instances: %w", err)
	}

	failures := 0

	ids := []string{}
	for _, d := range apiInts {
		ids = append(ids, d.R.UserDevice.DeviceDefinitionID)
	}

	deviceDefinitionResponse, err := ddSvc.GetDeviceDefinitionsByIDs(ctx, ids)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve all devices and definitions for event generation from grpc")
	}

	filterDeviceDefinition := func(id string, items []*ddgrpc.GetDeviceDefinitionItemResponse) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
		for _, dd := range items {
			if id == dd.DeviceDefinitionId {
				return dd, nil
			}
		}
		return nil, errors.Errorf("no device definition %s", id)
	}

	// For each of these, register the device's device definition with the data pipeline.
	for _, apiInt := range apiInts {

		ddInfo, err := filterDeviceDefinition(apiInt.R.UserDevice.DeviceDefinitionID, deviceDefinitionResponse)
		if err != nil {
			logger.Fatal().Err(err)
			continue
		}

		userDeviceID := apiInt.UserDeviceID

		region := ""

		if country := apiInt.R.UserDevice.CountryCode; country.Valid {
			countryData := constants.FindCountry(country.String)
			if countryData != nil {
				region = countryData.Region
			}
		}

		ddReg := services.DeviceDefinitionDTO{
			UserDeviceID:       userDeviceID,
			DeviceDefinitionID: ddInfo.DeviceDefinitionId,
			IntegrationID:      apiInt.IntegrationID,
			Make:               ddInfo.Type.Make,
			Model:              ddInfo.Type.Model,
			Year:               int(ddInfo.Type.Year),
			Region:             region,
			MakeSlug:           ddInfo.Type.MakeSlug,
			ModelSlug:          ddInfo.Type.ModelSlug,
		}

		err = reg.Register(ddReg)
		if err != nil {
			logger.Err(err).Str("userDeviceId", userDeviceID).Msg("Failed to register device's device definition.")
			failures++
		}
	}

	logger.Info().Int("attempted", len(apiInts)).Int("failed", failures).Msg("Finished device definition registration.")

	return nil
}
