package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestGeofencesController(t *testing.T) {
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
	testUserID := "123123"
	c := NewGeofencesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger)
	app := fiber.New()
	app.Post("/user/geofences", authInjectorTestHandler(testUserID), c.Create)
	app.Get("/user/geofences", authInjectorTestHandler(testUserID), c.GetAll)
	app.Put("/user/geofences/:geofenceID", authInjectorTestHandler(testUserID), c.Update)
	app.Delete("/user/geofences/:geofenceID", authInjectorTestHandler(testUserID), c.Delete)

	createdID := ""
	t.Run("POST - create geofence", func(t *testing.T) {
		req := CreateGeofence{
			Name:          "Home",
			Type:          "PrivacyFence",
			H3Indexes:     []string{"123", "321"},
			UserDeviceIDs: nil,
		}
		j, _ := json.Marshal(req)
		request := buildRequest("POST", "/user/geofences", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		createdID = gjson.Get(string(body), "id").String()
		assert.Len(t, createdID, 27)
	})
	t.Run("GET - get all geofences for user", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/user/geofences", nil)
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		// assert
		assert.Equal(t, fiber.StatusOK, response.StatusCode)
		get := gjson.Get(string(body), "geofences")
		if assert.True(t, get.IsArray()) == false {
			fmt.Println("body: " + string(body))
		}
		assert.Len(t, get.Array(), 1, "expected to find one item in response")
	})
	t.Run("PUT - update a geofence", func(t *testing.T) {
		req := CreateGeofence{
			Name:          "Work",
			Type:          "TriggerEntry",
			H3Indexes:     []string{"123", "321", "1234555"},
			UserDeviceIDs: nil,
		}
		j, _ := json.Marshal(req)
		request := buildRequest("PUT", "/user/geofences/"+createdID, string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		if assert.Equal(t, fiber.StatusNoContent, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
			fmt.Println("id: " + createdID)
		}
		// validate update was performed
		request, _ = http.NewRequest("GET", "/user/geofences", nil)
		response, _ = app.Test(request)
		body, _ = ioutil.ReadAll(response.Body)
		// assert changes
		assert.Equal(t, fiber.StatusOK, response.StatusCode)
		get := gjson.Get(string(body), "geofences").Array()
		assert.Equal(t, req.Name, get[0].Get("name").String(), "expected name to be updated")
		assert.Equal(t, req.Type, get[0].Get("type").String(), "expected type to be updated")
		assert.Len(t, get[0].Get("h3Indexes").Array(), 3)
	})
	t.Run("DELETE - delete the  geofence by id", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/user/geofences/"+createdID, nil)
		response, _ := app.Test(request)
		// assert
		assert.Equal(t, fiber.StatusNoContent, response.StatusCode)
	})
}
