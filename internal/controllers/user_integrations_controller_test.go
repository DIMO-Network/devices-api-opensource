package controllers

import (
	"context"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/stretchr/testify/assert"
)

func Test_createDeviceIntegrationIfAutoPi(t *testing.T) {
	ctx := context.Background()
	pdb := test.GetDBConnection(ctx)

	const region = "North America"
	t.Run("create with nothing existing returns nil, nil", func(t *testing.T) {
		di, err := createDeviceIntegrationIfAutoPi(ctx, "123", "123", region, pdb.DBS().Writer)

		assert.NoError(t, err)
		assert.Nil(t, di, "expected device integration to be nil")

		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("create with existing autopi integration returns new device_integration, and .R.Integration", func(t *testing.T) {
		autoPiInteg := test.SetupCreateAutoPiIntegration(t, 34, pdb)
		dm := test.SetupCreateMake(t, "Testla", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Model 4", 2022, pdb)
		// act
		di, err := createDeviceIntegrationIfAutoPi(ctx, autoPiInteg.ID, dd.ID, region, pdb.DBS().Writer)
		// assert
		assert.NoError(t, err)
		assert.NotNilf(t, di, "device integration should not be nil")
		assert.Equal(t, autoPiInteg.ID, di.IntegrationID)
		assert.Equal(t, dd.ID, di.DeviceDefinitionID)
		assert.Equal(t, region, di.Region)
		assert.NotNilf(t, di.R.Integration, "relationship to integration should not be nil")
		assert.Equal(t, services.AutoPiVendor, di.R.Integration.Vendor)

		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})

}
