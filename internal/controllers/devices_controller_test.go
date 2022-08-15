package controllers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
)

type DevicesControllerTestSuite struct {
	suite.Suite
	pdb         database.DbStore
	container   testcontainers.Container
	ctx         context.Context
	deviceDefID string
	mockCtrl    *gomock.Controller
	app         *fiber.App
	dbMake      models.DeviceMake
}

// SetupSuite starts container db
func (s *DevicesControllerTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pdb, s.container = test.StartContainerDatabase(s.ctx, s.T(), migrationsDirRelPath)
	s.mockCtrl = gomock.NewController(s.T())
	logger := test.Logger()

	nhtsaSvc := mock_services.NewMockINHTSAService(s.mockCtrl)
	deviceDefSvc := mock_services.NewMockIDeviceDefinitionService(s.mockCtrl)
	c := NewDevicesController(&config.Settings{Port: "3000"}, s.pdb.DBS, logger, nhtsaSvc, deviceDefSvc)

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
	request, _ := http.NewRequest("GET", "/device-definitions/"+s.deviceDefID, nil)
	response, _ := s.app.Test(request)
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

func TestNewDeviceDefinitionFromDatabase(t *testing.T) {
	dbMake := &models.DeviceMake{
		ID:   ksuid.New().String(),
		Name: "Mercedes",
	}
	dbDevice := models.DeviceDefinition{
		ID:           "123",
		DeviceMakeID: dbMake.ID,
		Model:        "R500",
		Year:         2020,
		Metadata:     null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
	}
	ds := models.DeviceStyle{
		SubModel:           "AMG",
		Name:               "C63 AMG",
		DeviceDefinitionID: dbDevice.ID,
	}
	di := models.DeviceIntegration{
		DeviceDefinitionID: "123",
		IntegrationID:      "123",
		CreatedAt:          time.Time{},
		UpdatedAt:          time.Time{},
	}
	di.R = di.R.NewStruct()
	di.R.Integration = &models.Integration{
		ID:     "123",
		Type:   "Hardware",
		Style:  "Addon",
		Vendor: "Autopi",
	}
	dbDevice.R = dbDevice.R.NewStruct()
	dbDevice.R.DeviceMake = dbMake
	dbDevice.R.DeviceIntegrations = append(dbDevice.R.DeviceIntegrations, &di)
	dbDevice.R.DeviceStyles = append(dbDevice.R.DeviceStyles, &ds)
	dd, err := NewDeviceDefinitionFromDatabase(&dbDevice)

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
	dbDevice := models.DeviceDefinition{
		ID:       "123",
		Model:    "R500",
		Year:     2020,
		Metadata: null.JSONFrom([]byte(`{"vehicle_info": {"fuel_type": "gas", "driven_wheels": "4", "number_of_doors":"5" } }`)),
	}
	dbDevice.R = dbDevice.R.NewStruct()
	_, err := NewDeviceDefinitionFromDatabase(&dbDevice)
	assert.Error(t, err)

	dbDevice.R = nil
	_, err = NewDeviceDefinitionFromDatabase(&dbDevice)
	assert.Error(t, err)
}
