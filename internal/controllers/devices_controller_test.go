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

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/services"
	mock_services "github.com/DIMO-INC/devices-api/internal/services/mocks"
	_ "github.com/DIMO-INC/devices-api/migrations"
	"github.com/DIMO-INC/devices-api/models"
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

func TestDevicesController_GetUsersDevices(t *testing.T) {
	ctx := context.Background()

	pdb, database := setupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := database.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	c := NewDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, nil, nil)

	app := fiber.New()
	app.Get("/devices", c.GetUsersDevices)

	request, _ := http.NewRequest("GET", "/devices", nil)
	response, _ := app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "{\"devices\":[{\"device_id\":\"123123\",\"name\":\"Johnny's Tesla\"}]}", string(body))
}

func TestDevicesController_LookupDeviceDefinitionByVIN(t *testing.T) {
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
	c := NewDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, nhtsaSvc)
	vinResp := services.NHTSADecodeVINResponse{}
	_ = json.Unmarshal([]byte(testNhtsaDecodedVin), &vinResp)
	const vin = "5YJYGDEF2LFR00942"

	nhtsaSvc.EXPECT().DecodeVIN(vin).Times(1).Return(&vinResp, nil)

	app := fiber.New()
	app.Get("/devices/lookup/vin/:vin", c.LookupDeviceDefinitionByVIN)

	request, _ := http.NewRequest("GET", "/devices/lookup/vin/"+vin, nil)
	response, _ := app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(t, 200, response.StatusCode)
	definition, err := models.DeviceDefinitions().One(ctx, pdb.DBS().Writer)
	assert.NoError(t, err, "expected to find one device def in DB")
	assert.NotNilf(t, definition, "expected device def not be nil")
	assert.Equal(t, vin[:10], definition.VinFirst10)

	fmt.Println(string(body))
}

func TestNewDeviceDefinitionFromNHTSA(t *testing.T) {
	vinResp := services.NHTSADecodeVINResponse{}
	_ = json.Unmarshal([]byte(testNhtsaDecodedVin), &vinResp)

	deviceDefinition := NewDeviceDefinitionFromNHTSA(&vinResp)

	assert.Equal(t, "", deviceDefinition.DeviceDefinitionID)
	assert.Equal(t, "2020 TESLA Model Y", deviceDefinition.Name)
	assert.Equal(t, "Vehicle", deviceDefinition.Type.Type)
	assert.Equal(t, 2020, deviceDefinition.Type.Year)
	assert.Equal(t, "TESLA", deviceDefinition.Type.Make)
	assert.Equal(t, "Model Y", deviceDefinition.Type.Model)
	assert.Equal(t, "", deviceDefinition.Type.SubModel)
	assert.Equal(t, "PASSENGER CAR", deviceDefinition.VehicleInfo.VehicleType)
	assert.Equal(t, 48000, deviceDefinition.VehicleInfo.BaseMSRP)
	assert.Equal(t, "5", deviceDefinition.VehicleInfo.NumberOfDoors)
	assert.Equal(t, "Electric", deviceDefinition.VehicleInfo.FuelType)
}

func TestNewDeviceDefinitionFromDatabase(t *testing.T) {
	dbDevice := models.DeviceDefinition{
		UUID:       "123",
		VinFirst10: "1231231231",
		Make:       "Merc",
		Model:      "R500",
		Year:       2020,
		SubModel:   null.StringFrom("AMG"),
		Metadata:   null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
	}
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
		VehicleInfo: DeviceVehicleInfo{
			FuelType:      "gas",
			DrivenWheels:  "4",
			NumberOfDoors: "5",
		},
	}
	dbDevice := NewDbModelFromDeviceDefinition(dd, "1231231")

	assert.Equal(t, "1231231", dbDevice.VinFirst10)
	assert.Equal(t, "R500", dbDevice.Model)
	assert.Equal(t, "Merc", dbDevice.Make)
	assert.Equal(t, int16(2020), dbDevice.Year)
	assert.Equal(t, "AMG", dbDevice.SubModel.String)
	assert.Equal(t, `{"vehicle_info":{"fuel_type":"gas","driven_wheels":"4","number_of_doors":"5"}}`, string(dbDevice.Metadata.JSON))
}
