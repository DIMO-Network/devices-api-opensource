package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"testing"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	mock_services "github.com/DIMO-Network/devices-api/internal/services/mocks"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UserIntegrationsControllerTestSuite struct {
	suite.Suite
	pdb                       database.DbStore
	container                 testcontainers.Container
	ctx                       context.Context
	mockCtrl                  *gomock.Controller
	app                       *fiber.App
	scClient                  *mock_services.MockSmartcarClient
	scTaskSvc                 *mock_services.MockSmartcarTaskService
	teslaSvc                  *mock_services.MockTeslaService
	teslaTaskService          *mock_services.MockTeslaTaskService
	autopiAPISvc              *mock_services.MockAutoPiAPIService
	autoPiIngest              *mock_services.MockIngestRegistrar
	drivlyTaskSvc             *mock_services.MockDrivlyTaskService
	blackbookTaskSvc          *mock_services.MockBlackbookTaskService
	eventSvc                  *mock_services.MockEventService
	deviceDefinitionRegistrar *mock_services.MockDeviceDefinitionRegistrar
	deviceDefSvc              *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc           *mock_services.MockDeviceDefinitionIntegrationService
}

const testUserID = "123123"
const testUser2 = "someOtherUser2"

// SetupSuite starts container db
func (s *UserIntegrationsControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)

	s.mockCtrl = gomock.NewController(s.T())

	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	s.scClient = mock_services.NewMockSmartcarClient(s.mockCtrl)
	s.scTaskSvc = mock_services.NewMockSmartcarTaskService(s.mockCtrl)
	s.teslaSvc = mock_services.NewMockTeslaService(s.mockCtrl)
	s.teslaTaskService = mock_services.NewMockTeslaTaskService(s.mockCtrl)
	s.autopiAPISvc = mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	s.autoPiIngest = mock_services.NewMockIngestRegistrar(s.mockCtrl)
	s.deviceDefinitionRegistrar = mock_services.NewMockDeviceDefinitionRegistrar(s.mockCtrl)
	s.drivlyTaskSvc = mock_services.NewMockDrivlyTaskService(s.mockCtrl)
	s.blackbookTaskSvc = mock_services.NewMockBlackbookTaskService(s.mockCtrl)
	s.eventSvc = mock_services.NewMockEventService(s.mockCtrl)
	autoPiTaskSvc := mock_services.NewMockAutoPiTaskService(s.mockCtrl)

	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		s.eventSvc, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), s.autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, autoPiTaskSvc, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	app.Post("/user2/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2), c.RegisterDeviceIntegration)
	app.Get("/integrations", c.GetIntegrations)
	app.Post("/user/devices/:userDeviceID/autopi/command", test.AuthInjectorTestHandler(testUserID), c.SendAutoPiCommand)
	app.Get("/user/devices/:userDeviceID/autopi/command/:jobID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiCommandStatus)
	s.app = app
}

// TearDownTest after each test truncate tables
func (s *UserIntegrationsControllerTestSuite) TearDownTest() {
	test.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
}

// TearDownSuite cleanup at end by terminating container
func (s *UserIntegrationsControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

// Test Runner
func TestUserIntegrationsControllerTestSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationsControllerTestSuite))
}

