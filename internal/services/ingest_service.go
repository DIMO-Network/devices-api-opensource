package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const deviceStatusEventType = "zone.dimo.device.status.update"

type IngestService struct {
	db           func() *database.DBReaderWriter
	log          *zerolog.Logger
	eventService EventService
}

func NewIngestService(db func() *database.DBReaderWriter, log *zerolog.Logger, eventService EventService) *IngestService {
	return &IngestService{db: db, log: log, eventService: eventService}
}

// ProcessDeviceStatusMessages works on channel stream of messages from watermill kafka consumer
func (i *IngestService) ProcessDeviceStatusMessages(messages <-chan *message.Message) {
	for msg := range messages {
		err := i.processDeviceStatus(msg)
		if err != nil {
			i.log.Err(err).Msg("error processing smartcar ingest msg")
		}
	}
}

func (i *IngestService) processDeviceStatus(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	defer appmetrics.SmartcarIngestTotalOps.Inc()

	ctx := context.Background() // should this be passed in so can cancel if application shutting down?
	log.Info().Msgf("Received message: %s, payload: %s", msg.UUID, string(msg.Payload))
	e := DeviceStatusEvent{}

	err := json.Unmarshal(msg.Payload, &e)
	if err != nil {
		return errors.Wrap(err, "error parsing device event payload")
	}

	if e.Type != deviceStatusEventType {
		return fmt.Errorf("received vehicle status event with unexpected type %s", e.Type)
	}

	userDeviceID := e.Subject
	udd := models.UserDeviceDatum{
		UserDeviceID: userDeviceID,
		Data:         null.JSONFrom(e.Data),
	}

	tx, err := i.db().Writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	// Horribly inefficient
	device, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.DeviceDefinition), // Only needed for the odometer event.
	).One(ctx, tx)
	if err != nil {
		return fmt.Errorf("couldn't find device %s for status update: %w", userDeviceID, err)
	}

	haveOldOdometer := false
	var oldOdometer float64
	oldUDD, err := models.FindUserDeviceDatum(ctx, tx, userDeviceID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			i.log.Err(err).Msg("Failed to look up old odometer value.")
		}
	} else if oldUDD.Data.Valid {
		oldOdometer, err = extractOdometer(oldUDD.Data.JSON)
		if err != nil {
			i.log.Err(err).Msg("Failed to grab odometer from existing status update")
		} else {
			haveOldOdometer = true
		}
	}

	err = udd.Upsert(ctx, tx, true, []string{"user_device_id"}, boil.Whitelist("data", "created_at", "updated_at"), boil.Infer())
	if err != nil {
		return fmt.Errorf("error upserting vehicle status event data: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	newOdometer, err := extractOdometer(e.Data)
	if err != nil {
		i.log.Err(err).Msg("Failed to grab odometer from status update")
	} else if !haveOldOdometer || newOdometer != oldOdometer {
		// If the Smartcar /odometer endpoint returned an error, we won't have a value.
		err = i.eventService.Emit(&Event{
			Type:    "com.dimo.zone.device.odometer.update",
			Subject: userDeviceID,
			Source:  e.Source, // Should be the integration
			Data: OdometerEvent{
				Timestamp: time.Now(),
				UserID:    device.UserID,
				Device: odometerEventDevice{
					ID:    userDeviceID,
					Make:  device.R.DeviceDefinition.Make,
					Model: device.R.DeviceDefinition.Model,
					Year:  int(device.R.DeviceDefinition.Year),
				},
				Odometer: newOdometer,
			},
		})
		if err != nil {
			i.log.Err(err).Msg("Failed to emit odometer event")
		}
	}

	appmetrics.SmartcarIngestSuccessOps.Inc()
	return nil
}

func extractOdometer(data []byte) (float64, error) {
	var partialData struct {
		Odometer *float64 `json:"odometer"`
	}
	err := json.Unmarshal(data, &partialData)
	if err != nil {
		return 0, fmt.Errorf("failed to parse data payload")
	}
	if partialData.Odometer == nil {
		return 0, fmt.Errorf("data payload did not have an odometer reading")
	}

	return *partialData.Odometer, nil
}

type odometerEventDevice struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

type OdometerEvent struct {
	Timestamp time.Time           `json:"timestamp"`
	UserID    string              `json:"userId"`
	Device    odometerEventDevice `json:"device"`
	Odometer  float64             `json:"odometer"`
}

type DeviceStatusEvent struct {
	ID          string          `json:"id"`
	Source      string          `json:"source"`
	Specversion string          `json:"specversion"`
	Subject     string          `json:"subject"`
	Time        time.Time       `json:"time"`
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}
