package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
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
)

type DevicesControllerTestSuite struct {
	suite.Suite
	pdb             database.DbStore
	container       testcontainers.Container
	ctx             context.Context
	deviceDefID     string
	mockCtrl        *gomock.Controller
	app             *fiber.App
	dbMake          models.DeviceMake
	deviceDefSvc    *mock_services.MockDeviceDefinitionService
	deviceDefIntSvc *mock_services.MockDeviceDefinitionIntegrationService
}

// SetupSuite starts container db
func (s *DevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	s.mockCtrl = gomock.NewController(s.T())
	logger := test.Logger()

	nhtsaSvc := mock_services.NewMockINHTSAService(s.mockCtrl)
	s.deviceDefSvc = mock_services.NewMockDeviceDefinitionService(s.mockCtrl)
	s.deviceDefIntSvc = mock_services.NewMockDeviceDefinitionIntegrationService(s.mockCtrl)
	c := NewDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, nhtsaSvc, s.deviceDefSvc, s.deviceDefIntSvc)

	// routes
	app := fiber.New()
	app.Get("/device-definitions/all", c.GetAllDeviceMakeModelYears)
	app.Get("/device-definitions/:id", c.GetDeviceDefinitionByID)
	app.Get("/device-definitions/:id/integrations", c.GetDeviceIntegrationsByID)

	s.app = app

	// arrange some data
	s.dbMake = test.SetupCreateMake(s.T(), "Testla", s.pdb)
	dbDeviceDef := test.SetupCreateDeviceDefinition(s.T(), s.dbMake, "MODEL Y", 2020, s.pdb)
	s.deviceDefID = dbDeviceDef.ID

	// note we do not want to truncate tables after each test for this one
}

// TearDownSuite cleanup at end by terminating container
func (s *DevicesControllerTestSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
	s.mockCtrl.Finish()
}

func TestDevicesControllerTestSuite(t *testing.T) {
	suite.Run(t, new(DevicesControllerTestSuite))
}

/* Actual tests*/

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionById() {

	ddGRPC := test.BuildDeviceDefinitionGRPC(s.deviceDefID, "Ford", "Ford", "Vehicle")

	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{s.deviceDefID}).Times(1).Return(ddGRPC, nil) // todo move to each test where used

	deviceCompatibilities := []services.DeviceCompatibility{}
	deviceCompatibilities = append(deviceCompatibilities, services.DeviceCompatibility{
		Vendor:       services.AutoPiVendor,
		Region:       "Americas",
		Capabilities: nil,
	})
	deviceCompatibilities = append(deviceCompatibilities, services.DeviceCompatibility{
		Vendor:       services.AutoPiVendor,
		Region:       "Europe",
		Capabilities: nil,
	})

	s.deviceDefIntSvc.EXPECT().AppendAutoPiCompatibility(gomock.Any(), gomock.Any(), s.deviceDefID).Times(1).Return(deviceCompatibilities, nil)

	request, _ := http.NewRequest("GET", "/device-definitions/"+s.deviceDefID, nil)
	response, errRes := s.app.Test(request)
	require.NoError(s.T(), errRes)

	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)

	v := gjson.GetBytes(body, "deviceDefinition")
	var dd services.DeviceDefinition
	err := json.Unmarshal([]byte(v.Raw), &dd)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), s.deviceDefID, dd.DeviceDefinitionID)
	if assert.True(s.T(), len(dd.CompatibleIntegrations) >= 2, "should be atleast 2 integrations for autopi") {
		assert.Equal(s.T(), services.AutoPiVendor, dd.CompatibleIntegrations[0].Vendor)
		assert.Equal(s.T(), "Americas", dd.CompatibleIntegrations[0].Region)
		assert.Equal(s.T(), services.AutoPiVendor, dd.CompatibleIntegrations[1].Vendor)
		assert.Equal(s.T(), "Europe", dd.CompatibleIntegrations[1].Region)
	} else {
		fmt.Printf("found integrations: %+v", dd.CompatibleIntegrations)
	}
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionDoesNotAddAutoPiForOldCars() {
	dbDdOldCar := test.SetupCreateDeviceDefinition(s.T(), s.dbMake, "Oldie", 1999, s.pdb)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{dbDdOldCar.ID}).Times(1).Return(test.BuildDeviceDefinitionGRPC(dbDdOldCar.ID, "Tesla", "Tesla", "Vehicle"), nil) // todo move to each test where used

	request, _ := http.NewRequest("GET", "/device-definitions/"+dbDdOldCar.ID, nil)
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)
	v := gjson.GetBytes(body, "deviceDefinition")
	var dd services.DeviceDefinition
	err := json.Unmarshal([]byte(v.Raw), &dd)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), dbDdOldCar.ID, dd.DeviceDefinitionID)
	assert.Len(s.T(), dd.CompatibleIntegrations, 0, "vehicles before 2020 should not auto inject autopi integrations")
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionDoesNotAddAutoPiForTesla() {
	tesla := test.SetupCreateMake(s.T(), "Tesla", s.pdb)
	teslaCar := test.SetupCreateDeviceDefinition(s.T(), tesla, "Cyber Truck never", 2022, s.pdb)
	s.deviceDefSvc.EXPECT().GetDeviceDefinitionsByIDs(gomock.Any(), []string{teslaCar.ID}).Times(1).Return(test.BuildDeviceDefinitionGRPC(tesla.ID, "Tesla", "Tesla", "Vehicle"), nil) // todo move to each test where used

	request, _ := http.NewRequest("GET", "/device-definitions/"+teslaCar.ID, nil)
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)
	v := gjson.GetBytes(body, "deviceDefinition")
	var dd services.DeviceDefinition
	err := json.Unmarshal([]byte(v.Raw), &dd)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), dd.CompatibleIntegrations, 0, "vehicles before 2020 should not auto inject autopi integrations")
}

