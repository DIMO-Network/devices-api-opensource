package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/test"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	saramamocks "github.com/Shopify/sarama/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type partialFenceCloudEvent struct {
	Data struct {
		H3Indexes []string `json:"h3Indexes"`
	} `json:"data"`
}

func checkForDeviceAndH3(userDeviceID string, h3Indexes []string) func(*sarama.ProducerMessage) error {
	return func(msg *sarama.ProducerMessage) error {
		kb, _ := msg.Key.Encode()
		if string(kb) != userDeviceID {
			return fmt.Errorf("expected message to be keyed with %s but got %s", userDeviceID, string(kb))
		}

		if len(h3Indexes) == 0 {
			if msg.Value != nil {
				return fmt.Errorf("non-nil body when nil was expected")
			}
			return nil
		}

		ev := new(partialFenceCloudEvent)
		vb, _ := msg.Value.Encode()
		if err := json.Unmarshal(vb, ev); err != nil {
			return err
		}
		if len(ev.Data.H3Indexes) != len(h3Indexes) {
			return fmt.Errorf("expected %d H3 indices but got %d", len(h3Indexes), len(ev.Data.H3Indexes))
		}

		set := shared.NewStringSet()
		for _, ind := range h3Indexes {
			set.Add(ind)
		}

		for _, ind := range ev.Data.H3Indexes {
			if !set.Contains(ind) {
				return fmt.Errorf("message contained unexpected H3 index %s", ind)
			}
		}

		return nil
	}
}

func TestGeofencesController(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()

	ctx := context.Background()
	pdb, db := test.SetupDatabase(ctx, t, migrationsDirRelPath)
	defer func() {
		if err := db.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	testUserID := "123123"
	otherUserID := "7734"
	producer := saramamocks.NewSyncProducer(t, sarama.NewConfig())
	c := NewGeofencesController(&config.Settings{Port: "3000"}, pdb.DBS, &logger, producer)
	app := fiber.New()
	app.Post("/user/geofences", test.AuthInjectorTestHandler(testUserID), c.Create)
	app.Get("/user/geofences", test.AuthInjectorTestHandler(testUserID), c.GetAll)
	app.Put("/user/geofences/:geofenceID", test.AuthInjectorTestHandler(testUserID), c.Update)
	app.Delete("/user/geofences/:geofenceID", test.AuthInjectorTestHandler(testUserID), c.Delete)
	// test data
	dm := models.DeviceMake{
		ID:   ksuid.New().String(),
		Name: "Mercedes-Benz",
	}
	_ = dm.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	deviceDef := models.DeviceDefinition{
		ID:           ksuid.New().String(),
		DeviceMakeID: dm.ID,
		Model:        "C300",
		Year:         2009,
		Verified:     true,
	}
	_ = deviceDef.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	userDevice := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             testUserID,
		DeviceDefinitionID: deviceDef.ID,
		Name:               null.StringFrom("chungus"),
		CountryCode:        null.StringFrom("USA"),
	}
	_ = userDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	otherUserDevice := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             otherUserID,
		DeviceDefinitionID: deviceDef.ID,
		Name:               null.StringFrom("hugh mungus"),
		CountryCode:        null.StringFrom("USA"),
	}
	_ = otherUserDevice.Insert(ctx, pdb.DBS().Writer, boil.Infer())

	createdID := ""
	t.Run("POST - create geofence", func(t *testing.T) {
		req := CreateGeofence{
			Name:          "Home",
			Type:          "PrivacyFence",
			H3Indexes:     []string{"123", "321"},
			UserDeviceIDs: []string{userDevice.ID},
		}
		j, _ := json.Marshal(req)

		producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(userDevice.ID, []string{"123", "321"}))

		request := test.BuildRequest("POST", "/user/geofences", string(j))
		response, _ := app.Test(request)
		body, _ := ioutil.ReadAll(response.Body)
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode) == false {
			fmt.Println("message: " + string(body))
		}
		createdID = gjson.Get(string(body), "id").String()
		assert.Len(t, createdID, 27)

		producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(userDevice.ID, []string{"123", "321"}))

		// create one without h3 indexes required
		req = CreateGeofence{
			Name:          "Work",
			Type:          "PrivacyFence",
			UserDeviceIDs: []string{userDevice.ID},
		}
		j, _ = json.Marshal(req)
		request = test.BuildRequest("POST", "/user/geofences", string(j))
		response, _ = app.Test(request)
		if assert.Equal(t, fiber.StatusCreated, response.StatusCode, "expected create OK without h3 indexes") == false {
			body, _ = ioutil.ReadAll(response.Body)
			fmt.Println("message: " + string(body))
		}
	})
	t.Run("POST - 400 if same name", func(t *testing.T) {
		req := CreateGeofence{
			Name:          "Home",
			Type:          "PrivacyFence",
			UserDeviceIDs: []string{userDevice.ID},
		}
		j, _ := json.Marshal(req)
		request := test.BuildRequest("POST", "/user/geofences", string(j))
		response, _ := app.Test(request)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode, "expected bad request on duplicate name")
	})
	t.Run("POST - 400 if not your device", func(t *testing.T) {
		req := CreateGeofence{
			Name:          "Home",
			Type:          "PrivacyFence",
			UserDeviceIDs: []string{otherUserDevice.ID},
		}
		j, _ := json.Marshal(req)
		request := test.BuildRequest("POST", "/user/geofences", string(j))
		response, _ := app.Test(request)
		assert.Equal(t, fiber.StatusBadRequest, response.StatusCode, "expected bad request when trying to attach a fence to a device that isn't ours")
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
		assert.Len(t, get.Array(), 2, "expected to find one item in response")
	})
	t.Run("PUT - update a geofence", func(t *testing.T) {
		// The fence is being detached from the device and it has type TriggerEntry anyway.
		producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(userDevice.ID, []string{}))

		req := CreateGeofence{
			Name:          "School",
			Type:          "TriggerEntry",
			H3Indexes:     []string{"123", "321", "1234555"},
			UserDeviceIDs: nil,
		}
		j, _ := json.Marshal(req)
		request := test.BuildRequest("PUT", "/user/geofences/"+createdID, string(j))
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
		// assert against second item in array, which was the created one
		assert.Equal(t, req.Name, get[1].Get("name").String(), "expected name to be updated")
		assert.Equal(t, req.Type, get[1].Get("type").String(), "expected type to be updated")
		assert.Len(t, get[1].Get("h3Indexes").Array(), 3)
	})
	t.Run("DELETE - delete the  geofence by id", func(t *testing.T) {
		producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(checkForDeviceAndH3(userDevice.ID, []string{}))

		request, _ := http.NewRequest("DELETE", "/user/geofences/"+createdID, nil)
		response, _ := app.Test(request)
		// assert
		assert.Equal(t, fiber.StatusNoContent, response.StatusCode)
	})

	_ = producer.Close()
}
