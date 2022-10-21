package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/api"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
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
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type UserDevicesController struct {
	Settings                  *config.Settings
	DBS                       func() *database.DBReaderWriter
	DeviceDefSvc              services.DeviceDefinitionService
	DeviceDefIntSvc           services.DeviceDefinitionIntegrationService
	log                       *zerolog.Logger
	eventService              services.EventService
	smartcarClient            services.SmartcarClient
	smartcarTaskSvc           services.SmartcarTaskService
	teslaService              services.TeslaService
	teslaTaskService          services.TeslaTaskService
	cipher                    shared.Cipher
	autoPiSvc                 services.AutoPiAPIService
	nhtsaService              services.INHTSAService
	autoPiIngestRegistrar     services.IngestRegistrar
	autoPiTaskService         services.AutoPiTaskService
	drivlyTaskService         services.DrivlyTaskService
	blackbookTaskService      services.BlackbookTaskService
	s3                        *s3.Client
	producer                  sarama.SyncProducer
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar
}

// NewUserDevicesController constructor
func NewUserDevicesController(
	settings *config.Settings,
	dbs func() *database.DBReaderWriter,
	logger *zerolog.Logger,
	ddSvc services.DeviceDefinitionService,
	ddIntSvc services.DeviceDefinitionIntegrationService,
	eventService services.EventService,
	smartcarClient services.SmartcarClient,
	smartcarTaskSvc services.SmartcarTaskService,
	teslaService services.TeslaService,
	teslaTaskService services.TeslaTaskService,
	cipher shared.Cipher,
	autoPiSvc services.AutoPiAPIService,
	nhtsaService services.INHTSAService,
	autoPiIngestRegistrar services.IngestRegistrar,
	deviceDefinitionRegistrar services.DeviceDefinitionRegistrar,
	autoPiTaskService services.AutoPiTaskService,
	producer sarama.SyncProducer,
	s3NFTClient *s3.Client,
	drivlyTaskService services.DrivlyTaskService,
	blackbookTaskService services.BlackbookTaskService,
) UserDevicesController {
	return UserDevicesController{
		Settings:                  settings,
		DBS:                       dbs,
		log:                       logger,
		DeviceDefSvc:              ddSvc,
		DeviceDefIntSvc:           ddIntSvc,
		eventService:              eventService,
		smartcarClient:            smartcarClient,
		smartcarTaskSvc:           smartcarTaskSvc,
		teslaService:              teslaService,
		teslaTaskService:          teslaTaskService,
		cipher:                    cipher,
		autoPiSvc:                 autoPiSvc,
		nhtsaService:              nhtsaService,
		autoPiIngestRegistrar:     autoPiIngestRegistrar,
		autoPiTaskService:         autoPiTaskService,
		s3:                        s3NFTClient,
		producer:                  producer,
		drivlyTaskService:         drivlyTaskService,
		blackbookTaskService:      blackbookTaskService,
		deviceDefinitionRegistrar: deviceDefinitionRegistrar,
	}
}

