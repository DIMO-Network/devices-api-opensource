package controllers

import (
	"context"
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
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestUserIntegrationsController(t *testing.T) {
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
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(mockCtrl)

	const (
		region     = "North America"
		testUserID = "123123"
	)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, deviceDefSvc, taskSvc, &fakeEventService{}, scClient, teslaSvc, teslaTaskService, &fakeEncrypter{}, autopiAPISvc)
	app := fiber.New()
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	app.Get("/integrations", c.GetIntegrations)

	t.Run("get integrations from db", func(t *testing.T) {
		autoPiInteg := test.SetupCreateAutoPiIntegration(t, 34, pdb)
		scInteg := test.SetupCreateSmartCarIntegration(t, pdb)

		request := test.BuildRequest("GET", "/integrations", "")
		response, err := app.Test(request)
		assert.NoError(t, err)
		body, _ := ioutil.ReadAll(response.Body)

		assert.Equal(t, fiber.StatusOK, response.StatusCode)

		jsonIntegrations := gjson.GetBytes(body, "integrations")
		assert.True(t, jsonIntegrations.IsArray())
		assert.Equal(t, gjson.GetBytes(body, "integrations.0.id").Str, autoPiInteg.ID)
		assert.Equal(t, gjson.GetBytes(body, "integrations.1.id").Str, scInteg.ID)

		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})

	t.Run("create device integration for autopi with nothing in db returns nil, nil", func(t *testing.T) {
		di, err := createDeviceIntegrationIfAutoPi(ctx, "123", "123", region, pdb.DBS().Writer)

		assert.NoError(t, err)
		assert.Nil(t, di, "expected device integration to be nil")

		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})
	t.Run("create device Integration with existing autopi integration returns new device_integration, and .R.Integration", func(t *testing.T) {
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

		assert.Equal(t, "myAccess", apiInt.AccessToken.String)
		assert.True(t, expiry.Equal(apiInt.AccessExpiresAt.Time))
		assert.Equal(t, "Pending", apiInt.Status)
		assert.Equal(t, "myRefresh", apiInt.RefreshToken.String)
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
		assert.Equal(t, "SECRETLOLabc", apiInt.AccessToken.String)
		assert.Equal(t, "1145", apiInt.ExternalID.String)
		assert.Equal(t, "SECRETLOLfffg", apiInt.RefreshToken.String)
		assert.True(t, within(&apiInt.AccessExpiresAt.Time, &expectedExpiry, 15*time.Second), "access token expires at %s, expected something close to %s", apiInt.AccessExpiresAt, expectedExpiry)
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

	t.Run("POST - AutoPi integration success", func(t *testing.T) {
		integration := test.SetupCreateAutoPiIntegration(t, 34, pdb)
		dm := test.SetupCreateMake(t, "Testla", pdb)
		dd := test.SetupCreateDeviceDefinition(t, dm, "Model 4", 2022, pdb)
		ud := test.SetupCreateUserDevice(t, testUserID, dd, pdb)
		const jobID = "123"
		const deviceID = "device123"
		const unitID = "qxyautopi"
		const vehicleID = "veh123"

		req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)
		// setup all autoPi mock expected calls.
		autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
			ID:       deviceID, // device id
			UnitID:   unitID,
			Vehicle:  services.AutoPiDongleVehicle{ID: vehicleID}, // vehicle profile id
			IMEI:     "IMEI321",
			Template: 1,
		}, nil)
		autopiAPISvc.EXPECT().PatchVehicleProfile(vehicleID, gomock.Any()).Times(1).Return(nil)
		autopiAPISvc.EXPECT().UnassociateDeviceTemplate(deviceID, 1).Times(1).Return(nil)
		autopiAPISvc.EXPECT().AssociateDeviceToTemplate(deviceID, 34).Times(1).Return(nil)
		autopiAPISvc.EXPECT().ApplyTemplate(deviceID, 34).Times(1).Return(nil)
		autopiAPISvc.EXPECT().CommandSyncDevice(deviceID).Times(1).Return(&services.AutoPiCommandResponse{
			Jid: jobID,
		}, nil)

		request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
		response, _ := app.Test(request)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode, "should return success") == false {
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("unexpected response: " + string(body) + "\n")
			fmt.Println("body sent to post: " + req)
		}

		apiInt, err := models.FindUserDeviceAPIIntegration(ctx, pdb.DBS().Writer, ud.ID, integration.ID)
		assert.NoError(t, err)
		fmt.Printf("found user device api int: %+v", *apiInt)

		metadata := new(services.UserDeviceAPIIntegrationsMetadata)
		err = apiInt.Metadata.Unmarshal(metadata)
		assert.NoError(t, err)

		assert.Equal(t, jobID, *metadata.AutoPiSyncCommandJobID)
		assert.Equal(t, deviceID, apiInt.ExternalID.String)
		assert.Equal(t, "PendingFirstData", apiInt.Status)
		//teardown
		test.TruncateTables(pdb.DBS().Writer.DB, t)
	})

}