/* Actual Tests */
func (s *UserIntegrationsControllerTestSuite) TestGetIntegrations() {
	autoPiInteg := test.SetupCreateAutoPiIntegration(s.T(), 34, nil, s.pdb)
	scInteg := test.SetupCreateSmartCarIntegration(s.T(), s.pdb)

	request := test.BuildRequest("GET", "/integrations", "")
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	body, _ := io.ReadAll(response.Body)

	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	jsonIntegrations := gjson.GetBytes(body, "integrations")
	assert.True(s.T(), jsonIntegrations.IsArray())
	assert.Equal(s.T(), gjson.GetBytes(body, "integrations.0.id").Str, autoPiInteg.ID)
	assert.Equal(s.T(), gjson.GetBytes(body, "integrations.1.id").Str, scInteg.ID)
}
func (s *UserIntegrationsControllerTestSuite) TestPostSmartCarFailure() {
	integration := test.SetupCreateSmartCarIntegration(s.T(), s.pdb)
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	test.SetupCreateDeviceIntegration(s.T(), dd, integration, s.pdb)

	req := `{
			"code": "qxyz",
			"redirectURI": "http://dimo.zone/cb"
		}`

	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxyz", "http://dimo.zone/cb").Times(1).Return(nil, errors.New("failure communicating with Smartcar"))

	ddIntegrations := []*ddgrpc.GetDeviceDefinitionIntegrationItemResponse{}
	deviceDefinitions := test.BuildDeviceDefinitionGRPC(ud.DeviceDefinitionID, "Ford", "Mach E", "Vehicle")

	integrationItem := &ddgrpc.GetIntegrationItemResponse{
		Id:     integration.ID,
		Vendor: integration.Vendor,
		Style:  integration.Style,
		Type:   integration.Type,
	}

	s.deviceDefIntSvc.EXPECT().GetDeviceDefinitionIntegration(gomock.Any(), gomock.Any()).Times(1).Return(ddIntegrations, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).Times(1).Return(deviceDefinitions, nil)
	s.deviceDefIntSvc.EXPECT().CreateDeviceDefinitionIntegration(gomock.Any(), integrationItem.Id, ud.DeviceDefinitionID, "Americas").Times(1).Return(integrationItem, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
	response, _ := s.app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should return bad request when given incorrect authorization code")
	exists, _ := models.UserDeviceAPIIntegrationExists(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.ID)
	assert.False(s.T(), exists, "no integration should have been created")
}
func (s *UserIntegrationsControllerTestSuite) TestPostSmartCar() {

	integration := test.SetupCreateSmartCarIntegration(s.T(), s.pdb)
	dm := test.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Mach E", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	test.SetupCreateDeviceIntegration(s.T(), dd, integration, s.pdb)
	const smartCarUserID = "smartCarUserId"
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	expiry, _ := time.Parse(time.RFC3339, "2022-03-01T12:00:00Z")
	s.scClient.EXPECT().ExchangeCode(gomock.Any(), "qxy", "http://dimo.zone/cb").Times(1).Return(&smartcar.Token{
		Access:        "myAccess",
		AccessExpiry:  expiry,
		Refresh:       "myRefresh",
		RefreshExpiry: expiry.Add(24 * time.Hour),
	}, nil)

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *services.Event) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), "Ford", data.Device.Make)
			assert.Equal(s.T(), "Mach E", data.Device.Model)
			assert.Equal(s.T(), 2022, data.Device.Year)
			assert.Equal(s.T(), "CARVIN", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), "SmartCar", data.Integration.Vendor)
			assert.Equal(s.T(), integration.ID, data.Integration.ID)
			return nil
		},
	)

	s.deviceDefinitionRegistrar.EXPECT().Register(services.DeviceDefinitionDTO{
		IntegrationID:      integration.ID,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dm.Name,
		Model:              dd.Model,
		Year:               int(dd.Year),
	}).Return(nil)

	s.scClient.EXPECT().GetUserID(gomock.Any(), "myAccess").Return(smartCarUserID, nil)
	s.scClient.EXPECT().GetExternalID(gomock.Any(), "myAccess").Return("smartcar-idx", nil)
	s.scClient.EXPECT().GetVIN(gomock.Any(), "myAccess", "smartcar-idx").Return("CARVIN", nil)
	s.scClient.EXPECT().GetEndpoints(gomock.Any(), "myAccess", "smartcar-idx").Return([]string{"/", "/vin"}, nil)
	s.scClient.EXPECT().HasDoorControl(gomock.Any(), "myAccess", "smartcar-idx").Return(false, nil)
	s.scClient.EXPECT().GetYear(gomock.Any(), "myAccess", "smartcar-idx").Return(2022, nil)
	s.drivlyTaskSvc.EXPECT().StartDrivlyUpdate(dd.ID, ud.ID, "CARVIN").Return("task-id-123", nil)
	s.blackbookTaskSvc.EXPECT().StartBlackbookUpdate(dd.ID, ud.ID, "CARVIN").Return("task-id-123", nil)

	oUdai := &models.UserDeviceAPIIntegration{}
	s.scTaskSvc.EXPECT().StartPoll(gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(udai *models.UserDeviceAPIIntegration) error {
			oUdai = udai
			return nil
		},
	)

	ddIntegrations := []*ddgrpc.GetDeviceDefinitionIntegrationItemResponse{}

	integrationItem := &ddgrpc.GetIntegrationItemResponse{
		Id:     integration.ID,
		Vendor: integration.Vendor,
		Style:  integration.Style,
		Type:   integration.Type,
	}
	deviceDefinitions := test.BuildDeviceDefinitionGRPC(ud.DeviceDefinitionID, "Ford", "Mach E", "Vehicle")
	deviceDefinitions[0].Type.Year = int32(dd.Year)

	deviceDefinitions[0].DeviceIntegrations = append(deviceDefinitions[0].DeviceIntegrations, &ddgrpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{
		Id:     integration.ID,
		Type:   models.IntegrationTypeAPI,
		Style:  models.IntegrationStyleWebhook,
		Vendor: "SmartCar",
	})

	s.deviceDefIntSvc.EXPECT().GetDeviceDefinitionIntegration(gomock.Any(), ud.DeviceDefinitionID).Times(1).Return(ddIntegrations, nil)

	s.deviceDefIntSvc.EXPECT().CreateDeviceDefinitionIntegration(gomock.Any(), integrationItem.Id, ud.DeviceDefinitionID, "Americas").Times(1).Return(integrationItem, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).AnyTimes().Return(deviceDefinitions, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	fmt.Println(response)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body))
	}
	apiInt, _ := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.ID)

	assert.Equal(s.T(), "zlNpprff", apiInt.AccessToken.String)
	assert.True(s.T(), expiry.Equal(apiInt.AccessExpiresAt.Time))
	assert.Equal(s.T(), "PendingFirstData", apiInt.Status)
	assert.Equal(s.T(), "zlErserfu", apiInt.RefreshToken.String)
}
func (s *UserIntegrationsControllerTestSuite) TestPostUnknownDevice() {
	integration := test.SetupCreateSmartCarIntegration(s.T(), s.pdb)
	req := `{
			"code": "qxy",
			"redirectURI": "http://dimo.zone/cb"
		}`
	request := test.BuildRequest("POST", "/user/devices/fakeDevice/integrations/"+integration.ID, req)
	response, _ := s.app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should fail")
}
func (s *UserIntegrationsControllerTestSuite) TestPostTesla() {
	dm := test.SetupCreateMake(s.T(), "Tesla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model Y", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	teslaInt := models.Integration{
		ID:     ksuid.New().String(),
		Type:   models.IntegrationTypeAPI,
		Style:  models.IntegrationStyleOEM,
		Vendor: "Tesla",
	}
	_ = teslaInt.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())

	di := models.DeviceIntegration{
		DeviceDefinitionID: dd.ID,
		IntegrationID:      teslaInt.ID,
		Region:             "Americas",
	}
	_ = di.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())

	req := `{
			"accessToken": "abc",
			"externalId": "1145",
			"expiresIn": 600,
			"refreshToken": "fffg"
		}`
	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+teslaInt.ID, req)

	oV := &services.TeslaVehicle{}
	oUdai := &models.UserDeviceAPIIntegration{}

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *services.Event) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), "Tesla", data.Device.Make)
			assert.Equal(s.T(), "Model Y", data.Device.Model)
			assert.Equal(s.T(), 2022, data.Device.Year)
			assert.Equal(s.T(), "5YJYGDEF9NF010423", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), "Tesla", data.Integration.Vendor)
			assert.Equal(s.T(), teslaInt.ID, data.Integration.ID)
			return nil
		},
	)

	s.deviceDefinitionRegistrar.EXPECT().Register(services.DeviceDefinitionDTO{
		IntegrationID:      teslaInt.ID,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dm.Name,
		Model:              dd.Model,
		Year:               int(dd.Year),
	}).Return(nil)

	s.teslaTaskService.EXPECT().StartPoll(gomock.AssignableToTypeOf(oV), gomock.AssignableToTypeOf(oUdai)).DoAndReturn(
		func(v *services.TeslaVehicle, udai *models.UserDeviceAPIIntegration) error {
			oV = v
			oUdai = udai
			return nil
		},
	)

	s.teslaSvc.EXPECT().GetVehicle("abc", 1145).Return(&services.TeslaVehicle{
		ID:        1145,
		VehicleID: 223,
		VIN:       "5YJYGDEF9NF010423",
	}, nil)
	s.teslaSvc.EXPECT().WakeUpVehicle("abc", 1145).Return(nil)

	ddIntegrations := []*ddgrpc.GetDeviceDefinitionIntegrationItemResponse{}

	integrationItem := &ddgrpc.GetIntegrationItemResponse{
		Id:     teslaInt.ID,
		Vendor: teslaInt.Vendor,
		Style:  teslaInt.Style,
		Type:   teslaInt.Type,
	}
	deviceDefinitions := test.BuildDeviceDefinitionGRPC(ud.DeviceDefinitionID, dm.Name, dd.Model, "Vehicle")
	deviceDefinitions[0].Type.Year = int32(dd.Year)
	deviceDefinitions[0].DeviceIntegrations = append(deviceDefinitions[0].DeviceIntegrations, &ddgrpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{
		Id:     teslaInt.ID,
		Type:   models.IntegrationTypeAPI,
		Style:  models.IntegrationStyleWebhook,
		Vendor: "SmartCar",
	})

	s.deviceDefIntSvc.EXPECT().GetDeviceDefinitionIntegration(gomock.Any(), ud.DeviceDefinitionID).Times(1).Return(ddIntegrations, nil)

	s.deviceDefIntSvc.EXPECT().CreateDeviceDefinitionIntegration(gomock.Any(), integrationItem.Id, ud.DeviceDefinitionID, "Americas").Times(1).Return(integrationItem, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).AnyTimes().Return(deviceDefinitions, nil)
	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(deviceDefinitions[0], nil)

	expectedExpiry := time.Now().Add(10 * time.Minute)
	response, err := s.app.Test(request)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success")

	assert.Equal(s.T(), 1145, oV.ID)
	assert.Equal(s.T(), 223, oV.VehicleID)

	within := func(test, reference *time.Time, d time.Duration) bool {
		return test.After(reference.Add(-d)) && test.Before(reference.Add(d))
	}

	apiInt, err := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, teslaInt.ID)
	if err != nil {
		s.T().Fatalf("Couldn't find API integration record: %v", err)
	}
	assert.Equal(s.T(), "nop", apiInt.AccessToken.String)
	assert.Equal(s.T(), "1145", apiInt.ExternalID.String)
	assert.Equal(s.T(), "ssst", apiInt.RefreshToken.String)
	assert.True(s.T(), within(&apiInt.AccessExpiresAt.Time, &expectedExpiry, 15*time.Second), "access token expires at %s, expected something close to %s", apiInt.AccessExpiresAt, expectedExpiry)

}

