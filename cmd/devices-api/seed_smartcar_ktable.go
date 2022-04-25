package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

func seedSmartcarCreds(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, producer sarama.SyncProducer) error {
	scInteg, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("SmartCar")).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	activeInts, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scInteg.ID),
		models.UserDeviceAPIIntegrationWhere.Status.EQ(models.UserDeviceAPIIntegrationStatusActive),
		models.UserDeviceAPIIntegrationWhere.TaskID.IsNotNull(),
	).All(ctx, pdb.DBS().Reader.DB)
	if err != nil {
		return err
	}

	success := 0

	for _, integ := range activeInts {
		logger := logger.With().Str("userDeviceId", integ.UserDeviceID).Logger()

		tc := services.TeslaCredentialsCloudEventV2{
			CloudEventHeaders: services.CloudEventHeaders{
				ID:          ksuid.New().String(),
				Source:      "dimo/integration/" + integ.IntegrationID,
				SpecVersion: "1.0",
				Subject:     integ.UserDeviceID,
				Time:        time.Now(),
				Type:        "zone.dimo.task.smartcar.poll.credential",
			},
			Data: services.TeslaCredentialsV2{
				TaskID:        integ.TaskID.String,
				UserDeviceID:  integ.UserDeviceID,
				IntegrationID: integ.IntegrationID,
				AccessToken:   integ.AccessToken.String,
				Expiry:        integ.AccessExpiresAt.Time,
				RefreshToken:  integ.RefreshToken.String,
			},
		}

		tcb, err := json.Marshal(tc)
		if err != nil {
			return err
		}

		_, _, err = producer.SendMessage(
			&sarama.ProducerMessage{
				Topic: settings.TaskCredentialTopic,
				Key:   sarama.StringEncoder(integ.TaskID.String),
				Value: sarama.ByteEncoder(tcb),
			},
		)
		if err != nil {
			logger.Err(err).Msg("Failed to update table.")
			continue
		}

		success++
	}

	logger.Info().Msgf("Seeded %d/%d Smartcar jobs.", success, len(activeInts))

	return nil
}
