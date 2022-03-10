package controllers

import (
	"context"
	"database/sql"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserDevicesController struct {
	Settings         *config.Settings
	DBS              func() *database.DBReaderWriter
	DeviceDefSvc     services.IDeviceDefinitionService
	log              *zerolog.Logger
	taskSvc          services.ITaskService
	eventService     services.EventService
	smartcarClient   services.SmartcarClient
	teslaService     services.TeslaService
	teslaTaskService services.TeslaTaskService
	encrypter        services.Encrypter
}

// NewUserDevicesController constructor
func NewUserDevicesController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger, ddSvc services.IDeviceDefinitionService, taskSvc services.ITaskService, eventService services.EventService, smartcarClient services.SmartcarClient, teslaSvc services.TeslaService, teslaTaskService services.TeslaTaskService, encrypter services.Encrypter) UserDevicesController {
	return UserDevicesController{
		Settings:         settings,
		DBS:              dbs,
		log:              logger,
		DeviceDefSvc:     ddSvc,
		taskSvc:          taskSvc,
		eventService:     eventService,
		smartcarClient:   smartcarClient,
		teslaService:     teslaSvc,
		teslaTaskService: teslaTaskService,
		encrypter:        encrypter,
	}
}

// GetUserDevices godoc
// @Description  gets all devices associated with current user - pulled from token
// @Tags           user-devices
// @Produce      json
// @Success      200  {object}  []controllers.UserDeviceFull
// @Security     BearerAuth
// @Router       /user/devices/me [get]
func (udc *UserDevicesController) GetUserDevices(c *fiber.Ctx) error {
	userID := getUserID(c)
	devices, err := models.UserDevices(qm.Where("user_id = ?", userID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
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
		dd, err := NewDeviceDefinitionFromDatabase(d.R.DeviceDefinition)
		if err != nil {
			return err
		}

		filteredIntegrations := []services.DeviceCompatibility{}
		if d.CountryCode.Valid {
			if countryRecord := services.FindCountry(d.CountryCode.String); countryRecord != nil {
				for _, integration := range dd.CompatibleIntegrations {
					if integration.Region == countryRecord.Region {
						integration.Country = d.CountryCode.String // Faking it until the UI updates for regions.
						filteredIntegrations = append(filteredIntegrations, integration)
					}
				}
			}
		}

		dd.CompatibleIntegrations = filteredIntegrations

		rp[i] = UserDeviceFull{
			ID:               d.ID,
			VIN:              d.VinIdentifier.Ptr(),
			VINConfirmed:     d.VinConfirmed,
			Name:             d.Name.Ptr(),
			CustomImageURL:   d.CustomImageURL.Ptr(),
			CountryCode:      d.CountryCode.Ptr(),
			DeviceDefinition: dd,
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
			CreatedAt:     udi.CreatedAt,
		}
	}

	return out
}

const UserDeviceCreationEventType = "com.dimo.zone.device.create"

type UserDeviceEvent struct {
	Timestamp time.Time                      `json:"timestamp"`
	UserID    string                         `json:"userId"`
	Device    services.UserDeviceEventDevice `json:"device"`
}

// RegisterDeviceForUser godoc
// @Description  adds a device to a user. can add with only device_definition_id or with MMY, which will create a device_definition on the fly
// @Tags           user-devices
// @Produce      json
// @Accept       json
// @Param        user_device  body  controllers.RegisterUserDevice  true  "add device to user. either MMY or id are required"
// @Security     ApiKeyAuth
// @Success      201  {object}  controllers.RegisterUserDeviceResponse
// @Security     BearerAuth
// @Router       /user/devices [post]
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
	// attach device def to user
	if reg.DeviceDefinitionID != nil {
		dd, err = models.DeviceDefinitions(qm.Load(models.DeviceDefinitionRels.DeviceMake),
			models.DeviceDefinitionWhere.ID.EQ(*reg.DeviceDefinitionID)).One(c.Context(), tx)
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
			dm, err := udc.DeviceDefSvc.GetOrCreateMake(c.Context(), tx, *reg.Make)
			if err != nil {
				return err
			}
			// since Definition does not exist, create one on the fly with userID as source and not verified
			dd = &models.DeviceDefinition{
				ID:           ksuid.New().String(),
				DeviceMakeID: dm.ID,
				Model:        *reg.Model,
				Year:         int16(*reg.Year),
				Source:       null.StringFrom("userID:" + userID),
				Verified:     false,
			}
			err = dd.Insert(c.Context(), tx, boil.Infer())
			if err != nil {
				return err
			}
			dd.R = dd.R.NewStruct()
			dd.R.DeviceMake = dm
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
		CountryCode:        null.StringFrom(reg.CountryCode),
	}
	err = ud.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return errorResponseHandler(c, errors.Wrapf(err, "could not create user device for def_id: %s", dd.ID), fiber.StatusInternalServerError)
	}
	region := ""
	if countryRecord := services.FindCountry(reg.CountryCode); countryRecord != nil {
		region = countryRecord.Region
	}
	// get device integrations to return in payload - helps frontend
	deviceInts, err := models.DeviceIntegrations(
		qm.Load(models.DeviceIntegrationRels.Integration),
		models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(dd.ID),
		models.DeviceIntegrationWhere.Region.EQ(region),
	).All(c.Context(), tx)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	err = tx.Commit()
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	if dd.R == nil {
		dd.R = dd.R.NewStruct()
	}
	dd.R.DeviceIntegrations = deviceInts

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
		Type:    UserDeviceCreationEventType,
		Subject: userID,
		Source:  "devices-api",
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  dd.R.DeviceMake.Name,
				Model: dd.Model,
				Year:  int(dd.Year), // Odd.
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed emitting device creation event")
	}

	ddNice, err := NewDeviceDefinitionFromDatabase(dd)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userDevice": UserDeviceFull{
			ID:               ud.ID,
			VIN:              ud.VinIdentifier.Ptr(),
			VINConfirmed:     ud.VinConfirmed,
			Name:             ud.Name.Ptr(),
			CustomImageURL:   ud.CustomImageURL.Ptr(),
			DeviceDefinition: ddNice,
			CountryCode:      ud.CountryCode.Ptr(),
			Integrations:     nil, // userDevice just created, there would never be any integrations setup
		},
	})
}

