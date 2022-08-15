package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/appmetrics"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	deviceStatusEventType = "zone.dimo.device.status.update"
	odometerCooldown      = time.Hour
)

type DeviceStatusIngestService struct {
	db           func() *database.DBReaderWriter
	log          *zerolog.Logger
	eventService EventService
	integrations models.IntegrationSlice
}

func NewDeviceStatusIngestService(db func() *database.DBReaderWriter, log *zerolog.Logger, eventService EventService) *DeviceStatusIngestService {
	// Cache the list of integrations.
	integrations, err := models.Integrations().All(context.Background(), db().Reader)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't retrieve global integration list.")
	}
	return &DeviceStatusIngestService{
		db:           db,
		log:          log,
		eventService: eventService,
		integrations: integrations,
	}
}

// ProcessDeviceStatusMessages works on channel stream of messages from watermill kafka consumer
func (i *DeviceStatusIngestService) ProcessDeviceStatusMessages(messages <-chan *message.Message) {
	for msg := range messages {
		if err := i.processMessage(msg); err != nil {
			i.log.Err(err).Msg("Error processing device status message.")
		}
	}
}

func (i *DeviceStatusIngestService) processMessage(msg *message.Message) error {
	// Keep the pipeline moving no matter what.
	defer func() { msg.Ack() }()

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

	switch integration.Vendor {
	case SmartCarVendor:
		defer appmetrics.SmartcarIngestTotalOps.Inc()
	case AutoPiVendor:
		defer appmetrics.AutoPiIngestTotalOps.Inc()
	}

	return i.processEvent(event)
}

var userDeviceDataPrimaryKeyColumns = []string{models.UserDeviceDatumColumns.UserDeviceID, models.UserDeviceDatumColumns.IntegrationID}

func (i *DeviceStatusIngestService) processEvent(event *DeviceStatusEvent) error {
	ctx := context.Background() // should this be passed in so can cancel if application shutting down?
	userDeviceID := event.Subject

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
		qm.Load(
			models.UserDeviceRels.UserDeviceData,
			models.UserDeviceDatumWhere.IntegrationID.EQ(integration.ID),
		),
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

	// Null for most AutoPis.
	var newOdometer null.Float64
	if o, err := extractOdometer(event.Data); err == nil {
		newOdometer = null.Float64From(o)
	} else if integration.Vendor == AutoPiVendor {
		// For AutoPis, for the purpose of odometer events we are pretending to always have
		// an odometer reading. Users became accustomed to seeing the associated events, even
		// though we mostly don't have odometer readings for AutoPis. For now, we fake it.
		newOdometer = null.Float64From(0.0)
	}

	var datum *models.UserDeviceDatum
	if len(device.R.UserDeviceData) > 0 {
		// Update the existing record.
		datum = device.R.UserDeviceData[0]
	} else {
		// Insert a new record.
		datum = &models.UserDeviceDatum{UserDeviceID: userDeviceID, IntegrationID: integration.ID}
	}

	i.processOdometer(datum, newOdometer, device, integration.ID)

	switch integration.Vendor {
	case SmartCarVendor, TeslaVendor:
		if newOdometer.Valid {
			datum.Data = null.JSONFrom(event.Data)
			datum.ErrorData = null.JSON{}
		} else {
			datum.ErrorData = null.JSONFrom(event.Data)
		}
	// Again, most AutoPis don't have decoded odometer readings, so just let updates through.
	case AutoPiVendor:
		// Not every AutoPi update has every signal. Merge the new into the old.
		compositeData := make(map[string]any)
		if err := datum.Data.Unmarshal(&compositeData); err != nil {
			return err
		}

		// This will preserve any mappings with keys present in datum.Data but not in
		// event.Data, but mappings in event.Data take precedence.
		//
		// For example, if in the database we have {A: 1, B: 2} and the new event has
		// {B: 4, C: 9} then the result should be {A: 1, B: 4, C: 9}.
		if err := json.Unmarshal(event.Data, &compositeData); err != nil {
			return err
		}

		if err := datum.Data.Marshal(compositeData); err != nil {
			return err
		}
		datum.ErrorData = null.JSON{}
	default:
		// Not sure what this is.
		datum.Data = null.JSONFrom(event.Data)
		datum.ErrorData = null.JSON{}
	}

	if err := datum.Upsert(ctx, tx, true, userDeviceDataPrimaryKeyColumns, boil.Infer(), boil.Infer()); err != nil {
		return fmt.Errorf("error upserting datum: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	switch integration.Vendor {
	case SmartCarVendor:
		appmetrics.SmartcarIngestSuccessOps.Inc()
	case AutoPiVendor:
		appmetrics.AutoPiIngestSuccessOps.Inc()
	}

	return nil
}

// processOdometer emits an odometer event and updates the last_odometer_event timestamp on the
// data record if the following conditions are met:
//   - there is no existing timestamp, or an hour has passed since that timestamp,
//   - the incoming status update has an odometer value, and
//   - the old status update lacks an odometer value, or has an odometer value that differs from
//     the new odometer reading
func (i *DeviceStatusIngestService) processOdometer(datum *models.UserDeviceDatum, newOdometer null.Float64, device *models.UserDevice, integrationID string) {
	if !newOdometer.Valid {
		return
	}

	var oldOdometer null.Float64
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

func (i *DeviceStatusIngestService) emitOdometerEvent(device *models.UserDevice, integrationID string, odometer float64) {
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

func (i *DeviceStatusIngestService) getIntegrationFromEvent(event *DeviceStatusEvent) (*models.Integration, error) {
	for _, integration := range i.integrations {
		if strings.HasSuffix(event.Source, integration.ID) {
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
