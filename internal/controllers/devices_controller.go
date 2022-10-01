package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/api"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	qm "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DevicesController struct {
	settings        *config.Settings
	dbs             func() *database.DBReaderWriter
	nhtsaSvc        services.INHTSAService
	edmundsSvc      services.EdmundsService
	deviceDefSvc    services.DeviceDefinitionService
	deviceDefIntSvc services.DeviceDefinitionIntegrationService
	log             *zerolog.Logger
}

const autoPiYearCutoff = 2000

// NewDevicesController constructor
func NewDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger, nhtsaSvc services.INHTSAService, ddSvc services.DeviceDefinitionService, ddIntSvc services.DeviceDefinitionIntegrationService) DevicesController {
	edmundsSvc := services.NewEdmundsService(settings.TorProxyURL, logger)

	return DevicesController{
		settings:        settings,
		dbs:             dbs,
		nhtsaSvc:        nhtsaSvc,
		log:             logger,
		edmundsSvc:      edmundsSvc,
		deviceDefSvc:    ddSvc,
		deviceDefIntSvc: ddIntSvc,
	}
}

// GetDeviceDefinitionByID godoc
// @Description gets a specific device definition by id, adds autopi integration on the fly if does not have it and year > cutoff
// @Tags        device-definitions
// @Produce     json
// @Param       id  path     string true "device definition id, KSUID format"
// @Success     200 {object} services.DeviceDefinition
// @Router      /device-definitions/:id [get]
func (d *DevicesController) GetDeviceDefinitionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if len(id) != 27 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "invalid device_definition_id",
		})
	}
	deviceDefinitionResponse, err := d.deviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{id})
	if err != nil {
		return api.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+id)
	}

	if len(deviceDefinitionResponse) == 0 {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("device definition with id %s not found", id))
	}

	dd := deviceDefinitionResponse[0]
	rp, err := NewDeviceDefinitionFromGRPC(dd)
	if err != nil {
		return errors.Wrapf(err, "could not convert device def for api response %+v", dd)
	}
	if dd.Type.Year >= autoPiYearCutoff && !strings.EqualFold(dd.Make.Name, "Tesla") {
		rp.CompatibleIntegrations, err = d.deviceDefIntSvc.AppendAutoPiCompatibility(c.Context(), rp.CompatibleIntegrations, dd.DeviceDefinitionId)
		if err != nil {
			return api.GrpcErrorToFiber(err, fmt.Sprintf("deviceDefIntSvc error when AppendAutoPiCompatibility. dd id: %s", dd.DeviceDefinitionId))
		}
	}
	return c.JSON(fiber.Map{
		"deviceDefinition": rp,
	})
}

// GetDeviceIntegrationsByID godoc
// @Description gets all the available integrations for a device definition. Includes the capabilities of the device with the integration
// @Tags        device-definitions
// @Produce     json
// @Param       id  path     string true "device definition id, KSUID format"
// @Success     200 {object} []services.DeviceCompatibility
// @Router      /device-definitions/{id}/integrations [get]
func (d *DevicesController) GetDeviceIntegrationsByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if len(id) != 27 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "invalid device definition id",
		})
	}
	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load("DeviceIntegrations.Integration")).
		One(c.Context(), d.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no device defintion with id %s found", id))
		}
		return err
	}
	// build object for integrations that have all the info
	var integrations []services.DeviceCompatibility
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			integrations = append(integrations, services.DeviceCompatibility{
				ID:           di.R.Integration.ID,
				Type:         di.R.Integration.Type,
				Style:        di.R.Integration.Style,
				Vendor:       di.R.Integration.Vendor,
				Region:       di.Region,
				Capabilities: jsonOrDefault(di.Capabilities),
			})
		}
	}
	if dd.Year >= autoPiYearCutoff && !strings.EqualFold(dd.R.DeviceMake.Name, "Tesla") {
		integrations, err = d.deviceDefIntSvc.AppendAutoPiCompatibility(c.Context(), integrations, dd.ID)
		if err != nil {
			return api.GrpcErrorToFiber(err, fmt.Sprintf("deviceDefIntSvc error when AppendAutoPiCompatibility. dd id: %s", dd.ID))
		}
	}
	return c.JSON(fiber.Map{
		"compatibleIntegrations": integrations,
	})
}

