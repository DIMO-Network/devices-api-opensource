package controllers

import (
	"database/sql"
	"strings"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserDevicesController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	log      *zerolog.Logger
}

// NewUserDevicesController constructor
func NewUserDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) UserDevicesController {
	return UserDevicesController{
		Settings: settings,
		DBS:      dbs,
		log:      logger,
	}
}

// GetUserDevices godoc
// @Description gets all devices associated with current user - pulled from token
// @Tags 	user-devices
// @Produce json
// @Success 200 {object} []controllers.UserDeviceFull
// @Security BearerAuth
// @Router  /user/devices/me [get]
func (udc *UserDevicesController) GetUserDevices(c *fiber.Ctx) error {
	userID := getUserID(c)
	devices, err := models.UserDevices(qm.Where("user_id = ?", userID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load("DeviceDefinition.DeviceIntegrations"),
		qm.Load("DeviceDefinition.DeviceIntegrations.Integration"),
	).
		All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	rp := make([]UserDeviceFull, len(devices))
	for i, d := range devices {
		rp[i] = UserDeviceFull{
			ID:               d.ID,
			VIN:              d.VinIdentifier.String,
			Name:             d.Name.String,
			CustomImageURL:   d.CustomImageURL.String,
			CountryCode:      d.CountryCode.String,
			DeviceDefinition: NewDeviceDefinitionFromDatabase(d.R.DeviceDefinition),
		}
	}

	return c.JSON(fiber.Map{
		"user_devices": rp,
	})
}

// RegisterDeviceForUser godoc
// @Description adds a device to a user. can add with only device_definition_id or with MMY, which will create a device_definition on the fly
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param user_device body controllers.RegisterUserDevice true "add device to user. either MMY or id are required"
// @Security ApiKeyAuth
// @Success 201 {object} controllers.RegisterUserDeviceResponse
// @Security BearerAuth
// @Router  /user/devices [post]
func (udc *UserDevicesController) RegisterDeviceForUser(c *fiber.Ctx) error {
	userID := getUserID(c)
	reg := &RegisterUserDevice{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := reg.validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	tx, err := udc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return err
	}
	var dd *models.DeviceDefinition
	if reg.DeviceDefinitionID != nil {
		// attach device def to user
		dd, err = models.FindDeviceDefinition(c.Context(), tx, *reg.DeviceDefinitionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errorResponseHandler(c, errors.Wrapf(err, "could not find device definition id: %s", *reg.DeviceDefinitionID), fiber.StatusBadRequest)
			}
			return errorResponseHandler(c, errors.Wrapf(err, "error querying for device definition id: %s", *reg.DeviceDefinitionID), fiber.StatusInternalServerError)
		}
		exists, err := models.UserDevices(qm.Where("user_id = ?", userID), qm.And("device_definition_id = ?", dd.ID)).Exists(c.Context(), tx)
		if err != nil {
			return errorResponseHandler(c, errors.Wrap(err, "error checking duplicate user device"), fiber.StatusInternalServerError)
		}
		if exists {
			return errorResponseHandler(c, errors.Wrap(err, "user already has this device registered"), fiber.StatusBadRequest)
		}
	} else {
		// check for existing MMY
		dd, err = models.DeviceDefinitions(
			qm.Where("make = ?", strings.ToUpper(*reg.Make)),
			qm.And("model = ?", strings.ToUpper(*reg.Model)),
			qm.And("year = ?", *reg.Year)).
			One(c.Context(), tx)
		if dd == nil {
			// since Definition does not exist, create one on the fly with userID as source and not verified
			dd = &models.DeviceDefinition{
				ID:       ksuid.New().String(),
				Make:     strings.ToUpper(*reg.Make),
				Model:    strings.ToUpper(*reg.Model),
				Year:     int16(*reg.Year),
				Source:   null.StringFrom("userID:" + userID),
				Verified: false,
			}
			err = dd.Insert(c.Context(), tx, boil.Infer())
		}
		if err != nil {
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}
	}
	// register device for the user
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             userID,
		DeviceDefinitionID: dd.ID,
		CountryCode:        null.StringFromPtr(reg.CountryCode),
	}
	err = ud.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, errors.Wrapf(err, "could not create user device for def_id: %s", dd.ID), fiber.StatusInternalServerError)
	}
	// get device integrations to return in payload - helps frontend
	deviceInts, err := models.DeviceIntegrations(qm.Load(models.DeviceIntegrationRels.Integration),
		qm.Where("device_definition_id = ?", dd.ID)).
		All(c.Context(), tx)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	err = tx.Commit()
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusCreated).JSON(
		RegisterUserDeviceResponse{
			UserDeviceID:            ud.ID,
			DeviceDefinitionID:      dd.ID,
			IntegrationCapabilities: DeviceCompatibilityFromDB(deviceInts),
		})
}

