package controllers

import (
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/gofiber/fiber/v2"
)

type DevicesController struct {
	// DB holder
	// redis cache?
	Settings *config.Settings
	DBS func() *database.DBReaderWriter
}

func NewDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter) DevicesController {
	return DevicesController{
		Settings: settings,
		DBS: dbs,
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

type DeviceRp struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}
