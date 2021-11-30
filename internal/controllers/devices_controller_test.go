package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"io/ioutil"
	"net/http"
	"testing"
)

//go:embed test_nhtsa_decoded_vin.json
var testNhtsaDecodedVin string

func TestDevicesController_GetUsersDevices(t *testing.T) {
	ctx := context.Background()
	pdb, database := setupDatabase(ctx, t)
	defer func() {
		if err := database.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	c := NewDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, nil)

	app := fiber.New()
	app.Get("/devices", c.GetUsersDevices)

	request, _ := http.NewRequest("GET", "/devices", nil)
	response, _ := app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "{\"devices\":[{\"device_id\":\"123123\",\"name\":\"Johnny's Tesla\"}]}", string(body))
}

func TestDevicesController_LookupDeviceDefinitionByVIN(t *testing.T) {
	// just use mock db instead of embedded pgsql
}

func TestNewDeviceDefinitionFromNHTSA(t *testing.T) {
	vinResp := services.NHTSADecodeVINResponse{}
	_ = json.Unmarshal([]byte(testNhtsaDecodedVin), &vinResp)

	deviceDefinition := NewDeviceDefinitionFromNHTSA(&vinResp)

	assert.Equal(t, "", deviceDefinition.DeviceDefinitionId)
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
		OtherData:  null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
	}
	dd := NewDeviceDefinitionFromDatabase(&dbDevice)

	assert.Equal(t, "123", dd.DeviceDefinitionId)
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
		Type:               DeviceType{
			Type:     "Vehicle",
			Make:     "Merc",
			Model:    "R500",
			Year:     2020,
			SubModel: "AMG",
		},
		VehicleInfo:        DeviceVehicleInfo{
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
	assert.Equal(t, `{"vehicle_info":{"fuel_type":"gas","driven_wheels":"4","number_of_doors":"5"}}`, string(dbDevice.OtherData.JSON))
}


