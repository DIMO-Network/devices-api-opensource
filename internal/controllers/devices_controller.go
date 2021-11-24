package controllers

import (
	"database/sql"
	"fmt"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	qm "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strconv"
)

type DevicesController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	NHTSASvc services.INHTSAService
}

// NewDevicesController constructor
func NewDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter) DevicesController {
	return DevicesController{
		Settings: settings,
		DBS:      dbs,
		NHTSASvc: services.NewNHTSAService(),
	}
}

// GetUsersDevices placeholder for endpoint to get devices that belong to a user
func (d *DevicesController) GetUsersDevices(c *fiber.Ctx) error {
	ds := make([]DeviceRp, 0)
	ds = append(ds, DeviceRp{
		DeviceID: "123123",
		Name:     "Johnny's Tesla",
	})

	return c.JSON(fiber.Map{
		"devices": ds,
	})
}

// LookupDeviceDefinitionByVIN decodes a VIN by first looking it up on our DB, and then calling out to external sources. If it does call out, it will backfill our DB
func (d *DevicesController) LookupDeviceDefinitionByVIN(c *fiber.Ctx) error {
	vin := c.Params("vin")
	if len(vin) != 17 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error_message": "vin is not 17 characters",
		})
	}
	squishVin := vin[:10]
	dd, err := models.DeviceDefinitions(qm.Where("vin_first_10 = ?", squishVin)).One(c.Context(), d.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			decodedVIN, err := d.NHTSASvc.DecodeVIN(vin)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error_message": err.Error(),
				})
			}
			rp := NewDeviceDefinitionFromNHTSA(decodedVIN)
			// todo: persist in our db
			return c.JSON(fiber.Map{
				"device_definition": rp,
			})
		} else {
			// todo: refactor error handling
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error_message": err.Error(),
			})
		}
	}
	rp := NewDeviceDefinitionFromDatabase(dd)
	return c.JSON(fiber.Map{
		"device_definition": rp,
	})
}

func NewDeviceDefinitionFromDatabase(dd *models.DeviceDefinition) DeviceDefinition {
	rp := DeviceDefinition{
		DeviceDefinitionId: dd.UUID,
		Name:               fmt.Sprintf("%d %s %s", dd.Year, dd.Make, dd.Model),
		ImageURL:           "",
		Compatibility:      DeviceCompatibility{},
		Type: DeviceType{
			Type:     "Vehicle",
			Make:     dd.Make,
			Model:    dd.Model,
			Year:     int(dd.Year),
			SubModel: dd.SubModel.String,
		},
		VehicleInfo: DeviceVehicleInfo{},
		MetaData:    string(dd.OtherData.JSON),
	}
	return rp
}

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

func NewDeviceDefinitionFromNHTSA(decodedVin *services.NHTSADecodeVINResponse) DeviceDefinition {
	dd := DeviceDefinition{}
	yr, _ := strconv.Atoi(decodedVin.LookupValue("Model Year"))
	msrp, _ := strconv.Atoi(decodedVin.LookupValue("Base Price ($)"))
	dd.Type = DeviceType{
		Type:     "Vehicle",
		Make:     decodedVin.LookupValue("Make"),
		Model:    decodedVin.LookupValue("Model"),
		Year:     yr,
	}
	dd.Name = fmt.Sprintf("%d %s %s", dd.Type.Year, dd.Type.Make, dd.Type.Model)
	dd.VehicleInfo = DeviceVehicleInfo{
		FuelType:      decodedVin.LookupValue("Fuel Type - Primary"),
		NumberOfDoors: decodedVin.LookupValue("Doors"),
		BaseMSRP:      msrp,
		VehicleType:   decodedVin.LookupValue("Vehicle Type"),
	}

	return dd
}

type DeviceDefinition struct {
	DeviceDefinitionId string              `json:"device_definition_id"`
	Name               string              `json:"name"`
	ImageURL           string              `json:"image_url"`
	Compatibility DeviceCompatibility `json:"compatibility"`
	Type          DeviceType          `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo DeviceVehicleInfo `json:"vehicle_data,omitempty"`
	MetaData    interface{}       `json:"meta_data"`
}

// DeviceCompatibility represents what systems we know this is compatible with
type DeviceCompatibility struct {
	IsSmartCarCompatible   bool `json:"is_smart_car_compatible"`
	IsDimoAutoPiCompatible bool `json:"is_dimo_auto_pi_compatible"`
}

// DeviceType whether it is a vehicle or other type and basic information
type DeviceType struct {
	// Type is eg. Vehicle, E-bike, roomba
	Type     string `json:"type"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	Year     int  `json:"year"`
	SubModel string `json:"sub_model"`
}

// DeviceVehicleInfo represents some standard vehicle specific properties
type DeviceVehicleInfo struct {
	FuelType      string  `json:"fuel_type,omitempty"`
	DrivenWheels  string  `json:"driven_wheels,omitempty"`
	NumberOfDoors string  `json:"number_of_doors,omitempty"`
	BaseMSRP      int `json:"base_msrp,omitempty"`
	EPAClass      string  `json:"epa_class,omitempty"`
	VehicleType   string  `json:"vehicle_type,omitempty"`
	MPGHighway    string  `json:"mpg_highway,omitempty"`
	MPGCity       string  `json:"mpg_city,omitempty"`
}