// GetUserDevices godoc
// @Description gets all devices associated with current user - pulled from token
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} []controllers.UserDeviceFull
// @Security    BearerAuth
// @Router      /user/devices/me [get]
func (udc *UserDevicesController) GetUserDevices(c *fiber.Ctx) error {
	// todo grpc call out to grpc service endpoint in the deviceDefinitionsService udc.deviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{ "todo"} )

	userID := api.GetUserID(c)
	devices, err := models.UserDevices(qm.Where("user_id = ?", userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations)),
		qm.Load(models.UserDeviceRels.MintRequest),
		qm.Load(models.UserDeviceRels.MintMetaTransactionRequest),
		qm.OrderBy("created_at"),
	).All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}
	rp := make([]UserDeviceFull, len(devices))
	ids := []string{}

	for _, d := range devices {
		ids = append(ids, d.DeviceDefinitionID)
	}

	if len(ids) == 0 {
		return c.JSON(fiber.Map{
			"userDevices": rp,
		})
	}

	deviceDefinitionResponse, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), ids)

	if err != nil {
		return api.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+ids[0])
	}

	filterDeviceDefinition := func(id string, items []*ddgrpc.GetDeviceDefinitionItemResponse) (*ddgrpc.GetDeviceDefinitionItemResponse, error) {
		for _, dd := range items {
			if id == dd.DeviceDefinitionId {
				return dd, nil
			}
		}
		return nil, errors.New("no device definition")
	}

	integrations, err2 := udc.DeviceDefSvc.GetIntegrations(c.Context())
	if err2 != nil {
		return api.GrpcErrorToFiber(err2, "failed to get integrations")
	}

	for i, d := range devices {

		deviceDefinition, err := filterDeviceDefinition(d.DeviceDefinitionID, deviceDefinitionResponse)

		if err != nil {
			return err
		}

		dd, err := NewDeviceDefinitionFromGRPC(deviceDefinition)
		if err != nil {
			return err
		}

		filteredIntegrations := []services.DeviceCompatibility{}
		if d.CountryCode.Valid {
			if countryRecord := constants.FindCountry(d.CountryCode.String); countryRecord != nil {
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
		if udc.Settings.Environment != "prod" {
			if mtr := d.R.MintMetaTransactionRequest; mtr != nil {
				nft = &NFTData{
					Status: mtr.Status,
				}
				if mtr.Hash.Valid {
					hash := hexutil.Encode(mtr.Hash.Bytes)
					nft.TxHash = &hash
				}
				if !d.TokenID.IsZero() {
					nft.TokenID = d.TokenID.Int(nil)
					nft.TokenURI = fmt.Sprintf("%s/v1/nfts/%s", udc.Settings.DeploymentBaseURL, nft.TokenID)
				}
			}
		} else {
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
		}

		rp[i] = UserDeviceFull{
			ID:               d.ID,
			VIN:              d.VinIdentifier.Ptr(),
			VINConfirmed:     d.VinConfirmed,
			Name:             d.Name.Ptr(),
			CustomImageURL:   d.CustomImageURL.Ptr(),
			CountryCode:      d.CountryCode.Ptr(),
			DeviceDefinition: dd,
			Integrations:     NewUserDeviceIntegrationStatusesFromDatabase(d.R.UserDeviceAPIIntegrations, integrations),
			Metadata:         *md,
			NFT:              nft,
			OptedInAt:        d.OptedInAt.Ptr(),
		}
	}

	return c.JSON(fiber.Map{
		"userDevices": rp,
	})
}

func NewUserDeviceIntegrationStatusesFromDatabase(udis []*models.UserDeviceAPIIntegration, integrations []*ddgrpc.Integration) []UserDeviceIntegrationStatus {
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

		for _, integration := range integrations {
			if integration.Id == udi.IntegrationID {
				out[i].IntegrationVendor = integration.Vendor
				break
			}
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
// @Description adds a device to a user. can add with only device_definition_id or with MMY, which will create a device_definition on the fly
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       user_device body controllers.RegisterUserDevice true "add device to user. either MMY or id are required"
// @Security    ApiKeyAuth
// @Success     201 {object} controllers.RegisterUserDeviceResponse
// @Security    BearerAuth
// @Router      /user/devices [post]
func (udc *UserDevicesController) RegisterDeviceForUser(c *fiber.Ctx) error {
	userID := api.GetUserID(c)
	reg := &RegisterUserDevice{}
	if err := c.BodyParser(reg); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := reg.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	tx, err := udc.DBS().Writer.DB.BeginTx(c.Context(), nil)
	defer tx.Rollback() //nolint
	if err != nil {
		return err
	}
	// attach device def to user

	deviceDefinitionResponse, err2 := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{*reg.DeviceDefinitionID})

	if err2 != nil {
		return api.GrpcErrorToFiber(err2, fmt.Sprintf("error querying for device definition id: %s ", *reg.DeviceDefinitionID))
	}

	dd := deviceDefinitionResponse[0]

	userDeviceID := ksuid.New().String()
	// register device for the user
	ud := models.UserDevice{
		ID:                 userDeviceID,
		UserID:             userID,
		DeviceDefinitionID: dd.DeviceDefinitionId,
		CountryCode:        null.StringFrom(reg.CountryCode),
	}
	err = ud.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create user device for def_id: "+dd.DeviceDefinitionId)
	}
	//region := ""
	//if countryRecord := services.FindCountry(reg.CountryCode); countryRecord != nil {
	//	region = countryRecord.Region
	//}
	err = tx.Commit() // commmit the transaction
	if err != nil {
		return errors.Wrapf(err, "error commiting transaction to create geofence")
	}

	// don't block, as image fetch could take a while
	go func() {
		// todo grpc update this service to call device-defintions over grpc to update the image
		err := udc.DeviceDefSvc.CheckAndSetImage(c.Context(), dd, false)
		if err != nil {
			udc.log.Error().Err(err).Msg("error getting device image upon user_device registration")
			return
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
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year), // Odd.
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed emitting device creation event")
	}

	ddNice, err := NewDeviceDefinitionFromGRPC(dd)
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

var opaqueInternalError = fiber.NewError(fiber.StatusInternalServerError, "Internal error.")

// DeviceOptIn godoc
// @Description Opts the device into data-sharing, and hence rewards.
// @Tags        user-devices
// @Produce     json
// @Param       userDeviceID path string                   true "user device id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/opt-in [post]
func (udc *UserDevicesController) DeviceOptIn(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	logger := udc.log.With().Str("routeName", c.Route().Name).Str("userId", userID).Str("userDeviceId", udi).Logger()

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

	if userDevice.OptedInAt.Valid {
		logger.Info().Time("previousTime", userDevice.OptedInAt.Time).Msg("Already opted in to data-sharing.")
		return c.SendStatus(fiber.StatusNoContent)
	}

	userDevice.OptedInAt = null.TimeFrom(time.Now())

	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Whitelist(models.UserDeviceColumns.OptedInAt))
	if err != nil {
		return err
	}

	logger.Info().Msg("Opted into data-sharing.")

	return nil
}

// UpdateVIN godoc
// @Description updates the VIN on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       vin          body controllers.UpdateVINReq true "VIN"
// @Param       userDeviceID path string                   true "user id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/vin [patch]
func (udc *UserDevicesController) UpdateVIN(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

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
	upperVIN := strings.ToUpper(*vinReq.VIN)
	vinReq.VIN = &upperVIN
	if err := vinReq.validate(); err != nil {
		if vinReq.VIN != nil {
			logger.Err(err).Str("vin", *vinReq.VIN).Msg("VIN failed validation.")
		}
		return fiber.NewError(fiber.StatusBadRequest, "Invalid VIN.")
	}

	userDevice.VinIdentifier = null.StringFrom(upperVIN)
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
	// todo grpc pull vin decoder via grpc from device definitions
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
// @Description updates the Name on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       name           body controllers.UpdateNameReq true "Name"
// @Param       user_device_id path string                    true "user id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/name [patch]
func (udc *UserDevicesController) UpdateName(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}
	name := &UpdateNameReq{}
	if err := c.BodyParser(name); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if name.Name == nil {
		return fiber.NewError(fiber.StatusBadRequest, "name cannot be empty")
	}
	*name.Name = strings.TrimSpace(*name.Name)

	if err := name.validate(); err != nil {
		if name.Name != nil {
			udc.log.Warn().Err(err).Str("userDeviceId", udi).Str("userId", userID).Str("name", *name.Name).Msg("Proposed device name is invalid.")
		}
		return fiber.NewError(fiber.StatusBadRequest, "Name field is limited to 16 alphanumeric characters.")
	}

	userDevice.Name = null.StringFromPtr(name.Name)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateCountryCode godoc
// @Description updates the CountryCode on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       name body controllers.UpdateCountryCodeReq true "Country code"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/country_code [patch]
func (udc *UserDevicesController) UpdateCountryCode(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)
	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.ErrorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return err
	}
	countryCode := &UpdateCountryCodeReq{}
	if err := c.BodyParser(countryCode); err != nil {
		// Return status 400 and error message.
		return api.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.CountryCode = null.StringFromPtr(countryCode.CountryCode)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateImage godoc
// @Description updates the ImageUrl on the user device record
// @Tags        user-devices
// @Produce     json
// @Accept      json
// @Param       name body controllers.UpdateImageURLReq true "Image URL"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/image [patch]
func (udc *UserDevicesController) UpdateImage(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	userDevice, err := models.UserDevices(models.UserDeviceWhere.ID.EQ(udi), models.UserDeviceWhere.UserID.EQ(userID)).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.ErrorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return err
	}
	req := &UpdateImageURLReq{}
	if err := c.BodyParser(req); err != nil {
		// Return status 400 and error message.
		return api.ErrorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	userDevice.CustomImageURL = null.StringFromPtr(req.ImageURL)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type DeviceValuation struct {
	// Contains a list of valuation sets, one for each vendor
	ValuationSets []ValuationSet `json:"valuationSets"`
}
type ValuationSet struct {
	// The source of the valuation (eg. "drivly" or "blackbook")
	Vendor string `json:"vendor"`
	// The time the valuation was pulled or in the case of blackbook, this may be the event time of the device odometer which was used for the valuation
	Updated string `json:"updated,omitempty"`
	// The mileage used for the valuation
	Mileage int `json:"mileage,omitempty"`
	// This will be the zip code used (if any) for the valuation request regardless if the vendor uses it
	ZipCode string `json:"zipCode,omitempty"`
	// Useful when Drivly returns multiple vendors and we've selected one (eg. "drivly:blackbook")
	TradeInSource string `json:"tradeInSource,omitempty"`
	// tradeIn is equal to tradeInAverage when available
	TradeIn int `json:"tradeIn,omitempty"`
	// tradeInClean, tradeInAverage, and tradeInRough my not always be available
	TradeInClean   int `json:"tradeInClean,omitempty"`
	TradeInAverage int `json:"tradeInAverage,omitempty"`
	TradeInRough   int `json:"tradeInRough,omitempty"`
	// Useful when Drivly returns multiple vendors and we've selected one (eg. "drivly:blackbook")
	RetailSource string `json:"retailSource,omitempty"`
	// retail is equal to retailAverage when available
	Retail int `json:"retail,omitempty"`
	// retailClean, retailAverage, and retailRough my not always be available
	RetailClean   int `json:"retailClean,omitempty"`
	RetailAverage int `json:"retailAverage,omitempty"`
	RetailRough   int `json:"retailRough,omitempty"`
}

// GetValuations godoc
// @Description gets valuations for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceValuation
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/valuations [get]
func (udc *UserDevicesController) GetValuations(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	// Ensure user is owner of user device
	userDeviceExists, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if !userDeviceExists {
		return c.SendStatus(fiber.StatusForbidden)
	}

	logger := udc.log.With().Str("route", c.Route().Path).Str("userId", userID).Str("userDeviceId", udi).Logger()

	dVal := DeviceValuation{
		ValuationSets: []ValuationSet{},
	}

	// Drivly data
	drivlyVinData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(udi)),
		models.ExternalVinDatumWhere.PricingMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if drivlyVinData != nil {
		drivlyVal := ValuationSet{
			Vendor:  "drivly",
			Updated: drivlyVinData.UpdatedAt.Format(time.RFC3339),
		}
		drivlyJSON := drivlyVinData.PricingMetadata.JSON
		requestJSON := drivlyVinData.RequestMetadata.JSON
		requestMileage := gjson.GetBytes(requestJSON, "mileage")
		if requestMileage.Exists() {
			drivlyVal.Mileage = int(requestMileage.Int())
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			drivlyVal.ZipCode = requestZipCode.String()
		}
		// Drivly Trade-In
		switch {
		case gjson.GetBytes(drivlyJSON, "trade.blackBook.totalAvg").Exists():
			drivlyVal.TradeInSource = "drivly:blackbook"
			values := gjson.GetManyBytes(drivlyJSON, "trade.blackBook.totalRough", "trade.blackBook.totalAvg", "trade.blackBook.totalClean")
			drivlyVal.TradeInRough = int(values[0].Int())
			drivlyVal.TradeInAverage = int(values[1].Int())
			drivlyVal.TradeInClean = int(values[2].Int())
			drivlyVal.TradeIn = drivlyVal.TradeInAverage
		case gjson.GetBytes(drivlyJSON, "trade.kelley.book").Exists():
			drivlyVal.TradeInSource = "drivly:kelley"
			drivlyVal.TradeIn = int(gjson.GetBytes(drivlyJSON, "trade.kelley.book").Int())
		case gjson.GetBytes(drivlyJSON, "trade.edmunds.average").Exists():
			drivlyVal.TradeInSource = "drivly:edmunds"
			values := gjson.GetManyBytes(drivlyJSON, "trade.edmunds.rough", "trade.edmunds.average", "trade.edmunds.clean")
			drivlyVal.TradeInRough = int(values[0].Int())
			drivlyVal.TradeInAverage = int(values[1].Int())
			drivlyVal.TradeInClean = int(values[2].Int())
			drivlyVal.TradeIn = drivlyVal.TradeInAverage
		case gjson.GetBytes(drivlyJSON, "trade").Exists() && !gjson.GetBytes(drivlyJSON, "trade").IsObject():
			drivlyVal.TradeInSource = "drivly"
			drivlyVal.TradeIn = int(gjson.GetBytes(drivlyJSON, "trade").Int())
		default:
			logger.Error().Msg("Unexpected structure for driv.ly pricing data trade values")
		}
		// Drivly Retail
		switch {
		case gjson.GetBytes(drivlyJSON, "retail.blackBook.totalAvg").Exists():
			drivlyVal.RetailSource = "drivly:blackbook"
			values := gjson.GetManyBytes(drivlyJSON, "retail.blackBook.totalRough", "retail.blackBook.totalAvg", "retail.blackBook.totalClean")
			drivlyVal.RetailRough = int(values[0].Int())
			drivlyVal.RetailAverage = int(values[1].Int())
			drivlyVal.RetailClean = int(values[2].Int())
			drivlyVal.Retail = drivlyVal.RetailAverage
		case gjson.GetBytes(drivlyJSON, "retail.kelley.book").Exists():
			drivlyVal.RetailSource = "drivly:kelley"
			drivlyVal.Retail = int(gjson.GetBytes(drivlyJSON, "retail.kelley.book").Int())
		case gjson.GetBytes(drivlyJSON, "retail.edmunds.average").Exists():
			drivlyVal.RetailSource = "drivly:edmunds"
			values := gjson.GetManyBytes(drivlyJSON, "retail.edmunds.rough", "retail.edmunds.average", "retail.edmunds.clean")
			drivlyVal.RetailRough = int(values[0].Int())
			drivlyVal.RetailAverage = int(values[1].Int())
			drivlyVal.RetailClean = int(values[2].Int())
			drivlyVal.Retail = drivlyVal.RetailAverage
		case gjson.GetBytes(drivlyJSON, "retail").Exists() && !gjson.GetBytes(drivlyJSON, "retail").IsObject():
			drivlyVal.RetailSource = "drivly"
			drivlyVal.Retail = int(gjson.GetBytes(drivlyJSON, "retail").Int())
		default:
			logger.Error().Msg("Unexpected structure for driv.ly pricing data retail values")
		}
		dVal.ValuationSets = append(dVal.ValuationSets, drivlyVal)
	}

	// Blackbook latest data
	var deviceMileage int
	var deviceOdometerEvent null.Time
	deviceData, err := models.UserDeviceData(
		models.UserDeviceDatumWhere.UserDeviceID.EQ(udi),
		models.UserDeviceDatumWhere.Data.IsNotNull(),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	} else {
		deviceOdometer := gjson.GetBytes(deviceData.Data.JSON, "odometer")
		if deviceOdometer.Exists() && deviceData.LastOdometerEventAt.Valid {
			deviceMileage = int(deviceOdometer.Float() / services.MilesToKmFactor)
			deviceOdometerEvent = deviceData.LastOdometerEventAt
		}
	}
	blackbookVinData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(udi)),
		models.ExternalVinDatumWhere.BlackbookMetadata.IsNotNull(),
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if blackbookVinData != nil {
		blackbookVal := ValuationSet{
			Vendor: "blackbook",
		}
		requestJSON := drivlyVinData.RequestMetadata.JSON
		if deviceOdometerEvent.Valid {
			blackbookVal.Updated = deviceOdometerEvent.Time.Format(time.RFC3339)
			blackbookVal.Mileage = deviceMileage
		} else {
			blackbookVal.Updated = blackbookVinData.UpdatedAt.Format(time.RFC3339)
			requestMileage := gjson.GetBytes(requestJSON, "mileage")
			if requestMileage.Exists() {
				blackbookVal.Mileage = int(requestMileage.Int())
			}
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			blackbookVal.ZipCode = requestZipCode.String()
		}
		type BlackbookValuation struct {
			UsedVehicles struct {
				UsedVehiclesList []struct {
					// BaseTradeinClean int `json:"base_tradein_clean"`
					// BaseTradeinAvg   int `json:"base_tradein_avg"`
					// BaseTradeinRough int `json:"base_tradein_rough"`
					MileageTradeinClean  int `json:"mileage_tradein_clean"`
					MileageTradeinAvg    int `json:"mileage_tradein_avg"`
					MileageTradeinRough  int `json:"mileage_tradein_rough"`
					AdjustedTradeinClean int `json:"adjusted_tradein_clean"`
					AdjustedTradeinAvg   int `json:"adjusted_tradein_avg"`
					AdjustedTradeinRough int `json:"adjusted_tradein_rough"`
					MileageList          []struct {
						RangeBegin int `json:"range_begin"`
						RangeEnd   int `json:"range_end"`
						Clean      int `json:"clean"`
						Avg        int `json:"avg"`
						Rough      int `json:"rough"`
					} `json:"mileage_list"`
				} `json:"used_vehicle_list"`
			} `json:"used_vehicles"`
		}
		bbvv := &BlackbookValuation{}
		err = json.Unmarshal(blackbookVinData.BlackbookMetadata.JSON, bbvv)
		if err != nil {
			return err
		}
		bbval := bbvv.UsedVehicles.UsedVehiclesList[0]

		// Using adjusted values to include regional adjustment if available
		blackbookVal.TradeInClean = bbval.AdjustedTradeinClean
		blackbookVal.TradeInAverage = bbval.AdjustedTradeinAvg
		blackbookVal.TradeInRough = bbval.AdjustedTradeinRough
		// Conditional prevents us from applying mileage adjustment on top of already mileage adjusted values
		if blackbookVal.Mileage > 0 && len(bbval.MileageList) > 1 && bbval.MileageTradeinClean == 0 && bbval.MileageTradeinAvg == 0 && bbval.MileageTradeinRough == 0 {
			for _, v := range bbval.MileageList {
				if v.RangeBegin <= blackbookVal.Mileage && v.RangeEnd >= blackbookVal.Mileage || v == bbval.MileageList[len(bbval.MileageList)-1] {
					blackbookVal.TradeInClean += v.Clean
					blackbookVal.TradeInAverage += v.Avg
					blackbookVal.TradeInRough += v.Rough
					break
				}
			}
		}
		blackbookVal.TradeIn = blackbookVal.TradeInAverage
		dVal.ValuationSets = append(dVal.ValuationSets, blackbookVal)
	}

	return c.JSON(dVal)

}

type DeviceOffer struct {
	// Contains a list of offer sets, one for each source
	OfferSets []OfferSet `json:"offerSets"`
}
type OfferSet struct {
	// The source of the offers (eg. "drivly")
	Source string `json:"source"`
	// The time the offers were pulled
	Updated string `json:"updated,omitempty"`
	// The mileage used for the offers
	Mileage int `json:"mileage,omitempty"`
	// This will be the zip code used (if any) for the offers request regardless if the source uses it
	ZipCode string `json:"zipCode,omitempty"`
	// Contains a list of offers from the source
	Offers []Offer `json:"offers"`
}
type Offer struct {
	// The vendor of the offer (eg. "carmax", "carvana", etc.)
	Vendor string `json:"vendor"`
	// The offer price from the vendor
	Price int `json:"price,omitempty"`
	// The offer URL from the vendor
	URL string `json:"url,omitempty"`
	// An error from the vendor (eg. when the VIN is invalid)
	Error string `json:"error,omitempty"`
	// The grade of the offer from the vendor (eg. "RETAIL")
	Grade string `json:"grade,omitempty"`
	// The reason the offer was declined from the vendor
	DeclineReason string `json:"declineReason,omitempty"`
}

// GetOffers godoc
// @Description gets offers for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceOffer
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/offers [get]
func (udc *UserDevicesController) GetOffers(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	// Ensure user is owner of user device
	userDeviceExists, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if !userDeviceExists {
		return c.SendStatus(fiber.StatusForbidden)
	}

	dOffer := DeviceOffer{
		OfferSets: []OfferSet{},
	}

	// Drivly data
	drivlyVinData, err := models.ExternalVinData(
		models.ExternalVinDatumWhere.UserDeviceID.EQ(null.StringFrom(udi)),
		models.ExternalVinDatumWhere.OfferMetadata.IsNotNull(), // offer_metadata is sourced from drivly
		qm.OrderBy("updated_at desc"),
		qm.Limit(1)).One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if drivlyVinData != nil {
		drivlyOffers := OfferSet{}
		drivlyOffers.Source = "drivly"
		drivlyJSON := drivlyVinData.OfferMetadata.JSON
		requestJSON := drivlyVinData.RequestMetadata.JSON
		drivlyOffers.Updated = drivlyVinData.UpdatedAt.Format(time.RFC3339)
		requestMileage := gjson.GetBytes(requestJSON, "mileage")
		if requestMileage.Exists() {
			drivlyOffers.Mileage = int(requestMileage.Int())
		}
		requestZipCode := gjson.GetBytes(requestJSON, "zipCode")
		if requestZipCode.Exists() {
			drivlyOffers.ZipCode = requestZipCode.String()
		}
		// Drivly Offers
		gjson.GetBytes(drivlyJSON, `@keys.#(%"*Price")#`).ForEach(func(key, value gjson.Result) bool {
			offer := Offer{}
			offer.Vendor = strings.TrimSuffix(value.String(), "Price") // eg. vroom, carvana, or carmax
			gjson.GetBytes(drivlyJSON, `@keys.#(%"`+offer.Vendor+`*")#`).ForEach(func(key, value gjson.Result) bool {
				prop := strings.TrimPrefix(value.String(), offer.Vendor)
				if prop == "Url" {
					prop = "URL"
				}
				if !reflect.ValueOf(&offer).Elem().FieldByName(prop).CanSet() {
					return true
				}
				val := gjson.GetBytes(drivlyJSON, value.String())
				switch val.Type {
				case gjson.Null: // ignore null values
					return true
				case gjson.Number: // for "Price"
					reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(int(val.Int())))
				case gjson.JSON: // for "Error"
					if prop == "Error" {
						val = gjson.GetBytes(drivlyJSON, value.String()+".error.title")
						if val.Exists() {
							offer.Error = val.String()
							// reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(val.String()))
						}
					}
				default: // for everything else
					reflect.ValueOf(&offer).Elem().FieldByName(prop).Set(reflect.ValueOf(val.String()))
				}
				return true
			})
			drivlyOffers.Offers = append(drivlyOffers.Offers, offer)
			return true
		})
		dOffer.OfferSets = append(dOffer.OfferSets, drivlyOffers)
	}

	return c.JSON(dOffer)

}

