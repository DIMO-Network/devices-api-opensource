package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type fakeEventService struct{}

func (f *fakeEventService) Emit(event *services.Event) error {
	fmt.Printf("Emitting %v\n", event)
	return nil
}

type UserDevicesControllerTestSuite struct {
	suite.Suite
	pdb          database.DbStore
	container    testcontainers.Container
	ctx          context.Context
	mockCtrl     *gomock.Controller
	app          *fiber.App
	deviceDefSvc *mock_services.MockIDeviceDefinitionService
	testUserID   string
	taskSvc      *mock_services.MockITaskService
	nhtsaService *mock_services.MockINHTSAService
}

// SetupSuite starts container db
func (s *UserDevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl

	s.deviceDefSvc = mock_services.NewMockIDeviceDefinitionService(mockCtrl)
	s.taskSvc = mock_services.NewMockITaskService(mockCtrl)
	scClient := mock_services.NewMockSmartcarClient(mockCtrl)
	scTaskSvc := mock_services.NewMockSmartcarTaskService(mockCtrl)
	teslaSvc := mock_services.NewMockTeslaService(mockCtrl)
	teslaTaskService := mock_services.NewMockTeslaTaskService(mockCtrl)
	s.nhtsaService = mock_services.NewMockINHTSAService(mockCtrl)
	autoPiIngest := mock_services.NewMockIngestRegistrar(mockCtrl)
	autoPiTaskSvc := mock_services.NewMockAutoPiTaskService(mockCtrl)

	s.testUserID = "123123"
	testUserID2 := "3232451"
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.taskSvc,
		&fakeEventService{}, scClient, scTaskSvc, teslaSvc, teslaTaskService, nil, nil,
		s.nhtsaService, autoPiIngest, autoPiTaskSvc)
	app := fiber.New()
	app.Post("/user/devices", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUser)
	app.Post("/user/devices/second", test.AuthInjectorTestHandler(testUserID2), c.RegisterDeviceForUser) // for different test user
	app.Get("/user/devices/me", test.AuthInjectorTestHandler(s.testUserID), c.GetUserDevices)
	app.Patch("/user/devices/:userDeviceID/vin", test.AuthInjectorTestHandler(s.testUserID), c.UpdateVIN)
	app.Patch("/user/devices/:userDeviceID/name", test.AuthInjectorTestHandler(s.testUserID), c.UpdateName)
	app.Post("/user/devices/:userDeviceID/commands/refresh", test.AuthInjectorTestHandler(s.testUserID), c.RefreshUserDeviceStatus)

	s.deviceDefSvc.EXPECT().CheckAndSetImage(gomock.Any(), false).AnyTimes().Return(nil) // todo move to each test where used
	s.app = app
}

//TearDownTest after each test truncate tables
func (s *UserDevicesControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

//TearDownSuite cleanup at end by terminating container
func (s *UserDevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

//Test Runner
func TestUserDevicesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(UserDevicesControllerTestSuite))
}

