package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TaskStatusListener struct {
	db  func() *database.DBReaderWriter
	log *zerolog.Logger
	cio CIOClient
}

type CIOClient interface {
	Track(customerID string, eventName string, data map[string]interface{}) error
}

type TaskStatusCloudEvent struct {
	CloudEventHeaders
	Data TaskStatusData `json:"data"`
}

type TaskStatusData struct {
	TaskID        string `json:"taskId"`
	UserDeviceID  string `json:"userDeviceId"`
	IntegrationID string `json:"integrationId"`
	Status        string `json:"status"`
}

func NewTaskStatusListener(db func() *database.DBReaderWriter, log *zerolog.Logger, cio CIOClient) *TaskStatusListener {
	return &TaskStatusListener{db: db, log: log, cio: cio}
}

func (i *TaskStatusListener) ProcessTaskUpdates(messages <-chan *message.Message) {
	for msg := range messages {
		err := i.processMessage(msg)
		if err != nil {
			i.log.Err(err).Msg("error processing task status message")
		}
	}
}

const smartcarStatusEventType = "zone.dimo.task.smartcar.poll.status.update"

func (i *TaskStatusListener) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	event := new(TaskStatusCloudEvent)
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing task status payload")
	}

	return i.processEvent(event)
}

func (i *TaskStatusListener) processEvent(event *TaskStatusCloudEvent) error {
	var (
		ctx          = context.Background()
		userDeviceID = event.Subject
	)

	// Smartcar-only for now.
	if event.Type != smartcarStatusEventType {
		return fmt.Errorf("unexpected event type %s", event.Type)
	}

	// Should we use data.integrationId instead?
	if !strings.HasPrefix(event.Source, sourcePrefix) {
		return fmt.Errorf("unexpected event source format: %s", event.Source)
	}
	integrationID := strings.TrimPrefix(event.Source, sourcePrefix)

	// Just one case for now.
	if event.Data.Status != models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
		return fmt.Errorf("unexpected task status %s", event.Data.Status)
	}

	where := models.UserDeviceAPIIntegrationWhere
	integ, err := models.UserDeviceAPIIntegrations(
		where.UserDeviceID.EQ(userDeviceID),
		where.IntegrationID.EQ(integrationID),
		qm.Load(
			qm.Rels(
				models.UserDeviceAPIIntegrationRels.UserDevice,
				models.UserDeviceRels.DeviceDefinition,
				models.DeviceDefinitionRels.DeviceMake,
			),
		), // Only need this for CustomerIO.
	).One(ctx, i.db().Writer)
	if err != nil {
		return fmt.Errorf("couldn't find device integration for device %s and integration %s: %w", userDeviceID, integrationID, err)
	}

	userDevice := integ.R.UserDevice

	i.log.Info().Str("userDeviceId", userDeviceID).Msg("Setting Smartcar integration to failed because credentials have changed.")

	if integ.TaskID.Valid && integ.TaskID.String == event.Data.TaskID {
		// Maybe you've restarted the task with new credentials already.
		// TODO: Delete credentials entry?
		integ.TaskID = null.String{}
	}
	integ.Status = models.UserDeviceAPIIntegrationStatusAuthenticationFailure
	if _, err := integ.Update(context.Background(), i.db().Writer, boil.Infer()); err != nil {
		return err
	}

	dd := userDevice.R.DeviceDefinition
	data := map[string]interface{}{
		"deviceId":     userDeviceID,
		"make_name":    dd.R.DeviceMake.Name,
		"model_name":   dd.Model,
		"model_year":   dd.Year,
		"country_code": userDevice.CountryCode.String,
	}

	if err := i.cio.Track(userDevice.UserID, "smartcar.Reauth.Required", data); err != nil {
		i.log.Err(err).Str("userId", userDevice.UserID).Str("userDeviceId", userDeviceID).Msg("Failed to emit reauthentication Customer.io event.")
	}

	return nil
}
