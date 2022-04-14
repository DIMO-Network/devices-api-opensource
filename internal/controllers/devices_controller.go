package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

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
	Settings     *config.Settings
	DBS          func() *database.DBReaderWriter
	NHTSASvc     services.INHTSAService
	EdmundsSvc   services.EdmundsService
	DeviceDefSvc services.IDeviceDefinitionService
	log          *zerolog.Logger
}

const autoPiYearCutoff = 2000

// NewDevicesController constructor
func NewDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger, nhtsaSvc services.INHTSAService, ddSvc services.IDeviceDefinitionService) DevicesController {
	edmundsSvc := services.NewEdmundsService(settings.TorProxyURL, logger)

	return DevicesController{
		Settings:     settings,
		DBS:          dbs,
		NHTSASvc:     nhtsaSvc,
		log:          logger,
		EdmundsSvc:   edmundsSvc,
		DeviceDefSvc: ddSvc,
	}
}

// GetAllDeviceMakeModelYears godoc
// @Description  returns a json tree of Makes, models, and years
// @Tags           device-definitions
// @Produce      json
// @Success      200  {object}  []controllers.DeviceMMYRoot
// @Router       /device-definitions/all [get]
func (d *DevicesController) GetAllDeviceMakeModelYears(c *fiber.Ctx) error {
	allMakes, err := models.DeviceMakes(qm.OrderBy(models.DeviceMakeColumns.Name)).All(c.Context(), d.DBS().Reader)
	if err != nil {
		return err
	}
	all, err := models.DeviceDefinitions(qm.Where("verified = true"),
		qm.OrderBy("device_make_id, model, year")).All(c.Context(), d.DBS().Reader)

	if err != nil {
		return err
	}
	var mmy []DeviceMMYRoot
	for _, dd := range all {
		makeName := ""
		for _, mk := range allMakes {
			if mk.ID == dd.DeviceMakeID {
				makeName = mk.Name
				break
			}
		}
		idx := indexOfMake(mmy, makeName)
		// append make if not found
		if idx == -1 {
			mmy = append(mmy, DeviceMMYRoot{
				Make:   makeName,
				Models: []DeviceModels{{Model: dd.Model, Years: []DeviceModelYear{{Year: dd.Year, DeviceDefinitionID: dd.ID}}}},
			})
		} else {
			// attach model or year to existing make, lookup model
			idx2 := indexOfModel(mmy[idx].Models, dd.Model)
			if idx2 == -1 {
				// append model if not found
				mmy[idx].Models = append(mmy[idx].Models, DeviceModels{
					Model: dd.Model,
					Years: []DeviceModelYear{{Year: dd.Year, DeviceDefinitionID: dd.ID}},
				})
			} else {
				// make and model already found, just add year
				mmy[idx].Models[idx2].Years = append(mmy[idx].Models[idx2].Years, DeviceModelYear{Year: dd.Year, DeviceDefinitionID: dd.ID})
			}
		}
	}

	return c.JSON(fiber.Map{
		"makes": mmy,
	})
}

// GetDeviceDefinitionByID godoc
// @Description  gets a specific device definition by id
// @Tags           device-definitions
// @Produce      json
// @Param             id        path  string  true  "device definition id, KSUID format"
// @Success      200  {object}  services.DeviceDefinition
// @Router       /device-definitions/{id} [get]
func (d *DevicesController) GetDeviceDefinitionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if len(id) != 27 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "invalid device_definition_id",
		})
	}
	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration))).
		One(c.Context(), d.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("device definition with id %s not found", id))
		}
		return err
	}

	rp, err := NewDeviceDefinitionFromDatabase(dd)
	if err != nil {
		return err
	}
	if dd.Year >= autoPiYearCutoff {
		rp.CompatibleIntegrations, err = services.AppendAutoPiCompatibility(c.Context(), rp.CompatibleIntegrations, d.DBS().Writer)
		if err != nil {
			return err
		}
	}
	return c.JSON(fiber.Map{
		"deviceDefinition": rp,
	})
}

