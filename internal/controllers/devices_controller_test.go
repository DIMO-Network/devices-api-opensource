package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/buger/jsonparser"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const migrationsDirRelPath = "../../migrations"

// integration tests using embedded pgsql, must be run in order
func TestDevicesController(t *testing.T) {
	// arrange global db and route setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ctx := context.Background()
	pdb, database := test.SetupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := database.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	nhtsaSvc := mock_services.NewMockINHTSAService(mockCtrl)
	deviceDefSvc := mock_services.NewMockIDeviceDefinitionService(mockCtrl)
	c := NewDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, nhtsaSvc, deviceDefSvc)
	// routes
	app := fiber.New()
	app.Get("/device-definitions/all", c.GetAllDeviceMakeModelYears)
	app.Get("/device-definitions/:id", c.GetDeviceDefinitionByID)
	app.Get("/device-definitions/:id/integrations", c.GetIntegrationsByID)

	createdID := ksuid.New().String()
	dbDeviceDef := models.DeviceDefinition{
		ID:       createdID,
		Make:     "TESLA",
		Model:    "MODEL Y",
		Year:     2020,
		Verified: true,
	}
	dbErr := dbDeviceDef.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, dbErr)
	fmt.Println("created device def id: " + createdID)

	t.Run("GET - device definition by id", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/device-definitions/"+createdID, nil)
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		assert.Equal(t, 200, response.StatusCode)
		v, _, _, _ := jsonparser.Get(body, "deviceDefinition")
		var dd services.DeviceDefinition
		err := json.Unmarshal(v, &dd)
		assert.NoError(t, err)
		assert.Equal(t, createdID, dd.DeviceDefinitionID)
	})
	t.Run("GET - device integrations by id", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/device-definitions/"+createdID+"/integrations", nil)
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		assert.Equal(t, 200, response.StatusCode)
		v, _, _, _ := jsonparser.Get(body, "compatibleIntegrations")
		var dc []services.DeviceCompatibility
		err := json.Unmarshal(v, &dc)
		assert.NoError(t, err)
	})
	t.Run("GET 400 - device definition by id invalid", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/device-definitions/caca", nil)
		response, _ := app.Test(request)
		// assert
		assert.Equal(t, 400, response.StatusCode)
	})
	t.Run("GET 400 - device definition integrations invalid", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/device-definitions/caca/integrations", nil)
		response, _ := app.Test(request)
		// assert
		assert.Equal(t, 400, response.StatusCode)
	})
	t.Run("GET - all make model years as a tree", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/device-definitions/all", nil)
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		assert.Equal(t, 200, response.StatusCode)
		v, _, _, _ := jsonparser.Get(body, "makes")
		var mmy []DeviceMMYRoot
		err := json.Unmarshal(v, &mmy)
		assert.NoError(t, err)
		assert.Len(t, mmy, 1)
		assert.Equal(t, "TESLA", mmy[0].Make)
		assert.Equal(t, "MODEL Y", mmy[0].Models[0].Model)
		assert.Equal(t, int16(2020), mmy[0].Models[0].Years[0].Year)
		assert.Equal(t, createdID, mmy[0].Models[0].Years[0].DeviceDefinitionID)
	})
}

func TestNewDeviceDefinitionFromDatabase(t *testing.T) {
	dbDevice := models.DeviceDefinition{
		ID:       "123",
		Make:     "Merc",
		Model:    "R500",
		Year:     2020,
		Metadata: null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
	}
	ds := models.DeviceStyle{
		SubModel:           "AMG",
		Name:               "C63 AMG",
		DeviceDefinitionID: dbDevice.ID,
	}
	di := models.DeviceIntegration{
		DeviceDefinitionID: "123",
		IntegrationID:      "123",
		CreatedAt:          time.Time{},
		UpdatedAt:          time.Time{},
	}
	di.R = di.R.NewStruct()
	di.R.Integration = &models.Integration{
		ID:     "123",
		Type:   "Hardware",
		Style:  "Addon",
		Vendor: "Autopi",
	}
	dbDevice.R = dbDevice.R.NewStruct()
	dbDevice.R.DeviceIntegrations = append(dbDevice.R.DeviceIntegrations, &di)
	dbDevice.R.DeviceStyles = append(dbDevice.R.DeviceStyles, &ds)
	dd := NewDeviceDefinitionFromDatabase(&dbDevice)

	assert.Equal(t, "123", dd.DeviceDefinitionID)
	assert.Equal(t, "gas", dd.VehicleInfo.FuelType)
	assert.Equal(t, "4", dd.VehicleInfo.DrivenWheels)
	assert.Equal(t, "5", dd.VehicleInfo.NumberOfDoors)
	assert.Equal(t, "Vehicle", dd.Type.Type)
	assert.Equal(t, 2020, dd.Type.Year)
	assert.Equal(t, "Merc", dd.Type.Make)
	assert.Equal(t, "R500", dd.Type.Model)
	assert.Contains(t, dd.Type.SubModels, "AMG")

	assert.Len(t, dd.CompatibleIntegrations, 1)
	assert.Equal(t, "Autopi", dd.CompatibleIntegrations[0].Vendor)
}

func TestNewDbModelFromDeviceDefinition(t *testing.T) {
	dd := services.DeviceDefinition{
		Type: services.DeviceType{
			Type:  "Vehicle",
			Make:  "Merc",
			Model: "R500",
			Year:  2020,
		},
		VehicleInfo: services.DeviceVehicleInfo{
			FuelType:      "gas",
			DrivenWheels:  "4",
			NumberOfDoors: "5",
		},
	}
	dbDevice := NewDbModelFromDeviceDefinition(dd, nil)

	assert.Equal(t, "R500", dbDevice.Model)
	assert.Equal(t, "Merc", dbDevice.Make)
	assert.Equal(t, int16(2020), dbDevice.Year)
	assert.Equal(t, `{"vehicle_info":{"fuel_type":"gas","driven_wheels":"4","number_of_doors":"5"}}`, string(dbDevice.Metadata.JSON))
}
