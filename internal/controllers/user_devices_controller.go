package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-INC/devices-api/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/tidwall/sjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserDevicesController struct {
	Settings     *config.Settings
	DBS          func() *database.DBReaderWriter
	DeviceDefSvc services.IDeviceDefinitionService
	log          *zerolog.Logger
	taskSvc      services.ITaskService
	eventService services.EventService
}

// NewUserDevicesController constructor
func NewUserDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger, ddSvc services.IDeviceDefinitionService, taskSvc services.ITaskService, eventService services.EventService) UserDevicesController {
	return UserDevicesController{
		Settings:     settings,
		DBS:          dbs,
		log:          logger,
		DeviceDefSvc: ddSvc,
		taskSvc:      taskSvc,
		eventService: eventService,
	}
}

// GetUserDevices godoc
// @Description gets all devices associated with current user - pulled from token
// @Tags 	user-devices
// @Produce json
// @Success 200 {object} []controllers.UserDeviceFull
// @Security BearerAuth
// @Router  /user/devices/me [get]
func (udc *UserDevicesController) GetUserDevices(c *fiber.Ctx) error {
	userID := getUserID(c)
	devices, err := models.UserDevices(qm.Where("user_id = ?", userID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load("DeviceDefinition.DeviceIntegrations"),
		qm.Load("DeviceDefinition.DeviceIntegrations.Integration"),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.OrderBy("created_at"),
	).
		All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	rp := make([]UserDeviceFull, len(devices))
	for i, d := range devices {
		rp[i] = UserDeviceFull{
			ID:               d.ID,
			VIN:              d.VinIdentifier.Ptr(),
			VINConfirmed:     d.VinConfirmed,
			Name:             d.Name.Ptr(),
			CustomImageURL:   d.CustomImageURL.Ptr(),
			CountryCode:      d.CountryCode.Ptr(),
			DeviceDefinition: NewDeviceDefinitionFromDatabase(d.R.DeviceDefinition),
			Integrations:     NewUserDeviceIntegrationStatusesFromDatabase(d.R.UserDeviceAPIIntegrations),
		}
	}

	return c.JSON(fiber.Map{
		"userDevices": rp,
	})
}

func NewUserDeviceIntegrationStatusesFromDatabase(udis []*models.UserDeviceAPIIntegration) []UserDeviceIntegrationStatus {
	out := make([]UserDeviceIntegrationStatus, len(udis))

	for i, udi := range udis {
		out[i] = UserDeviceIntegrationStatus{
			IntegrationID: udi.IntegrationID,
			Status:        udi.Status,
		}
	}

	return out
}

type userDeviceEvent struct {
	Timestamp time.Time             `json:"timestamp"`
	UserID    string                `json:"userId"`
	Device    userDeviceEventDevice `json:"device"`
}

type userDeviceEventDevice struct {
	ID    string `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

type userDeviceEventIntegration struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Style  string `json:"style"`
	Vendor string `json:"vendor"`
}

type userDeviceIntegrationEvent struct {
	Timestamp   time.Time                  `json:"timestamp"`
	UserID      string                     `json:"userId"`
	Device      userDeviceEventDevice      `json:"device"`
	Integration userDeviceEventIntegration `json:"integration"`
}

// RegisterDeviceForUser godoc
// @Description adds a device to a user. can add with only device_definition_id or with MMY, which will create a device_definition on the fly
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param user_device body controllers.RegisterUserDevice true "add device to user. either MMY or id are required"
// @Security ApiKeyAuth
// @Success 201 {object} controllers.RegisterUserDeviceResponse
// @Security BearerAuth
// @Router  /user/devices [post]
func (udc *UserDevicesController) RegisterDeviceForUser(c *fiber.Ctx) error {
	userID := getUserID(c)
	reg := &RegisterUserDevice{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := reg.Validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	tx, err := udc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return err
	}
	var dd *models.DeviceDefinition
	if reg.DeviceDefinitionID != nil {
		// attach device def to user
		dd, err = models.FindDeviceDefinition(c.Context(), tx, *reg.DeviceDefinitionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errorResponseHandler(c, errors.Wrapf(err, "could not find device definition id: %s", *reg.DeviceDefinitionID), fiber.StatusBadRequest)
			}
			return errorResponseHandler(c, errors.Wrapf(err, "error querying for device definition id: %s", *reg.DeviceDefinitionID), fiber.StatusInternalServerError)
		}
	} else {
		// check for existing MMY
		dd, err = udc.DeviceDefSvc.FindDeviceDefinitionByMMY(c.Context(), tx, *reg.Make, *reg.Model, *reg.Year, false)
		if dd == nil {
			// since Definition does not exist, create one on the fly with userID as source and not verified
			dd = &models.DeviceDefinition{
				ID:       ksuid.New().String(),
				Make:     strings.ToUpper(*reg.Make),
				Model:    strings.ToUpper(*reg.Model),
				Year:     int16(*reg.Year),
				Source:   null.StringFrom("userID:" + userID),
				Verified: false,
			}
			err = dd.Insert(c.Context(), tx, boil.Infer())
		}
		if err != nil {
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}
	}
	userDeviceID := ksuid.New().String()
	// register device for the user
	ud := models.UserDevice{
		ID:                 userDeviceID,
		UserID:             userID,
		DeviceDefinitionID: dd.ID,
		CountryCode:        null.StringFromPtr(reg.CountryCode),
	}
	err = ud.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, errors.Wrapf(err, "could not create user device for def_id: %s", dd.ID), fiber.StatusInternalServerError)
	}
	// get device integrations to return in payload - helps frontend
	deviceInts, err := models.DeviceIntegrations(qm.Load(models.DeviceIntegrationRels.Integration),
		qm.Where("device_definition_id = ?", dd.ID)).
		All(c.Context(), tx)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	err = tx.Commit()
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	// don't block, as image fetch could take a while
	go func() {
		err := udc.DeviceDefSvc.CheckAndSetImage(dd, false)
		if err != nil {
			udc.log.Error().Err(err).Msg("error getting device image upon user_device registration")
			return
		}
		_, err = dd.Update(context.Background(), udc.DBS().Writer, boil.Whitelist("image_url", "updated_at")) // only update image_url https://github.com/volatiletech/sqlboiler#update
		if err != nil {
			udc.log.Error().Err(err).Msg("error updating device image in DB for: " + dd.ID)
		}
	}()

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.create",
		Subject: userID,
		Source:  "devices-api",
		Data: userDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: userDeviceEventDevice{
				ID:    userDeviceID,
				Make:  dd.Make,
				Model: dd.Model,
				Year:  int(dd.Year), // Odd.
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed emitting device creation event")
	}

	return c.Status(fiber.StatusCreated).JSON(
		RegisterUserDeviceResponse{
			UserDeviceID:            ud.ID,
			DeviceDefinitionID:      dd.ID,
			IntegrationCapabilities: DeviceCompatibilityFromDB(deviceInts),
		})
}

type RegisterSmartcarRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirectURI"`
}

var smartcarScopes = []string{
	"read_engine_oil",
	"read_battery",
	"read_charge",
	"control_charge",
	"read_fuel",
	"read_location",
	"read_odometer",
	"read_tires",
	"read_vehicle_info",
	"read_vin",
}

// RegisterSmartcarIntegration godoc
// @Description Use a Smartcar auth code to connect to Smartcar and obtain access and refresh
// @Description tokens for use by the app.
// @Tags user-devices
// @Accept json
// @Param userDeviceIntegrationRegistration body controllers.RegisterSmartcarRequest true "Authorization code from Smartcar"
// @Success 204
// @Router /user/devices/:userDeviceID/integrations/:integrationID [post]
func (udc *UserDevicesController) RegisterSmartcarIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("integrationId", integrationID).
		Str("handler", "RegisterSmartcarIntegration").
		Logger()
	logger.Info().Msg("Attempting to register Smartcar integration")

	reqBody := RegisterSmartcarRequest{}
	if err := c.BodyParser(&reqBody); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	ud, err := models.UserDevices(
		qm.Where("id = ?", userDeviceID),
		qm.Where("user_id = ?", userID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, errors.Wrapf(err, "could not find user_device with id %s for user %s", userDeviceID, userID), fiber.StatusBadRequest)
		}
		logger.Err(err).Msg("Unexpected database error searching for user device")
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	integ, err := models.DeviceIntegrations(
		qm.Where("device_definition_id = ?", ud.DeviceDefinitionID),
		qm.And("integration_id = ?", integrationID),
		qm.And("country = ?", ud.CountryCode),
		qm.Load("Integration")).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn().Msg("Attempted to register a device integration that didn't exist")
			return errorResponseHandler(c,
				errors.Wrapf(err, "could not find device_integrations with device_definition_id %s, integration_id %s and country %s", ud.DeviceDefinitionID, integrationID, ud.CountryCode.String), fiber.StatusBadRequest)
		}
		logger.Err(err).Msg("Unexpected database error searching for device integration")
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	// This is the only integration we currently support. It is generated by the sync script in
	// /cmd/devices-api/load_smartcar.go with a random identifier.
	if integ.R.Integration.Type != "API" || integ.R.Integration.Vendor != "SmartCar" {
		logger.Warn().Msg("Attempted to register a non-Smartcar integration")
		return errorResponseHandler(c, errors.New("could not find SmartCar integration relation"), fiber.StatusBadRequest)
	}

	client := smartcar.NewClient() // Unclear whether we need one of these at the top level
	auth := client.NewAuth(&smartcar.AuthParams{
		ClientID:     udc.Settings.SmartcarClientID,
		ClientSecret: udc.Settings.SmartcarClientSecret,
		RedirectURI:  reqBody.RedirectURI,
		Scope:        smartcarScopes,
		TestMode:     udc.Settings.SmartcarTestMode,
	})
	token, err := auth.ExchangeCode(c.Context(), &smartcar.ExchangeCodeParams{Code: reqBody.Code})
	if err != nil {
		logger.Err(err).Msg("Error exchanging authorization code with Smartcar")
		return errorResponseHandler(c, errors.Wrap(err, "failure exchanging code with SmartCar"), fiber.StatusBadRequest)
	}

	// TODO: Probably replace this ugly block with an upsert later.
	temp, err := models.FindUserDeviceAPIIntegration(c.Context(), udc.DBS().Writer, userDeviceID, integrationID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Err(err).Msg("Unexpected database error looking for existing instance of integration")
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}
	} else {
		// This is mostly helpful for testing
		_, err := temp.Delete(c.Context(), udc.DBS().Writer)
		if err != nil {
			logger.Err(err).Msg("Unexpected database error deleting old instance of integration")
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}
	}

	// TODO: Encrypt the tokens. Note that you need the client id, client secret, and redirect
	// URL to make use of the tokens, but plain text is still a bad idea.
	integration := models.UserDeviceAPIIntegration{
		UserDeviceID:     userDeviceID,
		IntegrationID:    integrationID,
		Status:           models.UserDeviceAPIIntegrationStatusPending,
		AccessToken:      token.Access,
		AccessExpiresAt:  token.AccessExpiry,
		RefreshToken:     token.Refresh,
		RefreshExpiresAt: token.RefreshExpiry,
	}

	err = integration.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		logger.Err(err).Msg("Unexpected database error inserting new Smartcar integration registration")
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = udc.taskSvc.StartSmartcarRegistrationTasks(userDeviceID, integrationID)
	if err != nil {
		logger.Err(err).Msg("Unexpected error starting Smartcar Machinery tasks")
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.create",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: userDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: userDeviceEventDevice{
				ID:    userDeviceID,
				Make:  ud.R.DeviceDefinition.Make,
				Model: ud.R.DeviceDefinition.Model,
				Year:  int(ud.R.DeviceDefinition.Year),
			},
			Integration: userDeviceEventIntegration{
				ID:     integ.R.Integration.ID,
				Type:   integ.R.Integration.Type,
				Style:  integ.R.Integration.Style,
				Vendor: integ.R.Integration.Vendor,
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Failed sending device integration creation event")
	}

	logger.Info().Msg("Finished registering Smartcar integration")
	return c.SendStatus(fiber.StatusNoContent)
}