type DeviceRange struct {
	// Contains a list of range sets, one for each range basis (may be empty)
	RangeSets []RangeSet `json:"rangeSets"`
}

type RangeSet struct {
	// The time the data was collected
	Updated string `json:"updated"`
	// The basis for the range calculation (eg. "MPG" or "MPG Highway")
	RangeBasis string `json:"rangeBasis"`
	// The estimated range distance
	RangeDistance int `json:"rangeDistance"`
	// The unit used for the rangeDistance (eg. "miles" or "kilometers")
	RangeUnit string `json:"rangeUnit"`
}

// GetRange godoc
// @Description gets the estimated range for a particular user device
// @Tags        user-devices
// @Produce     json
// @Success     200 {object} controllers.DeviceRange
// @Security    BearerAuth
// @Param       userDeviceID path string true "user device id"
// @Router      /user/devices/{userDeviceID}/range [get]
func (udc *UserDevicesController) GetRange(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	// Ensure user is owner of user device
	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceData),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}

	dds, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{userDevice.DeviceDefinitionID})
	if err != nil {
		return api.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+userDevice.DeviceDefinitionID)
	}

	deviceRange := DeviceRange{
		RangeSets: []RangeSet{},
	}
	udd := userDevice.R.UserDeviceData
	if len(dds) > 0 && dds[0].VehicleData != nil && len(udd) > 0 {
		vd := dds[0].VehicleData
		sortByJSONFieldMostRecent(udd, "fuelPercentRemaining")
		fuelPercentRemaining := gjson.GetBytes(udd[0].Data.JSON, "fuelPercentRemaining")
		dataUpdatedOn := gjson.GetBytes(udd[0].Data.JSON, "timestamp").Time()
		if fuelPercentRemaining.Exists() && vd.FuelTankCapacityGal > 0 && vd.MPG > 0 {
			fuelTankAtGal := vd.FuelTankCapacityGal * float32(fuelPercentRemaining.Float())
			rangeSet := RangeSet{
				Updated:       dataUpdatedOn.Format(time.RFC3339),
				RangeBasis:    "MPG",
				RangeDistance: int(vd.MPG * fuelTankAtGal),
				RangeUnit:     "miles",
			}
			deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
			if vd.MPGHighway > 0 {
				rangeSet.RangeBasis = "MPG Highway"
				rangeSet.RangeDistance = int(vd.MPGHighway * fuelTankAtGal)
				deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
			}
		}
		sortByJSONFieldMostRecent(udd, "range")
		reportedRange := gjson.GetBytes(udd[0].Data.JSON, "range")
		dataUpdatedOn = gjson.GetBytes(udd[0].Data.JSON, "timestamp").Time()
		if reportedRange.Exists() {
			reportedRangeMiles := int(reportedRange.Float() / services.MilesToKmFactor)
			rangeSet := RangeSet{
				Updated:       dataUpdatedOn.Format(time.RFC3339),
				RangeBasis:    "Vehicle Reported",
				RangeDistance: reportedRangeMiles,
				RangeUnit:     "miles",
			}
			deviceRange.RangeSets = append(deviceRange.RangeSets, rangeSet)
		}
	}

	return c.JSON(deviceRange)

}

