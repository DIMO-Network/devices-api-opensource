package services

import (
	"encoding/json"
	"fmt"
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

func NewEventService(logger *zerolog.Logger, settings *config.Settings, producer sarama.SyncProducer) EventService {
	return &eventService{
		Settings: settings,
		Logger:   logger,
		Producer: producer,
	}
}

func (e *eventService) Emit(event *Event) error {
	msgBytes, err := json.Marshal(CloudEventMessage{
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

type UserDeviceEventDevice struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	VIN   string `json:"vin"`
}

type UserDeviceEventIntegration struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Style  string `json:"style"`
	Vendor string `json:"vendor"`
}

type UserDeviceIntegrationEvent struct {
	Timestamp   time.Time                  `json:"timestamp"`
	UserID      string                     `json:"userId"`
	Device      UserDeviceEventDevice      `json:"device"`
	Integration UserDeviceEventIntegration `json:"integration"`
}
