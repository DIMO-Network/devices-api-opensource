package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type fakeEventService struct{}

func (f *fakeEventService) Emit(event *services.Event) error {
	fmt.Printf("Emitting %v\n", event)
	return nil
}

type fakeEncrypter struct{}

func (e *fakeEncrypter) Encrypt(s string) (string, error) {
	return "SECRETLOL" + s, nil
}

func TestUserDevicesController(t *testing.T) {
	// arrange global db and route setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ctx := context.Background()
	pdb := test.GetDBConnection(ctx)

	deviceDefSvc := mock_services.NewMockIDeviceDefinitionService(mockCtrl)
	taskSvc := mock_services.NewMockITaskService(mockCtrl)
	scClient := mock_services.NewMockSmartcarClient(mockCtrl)
	teslaSvc := mock_services.NewMockTeslaService(mockCtrl)
	teslaTaskService := mock_services.NewMockTeslaTaskService(mockCtrl)
	nhtsaService := mock_services.NewMockINHTSAService(mockCtrl)

	testUserID := "123123"
	testUserID2 := "3232451"
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, deviceDefSvc, taskSvc, &fakeEventService{}, scClient, teslaSvc, teslaTaskService, &fakeEncrypter{}, nil, nhtsaService)
	app := fiber.New()
	app.Post("/user/devices", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceForUser)
	app.Post("/user/devices/second", test.AuthInjectorTestHandler(testUserID2), c.RegisterDeviceForUser) // for different test user
	app.Get("/user/devices/me", test.AuthInjectorTestHandler(testUserID), c.GetUserDevices)
	app.Patch("/user/devices/:userDeviceID/vin", test.AuthInjectorTestHandler(testUserID), c.UpdateVIN)
	app.Patch("/user/devices/:userDeviceID/name", test.AuthInjectorTestHandler(testUserID), c.UpdateName)
	app.Post("/user/devices/:userDeviceID/commands/refresh", test.AuthInjectorTestHandler(testUserID), c.RefreshUserDeviceStatus)

	deviceDefSvc.EXPECT().CheckAndSetImage(gomock.Any(), false).AnyTimes().Return(nil)

	t.Run("POST - register with existing device_definition_id", func(t *testing.T) {
		// arrange DB
		dm := test.SetupCreateMake(t, "Testla", pdb)
		integration := test.SetupCreateSmartCarIntegration(t, pdb)

		dd := models.DeviceDefinition{
			ID:           ksuid.New().String(),
			DeviceMakeID: dm.ID,
			Model:        "Model X",
			Year:         2020,
			Verified:     true,
		}
		err := dd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err, "database error")

		deviceInt := models.DeviceIntegration{
			DeviceDefinitionID: dd.ID,
			IntegrationID:      integration.ID,
			Region:             "Americas",
		}
		err = deviceInt.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err, "database error")
		// act request
		reg := RegisterUserDevice{
			DeviceDefinitionID: &dd.ID,
			CountryCode:        "USA",
		}
		j, _ := json.Marshal(reg)
		request := test.BuildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		regUserResp := UserDeviceFull{}
		jsonUD := gjson.Get(string(body), "userDevice")
		_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

		assert.Len(t, regUserResp.ID, 27)
		assert.Len(t, regUserResp.DeviceDefinition.DeviceDefinitionID, 27)
		assert.Equal(t, dd.ID, regUserResp.DeviceDefinition.DeviceDefinitionID)
		if assert.Len(t, regUserResp.DeviceDefinition.CompatibleIntegrations, 1) == false {
			fmt.Println("resp body: " + string(body))
		}
		assert.Equal(t, integration.Vendor, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Vendor)
		assert.Equal(t, integration.Type, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Type)
		assert.Equal(t, integration.ID, regUserResp.DeviceDefinition.CompatibleIntegrations[0].ID)
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("POST - register with MMY, create definition on the fly", func(t *testing.T) {
		mk := "Tesla"
		model := "Model Z"
		year := 2021
		deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), mk, model, year, false).
			Return(nil, nil)
		// create an existing make and then mock return the make we just created. Another option would be to have mock call real, but I feel this isolates a bit more.
		dm := test.SetupCreateMake(t, mk, pdb)
		deviceDefSvc.EXPECT().GetOrCreateMake(gomock.Any(), gomock.Any(), mk).Times(1).Return(&models.DeviceMake{
			ID:   dm.ID,
			Name: dm.Name,
		}, nil)

		reg := RegisterUserDevice{
			Make:        &mk,
			Model:       &model,
			Year:        &year,
			CountryCode: "USA",
		}
		j, _ := json.Marshal(reg)
		request := test.BuildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		regUserResp := UserDeviceFull{}
		jsonUD := gjson.Get(string(body), "userDevice")
		_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

		assert.Len(t, regUserResp.ID, 27)
		assert.Len(t, regUserResp.DeviceDefinition.DeviceDefinitionID, 27)
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("POST - register with MMY when definition exists - still works and does not duplicate definition", func(t *testing.T) {
		mk := "Ford"
		model := "Mach E"
		year := 2021
		dm := test.SetupCreateMake(t, mk, pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, model, year, pdb)

		dd.R = dd.R.NewStruct()
		dd.R.DeviceMake = &dm
		deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), mk, model, year, false).
			Return(dd, nil)
		reg := RegisterUserDevice{
			Make:        &mk,
			Model:       &model,
			Year:        &year,
			CountryCode: "USA",
		}
		j, _ := json.Marshal(reg)
		request := test.BuildRequest("POST", "/user/devices/second", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		regUserResp := UserDeviceFull{}
		jsonUD := gjson.Get(string(body), "userDevice")
		_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

		assert.Len(t, regUserResp.ID, 27)
		assert.Equal(t, dd.ID, regUserResp.DeviceDefinition.DeviceDefinitionID)
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("POST - bad payload", func(t *testing.T) {
		request := test.BuildRequest("POST", "/user/devices", "{}")
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
		msg := gjson.Get(string(body), "errorMessage").String()
		assert.Contains(t, msg, "cannot be blank")
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("POST - bad device_definition_id", func(t *testing.T) {
		ddID := "caca"
		reg := RegisterUserDevice{
			DeviceDefinitionID: &ddID,
			CountryCode:        "USA",
		}
		j, _ := json.Marshal(reg)
		request := test.BuildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
		msg := gjson.Get(string(body), "errorMessage").String()
		fmt.Println("message: " + msg)
		assert.Contains(t, msg, "caca")
	})
	t.Run("GET - user devices", func(t *testing.T) {
		// arrange db, insert some user_devices
		dm := test.SetupCreateMake(t, "Ford", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Mach E", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)

		request := test.BuildRequest("GET", "/user/devices/me", "")
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)

		assert.Equal(t, fiber.StatusOK, response.StatusCode)

		result := gjson.Get(string(body), "userDevices.#.id")
		assert.Len(t, result.Array(), 1)
		for _, id := range result.Array() {
			assert.True(t, id.Exists(), "expected to find the ID")
			assert.Equal(t, ud.ID, id.String(), "expected user device ID to match")
		}
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("PATCH - update VIN", func(t *testing.T) {
		dm := test.SetupCreateMake(t, "Ford", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Mach E", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)

		evID := "4"
		nhtsaService.EXPECT().DecodeVIN("5YJYGDEE5MF085533").Return(&services.NHTSADecodeVINResponse{
			Results: []services.NHTSAResult{
				{
					VariableID: 126,
					ValueID:    &evID,
				},
			},
		}, nil)
		payload := `{ "vin": "5YJYGDEE5MF085533" }`
		request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/vin", payload)
		response, _ := app.Test(request)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode) == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("message: " + string(body))
		}
		request = test.BuildRequest("GET", "/user/devices/me", "")
		response, _ = app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(body))
		pt := gjson.GetBytes(body, "userDevices.0.metadata.powertrainType").String()
		assert.Equal(t, "BEV", pt)
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("PATCH - update Name", func(t *testing.T) {
		dm := test.SetupCreateMake(t, "Ford", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Mach E", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)

		payload := `{ "name": "Queens Charriot" }`
		request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/name", payload)
		response, _ := app.Test(request)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode) == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("message: " + string(body))
		}
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("POST - refresh smartcar data", func(t *testing.T) {
		dm := test.SetupCreateMake(t, "Ford", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Mach E", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)
		// arrange some additional data for this to work
		smartCarInt := test.SetupCreateSmartCarIntegration(t, pdb)
		udiai := models.UserDeviceAPIIntegration{
			UserDeviceID:     ud.ID,
			IntegrationID:    smartCarInt.ID,
			Status:           models.UserDeviceAPIIntegrationStatusActive,
			AccessToken:      null.StringFrom("caca-token"),
			AccessExpiresAt:  null.TimeFrom(time.Now().Add(time.Duration(10) * time.Hour)),
			RefreshToken:     null.StringFrom("caca-refresh"),
			RefreshExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(100) * time.Hour)),
			ExternalID:       null.StringFrom("caca-external-id"),
		}
		_ = udiai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		// arrange mock
		taskSvc.EXPECT().StartSmartcarRefresh(ud.ID, smartCarInt.ID).Return(nil)
		payload := `{}`
		request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
		response, _ := app.Test(request)

		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode) == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("unexpected response: " + string(body))
		}
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("POST - refresh smartcar data rate limited", func(t *testing.T) {
		integration := test.SetupCreateSmartCarIntegration(t, pdb)
		dm := test.SetupCreateMake(t, "Ford", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Mach E", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)
		test.SetupCreateDeviceIntegration(t, dd, integration, pdb)

		udiai := models.UserDeviceAPIIntegration{
			UserDeviceID:     ud.ID,
			IntegrationID:    integration.ID,
			Status:           models.UserDeviceAPIIntegrationStatusActive,
			AccessToken:      null.StringFrom("caca-token"),
			AccessExpiresAt:  null.TimeFrom(time.Now().Add(time.Duration(10) * time.Hour)),
			RefreshToken:     null.StringFrom("caca-refresh"),
			RefreshExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(100) * time.Hour)),
			ExternalID:       null.StringFrom("caca-external-id"),
		}
		_ = udiai.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		// arrange data to cause condition
		udd := models.UserDeviceDatum{
			UserDeviceID: ud.ID,
			Data:         null.JSON{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_ = udd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		payload := `{}`
		request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
		response, _ := app.Test(request)
		// todo not getting rate limited - 400 vs 429? production code broke?
		if assert.Equal(t, fiber.StatusTooManyRequests, response.StatusCode) == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("unexpected response: " + string(body))
		}
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
}