// DeleteUserDevice godoc
// @Description delete the user device record (hard delete)
// @Tags        user-devices
// @Param       userDeviceID path string true "user id"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID} [delete]
func (udc *UserDevicesController) DeleteUserDevice(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint
	// todo grpc pull device-definitions via grpc
	userDevice, err := models.UserDevices(
		qm.Where("id = ?", udi),
		qm.And("user_id = ?", userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations), // Probably don't need this one.
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return api.ErrorResponseHandler(c, err, fiber.StatusNotFound)
		}
		return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	deviceDefinitionResponse, err := udc.DeviceDefSvc.GetDeviceDefinitionsByIDs(c.Context(), []string{userDevice.DeviceDefinitionID})

	if err != nil {
		return api.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+userDevice.DeviceDefinitionID)
	}

	if len(deviceDefinitionResponse) == 0 {
		udc.log.Err(err).
			Str("userDeviceID", udi).
			Str("deviceDefinitionID", userDevice.DeviceDefinitionID).
			Msg("unexpected error deregistering autopi")

		return api.ErrorResponseHandler(c, errors.New("no device definition"), fiber.StatusBadRequest)
	}

	var dd = deviceDefinitionResponse[0]

	for _, apiInteg := range userDevice.R.UserDeviceAPIIntegrations {

		integration, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), apiInteg.IntegrationID)

		if err != nil {
			return api.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+apiInteg.IntegrationID)
		}

		if integration.Vendor == constants.SmartCarVendor {
			if apiInteg.ExternalID.Valid {
				if apiInteg.TaskID.Valid {
					err = udc.smartcarTaskSvc.StopPoll(apiInteg)
					if err != nil {
						return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
					}
				}
				// Otherwise, it was on a webhook and we were never able to create a task for it.
			}
		} else if integration.Vendor == "Tesla" {
			if apiInteg.ExternalID.Valid {
				if err := udc.teslaTaskService.StopPoll(apiInteg); err != nil {
					return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
				}
			}
		} else if integration.Vendor == constants.AutoPiVendor {
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
					Make:  dd.Make.Name,
					Model: dd.Type.Model,
					Year:  int(dd.Type.Year),
				},
				Integration: services.UserDeviceEventIntegration{
					ID:     integration.Id,
					Type:   integration.Type,
					Style:  integration.Style,
					Vendor: integration.Vendor,
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
		return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		return api.ErrorResponseHandler(c, err, fiber.StatusInternalServerError)
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.delete",
		Subject: userID,
		Source:  "devices-api",
		Data: UserDeviceEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    udi,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year), // Odd.
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed emitting device deletion event")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetMintDataToSign godoc
// @Description Returns the data the user must sign in order to mint this device.
// @Tags        user-devices
// @Param       userDeviceID path     string true "user device ID"
// @Success     200          {object} signer.TypedData
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [get]
func (udc *UserDevicesController) GetMintDataToSign(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	dd, err2 := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err2 != nil {
		return api.GrpcErrorToFiber(err2, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	if dd.Make.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Device make %s not yet minted.", dd.Make.Name))
	}

	mkTok := big.NewInt(int64(dd.Make.TokenId))

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
				dd.Make.Name,
				dd.Type.Model,
				strconv.Itoa(int(dd.Type.Year)),
			},
		},
	}

	return c.JSON(typedData)
}