// UpdateVIN godoc
// @Description  updates the VIN on the user device record
// @Tags           user-devices
// @Produce      json
// @Accept       json
// @Param        vin           body  controllers.UpdateVINReq  true  "VIN"
// @Param        userDeviceID  path  string                    true  "user id"
// @Success      204
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/vin [patch]
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
// @Description  updates the Name on the user device record
// @Tags           user-devices
// @Produce      json
// @Accept       json
// @Param        name            body  controllers.UpdateNameReq  true  "Name"
// @Param        user_device_id  path  string                     true  "user id"
// @Success      204
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/name [patch]
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

// UpdateCountryCode godoc
// @Description  updates the CountryCode on the user device record
// @Tags           user-devices
// @Produce      json
// @Accept       json
// @Param        name  body  controllers.UpdateCountryCodeReq  true  "Country code"
// @Success      204
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/country_code [patch]
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
// @Description  delete the user device record (hard delete)
// @Tags                       user-devices
// @Param        userDeviceID  path  string  true  "user id"
// @Success      204
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID} [delete]
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
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations), // Probably don't need this one.
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationRels.Integration)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return errorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	for _, apiInteg := range userDevice.R.UserDeviceAPIIntegrations {
		if apiInteg.R.Integration.Vendor == services.SmartCarVendor {
			if apiInteg.ExternalID.Valid {
				err = udc.taskSvc.StartSmartcarDeregistrationTasks(udi, apiInteg.IntegrationID, apiInteg.ExternalID.String, apiInteg.AccessToken)
				if err != nil {
					return errorResponseHandler(c, err, fiber.StatusInternalServerError)
				}
			}
		} else if apiInteg.R.Integration.Vendor == "Tesla" {
			if apiInteg.ExternalID.Valid {
				if err := udc.teslaTaskService.StopPoll(apiInteg); err != nil {
					return errorResponseHandler(c, err, fiber.StatusInternalServerError)
				}
			}
		} else {
			udc.log.Warn().Msgf("Don't know how to deregister integration %s for device %s", apiInteg.IntegrationID, udi)
		}
		err = udc.eventService.Emit(&services.Event{
			Type:    "com.dimo.zone.device.integration.delete",
			Source:  "devices-api",
			Subject: udi,
			Data: services.UserDeviceIntegrationEvent{
				Timestamp: time.Now(),
				UserID:    userID,
				Device: services.UserDeviceEventDevice{
					ID:    udi,
					Make:  userDevice.R.DeviceDefinition.R.DeviceMake.Name,
					Model: userDevice.R.DeviceDefinition.Model,
					Year:  int(userDevice.R.DeviceDefinition.Year),
				},
				Integration: services.UserDeviceEventIntegration{
					ID:     apiInteg.R.Integration.ID,
					Type:   apiInteg.R.Integration.Type,
					Style:  apiInteg.R.Integration.Style,
					Vendor: apiInteg.R.Integration.Vendor,
				},
			},
		})
		if err != nil {
			udc.log.Err(err).Msg("Failed to emit integration deletion")
		}
	}

	// This will delete the associated integrations as well.
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
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    udi,
				Make:  dd.R.DeviceMake.Name,
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
	CountryCode        string  `json:"countryCode"`
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
		validation.Field(&reg.CountryCode, validation.Required, validation.Length(3, 3)),
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
	IntegrationID string    `json:"integrationId"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

// RegisterDeviceIntegrationRequest carries credentials used to connect the device to a given
// integration.
type RegisterDeviceIntegrationRequest struct {
	// Code is an OAuth authorization code. Not used in all integrations.
	Code string `json:"code"`
	// RedirectURI is the OAuth redirect URI used by the frontend. Not used in all integrations.
	RedirectURI string `json:"redirectURI"`

	ExternalID   string `json:"externalId"`
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int    `json:"expiresIn"`
	RefreshToken string `json:"refreshToken"`
}

type GetUserDeviceIntegrationResponse struct {
	// Status is one of "Pending", "PendingFirstData", "Active", "Failed", "DuplicateIntegration".
	Status string `json:"status"`
	// ExternalID is the identifier used by the third party for the device. It may be absent if we
	// haven't authorized yet.
	ExternalID null.String `json:"externalId" swaggertype:"string"`
	// CreatedAt is the creation time of this integration for this device.
	CreatedAt time.Time `json:"createdAt"`
}
