package services

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
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

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ctx := context.Background()
	pdb, db := test.SetupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := db.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	is := NewIngestService(pdb.DBS, &logger, mes)

	scIntegration := test.SetupCreateSmartCarIntegration(t, pdb)
	dm := test.SetupCreateMake(t, "Tesla", pdb)
	dd := test.SetupCreateDeviceDefinition(t, dm, "Model Y", 2021, pdb)
	ud := test.SetupCreateUserDevice(t, "dylan", dd, pdb)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: scIntegration.ID,
		Status:        models.UserDeviceAPIIntegrationStatusPendingFirstData,
	}
	err := udai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	// "No API integration found for device 264bqKKB5rFp8ztfJw1gVkAid4x and integration 264bqJfPTi7UsdurCtRA8ucH54i"
	err = is.processEvent(&DeviceStatusEvent{
		Source:      "dimo/integration/" + scIntegration.ID,
		Specversion: "1.0",
		Subject:     ud.ID,
		Type:        deviceStatusEventType,
		Data:        json.RawMessage(`{"odometer": 45.1}`),
	})
	assert.NoError(t, err, "expected no errors from first status event")

	newUDAI, _ := models.FindUserDeviceAPIIntegration(ctx, pdb.DBS().Writer, ud.ID, scIntegration.ID)

	assert.Equal(t, models.UserDeviceAPIIntegrationStatusActive, newUDAI.Status, "integration should be set to active")

	data, _ := models.FindUserDeviceDatum(ctx, pdb.DBS().Writer, ud.ID)

	assert.Equal(t, []byte(`{"odometer": 45.1}`), data.Data.JSON, "should have updated the data field")

	assert.Equal(t, 45.1, mes.Buffer[0].Data.(OdometerEvent).Odometer)

	err = is.processEvent(&DeviceStatusEvent{
		Source:      "dimo/integration/" + scIntegration.ID,
		Specversion: "1.0",
		Subject:     ud.ID,
		Type:        deviceStatusEventType,
		Data:        json.RawMessage(`{"odometer": 55.2}`),
	})
	assert.NoError(t, err, "expected no errors from second status event")

	assert.Equal(t, 55.2, mes.Buffer[1].Data.(OdometerEvent).Odometer)

}
