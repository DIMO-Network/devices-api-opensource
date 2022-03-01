package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

func migrateSmartcarWebhooks(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, oldWebhookID string) error {
	db := pdb.DBS().Reader
	logger.Info().Msgf("Migrating Smartcar webhooks from %q to %q", oldWebhookID, settings.SmartcarWebhookID)

	httpClient := &http.Client{Timeout: 10 * time.Second}

	oldClient := services.SmartcarWebhookClient{
		HTTPClient: httpClient,
		WebhookID:  oldWebhookID,
	}

	newClient := services.SmartcarWebhookClient{
		HTTPClient: httpClient,
		WebhookID:  settings.SmartcarWebhookID,
	}

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
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs and status Active: %w", err)
	}

	success := 0
	fail := 0

	// For each of these send a new registration message, keyed by Smartcar vehicle ID.
	for _, apiInt := range apiInts {
		good := true
		vehicleID := apiInt.ExternalID.String
		if err := oldClient.Unsubscribe(vehicleID, apiInt.AccessToken); err != nil {
			logger.Err(err).Msgf("Failed to unsubscribe %s from the old webhook", vehicleID)
		} else {
			logger.Info().Msgf("Successfully unsubscribed %s from old webhook", vehicleID)
		}
		if err := newClient.Subscribe(vehicleID, apiInt.AccessToken); err != nil {
			logger.Err(err).Msgf("Failed to subscribe %s to the new webhook", vehicleID)
			good = false
		} else {
			logger.Info().Msgf("Successfully subscribed %s to new webhook", vehicleID)
		}

		if good {
			success++
		} else {
			fail++
		}

		logger.Info().Msgf("Succeed: %d, Fail: %d", success, fail)
	}

	return nil
}