// GetMintDataToSignV2 godoc
// @Description Returns the data the user must sign in order to mint this device.
// @Tags        user-devices
// @Param       userDeviceID path     string true "user device ID"
// @Success     200          {object} signer.TypedData
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [get]
func (udc *UserDevicesController) GetMintDataToSignV2(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err != nil {
		return api.GrpcErrorToFiber(err, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	if dd.Make.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, "Device make not yet minted.")
	}
	makeTokenID := big.NewInt(int64(dd.Make.TokenId))

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
		return fiber.NewError(fiber.StatusBadRequest, "User does not have an Ethereum address on file.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(int64(udc.Settings.NFTChainID)),
			Address: common.HexToAddress("0x72b7268bD15EC670BfdA1445bD380C9400F4b1A6"),
			Name:    "DIMO",
			Version: "1",
		},
	}

	deviceMake := dd.Make.Name
	deviceModel := dd.Type.Model
	deviceYear := strconv.Itoa(int(dd.Type.Year))

	mvs := registry.MintVehicleSign{
		ManufacturerNode: makeTokenID,
		Owner:            common.HexToAddress(*user.EthereumAddress),
		Attributes:       []string{"Make", "Model", "Year"},
		Infos:            []string{deviceMake, deviceModel, deviceYear},
	}

	return c.JSON(client.GetPayload(&mvs))
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

