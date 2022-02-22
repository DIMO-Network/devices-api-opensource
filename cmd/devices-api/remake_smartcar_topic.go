package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

func remakeSmartcarTopic(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, producer sarama.SyncProducer) error {
	reg := services.SmartcarIngestRegistrar{Producer: producer}
	db := pdb.DBS().Reader

	// Grab the Smartcar integration ID, there should be exactly one.
	var scIntID string
	scInt, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("SmartCar")).One(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve Smartcar integration: %w", err)
	}
	scIntID = scInt.ID

	// Find all integration instances that have acquired Smartcar ids.
	apiInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scIntID),
		models.UserDeviceAPIIntegrationWhere.ExternalID.NEQ(null.StringFromPtr(nil)),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}

	// For each of these send a new registration message, keyed by Smartcar vehicle ID.
	for _, apiInt := range apiInts {
		if err := reg.Register(apiInt.ExternalID.String, apiInt.UserDeviceID, scIntID); err != nil {
			return fmt.Errorf("failed to register Smartcar-DIMO id link for device %s: %w", apiInt.UserDeviceID, err)
		}
	}

	return nil
}
