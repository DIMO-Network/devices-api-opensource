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
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
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
	pdb, db := test.SetupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := db.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	deviceDefSvc := mock_services.NewMockIDeviceDefinitionService(mockCtrl)
	taskSvc := mock_services.NewMockITaskService(mockCtrl)
	scClient := mock_services.NewMockSmartcarClient(mockCtrl)
	teslaSvc := mock_services.NewMockTeslaService(mockCtrl)
	teslaTaskService := mock_services.NewMockTeslaTaskService(mockCtrl)

	testUserID := "123123"
	testUserID2 := "3232451"
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, deviceDefSvc, taskSvc, &fakeEventService{}, scClient, teslaSvc, teslaTaskService, &fakeEncrypter{})
	app := fiber.New()
	app.Post("/user/devices", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceForUser)
	app.Post("/user/devices/second", test.AuthInjectorTestHandler(testUserID2), c.RegisterDeviceForUser) // for different test user
	app.Get("/user/devices/me", test.AuthInjectorTestHandler(testUserID), c.GetUserDevices)
	app.Patch("/user/devices/:userDeviceID/vin", test.AuthInjectorTestHandler(testUserID), c.UpdateVIN)
	app.Patch("/user/devices/:userDeviceID/name", test.AuthInjectorTestHandler(testUserID), c.UpdateName)
	app.Post("/user/devices/:userDeviceID/commands/refresh", test.AuthInjectorTestHandler(testUserID), c.RefreshUserDeviceStatus)
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)

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

		payload := `{ "vin": "5YJYGDEE5MF085533" }`
		request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/vin", payload)
		response, _ := app.Test(request)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode) == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("message: " + string(body))
		}
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
			AccessToken:      "caca-token",
			AccessExpiresAt:  time.Now().Add(time.Duration(10) * time.Hour),
			RefreshToken:     "caca-refresh",
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
		test.SetupCreateDeviceIntegration(t, dd, integration, ud, pdb)

		udiai := models.UserDeviceAPIIntegration{
			UserDeviceID:     ud.ID,
			IntegrationID:    integration.ID,
			Status:           models.UserDeviceAPIIntegrationStatusActive,
			AccessToken:      "caca-token",
			AccessExpiresAt:  time.Now().Add(time.Duration(10) * time.Hour),
			RefreshToken:     "caca-refresh",
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

	t.Run("POST - Smartcar integration failure", func(t *testing.T) {
		integration := test.SetupCreateSmartCarIntegration(t, pdb)
		dm := test.SetupCreateMake(t, "Ford", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Mach E", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)
		test.SetupCreateDeviceIntegration(t, dd, integration, ud, pdb)

		req := `{
			"code": "qxyz",
			"redirectURI": "http://dimo.zone/cb"
		}`

		scClient.EXPECT().ExchangeCode(gomock.Any(), "qxyz", "http://dimo.zone/cb").Times(1).Return(nil, errors.New("failure communicating with Smartcar"))
		request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
		response, _ := app.Test(request)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode, "should return bad request when given incorrect authorization code")
		exists, _ := models.UserDeviceAPIIntegrationExists(ctx, pdb.DBS().Writer, ud.ID, integration.ID)
		assert.False(t, exists, "no integration should have been created")
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})

	t.Run("POST - Smartcar integration success", func(t *testing.T) {
		integration := test.SetupCreateSmartCarIntegration(t, pdb)
		dm := test.SetupCreateMake(t, "Ford", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Mach E", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)
		test.SetupCreateDeviceIntegration(t, dd, integration, ud, pdb)
		req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
		expiry, _ := time.Parse(time.RFC3339, "2022-03-01T12:00:00Z")
		scClient.EXPECT().ExchangeCode(gomock.Any(), "qxy", "http://dimo.zone/cb").Times(1).Return(&smartcar.Token{
			Access:        "myAccess",
			AccessExpiry:  expiry,
			Refresh:       "myRefresh",
			RefreshExpiry: expiry.Add(24 * time.Hour),
		}, nil)

		taskSvc.EXPECT().StartSmartcarRegistrationTasks(ud.ID, integration.ID).Times(1).Return(nil)
		request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
		response, _ := app.Test(request)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode, "should return success") == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("unexpected response: " + string(body))
		}
		apiInt, _ := models.FindUserDeviceAPIIntegration(ctx, pdb.DBS().Writer, ud.ID, integration.ID)

		assert.Equal(t, "myAccess", apiInt.AccessToken)
		assert.True(t, expiry.Equal(apiInt.AccessExpiresAt))
		assert.Equal(t, "Pending", apiInt.Status)
		assert.Equal(t, "myRefresh", apiInt.RefreshToken)
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})

	t.Run("POST - integration for unknown device", func(t *testing.T) {
		integration := test.SetupCreateSmartCarIntegration(t, pdb)
		req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
		request := test.BuildRequest("POST", "/user/devices/fakeDevice/integrations/"+integration.ID, req)
		response, _ := app.Test(request)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode, "should fail")
	})

	t.Run("POST - register Tesla integration", func(t *testing.T) {
		dm := test.SetupCreateMake(t, "Tesla", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Model Y", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)
		teslaInt := models.Integration{
			ID:     ksuid.New().String(),
			Type:   models.IntegrationTypeAPI,
			Style:  models.IntegrationStyleOEM,
			Vendor: "Tesla",
		}
		_ = teslaInt.Insert(ctx, pdb.DBS().Writer, boil.Infer())

		di := models.DeviceIntegration{
			DeviceDefinitionID: dd.ID,
			IntegrationID:      teslaInt.ID,
			Region:             "Americas",
		}
		_ = di.Insert(ctx, pdb.DBS().Writer, boil.Infer())

		req := `{
			"accessToken": "abc",
			"externalId": "1145",
			"expiresIn": 600,
			"refreshToken": "fffg"
		}`
		request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+teslaInt.ID, req)

		oV := &services.TeslaVehicle{}
		oUdai := &models.UserDeviceAPIIntegration{}

		teslaTaskService.EXPECT().StartPoll(gomock.AssignableToTypeOf(oV), gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
			func(v *services.TeslaVehicle, udai *models.UserDeviceAPIIntegration) error {
				oV = v
				oUdai = udai
				return nil
			},
		)

		teslaSvc.EXPECT().GetVehicle("abc", 1145).Return(&services.TeslaVehicle{
			ID:        1145,
			VehicleID: 223,
			VIN:       "5YJYGDEF9NF010423",
		}, nil)
		teslaSvc.EXPECT().WakeUpVehicle("abc", 1145).Return(nil)
		expectedExpiry := time.Now().Add(10 * time.Minute)
		response, _ := app.Test(request)
		assert.Equal(t, fiber.StatusNoContent, response.StatusCode, "should return success")

		assert.Equal(t, 1145, oV.ID)
		assert.Equal(t, 223, oV.VehicleID)

		within := func(test, reference *time.Time, d time.Duration) bool {
			return test.After(reference.Add(-d)) && test.Before(reference.Add(d))
		}

		apiInt, err := models.FindUserDeviceAPIIntegration(ctx, pdb.DBS().Writer, ud.ID, teslaInt.ID)
		if err != nil {
			t.Fatalf("Couldn't find API integration record: %v", err)
		}
		assert.Equal(t, "SECRETLOLabc", apiInt.AccessToken)
		assert.Equal(t, "1145", apiInt.ExternalID.String)
		assert.Equal(t, "SECRETLOLfffg", apiInt.RefreshToken)
		assert.True(t, within(&apiInt.AccessExpiresAt, &expectedExpiry, 15*time.Second), "access token expires at %s, expected something close to %s", apiInt.AccessExpiresAt, expectedExpiry)
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})

	t.Run("POST - register Tesla integration, update device definition", func(t *testing.T) {
		dm := test.SetupCreateMake(t, "Tesla", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Model Y", 2022, pdb)
		dd.R = dd.R.NewStruct()
		dd.R.DeviceMake = &dm

		dd2 := test.SetupCreateDeviceDefinition(t, dm, "Roadster", 2010, pdb)

		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)
		ud.R = ud.R.NewStruct()
		ud.R.DeviceDefinition = dd

		teslaInt := models.Integration{
			ID:     ksuid.New().String(),
			Type:   models.IntegrationTypeAPI,
			Style:  models.IntegrationStyleOEM,
			Vendor: "Tesla",
		}
		_ = teslaInt.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		di := models.DeviceIntegration{
			DeviceDefinitionID: dd.ID,
			IntegrationID:      teslaInt.ID,
			Region:             "Americas",
		}
		_ = di.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		di2 := models.DeviceIntegration{
			DeviceDefinitionID: dd2.ID,
			IntegrationID:      teslaInt.ID,
			Region:             "Americas",
		}
		dd.R.DeviceIntegrations = []*models.DeviceIntegration{&di, &di2}

		_ = di2.Insert(ctx, pdb.DBS().Writer, boil.Infer())

		err := c.fixTeslaDeviceDefinition(ctx, &logger, pdb.DBS().Writer.DB, &teslaInt, &ud, "5YJRE1A31A1P01234")
		if err != nil {
			t.Fatalf("Got an error while fixing device definition: %v", err)
		}

		_ = ud.Reload(ctx, pdb.DBS().Writer.DB)
		if ud.DeviceDefinitionID != dd2.ID {
			t.Fatalf("Failed to switch device definition to the correct one")
		}
	})
}