type GetUserDeviceIntegrationResponse struct {
	// Status is one of "Pending", "PendingFirstData", "Active"
	Status string `json:"status"`
	// ExternalID is the identifier used by the third party for the device. It may be absent if we
	// haven't authorized yet.
	ExternalID null.String `json:"externalId" swaggertype:"string"`
}

// GetUserDeviceIntegration godoc
// @Description Receive status updates about a Smartcar integration
// @Tags user-devices
// @Success 200 {object} controllers.GetUserDeviceIntegrationResponse
// @Router /user/devices/:userDeviceID/integrations/:integrationID [get]
func (udc *UserDevicesController) GetUserDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")
	deviceExists, err := models.UserDevices(
		qm.Where("user_id = ?", userID),
		qm.And("id = ?", userDeviceID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	if !deviceExists {
		return errorResponseHandler(c, fmt.Errorf("no user device with ID %s", userDeviceID), fiber.StatusBadRequest)
	}

	apiIntegration, err := models.UserDeviceAPIIntegrations(
		qm.Where("user_device_id = ?", userDeviceID),
		qm.Where("integration_id = ?", integrationID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, fmt.Errorf("user device %s does not have integration %s", userDeviceID, integrationID), fiber.StatusBadRequest)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	return c.JSON(GetUserDeviceIntegrationResponse{Status: apiIntegration.Status, ExternalID: apiIntegration.ExternalID})
}

// DeleteUserDeviceIntegration godoc
// @Description Remove an user device's integration
// @Tags user-devices
// @Success 204
// @Router /user/devices/:userDeviceID/integrations/:integrationID [delete]
func (udc *UserDevicesController) DeleteUserDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	defer tx.Rollback() //nolint

	device, err := models.UserDevices(
		qm.Where("user_id = ?", userID),
		qm.And("id = ?", userDeviceID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, fmt.Errorf("no user device with ID %s", userDeviceID), fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	// Probably don't need two queries if you're smart
	apiIntegration, err := models.UserDeviceAPIIntegrations(
		qm.Where("user_device_id = ?", userDeviceID),
		qm.Where("integration_id = ?", integrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.Integration),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, fmt.Errorf("user device %s does not have integration %s", userDeviceID, integrationID), fiber.StatusBadRequest)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = udc.taskSvc.StartSmartcarDeregistrationTasks(userDeviceID, integrationID, apiIntegration.ExternalID.String, apiIntegration.AccessToken)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	_, err = apiIntegration.Delete(c.Context(), tx)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.delete",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: userDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: userDeviceEventDevice{
				ID:    userDeviceID,
				Make:  device.R.DeviceDefinition.Make,
				Model: device.R.DeviceDefinition.Model,
				Year:  int(device.R.DeviceDefinition.Year),
			},
			Integration: userDeviceEventIntegration{
				ID:     apiIntegration.R.Integration.ID,
				Type:   apiIntegration.R.Integration.Type,
				Style:  apiIntegration.R.Integration.Style,
				Vendor: apiIntegration.R.Integration.Vendor,
			},
		},
	})
	if err != nil {
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateVIN godoc
// @Description updates the VIN on the user device record
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param vin body controllers.UpdateVINReq true "VIN"
// @Param userDeviceID path string true "user id"
// @Success 204
// @Security BearerAuth
// @Router  /user/devices/:userDeviceID/vin [patch]
func (udc *UserDevicesController) UpdateVIN(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(qm.Where("id = ?", udi), qm.And("user_id = ?", userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	if userDevice.VinConfirmed || userDevice.VinIdentifier.Ptr() != nil && userDevice.CountryCode.String == "USA" {
		return errorResponseHandler(c, errors.New("VIN cannot be changed at this point"), fiber.StatusBadRequest)
	}
	vin := &UpdateVINReq{}
	if err := c.BodyParser(vin); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	if err := vin.validate(); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}
	userDevice.VinIdentifier = null.StringFromPtr(vin.VIN)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateName godoc
// @Description updates the Name on the user device record
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param name body controllers.UpdateNameReq true "Name"
// @Param user_device_id path string true "user id"
// @Success 204
// @Security BearerAuth
// @Router  /user/devices/:userDeviceID/name [patch]
func (udc *UserDevicesController) UpdateName(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(qm.Where("id = ?", udi), qm.And("user_id = ?", userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	name := &UpdateNameReq{}
	if err := c.BodyParser(name); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.Name = null.StringFromPtr(name.Name)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetUserDeviceStatus godoc
// @Description Returns the latest status update for the device. May return 404 if the
// @Description user does not have a device with the ID, or if no status updates have come
// @Tags user-devices
// @Produce json
// @Param user_device_id path string true "user device ID"
// @Success 200
// @Security BearerAuth
// @Router  /user/devices/:userDeviceID/status [get]
func (udc *UserDevicesController) GetUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceDatum),
	).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	if userDevice.R.UserDeviceDatum == nil || !userDevice.R.UserDeviceDatum.Data.Valid {
		return errorResponseHandler(c, errors.New("no status updates yet"), fiber.StatusNotFound)
	}
	// date formatting defaults to encoding/json
	json, _ := sjson.Set(string(userDevice.R.UserDeviceDatum.Data.JSON), "recordUpdatedAt", userDevice.R.UserDeviceDatum.UpdatedAt)
	json, _ = sjson.Set(json, "recordCreatedAt", userDevice.R.UserDeviceDatum.CreatedAt)

	c.Set("Content-Type", "application/json")
	return c.Send([]byte(json))
}

// RefreshUserDeviceStatus godoc
// @Description Starts the process of refreshing device status from Smartcar
// @Tags user-devices
// @Param user_device_id path string true "user device ID"
// @Success 204
// @Failure 429 "rate limit hit for integration"
// @Security BearerAuth
// @Router  /user/devices/:userDeviceID/commands/refresh [post]
func (udc *UserDevicesController) RefreshUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	// We could probably do a smarter join here, but it's unclear to me how to handle that
	// in SQLBoiler.
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.UserDeviceDatum),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationRels.Integration)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	// note: the UserDeviceDatum is not tied to the integration table

	for _, devInteg := range ud.R.UserDeviceAPIIntegrations {
		if devInteg.R.Integration.Type == models.IntegrationTypeAPI && devInteg.R.Integration.Vendor == "SmartCar" && devInteg.Status == models.UserDeviceAPIIntegrationStatusActive {
			if ud.R.UserDeviceDatum != nil {
				nextAvailableTime := ud.R.UserDeviceDatum.UpdatedAt.Add(time.Second * time.Duration(devInteg.R.Integration.RefreshLimitSecs))
				if time.Now().Before(nextAvailableTime) {
					return errorResponseHandler(c, errors.New("rate limit for integration refresh hit"), fiber.StatusTooManyRequests)
				}
			}
			err = udc.taskSvc.StartSmartcarRefresh(udi, devInteg.R.Integration.ID)
			if err != nil {
				return errorResponseHandler(c, err, fiber.StatusInternalServerError)
			}
			return c.SendStatus(204)
		}
	}

	return errorResponseHandler(c, errors.New("no active Smartcar integration found for this device"), fiber.StatusBadRequest)
}

