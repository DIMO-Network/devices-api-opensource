package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// The situation here is that a refresh token has been inserted, in the plain, into the refresh_token column. The
// job is still running but failing.
//
// * Exchange the token
// * Encrypt and save the new tokens
// * Update the credentials in the job.
func redoSmartcarToken(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, cipher shared.Cipher, producer sarama.SyncProducer, userDeviceID string, smartcarClient services.SmartcarClient) error {
	db := pdb.DBS().Writer

	// Grab the Smartcar integration ID, there should be exactly one.
	var scIntID string
	scInt, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("SmartCar")).One(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to retrieve Smartcar integration: %w", err)
	}
	scIntID = scInt.ID

	// Find all integration instances that have acquired Smartcar ids.
	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scIntID),
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
	).One(ctx, db)
	if err != nil {
		return fmt.Errorf("couldn't find Smartcar integration: %w", err)
	}

	client := smartcar.NewClient()
	auth := client.NewAuth(&smartcar.AuthParams{
		ClientID:     settings.SmartcarClientID,
		ClientSecret: settings.SmartcarClientSecret,
	})

	newToken, err := auth.ExchangeRefreshToken(ctx, &smartcar.ExchangeRefreshTokenParams{Token: udai.RefreshToken.String})
	if err != nil {
		return err
	}
	logger.Info().Str("userDeviceID", userDeviceID).Str("newRefreshToken", newToken.Refresh).Msg("Exchanged token.")

	externalID, err := smartcarClient.GetExternalID(ctx, newToken.Access)
	if err != nil {
		return err
	}

	if externalID != udai.ExternalID.String {
		return fmt.Errorf("expected external ID %s but token has %s", udai.ExternalID.String, externalID)
	}

	encAccess, err := cipher.Encrypt(newToken.Access)
	if err != nil {
		return err
	}

	encRefresh, err := cipher.Encrypt(newToken.Refresh)
	if err != nil {
		return err
	}

	udai.AccessToken = null.StringFrom(encAccess)
	udai.RefreshToken = null.StringFrom(encRefresh)
	udai.AccessExpiresAt = null.TimeFrom(newToken.AccessExpiry)
	if _, err := udai.Update(ctx, db, boil.Infer()); err != nil {
		return err
	}

	tc := services.TeslaCredentialsCloudEventV2{
		CloudEventHeaders: services.CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + scIntID,
			SpecVersion: "1.0",
			Subject:     userDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.smartcar.poll.credential",
		},
		Data: services.TeslaCredentialsV2{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			AccessToken:   encAccess,
			Expiry:        newToken.AccessExpiry,
			RefreshToken:  encRefresh,
		},
	}

	tcb, err := json.Marshal(tc)
	if err != nil {
		return err
	}

	_, _, err = producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: settings.TaskCredentialTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(tcb),
		},
	)
	if err != nil {
		return err
	}

	logger.Info().Msgf("Refreshed Smartcar token for %s", userDeviceID)

	return nil
}
