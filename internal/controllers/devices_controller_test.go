package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
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


