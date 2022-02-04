package controllers

import (
	"fmt"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GeofencesController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	log      *zerolog.Logger
}

// NewGeofencesController constructor
func NewGeofencesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) GeofencesController {
	return GeofencesController{
		Settings: settings,
		DBS:      dbs,
		log:      logger,
	}
}

// Create godoc
// @Description adds a new geofence to the user's account, optionally attached to specific user_devices
// @Tags 	geofence
// @Produce json
// @Accept json
// @Param geofence body controllers.CreateGeofence true "add geofence to user."
// @Success 201 {object} controllers.CreateResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router  /user/geofences [post]
func (g *GeofencesController) Create(c *fiber.Ctx) error {
	userID := getUserID(c)
	create := CreateGeofence{}
	if err := c.BodyParser(&create); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := create.Validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	tx, err := g.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return err
	}

	// check if already exists
	exists, err := models.Geofences(models.GeofenceWhere.UserID.EQ(userID), models.GeofenceWhere.Name.EQ(create.Name)).Exists(c.Context(), tx)
	if err != nil {
		return err
	}
	if exists {
		return errorResponseHandler(c, errors.New("Geofence with that name already exists for this user"), fiber.StatusBadRequest)
	}
	geofence := models.Geofence{
		ID:        ksuid.New().String(),
		UserID:    userID,
		Name:      create.Name,
		Type:      create.Type,
		H3Indexes: create.H3Indexes,
	}
	err = geofence.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "error inserting geofence")
	}
	for _, uID := range create.UserDeviceIDs {
		geoToUser := models.UserDeviceToGeofence{
			UserDeviceID: uID,
			GeofenceID:   geofence.ID,
		}
		err = geoToUser.Upsert(c.Context(), tx, true, []string{"user_device_id", "geofence_id"}, boil.Infer(), boil.Infer())
		if err != nil {
			return errors.Wrapf(err, "error upserting user_device_to_geofence")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "error commiting transaction to create geofence")
	}

	return c.Status(fiber.StatusCreated).JSON(CreateResponse{ID: geofence.ID})
}

// GetAll godoc
// @Description gets all geofences for the current user
// @Tags 	geofence
// @Produce json
// @Success 200 {object} []controllers.GetGeofence
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router  /user/geofences [get]
func (g *GeofencesController) GetAll(c *fiber.Ctx) error {
	userID := getUserID(c)
	//could not find LoadUserDevices method for eager loading
	items, err := models.Geofences(models.GeofenceWhere.UserID.EQ(userID),
		qm.Load(models.GeofenceRels.UserDeviceToGeofences),
		qm.Load("UserDeviceToGeofences.UserDevice"),
		qm.Load("UserDeviceToGeofences.UserDevice.DeviceDefinition")).All(c.Context(), g.DBS().Reader)
	if err != nil {
		return err
	}
	fences := make([]GetGeofence, len(items))
	for i, item := range items {
		f := GetGeofence{
			ID:        item.ID,
			Name:      item.Name,
			Type:      item.Type,
			H3Indexes: item.H3Indexes,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		for _, udtg := range item.R.UserDeviceToGeofences {
			f.UserDevices = append(f.UserDevices, GeoFenceUserDevice{
				UserDeviceID: udtg.UserDeviceID,
				Name:         udtg.R.UserDevice.Name.Ptr(),
				MMY: fmt.Sprintf("%d %s %s", udtg.R.UserDevice.R.DeviceDefinition.Year,
					udtg.R.UserDevice.R.DeviceDefinition.Make,
					udtg.R.UserDevice.R.DeviceDefinition.Model),
			})
		}
		fences[i] = f
	}

	return c.JSON(fiber.Map{
		"geofences": fences,
	})
}

// Update godoc
// @Description updates an existing geofence for the current user
// @Tags geofence
// @Produce json
// @Accept json
// @Param geofenceID path string true "geofence id"
// @Param geofence body controllers.CreateGeofence true "add geofence to user."
// @Success 204
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /user/geofences/:geofenceID [put]
func (g *GeofencesController) Update(c *fiber.Ctx) error {
	userID := getUserID(c)
	id := c.Params("geofenceID")
	update := CreateGeofence{}
	if err := c.BodyParser(&update); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := update.Validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	tx, err := g.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return err
	}
	// Return status 400 and error message.
	geofence, err := models.Geofences(models.GeofenceWhere.UserID.EQ(userID), models.GeofenceWhere.ID.EQ(id),
		qm.Load(models.GeofenceRels.UserDeviceToGeofences)).One(c.Context(), tx)
	if err != nil {
		return err
	}
	geofence.Name = update.Name
	geofence.Type = update.Type
	geofence.H3Indexes = update.H3Indexes
	_, err = geofence.Update(c.Context(), tx, boil.Whitelist(
		models.GeofenceColumns.Name,
		models.GeofenceColumns.Type,
		models.GeofenceColumns.H3Indexes,
		models.GeofenceColumns.UpdatedAt))
	if err != nil {
		return errors.Wrap(err, "error updating geofence")
	}
	for _, uID := range update.UserDeviceIDs {
		geoToUser := models.UserDeviceToGeofence{
			UserDeviceID: uID,
			GeofenceID:   geofence.ID,
		}
		err = geoToUser.Upsert(c.Context(), tx, true,
			[]string{models.UserDeviceToGeofenceColumns.UserDeviceID, models.UserDeviceToGeofenceColumns.GeofenceID}, boil.Infer(), boil.Infer())
		if err != nil {
			return errors.Wrapf(err, "error upserting user_device_to_geofence")
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "error commiting transaction to create geofence")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Delete godoc
// @Description hard deletes a geofence from db
// @Tags geofence
// @Param geofenceID path string true "geofence id"
// @Success 204
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /user/geofences/:geofenceID [delete]
func (g *GeofencesController) Delete(c *fiber.Ctx) error {
	userID := getUserID(c)
	id := c.Params("geofenceID")
	_, err := models.Geofences(models.GeofenceWhere.UserID.EQ(userID), models.GeofenceWhere.ID.EQ(id)).DeleteAll(c.Context(), g.DBS().Writer)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type GetGeofence struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	H3Indexes   []string             `json:"h3Indexes"`
	UserDevices []GeoFenceUserDevice `json:"userDevices"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

type GeoFenceUserDevice struct {
	UserDeviceID string  `json:"userDeviceId"`
	Name         *string `json:"name"`
	MMY          string  `json:"mmy"`
}

type CreateGeofence struct {
	// required: true
	Name string `json:"name"`
	// one of following: "PrivacyFence", "TriggerEntry", "TriggerExit"
	// required: true
	Type string `json:"type"`
	// required: false
	H3Indexes []string `json:"h3Indexes"`
	// Optionally link the geofence with a list of user device ID
	UserDeviceIDs []string `json:"userDeviceIds"`
}

func (g *CreateGeofence) Validate() error {
	return validation.ValidateStruct(g,
		validation.Field(&g.Name, validation.Required),
		validation.Field(&g.Type, validation.Required, validation.In("PrivacyFence", "TriggerEntry", "TriggerExit")),
	)
}