func (s *UserIntegrationsControllerTestSuite) TestPostTeslaAndUpdateDD() {
	dm := test.SetupCreateMake(s.T(), "Tesla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model Y", 2022, s.pdb)
	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &dm

	dd2 := test.SetupCreateDeviceDefinition(s.T(), dm, "Roadster", 2010, s.pdb)

	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	ud.R = ud.R.NewStruct()
	ud.R.DeviceDefinition = dd

	teslaInt := ddgrpc.GetIntegrationItemResponse{
		Id:     ksuid.New().String(),
		Type:   models.IntegrationTypeAPI,
		Style:  models.IntegrationStyleOEM,
		Vendor: "Tesla",
	}

	deviceDefinitions := test.BuildDeviceDefinitionGRPC(dd2.ID, "Ford", "Mach E", "Vehicle")
	deviceDefinitions[0].DeviceIntegrations = append(deviceDefinitions[0].DeviceIntegrations, &ddgrpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{
		Id:     teslaInt.Id,
		Type:   models.IntegrationTypeAPI,
		Style:  models.IntegrationStyleWebhook,
		Vendor: "SmartCar",
	})

	s.deviceDefSvc.EXPECT().FindDeviceDefinitionByMMY(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(deviceDefinitions[0], nil)

	err := fixTeslaDeviceDefinition(s.ctx, test.Logger(), s.deviceDefSvc, s.pdb.DBS().Writer.DB, &teslaInt, &ud, "5YJRE1A31A1P01234")
	if err != nil {
		s.T().Fatalf("Got an error while fixing device definition: %v", err)
	}

	_ = ud.Reload(s.ctx, s.pdb.DBS().Writer.DB)
	if ud.DeviceDefinitionID != dd2.ID {
		s.T().Fatalf("Failed to switch device definition to the correct one")
	}
}

func (s *UserIntegrationsControllerTestSuite) TestPostAutoPi_HappyPath() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		s.eventSvc, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	// arrange
	const templateID = 34
	integration := test.SetupCreateAutoPiIntegration(s.T(), templateID, nil, s.pdb)
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	const (
		jobID     = "123"
		deviceID  = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID    = "431d2e89-46f1-6884-6226-5d1ad20c84d9" // note lowercase & 36 char, as we always lowercase input bc that is what autopi uses
		vehicleID = 123
		imei      = "IMEI321"
		nftAddr   = "0x323b5d4c32345ced77393b3530b1eed0f346429d"
	)

	ud.VinIdentifier = null.StringFrom("CARVANA")
	_, _ = ud.Update(context.Background(), s.pdb.DBS().Writer, boil.Infer())

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)
	// setup all autoPi mock expected calls.
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		ID:                deviceID, // device id
		UnitID:            unitID,
		Vehicle:           services.AutoPiDongleVehicle{ID: vehicleID}, // vehicle profile id
		IMEI:              imei,
		Template:          1,
		EthereumAddress:   nftAddr,
		LastCommunication: time.Now().Add(time.Second * -15).UTC(),
	}, nil)
	autopiAPISvc.EXPECT().PatchVehicleProfile(vehicleID, gomock.Any()).Times(1).Return(nil)
	autopiAPISvc.EXPECT().UnassociateDeviceTemplate(deviceID, 1).Times(1).Return(nil)
	autopiAPISvc.EXPECT().AssociateDeviceToTemplate(deviceID, 34).Times(1).Return(nil)
	autopiAPISvc.EXPECT().ApplyTemplate(deviceID, 34).Times(1).Return(nil)
	autopiAPISvc.EXPECT().CommandSyncDevice(gomock.Any(), unitID, deviceID, ud.ID).Times(1).Return(&services.AutoPiCommandResponse{
		Jid: jobID,
	}, nil)
	s.autoPiIngest.EXPECT().Register(unitID, ud.ID, integration.ID).Return(nil)
	s.deviceDefinitionRegistrar.EXPECT().Register(services.DeviceDefinitionDTO{
		IntegrationID:      integration.ID,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dm.Name,
		Model:              dd.Model,
		Year:               int(dd.Year),
	}).Return(nil)

	s.eventSvc.EXPECT().Emit(gomock.Any()).Return(nil).Do(
		func(event *services.Event) error {
			assert.Equal(s.T(), ud.ID, event.Subject)
			assert.Equal(s.T(), "com.dimo.zone.device.integration.create", event.Type)

			data := event.Data.(services.UserDeviceIntegrationEvent)

			assert.Equal(s.T(), "Testla", data.Device.Make)
			assert.Equal(s.T(), "Model 4", data.Device.Model)
			assert.Equal(s.T(), 2022, data.Device.Year)
			assert.Equal(s.T(), "CARVANA", data.Device.VIN)
			assert.Equal(s.T(), ud.ID, data.Device.ID)

			assert.Equal(s.T(), "AutoPi", data.Integration.Vendor)
			assert.Equal(s.T(), integration.ID, data.Integration.ID)
			return nil
		},
	)

	ddIntegrations := []*ddgrpc.GetDeviceDefinitionIntegrationItemResponse{}

	integrationItem := &ddgrpc.GetIntegrationItemResponse{
		Id:                      integration.ID,
		Vendor:                  integration.Vendor,
		Style:                   integration.Style,
		Type:                    integration.Type,
		AutoPiDefaultTemplateId: templateID,
	}

	deviceDefinitions := test.BuildDeviceDefinitionGRPC(ud.DeviceDefinitionID, dm.Name, dd.Model, "Vehicle")
	deviceDefinitions[0].Type.Year = int32(dd.Year)

	s.deviceDefIntSvc.EXPECT().GetDeviceDefinitionIntegration(gomock.Any(), gomock.Any()).Times(1).Return(ddIntegrations, nil)

	s.deviceDefIntSvc.EXPECT().CreateDeviceDefinitionIntegration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(integrationItem, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).Times(1).Return(deviceDefinitions, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
	response, err := app.Test(request, 2000000)
	require.NoError(s.T(), err)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body) + "\n")
		fmt.Println("body sent to post: " + req)
	}

	apiInt, err := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.ID)
	require.NoError(s.T(), err)
	fmt.Printf("found user device api int: %+v", *apiInt)

	autoPiUnit, err := models.FindAutopiUnit(s.ctx, s.pdb.DBS().Writer, unitID)
	require.NoError(s.T(), err)

	metadata := new(services.UserDeviceAPIIntegrationsMetadata)
	err = apiInt.Metadata.Unmarshal(metadata)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), deviceID, apiInt.ExternalID.String)
	assert.Equal(s.T(), unitID, apiInt.AutopiUnitID.String)
	assert.Equal(s.T(), unitID, autoPiUnit.AutopiUnitID)
	assert.Equal(s.T(), ud.UserID, autoPiUnit.UserID.String)
	assert.Equal(s.T(), deviceID, autoPiUnit.AutopiDeviceID.String)
	assert.Equal(s.T(), nftAddr, "0x"+common.Bytes2Hex(autoPiUnit.EthereumAddress.Bytes))
	assert.Equal(s.T(), "Pending", apiInt.Status)
	assert.Equal(s.T(), templateID, *metadata.AutoPiTemplateApplied)
	assert.Equal(s.T(), unitID, *metadata.AutoPiUnitID)
	assert.Equal(s.T(), imei, *metadata.AutoPiIMEI)
	assert.Equal(s.T(), services.PendingTemplateConfirm.String(), *metadata.AutoPiSubStatus)
}
func (s *UserIntegrationsControllerTestSuite) TestPostAutoPiCustomPowerTrain() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	// arrange
	evTemplateID := 12
	powertrain := "BEV"
	integration := test.SetupCreateAutoPiIntegration(s.T(), 34, &evTemplateID, s.pdb)
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, &powertrain, s.pdb)
	const (
		jobID     = "123"
		deviceID  = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID    = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
		vehicleID = 123
	)

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)
	// setup all autoPi mock expected calls.
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		ID:                deviceID, // device id
		UnitID:            unitID,
		Vehicle:           services.AutoPiDongleVehicle{ID: vehicleID}, // vehicle profile id
		IMEI:              "IMEI321",
		Template:          1,
		LastCommunication: time.Now().UTC().Add(time.Second * -20),
	}, nil)

	autopiAPISvc.EXPECT().PatchVehicleProfile(vehicleID, gomock.Any()).Times(1).Return(nil)
	autopiAPISvc.EXPECT().UnassociateDeviceTemplate(deviceID, 1).Times(1).Return(nil)

	// todo: validate with james
	//autopiAPISvc.EXPECT().ApplyTemplate(deviceID, evTemplateID).Times(1).Return(nil)
	autopiAPISvc.EXPECT().AssociateDeviceToTemplate(deviceID, gomock.Any()).Times(1).Return(nil)
	autopiAPISvc.EXPECT().ApplyTemplate(deviceID, gomock.Any()).Times(1).Return(nil)
	autopiAPISvc.EXPECT().CommandSyncDevice(gomock.Any(), unitID, deviceID, ud.ID).Times(1).Return(&services.AutoPiCommandResponse{
		Jid: jobID,
	}, nil)
	s.autoPiIngest.EXPECT().Register(unitID, ud.ID, integration.ID).Return(nil)

	s.deviceDefinitionRegistrar.EXPECT().Register(services.DeviceDefinitionDTO{
		IntegrationID:      integration.ID,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dm.Name,
		Model:              dd.Model,
		Year:               int(dd.Year),
	}).Return(nil)

	ddIntegrations := []*ddgrpc.GetDeviceDefinitionIntegrationItemResponse{}

	integrationItem := &ddgrpc.GetIntegrationItemResponse{
		Id:                      integration.ID,
		Vendor:                  integration.Vendor,
		Style:                   integration.Style,
		Type:                    integration.Type,
		AutoPiDefaultTemplateId: int32(evTemplateID),
		AutoPiPowertrainTemplate: &ddgrpc.GetIntegrationItemResponse_GetAutoPiPowertrainTemplate{
			BEV: 14,
		},
	}

	deviceDefinitions := test.BuildDeviceDefinitionGRPC(ud.DeviceDefinitionID, dm.Name, dd.Model, "Vehicle")
	deviceDefinitions[0].Type.Year = int32(dd.Year)

	s.deviceDefIntSvc.EXPECT().GetDeviceDefinitionIntegration(gomock.Any(), gomock.Any()).Times(1).Return(ddIntegrations, nil)

	s.deviceDefIntSvc.EXPECT().CreateDeviceDefinitionIntegration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(integrationItem, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd.ID}).Times(1).Return(deviceDefinitions, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
	response, _ := app.Test(request, 200000)
	if assert.Equal(s.T(), fiber.StatusNoContent, response.StatusCode, "should return success") == false {
		body, _ := io.ReadAll(response.Body)
		fmt.Println("unexpected response: " + string(body) + "\n")
		fmt.Println("body sent to post: " + req)
	}

	apiInt, err := models.FindUserDeviceAPIIntegration(s.ctx, s.pdb.DBS().Writer, ud.ID, integration.ID)
	require.NoError(s.T(), err)
	fmt.Printf("found user device api int: %+v", *apiInt)

	metadata := new(services.UserDeviceAPIIntegrationsMetadata)
	err = apiInt.Metadata.Unmarshal(metadata)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), deviceID, apiInt.ExternalID.String)
	assert.Equal(s.T(), "Pending", apiInt.Status)
	//assert.Equal(s.T(), evTemplateID, *metadata.AutoPiTemplateApplied)
}
func (s *UserIntegrationsControllerTestSuite) TestPostAutoPiBlockedForDuplicateDeviceSameUser() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUserID), c.RegisterDeviceIntegration)
	// arrange
	integration := test.SetupCreateAutoPiIntegration(s.T(), 34, nil, s.pdb)
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integration.ID, s.pdb)

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)
	// no calls should be made to autopi api

	ddIntegrations := []*ddgrpc.GetDeviceDefinitionIntegrationItemResponse{}
	ddIntegrations = append(ddIntegrations, &ddgrpc.GetDeviceDefinitionIntegrationItemResponse{
		Id:     integration.ID,
		Vendor: integration.Vendor,
		Style:  integration.Style,
		Type:   integration.Type,
	})

	deviceDefinitions := test.BuildDeviceDefinitionGRPC(ud.DeviceDefinitionID, "Ford", "Mach E", "Vehicle")
	deviceDefinitions[0].DeviceIntegrations = append(deviceDefinitions[0].DeviceIntegrations, &ddgrpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{
		Id:     integration.ID,
		Type:   models.IntegrationTypeAPI,
		Style:  models.IntegrationStyleWebhook,
		Vendor: "SmartCar",
	})

	s.deviceDefIntSvc.EXPECT().GetDeviceDefinitionIntegration(gomock.Any(), ud.DeviceDefinitionID).Times(1).Return(ddIntegrations, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{ud.DeviceDefinitionID}).AnyTimes().Return(deviceDefinitions, nil)

	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/integrations/"+integration.ID, req)
	response, _ := app.Test(request)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should return failure")
	body, _ := io.ReadAll(response.Body)
	fmt.Println("body response: " + string(body) + "\n")
}
func (s *UserIntegrationsControllerTestSuite) TestPostAutoPiBlockedForDuplicateDeviceDifferentUser() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Post("/user/devices/:userDeviceID/integrations/:integrationID", test.AuthInjectorTestHandler(testUser2), c.RegisterDeviceIntegration)
	// arrange
	integration := test.SetupCreateAutoPiIntegration(s.T(), 34, nil, s.pdb)
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
	// the other user that already claimed unit
	_ = test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	// test user
	ud2 := test.SetupCreateUserDevice(s.T(), testUser2, dd, nil, s.pdb)

	req := fmt.Sprintf(`{
			"externalId": "%s"
		}`, unitID)

	ddIntegrations := []*ddgrpc.GetDeviceDefinitionIntegrationItemResponse{}
	ddIntegrations = append(ddIntegrations, &ddgrpc.GetDeviceDefinitionIntegrationItemResponse{
		Id:     integration.ID,
		Vendor: integration.Vendor,
		Style:  integration.Style,
		Type:   integration.Type,
	})
	grpcIntegration := &ddgrpc.GetIntegrationItemResponse{
		Id:                       integration.ID,
		Type:                     integration.Type,
		Style:                    integration.Style,
		Vendor:                   integration.Vendor,
		AutoPiDefaultTemplateId:  0,
		AutoPiPowertrainTemplate: nil,
	}

	deviceDefinitions := test.BuildDeviceDefinitionGRPC(dd.ID, "Ford", "Mach E", "Vehicle")

	s.deviceDefIntSvc.EXPECT().GetDeviceDefinitionIntegration(gomock.Any(), gomock.Any()).Times(1).Return(ddIntegrations, nil)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dd.ID}).Times(1).Return(deviceDefinitions, nil)
	s.deviceDefSvc.EXPECT().GetIntegrationByID(gomock.Any(), integration.ID).Return(grpcIntegration, nil)

	// no calls should be made to autopi api
	request := test.BuildRequest("POST", "/user/devices/"+ud2.ID+"/integrations/"+integration.ID, req)
	response, err := app.Test(request, 2000)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode, "should return bad request")
	body, _ := io.ReadAll(response.Body)
	fmt.Println("body response: " + string(body) + "\n")
}

