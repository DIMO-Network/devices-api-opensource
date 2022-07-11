package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/Shopify/sarama"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type UserDevicesController struct {
	Settings              *config.Settings
	DBS                   func() *database.DBReaderWriter
	DeviceDefSvc          services.IDeviceDefinitionService
	log                   *zerolog.Logger
	taskSvc               services.ITaskService
	eventService          services.EventService
	smartcarClient        services.SmartcarClient
	smartcarTaskSvc       services.SmartcarTaskService
	teslaService          services.TeslaService
	teslaTaskService      services.TeslaTaskService
	cipher                shared.Cipher
	autoPiSvc             services.AutoPiAPIService
	nhtsaService          services.INHTSAService
	autoPiIngestRegistrar services.IngestRegistrar
	autoPiTaskService     services.AutoPiTaskService
	s3                    *s3.Client
	producer              sarama.SyncProducer
}

// NewUserDevicesController constructor
func NewUserDevicesController(
	settings *config.Settings,
	dbs func() *database.DBReaderWriter,
	logger *zerolog.Logger,
	ddSvc services.IDeviceDefinitionService,
	taskSvc services.ITaskService,
	eventService services.EventService,
	smartcarClient services.SmartcarClient,
	smartcarTaskSvc services.SmartcarTaskService,
	teslaService services.TeslaService,
	teslaTaskService services.TeslaTaskService,
	cipher shared.Cipher,
	autoPiSvc services.AutoPiAPIService,
	nhtsaService services.INHTSAService,
	autoPiIngestRegistrar services.IngestRegistrar,
	autoPiTaskService services.AutoPiTaskService,
	producer sarama.SyncProducer,
	s3NFTClient *s3.Client,
) UserDevicesController {
	return UserDevicesController{
		Settings:              settings,
		DBS:                   dbs,
		log:                   logger,
		DeviceDefSvc:          ddSvc,
		taskSvc:               taskSvc,
		eventService:          eventService,
		smartcarClient:        smartcarClient,
		smartcarTaskSvc:       smartcarTaskSvc,
		teslaService:          teslaService,
		teslaTaskService:      teslaTaskService,
		cipher:                cipher,
		autoPiSvc:             autoPiSvc,
		nhtsaService:          nhtsaService,
		autoPiIngestRegistrar: autoPiIngestRegistrar,
		autoPiTaskService:     autoPiTaskService,
		s3:                    s3NFTClient,
		producer:              producer,
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
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationRels.Integration)),
		qm.Load(models.UserDeviceRels.MintRequest),
		qm.OrderBy("created_at"),
	).All(c.Context(), udc.DBS().Reader)
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

		md := new(services.UserDeviceMetadata)
		if d.Metadata.Valid {
			if err := d.Metadata.Unmarshal(md); err != nil {
				return opaqueInternalError
			}
		}

		var nft *NFTData
		if mr := d.R.MintRequest; mr != nil {
			nft = &NFTData{
				Status: mr.TXState,
			}
			if mr.TXHash.Valid {
				txHash := common.BytesToHash(mr.TXHash.Bytes).String()
				nft.TxHash = &txHash
			}
			if !mr.TokenID.IsZero() {
				nft.TokenID = mr.TokenID.Big.Int(new(big.Int))
				nft.TokenURI = fmt.Sprintf("%s/v1/nfts/%s", udc.Settings.DeploymentBaseURL, nft.TokenID)
			}
		}

		rp[i] = UserDeviceFull{
			ID:               d.ID,
			VIN:              d.VinIdentifier.Ptr(),
			VINConfirmed:     d.VinConfirmed,
			Name:             d.Name.Ptr(),
			CustomImageURL:   d.CustomImageURL.Ptr(),
			CountryCode:      d.CountryCode.Ptr(),
			DeviceDefinition: dd,
			Integrations:     NewUserDeviceIntegrationStatusesFromDatabase(d.R.UserDeviceAPIIntegrations),
			Metadata:         *md,
			NFT:              nft,
		}
	}

	return c.JSON(fiber.Map{
		"userDevices": rp,
	})
}

