package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/segmentio/ksuid"
)

type TaskService interface {
	StartPoll(udai *models.UserDeviceAPIIntegration, vendor string) error
	StopPoll(udai *models.UserDeviceAPIIntegration) error
	UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error)
	LockDoors(udai *models.UserDeviceAPIIntegration) (string, error)
	OpenTrunk(udai *models.UserDeviceAPIIntegration) (string, error)
	OpenFrunk(udai *models.UserDeviceAPIIntegration) (string, error)
	Refresh(udai *models.UserDeviceAPIIntegration) error
}

type taskService struct {
	Producer sarama.SyncProducer
	Settings *config.Settings
}

type Identifiers struct {
	ID             string `json:"id,omitempty"`
	TeslaVehicleID int    `json:"vehicleId,omitempty"`
}

type DoorTask struct {
	TaskID        string      `json:"taskId"`
	SubTaskID     string      `json:"subTaskId"`
	UserDeviceID  string      `json:"userDeviceId"`
	IntegrationID string      `json:"integrationId"`
	Identifiers   Identifiers `json:"identifiers"`
	ChargeLimit   *float64    `json:"chargeLimit,omitempty"`
}

func NewTaskServiceController(settings *config.Settings, producer sarama.SyncProducer) TaskService {
	return &taskService{
		Producer: producer,
		Settings: settings,
	}
}

func (t *taskService) StartPoll(udai *models.UserDeviceAPIIntegration, vendor string) error {
	m := new(services.UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(m); err != nil {
		return err
	}

	var taskData, credentialData interface{}

	switch vendor {
	case constants.SmartCarVendor:
		taskData = services.SmartcarTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: services.SmartcarIdentifiers{
				ID: udai.ExternalID.String,
			},
			Paths: m.SmartcarEndpoints,
		}
		credentialData = services.TeslaCredentialsV2{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			AccessToken:   udai.AccessToken.String,
			Expiry:        udai.AccessExpiresAt.Time,
			RefreshToken:  udai.RefreshToken.String,
		}
	case constants.TeslaVendor:
		extID, err := strconv.Atoi(udai.ExternalID.String)
		if err != nil {
			return err
		}

		var metadata services.UserDeviceAPIIntegrationsMetadata
		err = json.Unmarshal(udai.Metadata.JSON, &metadata)
		if err != nil {
			return err
		}

		taskData = services.TeslaTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: services.TeslaIdentifiers{
				ID:        extID,
				VehicleID: metadata.TeslaVehicleID,
			},
			OnlineIdleLastPoll: false,
		}

		credentialData = services.TeslaCredentialsV2{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			AccessToken:   udai.AccessToken.String,
			Expiry:        udai.AccessExpiresAt.Time,
			RefreshToken:  udai.RefreshToken.String,
		}
	}

	tt := struct {
		shared.CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: shared.CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.smartcar.poll.scheduled",
		},
		Data: taskData,
	}

	tc := struct {
		shared.CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: shared.CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.smartcar.poll.credential",
		},
		Data: credentialData,
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

func (t *taskService) Refresh(udai *models.UserDeviceAPIIntegration) error {
	m := new(services.UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(m); err != nil {
		return err
	}

	tt := struct {
		shared.CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: shared.CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/" + udai.IntegrationID,
			SpecVersion: "1.0",
			Subject:     udai.UserDeviceID,
			Time:        time.Now(),
			Type:        "zone.dimo.task.smartcar.poll.refresh",
		},
		Data: services.SmartcarTask{
			TaskID:        udai.TaskID.String,
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: services.SmartcarIdentifiers{
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

func (t *taskService) StopPoll(udai *models.UserDeviceAPIIntegration) error {
	var taskKey string
	if udai.TaskID.Valid {
		taskKey = udai.TaskID.String
	} else {
		taskKey = fmt.Sprintf("device/%s/integration/%s", udai.UserDeviceID, udai.IntegrationID)
	}

	tt := struct {
		shared.CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: shared.CloudEventHeaders{
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

func (t *taskService) UnlockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	tt := shared.CloudEvent[DoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.doors.unlock",
		Data: DoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: Identifiers{
				ID: udai.ExternalID.String,
			},
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return "", err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return tt.Data.SubTaskID, err
}

func (t *taskService) LockDoors(udai *models.UserDeviceAPIIntegration) (string, error) {
	tt := shared.CloudEvent[DoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.smartcar.doors.lock",
		Data: DoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: Identifiers{
				ID: udai.ExternalID.String,
			},
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return "", err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return tt.Data.SubTaskID, err
}

func (t *taskService) OpenTrunk(udai *models.UserDeviceAPIIntegration) (string, error) {
	tt := shared.CloudEvent[DoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.trunk.open",
		Data: DoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: Identifiers{
				ID: udai.ExternalID.String,
			},
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return "", err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return tt.Data.SubTaskID, err
}

func (t *taskService) OpenFrunk(udai *models.UserDeviceAPIIntegration) (string, error) {
	tt := shared.CloudEvent[DoorTask]{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + udai.IntegrationID,
		SpecVersion: "1.0",
		Subject:     udai.UserDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.task.tesla.frunk.open",
		Data: DoorTask{
			TaskID:        udai.TaskID.String,
			SubTaskID:     ksuid.New().String(),
			UserDeviceID:  udai.UserDeviceID,
			IntegrationID: udai.IntegrationID,
			Identifiers: Identifiers{
				ID: udai.ExternalID.String,
			},
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return "", err
	}

	_, _, err = t.Producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: t.Settings.TaskRunNowTopic,
			Key:   sarama.StringEncoder(udai.TaskID.String),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return tt.Data.SubTaskID, err
}
