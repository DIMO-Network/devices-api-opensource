package controllers

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	qm "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DevicesController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	NHTSASvc services.INHTSAService
	log      *zerolog.Logger
}

// NewDevicesController constructor
func NewDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger, nhtsaSvc services.INHTSAService) DevicesController {
	return DevicesController{
		Settings: settings,
		DBS:      dbs,
		NHTSASvc: nhtsaSvc,
		log:      logger,
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
	dd, err := models.DeviceDefinitions(qm.Where("vin_first_10 = ?", squishVin),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load("DeviceIntegrations.Integrations")).
		One(c.Context(), d.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			decodedVIN, err := d.NHTSASvc.DecodeVIN(vin)
			if err != nil {
				return errorResponseHandler(c, err, fiber.StatusInternalServerError)
			}
			rp := NewDeviceDefinitionFromNHTSA(decodedVIN)
			// save to database, if error just log do not block, execute in go func routine to not block
			go func() {
				dbDevice := NewDbModelFromDeviceDefinition(rp, squishVin)
				err := dbDevice.Insert(c.Context(), d.DBS().Writer, boil.Infer())
				if err != nil {
					d.log.Error().Err(err).Msg("error inserting device definition to db")
				}
			}()
			return c.JSON(fiber.Map{
				"device_definition": rp,
			})
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	rp := NewDeviceDefinitionFromDatabase(dd)
	return c.JSON(fiber.Map{
		"device_definition": rp,
	})
}

const vehicleInfoJSONNode = "vehicle_info"

func NewDeviceDefinitionFromDatabase(dd *models.DeviceDefinition) DeviceDefinition {
	rp := DeviceDefinition{
		DeviceDefinitionID: dd.UUID,
		Name:               fmt.Sprintf("%d %s %s", dd.Year, dd.Make, dd.Model),
		ImageURL:           "",
		Compatibility:      []DeviceCompatibility{},
		Type: DeviceType{
			Type:     "Vehicle",
			Make:     dd.Make,
			Model:    dd.Model,
			Year:     int(dd.Year),
			SubModel: dd.SubModel.String,
		},
		Metadata: string(dd.Metadata.JSON),
	}
	// vehicle info
	var vi map[string]DeviceVehicleInfo
	err := dd.Metadata.Unmarshal(&vi)
	if err == nil {
		rp.VehicleInfo = vi[vehicleInfoJSONNode]
	}
	// compatible integrations
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			rp.Compatibility = append(rp.Compatibility, DeviceCompatibility{
				Id:      di.R.Integration.UUID,
				Type:    di.R.Integration.Type,
				Style:   di.R.Integration.Style,
				Vendors: di.R.Integration.Vendors,
			})
		}
	}

	return rp
}

// NewDbModelFromDeviceDefinition converts a DeviceDefinition response object to a new database model for the given squishVin
func NewDbModelFromDeviceDefinition(dd DeviceDefinition, squishVin string) *models.DeviceDefinition {
	dbDevice := models.DeviceDefinition{
		VinFirst10: squishVin,
		Make:       dd.Type.Make,
		Model:      dd.Type.Model,
		Year:       int16(dd.Type.Year),
		SubModel:   null.StringFrom(dd.Type.SubModel),
	}
	_ = dbDevice.Metadata.Marshal(map[string]interface{}{vehicleInfoJSONNode: dd.VehicleInfo})
	// next: figure out how we store compatibility

	return &dbDevice
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
		Type:  "Vehicle",
		Make:  decodedVin.LookupValue("Make"),
		Model: decodedVin.LookupValue("Model"),
		Year:  yr,
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
	DeviceDefinitionID string              `json:"device_definition_id"`
	Name               string              `json:"name"`
	ImageURL           string              `json:"image_url"`
	// Compatibility has systems this vehicle can integrate with
	Compatibility      []DeviceCompatibility `json:"compatibility"`
	Type               DeviceType          `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo DeviceVehicleInfo `json:"vehicle_data,omitempty"`
	Metadata    interface{}       `json:"metadata"`
}

// DeviceCompatibility represents what systems we know this is compatible with
type DeviceCompatibility struct {
	Id string `json:"id"`
	Type string `json:"type"`
	Style string `json:"style"`
	Vendors string `json:"vendors"`
}

// DeviceType whether it is a vehicle or other type and basic information
type DeviceType struct {
	// Type is eg. Vehicle, E-bike, roomba
	Type     string `json:"type"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	SubModel string `json:"sub_model"`
}

// DeviceVehicleInfo represents some standard vehicle specific properties
type DeviceVehicleInfo struct {
	FuelType      string `json:"fuel_type,omitempty"`
	DrivenWheels  string `json:"driven_wheels,omitempty"`
	NumberOfDoors string `json:"number_of_doors,omitempty"`
	BaseMSRP      int    `json:"base_msrp,omitempty"`
	EPAClass      string `json:"epa_class,omitempty"`
	VehicleType   string `json:"vehicle_type,omitempty"`
	MPGHighway    string `json:"mpg_highway,omitempty"`
	MPGCity       string `json:"mpg_city,omitempty"`
}
