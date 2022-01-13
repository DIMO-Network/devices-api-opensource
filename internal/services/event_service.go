package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

type Event struct {
	Type    string
	Subject string
	Source  string
	Data    interface{}
}

type EventService interface {
	Emit(event *Event) error
}

type eventService struct {
	Settings *config.Settings
	Logger   *zerolog.Logger
	Producer sarama.SyncProducer
}

func NewEventService(logger *zerolog.Logger, settings *config.Settings) EventService {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	// Would like to move to AsyncProducer but this needs more thought. These are not mere tracing
	// messages, the user does expect to see them.
	producer, err := sarama.NewSyncProducer(strings.Split(settings.KafkaBrokers, ","), kafkaConfig)
	if err != nil {
		panic(err)
	}
	return &eventService{
		Settings: settings,
		Logger:   logger,
		Producer: producer,
	}
}

func (e *eventService) Emit(event *Event) error {
	msgBytes, err := json.Marshal(cloudEventMessage{
		ID:          ksuid.New().String(),
		Source:      event.Source,
		SpecVersion: "1.0",
		Subject:     event.Subject,
		Time:        time.Now(),
		Type:        event.Type,
		Data:        event.Data,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal CloudEvent: %w", err)
	}
	msg := &sarama.ProducerMessage{
		Topic: e.Settings.EventsTopic,
		Value: sarama.ByteEncoder(msgBytes),
	}
	_, _, err = e.Producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to produce CloudEvent to Kafka: %w", err)
	}
	return nil
}