func recoverAddress2(hash []byte, sig []byte) (common.Address, error) {
	fixedSig := make([]byte, len(sig))
	copy(fixedSig, sig)
	fixedSig[64] -= 27

	uncPubKey, err := crypto.Ecrecover(hash, fixedSig)
	if err != nil {
		return common.Address{}, err
	}

	pubKey, err := crypto.UnmarshalPubkey(uncPubKey)
	if err != nil {
		return common.Address{}, err
	}

	return crypto.PubkeyToAddress(*pubKey), nil
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
// @Description Sends a mint device request to the blockchain
// @Tags        user-devices
// @Param       userDeviceID path string                  true "user device ID"
// @Param       mintRequest  body controllers.MintRequest true "Signature and NFT data"
// @Success     200
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [post]
func (udc *UserDevicesController) MintDevice(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.MintRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err != nil {
		return api.GrpcErrorToFiber(err, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	if dd.Make.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, "Device make not yet minted.")
	}

	mkBI := big.NewInt(int64(dd.Make.TokenId))
	makeTokenID := (*math.HexOrDecimal256)(mkBI)

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

		integration, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), apiInt.IntegrationID)

		if err != nil {
			return api.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+apiInt.IntegrationID)
		}

		switch integration.Vendor {
		case constants.SmartCarVendor, constants.TeslaVendor:
			eligible = true
			// Sure hope this works!
			mreq.Vin = userDevice.VinIdentifier
		case constants.AutoPiVendor:
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

	if len(image) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Empty image field.")
	}

	// This may not be there, but if it is we should delete it.
	imageDataTransparent := strings.TrimPrefix(mr.ImageDataTransparent, "data:image/png;base64,")

	imageTransparent, err := base64.StdEncoding.DecodeString(imageDataTransparent)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Field imageDataTransparent not properly base64-encoded.")
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

	if len(imageTransparent) != 0 {
		_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
			Bucket: &udc.Settings.NFTS3Bucket,
			Key:    aws.String(mintRequestID + "_transparent.png"), // This will be the request ID.
			Body:   bytes.NewReader(imageTransparent),
		})
		if err != nil {
			logger.Err(err).Msg("Failed to save transparent image to S3.")
			return opaqueInternalError
		}
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
			"rootNode":   makeTokenID,
			"_owner":     *user.EthereumAddress,
			"attributes": []any{"Make", "Model", "Year"},
			"infos": []any{
				dd.Make.Name,
				dd.Type.Model,
				strconv.Itoa(int(dd.Type.Year)),
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
				dd.Make.Name,
				dd.Type.Model,
				strconv.Itoa(int(dd.Type.Year)),
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

// MintDeviceV2 godoc
// @Description Sends a mint device request to the blockchain
// @Tags        user-devices
// @Param       userDeviceID path string                  true "user device ID"
// @Param       mintRequest  body controllers.MintRequest true "Signature and NFT data"
// @Success     200
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/commands/mint [post]
func (udc *UserDevicesController) MintDeviceV2(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	userID := api.GetUserID(c)

	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "No device with that ID found.")
	}

	dd, err2 := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), userDevice.DeviceDefinitionID)
	if err2 != nil {
		return api.GrpcErrorToFiber(err2, fmt.Sprintf("error querying for device definition id: %s ", userDevice.DeviceDefinitionID))
	}

	if dd.Make.TokenId == 0 {
		return fiber.NewError(fiber.StatusConflict, "Device make not yet minted.")
	}

	makeTokenID := big.NewInt(int64(dd.Make.TokenId))

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
		return fiber.NewError(fiber.StatusBadRequest, "User does not have an Ethereum address on file.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(int64(udc.Settings.NFTChainID)),
			Address: common.HexToAddress("0x72b7268bD15EC670BfdA1445bD380C9400F4b1A6"),
			Name:    "DIMO",
			Version: "1",
		},
	}

	deviceMake := dd.Make.Name
	deviceModel := dd.Type.Model
	deviceYear := strconv.Itoa(int(dd.Type.Year))

	mvs := registry.MintVehicleSign{
		ManufacturerNode: makeTokenID,
		Owner:            common.HexToAddress(*user.EthereumAddress),
		Attributes:       []string{"Make", "Model", "Year"},
		Infos:            []string{deviceMake, deviceModel, deviceYear},
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

	if len(image) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Empty image field.")
	}

	requestID := ksuid.New().String()

	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("requestId", requestID).
		Str("handler", "MintDevice").
		Str("feature", "identity").
		Logger()

	logger.Info().
		Interface("httpRequestBody", mr).
		Interface("client", client).
		Interface("mintVehicleSign", mvs).
		Interface("typedData", client.GetPayload(&mvs)).
		Msg("Got request.")

	_, err = udc.s3.PutObject(c.Context(), &s3.PutObjectInput{
		Bucket: &udc.Settings.NFTS3Bucket,
		Key:    aws.String(userDeviceID + ".png"),
		Body:   bytes.NewReader(image),
	})
	if err != nil {
		logger.Err(err).Msg("Failed to save image to S3.")
		return opaqueInternalError
	}

	hash, err := client.Hash(&mvs)
	if err != nil {
		return opaqueInternalError
	}

	sigBytes := common.FromHex(mr.Signature)

	sigBytesYellowPaper := make([]byte, len(sigBytes))
	copy(sigBytesYellowPaper, sigBytes)
	sigBytesYellowPaper[64] -= 27

	recUncPubKey, err := crypto.Ecrecover(hash[:], sigBytesYellowPaper)
	if err != nil {
		return err
	}

	recPubKey, err := crypto.UnmarshalPubkey(recUncPubKey)
	if err != nil {
		return err
	}

	recAddr := crypto.PubkeyToAddress(*recPubKey)
	realAddr := common.HexToAddress(*user.EthereumAddress)

	if recAddr != realAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Signature incorrect.")
	}

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: "Unsubmitted",
	}
	err = mtr.Insert(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	userDevice.MintMetaTransactionRequestID = null.StringFrom(requestID)
	_, err = userDevice.Update(c.Context(), udc.DBS().Writer, boil.Infer())
	if err != nil {
		return err
	}

	udc.log.Info().Str("userDeviceId", userDevice.ID).Str("requestId", requestID).Msg("Submitted metatransaction request.")

	return client.MintVehicleSign(requestID, makeTokenID, realAddr, mvs.Attributes, mvs.Infos, sigBytes)
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
	// ImageDataTransparent contains the base64-encoded NFT PNG image
	// with a transparent background, for use in the app.
	ImageDataTransparent string `json:"imageDataTransparent"`
}

