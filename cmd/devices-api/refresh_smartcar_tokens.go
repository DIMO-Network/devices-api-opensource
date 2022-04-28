package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func refreshSmartcarTokens(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	db := pdb.DBS().Writer
	logger.Info().Msg("Refreshing Smartcar tokens")

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
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNull(),
	).All(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs and status Active: %w", err)
	}

	client := smartcar.NewClient()
	auth := client.NewAuth(&smartcar.AuthParams{
		ClientID:     settings.SmartcarClientID,
		ClientSecret: settings.SmartcarClientSecret,
	})

	// For each of these, try the exchange and save the result if successul.
	for _, apiInt := range apiInts {
		token, err := auth.ExchangeRefreshToken(context.Background(), &smartcar.ExchangeRefreshTokenParams{
			Token: apiInt.RefreshToken.String,
		})
		if err != nil {
			logger.Err(err).Str("externalID", apiInt.ExternalID.String).Msg("Failed refreshing Smartcar token")
			continue
		}

		apiInt.AccessToken = null.StringFrom(token.Access)
		apiInt.AccessExpiresAt = null.TimeFrom(token.AccessExpiry)
		apiInt.RefreshToken = null.StringFrom(token.Refresh)

		logger.Info().Str("userDeviceId", apiInt.UserDeviceID).Str("refreshToken", token.Refresh).Msg("Refresh succeeded")

		_, err = apiInt.Update(ctx, db, boil.Infer())
		if err != nil {
			logger.Err(err).Str("externalID", apiInt.ExternalID.String).Msgf("Failed saving new Smartcar token to database")
		}
	}

	logger.Info().Msg("Refreshing Smartcar tokens")

	return nil
}