/* Actual Tests */
func (s *UserDevicesControllerTestSuite) TestPostWithExistingDefinitionID() {
	// arrange DB
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	integration := test.SetupCreateSmartCarIntegration(s.T(), s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model X", 2020, s.pdb)
	test.SetupCreateDeviceIntegration(s.T(), dd, integration, s.pdb)

	// act request
	reg := RegisterUserDevice{
		DeviceDefinitionID: &dd.ID,
		CountryCode:        "USA",
	}
	j, _ := json.Marshal(reg)
	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, _ := s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Len(s.T(), regUserResp.DeviceDefinition.DeviceDefinitionID, 27)
	assert.Equal(s.T(), dd.ID, regUserResp.DeviceDefinition.DeviceDefinitionID)
	if assert.Len(s.T(), regUserResp.DeviceDefinition.CompatibleIntegrations, 1) == false {
		fmt.Println("resp body: " + string(body))
	}
	assert.Equal(s.T(), integration.Vendor, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Vendor)
	assert.Equal(s.T(), integration.Type, regUserResp.DeviceDefinition.CompatibleIntegrations[0].Type)
	assert.Equal(s.T(), integration.ID, regUserResp.DeviceDefinition.CompatibleIntegrations[0].ID)
}

func (s *UserDevicesControllerTestSuite) TestPostWithMMYOnTheFlyCreateDD() {
	mk := "Tesla"
	model := "Model Z"
	year := 2021
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), mk, model, year, false).
		Return(nil, nil)
	// create an existing make and then mock return the make we just created. Another option would be to have mock call real, but I feel this isolates a bit more.
	dm := test.SetupCreateMake(s.T(), mk, s.pdb)
	s.deviceDefSvc.EXPECT().GetOrCreateMake(gomock.Any(), gomock.Any(), mk).Times(1).Return(&models.DeviceMake{
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
	response, _ := s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Len(s.T(), regUserResp.DeviceDefinition.DeviceDefinitionID, 27)
}

func (s *UserDevicesControllerTestSuite) TestPostWithMMYExistingDD() {
	mk := "Tesla"
	model := "Model Z"
	year := 2021
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), mk, model, year, false).
		Return(nil, nil)
	// create an existing make and then mock return the make we just created. Another option would be to have mock call real, but I feel this isolates a bit more.
	dm := test.SetupCreateMake(s.T(), mk, s.pdb)
	s.deviceDefSvc.EXPECT().GetOrCreateMake(gomock.Any(), gomock.Any(), mk).Times(1).Return(&models.DeviceMake{
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
	response, _ := s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Len(s.T(), regUserResp.DeviceDefinition.DeviceDefinitionID, 27)
}

func (s *UserDevicesControllerTestSuite) TestPostWithMMYDoesNotDuplicateDD() {
	mk := "Ford"
	model := "Mach E"
	year := 2021
	dm := test.SetupCreateMake(s.T(), mk, s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, model, year, s.pdb)

	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &dm
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), mk, model, year, false).
		Return(dd, nil)
	reg := RegisterUserDevice{
		Make:        &mk,
		Model:       &model,
		Year:        &year,
		CountryCode: "USA",
	}
	j, _ := json.Marshal(reg)
	request := test.BuildRequest("POST", "/user/devices/second", string(j))
	response, _ := s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	// assert
	if assert.Equal(s.T(), fiber.StatusCreated, response.StatusCode) == false {
		fmt.Println("message: " + string(body))
	}
	regUserResp := UserDeviceFull{}
	jsonUD := gjson.Get(string(body), "userDevice")
	_ = json.Unmarshal([]byte(jsonUD.String()), &regUserResp)

	assert.Len(s.T(), regUserResp.ID, 27)
	assert.Equal(s.T(), dd.ID, regUserResp.DeviceDefinition.DeviceDefinitionID)
}

func (s *UserDevicesControllerTestSuite) TestPostBadPayload() {
	request := test.BuildRequest("POST", "/user/devices", "{}")
	response, _ := s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	msg := gjson.Get(string(body), "errorMessage").String()
	assert.Contains(s.T(), msg, "cannot be blank")
}

func (s *UserDevicesControllerTestSuite) TestPostInvalidDefinitionID() {
	ddID := "caca"
	reg := RegisterUserDevice{
		DeviceDefinitionID: &ddID,
		CountryCode:        "USA",
	}
	j, _ := json.Marshal(reg)
	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, _ := s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	msg := gjson.Get(string(body), "errorMessage").String()
	fmt.Println("message: " + msg)
	assert.Contains(s.T(), msg, "caca")
}