func (udc *UserDevicesController) RegisterSmartCarIntegration(c *fiber.Ctx) error {
	return nil
}

// AdminRegisterUserDevice godoc
// @Description meant for internal admin use - adds a device to a user. can add with only device_definition_id or with MMY, which will create a device_definition on the fly
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param user_device body controllers.AdminRegisterUserDevice true "add device to user. either MMY or id are required"
// @Param user_id path string true "user id"
// @Success 201 {object} controllers.RegisterUserDeviceResponse
// @Router  /admin/user/:user_id/devices [post]
func (udc *UserDevicesController) AdminRegisterUserDevice(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	reg := &AdminRegisterUserDevice{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := reg.validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	tx, err := udc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return err
	}
	var dd *models.DeviceDefinition
	if reg.DeviceDefinitionID != nil {
		// attach device def to user
		dd, err = models.FindDeviceDefinition(c.Context(), tx, *reg.DeviceDefinitionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errorResponseHandler(c, errors.Wrapf(err, "could not find device definition id: %s", *reg.DeviceDefinitionID), fiber.StatusBadRequest)
			}
			return errorResponseHandler(c, errors.Wrapf(err, "error querying for device definition id: %s", *reg.DeviceDefinitionID), fiber.StatusInternalServerError)
		}
		exists, err := models.UserDevices(qm.Where("user_id = ?", userID), qm.And("device_definition_id = ?", dd.ID)).Exists(c.Context(), tx)
		if err != nil {
			return errorResponseHandler(c, errors.Wrap(err, "error checking duplicate user device"), fiber.StatusInternalServerError)
		}
		if exists {
			return errorResponseHandler(c, errors.Wrap(err, "user already has this device registered"), fiber.StatusBadRequest)
		}
	} else {
		// lookup existing MMY
		dd, err = models.DeviceDefinitions(
			qm.Where("make = ?", strings.ToUpper(*reg.Make)),
			qm.And("model = ?", strings.ToUpper(*reg.Model)),
			qm.And("year = ?", *reg.Year)).
			One(c.Context(), tx)
		if dd == nil {
			// since Definition does not exist, create one on the fly with userID as source and not verified
			dd = &models.DeviceDefinition{
				ID:       ksuid.New().String(),
				Make:     *reg.Make,
				Model:    *reg.Model,
				Year:     int16(*reg.Year),
				Source:   null.StringFrom("userID:" + userID),
				Verified: reg.Verified,
				ImageURL: null.StringFromPtr(reg.ImageURL),
			}
			if len(reg.VIN) == 17 {
				dd.VinFirst10 = null.StringFrom(reg.VIN[:10])
			}
			err = dd.Insert(c.Context(), tx, boil.Infer())
		}
		if err != nil {
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}
	}
	// register device for the user
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             userID,
		DeviceDefinitionID: dd.ID,
		CountryCode:        null.StringFromPtr(reg.CountryCode),
		Name:               null.StringFromPtr(reg.VehicleName),
		CreatedAt:          time.Unix(reg.CreatedDate, 0),
	}
	if len(reg.VIN) == 17 {
		ud.VinIdentifier = null.StringFrom(reg.VIN)
	}
	err = ud.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, errors.Wrapf(err, "could not create user device for def_id: %s", dd.ID), fiber.StatusInternalServerError)
	}
	// get device integrations to return in payload - helps frontend
	deviceInts, err := models.DeviceIntegrations(qm.Load(models.DeviceIntegrationRels.Integration),
		qm.Where("device_definition_id = ?", dd.ID)).
		All(c.Context(), tx)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	err = tx.Commit()
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusCreated).JSON(
		RegisterUserDeviceResponse{
			UserDeviceID:            ud.ID,
			DeviceDefinitionID:      dd.ID,
			IntegrationCapabilities: DeviceCompatibilityFromDB(deviceInts),
		})
}

