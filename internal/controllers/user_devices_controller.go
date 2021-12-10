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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user_device_id":           ud.ID,
		"device_definition_id":     dd.ID,
		"integration_capabilities": DeviceCompatibilityFromDB(deviceInts),
	})
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