// todo: validate with james
//func (s *UserIntegrationsControllerTestSuite) TestPostAutoPiCommand() {
//	// specific dependency and controller
//	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
//	c := NewUserDevicesController(&config.settings{Port: "3000"}, s.pdb.dbs, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
//		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
//		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
//	app := fiber.New()
//	app.Post("/user/devices/:userDeviceID/autopi/command", test.AuthInjectorTestHandler(testUserID), c.SendAutoPiCommand)
//	// arrange
//	integ := test.SetupCreateAutoPiIntegration(s.T(), 34, nil, s.pdb)
//	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
//	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
//	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
//	test.SetupCreateDeviceIntegration(s.T(), dd, integ, s.pdb)
//	const (
//		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
//		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
//	)
//	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
//	udapiInt := test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integ.ID, s.pdb)
//
//	udMetadata := services.UserDeviceAPIIntegrationsMetadata{
//		AutoPiUnitID: func(s string) *string { return &s }(unitID),
//	}
//	_ = udapiInt.Metadata.Marshal(udMetadata)
//	_, err := udapiInt.Update(s.ctx, s.pdb.dbs().Writer, boil.Infer())
//	require.NoError(s.T(), err)
//	autoPiJob := models.AutopiJob{
//		ID:                 "somepreviousjobId",
//		AutopiDeviceID:     deviceID,
//		Command:            "raw",
//		State:              "COMMAND_EXECUTED",
//		CommandLastUpdated: null.TimeFrom(time.Now().UTC()),
//		UserDeviceID:       null.StringFrom(ud.ID),
//	}
//	err = autoPiJob.Insert(s.ctx, s.pdb.dbs().Writer, boil.Infer())
//	require.NoError(s.T(), err)
//	// test job can be retrieved
//	apSvc := services.NewAutoPiAPIService(&config.settings{}, s.pdb.dbs)
//	status, _, err := apSvc.GetCommandStatus(s.ctx, "somepreviousjobId")
//	require.NoError(s.T(), err)
//	assert.Equal(s.T(), "somepreviousjobId", status.CommandJobID)
//	assert.Equal(s.T(), autoPiJob.State, status.CommandState)
//	assert.Equal(s.T(), "raw", status.CommandRaw)
//
//	// test sending a command from api
//	const jobID = "123"
//	// mock expectations
//	const cmd = "raw test"
//	autopiAPISvc.EXPECT().CommandRaw(gomock.Any(), unitID, deviceID, cmd, ud.ID).Return(&services.AutoPiCommandResponse{
//		Jid:     jobID,
//		Minions: nil,
//	}, nil)
//	// act: send request
//	req := fmt.Sprintf(`{
//			"command": "%s"
//		}`, cmd)
//
//	s.deviceDefIntSvc.EXPECT().FindUserDeviceAutoPiIntegration(gomock.Any(), gomock.Any(), ud.ID, testUserID).Times(1).Return(udapiInt, nil)
//
//	request := test.BuildRequest("POST", "/user/devices/"+ud.ID+"/autopi/command", req)
//	response, _ := app.Test(request, 20000)
//	body, _ := io.ReadAll(response.Body)
//	//assert
//	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
//	jid := gjson.GetBytes(body, "jid")
//	assert.Equal(s.T(), jobID, jid.String())
//}