func (s *DevicesControllerTestSuite) TestGetDeviceIntegrationsById() {

	deviceCompatibilities := []services.DeviceCompatibility{}
	deviceCompatibilities = append(deviceCompatibilities, services.DeviceCompatibility{
		Vendor:       services.AutoPiVendor,
		Region:       "Americas",
		Capabilities: nil,
	})
	deviceCompatibilities = append(deviceCompatibilities, services.DeviceCompatibility{
		Vendor:       services.AutoPiVendor,
		Region:       "Europe",
		Capabilities: nil,
	})

	s.deviceDefIntSvc.EXPECT().AppendAutoPiCompatibility(gomock.Any(), gomock.Any(), s.deviceDefID).Times(1).Return(deviceCompatibilities, nil)

	request, _ := http.NewRequest("GET", "/device-definitions/"+s.deviceDefID+"/integrations", nil)
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)
	v := gjson.GetBytes(body, "compatibleIntegrations")
	var dc []services.DeviceCompatibility
	err := json.Unmarshal([]byte(v.Raw), &dc)
	assert.NoError(s.T(), err)
	if assert.True(s.T(), len(dc) >= 2, "should be atleast 2 integrations for autopi") {
		assert.Equal(s.T(), services.AutoPiVendor, dc[0].Vendor)
		assert.Equal(s.T(), "Americas", dc[0].Region)
		assert.Equal(s.T(), services.AutoPiVendor, dc[1].Vendor)
		assert.Equal(s.T(), "Europe", dc[1].Region)
	}
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefinitionWithInvalidID() {
	request, _ := http.NewRequest("GET", "/device-definitions/caca", nil)
	response, _ := s.app.Test(request)
	// assert
	assert.Equal(s.T(), 400, response.StatusCode)
}

func (s *DevicesControllerTestSuite) TestGetDeviceDefIntegrationWithInvalidID() {
	request, _ := http.NewRequest("GET", "/device-definitions/caca/integrations", nil)
	response, _ := s.app.Test(request)
	// assert
	assert.Equal(s.T(), 400, response.StatusCode)
}

func (s *DevicesControllerTestSuite) TestGetAll() {
	request, _ := http.NewRequest("GET", "/device-definitions/all", nil)
	response, _ := s.app.Test(request)
	body, _ := io.ReadAll(response.Body)
	// assert
	assert.Equal(s.T(), 200, response.StatusCode)
	v := gjson.GetBytes(body, "makes")
	var mmy []DeviceMMYRoot
	err := json.Unmarshal([]byte(v.Raw), &mmy)
	assert.NoError(s.T(), err)
	if assert.True(s.T(), len(mmy) >= 1, "should be at least one device definition") {
		assert.Equal(s.T(), "Testla", mmy[0].Make)
		assert.Equal(s.T(), "MODEL Y", mmy[0].Models[0].Model)
		assert.Equal(s.T(), int16(2020), mmy[0].Models[0].Years[0].Year)
		assert.Equal(s.T(), s.deviceDefID, mmy[0].Models[0].Years[0].DeviceDefinitionID)
	}
}

func TestNewDeviceDefinitionFromGrpc(t *testing.T) {
	subModels := []string{"AMG"}
	dbDevice := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "123",
		Type: &grpc.GetDeviceDefinitionItemResponse_Type{
			Type:      "Vehicle",
			Model:     "R500",
			Year:      2020,
			Make:      "Mercedes",
			SubModels: subModels,
		},
		VehicleData: &grpc.VehicleInfo{
			FuelType:      "gas",
			DrivenWheels:  "4",
			NumberOfDoors: 5,
		},
		Make: &grpc.GetDeviceDefinitionItemResponse_Make{
			Id:   "1",
			Name: "Mercedes",
		},
		DeviceIntegrations: append([]*grpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{}, &grpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{
			Id: "123",
		}),
		CompatibleIntegrations: append([]*grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations{}, &grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations{
			Vendor: "Autopi",
		}),
		//Metadata:     null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
	}

	dd, err := NewDeviceDefinitionFromGRPC(dbDevice)

	assert.NoError(t, err)
	assert.Equal(t, "123", dd.DeviceDefinitionID)
	assert.Equal(t, "gas", dd.VehicleInfo.FuelType)
	assert.Equal(t, "4", dd.VehicleInfo.DrivenWheels)
	assert.Equal(t, "5", dd.VehicleInfo.NumberOfDoors)
	assert.Equal(t, "Vehicle", dd.Type.Type)
	assert.Equal(t, 2020, dd.Type.Year)
	assert.Equal(t, "Mercedes", dd.Type.Make)
	assert.Equal(t, "R500", dd.Type.Model)
	assert.Contains(t, dd.Type.SubModels, "AMG")

	assert.Len(t, dd.CompatibleIntegrations, 1)
	assert.Equal(t, "Autopi", dd.CompatibleIntegrations[0].Vendor)
}

func TestNewDeviceDefinitionFromDatabase_Error(t *testing.T) {
	dbDevice := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: "123",
		VehicleData: &grpc.VehicleInfo{
			FuelType:      "gas",
			DrivenWheels:  "4",
			NumberOfDoors: 5,
		},
	}
	_, err := NewDeviceDefinitionFromGRPC(dbDevice)
	assert.Error(t, err)
}
