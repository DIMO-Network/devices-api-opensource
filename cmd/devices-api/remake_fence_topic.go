package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/controllers"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func remakeFenceTopic(logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, producer sarama.SyncProducer) error {
	ctx := context.Background()

	rels, err := models.UserDeviceToGeofences(
		qm.Load(models.UserDeviceToGeofenceRels.Geofence),
	).All(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	deviceIDToIndexes := make(map[string]*shared.StringSet)

	for _, rel := range rels {
		if _, ok := deviceIDToIndexes[rel.UserDeviceID]; !ok {
			deviceIDToIndexes[rel.UserDeviceID] = shared.NewStringSet()
		}
		for _, ind := range rel.R.Geofence.H3Indexes {
			deviceIDToIndexes[rel.UserDeviceID].Add(ind)
		}
	}

	for userDeviceID, indexes := range deviceIDToIndexes {
		if indexes.Len() == 0 {
			continue
		}
		ce := services.CloudEventMessage{
			ID:          ksuid.New().String(),
			Source:      "devices-api",
			SpecVersion: "1.0",
			Subject:     userDeviceID,
			Time:        time.Now(),
			Type:        controllers.PrivacyFenceEventType,
			Data: controllers.FenceData{
				H3Indexes: indexes.Slice(),
			},
		}
		b, err := json.Marshal(ce)
		if err != nil {
			return err
		}
		msg := &sarama.ProducerMessage{
			Topic: settings.PrivacyFenceTopic,
			Key:   sarama.StringEncoder(userDeviceID),
			Value: sarama.ByteEncoder(b),
		}
		if _, _, err := producer.SendMessage(msg); err != nil {
			return err
		}
	}

	return nil
}
