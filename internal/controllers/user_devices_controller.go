package controllers

import (
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
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
