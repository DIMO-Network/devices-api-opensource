package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/segmentio/ksuid"
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
	pdb              database.DbStore
	container        testcontainers.Container
	ctx              context.Context
	mockCtrl         *gomock.Controller
	app              *fiber.App
	deviceDefSvc     *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc  *mock_services.MockDeviceDefinitionIntegrationService
	testUserID       string
	scTaskSvc        *mock_services.MockSmartcarTaskService
	nhtsaService     *mock_services.MockINHTSAService
	drivlyTaskSvc    *mock_services.MockDrivlyTaskService
	blackbookTaskSvc *mock_services.MockBlackbookTaskService
}

// SetupSuite starts container db
func (s *UserDevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	logger := test.Logger()
	mockCtrl := gomock.NewController(s.T())
	s.mockCtrl = mockCtrl

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(mockCtrl)
	scClient := mock_services.NewMockSmartcarClient(mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(mockCtrl)
	teslaSvc := mock_services.NewMockTeslaService(mockCtrl)
	teslaTaskService := mock_services.NewMockTeslaTaskService(mockCtrl)
	s.nhtsaService = mock_services.NewMockINHTSAService(mockCtrl)
	autoPiIngest := mock_services.NewMockIngestRegistrar(mockCtrl)
	deviceDefinitionIngest := mock_services.NewMockDeviceDefinitionRegistrar(mockCtrl)
	autoPiTaskSvc := mock_services.NewMockAutoPiTaskService(mockCtrl)

	s.testUserID = "123123"
	testUserID2 := "3232451"
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, scClient, s.scTaskSvc, teslaSvc, teslaTaskService, nil, nil,
		s.nhtsaService, autoPiIngest, deviceDefinitionIngest, autoPiTaskSvc, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Post("/user/devices", test.AuthInjectorTestHandler(s.testUserID), c.RegisterDeviceForUser)
	app.Post("/user/devices/second", test.AuthInjectorTestHandler(testUserID2), c.RegisterDeviceForUser) // for different test user
	app.Get("/user/devices/me", test.AuthInjectorTestHandler(s.testUserID), c.GetUserDevices)
	app.Patch("/user/devices/:userDeviceID/vin", test.AuthInjectorTestHandler(s.testUserID), c.UpdateVIN)
	app.Patch("/user/devices/:userDeviceID/name", test.AuthInjectorTestHandler(s.testUserID), c.UpdateName)
	app.Patch("/user/devices/:userDeviceID/image", test.AuthInjectorTestHandler(s.testUserID), c.UpdateImage)
	app.Get("/user/devices/:userDeviceID/offers", test.AuthInjectorTestHandler(s.testUserID), c.GetOffers)
	app.Get("/user/devices/:userDeviceID/valuations", test.AuthInjectorTestHandler(s.testUserID), c.GetValuations)
	app.Get("/user/devices/:userDeviceID/range", test.AuthInjectorTestHandler(s.testUserID), c.GetRange)
	app.Post("/user/devices/:userDeviceID/commands/refresh", test.AuthInjectorTestHandler(s.testUserID), c.RefreshUserDeviceStatus)

	s.deviceDefSvc.EXPECT().CheckAndSetImage(s.ctx, gomock.Any(), false).AnyTimes().Return(nil) // todo move to each test where used
	s.app = app
}

// TearDownTest after each test truncate tables
func (s *UserDevicesControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserDevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish() // might need to do mockctrl on every test, and refactor setup into one method
}

// Test Runner
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

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd.ID}).Times(1).Return(test.BuildDeviceDefinitionWithIntegrationGRPC(dd.ID, "Ford", "Ford", "Vehicle", integration.ID), nil) // todo move to each test where used
	s.deviceDefSvc.EXPECT().CheckAndSetImage(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, responseError := s.app.Test(request)
	fmt.Println(responseError)
	body, _ := io.ReadAll(response.Body)
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

	userDevice, err := models.UserDevices().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.NotNilf(s.T(), userDevice, "expected a user device in the database to exist")
	assert.Equal(s.T(), s.testUserID, userDevice.UserID)
}