func NewUserDeviceIntegrationStatusesFromDatabase(udis []*models.UserDeviceAPIIntegration) []UserDeviceIntegrationStatus {
	out := make([]UserDeviceIntegrationStatus, len(udis))

	for i, udi := range udis {
		// TODO(elffjs): Remove this translation when the frontend is ready for "AuthenticationFailure".
		status := udi.Status
		if status == models.UserDeviceAPIIntegrationStatusAuthenticationFailure {
			status = models.UserDeviceAPIIntegrationStatusFailed
		}

		out[i] = UserDeviceIntegrationStatus{
			IntegrationID: udi.IntegrationID,
			Status:        status,
			ExternalID:    udi.ExternalID.Ptr(),
			CreatedAt:     udi.CreatedAt,
			UpdatedAt:     udi.UpdatedAt,
			Metadata:      udi.Metadata,
		}
		if udi.R != nil && udi.R.Integration != nil {
			out[i].IntegrationVendor = udi.R.Integration.Vendor
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

	// Baby the frontend.
	for i := range ddNice.CompatibleIntegrations {
		ddNice.CompatibleIntegrations[i].Country = reg.CountryCode
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

var opaqueInternalError = fiber.NewError(fiber.StatusBadGateway, "Internal error.")

// UpdateVIN godoc
// @Description  updates the VIN on the user device record
// @Tags         user-devices
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

	logger := udc.log.With().Str("route", c.Route().Path).Str("userId", userID).Str("userDeviceId", udi).Logger()

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		models.UserDeviceWhere.ID.EQ(udi),
	).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
		logger.Err(err).Msg("Database error searching for device.")
		return err
	}

	if userDevice.VinConfirmed {
		return fiber.NewError(fiber.StatusBadRequest, "Can't update a VIN that was previously confirmed.")
	}

	vinReq := &UpdateVINReq{}
	if err := c.BodyParser(vinReq); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Could not parse request body.")
	}
	if err := vinReq.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid VIN.")
	}

	userDevice.VinIdentifier = null.StringFrom(strings.ToUpper(*vinReq.VIN))
	if _, err := userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer()); err != nil {
		// Okay to dereference here, since we validated the field.
		logger.Err(err).Msgf("Database error updating VIN to %s.", *vinReq.VIN)
		return opaqueInternalError
	}

	// TODO: Genericize this for more countries.
	if userDevice.CountryCode.Valid && userDevice.CountryCode.String == "USA" {
		if err := udc.updateUSAPowertrain(c.Context(), userDevice); err != nil {
			logger.Err(err).Msg("Failed to update American powertrain type.")
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) updateUSAPowertrain(ctx context.Context, userDevice *models.UserDevice) error {
	resp, err := udc.nhtsaService.DecodeVIN(userDevice.VinIdentifier.String)
	if err != nil {
		return err
	}

	dt, err := resp.DriveType()
	if err != nil {
		return err
	}

	md := new(services.UserDeviceMetadata)
	if err := userDevice.Metadata.Unmarshal(md); err != nil {
		return err
	}

	md.PowertrainType = &dt
	if err := userDevice.Metadata.Marshal(md); err != nil {
		return err
	}
	if _, err := userDevice.Update(ctx, udc.DBS().Writer, boil.Infer()); err != nil {
		return err
	}

	return nil
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

	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return err
	}
	name := &UpdateNameReq{}
	if err := c.BodyParser(name); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	if err := name.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Name field is limited to 16 alphanumeric characters.")
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
	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return err
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

// UpdateImage godoc
// @Description  updates the ImageUrl on the user device record
// @Tags         user-devices
// @Produce      json
// @Accept       json
// @Param        name  body  controllers.UpdateImageURLReq  true  "Image URL"
// @Success      204
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/image [patch]
func (udc *UserDevicesController) UpdateImage(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)

	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return err
	}
	req := &UpdateImageURLReq{}
	if err := c.BodyParser(req); err != nil {
		// Return status 400 and error message.
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.CustomImageURL = null.StringFromPtr(req.ImageURL)
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
				if apiInteg.TaskID.Valid {
					err = udc.smartcarTaskSvc.StopPoll(apiInteg)
					if err != nil {
						return errorResponseHandler(c, err, fiber.StatusInternalServerError)
					}
				} else {
					err = udc.taskSvc.StartSmartcarDeregistrationTasks(udi, apiInteg.IntegrationID, apiInteg.ExternalID.String, apiInteg.AccessToken.String)
					if err != nil {
						return errorResponseHandler(c, err, fiber.StatusInternalServerError)
					}
				}
			}
		} else if apiInteg.R.Integration.Vendor == "Tesla" {
			if apiInteg.ExternalID.Valid {
				if err := udc.teslaTaskService.StopPoll(apiInteg); err != nil {
					return errorResponseHandler(c, err, fiber.StatusInternalServerError)
				}
			}
		} else if apiInteg.R.Integration.Vendor == services.AutoPiVendor {
			err = udc.autoPiIngestRegistrar.Deregister(apiInteg.ExternalID.String, apiInteg.UserDeviceID, apiInteg.IntegrationID)
			if err != nil {
				udc.log.Err(err).Msgf("unexpected error deregistering autopi device from ingest. userDeviceID: %s", apiInteg.UserDeviceID)
				return err
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

// GetMintDataToSign godoc
// @Description  Returns the data the user must sign in order to mint this device.
// @Tags         user-devices
// @Param        userDeviceID path string true "user device ID"
// @Success      200 {object} signer.TypedData
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/commands/mint [get]
func (udc *UserDevicesController) GetMintDataToSign(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := getUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	mk := userDevice.R.DeviceDefinition.R.DeviceMake
	if mk.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Device make %s not yet minted.", mk.Name))
	}
	mkTok := mk.TokenID.Int(nil)

	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		udc.log.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusBadRequest, "user does not have an ethereum address on file")
	}

	// Can't use signer.TypedData because the serialization of math.HexOrDecimal256
	// makes Trust Wallet go nuts.
	typedData := map[string]any{
		"types": signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintVehicleSign": {
				{Name: "rootNode", Type: "uint256"},
				{Name: "_owner", Type: "address"},
				{Name: "attributes", Type: "string[]"},
				{Name: "infos", Type: "string[]"},
			},
		},
		"primaryType": "MintVehicleSign",
		"domain": signer.TypedDataMessage{
			"name":              udc.Settings.NFTContractName,
			"version":           udc.Settings.NFTContractVersion,
			"chainId":           udc.Settings.NFTChainID,
			"verifyingContract": udc.Settings.NFTContractAddr,
		},
		"message": signer.TypedDataMessage{
			"rootNode":   mkTok,
			"_owner":     *user.EthereumAddress,
			"attributes": []string{"Make", "Model", "Year"},
			"infos": []string{
				userDevice.R.DeviceDefinition.R.DeviceMake.Name,
				userDevice.R.DeviceDefinition.Model,
				strconv.Itoa(int(userDevice.R.DeviceDefinition.Year)),
			},
		},
	}

	return c.JSON(typedData)
}

