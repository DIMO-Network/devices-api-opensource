package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
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

var odometerCooldown = time.Hour

type IngestService struct {
	db           func() *database.DBReaderWriter
	log          *zerolog.Logger
	eventService EventService
	integrations models.IntegrationSlice
}

func NewIngestService(db func() *database.DBReaderWriter, log *zerolog.Logger, eventService EventService) *IngestService {
	// save integrations available
	integrations, _ := models.Integrations().All(context.Background(), db().Reader)
	return &IngestService{db: db, log: log, eventService: eventService, integrations: integrations}
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

	log.Info().Msgf("Received message: %s, payload: %s", msg.UUID, string(msg.Payload))

	event := new(DeviceStatusEvent)
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		return errors.Wrap(err, "error parsing device event payload")
	}

	if event.Type != deviceStatusEventType {
		return fmt.Errorf("received vehicle status event with unexpected type %s", event.Type)
	}
	integration, err := i.getIntegrationFromEvent(event)
	if err != nil {
		return err
	}
	if integration.Vendor == SmartCarVendor {
		defer appmetrics.SmartcarIngestTotalOps.Inc()
	}
	if integration.Vendor == AutoPiVendor {
		defer appmetrics.AutoPiIngestTotalOps.Inc()
	}

	return i.processEvent(event)
}

// integrationIDregexp is used to parse out the KSUID of the integration from the CloudEvent
// source field.
var integrationIDregexp = regexp.MustCompile("^dimo/integration/([a-zA-Z0-9]{27})$")

func (i *IngestService) processEvent(event *DeviceStatusEvent) error {
	ctx := context.Background() // should this be passed in so can cancel if application shutting down?
	userDeviceID := event.Subject

	match := integrationIDregexp.FindStringSubmatch(event.Source)
	if match == nil {
		return fmt.Errorf("failed to parse integration from event source %q", event.Source)
	}
	integrationID := match[1]
	integration, err := i.getIntegrationFromEvent(event)
	if err != nil {
		return err
	}

	tx, err := i.db().Writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	device, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load(
			models.UserDeviceRels.UserDeviceAPIIntegrations,
			models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integration.ID),
		),
		qm.Load(models.UserDeviceRels.UserDeviceDatum),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).One(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to find device: %w", err)
	}

	if len(device.R.UserDeviceAPIIntegrations) == 0 {
		return fmt.Errorf("can't find API integration for device %s and integration %s", userDeviceID, integration.ID)
	}

	// update status to Active if not alrady set
	apiIntegration := device.R.UserDeviceAPIIntegrations[0]
	if apiIntegration.Status != models.UserDeviceAPIIntegrationStatusActive {
		apiIntegration.Status = models.UserDeviceAPIIntegrationStatusActive
		if _, err := apiIntegration.Update(ctx, tx, boil.Infer()); err != nil {
			return fmt.Errorf("failed to update API integration: %w", err)
		}
	}

	// currently this will be 0 with autopi
	var newOdometer null.Float64
	if o, err := extractOdometer(event.Data); err == nil {
		newOdometer = null.Float64From(o)
	}

	datum := device.R.UserDeviceDatum
	if datum == nil {
		datum = &models.UserDeviceDatum{
			UserDeviceID: userDeviceID,
		}
	}
	if datum.IntegrationID.IsZero() {
		datum.IntegrationID = null.StringFrom(integration.ID) // once we have data make this field not nullable
	}
	// save autopi data
	if integration.Vendor == AutoPiVendor {
		datum.Data = null.JSONFrom(event.Data)
		datum.ErrorData = null.JSON{}
		i.processOdometer(datum, newOdometer, device, integrationID)
	}
	// do smartcar
	if integration.Vendor == SmartCarVendor || integration.Vendor == TeslaVendor {
		if newOdometer.Valid {
			datum.Data = null.JSONFrom(event.Data)
			datum.ErrorData = null.JSON{}

			i.processOdometer(datum, newOdometer, device, integrationID)
		} else {
			datum.ErrorData = null.JSONFrom(event.Data)
		}
	}

	if err := datum.Upsert(ctx, tx, true, []string{models.UserDeviceDatumColumns.UserDeviceID}, boil.Infer(), boil.Infer()); err != nil {
		return fmt.Errorf("error upserting datum: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	if integration.Vendor == SmartCarVendor {
		appmetrics.SmartcarIngestSuccessOps.Inc()
	}
	if integration.Vendor == AutoPiVendor {
		appmetrics.AutoPiIngestSuccessOps.Inc()
	}
	return nil
}

// processOdometer get the odometer diff
func (i *IngestService) processOdometer(datum *models.UserDeviceDatum, newOdometer null.Float64, device *models.UserDevice, integrationID string) {
	oldOdometer := null.Float64FromPtr(nil)
	if datum.Data.Valid {
		if o, err := extractOdometer(datum.Data.JSON); err == nil {
			oldOdometer = null.Float64From(o)
		}
	}

	now := time.Now()
	odometerOffCooldown := !datum.LastOdometerEventAt.Valid || now.Sub(datum.LastOdometerEventAt.Time) >= odometerCooldown
	odometerChanged := !oldOdometer.Valid || newOdometer.Float64 > oldOdometer.Float64

	if odometerOffCooldown && odometerChanged {
		datum.LastOdometerEventAt = null.TimeFrom(now)
		i.emitOdometerEvent(device, integrationID, newOdometer.Float64)
	}
}

func (i *IngestService) emitOdometerEvent(device *models.UserDevice, integrationID string, odometer float64) {
	event := &Event{
		Type:    "com.dimo.zone.device.odometer.update",
		Subject: device.ID,
		Source:  "dimo/integration/" + integrationID,
		Data: OdometerEvent{
			Timestamp: time.Now(),
			UserID:    device.UserID,
			Device: odometerEventDevice{
				ID:    device.ID,
				Make:  device.R.DeviceDefinition.R.DeviceMake.Name,
				Model: device.R.DeviceDefinition.Model,
				Year:  int(device.R.DeviceDefinition.Year),
			},
			Odometer: odometer,
		},
	}
	if err := i.eventService.Emit(event); err != nil {
		i.log.Err(err).Msgf("Failed to emit odometer event for device %s", device.ID)
	}
}

func extractOdometer(data []byte) (float64, error) {
	partialData := new(struct {
		Odometer *float64 `json:"odometer"`
	})
	if err := json.Unmarshal(data, partialData); err != nil {
		return 0, fmt.Errorf("failed parsing data field: %w", err)
	}
	if partialData.Odometer == nil {
		return 0, errors.New("data payload did not have an odometer reading")
	}

	return *partialData.Odometer, nil
}

func (i *IngestService) getIntegrationFromEvent(event *DeviceStatusEvent) (*models.Integration, error) {
	for _, integration := range i.integrations {
		if strings.Contains(event.Source, integration.ID) {
			return integration, nil
		}
	}
	return nil, fmt.Errorf("no matching integration found in DB for event source: %s", event.Source)
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