func (s *UserDevicesControllerTestSuite) TestPostWithoutDefinitionID_BadRequest() {
	// act request
	reg := RegisterUserDevice{
		CountryCode: "USA",
	}
	j, _ := json.Marshal(reg)
	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	// assert
	require.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)

	errorMessage := gjson.Get(string(body), "errorMessage")
	if assert.Equal(s.T(), "deviceDefinitionId: cannot be blank.", errorMessage.String()) == false {
		fmt.Println(string(body))
	}
}

func (s *UserDevicesControllerTestSuite) TestPostBadPayload() {
	request := test.BuildRequest("POST", "/user/devices", "{}")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
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

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ddID}).Times(1).Return(test.BuildDeviceDefinitionGRPC(ddID, "Tesla", "Tesla", "Vehicle"), nil) // todo move to each test where used

	request := test.BuildRequest("POST", "/user/devices", string(j))
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	assert.Equal(s.T(), fiber.StatusInternalServerError, response.StatusCode)
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
	const (
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
		deviceID = "device123"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integ.ID, s.pdb)

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd.ID}).Times(1).Return(test.BuildDeviceDefinitionGRPC(dd.ID, "Ford", "Ford", "Vehicle"), nil) // todo move to each test where used

	request := test.BuildRequest("GET", "/user/devices/me", "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

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
	response, responseError := s.app.Test(request)
	fmt.Println(responseError)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd.ID}).Times(1).Return(test.BuildDeviceDefinitionGRPC(dd.ID, "Ford", "Ford", "Vehicle"), nil) // todo move to each test where used

	request = test.BuildRequest("GET", "/user/devices/me", "")
	response, responseError = s.app.Test(request)
	fmt.Println(responseError)
	body, _ := io.ReadAll(response.Body)
	fmt.Println(string(body))
	pt := gjson.GetBytes(body, "userDevices.0.metadata.powertrainType").String()
	assert.Equal(s.T(), "BEV", pt)
}

func (s *UserDevicesControllerTestSuite) TestVINValidate() {

	type test struct {
		vin    string
		want   bool
		reason string
	}

	tests := []test{
		{vin: "5YJYGDEE5MF085533", want: true, reason: "valid vin number"},
		{vin: "5YJYGDEE5MF08553", want: false, reason: "too short"},
		{vin: "JA4AJ3AUXKU602608", want: true, reason: "valid vin number"},
		{vin: "2T1BU4EE2DC071057", want: true, reason: "valid vin number"},
		{vin: "5YJ3E1EA1NF156661", want: true, reason: "valid vin number"},
		{vin: "7AJ3E1EB3JF110865", want: true, reason: "valid vin number"},
		{vin: "", want: false, reason: "empty vin string"},
		{vin: "7FJ3E1EB3JF1108651234", want: false, reason: "vin string too long"},
	}

	for _, tc := range tests {
		vinReq := UpdateVINReq{VIN: &tc.vin}
		err := vinReq.validate()
		if tc.want == true {
			assert.NoError(s.T(), err, tc.reason)
		} else {
			assert.Error(s.T(), err, tc.reason)
		}
	}
}

func (s *UserDevicesControllerTestSuite) TestNameValidate() {

	type test struct {
		name   string
		want   bool
		reason string
	}

	tests := []test{
		{name: "ValidNameHere", want: true, reason: "valid name"},
		{name: "MyCar2022", want: true, reason: "valid name"},
		{name: "16CharactersLong", want: true, reason: "valid name"},
		{name: "12345", want: true, reason: "valid name"},
		{name: "a", want: true, reason: "valid name"},
		{name: "เร็ว", want: true, reason: "valid name"},
		{name: "快速地", want: true, reason: "valid name"},
		{name: "швидко", want: true, reason: "valid name"},
		{name: "سريع", want: true, reason: "valid name"},
		{name: "Dimo's Fav Car", want: true, reason: "valid name"},
		{name: "My Car: 2022", want: true, reason: "valid name"},
		{name: "Car #2", want: true, reason: "valid name"},
		{name: `Sally "Speed Demon" Sedan`, want: true, reason: "valid name"},
		{name: "Valid Car Name", want: true, reason: "valid name"},
		{name: " Invalid Name", want: false, reason: "starts with space"},
		{name: "My Car!!!", want: true, reason: "valid name with !"},
		{name: "", want: false, reason: "empty name"},
		{name: "ThisNameIsTooLong--CanOnlyBe25CharactersInLength", want: false, reason: "too long"},
		{name: "no\nNewLine", want: false, reason: "no new lines allowed"},
	}

	for _, tc := range tests {
		vinReq := UpdateNameReq{Name: &tc.name}
		err := vinReq.validate()
		if tc.want {
			assert.NoError(s.T(), err, tc.reason)
		} else {
			assert.Error(s.T(), err, tc.reason)
		}
	}
}