// GetDeviceDefinitionByMMY godoc
// @Description gets a specific device definition by make model and year
// @Tags        device-definitions
// @Produce     json
// @Param       make  query    string true "make eg TESLA"
// @Param       model query    string true "model eg MODEL Y"
// @Param       year  query    string true "year eg 2021"
// @Success     200   {object} services.DeviceDefinition
// @Router      /device-definitions [get]
func (d *DevicesController) GetDeviceDefinitionByMMY(c *fiber.Ctx) error {
	mk := c.Query("make")
	model := c.Query("model")
	year := c.Query("year")
	if mk == "" || model == "" || year == "" {
		return api.ErrorResponseHandler(c, errors.New("make, model, and year are required"), fiber.StatusBadRequest)
	}
	yrInt, err := strconv.Atoi(year)
	if err != nil {
		return api.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	dd, err := d.deviceDefSvc.FindDeviceDefinitionByMMY(c.Context(), mk, model, yrInt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.ErrorResponseHandler(c, errors.Wrapf(err, "device with %s %s %s not found", mk, model, year), fiber.StatusNotFound)
		}
		return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	rp, err := NewDeviceDefinitionFromGRPC(dd)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"deviceDefinition": rp,
	})
}

func NewDeviceDefinitionFromGRPC(dd *grpc.GetDeviceDefinitionItemResponse) (services.DeviceDefinition, error) {
	if dd.Make == nil {
		return services.DeviceDefinition{}, errors.New("required DeviceMake relation is not set")
	}
	rp := services.DeviceDefinition{
		DeviceDefinitionID:     dd.DeviceDefinitionId,
		Name:                   dd.Name,
		ImageURL:               &dd.ImageUrl,
		CompatibleIntegrations: []services.DeviceCompatibility{},
		DeviceMake: services.DeviceMake{
			ID:              dd.Make.Id,
			Name:            dd.Make.Name,
			LogoURL:         null.StringFrom(dd.Make.LogoUrl),
			OemPlatformName: null.StringFrom(dd.Make.OemPlatformName),
		},
		VehicleInfo: services.DeviceVehicleInfo{
			MPG:                 fmt.Sprintf("%f", dd.VehicleData.MPG),
			MPGHighway:          fmt.Sprintf("%f", dd.VehicleData.MPGHighway),
			MPGCity:             fmt.Sprintf("%f", dd.VehicleData.MPGCity),
			FuelTankCapacityGal: fmt.Sprintf("%f", dd.VehicleData.FuelTankCapacityGal),
			FuelType:            dd.VehicleData.FuelType,
			BaseMSRP:            int(dd.VehicleData.Base_MSRP),
			DrivenWheels:        dd.VehicleData.DrivenWheels,
			NumberOfDoors:       fmt.Sprintf("%d", dd.VehicleData.NumberOfDoors),
			EPAClass:            dd.VehicleData.EPAClass,
			VehicleType:         dd.VehicleData.VehicleType,
		},
		Type: services.DeviceType{
			Type:  dd.Type.Type,
			Make:  dd.Type.Make,
			Model: dd.Type.Model,
			Year:  int(dd.Type.Year),
		},
		//Metadata: dd.Metadata,
		Verified: dd.Verified,
	}
	//// vehicle info
	//var vi map[string]services.DeviceVehicleInfo
	//rp.VehicleInfo = vi[vehicleInfoJSONNode]

	// compatible integrations
	rp.CompatibleIntegrations = DeviceCompatibilityFromDB(dd.CompatibleIntegrations)
	// sub_models
	rp.Type.SubModels = dd.Type.SubModels

	return rp, nil
}

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func DeviceCompatibilityFromDB(dbDIS []*grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations) []services.DeviceCompatibility {
	if len(dbDIS) == 0 {
		return []services.DeviceCompatibility{}
	}
	compatibilities := make([]services.DeviceCompatibility, len(dbDIS))
	for i, di := range dbDIS {
		compatibilities[i] = services.DeviceCompatibility{
			ID:     di.Id,
			Type:   di.Type,
			Style:  di.Style,
			Vendor: di.Vendor,
			Region: di.Region,
			//Capabilities: di.Capabilities,
		}
	}
	return compatibilities
}

// jsonOrDefault returns the raw JSON bytes if there is a value, and otherwise returns the byte
// representation of the empty JSON object {}.
func jsonOrDefault(j null.JSON) json.RawMessage {
	if !j.Valid || len(j.JSON) == 0 {
		return []byte(`{}`)
	}
	return j.JSON
}

type DeviceMMYRoot struct {
	Make   string         `json:"make"`
	Models []DeviceModels `json:"models"`
}

type DeviceModels struct {
	Model string            `json:"model"`
	Years []DeviceModelYear `json:"years"`
}

type DeviceModelYear struct {
	Year               int16  `json:"year"`
	DeviceDefinitionID string `json:"id"`
}
