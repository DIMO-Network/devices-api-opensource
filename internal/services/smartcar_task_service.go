package services

import (
	"encoding/json"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"
)

//go:generate mockgen -source smartcar_task_service.go -destination mocks/smartcar_task_service_mock.go

type SmartcarTaskService interface {
	StartPoll(udai *models.UserDeviceAPIIntegration) error
	StopPoll(udai *models.UserDeviceAPIIntegration) error
	Refresh(udai *models.UserDeviceAPIIntegration) error
}

func NewSmartcarTaskService(settings *config.Settings, producer sarama.SyncProducer) SmartcarTaskService {
	return &smartcarTaskService{
		Producer: producer,
		Settings: settings,
	}
}

type smartcarTaskService struct {
	Producer sarama.SyncProducer
	Settings *config.Settings
}

type SmartcarIdentifiers struct {
	ID string `json:"id"`
}

type SmartcarCredentials struct {
	TaskID        string    `json:"taskId"`
	UserDeviceID  string    `json:"userDeviceId"`
	IntegrationID string    `json:"integrationId"`
	AccessToken   string    `json:"accessToken"`
	Expiry        time.Time `json:"expiry"`
	RefreshToken  string    `json:"refreshToken"`
}

type SmartcarTask struct {
	TaskID        string              `json:"taskId"`
	UserDeviceID  string              `json:"userDeviceId"`
	IntegrationID string              `json:"integrationId"`
	Identifiers   SmartcarIdentifiers `json:"identifiers"`
	Paths         []string            `json:"paths"`
}

type SmartcarTaskCloudEvent struct {
	CloudEventHeaders
	Data SmartcarTask `json:"data"`
}

type SmartcarCredentialsCloudEvent struct {
	CloudEventHeaders
	Data SmartcarCredentials `json:"data"`
}

func (t *smartcarTaskService) StartPoll(udai *models.UserDeviceAPIIntegration) error {
	m := new(UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(m); err != nil {
		return err
	}

	tt := SmartcarTaskCloudEvent{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.smartcar.poll.scheduled",
		},
		Data: SmartcarTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: SmartcarIdentifiers{
				ID: udai.ExternalID.String,
			},
			Paths: m.SmartcarEndpoints,
		},
	}

	tc := TeslaCredentialsCloudEventV2{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.smartcar.poll.credential",
		},
		Data: TeslaCredentialsV2{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			AccessToken:   udai.AccessToken.String,
			Expiry:        udai.AccessExpiresAt.Time,
			RefreshToken:  udai.RefreshToken.String,
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	tcb, err := json.Marshal(tc)
	if err != nil {
		return err
	}

	err = t.Producer.SendMessages(
		[]*sarama.ProducerMessage{
			{
				Topic: t.Settings.TaskRunNowTopic,
				Key:   sarama.StringEncoder(udai.TaskID.String),
				Value: sarama.ByteEncoder(ttb),
			},
			{
				Topic: t.Settings.TaskCredentialTopic,
				Key:   sarama.StringEncoder(udai.TaskID.String),
				Value: sarama.ByteEncoder(tcb),
			},
		},
	)

	return err
}

func (t *smartcarTaskService) Refresh(udai *models.UserDeviceAPIIntegration) error {
	m := new(UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(m); err != nil {
		return err
	}

	tt := SmartcarTaskCloudEvent{
		CloudEventHeaders: CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.smartcar.poll.refresh",
		},
		Data: SmartcarTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: SmartcarIdentifiers{
				ID: udai.ExternalID.String,
			},
			Paths: m.SmartcarEndpoints,
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return err
}

func (t *smartcarTaskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	var taskKey = udai.TaskID.String

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
			Type:        "zone.dimo.task.smartcar.poll.stop",
		},
		Data: struct {
			TaskID        string `json:"taskId"`
			UserDeviceID  string `json:"userDeviceId"`
			IntegrationID string `json:"integrationId"`
		}{
			TaskID:        taskKey,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	err = t.Producer.SendMessages(
		[]*sarama.ProducerMessage{
			{
				Topic: t.Settings.TaskStopTopic,
				Key:   sarama.StringEncoder(taskKey),
				Value: sarama.ByteEncoder(ttb),
			},
			{
				Topic: t.Settings.TaskCredentialTopic,
				Key:   sarama.StringEncoder(taskKey),
				Value: nil,
			},
		},
	)

	return err
}
