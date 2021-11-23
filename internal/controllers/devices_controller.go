package controllers

import (
	"database/sql"
	"fmt"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	qm "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DevicesController struct {
	// DB holder
	// redis cache?
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
}

func NewDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter) DevicesController {
	return DevicesController{
		Settings: settings,
		DBS:      dbs,
	}
}

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

func (d *DevicesController) LookupDeviceDefinitionByVIN(c *fiber.Ctx) error {
	vin := c.Params("vin")
	squishVin := vin[0:9]
	dd, err := models.DeviceDefinitions(qm.Where("vin_first_10 = ?", squishVin)).One(c.Context(), d.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// call out to nhtsa
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error_message": err.Error(),
			})
		}
	}
	rp := DeviceDefinitionRp{
		DeviceDefinitionId: dd.UUID,
		Name:               fmt.Sprintf("%d %s %s", dd.Year, dd.Make, dd.Model),
		ImageURL:           "",
		Compatibility:      DeviceCompatibilityRp{},
		Type: DeviceTypeRp{
			Type:     "Vehicle",
			Make:     dd.Make,
			Model:    dd.Model,
			Year:     dd.Year,
			SubModel: dd.SubModel.String,
		},
		VehicleInfo: DeviceVehicleInfoRp{},
		MetaData:    string(dd.OtherData.JSON),
	}
	return c.JSON(fiber.Map{
		"device_definition": rp,
	})
}

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

type DeviceDefinitionRp struct {
	DeviceDefinitionId string                `json:"device_definition_id"`
	Name               string                `json:"name"`
	ImageURL           string                `json:"image_url"`
	Compatibility      DeviceCompatibilityRp `json:"compatibility"`
	Type               DeviceTypeRp          `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo DeviceVehicleInfoRp `json:"vehicle_data,omitempty"`
	MetaData    interface{}         `json:"meta_data"`
}

// DeviceCompatibilityRp represents what systems we know this is compatible with
type DeviceCompatibilityRp struct {
	IsSmartCarCompatible   bool `json:"is_smart_car_compatible"`
	IsDimoAutoPiCompatible bool `json:"is_dimo_auto_pi_compatible"`
}

// DeviceTypeRp whether it is a vehicle or other type and basic information
type DeviceTypeRp struct {
	// Type is eg. Vehicle, E-bike, roomba
	Type     string `json:"type"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	Year     int16  `json:"year"`
	SubModel string `json:"sub_model"`
}

// DeviceVehicleInfoRp represents some standard vehicle specific properties
type DeviceVehicleInfoRp struct {
	FuelType      string  `json:"fuel_type,omitempty"`
	DrivenWheels  string  `json:"driven_wheels,omitempty"`
	NumberOfDoors string  `json:"number_of_doors,omitempty"`
	BaseMSRP      float32 `json:"base_msrp,omitempty"`
	EPAClass      string  `json:"epa_class,omitempty"`
	VehicleType   string  `json:"vehicle_type,omitempty"`
	MPGHighway    string  `json:"mpg_highway,omitempty"`
	MPGCity       string  `json:"mpg_city,omitempty"`
}