type RegisterUserDevice struct {
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
		validation.Field(&reg.DeviceDefinitionID, validation.Required),
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

	validateLengthAndChars := validation.ValidateStruct(u,
		// vin must be 17 characters in length, alphanumeric
		validation.Field(&u.VIN, validation.Required, validation.Match(regexp.MustCompile("^[A-Z0-9]{17}$"))),
	)
	if validateLengthAndChars != nil {
		return validateLengthAndChars
	}

	return nil
}

func (u *UpdateNameReq) validate() error {

	return validation.ValidateStruct(u,
		// name must be between 1 and 16 alphanumeric characters in length (spaces are not allowed)
		// NOTE: this captures characters in the latin/ chinese/ cyrillic alphabet but doesn't work as well for thai or arabic
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`^[\p{L}\p{N}\p{M}# ,’.@!$'":_-]{1,25}$`))),
		// cannot start with space
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`^[^\s]`))),
		// cannot end with space
		validation.Field(&u.Name, validation.Required, validation.Match(regexp.MustCompile(`.+[^\s]$|[^\s]$`))),
	)
}

// sortByJSONFieldMostRecent Sort user device data so the latest that has the specified field is first
func sortByJSONFieldMostRecent(udd models.UserDeviceDatumSlice, field string) {
	sort.Slice(udd, func(i, j int) bool {
		fpri := gjson.GetBytes(udd[i].Data.JSON, field)
		fprj := gjson.GetBytes(udd[j].Data.JSON, field)
		if fpri.Exists() && !fprj.Exists() {
			return true
		} else if !fpri.Exists() && fprj.Exists() {
			return false
		}
		return udd[i].UpdatedAt.After(udd[j].UpdatedAt)
	})
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
	OptedInAt        *time.Time                    `json:"optedInAt"`
}

type NFTData struct {
	TokenID  *big.Int `json:"tokenId,omitempty" swaggertype:"number" example:"37"`
	TokenURI string   `json:"tokenUri,omitempty" example:"https://nft.dimo.zone/37"`
	// TxHash is the hash of the minting transaction.
	TxHash *string `json:"txHash,omitempty" example:"0x30bce3da6985897224b29a0fe064fd2b426bb85a394cc09efe823b5c83326a8e"`
	// Status is the minting status of the NFT.
	Status string `json:"status" enums:"Unstarted,Submitted,Mined,Confirmed" example:"Confirmed"`
}
