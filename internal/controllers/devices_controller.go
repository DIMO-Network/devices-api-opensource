package controllers

import (
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/gofiber/fiber/v2"
)

type DevicesController struct {
	// DB holder
	// redis cache?
	Settings *config.Settings
}

func NewDevicesController(settings *config.Settings) DevicesController {
	return DevicesController{
		Settings: settings,
	}
}

func (d *DevicesController) GetUsersDevices(c *fiber.Ctx) error {
	ds := make([]DeviceRp, 0)
	ds = append(ds, DeviceRp{
		DeviceId: "123123",
		Name:     "Johnny's Tesla",
	})

	return c.JSON(fiber.Map{
		"devices": ds,
	})
}

type DeviceRp struct {
	DeviceId string `json:"device_id"`
	Name     string `json:"name"`
}
