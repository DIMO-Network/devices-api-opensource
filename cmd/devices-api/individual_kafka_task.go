package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

func stopTaskByKey(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, taskKey string, producer sarama.SyncProducer) error {
	tt := struct {
		services.CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: services.CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/FAKE",
			SpecVersion: "1.0",
			Subject:     "FAKE",
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.stop",
		},
		Data: struct {
			TaskID        string `json:"taskId"`
			UserDeviceID  string `json:"userDeviceId"`
			IntegrationID string `json:"integrationId"`
		}{
			TaskID:        taskKey,
			UserDeviceID:  "FAKE",
			IntegrationID: "FAKE",
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	_, _, err = producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: settings.TaskStopTopic,
			Key:   sarama.StringEncoder(taskKey),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return err
}