// UpdateCountryCode godoc
// @Description updates the CountryCode on the user device record
// @Tags 	user-devices
// @Produce json
// @Accept json
// @Param name body controllers.UpdateCountryCodeReq true "Country code"
// @Success 204
// @Security BearerAuth
// @Router  /user/devices/:userDeviceID/country_code [patch]
func (udc *UserDevicesController) UpdateCountryCode(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(qm.Where("id = ?", udi), qm.And("user_id = ?", userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	countryCode := &UpdateCountryCodeReq{}
	if err := c.BodyParser(countryCode); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.CountryCode = null.StringFromPtr(countryCode.CountryCode)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteUserDevice godoc
// @Description delete the user device record (hard delete)
// @Tags 	user-devices
// @Param userDeviceID path string true "user id"
// @Success 204
// @Security BearerAuth
// @Router  /user/devices/:userDeviceID [delete]
func (udc *UserDevicesController) DeleteUserDevice(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint
	userDevice, err := models.UserDevices(
		qm.Where("id = ?", udi),
		qm.And("user_id = ?", userID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	for _, apiInteg := range userDevice.R.UserDeviceAPIIntegrations {
		// For now, there are only Smartcar integrations. We will probably regret this
		// line later.
		err = udc.taskSvc.StartSmartcarDeregistrationTasks(udi, apiInteg.IntegrationID, apiInteg.ExternalID.String, apiInteg.AccessToken)
		if err != nil {
			return errorResponseHandler(c, err, fiber.StatusInternalServerError)
		}
	}

	_, err = userDevice.Delete(c.Context(), tx)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	dd := userDevice.R.DeviceDefinition
	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.delete",
		Subject: userID,
		Source:  "devices-api",
		Data: userDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: userDeviceEventDevice{
				ID:    udi,
				Make:  dd.Make,
				Model: dd.Model,
				Year:  int(dd.Year), // Odd.
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed emitting device deletion event")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type RegisterUserDevice struct {
	Make               *string `json:"make"`
	Model              *string `json:"model"`
	Year               *int    `json:"year"`
	DeviceDefinitionID *string `json:"deviceDefinitionId"`
	CountryCode        *string `json:"countryCode"`
}

type RegisterUserDeviceResponse struct {
	UserDeviceID            string                         `json:"userDeviceId"`
	DeviceDefinitionID      string                         `json:"deviceDefinitionId"`
	IntegrationCapabilities []services.DeviceCompatibility `json:"integrationCapabilities"`
}

type AdminRegisterUserDevice struct {
	RegisterUserDevice
	ID          string  `json:"id"`          // KSUID from client,
	CreatedDate int64   `json:"createdDate"` // unix timestamp
	VehicleName *string `json:"vehicleName"`
	VIN         string  `json:"vin"`
	ImageURL    *string `json:"imageUrl"`
	Verified    bool    `json:"verified"`
}

type UpdateVINReq struct {
	VIN *string `json:"vin"`
}

type UpdateNameReq struct {
	Name *string `json:"name"`
}

type UpdateCountryCodeReq struct {
	CountryCode *string `json:"countryCode"`
}

func (reg *RegisterUserDevice) Validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.Make, validation.When(reg.DeviceDefinitionID == nil, validation.Required)),
		validation.Field(&reg.Model, validation.When(reg.DeviceDefinitionID == nil, validation.Required)),
		validation.Field(&reg.Year, validation.When(reg.DeviceDefinitionID == nil, validation.Required)),
		validation.Field(&reg.DeviceDefinitionID, validation.When(reg.Make == nil && reg.Model == nil && reg.Year == nil, validation.Required)),
		validation.Field(&reg.CountryCode, validation.When(reg.CountryCode != nil, validation.Length(3, 3))),
	)
}

func (reg *AdminRegisterUserDevice) Validate() error {
	return validation.ValidateStruct(reg,
		validation.Field(&reg.RegisterUserDevice),
		validation.Field(&reg.ID, validation.Required, validation.Length(27, 27), is.Alphanumeric),
	)
}

func (u *UpdateVINReq) validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.VIN, validation.Required, validation.Length(17, 17)))
}

// UserDeviceFull represents object user's see on frontend for listing of their devices
type UserDeviceFull struct {
	ID               string                        `json:"id"`
	VIN              *string                       `json:"vin"`
	VINConfirmed     bool                          `json:"vinConfirmed"`
	Name             *string                       `json:"name"`
	CustomImageURL   *string                       `json:"customImageUrl"`
	DeviceDefinition services.DeviceDefinition     `json:"deviceDefinition"`
	CountryCode      *string                       `json:"countryCode"`
	Integrations     []UserDeviceIntegrationStatus `json:"integrations"`
}

type UserDeviceIntegrationStatus struct {
	IntegrationID string `json:"integrationId"`
	Status        string `json:"status"`
}