func (s *UserDevicesControllerTestSuite) TestGetMyUserDevices() {
	// arrange db, insert some user_devices
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)
	integ := test.SetupCreateAutoPiIntegration(s.T(), 10, nil, s.pdb)
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), "123", "device123", ud.ID, integ.ID, s.pdb)

	request := test.BuildRequest("GET", "/user/devices/me", "")
	response, _ := s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	result := gjson.Get(string(body), "userDevices.#.id")
	assert.Len(s.T(), result.Array(), 1)
	for _, id := range result.Array() {
		assert.True(s.T(), id.Exists(), "expected to find the ID")
		assert.Equal(s.T(), ud.ID, id.String(), "expected user device ID to match")
	}
	assert.Equal(s.T(), integ.ID, gjson.GetBytes(body, "userDevices.0.integrations.0.integrationId").String())
	assert.Equal(s.T(), "device123", gjson.GetBytes(body, "userDevices.0.integrations.0.externalId").String())
	assert.Equal(s.T(), integ.Vendor, gjson.GetBytes(body, "userDevices.0.integrations.0.integrationVendor").String())
}

func (s *UserDevicesControllerTestSuite) TestPatchVIN() {
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)

	evID := "4"
	s.nhtsaService.EXPECT().DecodeVIN("5YJYGDEE5MF085533").Return(&services.NHTSADecodeVINResponse{
		Results: []services.NHTSAResult{
			{
				VariableID: 126,
				ValueID:    &evID,
			},
		},
	}, nil)
	payload := `{ "vin": "5YJYGDEE5MF085533" }`
	request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/vin", payload)
	response, _ := s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}
	request = test.BuildRequest("GET", "/user/devices/me", "")
	response, _ = s.app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	pt := gjson.GetBytes(body, "userDevices.0.metadata.powertrainType").String()
	assert.Equal(s.T(), "BEV", pt)
}

func (s *UserDevicesControllerTestSuite) TestPatchName() {
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)

	payload := `{ "name": "Queens Charriot" }`
	request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/name", payload)
	response, _ := s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}
}

func (s *UserDevicesControllerTestSuite) TestPostRefreshSmartCar() {
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)
	// arrange some additional data for this to work
	smartCarInt := test.SetupCreateSmartCarIntegration(s.T(), s.pdb)
	udiai := models.UserDeviceAPIIntegration{
		UserDeviceID:    ud.ID,
		IntegrationID:   smartCarInt.ID,
		Status:          models.UserDeviceAPIIntegrationStatusActive,
		AccessToken:     null.StringFrom("caca-token"),
		AccessExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(10) * time.Hour)),
		RefreshToken:    null.StringFrom("caca-refresh"),
		ExternalID:      null.StringFrom("caca-external-id"),
	}
	err := udiai.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	udd := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Data:          null.JSONFrom([]byte(`{"odometer": 123.223}`)),
		IntegrationID: smartCarInt.ID,
		CreatedAt:     time.Now().UTC().Add(time.Hour * -4),
		UpdatedAt:     time.Now().UTC().Add(time.Hour * -4),
	}
	err = udd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	// arrange mock
	s.taskSvc.EXPECT().StartSmartcarRefresh(ud.ID, smartCarInt.ID).Return(nil)
	payload := `{}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
	response, _ := s.app.Test(request)

	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body))
	}
}

func (s *UserDevicesControllerTestSuite) TestPostRefreshSmartCarRateLimited() {
	integration := test.SetupCreateSmartCarIntegration(s.T(), s.pdb)
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)
	test.SetupCreateDeviceIntegration(s.T(), dd, integration, s.pdb)

	udiai := models.UserDeviceAPIIntegration{
		UserDeviceID:    ud.ID,
		IntegrationID:   integration.ID,
		Status:          models.UserDeviceAPIIntegrationStatusActive,
		AccessToken:     null.StringFrom("caca-token"),
		AccessExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(10) * time.Hour)),
		RefreshToken:    null.StringFrom("caca-refresh"),
		ExternalID:      null.StringFrom("caca-external-id"),
	}
	err := udiai.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	// arrange data to cause condition
	udd := models.UserDeviceDatum{
		UserDeviceID:  ud.ID,
		Data:          null.JSON{},
		IntegrationID: integration.ID,
	}
	err = udd.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	payload := `{}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
	response, _ := s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusTooManyRequests, response.StatusCode) == false {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body))
	}
}
