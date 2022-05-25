package services

import (
	"context"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestTaskStatusListener(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "devices-api").Logger()

	ctx := context.Background()
	pdb, container := test.StartContainerDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	ingest := NewTaskStatusListener(pdb.DBS, &logger)

	scIntegration := test.SetupCreateSmartCarIntegration(t, pdb)
	dm := test.SetupCreateMake(t, "Tesla", pdb)
	dd := test.SetupCreateDeviceDefinition(t, dm, "Model Y", 2021, pdb)
	ud := test.SetupCreateUserDevice(t, "dylan", dd, nil, pdb)

	udai := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: scIntegration.ID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
	}
	err := udai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)

	input := &TaskStatusCloudEvent{
		CloudEventHeaders: CloudEventHeaders{
			Source:      "dimo/integration/" + scIntegration.ID,
			SpecVersion: "1.0",
			Subject:     ud.ID,
			Type:        "zone.dimo.task.smartcar.poll.status.update",
		},
		Data: TaskStatusData{
			TaskID:        ksuid.New().String(),
			UserDeviceID:  ud.ID,
			IntegrationID: scIntegration.ID,
			Status:        "AuthenticationFailure",
		},
	}

	if err := ingest.processEvent(input); err != nil {
		t.Fatalf("Got an unexpected error processing status update: %v", err)
	}

	if err := udai.Reload(ctx, pdb.DBS().Writer); err != nil {
		t.Fatalf("Couldn't reload UDAI: %v", err)
	}

	assert.Equal(t, models.UserDeviceAPIIntegrationStatusAuthenticationFailure, udai.Status, "New status should be Failed.")
}