func (s *UserDevicesControllerTestSuite) TestPatchName() {
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)
	// nil check test
	payload := `{}`
	request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/name", payload)
	response, _ := s.app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
	// name with spaces happy path test
	payload = `{ "name": " Queens Charriot,.@!$’ " }`
	request = test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/name", payload)
	response, _ = s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}
	require.NoError(s.T(), ud.Reload(s.ctx, s.pdb.DBS().Reader))
	assert.Equal(s.T(), "Queens Charriot,.@!$’", ud.Name.String)
}

func (s *UserDevicesControllerTestSuite) TestPatchImageURL() {
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)

	payload := `{ "imageUrl": "https://ipfs.com/planetary/car.jpg" }`
	request := test.BuildRequest("PATCH", "/user/devices/"+ud.ID+"/image", payload)
	response, _ := s.app.Test(request)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("message: " + string(body))
	}
}

//go:embed test_drivly_pricing_by_vin.json
var testDrivlyPricingJSON string

//go:embed test_blackbook_by_vin.json
var testBlackbookJSON string

func (s *UserDevicesControllerTestSuite) TestGetDeviceValuations() {
	// arrange db, insert some user_devices
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)
	_ = test.SetupCreateExternalVINData(s.T(), dd, &ud, map[string][]byte{
		"PricingMetadata":   []byte(testDrivlyPricingJSON),
		"BlackbookMetadata": []byte(testBlackbookJSON),
	}, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/valuations", ud.ID), "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 2, int(gjson.GetBytes(body, "valuationSets.#").Int()))
	assert.Equal(s.T(), 49957, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).mileage").Int()))
	assert.Equal(s.T(), 49040, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeIn").Int()))
	assert.Equal(s.T(), 49040, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).tradeInAverage").Int()))
	assert.Equal(s.T(), 54123, int(gjson.GetBytes(body, "valuationSets.#(vendor=drivly).retail").Int()))
	assert.Equal(s.T(), 49957, int(gjson.GetBytes(body, "valuationSets.#(vendor=blackbook).mileage").Int()))
	assert.Equal(s.T(), 42915+1225, int(gjson.GetBytes(body, "valuationSets.#(vendor=blackbook).tradeIn").Int()))
	assert.Equal(s.T(), 42915+1225, int(gjson.GetBytes(body, "valuationSets.#(vendor=blackbook).tradeInAverage").Int()))
	assert.False(s.T(), gjson.GetBytes(body, "valuationSets.#(vendor=blackbook).retail").Exists())
}

//go:embed test_drivly_offers_by_vin.json
var testDrivlyOffersJSON string

func (s *UserDevicesControllerTestSuite) TestGetDeviceOffers() {
	// arrange db, insert some user_devices
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, nil, s.pdb)
	_ = test.SetupCreateExternalVINData(s.T(), dd, &ud, map[string][]byte{
		"OfferMetadata": []byte(testDrivlyOffersJSON),
		// "PricingMetadata":   nil,
		// "BlackbookMetadata": nil,
	}, s.pdb)

	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/offers", ud.ID), "")
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 1, int(gjson.GetBytes(body, "offerSets.#").Int()))
	assert.Equal(s.T(), "drivly", gjson.GetBytes(body, "offerSets.0.source").String())
	assert.Equal(s.T(), 3, int(gjson.GetBytes(body, "offerSets.0.offers.#").Int()))
	assert.Equal(s.T(), "Error in v1/acquisition/appraisal POST",
		gjson.GetBytes(body, "offerSets.0.offers.#(vendor=vroom).error").String())
	assert.Equal(s.T(), 10123, int(gjson.GetBytes(body, "offerSets.0.offers.#(vendor=carvana).price").Int()))
	assert.Equal(s.T(), "Make[Ford],Model[Mustang Mach-E],Year[2022] is not eligible for offer.",
		gjson.GetBytes(body, "offerSets.0.offers.#(vendor=carmax).declineReason").String())
}

//go:embed test_user_device_metadata.json
var testUserDeviceMetadata []byte

