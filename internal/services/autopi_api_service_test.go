package services

import (
	"context"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestFindUserDeviceAutoPiIntegration(t *testing.T) {
	ctx := context.Background()
	pdb, db := test.SetupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := db.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	// arrange some data
	const testUserID = "123123"
	const autoPiDeviceID = "321"
	autoPiUnitID := "456"
	apInt := test.SetupCreateAutoPiIntegration(t, 10, nil, pdb)
	scInt := test.SetupCreateSmartCarIntegration(t, pdb)
	dm := test.SetupCreateMake(t, "Tesla", pdb)
	dd := test.SetupCreateDeviceDefinition(t, dm, "Model 3", 2020, pdb)
	test.SetupCreateDeviceIntegration(t, dd, apInt, pdb)
	test.SetupCreateDeviceIntegration(t, dd, scInt, pdb)
	ud := test.SetupCreateUserDevice(t, testUserID, dd, nil, pdb)
	// now create the api ints
	scUdai := &models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: scInt.ID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
		ExternalID:    null.StringFrom("423324"),
	}
	err := scUdai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	amd := UserDeviceAPIIntegrationsMetadata{
		AutoPiUnitID: &autoPiUnitID,
	}
	apUdai := &models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: apInt.ID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
		ExternalID:    null.StringFrom(autoPiDeviceID),
	}
	_ = apUdai.Metadata.Marshal(amd)
	err = apUdai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	// act  now call the method
	udIntegration, metadata, err := FindUserDeviceAutoPiIntegration(ctx, pdb.DBS().Writer, ud.ID, testUserID)
	assert.NoError(t, err)
	assert.NotNil(t, udIntegration, "expected user_device_api_integration not be nil")
	assert.NotNilf(t, metadata, "expected metadata not be nil")
	assert.Equal(t, ud.ID, udIntegration.UserDeviceID)
	assert.Equal(t, apInt.ID, udIntegration.IntegrationID)
	assert.Equal(t, autoPiDeviceID, udIntegration.ExternalID.String)
	// check some values
	test.TruncateTables(pdb.DBS().Writer.DB, t)
}

func TestAppendAutoPiCompatibility(t *testing.T) {
	ctx := context.Background()
	pdb, db := test.SetupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := db.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	dm := test.SetupCreateMake(t, "Ford", pdb)
	dd := test.SetupCreateDeviceDefinition(t, dm, "MachE", 2020, pdb)
	var dcs []DeviceCompatibility
	compatibility, err := AppendAutoPiCompatibility(ctx, dcs, dd.ID, pdb.DBS().Writer)

	assert.NoError(t, err)
	assert.Len(t, compatibility, 2)
	all, err := models.DeviceIntegrations().All(ctx, pdb.DBS().Reader)
	assert.NoError(t, err)
	assert.Len(t, all, 2)

	test.TruncateTables(pdb.DBS().Writer.DB, t)
}