// GetDeviceIntegrationsByID godoc
// @Description  gets all the available integrations for a device definition. Includes the capabilities of the device with the integration
// @Tags           device-definitions
// @Produce      json
// @Param             id        path  string  true  "device definition id, KSUID format"
// @Success      200  {object}  []services.DeviceCompatibility
// @Router       /device-definitions/{id}/integrations [get]
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
		qm.Load("DeviceIntegrations.Integration")).
		One(c.Context(), d.DBS().Reader)
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
	if dd.Year >= autoPiYearCutoff {
		integrations, err = services.AppendAutoPiCompatibility(c.Context(), integrations, d.DBS().Writer)
		if err != nil {
			return err
		}
	}
	return c.JSON(fiber.Map{
		"compatibleIntegrations": integrations,
	})
}

// GetDeviceDefinitionByMMY godoc
// @Description  gets a specific device definition by make model and year
// @Tags           device-definitions
// @Produce      json
// @Param             make      query  string  true  "make eg TESLA"
// @Param             model     query  string  true  "model eg MODEL Y"
// @Param             year      query  string  true  "year eg 2021"
// @Success      200  {object}  services.DeviceDefinition
// @Router       /device-definitions [get]
func (d *DevicesController) GetDeviceDefinitionByMMY(c *fiber.Ctx) error {
	mk := c.Query("make")
	model := c.Query("model")
	year := c.Query("year")
	if mk == "" || model == "" || year == "" {
		return errorResponseHandler(c, errors.New("make, model, and year are required"), fiber.StatusBadRequest)
	}
	yrInt, err := strconv.Atoi(year)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	dd, err := d.DeviceDefSvc.FindDeviceDefinitionByMMY(c.Context(), nil, mk, model, yrInt, true)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, errors.Wrapf(err, "device with %s %s %s not found", mk, model, year), fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	rp, err := NewDeviceDefinitionFromDatabase(dd)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"deviceDefinition": rp,
	})
}

func indexOfMake(makes []DeviceMMYRoot, make string) int {
	for i, root := range makes {
		if root.Make == make {
			return i
		}
	}
	return -1
}
func indexOfModel(models []DeviceModels, model string) int {
	for i, m := range models {
		if m.Model == model {
			return i
		}
	}
	return -1
}

const vehicleInfoJSONNode = "vehicle_info"

func NewDeviceDefinitionFromDatabase(dd *models.DeviceDefinition) (services.DeviceDefinition, error) {
	if dd.R == nil || dd.R.DeviceMake == nil {
		return services.DeviceDefinition{}, errors.New("required DeviceMake relation is not set")
	}
	rp := services.DeviceDefinition{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:               dd.ImageURL.Ptr(),
		CompatibleIntegrations: []services.DeviceCompatibility{},
		Type: services.DeviceType{
			Type:  "Vehicle",
			Make:  dd.R.DeviceMake.Name,
			Model: dd.Model,
			Year:  int(dd.Year),
		},
		Metadata: string(dd.Metadata.JSON),
		Verified: dd.Verified,
	}
	// vehicle info
	var vi map[string]services.DeviceVehicleInfo
	if err := dd.Metadata.Unmarshal(&vi); err == nil {
		rp.VehicleInfo = vi[vehicleInfoJSONNode]
	}
	// relational properties
	if dd.R != nil {
		// compatible integrations
		rp.CompatibleIntegrations = DeviceCompatibilityFromDB(dd.R.DeviceIntegrations)
		// sub_models
		rp.Type.SubModels = services.SubModelsFromStylesDB(dd.R.DeviceStyles)
	}

	return rp, nil
}

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func DeviceCompatibilityFromDB(dbDIS models.DeviceIntegrationSlice) []services.DeviceCompatibility {
	if len(dbDIS) == 0 {
		return []services.DeviceCompatibility{}
	}
	compatibilities := make([]services.DeviceCompatibility, len(dbDIS))
	for i, di := range dbDIS {
		compatibilities[i] = services.DeviceCompatibility{
			ID:           di.IntegrationID,
			Type:         di.R.Integration.Type,
			Style:        di.R.Integration.Style,
			Vendor:       di.R.Integration.Vendor,
			Region:       di.Region,
			Capabilities: jsonOrDefault(di.Capabilities),
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
