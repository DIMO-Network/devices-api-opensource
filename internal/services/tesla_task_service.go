package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source tesla_task_service.go -destination mocks/tesla_task_service_mock.go

type TeslaTaskService interface {
	StartPoll(vehicle *TeslaVehicle, udai *models.UserDeviceAPIIntegration) error
	StopPoll(udai *models.UserDeviceAPIIntegration) error
}

func NewTeslaTaskService(settings *config.Settings, producer sarama.SyncProducer) TeslaTaskService {
	return &teslaTaskService{
		Producer: producer,
		Settings: settings,
	}
}

type teslaTaskService struct {
	Producer sarama.SyncProducer
	Settings *config.Settings
}

type TeslaIdentifiers struct {
	ID        int `json:"id"`
	VehicleID int `json:"vehicleId"`
}

type TeslaCredentials struct {
	OwnerAccessToken          string    `json:"ownerAccessToken"`
	OwnerAccessTokenExpiresAt time.Time `json:"ownerAccessTokenExpiresAt"`
	AuthRefreshToken          string    `json:"authRefreshToken"`
}

type TeslaTask struct {
	UserDeviceID   string           `json:"userDeviceId"`
	IntegrationID  string           `json:"integrationId"`
	Identifiers    TeslaIdentifiers `json:"identifiers"`
	Credentials    TeslaCredentials `json:"credentials"`
	ActiveLastPoll bool             `json:"activeLastPoll"`
}

// CloudEventHeaders contains the fields common to all CloudEvent messages.
type CloudEventHeaders struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	SpecVersion string    `json:"specversion"`
	Subject     string    `json:"subject"`
	Time        time.Time `json:"time"`
	Type        string    `json:"type"`
}

type TeslaTaskCloudEvent struct {
	CloudEventHeaders
	Data TeslaTask `json:"data"`
}

func (t *teslaTaskService) StartPoll(vehicle *TeslaVehicle, udai *models.UserDeviceAPIIntegration) error {
	tt := TeslaTaskCloudEvent{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.start",
		},
		Data: TeslaTask{
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: TeslaIdentifiers{
				ID:        vehicle.ID,
				VehicleID: vehicle.VehicleID,
			},
			Credentials: TeslaCredentials{
				OwnerAccessToken:          udai.AccessToken,
				OwnerAccessTokenExpiresAt: udai.AccessExpiresAt,
				AuthRefreshToken:          udai.RefreshToken,
			},
			ActiveLastPoll: false,
		},
	}

	taskKey := fmt.Sprintf("device/%s/integration/%s", udai.UserDeviceID, udai.IntegrationID)

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	_, _, err = t.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: t.Settings.TaskRunNowTopic,
		Key:   sarama.StringEncoder(taskKey),
		Value: sarama.ByteEncoder(ttb),
	})

	return err
}

func (t *teslaTaskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	taskKey := fmt.Sprintf("device/%s/integration/%s", udai.UserDeviceID, udai.IntegrationID)

	tt := struct {
		CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.stop",
		},
		Data: struct {
			LOL string `json:"LOL"`
		}{},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	_, _, err = t.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: t.Settings.TaskStopTopic,
		Key:   sarama.StringEncoder(taskKey),
		Value: sarama.ByteEncoder(ttb),
	})

	return err
}