// TODO(elffjs): Do not keep these functions in this file!
func computeTypedDataHash(td *signer.TypedData) (hash common.Hash, err error) {
	domainSep, err := td.HashStruct("EIP712Domain", td.Domain.Map())
	if err != nil {
		return
	}
	msgHash, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		return
	}

	payload := []byte{0x19, 0x01}
	payload = append(payload, domainSep...)
	payload = append(payload, msgHash...)

	hash = crypto.Keccak256Hash(payload)
	return
}

func recoverAddress(td *signer.TypedData, signature []byte) (addr common.Address, err error) {
	hash, err := computeTypedDataHash(td)
	if err != nil {
		return
	}
	signature[64] -= 27
	rawPub, err := crypto.Ecrecover(hash[:], signature)
	if err != nil {
		return
	}

	pub, err := crypto.UnmarshalPubkey(rawPub)
	if err != nil {
		return
	}
	addr = crypto.PubkeyToAddress(*pub)
	return
}

// MintDevice godoc
// @Description  Sends a mint device request to the blockchain
// @Tags         user-devices
// @Param        userDeviceID path string true "user device ID"
// @Param        mintRequest body controllers.MintRequest true "Signature and NFT data"
// @Success      200
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/commands/mint [post]
func (udc *UserDevicesController) MintDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := getUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationRels.Integration)),
		qm.Load(models.UserDeviceRels.MintRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	mk := userDevice.R.DeviceDefinition.R.DeviceMake
	if mk.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Device make %s not yet minted.", mk.Name))
	}
	mkBI := mk.TokenID.Int(nil)
	mkTok := (*math.HexOrDecimal256)(mkBI)

	mintRequestID := ksuid.New().String()
	mreq := &models.MintRequest{
		ID:           mintRequestID,
		UserDeviceID: null.StringFrom(userDeviceID),
		TXState:      models.TxstateUnstarted,
	}

	eligible := false
	// Check ability based on completed integrations.
	for _, apiInt := range userDevice.R.UserDeviceAPIIntegrations {
		// Might be able to do this check in the DB.
		if apiInt.Status != models.UserDeviceAPIIntegrationStatusActive {
			continue
		}
		switch apiInt.R.Integration.Vendor {
		case services.SmartCarVendor, services.TeslaVendor:
			eligible = true
			// Sure hope this works!
			mreq.Vin = userDevice.VinIdentifier
		case services.AutoPiVendor:
			eligible = true
			mreq.ChildDeviceID = apiInt.ExternalID
		}
	}

	if !eligible {
		return fiber.NewError(fiber.StatusBadRequest, "Device does not have an active, eligible integration.")
	}

	// Check historical mints in prod.
	if udc.Settings.Environment == "prod" {
		var rateControlConds []qm.QueryMod
		if mreq.Vin.Valid {
			rateControlConds = []qm.QueryMod{models.MintRequestWhere.Vin.EQ(mreq.Vin)}
		}
		if mreq.ChildDeviceID.Valid {
			if len(rateControlConds) == 0 {
				rateControlConds = []qm.QueryMod{models.MintRequestWhere.ChildDeviceID.EQ(mreq.ChildDeviceID)}
			} else {
				rateControlConds = append(rateControlConds, qm.Or2(models.MintRequestWhere.ChildDeviceID.EQ(mreq.ChildDeviceID)))
			}
		}

		conflict, err := models.MintRequests(rateControlConds...).Exists(c.Context(), udc.DBS().Reader)
		if err != nil {
			udc.log.Err(err).Msg("Couldn't search for old, conflicting records.")
			return opaqueInternalError
		}

		if conflict {
			return fiber.NewError(fiber.StatusConflict, "Already minted.")
		}
	}

	mr := new(MintRequest)
	if err := c.BodyParser(mr); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	// This may not be there, but if it is we should delete it.
	imageData := strings.TrimPrefix(mr.ImageData, "data:image/png;base64,")

	image, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Field imageData not properly base64-encoded.")
	}

	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("mintRequestId", mintRequestID).
		Str("handler", "MintDevice").
		Logger()

	logger.Info().Msg("Mint request received.")

	_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
		Bucket: &udc.Settings.NFTS3Bucket,
		Key:    aws.String(mintRequestID + ".png"), // This will be the request ID.
		Body:   bytes.NewReader(image),
	})
	if err != nil {
		logger.Err(err).Msg("Failed to save image to S3.")
		return opaqueInternalError
	}

	conn, err := grpc.Dial(udc.Settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Err(err).Msg("Failed to create users API client.")
		return opaqueInternalError
	}
	defer conn.Close()

	usersClient := pb.NewUserServiceClient(conn)

	user, err := usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		logger.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusBadRequest, "user does not have an ethereum address on file")
	}

	typedData := &signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"MintVehicleSign": {
				{Name: "rootNode", Type: "uint256"},
				{Name: "_owner", Type: "address"},
				{Name: "attributes", Type: "string[]"},
				{Name: "infos", Type: "string[]"},
			},
		},
		PrimaryType: "MintVehicleSign",
		Domain: signer.TypedDataDomain{
			Name:              udc.Settings.NFTContractName,
			Version:           udc.Settings.NFTContractVersion,
			ChainId:           math.NewHexOrDecimal256(int64(udc.Settings.NFTChainID)),
			VerifyingContract: udc.Settings.NFTContractAddr,
		},
		Message: signer.TypedDataMessage{
			"rootNode":   mkTok,
			"_owner":     *user.EthereumAddress,
			"attributes": []any{"Make", "Model", "Year"},
			"infos": []any{
				userDevice.R.DeviceDefinition.R.DeviceMake.Name,
				userDevice.R.DeviceDefinition.Model,
				strconv.Itoa(int(userDevice.R.DeviceDefinition.Year)),
			},
		},
	}

	sigBytes := common.FromHex(mr.Signature)
	if len(sigBytes) != 65 {
		return fiber.NewError(fiber.StatusBadRequest, "Signature has incorrect length, should be 65.")
	}

	logger.Info().Bytes("bytes", sigBytes).Msg("Hex Signature")

	recAddr, err := recoverAddress(typedData, sigBytes)
	if err != nil {
		logger.Err(err).Msg("Failed recovering address.")
		return fiber.NewError(fiber.StatusBadRequest, "Signature incorrect.")
	}

	realAddr := common.HexToAddress(*user.EthereumAddress)

	if recAddr != realAddr {
		logger.Err(err).Str("recAddr", recAddr.String()).Msg("Recovered address, but incorrect.")
		return fiber.NewError(fiber.StatusBadRequest, "Signature incorrect.")
	}

	me := shared.CloudEvent[MintEventData]{
		ID:          ksuid.New().String(),
		Source:      "devices-api",
		SpecVersion: "1.0",
		Subject:     userDeviceID,
		Time:        time.Now(),
		Type:        "zone.dimo.device.mint.request",
		Data: MintEventData{
			RequestID:    mintRequestID,
			UserDeviceID: userDeviceID,
			Owner:        *user.EthereumAddress,
			RootNode:     mkBI,
			Attributes:   []string{"Make", "Model", "Year"},
			Infos: []string{
				userDevice.R.DeviceDefinition.R.DeviceMake.Name,
				userDevice.R.DeviceDefinition.Model,
				strconv.Itoa(int(userDevice.R.DeviceDefinition.Year)),
			},
			Signature: mr.Signature,
		},
	}

	b, err := json.Marshal(me)
	if err != nil {
		logger.Err(err).Msg("Failed to serialize mint request.")
		return opaqueInternalError
	}

	_, _, err = udc.producer.SendMessage(&sarama.ProducerMessage{
		Topic: udc.Settings.NFTInputTopic,
		Key:   sarama.StringEncoder(userDeviceID),
		Value: sarama.ByteEncoder(b),
	})
	if err != nil {
		logger.Err(err).Msgf("Couldn't produce mint request to Kafka.")
		return opaqueInternalError
	}

	if err := mreq.Insert(c.Context(), udc.DBS().Writer, boil.Infer()); err != nil {
		logger.Err(err).Msg("Failed to insert mint record.")
		return opaqueInternalError
	}

	return c.JSON(map[string]any{"mintRequestId": mintRequestID})
}

