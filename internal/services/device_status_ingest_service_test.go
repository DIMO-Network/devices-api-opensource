package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type testEventService struct {
	Buffer []*Event
}

func (e *testEventService) Emit(event *Event) error {
	e.Buffer = append(e.Buffer, event)
	return nil
}

const migrationsDirRelPath = "../../migrations"

func TestIngestDeviceStatus(t *testing.T) {
	mes := &testEventService{
		Buffer: make([]*Event, 0),
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "devices-api").Logger()
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	scIntegration := test.SetupCreateSmartCarIntegration(t, pdb)

	ingest := NewDeviceStatusIngestService(pdb.DBS, &logger, mes)

	dm := test.SetupCreateMake(t, "Tesla", pdb)
	dd := test.SetupCreateDeviceDefinition(t, dm, "Model Y", 2021, pdb)
	ud := test.SetupCreateUserDevice(t, "dylan", dd, nil, pdb)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: scIntegration.ID,
		Status:        models.UserDeviceAPIIntegrationStatusPendingFirstData,
	}
	err := udai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	testCases := []struct {
		Name                string
		ExistingData        null.JSON
		NewData             null.JSON
		LastOdometerEventAt null.Time
		ExpectedEvent       null.Float64
	}{
		// todo commenting these out because they are failing and not sure if this is still something we need or conflicts with latest functionality
		{
			Name:                "New reading, none prior",
			ExistingData:        null.JSON{},
			NewData:             null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			LastOdometerEventAt: null.Time{},
			ExpectedEvent:       null.Float64From(12.5),
		},
		{
			Name:                "Odometer changed, event off cooldown",
			ExistingData:        null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			NewData:             null.JSONFrom([]byte(`{"odometer": 14.5}`)),
			LastOdometerEventAt: null.TimeFrom(time.Now().Add(-2 * odometerCooldown)),
			ExpectedEvent:       null.Float64From(14.5),
		},
		{
			Name:                "Event off cooldown, odometer unchanged",
			ExistingData:        null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			NewData:             null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			LastOdometerEventAt: null.TimeFrom(time.Now().Add(-2 * odometerCooldown)),
			ExpectedEvent:       null.Float64{},
		},
		{
			Name:                "Odometer changed, but event on cooldown",
			ExistingData:        null.JSONFrom([]byte(`{"odometer": 12.5}`)),
			NewData:             null.JSONFrom([]byte(`{"odometer": 14.5}`)),
			LastOdometerEventAt: null.TimeFrom(time.Now().Add(odometerCooldown / 2)),
			ExpectedEvent:       null.Float64{},
		},
	}

	tx := pdb.DBS().Writer

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			defer func() { mes.Buffer = nil }()

			datum := models.UserDeviceDatum{
				UserDeviceID:        ud.ID,
				Data:                c.ExistingData,
				LastOdometerEventAt: c.LastOdometerEventAt,
				IntegrationID:       scIntegration.ID,
			}

			err := datum.Upsert(ctx, tx, true, []string{models.UserDeviceDatumColumns.UserDeviceID, models.UserDeviceDatumColumns.IntegrationID},
				boil.Infer(), boil.Infer())
			if err != nil {
				t.Fatalf("Failed setting up existing data row: %v", err)
			}

			input := &DeviceStatusEvent{
				Source:      "dimo/integration/" + scIntegration.ID,
				Specversion: "1.0",
				Subject:     ud.ID,
				Type:        deviceStatusEventType,
				Data:        c.NewData.JSON,
			}

			if err := ingest.processEvent(input); err != nil {
				t.Fatalf("Got an unexpected error processing status update: %v", err)
			}
			if c.ExpectedEvent.Valid {
				if len(mes.Buffer) != 1 {
					t.Fatalf("Expected one odometer event, but got %d", len(mes.Buffer))
				}
				// A bit ugly to have to cast like this.
				actualOdometer := mes.Buffer[0].Data.(OdometerEvent).Odometer
				if actualOdometer != c.ExpectedEvent.Float64 {
					t.Fatalf("Expected an odometer reading of %f but got %f", c.ExpectedEvent.Float64, actualOdometer)
				}
			} else if len(mes.Buffer) != 0 {
				t.Fatalf("Expected no odometer events, but got %d", len(mes.Buffer))
			}
		})
	}
}

func TestAutoPiStatusMerge(t *testing.T) {
	assert := assert.New(t)

	mes := &testEventService{
		Buffer: make([]*Event, 0),
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "devices-api").Logger()
	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	// Only making use the last parameter.
	apInt := test.SetupCreateAutoPiIntegration(t, 1, nil, pdb)

	ingest := NewDeviceStatusIngestService(pdb.DBS, &logger, mes)

	dm := test.SetupCreateMake(t, "Toyota", pdb)
	dd := test.SetupCreateDeviceDefinition(t, dm, "RAV4", 2021, pdb)
	ud := test.SetupCreateUserDevice(t, "dylan", dd, nil, pdb)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: apInt.ID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
	}

	err := udai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(err)

	tx := pdb.DBS().Writer

	dat1 := models.UserDeviceDatum{
		UserDeviceID:        ud.ID,
		Data:                null.JSONFrom([]byte(`{"odometer": 45.22, "latitude": 11.0, "longitude": -7.0}`)),
		LastOdometerEventAt: null.TimeFrom(time.Now().Add(-10 * time.Second)),
		IntegrationID:       apInt.ID,
	}

	err = dat1.Insert(ctx, tx, boil.Infer())
	assert.NoError(err)

	input := &DeviceStatusEvent{
		Source:      "dimo/integration/" + apInt.ID,
		Specversion: "1.0",
		Subject:     ud.ID,
		Type:        deviceStatusEventType,
		Time:        time.Now(),
		Data:        []byte(`{"latitude": 2.0, "longitude": 3.0}`),
	}

	err = ingest.processEvent(input)
	assert.NoError(err)

	err = dat1.Reload(ctx, tx)
	assert.NoError(err)

	assert.JSONEq(`{"odometer": 45.22, "latitude": 2.0, "longitude": 3.0}`, string(dat1.Data.JSON))
}