// UpdateVIN godoc
// @Description updates the VIN on the user device record
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param vin body controllers.UpdateVINReq true "VIN"
// @Success 204
// @Security BearerAuth
// @Router  /user/devices/:user_device_id/vin [patch]
func (udc *UserDevicesController) UpdateVIN(c *fiber.Ctx) error {
	udi := c.Params("user_device_id")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(qm.Where("id = ?", udi), qm.And("user_id = ?", userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	if userDevice.VinIdentifier.Ptr() != nil && userDevice.CountryCode.String == "USA" {
		return errorResponseHandler(c, errors.New("VIN cannot be changed at this point"), fiber.StatusBadRequest)
	}
	vin := &UpdateVINReq{}
	if err := c.BodyParser(vin); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := vin.validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	userDevice.VinIdentifier = null.StringFromPtr(vin.VIN)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateName godoc
// @Description updates the Name on the user device record
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param name body controllers.UpdateNameReq true "Name"
// @Success 204
// @Security BearerAuth
// @Router  /user/devices/:user_device_id/name [patch]
func (udc *UserDevicesController) UpdateName(c *fiber.Ctx) error {
	udi := c.Params("user_device_id")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(qm.Where("id = ?", udi), qm.And("user_id = ?", userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	name := &UpdateNameReq{}
	if err := c.BodyParser(name); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.Name = null.StringFromPtr(name.Name)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateCountryCode godoc
// @Description updates the CountryCode on the user device record
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param name body controllers.UpdateCountryCodeReq true "Country code"
// @Success 204
// @Security BearerAuth
// @Router  /user/devices/:user_device_id/country_code [patch]
func (udc *UserDevicesController) UpdateCountryCode(c *fiber.Ctx) error {
	udi := c.Params("user_device_id")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(qm.Where("id = ?", udi), qm.And("user_id = ?", userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	countryCode := &UpdateCountryCodeReq{}
	if err := c.BodyParser(countryCode); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.CountryCode = null.StringFromPtr(countryCode.CountryCode)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type RegisterUserDevice struct {
	Make               *string `json:"make"`
	Model              *string `json:"model"`
	Year               *int    `json:"year"`
	DeviceDefinitionID *string `json:"device_definition_id"`
	CountryCode        *string `json:"country_code"`
}

type RegisterUserDeviceResponse struct {
	UserDeviceID            string                `json:"user_device_id"`
	DeviceDefinitionID      string                `json:"device_definition_id"`
	IntegrationCapabilities []DeviceCompatibility `json:"integration_capabilities"`
}

type AdminRegisterUserDevice struct {
	RegisterUserDevice
	CreatedDate int64   `json:"created_date"` // unix timestamp
	VehicleName *string `json:"vehicle_name"`
	VIN         string  `json:"vin"`
	ImageURL    *string `json:"image_url"`
	Verified    bool    `json:"verified"`
}

type UpdateVINReq struct {
	VIN *string `json:"vin"`
}

type UpdateNameReq struct {
	Name *string `json:"name"`
}

type UpdateCountryCodeReq struct {
	CountryCode *string `json:"countryCode"`
}

func (reg *RegisterUserDevice) validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.Make, validation.When(reg.DeviceDefinitionID == nil, validation.Required)),
		validation.Field(&reg.Model, validation.When(reg.DeviceDefinitionID == nil, validation.Required)),
		validation.Field(&reg.Year, validation.When(reg.DeviceDefinitionID == nil, validation.Required)),
		validation.Field(&reg.DeviceDefinitionID, validation.When(reg.Make == nil && reg.Model == nil && reg.Year == nil, validation.Required)),
		validation.Field(&reg.CountryCode, validation.When(reg.CountryCode != nil, validation.Length(3, 3))),
	)
}

func (u *UpdateVINReq) validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.VIN, validation.Required, validation.Length(17, 17)))
}

// UserDeviceFull represents object user's see on frontend for listing of their devices
type UserDeviceFull struct {
	ID               string           `json:"id"`
	VIN              string           `json:"vin"`
	Name             string           `json:"name"`
	CustomImageURL   string           `json:"custom_image_url"`
	DeviceDefinition DeviceDefinition `json:"device_definition"`
	CountryCode      string           `json:"country_code"`
}
