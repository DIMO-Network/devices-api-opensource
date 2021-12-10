package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func TestUserDevicesController(t *testing.T) {
	// arrange global db and route setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ctx := context.Background()
	pdb, database := setupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := database.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	testUserId := "123123"
	c := NewUserDevicesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger)
	app := fiber.New()
	app.Post("/user/devices", authInjectorTestHandler(testUserId), c.RegisterDeviceForUser)
	app.Get("/user/devices/me", authInjectorTestHandler(testUserId), c.GetUserDevices)

	t.Run("POST - register with device_definition_id", func(t *testing.T) {
		ddID := "123"
		dd := models.DeviceDefinition{
			ID:       ddID,
			Make:     "Tesla",
			Model:    "Model X",
			Year:     2020,
			Verified: true,
		}
		err := dd.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		assert.NoError(t, err, "database error")
		reg := RegisterUserDevice{
			DeviceDefinitionId: &ddID,
		}
		j, _ := json.Marshal(reg)
		request := buildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		udi := gjson.Get(string(body), "user_device_id")
		assert.True(t, udi.Exists(), "expected to find user_device_id")
	})
	t.Run("POST - register with MMY", func(t *testing.T) {
		mk := "Tesla"
		model := "Model"
		year := 2021
		reg := RegisterUserDevice{
			Make:  &mk,
			Model: &model,
			Year:  &year,
		}
		j, _ := json.Marshal(reg)
		request := buildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		udi := gjson.Get(string(body), "user_device_id")
		assert.True(t, udi.Exists(), "expected to find user_device_id")
	})
	t.Run("POST - bad payload", func(t *testing.T) {
		request := buildRequest("POST", "/user/devices", "{}")
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
		msg := gjson.Get(string(body), "error_message").String()
		assert.Contains(t, msg, "cannot be blank")
	})
	t.Run("POST - bad device_definition_id", func(t *testing.T) {
		ddID := "caca"
		reg := RegisterUserDevice{
			DeviceDefinitionId: &ddID,
		}
		j, _ := json.Marshal(reg)
		request := buildRequest("POST", "/user/devices", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode)
		msg := gjson.Get(string(body), "error_message").String()
		fmt.Println("message: " + msg)
		assert.Contains(t, msg, "caca")
	})
}