func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiCommand() {
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Get("/user/devices/:userDeviceID/autopi/command/:jobID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiCommandStatus)
	//arrange
	integ := test.SetupCreateAutoPiIntegration(s.T(), 34, nil, s.pdb)
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	test.SetupCreateDeviceIntegration(s.T(), dd, integ, s.pdb)
	const (
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		jobID    = "somepreviousjobId"
	)
	_ = test.SetupCreateUserDeviceAPIIntegration(s.T(), "", deviceID, ud.ID, integ.ID, s.pdb)

	lastUpdated := time.Now()

	autopiAPISvc.EXPECT().GetCommandStatus(gomock.Any(), jobID).Return(&services.AutoPiCommandJob{
		CommandJobID: jobID,
		CommandState: "COMMAND_EXECUTED",
		CommandRaw:   "raw",
		LastUpdated:  &lastUpdated,
	}, &models.AutopiJob{
		ID:                 jobID,
		AutopiDeviceID:     deviceID,
		Command:            "raw",
		State:              "COMMAND_EXECUTED",
		CommandLastUpdated: null.TimeFrom(lastUpdated),
		UserDeviceID:       null.StringFrom(ud.ID),
	}, nil)

	// act: send request
	request := test.BuildRequest("GET", "/user/devices/"+ud.ID+"/autopi/command/"+jobID, "")
	response, _ := app.Test(request)
	require.Equal(s.T(), fiber.StatusOK, response.StatusCode)

	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), jobID, gjson.GetBytes(body, "commandJobId").String())
	assert.Equal(s.T(), "COMMAND_EXECUTED", gjson.GetBytes(body, "commandState").String())
	assert.Equal(s.T(), "raw", gjson.GetBytes(body, "commandRaw").String())

}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiCommandNoResults400() {
	//arrange
	integ := test.SetupCreateAutoPiIntegration(s.T(), 34, nil, s.pdb)
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), testUserID, dd, nil, s.pdb)
	test.SetupCreateDeviceIntegration(s.T(), dd, integ, s.pdb)
	const (
		jobID    = "somepreviousjobId2"
		deviceID = "1dd96159-3bb2-9472-91f6-72fe9211cfeb"
		unitID   = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }(deviceID), s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, deviceID, ud.ID, integ.ID, s.pdb)

	s.autopiAPISvc.EXPECT().GetCommandStatus(gomock.Any(), jobID).Return(nil, nil, sql.ErrNoRows)

	// act: send request
	request := test.BuildRequest("GET", "/user/devices/"+ud.ID+"/autopi/command/"+jobID, "")
	response, _ := s.app.Test(request)
	//assert
	assert.Equal(s.T(), fiber.StatusBadRequest, response.StatusCode)
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_ShouldUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         false,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "1.21.6"}),
	}, nil)
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(nil, nil)
	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), false, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), unitID, gjson.GetBytes(body, "unitId").String())
	assert.Equal(s.T(), "4321", gjson.GetBytes(body, "deviceId").String())
	assert.Equal(s.T(), "1.23", gjson.GetBytes(body, "hwRevision").String())
	assert.Equal(s.T(), "1.21.6", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), true, gjson.GetBytes(body, "shouldUpdate").Bool()) // this because releaseVersion below 1.21.9
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_UpToDate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         true,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "1.21.9"}),
	}, nil)
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(nil, nil)
	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), true, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), "1.21.9", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), false, gjson.GetBytes(body, "shouldUpdate").Bool()) // returned version is 1.21.9 which is our cutoff
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoUDAI_FutureUpdate() {
	const environment = "prod" // shouldUpdate only applies in prod
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000", Environment: environment}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	autopiAPISvc.EXPECT().GetDeviceByUnitID(unitID).Times(1).Return(&services.AutoPiDongleDevice{
		IsUpdated:         false,
		UnitID:            unitID,
		ID:                "4321",
		HwRevision:        "1.23",
		Template:          10,
		LastCommunication: time.Now(),
		Release: struct {
			Version string `json:"version"`
		}(struct{ Version string }{Version: "1.23.1"}),
	}, nil)
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(nil, nil)
	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusOK, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	//assert
	assert.Equal(s.T(), false, gjson.GetBytes(body, "isUpdated").Bool())
	assert.Equal(s.T(), "1.23.1", gjson.GetBytes(body, "releaseVersion").String())
	assert.Equal(s.T(), false, gjson.GetBytes(body, "shouldUpdate").Bool())
}
func (s *UserIntegrationsControllerTestSuite) TestGetAutoPiInfoNoMatchUDAI() {
	// specific dependency and controller
	autopiAPISvc := mock_services.NewMockAutoPiAPIService(s.mockCtrl)
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, test.Logger(), s.deviceDefSvc, s.deviceDefIntSvc,
		&fakeEventService{}, s.scClient, s.scTaskSvc, s.teslaSvc, s.teslaTaskService, new(shared.ROT13Cipher), autopiAPISvc,
		nil, s.autoPiIngest, s.deviceDefinitionRegistrar, nil, nil, nil, s.drivlyTaskSvc, s.blackbookTaskSvc)
	app := fiber.New()
	app.Get("/autopi/unit/:unitID", test.AuthInjectorTestHandler(testUserID), c.GetAutoPiUnitInfo)
	// arrange
	const unitID = "431d2e89-46f1-6884-6226-5d1ad20c84d9"
	integ := test.SetupCreateAutoPiIntegration(s.T(), 34, nil, s.pdb)
	dm := test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dd := test.SetupCreateDeviceDefinition(s.T(), dm, "Model 4", 2022, s.pdb)
	ud := test.SetupCreateUserDevice(s.T(), "some-other-user", dd, nil, s.pdb)
	test.SetupCreateDeviceIntegration(s.T(), dd, integ, s.pdb)
	_ = test.SetupCreateAutoPiUnit(s.T(), testUserID, unitID, func(s string) *string { return &s }("1234"), s.pdb)
	test.SetupCreateUserDeviceAPIIntegration(s.T(), unitID, "321", ud.ID, integ.ID, s.pdb)

	udai := models.UserDeviceAPIIntegration{}
	udai.R = udai.R.NewStruct()
	udai.R.UserDevice = &ud
	autopiAPISvc.EXPECT().GetUserDeviceIntegrationByUnitID(gomock.Any(), unitID).Return(&udai, nil)

	// act
	request := test.BuildRequest("GET", "/autopi/unit/"+unitID, "")
	response, err := app.Test(request)
	require.NoError(s.T(), err)
	// assert
	assert.Equal(s.T(), fiber.StatusForbidden, response.StatusCode)
}

