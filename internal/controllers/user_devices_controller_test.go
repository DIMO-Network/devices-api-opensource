package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/DIMO-INC/devices-api/internal/config"
	mock_services "github.com/DIMO-INC/devices-api/internal/services/mocks"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestUserDevicesController(t *testing.T) {
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
	deviceDefSvc := mock_services.NewMockIDeviceDefinitionService(mockCtrl)

	testUserID := "123123"
	testUserID2 := "3232451"
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, deviceDefSvc)
	app := fiber.New()
	app.Post("/user/devices", authInjectorTestHandler(testUserID), c.RegisterDeviceForUser)
	app.Post("/user/devices/second", authInjectorTestHandler(testUserID2), c.RegisterDeviceForUser)
	app.Post("/admin/user/:user_id/devices", c.AdminRegisterUserDevice)
	app.Get("/user/devices/me", authInjectorTestHandler(testUserID), c.GetUserDevices)
	app.Patch("/user/devices/:user_device_id/vin", authInjectorTestHandler(testUserID), c.UpdateVIN)
	app.Patch("/user/devices/:user_device_id/name", authInjectorTestHandler(testUserID), c.UpdateName)

	deviceDefSvc.EXPECT().CheckAndSetImage(gomock.Any()).Return(nil)
	// todo: put this expectation in right place with correct values.
	deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false).Return(nil, nil)
	createdUserDeviceID := ""

	t.Run("POST - register with existing device_definition_id", func(t *testing.T) {
		// arrange DB
		ddID := ksuid.New().String()
		dd := models.DeviceDefinition{
			ID:       ddID,
			Make:     "Tesla",
			Model:    "Model X",
			Year:     2020,
			Verified: true,
		}
		err := dd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err, "database error")
		integration := models.Integration{
			ID:     ksuid.New().String(),
			Type:   models.IntegrationTypeAPI,
			Style:  models.IntegrationStyleWebhook,
			Vendor: "SmartCar",
		}
		err = integration.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err, "database error")
		deviceInt := models.DeviceIntegration{
			DeviceDefinitionID: ddID,
			IntegrationID:      integration.ID,
			Country:            "USA",
		}
		err = deviceInt.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err, "database error")
		// act request
		cc := "USA"
		reg := RegisterUserDevice{
			DeviceDefinitionID: &ddID,
			CountryCode:        &cc,
		}
		j, _ := json.Marshal(reg)
		request := buildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		regUserResp := RegisterUserDeviceResponse{}
		_ = json.Unmarshal(body, &regUserResp)
		assert.Len(t, regUserResp.UserDeviceID, 27)
		assert.Len(t, regUserResp.DeviceDefinitionID, 27)
		assert.Equal(t, ddID, regUserResp.DeviceDefinitionID)
		if assert.Len(t, regUserResp.IntegrationCapabilities, 1) == false {
			fmt.Println("resp body: " + string(body))
		}
		assert.Equal(t, integration.Vendor, regUserResp.IntegrationCapabilities[0].Vendor)
		assert.Equal(t, integration.Type, regUserResp.IntegrationCapabilities[0].Type)
		assert.Equal(t, integration.ID, regUserResp.IntegrationCapabilities[0].ID)
		createdUserDeviceID = regUserResp.UserDeviceID
	})
	t.Run("POST - register with MMY, twice don't duplicate definition", func(t *testing.T) {
		mk := "Tesla"
		model := "Model Z"
		year := 2021
		reg := RegisterUserDevice{
			Make:  &mk,
			Model: &model,
			Year:  &year,
		}
		j, _ := json.Marshal(reg)
		request := buildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		regUserResp := RegisterUserDeviceResponse{}
		_ = json.Unmarshal(body, &regUserResp)
		assert.Len(t, regUserResp.UserDeviceID, 27)
		assert.Len(t, regUserResp.DeviceDefinitionID, 27)

		// second pass, assert get same device_definition_id
		request = buildRequest("POST", "/user/devices/second", string(j))
		response, _ = app.Test(request)
		body, _ = ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		regUserResp2 := RegisterUserDeviceResponse{}
		_ = json.Unmarshal(body, &regUserResp2)
		assert.Equal(t, regUserResp.DeviceDefinitionID, regUserResp2.DeviceDefinitionID)
		assert.NotEqual(t, regUserResp.UserDeviceID, regUserResp2.UserDeviceID)
	})
	t.Run("POST - bad payload", func(t *testing.T) {
		request := buildRequest("POST", "/user/devices", "{}")
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
		msg := gjson.Get(string(body), "error_message").String()
		assert.Contains(t, msg, "cannot be blank")
	})
	t.Run("POST - bad device_definition_id", func(t *testing.T) {
		ddID := "caca"
		reg := RegisterUserDevice{
			DeviceDefinitionID: &ddID,
		}
		j, _ := json.Marshal(reg)
		request := buildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
		msg := gjson.Get(string(body), "error_message").String()
		fmt.Println("message: " + msg)
		assert.Contains(t, msg, "caca")
	})
	t.Run("GET - user devices", func(t *testing.T) {
		request := buildRequest("GET", "/user/devices/me", "")
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)

		assert.Equal(t, fiber.StatusOK, response.StatusCode)

		result := gjson.Get(string(body), "user_devices.#.id")
		fmt.Println(string(body))
		assert.Len(t, result.Array(), 2)
		for _, id := range result.Array() {
			assert.True(t, id.Exists(), "expected to find the ID")
		}
	})
	t.Run("POST - admin register with MMY", func(t *testing.T) {
		payload := `{
  "country_code": "USA",
  "created_date": 1634835455,
  "device_definition_id": null,
  "image_url": null,
  "make": "HYUNDAI",
  "model": "KONA ELECTRIC",
  "vehicle_name": "Test Name",
  "verified": false,
  "vin": null,
  "year": 2020
}`
		request := buildRequest("POST", "/admin/user/1234/devices", payload)
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		udi := gjson.Get(string(body), "user_device_id")
		fmt.Println("MMY user_device_id created: " + udi.String())
		assert.True(t, udi.Exists(), "expected to find user_device_id")
	})
	t.Run("PATCH - update VIN", func(t *testing.T) {
		payload := `{ "vin": "5YJYGDEE5MF085533" }`
		request := buildRequest("PATCH", "/user/devices/"+createdUserDeviceID+"/vin", payload)
		response, _ := app.Test(request)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode) == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("message: " + string(body))
		}
	})
	t.Run("PATCH - update Name", func(t *testing.T) {
		payload := `{ "name": "Queens Charriot" }`
		request := buildRequest("PATCH", "/user/devices/"+createdUserDeviceID+"/name", payload)
		response, _ := app.Test(request)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode) == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("message: " + string(body))
		}
	})
}
