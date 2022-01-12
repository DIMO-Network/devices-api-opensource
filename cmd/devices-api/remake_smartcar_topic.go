package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

func remakeSmartcarTopic(logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	ctx := context.Background()
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(strings.Split(settings.KafkaBrokers, ","), kafkaConfig)
	if err != nil {
		return err
	}

	reg := services.SmartcarIngestRegistrar{Producer: producer}
	db := pdb.DBS().Reader

	// Grab the Smartcar integration, there should be exactly one.
	sc, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("SmartCar")).One(ctx, db)
	if err != nil {
		return err
	}
	scID := sc.ID

	// Clear out any old messages that were keyed by our device ID.
	devices, err := models.UserDevices().All(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed retrieving devices: %w", err)
	}
	for _, device := range devices {
		err = reg.Deregister(
			device.ID,
			device.ID, // This looks a bit odd because it is not a Smartcar ID.
			scID,
		)
		if err != nil {
			return fmt.Errorf("failed clearing out any old messages for %s: %w", device.ID, err)
		}
	}

	integs, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(scID), // At the time of writing, this will get everything.
		models.UserDeviceAPIIntegrationWhere.ExternalID.NEQ(null.StringFromPtr(nil)),
	).All(context.Background(), db)
	if err != nil {
		return fmt.Errorf("failed to retrieve all API integrations with external IDs: %w", err)
	}

	// For each of these send a new registration message, keyed by Smartcar vehicle ID.
	for _, integ := range integs {
		err = reg.Register(integ.ExternalID.String, integ.UserDeviceID, scID)
		if err != nil {
			return fmt.Errorf("failed sending registration for device %s: %w", integ.UserDeviceID, err)
		}
	}

	return nil
}
