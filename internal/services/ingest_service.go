package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
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
		err := i.processMessage(msg)
		if err != nil {
			i.log.Err(err).Msg("error processing smartcar ingest msg")
		}
	}
}

func (i *IngestService) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

	defer appmetrics.SmartcarIngestTotalOps.Inc()

	log.Info().Msgf("Received message: %s, payload: %s", msg.UUID, string(msg.Payload))
	e := new(DeviceStatusEvent)

	err := json.Unmarshal(msg.Payload, e)
	if err != nil {
		return errors.Wrap(err, "error parsing device event payload")
	}

	if e.Type != deviceStatusEventType {
		return fmt.Errorf("received vehicle status event with unexpected type %s", e.Type)
	}

	return i.processEvent(e)
}

var integrationIDregex = regexp.MustCompile("^dimo/integration/([a-zA-Z0-9]{27})$")

func (i *IngestService) processEvent(event *DeviceStatusEvent) error {
	ctx := context.Background() // should this be passed in so can cancel if application shutting down?

	userDeviceID := event.Subject
	udd := models.UserDeviceDatum{
		UserDeviceID: userDeviceID,
		Data:         null.JSONFrom(event.Data),
	}
	whiteList := []string{"error_data", "updated_at"} // always update error_data, even if no errors so gets set to null
	newOdometer, errOdo := extractOdometer(event.Data)
	if errOdo != nil {
		i.log.Err(errOdo).Msg("Failed to grab odometer from status update, will not update Data")
		udd.ErrorData = null.JSONFrom(event.Data)
	} else {
		whiteList = append(whiteList, "data") // set data only if can find odometer, meaning good data
	}

	tx, err := i.db().Writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	match := integrationIDregex.FindStringSubmatch(event.Source)
	if match == nil {
		i.log.Error().Msgf("Failed to parse out integration from device status event source %q", event.Source)
	} else {
		integrationID := match[1]
		udai, err := models.FindUserDeviceAPIIntegration(ctx, tx, userDeviceID, integrationID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				i.log.Err(err).Msgf("No API integration found for device %s and integration %s", userDeviceID, integrationID)
			} else {
				i.log.Err(err).Msg("Failed to search for device integration, cannot check status")
			}
		} else if udai.Status != models.UserDeviceAPIIntegrationStatusActive {
			udai.Status = models.UserDeviceAPIIntegrationStatusActive
			if _, err := udai.Update(ctx, tx, boil.Whitelist("status")); err != nil {
				i.log.Err(err).Msg("Failed to update user device API integration to active")
			}
		}
	}

	// Horribly inefficient
	device, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.DeviceDefinition), // Only needed for the odometer event.
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
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

	err = udd.Upsert(ctx, tx, true, []string{"user_device_id"}, boil.Whitelist(whiteList...), boil.Infer())
	if err != nil {
		return fmt.Errorf("error upserting vehicle status event data: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	if errOdo == nil && (!haveOldOdometer || newOdometer != oldOdometer) {
		// If the Smartcar /odometer endpoint returned an error, we won't have a value.
		err = i.eventService.Emit(&Event{
			Type:    "com.dimo.zone.device.odometer.update",
			Subject: userDeviceID,
			Source:  event.Source, // Should be the integration
			Data: OdometerEvent{
				Timestamp: time.Now(),
				UserID:    device.UserID,
				Device: odometerEventDevice{
					ID:    userDeviceID,
					Make:  device.R.DeviceDefinition.R.DeviceMake.Name,
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
