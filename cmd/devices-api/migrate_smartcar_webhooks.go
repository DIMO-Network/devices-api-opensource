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
	logger.Info().Msgf("Migrating Smartcar webhooks from %s to %s", oldWebhookID, settings.SmartcarWebhookID)

	httpClient := &http.Client{Timeout: 10 * time.Second}

	oldClient := services.SmartcarWebhookClient{
		HTTPClient:      httpClient,
		WebhookID:       oldWebhookID,
		ManagementToken: settings.SmartcarManagementToken,
	}

	newClient := services.SmartcarWebhookClient{
		HTTPClient:      httpClient,
		WebhookID:       settings.SmartcarWebhookID,
		ManagementToken: settings.SmartcarManagementToken, // Won't use it here.
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

	// For each of these send a new registration message, keyed by Smartcar vehicle ID.
	for _, apiInt := range apiInts {
		dimoID := apiInt.UserDeviceID
		vehicleID := apiInt.ExternalID.String
		accessToken := apiInt.AccessToken.String

		if err := oldClient.Unsubscribe(vehicleID); err != nil {
			logger.Err(err).Msgf("Failed to unsubscribe %s from the old webhook", dimoID)
		}
		if err := newClient.Subscribe(vehicleID, accessToken); err != nil {
			logger.Err(err).Msgf("Failed to subscribe %s to the new webhook", dimoID)
		}
	}

	return nil
}
