package controllers

import (
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UserDevicesController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	log      *zerolog.Logger
}

func (udc *UserDevicesController) GetUserDevices(c *fiber.Ctx) error {
	return nil
}

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
	if reg.DeviceDefinitionId != nil {
		// attach device def to user
		dd, err := models.FindDeviceDefinition(c.Context(), tx, *reg.DeviceDefinitionId)
		if err != nil {
			return errorResponseHandler(c, errors.Wrap(err, "could not find provided device definition id"), fiber.StatusBadRequest)
		}
		// todo: check that device def doesn't already exist for
		ud := models.UserDevice{
			ID:                 ksuid.New().String(),
			UserID:             userId,
			DeviceDefinitionID: dd.ID,
		}
		err = ud.Insert(c.Context(), tx, boil.Infer())
		if err != nil {
			return errorResponseHandler(c, errors.Wrap(err, "could not create device"), fiber.StatusInternalServerError)
		}

		return c.JSON(fiber.Map{
			"device_id": ud.ID,
			"device_definition_id": dd.ID,
			"integration_capabilities": "array of integrations for device dev", // todo
		})
	}
	// todo: handle else case of creating a Device def on the fly, previous lookup

	return nil
}

func (udc *UserDevicesController) RegisterSmartCarIntegration(c *fiber.Ctx) error {
	return nil
}

// NewUserDevicesController constructor
func NewUserDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) UserDevicesController {
	return UserDevicesController{
		Settings: settings,
		DBS:      dbs,
		log:      logger,
	}
}

type RegisterUserDevice struct {
	Make               *string `json:"make"`
	Model              *string `json:"model"`
	Year               *int    `json:"year"`
	DeviceDefinitionId *string `json:"device_definition_id"`
}

func (reg *RegisterUserDevice) validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.Make, validation.When(reg.DeviceDefinitionId == nil, validation.Required)),
		validation.Field(&reg.Model, validation.When(reg.DeviceDefinitionId == nil, validation.Required)),
		validation.Field(&reg.Year, validation.When(reg.DeviceDefinitionId == nil, validation.Required)),
		validation.Field(&reg.DeviceDefinitionId, validation.When(reg.Make == nil && reg.Model == nil && reg.Year == nil, validation.Required)),
	)
}