func TestUserDevicesController_ClaimAutoPi(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.ClaimAutoPi(tt.args.c), fmt.Sprintf("ClaimAutoPi(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_DeleteUserDeviceIntegration(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.DeleteUserDeviceIntegration(tt.args.c), fmt.Sprintf("DeleteUserDeviceIntegration(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetAutoPiClaimMessage(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetAutoPiClaimMessage(tt.args.c), fmt.Sprintf("GetAutoPiClaimMessage(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetAutoPiCommandStatus(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetAutoPiCommandStatus(tt.args.c), fmt.Sprintf("GetAutoPiCommandStatus(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetAutoPiTask(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetAutoPiTask(tt.args.c), fmt.Sprintf("GetAutoPiTask(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetAutoPiUnitInfo(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetAutoPiUnitInfo(tt.args.c), fmt.Sprintf("GetAutoPiUnitInfo(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetCommandRequestStatus(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetCommandRequestStatus(tt.args.c), fmt.Sprintf("GetCommandRequestStatus(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetIntegrations(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetIntegrations(tt.args.c), fmt.Sprintf("GetIntegrations(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetIsAutoPiOnline(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetIsAutoPiOnline(tt.args.c), fmt.Sprintf("GetIsAutoPiOnline(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_GetUserDeviceIntegration(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.GetUserDeviceIntegration(tt.args.c), fmt.Sprintf("GetUserDeviceIntegration(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_LockDoors(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.LockDoors(tt.args.c), fmt.Sprintf("LockDoors(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_OpenFrunk(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.OpenFrunk(tt.args.c), fmt.Sprintf("OpenFrunk(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_OpenTrunk(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.OpenTrunk(tt.args.c), fmt.Sprintf("OpenTrunk(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_RegisterDeviceIntegration(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.RegisterDeviceIntegration(tt.args.c), fmt.Sprintf("RegisterDeviceIntegration(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_SendAutoPiCommand(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.SendAutoPiCommand(tt.args.c), fmt.Sprintf("SendAutoPiCommand(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_StartAutoPiUpdateTask(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.StartAutoPiUpdateTask(tt.args.c), fmt.Sprintf("StartAutoPiUpdateTask(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_UnlockDoors(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.UnlockDoors(tt.args.c), fmt.Sprintf("UnlockDoors(%v)", tt.args.c))
		})
	}
}

func TestUserDevicesController_fixSmartcarDeviceYear(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		ctx    context.Context
		logger *zerolog.Logger
		exec   boil.ContextExecutor
		integ  *ddgrpc.GetIntegrationItemResponse
		ud     *models.UserDevice
		year   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.fixSmartcarDeviceYear(tt.args.ctx, tt.args.logger, tt.args.exec, tt.args.integ, tt.args.ud, tt.args.year), fmt.Sprintf("fixSmartcarDeviceYear(%v, %v, %v, %v, %v, %v)", tt.args.ctx, tt.args.logger, tt.args.exec, tt.args.integ, tt.args.ud, tt.args.year))
		})
	}
}

func TestUserDevicesController_handleEnqueueCommand(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c           *fiber.Ctx
		commandPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.handleEnqueueCommand(tt.args.c, tt.args.commandPath), fmt.Sprintf("handleEnqueueCommand(%v, %v)", tt.args.c, tt.args.commandPath))
		})
	}
}

func TestUserDevicesController_registerAutoPiUnit(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c             *fiber.Ctx
		logger        *zerolog.Logger
		tx            *sql.Tx
		ud            *models.UserDevice
		integrationID string
		dd            *ddgrpc.GetDeviceDefinitionItemResponse
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.registerAutoPiUnit(tt.args.c, tt.args.logger, tt.args.tx, tt.args.ud, tt.args.integrationID, tt.args.dd), fmt.Sprintf("registerAutoPiUnit(%v, %v, %v, %v, %v, %v)", tt.args.c, tt.args.logger, tt.args.tx, tt.args.ud, tt.args.integrationID, tt.args.dd))
		})
	}
}

func TestUserDevicesController_registerDeviceTesla(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c            *fiber.Ctx
		logger       *zerolog.Logger
		tx           *sql.Tx
		userDeviceID string
		integ        *ddgrpc.GetIntegrationItemResponse
		ud           *models.UserDevice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.registerDeviceTesla(tt.args.c, tt.args.logger, tt.args.tx, tt.args.userDeviceID, tt.args.integ, tt.args.ud), fmt.Sprintf("registerDeviceTesla(%v, %v, %v, %v, %v, %v)", tt.args.c, tt.args.logger, tt.args.tx, tt.args.userDeviceID, tt.args.integ, tt.args.ud))
		})
	}
}

func TestUserDevicesController_registerSmartcarIntegration(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		c      *fiber.Ctx
		logger *zerolog.Logger
		tx     *sql.Tx
		integ  *ddgrpc.GetIntegrationItemResponse
		ud     *models.UserDevice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			tt.wantErr(t, udc.registerSmartcarIntegration(tt.args.c, tt.args.logger, tt.args.tx, tt.args.integ, tt.args.ud), fmt.Sprintf("registerSmartcarIntegration(%v, %v, %v, %v, %v)", tt.args.c, tt.args.logger, tt.args.tx, tt.args.integ, tt.args.ud))
		})
	}
}

func TestUserDevicesController_runPostRegistration(t *testing.T) {
	type fields struct {
		Settings                  *config.Settings
		DBS                       func() *database.DBReaderWriter
		DeviceDefSvc              services.DeviceDefinitionService
		DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
		log                       *zerolog.Logger
		eventService              services.EventService
		smartcarClient            services.SmartcarClient
		smartcarTaskSvc           services.SmartcarTaskService
		teslaService              services.TeslaService
		teslaTaskService          services.TeslaTaskService
		cipher                    shared.Cipher
		autoPiSvc                 services.AutoPiAPIService
		nhtsaService              services.INHTSAService
		autoPiIngestRegistrar     services.IngestRegistrar
		autoPiTaskService         services.AutoPiTaskService
		drivlyTaskService         services.DrivlyTaskService
		blackbookTaskService      services.BlackbookTaskService
		s3                        *s3.Client
		producer                  sarama.SyncProducer
		deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
	}
	type args struct {
		ctx           context.Context
		logger        *zerolog.Logger
		userDeviceID  string
		integrationID string
		integ         *ddgrpc.GetIntegrationItemResponse
		dd            *ddgrpc.GetDeviceDefinitionItemResponse
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			udc := &UserDevicesController{
				Settings:                  tt.fields.Settings,
				DBS:                       tt.fields.DBS,
				DeviceDefSvc:              tt.fields.DeviceDefSvc,
				DeviceDefIntSvc:           tt.fields.DeviceDefIntSvc,
				log:                       tt.fields.log,
				eventService:              tt.fields.eventService,
				smartcarClient:            tt.fields.smartcarClient,
				smartcarTaskSvc:           tt.fields.smartcarTaskSvc,
				teslaService:              tt.fields.teslaService,
				teslaTaskService:          tt.fields.teslaTaskService,
				cipher:                    tt.fields.cipher,
				autoPiSvc:                 tt.fields.autoPiSvc,
				nhtsaService:              tt.fields.nhtsaService,
				autoPiIngestRegistrar:     tt.fields.autoPiIngestRegistrar,
				autoPiTaskService:         tt.fields.autoPiTaskService,
				drivlyTaskService:         tt.fields.drivlyTaskService,
				blackbookTaskService:      tt.fields.blackbookTaskService,
				s3:                        tt.fields.s3,
				producer:                  tt.fields.producer,
				deviceDefinitionRegistrar: tt.fields.deviceDefinitionRegistrar,
			}
			udc.runPostRegistration(tt.args.ctx, tt.args.logger, tt.args.userDeviceID, tt.args.integrationID, tt.args.integ, tt.args.dd)
		})
	}
}

func Test_fixTeslaDeviceDefinition(t *testing.T) {
	type args struct {
		ctx    context.Context
		logger *zerolog.Logger
		ddSvc  services.DeviceDefinitionService
		exec   boil.ContextExecutor
		integ  *ddgrpc.GetIntegrationItemResponse
		ud     *models.UserDevice
		vin    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, fixTeslaDeviceDefinition(tt.args.ctx, tt.args.logger, tt.args.ddSvc, tt.args.exec, tt.args.integ, tt.args.ud, tt.args.vin), fmt.Sprintf("fixTeslaDeviceDefinition(%v, %v, %v, %v, %v, %v, %v)", tt.args.ctx, tt.args.logger, tt.args.ddSvc, tt.args.exec, tt.args.integ, tt.args.ud, tt.args.vin))
		})
	}
}