type MintEventData struct {
	RequestID    string   `json:"requestId"`
	UserDeviceID string   `json:"userDeviceId"`
	Owner        string   `json:"owner"`
	RootNode     *big.Int `json:"rootNode"`
	Attributes   []string `json:"attributes"`
	Infos        []string `json:"infos"`
	// Signature is the EIP-712 signature of the RootNode, Attributes, and Infos fields.
	Signature string `json:"signature"`
}

// MintRequest contains the user's signature for the mint request as well as the
// NFT image.
type MintRequest struct {
	// Signature is the hex encoding of the EIP-712 signature result.
	Signature string `json:"signature"`
	// ImageData contains the base64-encoded NFT PNG image.
	ImageData string `json:"imageData"`
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

type UpdateImageURLReq struct {
	ImageURL *string `json:"imageUrl"`
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

var vinRegex = regexp.MustCompile("^(?:[1-5]|7[F-Z0-9])")

func (u *UpdateVINReq) validate() error {

	validateLengthAndChars := validation.ValidateStruct(u,
		// vin must be 17 characters in length, alphanumeric, without characters I, O, Q
		validation.Field(&u.VIN, validation.Required, validation.Match(regexp.MustCompile("^[A-HJ-NPR-Z0-9]{17}$"))),
		// in addition to three excluded characters above, 10th character must not eual U, Z or 0
		validation.Field(&u.VIN, validation.Required, validation.Match(regexp.MustCompile("^.{9}[A-HJ-NPR-TV-Y1-9]"))),
	)
	if validateLengthAndChars != nil {
		return validateLengthAndChars
	}

	// if car is made in North America, apply additional checksum validation (character 9)
	// world manufacturer identifier is first 2 digits of vin
	wmi := (*u.VIN)[:2]
	checkSum := (*u.VIN)[8:9]
	northAmerDevice := vinRegex.MatchString(wmi)

	if northAmerDevice {
		var derivedCheck string
		check := transcodeDigits(*u.VIN)
		checkNum := check % 11

		if checkNum == 10 {
			derivedCheck = "X"
		} else {
			derivedCheck = strconv.Itoa(int(checkNum))
		}

		return validation.Validate(checkSum, validation.In(derivedCheck))

	}

	return nil
}

func (u *UpdateNameReq) validate() error {

	return validation.ValidateStruct(u,
		// name must be between 1 and 16 alphanumeric characters in length (spaces are not allowed)
		// NOTE: this captures characters in the latin/ chinese/ cyrillic alphabet but doesn't work as well for thai or arabic
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`^[\s\p{L}\p{N}\p{M}#'":_-]{1,25}$`))),
		// cannot start with space
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`^[^\s]`))),
		// cannot end with space
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`.+[^\s]$|[^\s]$`))),
	)
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
	Metadata         services.UserDeviceMetadata   `json:"metadata"`
	NFT              *NFTData                      `json:"nft,omitempty"`
}

