package controllers

import (
	"database/sql"
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
	userId := getUserId(c)
	devices, err := models.UserDevices(qm.Where("user_id = ?", userId),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load("DeviceDefinition.DeviceIntegrations"),
		qm.Load("DeviceDefinition.DeviceIntegrations.Integration"),
	).
		All(c.Context(), udc.DBS().Reader)
	//qm.Load("DeviceIntegrations.Integration")
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	var rp []UserDeviceFull
	for _, d := range devices {
		rp = append(rp, UserDeviceFull{
			ID:               d.ID,
			VIN:              d.VinIdentifier.String,
			Name:             d.Name.String,
			CustomImageUrl:   d.CustomImageURL.String,
			Region:           d.Region.String,
			DeviceDefinition: NewDeviceDefinitionFromDatabase(d.R.DeviceDefinition),
		})
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
// @Success 200 {object} controllers.RegisterUserDeviceResponse
// @Security BearerAuth
// @Router  /user/devices [post]
func (udc *UserDevicesController) RegisterDeviceForUser(c *fiber.Ctx) error {
	userId := getUserId(c)
	reg := &RegisterUserDevice{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := reg.validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	tx, err := udc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback()
	if err != nil {
		return err
	}
	var dd *models.DeviceDefinition
	if reg.DeviceDefinitionId != nil {
		// attach device def to user
		dd, err = models.FindDeviceDefinition(c.Context(), tx, *reg.DeviceDefinitionId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errorResponseHandler(c, errors.Wrapf(err, "could not find device definition id: %s", *reg.DeviceDefinitionId), fiber.StatusBadRequest)
			}
			return errorResponseHandler(c, errors.Wrapf(err, "error querying for device definition id: %s", *reg.DeviceDefinitionId), fiber.StatusInternalServerError)
		}
		exists, err := models.UserDevices(qm.Where("user_id = ?", userId), qm.And("device_definition_id = ?", dd.ID)).Exists(c.Context(), tx)
		if err != nil {
			return errorResponseHandler(c, errors.Wrap(err, "error checking duplicate user device"), fiber.StatusInternalServerError)
		}
		if exists {
			return errorResponseHandler(c, errors.Wrap(err, "user already has this device registered"), fiber.StatusBadRequest)
		}
	} else {
		// since Definition does not exist, create one on the fly with userId as source and not verified
		dd = &models.DeviceDefinition{
			ID:     ksuid.New().String(),
			Make:   *reg.Make,
			Model:  *reg.Model,
			Year:   int16(*reg.Year),
			Source: null.StringFrom("userId:" + userId),
		}
		err = dd.Insert(c.Context(), tx, boil.Infer())
		if err != nil {
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}
	}
	// register device for the user
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             userId,
		DeviceDefinitionID: dd.ID,
		Region:             null.StringFromPtr(reg.Region),
	}
	err = ud.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, errors.Wrapf(err, "could not create user device for def_id: %s", dd.ID), fiber.StatusInternalServerError)
	}
	// get device integrations to return in payload - helps frontend
	deviceInts, err := models.DeviceIntegrations(qm.Load(models.DeviceIntegrationRels.Integration), qm.Where("device_definition_id = ?", dd.ID)).All(c.Context(), tx)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	err = tx.Commit()
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusCreated).JSON(
		RegisterUserDeviceResponse{
			UserDeviceId:            ud.ID,
			DeviceDefinitionId:      dd.ID,
			IntegrationCapabilities: DeviceCompatibilityFromDB(deviceInts),
		})
}

func (udc *UserDevicesController) RegisterSmartCarIntegration(c *fiber.Ctx) error {
	return nil
}

type RegisterUserDevice struct {
	Make               *string `json:"make"`
	Model              *string `json:"model"`
	Year               *int    `json:"year"`
	DeviceDefinitionId *string `json:"device_definition_id"`
	Region             *string `json:"region"`
}

type RegisterUserDeviceResponse struct {
	UserDeviceId            string                `json:"user_device_id"`
	DeviceDefinitionId      string                `json:"device_definition_id"`
	IntegrationCapabilities []DeviceCompatibility `json:"integration_capabilities"`
}

func (reg *RegisterUserDevice) validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.Make, validation.When(reg.DeviceDefinitionId == nil, validation.Required)),
		validation.Field(&reg.Model, validation.When(reg.DeviceDefinitionId == nil, validation.Required)),
		validation.Field(&reg.Year, validation.When(reg.DeviceDefinitionId == nil, validation.Required)),
		validation.Field(&reg.DeviceDefinitionId, validation.When(reg.Make == nil && reg.Model == nil && reg.Year == nil, validation.Required)),
	)
}

// UserDeviceFull represents object user's see on frontend for listing of their devices
type UserDeviceFull struct {
	ID               string           `json:"id"`
	VIN              string           `json:"vin"`
	Name             string           `json:"name"`
	CustomImageUrl   string           `json:"custom_image_url"`
	DeviceDefinition DeviceDefinition `json:"device_definition"`
	Region           string           `json:"region"`
}
