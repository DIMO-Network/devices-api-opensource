package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/services"
	mock_services "github.com/DIMO-INC/devices-api/internal/services/mocks"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/buger/jsonparser"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
)

//go:embed test_nhtsa_decoded_vin.json
var testNhtsaDecodedVin string

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
	pdb, database := setupDatabase(ctx, t, migrationsDirRelPath)
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
		v, _, _, _ := jsonparser.Get(body, "device_definition")
		var dd DeviceDefinition
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
		v, _, _, _ := jsonparser.Get(body, "compatible_integrations")
		var dc []DeviceCompatibility
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

func TestNewDeviceDefinitionFromNHTSA(t *testing.T) {
	vinResp := services.NHTSADecodeVINResponse{}
	_ = json.Unmarshal([]byte(testNhtsaDecodedVin), &vinResp)

	deviceDefinition := NewDeviceDefinitionFromNHTSA(&vinResp)

	assert.Equal(t, "", deviceDefinition.DeviceDefinitionID)
	assert.Equal(t, "2020 TESLA MODEL Y", deviceDefinition.Name)
	assert.Equal(t, "Vehicle", deviceDefinition.Type.Type)
	assert.Equal(t, 2020, deviceDefinition.Type.Year)
	assert.Equal(t, "TESLA", deviceDefinition.Type.Make)
	assert.Equal(t, "MODEL Y", deviceDefinition.Type.Model)
	assert.Equal(t, "", deviceDefinition.Type.SubModel)
	assert.Equal(t, "PASSENGER CAR", deviceDefinition.VehicleInfo.VehicleType)
	assert.Equal(t, 48000, deviceDefinition.VehicleInfo.BaseMSRP)
	assert.Equal(t, "5", deviceDefinition.VehicleInfo.NumberOfDoors)
	assert.Equal(t, "ELECTRIC", deviceDefinition.VehicleInfo.FuelType)
}

func TestNewDeviceDefinitionFromDatabase(t *testing.T) {
	dbDevice := models.DeviceDefinition{
		ID:       "123",
		Make:     "Merc",
		Model:    "R500",
		Year:     2020,
		SubModel: null.StringFrom("AMG"),
		Metadata: null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
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
	dd := NewDeviceDefinitionFromDatabase(&dbDevice)

	assert.Equal(t, "123", dd.DeviceDefinitionID)
	assert.Equal(t, "gas", dd.VehicleInfo.FuelType)
	assert.Equal(t, "4", dd.VehicleInfo.DrivenWheels)
	assert.Equal(t, "5", dd.VehicleInfo.NumberOfDoors)
	assert.Equal(t, "Vehicle", dd.Type.Type)
	assert.Equal(t, 2020, dd.Type.Year)
	assert.Equal(t, "Merc", dd.Type.Make)
	assert.Equal(t, "R500", dd.Type.Model)
	assert.Equal(t, "AMG", dd.Type.SubModel)

	assert.Len(t, dd.CompatibleIntegrations, 1)
	assert.Equal(t, "Autopi", dd.CompatibleIntegrations[0].Vendor)
}

func TestNewDbModelFromDeviceDefinition(t *testing.T) {
	dd := DeviceDefinition{
		Type: DeviceType{
			Type:     "Vehicle",
			Make:     "Merc",
			Model:    "R500",
			Year:     2020,
			SubModel: "AMG",
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
	assert.Equal(t, "AMG", dbDevice.SubModel.String)
	assert.Equal(t, `{"vehicle_info":{"fuel_type":"gas","driven_wheels":"4","number_of_doors":"5"}}`, string(dbDevice.Metadata.JSON))
}
