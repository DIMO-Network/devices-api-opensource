package controllers

import (
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
	"net/http"
	"testing"
)

func TestDevicesController_GetUsersDevices(t *testing.T) {
	c :=NewDevicesController(&config.Settings{Port: "3000"})

	app := fiber.New()
	app.Get("/devices", c.GetUsersDevices)

	request, _ := http.NewRequest("GET", "/devices", nil)
	response, _ := app.Test(request)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "{\"devices\":[{\"device_id\":\"123123\",\"name\":\"Johnny's Tesla\"}]}", string(body))
}