func (s *UserDevicesControllerTestSuite) TestGetDeviceRange() {
	// arrange db, insert some user_devices
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)

	gddir := []*grpc.GetDeviceDefinitionItemResponse{
		{
			VehicleData: &grpc.VehicleInfo{
				MPG:                 40.0,
				MPGHighway:          38.0,
				FuelTankCapacityGal: 14.5,
			},
			Make: &grpc.GetDeviceDefinitionItemResponse_Make{
				Name: "Ford",
			},
			Name:               "F-150",
			DeviceDefinitionId: dd.ID,
		},
	}
	ud := test.SetupCreateUserDevice(s.T(), s.testUserID, dd, &testUserDeviceMetadata, s.pdb)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd.ID}).Return(gddir, nil)
	request := test.BuildRequest("GET", fmt.Sprintf("/user/devices/%s/range", ud.ID), "")
	response, err := s.app.Test(request)
	assert.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	assert.Equal(s.T(), 3, int(gjson.GetBytes(body, "rangeSets.#").Int()))
	assert.Equal(s.T(), "MPG", gjson.GetBytes(body, "rangeSets.0.rangeBasis").String())
	assert.Equal(s.T(), "miles", gjson.GetBytes(body, "rangeSets.0.rangeUnit").String())
	assert.Equal(s.T(), 391, int(gjson.GetBytes(body, "rangeSets.1.rangeDistance").Int()))
	assert.Equal(s.T(), 380, int(gjson.GetBytes(body, "rangeSets.2.rangeDistance").Int()))
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
		TaskID:          null.StringFrom(ksuid.New().String()),
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

	var oUdai *models.UserDeviceAPIIntegration

	// arrange mock
	s.scTaskSvc.EXPECT().Refresh(gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(udai *models.UserDeviceAPIIntegration) error {
			oUdai = udai
			return nil
		},
	)

	payload := `{}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/commands/refresh", payload)
	response, err := s.app.Test(request)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), ud.ID, oUdai.UserDeviceID)

	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode) == false {
		body, _ := io.ReadAll(response.Body)
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
		body, _ := io.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body))
	}
}

func TestEIP712Hash(t *testing.T) {
	td := &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintDevice": {
				{Name: "rootNode", Type: "uint256"},
				{Name: "attributes", Type: "string[]"},
				{Name: "infos", Type: "string[]"},
			},
		},
		PrimaryType: "MintDevice",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x5fbdb2315678afecb367f032d93f642f64180aa3",
		},
		Message: signer.TypedDataMessage{
			"rootNode":   math.NewHexOrDecimal256(7), // Just hardcoding this. We need a node for each make, and to keep these in sync.
			"attributes": []any{"Make", "Model", "Year"},
			"infos":      []any{"Tesla", "Model 3", "2020"},
		},
	}
	hash, err := computeTypedDataHash(td)
	if assert.NoError(t, err) {
		realHash := common.HexToHash("0x8258cd28afb13c201c07bf80c717d55ce13e226b725dd8a115ae5ab064e537da")
		assert.Equal(t, realHash, hash)
	}
}

func TestEIP712Recover(t *testing.T) {
	td := &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintDevice": {
				{Name: "rootNode", Type: "uint256"},
				{Name: "attributes", Type: "string[]"},
				{Name: "infos", Type: "string[]"},
			},
		},
		PrimaryType: "MintDevice",
		Domain: signer.TypedDataDomain{
			Name:              "DIMO",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x5fbdb2315678afecb367f032d93f642f64180aa3",
		},
		Message: signer.TypedDataMessage{
			"rootNode":   math.NewHexOrDecimal256(7), // Just hardcoding this. We need a node for each make, and to keep these in sync.
			"attributes": []any{"Make", "Model", "Year"},
			"infos":      []any{"Tesla", "Model 3", "2020"},
		},
	}
	sig := common.FromHex("0x558266d4d8cd994c9eab2dee0efeb3ee33c839e4ce77c64da544679a85bd4a864805dd1fab769e9888fdfc0ed6502f685dc43ddda1add760febd749acfcd517b1b")
	addr, err := recoverAddress(td, sig)
	if assert.NoError(t, err) {
		realAddr := common.HexToAddress("0x969602c4f39D345Cbe47E7fe0dd8F1f16f984D65")
		assert.Equal(t, realAddr, addr)
	}
}