type NFTData struct {
	TokenID  *big.Int `json:"tokenId,omitempty" swaggertype:"number" example:"37"`
	TokenURI string   `json:"tokenUri,omitempty" example:"https://nft.dimo.zone/37"`
	// TxHash is the hash of the minting transaction.
	TxHash *string `json:"txHash,omitempty" example:"0x30bce3da6985897224b29a0fe064fd2b426bb85a394cc09efe823b5c83326a8e"`
	// Status is the minting status of the NFT.
	Status string `json:"status" enums:"Unstarted,Submitted,Mined,Confirmed" example:"Confirmed"`
}

func transcodeDigits(vin string) int {
	var digitSum = 0
	var code int
	for i, chr := range vin {
		switch chr {
		case 'A', 'J', '1':
			code = 1
		case 'B', 'K', 'S', '2':
			code = 2
		case 'C', 'L', 'T', '3':
			code = 3
		case 'D', 'M', 'U', '4':
			code = 4
		case 'E', 'N', 'V', '5':
			code = 5
		case 'F', 'W', '6':
			code = 6
		case 'G', 'P', 'X', '7':
			code = 7
		case 'H', 'Y', '8':
			code = 8
		case 'R', 'Z', '9':
			code = 9
		default:
			code = 0
		}
		switch i + 1 {
		case 1, 11:
			digitSum += code * 8
		case 2, 12:
			digitSum += code * 7
		case 3, 13:
			digitSum += code * 6
		case 4, 14:
			digitSum += code * 5
		case 5, 15:
			digitSum += code * 4
		case 6, 16:
			digitSum += code * 3
		case 7, 17:
			digitSum += code * 2
		case 8:
			digitSum += code * 10
		case 9:
			digitSum += code * 0
		case 10:
			digitSum += code * 9
		}
	}
	return digitSum
}
