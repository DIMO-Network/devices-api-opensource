package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	smartcar "github.com/smartcar/go-sdk"

	ddgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/devices-api/internal/constants"
	"github.com/DIMO-Network/devices-api/internal/controllers/helpers"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/internal/services/registry"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/users"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	signer "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/semver"
)

// GetUserDeviceIntegration godoc
// @Description Receive status updates about a Smartcar integration
// @Tags        integrations
// @Success     200 {object} controllers.GetUserDeviceIntegrationResponse
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID} [get]
func (udc *UserDevicesController) GetUserDeviceIntegration(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")
	deviceExists, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if !deviceExists {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no user device with ID %s", userDeviceID))
	}

	apiIntegration, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("user device %s does not have integration %s", userDeviceID, integrationID))
		}
		return err
	}
	return c.JSON(GetUserDeviceIntegrationResponse{Status: apiIntegration.Status, ExternalID: apiIntegration.ExternalID, CreatedAt: apiIntegration.CreatedAt})
}

func (udc *UserDevicesController) deleteDeviceIntegration(ctx context.Context, userID, userDeviceID, integrationID string, dd *ddgrpc.GetDeviceDefinitionItemResponse) error {
	tx, err := udc.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	apiInt, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.SerialAftermarketDevice),
	).One(ctx, tx)
	if err != nil {
		return err
	}

	integ, err := udc.DeviceDefSvc.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+integrationID)
	}

	switch integ.Vendor {
	case constants.SmartCarVendor:
		if apiInt.TaskID.Valid {
			err = udc.smartcarTaskSvc.StopPoll(apiInt)
			if err != nil {
				return err
			}
		}
	case constants.TeslaVendor:
		if apiInt.TaskID.Valid {
			err = udc.teslaTaskService.StopPoll(apiInt)
			if err != nil {
				return err
			}
		}
	case constants.AutoPiVendor:
		if unit := apiInt.R.SerialAftermarketDevice; unit != nil && unit.PairRequestID.Valid {
			return fiber.NewError(fiber.StatusConflict, "Must un-pair on-chain before deleting integration.")
		}

		err = udc.autoPiIngestRegistrar.Deregister(apiInt.ExternalID.String, apiInt.UserDeviceID, apiInt.IntegrationID)
		if err != nil {
			return err
		}
	}

	_, err = apiInt.Delete(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = udc.eventService.Emit(&shared.CloudEvent[any]{
		Type:    "com.dimo.zone.device.integration.delete",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  dd.Make.Name,
				Model: dd.Type.Model,
				Year:  int(dd.Type.Year),
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     integ.Id,
				Type:   integ.Type,
				Style:  integ.Style,
				Vendor: integ.Vendor,
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed to emit integration deletion")
	}

	return err
}

// DeleteUserDeviceIntegration godoc
// @Description Remove an integration from a device.
// @Tags        integrations
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID} [delete]
func (udc *UserDevicesController) DeleteUserDeviceIntegration(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint

	device, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID)),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenSyntheticDevice)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id.")
		}
		return err
	}

	if len(device.R.UserDeviceAPIIntegrations) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Device does not have that integration.")
	}

	if device.R.VehicleNFT != nil && device.R.VehicleNFT.R.VehicleTokenSyntheticDevice != nil {
		sd := device.R.VehicleNFT.R.VehicleTokenSyntheticDevice

		integr, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), integrationID)
		if err != nil {
			return err
		}

		integrTokenID, _ := device.R.VehicleNFT.R.VehicleTokenSyntheticDevice.IntegrationTokenID.Uint64()
		if integr.TokenId == integrTokenID {
			if sd.BurnRequestID.Valid {
				return fiber.NewError(fiber.StatusConflict, "Synthetic device burn in progress.")
			}
			return fiber.NewError(fiber.StatusConflict, "Burn synthetic device before deleting integration.")
		}
	}

	// Need this for activity log.
	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), device.DeviceDefinitionID)
	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting definition id: "+device.DeviceDefinitionID)
	}

	err = udc.deleteDeviceIntegration(c.Context(), userID, userDeviceID, integrationID, dd)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetCommandRequestStatus godoc
// @Summary     Get the status of a submitted command.
// @Description Get the status of a submitted command by request id.
// @ID          get-command-request-status
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandRequestStatusResp
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Param       requestID path string true "Command request ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/{requestID} [get]
func (udc *UserDevicesController) GetCommandRequestStatus(c *fiber.Ctx) error {
	requestID := c.Params("requestID")

	// Don't actually validate userDeviceID or integrationID, just following a URL pattern.
	// Is this beyond the pale?
	cr, err := models.DeviceCommandRequests(
		models.DeviceCommandRequestWhere.ID.EQ(requestID),
		qm.Load(models.DeviceCommandRequestRels.UserDevice),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No command request with that id found.")
		}
		udc.log.Err(err).Msg("Failed to search for command status.")
		return opaqueInternalError
	}

	dcr := CommandRequestStatusResp{
		ID:        requestID,
		Command:   cr.Command,
		Status:    cr.Status,
		CreatedAt: cr.CreatedAt,
		UpdatedAt: cr.UpdatedAt,
	}

	return c.JSON(dcr)
}

type CommandRequestStatusResp struct {
	ID        string    `json:"id" example:"2D8LqUHQtaMHH6LYPqznmJMBeZm"`
	Command   string    `json:"command" example:"doors/unlock"`
	Status    string    `json:"status" enums:"Pending,Complete,Failed" example:"Complete"`
	CreatedAt time.Time `json:"createdAt" example:"2022-08-09T19:38:39Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2022-08-09T19:39:22Z"`
}

// handleEnqueueCommand enqueues the command specified by commandPath with the
// appropriate task service.
//
// Grabs user ID, device ID, and integration ID from Ctx.
func (udc *UserDevicesController) handleEnqueueCommand(c *fiber.Ctx, commandPath string) error {
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	logger := helpers.GetLogger(c, udc.log)

	logger.Info().Msg("Received command request.")

	// Checking both that the device exists and that the user owns it.
	deviceOK, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
	).Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		logger.Err(err).Msg("Failed to search for device.")
		return opaqueInternalError
	}

	if !deviceOK {
		return fiber.NewError(fiber.StatusNotFound, "Device not found.")
	}

	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Integration not found for this device.")
		}
		logger.Err(err).Msg("Failed to search for device integration record.")
		return opaqueInternalError
	}

	if udai.Status != models.UserDeviceAPIIntegrationStatusActive {
		return fiber.NewError(fiber.StatusConflict, "Integration is not active for this device.")
	}

	md := new(services.UserDeviceAPIIntegrationsMetadata)
	if err := udai.Metadata.Unmarshal(md); err != nil {
		logger.Err(err).Msg("Couldn't parse metadata JSON.")
		return opaqueInternalError
	}

	// TODO(elffjs): This map is ugly. Surely we interface our way out of this?
	commandMap := map[string]map[string]func(udai *models.UserDeviceAPIIntegration) (string, error){
		constants.SmartCarVendor: {
			"doors/unlock": udc.smartcarTaskSvc.UnlockDoors,
			"doors/lock":   udc.smartcarTaskSvc.LockDoors,
		},
		constants.TeslaVendor: {
			"doors/unlock": udc.teslaTaskService.UnlockDoors,
			"doors/lock":   udc.teslaTaskService.LockDoors,
			"trunk/open":   udc.teslaTaskService.OpenTrunk,
			"frunk/open":   udc.teslaTaskService.OpenFrunk,
		},
	}

	integration, err := udc.DeviceDefSvc.GetIntegrationByID(c.Context(), udai.IntegrationID)

	if err != nil {
		return shared.GrpcErrorToFiber(err, "deviceDefSvc error getting integration id: "+udai.IntegrationID)
	}

	vendorCommandMap, ok := commandMap[integration.Vendor]
	if !ok {
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command.")
	}

	// This correctly handles md.Commands.Enabled being nil.
	if !slices.Contains(md.Commands.Enabled, commandPath) {
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command with this device.")
	}

	commandFunc, ok := vendorCommandMap[commandPath]
	if !ok {
		// Should never get here.
		logger.Error().Msg("Command was enabled for this device, but there is no function to execute it.")
		return fiber.NewError(fiber.StatusConflict, "Integration is not capable of this command.")
	}

	subTaskID, err := commandFunc(udai)
	if err != nil {
		logger.Err(err).Msg("Failed to start command task.")
		return opaqueInternalError
	}

	comRow := &models.DeviceCommandRequest{
		ID:            subTaskID,
		UserDeviceID:  userDeviceID,
		IntegrationID: integrationID,
		Command:       commandPath,
		Status:        models.DeviceCommandRequestStatusPending,
	}

	if err := comRow.Insert(c.Context(), udc.DBS().Writer, boil.Infer()); err != nil {
		logger.Err(err).Msg("Couldn't insert device command request record.")
		return opaqueInternalError
	}

	logger.Info().Msg("Successfully enqueued command.")

	return c.JSON(CommandResponse{RequestID: subTaskID})
}

type CommandResponse struct {
	RequestID string `json:"requestId"`
}

// UnlockDoors godoc
// @Summary     Unlock the device's doors
// @Description Unlock the device's doors.
// @ID          unlock-doors
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/doors/unlock [post]
func (udc *UserDevicesController) UnlockDoors(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "doors/unlock")
}

// LockDoors godoc
// @Summary     Lock the device's doors
// @Description Lock the device's doors.
// @ID          lock-doors
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/doors/lock [post]
func (udc *UserDevicesController) LockDoors(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "doors/lock")
}

// OpenTrunk godoc
// @Summary     Open the device's rear trunk
// @Description Open the device's front trunk. Currently, this only works for Teslas connected through Tesla.
// @ID          open-trunk
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/trunk/open [post]
func (udc *UserDevicesController) OpenTrunk(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "trunk/open")
}

// OpenFrunk godoc
// @Summary     Open the device's front trunk
// @Description Open the device's front trunk. Currently, this only works for Teslas connected through Tesla.
// @ID          open-frunk
// @Tags        device,integration,command
// @Success 200 {object} controllers.CommandResponse
// @Produce     json
// @Param       userDeviceID  path string true "Device ID"
// @Param       integrationID path string true "Integration ID"
// @Router      /user/devices/{userDeviceID}/integrations/{integrationID}/commands/frunk/open [post]
func (udc *UserDevicesController) OpenFrunk(c *fiber.Ctx) error {
	return udc.handleEnqueueCommand(c, "frunk/open")
}

// GetAutoPiUnitInfo godoc
// @Description gets the information about the aftermarket device by the hw serial
// @Tags        integrations
// @Produce     json
// @Param       serial path     string true "autopi unit id or macaron serial"
// @Success     200    {object} controllers.AutoPiDeviceInfo
// @Security    BearerAuth
// @Router      /aftermarket/device/by-serial/:serial [get]
func (udc *UserDevicesController) GetAutoPiUnitInfo(c *fiber.Ctx) error {
	const minimumAutoPiRelease = "v1.22.8" // correct semver has leading v

	serial := c.Locals("serial").(string)

	var claim, pair, unpair *AutoPiTransactionStatus

	var tokenID *big.Int
	var ethereumAddress, beneficiaryAddress *common.Address
	var ownerAddress *string // Frontend is doing a case-sensitive match.

	var mfr *ManufacturerInfo

	dbUnit, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Serial.EQ(serial),
		qm.Load(models.AftermarketDeviceRels.ClaimMetaTransactionRequest),
		qm.Load(models.AftermarketDeviceRels.PairRequest),
		qm.Load(models.AftermarketDeviceRels.UnpairRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		tokenID = dbUnit.TokenID.Int(nil)

		addr := common.BytesToAddress(dbUnit.EthereumAddress)
		ethereumAddress = &addr

		if dbUnit.OwnerAddress.Valid {
			addr := common.BytesToAddress(dbUnit.OwnerAddress.Bytes)
			addrStr := addr.Hex()
			ownerAddress = &addrStr
			beneficiaryAddress = &addr
			claim = &AutoPiTransactionStatus{
				Status: models.MetaTransactionRequestStatusConfirmed,
			}
		}

		if dbUnit.Beneficiary.Valid {
			addr := common.BytesToAddress(dbUnit.Beneficiary.Bytes)
			beneficiaryAddress = &addr
		}

		if req := dbUnit.R.ClaimMetaTransactionRequest; req != nil {
			claim = &AutoPiTransactionStatus{
				Status:    req.Status,
				CreatedAt: req.CreatedAt,
				UpdatedAt: req.UpdatedAt,
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				claim.Hash = &hash
			}
		}

		// Check for pair.
		if req := dbUnit.R.PairRequest; req != nil {
			pair = &AutoPiTransactionStatus{
				Status:    req.Status,
				CreatedAt: req.CreatedAt,
				UpdatedAt: req.UpdatedAt,
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				pair.Hash = &hash
			}
		}

		// Check for unpair.
		if req := dbUnit.R.UnpairRequest; req != nil {
			unpair = &AutoPiTransactionStatus{
				Status:    req.Status,
				CreatedAt: req.CreatedAt,
				UpdatedAt: req.UpdatedAt,
			}
			if req.Status != models.MetaTransactionRequestStatusUnsubmitted {
				hash := hexutil.Encode(req.Hash.Bytes)
				unpair.Hash = &hash
			}
		}

		tib := dbUnit.DeviceManufacturerTokenID.Int(nil)

		dm, err := udc.DeviceDefSvc.GetMakeByTokenID(c.Context(), tib)
		if err != nil {
			return err
		}

		mfr = &ManufacturerInfo{
			TokenID: tib,
			Name:    dm.Name,
		}
	}

	if mfr != nil && mfr.Name != constants.AutoPiVendor {
		// Might be a Macaron
		adi := AutoPiDeviceInfo{
			IsUpdated:          true,
			UnitID:             serial,
			ShouldUpdate:       false,
			TokenID:            tokenID,
			EthereumAddress:    ethereumAddress,
			OwnerAddress:       ownerAddress,
			BeneficiaryAddress: beneficiaryAddress,
			Claim:              claim,
			Pair:               pair,
			Unpair:             unpair,
			Manufacturer:       mfr,
		}

		return c.JSON(adi)
	}

	// This is hitting AutoPi.
	unit, err := udc.autoPiSvc.GetDeviceByUnitID(serial)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Serial %s unknown to AutoPi.", serial))
		}
		return err
	}

	// Must be an AutoPi.
	shouldUpdate := false
	if udc.Settings.IsProduction() {
		version := unit.Release.Version
		if string(unit.Release.Version[0]) != "v" {
			version = "v" + version
		}
		shouldUpdate = semver.Compare(version, minimumAutoPiRelease) < 0
	}

	adi := AutoPiDeviceInfo{
		IsUpdated:          unit.IsUpdated,
		DeviceID:           unit.ID,
		UnitID:             unit.UnitID,
		DockerReleases:     unit.DockerReleases,
		HwRevision:         unit.HwRevision,
		Template:           unit.Template,
		LastCommunication:  unit.LastCommunication,
		ReleaseVersion:     unit.Release.Version,
		ShouldUpdate:       shouldUpdate,
		TokenID:            tokenID,
		EthereumAddress:    ethereumAddress,
		OwnerAddress:       ownerAddress,
		BeneficiaryAddress: beneficiaryAddress,
		Claim:              claim,
		Pair:               pair,
		Unpair:             unpair,
		Manufacturer:       mfr,
	}
	return c.JSON(adi)
}

// StartAutoPiUpdateTask godoc
// @Description checks to see if aftermarket device needs to be updated, and starts update process if so.
// @Tags        integrations
// @Produce     json
// @Param       serial path     string true "autopi unit id", ie. physical barcode
// @Success     200    {object} services.AutoPiTask
// @Security    BearerAuth
// @Router      /aftermarket/device/by-serial/:serial/update [post]
func (udc *UserDevicesController) StartAutoPiUpdateTask(c *fiber.Ctx) error {
	unitID := c.Locals("serial").(string)
	userID := helpers.GetUserID(c)

	// check if device already updated
	unit, err := udc.autoPiSvc.GetDeviceByUnitID(unitID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Device not found.")
		}
		return err
	}

	if unit.IsUpdated {
		return c.JSON(services.AutoPiTask{
			TaskID:      "0",
			Status:      string(services.Success),
			Description: "autopi device is already up to date running version " + unit.Release.Version,
			Code:        200,
		})
	}

	// fire off task
	taskID, err := udc.autoPiTaskService.StartAutoPiUpdate(unit.ID, userID, unitID)
	if err != nil {
		return err
	}

	return c.JSON(services.AutoPiTask{
		TaskID:      taskID,
		Status:      "Pending",
		Description: "",
		Code:        100,
	})
}

// GetAutoPiClaimMessage godoc
// @Description Return the EIP-712 payload to be signed for Aftermarket device claiming.
// @Produce json
// @Param serial path string true "AutoPi unit id"
// @Success 200 {object} signer.TypedData
// @Security BearerAuth
// @Router /aftermarket/device/by-serial/:serial/commands/claim [get]
func (udc *UserDevicesController) GetAutoPiClaimMessage(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	unitID := c.Params("serial")

	logger := udc.log.With().Str("userId", userID).Str("serial", unitID).Logger()
	logger.Info().Msg("Got AutoPi claim request.")

	unit, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Serial.EQ(unitID),
		qm.Load(models.AftermarketDeviceRels.ClaimMetaTransactionRequest),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info().Msg("Unknown unit id.")
			return fiber.NewError(fiber.StatusNotFound, "Aftermarket device not minted.")
		}
		logger.Err(err).Msg("Database failure searching for AutoPi.")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal error.")
	}

	if unit.OwnerAddress.Valid {
		return fiber.NewError(fiber.StatusConflict, "Device already claimed.")
	}

	if unit.R.ClaimMetaTransactionRequest != nil && unit.R.ClaimMetaTransactionRequest.Status != "Failed" {
		return fiber.NewError(fiber.StatusConflict, "Claiming transaction in progress.")
	}

	apToken := unit.TokenID.Int(nil)

	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		udc.log.Error().Msg("No Ethereum address on file for user.")
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	cads := &registry.ClaimAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		Owner:                 common.HexToAddress(*user.EthereumAddress),
	}

	var out *signer.TypedData = client.GetPayload(cads)

	return c.JSON(out)
}

// GetAutoPiPairMessage godoc
// @Description Return the EIP-712 payload to be signed for Aftermarket device pairing. The device must
// @Description either already be integrated with the vehicle, or you must provide its unit id
// @Description as a query parameter. In the latter case, the integration process will start
// @Description once the transaction confirms.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Param external_id query string false "External id, for now AutoPi unit id"
// @Success 200 {object} signer.TypedData "EIP-712 message for pairing."
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/aftermarket/commands/pair [get]
func (udc *UserDevicesController) GetAutoPiPairMessage(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	logger := helpers.GetLogger(c, udc.log)

	logger.Info().Msg("Got aftermarket device pair payload request.")

	// This is only a query parameter because we want to maintain the path shape for POST.
	// We also had a legacy mode for web2-paired devices. This was never used in production.
	externalID := c.Query("external_id")

	vnft, ad, err := udc.checkPairable(c.Context(), udc.DBS().Reader, userDeviceID, externalID)
	if err != nil {
		return err
	}

	vehicleToken := vnft.TokenID.Int(nil)
	apToken := ad.TokenID.Int(nil)

	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	pads := &registry.PairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	return c.JSON(client.GetPayload(pads))
}

// PostPairAutoPi godoc
// @Description Submit the signature for pairing this device with its attached Aftermarket.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Param userSignature body controllers.AutoPiPairRequest true "User signature."
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/aftermarket/commands/pair [post]
func (udc *UserDevicesController) PostPairAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	logger := helpers.GetLogger(c, udc.log)

	logger.Info().Msg("Got aftermarket device pair submission request.")

	var pairReq AutoPiPairRequest
	err := c.BodyParser(&pairReq)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	logger.Info().Interface("request", pairReq).Msg("Pairing request body.")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	vnft, ad, err := udc.checkPairable(c.Context(), tx, userDeviceID, pairReq.ExternalID)
	if err != nil {
		return err
	}

	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	userAddr := common.HexToAddress(*user.EthereumAddress)

	apToken := ad.TokenID.Int(nil)
	vehicleToken := vnft.TokenID.Int(nil)

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	pads := registry.PairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	hash, err := client.Hash(&pads)
	if err != nil {
		return err
	}

	vehicleOwnerSig := pairReq.Signature

	if len(vehicleOwnerSig) != 65 {
		return fiber.NewError(fiber.StatusBadRequest, "User signature was not 65 bytes long.")
	}

	if recAddr, err := helpers.Ecrecover(hash, vehicleOwnerSig); err != nil {
		return err
	} else if recAddr != userAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Incorrect user signature.")
	}

	if common.BytesToAddress(ad.OwnerAddress.Bytes) != common.BytesToAddress(vnft.OwnerAddress.Bytes) {
		// We must be trying to do a host pair.
		aftermarketDeviceSig := pairReq.AftermarketDeviceSignature
		if len(aftermarketDeviceSig) != 65 {
			// Not great English.
			return fiber.NewError(fiber.StatusBadRequest, "Aftermarket device signature was not 65 bytes long.")
		}

		if recAddr, err := helpers.Ecrecover(hash, aftermarketDeviceSig); err != nil {
			return err
		} else if recAddr != common.BytesToAddress(ad.EthereumAddress) {
			return fiber.NewError(fiber.StatusBadRequest, "Incorrect aftermarket device signature.")
		}

		requestID := ksuid.New().String()

		mtr := models.MetaTransactionRequest{
			ID:     requestID,
			Status: models.MetaTransactionRequestStatusUnsubmitted,
		}
		err = mtr.Insert(c.Context(), tx, boil.Infer())
		if err != nil {
			return err
		}

		ad.UnpairRequestID = null.String{}
		ad.PairRequestID = null.StringFrom(requestID)
		_, err = ad.Update(c.Context(), tx, boil.Whitelist(models.AftermarketDeviceColumns.UnpairRequestID, models.AftermarketDeviceColumns.PairRequestID, models.AftermarketDeviceColumns.UpdatedAt))
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return client.PairAftermarketDeviceSignTwoOwners(requestID, apToken, vehicleToken, aftermarketDeviceSig, vehicleOwnerSig)
	}

	// Yes, this is ugly, we'll fix it.
	requestID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}
	err = mtr.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	ad.UnpairRequestID = null.String{}
	ad.PairRequestID = null.StringFrom(requestID)
	_, err = ad.Update(c.Context(), tx, boil.Whitelist(models.AftermarketDeviceColumns.UnpairRequestID, models.AftermarketDeviceColumns.PairRequestID, models.AftermarketDeviceColumns.UpdatedAt))
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return client.PairAftermarketDeviceSignSameOwner(requestID, apToken, vehicleToken, vehicleOwnerSig)
}

func (udc *UserDevicesController) checkPairable(ctx context.Context, exec boil.ContextExecutor, userDeviceID, serial string) (*models.VehicleNFT, *models.AftermarketDevice, error) {
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAftermarketDevice)),
	).One(ctx, exec)
	if err != nil {
		// Access middleware will catch "not found".
		return nil, nil, err
	}

	vnft := ud.R.VehicleNFT

	// Vehicle must be minted.
	if vnft == nil || vnft.TokenID.IsZero() {
		return nil, nil, fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	if serial == "" {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "Serial required.")
	}

	serial = strings.TrimSpace(strings.ToLower(serial))

	ad, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Serial.EQ(serial),
		qm.Load(models.AftermarketDeviceRels.PairRequest),
	).One(ctx, exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("No aftermarket device with serial %q known.", serial))
		}
		return nil, nil, err
	}

	if !ad.OwnerAddress.Valid {
		return nil, nil, fiber.NewError(fiber.StatusConflict, "Aftermarket device not yet claimed.")
	}

	// TODO(elffjs): It's difficult to tell if the vehicle is in the process of being paired.
	if vad := ud.R.VehicleNFT.R.VehicleTokenAftermarketDevice; vad != nil {
		if vad.TokenID.Cmp(vad.TokenID.Big) == 0 {
			return nil, nil, fiber.NewError(fiber.StatusConflict, "Specified vehicle and aftermarket device are already paired.")
		}
		return nil, nil, fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Vehicle already paired with aftermarket device %s.", vad.TokenID))
	}

	if ad.R.PairRequest != nil {
		if ad.R.PairRequest.Status == models.MetaTransactionRequestStatusConfirmed {
			return nil, nil, fiber.NewError(fiber.StatusConflict, "Aftermarket device already paired.")
		}
		return nil, nil, fiber.NewError(fiber.StatusConflict, "Aftermarket device already in the pairing process.")
	}

	return vnft, ad, nil
}

func (udc *UserDevicesController) checkUnpairable(ctx context.Context, exec boil.ContextExecutor, userDeviceID string) (*models.VehicleNFT, *models.AftermarketDevice, error) {
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAftermarketDevice)),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fiber.NewError(fiber.StatusNotFound, "No vehicle with that id found.")
		}
		return nil, nil, err
	}

	vnft := ud.R.VehicleNFT

	if vnft == nil || vnft.TokenID.IsZero() {
		return nil, nil, fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	if vnft.R.VehicleTokenAftermarketDevice == nil {
		return nil, nil, fiber.NewError(fiber.StatusConflict, "Vehicle not paired with an aftermarket device.")
	}

	ad := vnft.R.VehicleTokenAftermarketDevice

	if ad.UnpairRequestID.Valid {
		return nil, nil, fiber.NewError(fiber.StatusConflict, "Unpairing already in progress.")
	}

	return vnft, ad, nil
}

// CloudRepairAutoPi godoc
// @Description Re-apply AutoPi cloud actions in an attempt to get the device transmitting data again.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Success 204
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/aftermarket/commands/cloud-repair [post]
func (udc *UserDevicesController) CloudRepairAutoPi(c *fiber.Ctx) error {
	userDeviceID := c.Params("userDeviceID")

	logger := helpers.GetLogger(c, udc.log)
	logger.Info().Msg("Got Aftermarket pair request.")

	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(qm.Rels(models.UserDeviceRels.VehicleNFT, models.VehicleNFTRels.VehicleTokenAftermarketDevice)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "No device with that id found.")
		}
		logger.Err(err).Msg("Database failure searching for device.")
		return opaqueInternalError
	}

	if ud.R.VehicleNFT == nil || ud.R.VehicleNFT.TokenID.IsZero() {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not yet minted.")
	}

	if ud.R.VehicleNFT.R.VehicleTokenAftermarketDevice == nil {
		return fiber.NewError(fiber.StatusConflict, "Vehicle not paired on-chain with an aftermarket device.")
	}

	vehicleID := ud.R.VehicleNFT.TokenID.Int(nil)
	autoPiID := ud.R.VehicleNFT.R.VehicleTokenAftermarketDevice.TokenID.Int(nil)

	err = udc.autoPiIntegration.Pair(c.Context(), autoPiID, vehicleID)
	if err != nil {
		return err
	}

	return c.SendStatus(204)
}

// UnpairAutoPi godoc
// @Description Submit the signature for unpairing this device from its attached Aftermarket.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Param userSignature body controllers.AutoPiPairRequest true "User signature."
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/aftermarket/commands/unpair [post]
func (udc *UserDevicesController) UnpairAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	realAddr := common.HexToAddress(*user.EthereumAddress)

	userDeviceID := c.Params("userDeviceID")

	logger := helpers.GetLogger(c, udc.log)
	logger.Info().Msg("Got Aftermarket unpair request.")

	// TODO(elffjs): Is SELECT ... FOR UPDATE better here?
	tx, err := udc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	vnft, apnft, err := udc.checkUnpairable(c.Context(), tx, userDeviceID)
	if err != nil {
		return err
	}

	vehicleToken := vnft.TokenID.Int(nil)
	apToken := apnft.TokenID.Int(nil)

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	uads := registry.UnPairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	var pairReq AutoPiPairRequest
	err = c.BodyParser(&pairReq)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	hash, err := client.Hash(&uads)
	if err != nil {
		return err
	}

	sigBytes := pairReq.Signature
	if len(sigBytes) != 65 {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Signature has length %d, not the required 65.", len(sigBytes)))
	}

	recAddr, err := helpers.Ecrecover(hash, sigBytes)
	if err != nil {
		return err
	}

	if recAddr != realAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signature.")
	}

	requestID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}
	err = mtr.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	apnft.UnpairRequestID = null.StringFrom(requestID)
	_, err = apnft.Update(c.Context(), tx, boil.Whitelist(models.AftermarketDeviceColumns.UnpairRequestID, models.AftermarketDeviceColumns.UpdatedAt))
	if err != nil {
		return err
	}

	// This is a bit iffy, since we don't want to save this record and then fail to send to Kafka.
	// But the opposite is worse, I think.
	err = tx.Commit()
	if err != nil {
		return err
	}

	return client.UnPairAftermarketDeviceSign(requestID, apToken, vehicleToken, sigBytes)
}

// GetAutoPiUnpairMessage godoc
// @Description Return the EIP-712 payload to be signed for Aftermarket device unpairing.
// @Produce json
// @Param userDeviceID path string true "Device id"
// @Success 200 {object} signer.TypedData
// @Security BearerAuth
// @Router /user/devices/:userDeviceID/aftermarket/commands/unpair [get]
func (udc *UserDevicesController) GetAutoPiUnpairMessage(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)

	userDeviceID := c.Params("userDeviceID")
	logger := helpers.GetLogger(c, udc.log)
	logger.Info().Msg("Got Aftermarket unpair request.")

	vnft, autoPiUnit, err := udc.checkUnpairable(c.Context(), udc.DBS().Writer, userDeviceID)
	if err != nil {
		return err
	}

	apToken := autoPiUnit.TokenID.Int(nil)
	vehicleToken := vnft.TokenID.Int(nil)

	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Failed to retrieve user information.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusConflict, "User does not have an Ethereum address.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	uads := &registry.UnPairAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		VehicleNode:           vehicleToken,
	}

	var out *signer.TypedData = client.GetPayload(uads)

	return c.JSON(out)
}

type AutoPiClaimRequest struct {
	// UserSignature is the signature from the user, using their private key.
	UserSignature string `json:"userSignature"`
	// AftermarketDeviceSignature is the signature from the aftermarket device.
	AftermarketDeviceSignature string `json:"aftermarketDeviceSignature"`
}

type AutoPiPairRequest struct {
	ExternalID string        `json:"externalId"`
	Signature  hexutil.Bytes `json:"signature"`
	// AftermarketDeviceSignature is the 65-byte, hex-encoded Ethereum signature of
	// the pairing payload by the device. Only needed if the vehicle owner and aftermarket
	// device owner are not the same.
	AftermarketDeviceSignature hexutil.Bytes `json:"aftermarketDeviceSignature"`
}

// PostUnclaimAutoPi godoc
// @Description Dev-only endpoint for removing a claim. Removes the flag on-chain and clears
// @Description the owner in the database.
// @Produce json
// @Param serial path string true "AutoPi unit id"
// @Success 204
// @Security BearerAuth
// @Router /aftermarket/device/by-serial/:serial/commands/unclaim [post]
func (udc *UserDevicesController) PostUnclaimAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	unitID := c.Params("serial")

	logger := udc.log.With().Str("userId", userID).Str("serial", unitID).Str("route", c.Route().Name).Logger()

	logger.Info().Msg("Got unclaim request.")

	unit, err := models.AftermarketDevices(models.AftermarketDeviceWhere.Serial.EQ(unitID)).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "Aftermarket device not minted.")
		}
		return err
	}

	apToken := unit.TokenID.Int(nil)

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	requestID := ksuid.New().String()

	unit.OwnerAddress = null.Bytes{}
	unit.UserID = null.String{}
	unit.ClaimMetaTransactionRequestID = null.String{}
	unit.UnpairRequestID = null.String{}
	if _, err := unit.Update(c.Context(), udc.DBS().Writer, boil.Infer()); err != nil {
		return err
	}

	return client.UnclaimAftermarketDeviceNode(requestID, []*big.Int{apToken})
}

// PostClaimAutoPi godoc
// @Description Return the EIP-712 payload to be signed for Aftermarket device claiming.
// @Produce json
// @Param serial path string true "AutoPi unit id"
// @Param claimRequest body controllers.AutoPiClaimRequest true "Signatures from the user and AutoPi"
// @Success 204
// @Security BearerAuth
// @Router /aftermarket/device/by-serial/:serial/commands/claim [post]
func (udc *UserDevicesController) PostClaimAutoPi(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	unitID := c.Params("serial")

	logger := udc.log.With().Str("userId", userID).Str("serial", unitID).Str("route", c.Route().Name).Logger()

	reqBody := AutoPiClaimRequest{}
	err := c.BodyParser(&reqBody)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	udc.log.Info().Interface("payload", reqBody).Msg("Got claim request.")

	unit, err := models.AftermarketDevices(
		models.AftermarketDeviceWhere.Serial.EQ(unitID),
		qm.Load(models.AftermarketDeviceRels.ClaimMetaTransactionRequest),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Aftermarket device not minted.")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Internal error.")
	}

	if unit.OwnerAddress.Valid {
		return fiber.NewError(fiber.StatusConflict, "Device already claimed.")
	}

	if unit.R.ClaimMetaTransactionRequest != nil && unit.R.ClaimMetaTransactionRequest.Status != "Failed" {
		return fiber.NewError(fiber.StatusConflict, "Claiming transaction in progress.")
	}

	apToken := unit.TokenID.Int(nil)

	user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: userID})
	if err != nil {
		udc.log.Err(err).Msg("Couldn't retrieve user record.")
		return opaqueInternalError
	}

	if user.EthereumAddress == nil {
		return fiber.NewError(fiber.StatusBadRequest, "User does not have an Ethereum address.")
	}

	client := registry.Client{
		Producer:     udc.producer,
		RequestTopic: "topic.transaction.request.send",
		Contract: registry.Contract{
			ChainID: big.NewInt(udc.Settings.DIMORegistryChainID),
			Address: common.HexToAddress(udc.Settings.DIMORegistryAddr),
			Name:    "DIMO",
			Version: "1",
		},
	}

	realUserAddr := common.HexToAddress(*user.EthereumAddress)

	cads := &registry.ClaimAftermarketDeviceSign{
		AftermarketDeviceNode: apToken,
		Owner:                 realUserAddr,
	}

	hash, err := client.Hash(cads)
	if err != nil {
		return err
	}

	userSig := common.FromHex(reqBody.UserSignature)

	if len(userSig) != 65 {
		logger.Error().Str("rawSignature", reqBody.UserSignature).Msg("User signature was not 65 bytes.")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("User signature has invalid length %d.", len(userSig)))
	}

	recUserAddr, err := helpers.Ecrecover(hash, userSig)
	if err != nil {
		return err
	}

	if recUserAddr != realUserAddr {
		return fiber.NewError(fiber.StatusBadRequest, "User signature invalid.")
	}

	amSig := common.FromHex(reqBody.AftermarketDeviceSignature)

	if len(amSig) != 65 {
		logger.Error().Str("rawSignature", reqBody.AftermarketDeviceSignature).Msg("Device signature was not 65 bytes.")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Device signature has invalid length %d.", len(amSig)))
	}

	recAmAddr, err := helpers.Ecrecover(hash, amSig)
	if err != nil {
		return err
	}

	realAmAddr := common.BytesToAddress(unit.EthereumAddress)

	if recAmAddr != realAmAddr {
		return fiber.NewError(fiber.StatusBadRequest, "Aftermarket device signature invalid.")
	}

	requestID := ksuid.New().String()

	mtr := models.MetaTransactionRequest{
		ID:     requestID,
		Status: models.MetaTransactionRequestStatusUnsubmitted,
	}
	err = mtr.Insert(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	unit.UserID = null.StringFrom(userID)
	unit.ClaimMetaTransactionRequestID = null.StringFrom(requestID)
	_, err = unit.Update(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return client.ClaimAftermarketDeviceSign(requestID, apToken, realUserAddr, userSig, amSig)
}

func (udc *UserDevicesController) registerDeviceIntegrationInner(c *fiber.Ctx, userID, userDeviceID, integrationID string) error {
	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("integrationId", integrationID).
		Str("route", c.Route().Path).
		Logger()
	logger.Info().Msg("Attempting to register device integration")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to create transaction: %s", err))
	}
	defer tx.Rollback() //nolint
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(userDeviceID),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("could not find device with id %s for user %s", userDeviceID, userID))
		}
		logger.Err(err).Msg("Unexpected database error searching for user device")
		return err
	}

	if !ud.CountryCode.Valid {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("device %s does not have a country code, can't check compatibility", ud.ID))
	}

	countryRecord := constants.FindCountry(ud.CountryCode.String)
	if countryRecord == nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("can't find compatibility region for country %s", ud.CountryCode.String))
	}
	logger = logger.With().Str("region", countryRecord.Region).Logger()

	dd, err := udc.DeviceDefSvc.GetDeviceDefinitionByID(c.Context(), ud.DeviceDefinitionID)
	if err != nil {
		logger.Err(err).Msg("grpc error searching for device definition")
		return shared.GrpcErrorToFiber(err, "failed to get device definition with id: "+ud.DeviceDefinitionID)
	}
	// block SC kia's
	if dd.Make.NameSlug == "kia" && integrationID == "22N2xaPOq2WW2gAHBHd0Ikn4Zob" {
		logger.Warn().Msg("kia blocked, smartcar connection")
		return fiber.NewError(fiber.StatusConflict, "kia software connection blocked, only hardware connections are allowed for this car")
	}
	logger.Info().Msgf("get device definition id result during registration %+v", dd)

	// filter out the desired integration from the compatible ones
	var deviceInteg *ddgrpc.Integration
	for _, integration := range dd.DeviceIntegrations {
		if integration.Integration.Id == integrationID {
			deviceInteg = &ddgrpc.Integration{
				Id:     integration.Integration.Id,
				Type:   integration.Integration.Type,
				Style:  integration.Integration.Style,
				Vendor: integration.Integration.Vendor,
			}
			break
		}
	}

	if deviceInteg == nil {
		// todo need a test for this
		return fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("deviceDefinitionId %s does not support integrationId %s for region %s", ud.DeviceDefinitionID, integrationID, countryRecord.Region))
	}

	if exists, err := models.UserDeviceAPIIntegrationExists(c.Context(), tx, userDeviceID, integrationID); err != nil {
		logger.Err(err).Msg("Unexpected database error looking for existing instance of integration")
		return err
	} else if exists {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("userDeviceId %s already has a user_device_api_integration with integrationId %s, please delete that first", userDeviceID, integrationID))
	}

	var regErr error
	// The per-integration handler is responsible for handling the fiber context and committing the
	// transaction.
	switch vendor := deviceInteg.Vendor; vendor {
	case constants.SmartCarVendor:
		regErr = udc.registerSmartcarIntegration(c, &logger, tx, deviceInteg, ud)
	case constants.TeslaVendor:
		regErr = udc.registerDeviceTesla(c, &logger, tx, userDeviceID, deviceInteg, ud)
	case constants.AutoPiVendor:
		logger.Error().Msg("autopi register request via invalid route: /user/devices/:userDeviceID/integrations/:integrationID")
		return errors.New("this route cannot be used to register an autopi anymore - update your client")
	default:
		logger.Error().Str("vendor", vendor).Msg("Attempted to register an unsupported integration")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("unsupported integration %s", integrationID))
	}

	if regErr != nil {
		return regErr
	}

	udc.runPostRegistration(c.Context(), &logger, userDeviceID, integrationID, deviceInteg)

	return nil
}

// RegisterDeviceIntegration godoc
// @Description Submit credentials for registering a device with a given integration.
// @Tags        integrations
// @Accept      json
// @Param       userDeviceIntegrationRegistration body controllers.RegisterDeviceIntegrationRequest true "Integration credentials"
// @Success     204
// @Security    BearerAuth
// @Router      /user/devices/:userDeviceID/integrations/:integrationID [post]
func (udc *UserDevicesController) RegisterDeviceIntegration(c *fiber.Ctx) error {
	userID := helpers.GetUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	return udc.registerDeviceIntegrationInner(c, userID, userDeviceID, integrationID)
}

/** Refactored / helper methods **/

// runPostRegistration runs tasks that should be run after a successful integration. For now, this
// just means emitting an event to topic.event for the activity log.
func (udc *UserDevicesController) runPostRegistration(ctx context.Context, logger *zerolog.Logger, userDeviceID, integrationID string, integ *ddgrpc.Integration) {
	// Just reload the entire tree of attributes. Too many things modify this during the registration flow.
	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(qm.Rels(models.UserDeviceAPIIntegrationRels.UserDevice)),
	).One(ctx, udc.DBS().Reader)
	if err != nil {
		logger.Err(err).Msg("Couldn't retrieve UDAI for post-registration tasks.")
		return
	}

	ud := udai.R.UserDevice
	// pull dd info again - don't pass it in, as it may have changed
	dd, err2 := udc.DeviceDefSvc.GetDeviceDefinitionByID(ctx, ud.DeviceDefinitionID)
	if err2 != nil {
		logger.Err(err2).Str("deviceDefinitionId", ud.DeviceDefinitionID).Msg("failed to retrieve device defintion")
	}

	err = udc.eventService.Emit(
		&shared.CloudEvent[any]{
			Type:    "com.dimo.zone.device.integration.create",
			Source:  "devices-api",
			Subject: userDeviceID,
			Data: services.UserDeviceIntegrationEvent{
				Timestamp: time.Now(),
				UserID:    ud.UserID,
				Device: services.UserDeviceEventDevice{
					ID:                 userDeviceID,
					DeviceDefinitionID: dd.DeviceDefinitionId,
					Make:               dd.Type.Make,
					Model:              dd.Type.Model,
					Year:               int(dd.Type.Year),
					VIN:                ud.VinIdentifier.String,
				},
				Integration: services.UserDeviceEventIntegration{
					ID:     integ.Id,
					Type:   integ.Type,
					Style:  integ.Style,
					Vendor: integ.Vendor,
				},
			},
		},
	)
	if err != nil {
		logger.Err(err).Msg("Failed to emit integration event.")
	}

	region := ""
	if ud.CountryCode.Valid {
		countryRecord := constants.FindCountry(ud.CountryCode.String)
		if countryRecord != nil {
			region = countryRecord.Region
		}
	}
	err = udc.deviceDefinitionRegistrar.Register(services.DeviceDefinitionDTO{
		IntegrationID:      integ.Id,
		UserDeviceID:       ud.ID,
		DeviceDefinitionID: ud.DeviceDefinitionID,
		Make:               dd.Type.Make,
		Model:              dd.Type.Model,
		Year:               int(dd.Type.Year),
		Region:             region,
		MakeSlug:           dd.Type.MakeSlug,
		ModelSlug:          dd.Type.ModelSlug,
	})
	if err != nil {
		logger.Err(err).Msg("Failed to set values in device definition tables.")
	}
}

var smartcarCallErr = fiber.NewError(fiber.StatusInternalServerError, "Error communicating with Smartcar.")

func (udc *UserDevicesController) registerSmartcarIntegration(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, integ *ddgrpc.Integration, ud *models.UserDevice) error {
	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request JSON body.")
	}
	var token *smartcar.Token
	// check for token in redis, if exists do not call this.
	if ud.VinIdentifier.Valid {
		scTokenGet, err := udc.redisCache.Get(c.Context(), buildSmartcarTokenKey(ud.VinIdentifier.String, ud.UserID)).Result()
		if err == nil && len(scTokenGet) > 0 {
			decrypted, err := udc.cipher.Decrypt(scTokenGet)
			if err != nil {
				return errors.Wrap(err, "failed to decrypt sc token")
			}
			// found existing token
			token = &smartcar.Token{}
			err = json.Unmarshal([]byte(decrypted), token)
			if err != nil {
				udc.log.Err(err).Msgf("failed to unmarshal smartcar token found in redis cache for vin: %s", ud.VinIdentifier.String)
			}
			// clear cache
			udc.redisCache.Del(c.Context(), buildSmartcarTokenKey(ud.VinIdentifier.String, ud.UserID))
		}
	}
	if token == nil {
		// no token found or could be unmarshalled so try exchangecode, assumption is it has not been called before for this code
		var err error
		token, err = udc.smartcarClient.ExchangeCode(c.Context(), reqBody.Code, reqBody.RedirectURI)
		if err != nil {
			var scErr *services.SmartcarError
			if errors.As(err, &scErr) {
				logger.Error().Msgf("Failed exchanging Authorization code. Status code %d, request id %s, and body `%s`.", scErr.Code, scErr.RequestID, string(scErr.Body))
			} else {
				logger.Err(err).Msg("Failed to exchange authorization code with Smartcar.")
			}

			// This may not be the user's fault, but 400 for now.
			return fiber.NewError(fiber.StatusBadRequest, "Failed to exchange authorization code with Smartcar.")
		}
	}

	scUserID, err := udc.smartcarClient.GetUserID(c.Context(), token.Access)
	if err != nil {
		logger.Err(err).Msg("Failed to retrieve user ID from Smartcar.")
		return smartcarCallErr
	}

	externalID, err := udc.smartcarClient.GetExternalID(c.Context(), token.Access)
	if err != nil {
		logger.Err(err).Msg("Failed to retrieve vehicle ID from Smartcar.")
		return smartcarCallErr
	}
	// by default use vin from userdevice, unless if it is not confirmed, in that case pull from SC
	vin := ud.VinIdentifier.String
	if !ud.VinConfirmed {
		vin, err = udc.smartcarClient.GetVIN(c.Context(), token.Access, externalID)
		if err != nil {
			logger.Err(err).Msg("Failed to retrieve VIN from Smartcar.")
			return smartcarCallErr
		}

		if ud.VinConfirmed && ud.VinIdentifier.String != vin {
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Vehicle's confirmed VIN does not match Smartcar's %s.", vin))
		}
	}
	localLog := logger.With().Str("vin", vin).Str("userId", ud.UserID).Logger()

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.IsProduction() {
		if vin[0:3] == "0SC" {
			localLog.Error().Msgf("Smartcar test VIN %s is not allowed.", vin)
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("Smartcar test VIN %s is not allowed.", vin))
		}
		// Probably a race condition here. Need to either lock something or impose a greater
		// isolation level.
		conflict, err := models.UserDevices(
			models.UserDeviceWhere.ID.NEQ(ud.ID), // If you want to re-register, or register a different integration, that's okay.
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(c.Context(), tx)
		if err != nil {
			localLog.Err(err).Msg("Failed to search for VIN conflicts.")
			return opaqueInternalError
		}

		if conflict {
			localLog.Error().Msg("VIN %s already in use.")
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("VIN %s in use by a previously connected device.", ud.VinIdentifier.String))
		}
	}

	endpoints, err := udc.smartcarClient.GetEndpoints(c.Context(), token.Access, externalID)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve permissions from Smartcar.")
		return smartcarCallErr
	}

	var commands *services.UserDeviceAPIIntegrationsMetadataCommands

	doorControl, err := udc.smartcarClient.HasDoorControl(c.Context(), token.Access, externalID)
	if err != nil {
		localLog.Err(err).Msg("Failed to retrieve door control permissions from Smartcar.")
		return smartcarCallErr
	}

	if doorControl {
		commands = &services.UserDeviceAPIIntegrationsMetadataCommands{
			Enabled: []string{"doors/unlock", "doors/lock"},
		}
	}

	meta := services.UserDeviceAPIIntegrationsMetadata{
		SmartcarUserID:    &scUserID,
		SmartcarEndpoints: endpoints,
		Commands:          commands,
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	encAccess, err := udc.cipher.Encrypt(token.Access)
	if err != nil {
		return opaqueInternalError
	}

	encRefresh, err := udc.cipher.Encrypt(token.Refresh)
	if err != nil {
		return opaqueInternalError
	}

	taskID := ksuid.New().String()

	integration := &models.UserDeviceAPIIntegration{
		TaskID:          null.StringFrom(taskID),
		ExternalID:      null.StringFrom(externalID),
		UserDeviceID:    ud.ID,
		IntegrationID:   integ.Id,
		Status:          models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:     null.StringFrom(encAccess),
		AccessExpiresAt: null.TimeFrom(token.AccessExpiry),
		RefreshToken:    null.StringFrom(encRefresh),
		Metadata:        null.JSONFrom(b),
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		localLog.Err(err).Msg("Unexpected database error inserting new Smartcar integration registration.")
		return opaqueInternalError
	}

	if !ud.VinConfirmed {
		ud.VinIdentifier = null.StringFrom(strings.ToUpper(vin))
		ud.VinConfirmed = true
		_, err = ud.Update(c.Context(), tx, boil.Infer())
		if err != nil {
			return opaqueInternalError
		}
	}

	if err := udc.smartcarTaskSvc.StartPoll(integration); err != nil {
		localLog.Err(err).Msg("Couldn't start Smartcar polling.")
		return opaqueInternalError
	}

	if err := tx.Commit(); err != nil {
		localLog.Error().Msg("Failed to commit new user device integration.")
		return opaqueInternalError
	}

	localLog.Info().Msg("Finished Smartcar device registration.")

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) registerDeviceTesla(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, userDeviceID string, integ *ddgrpc.Integration, ud *models.UserDevice) error {
	if existingIntegrations, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
	).Count(c.Context(), tx); err != nil {
		return err
	} else if existingIntegrations > 0 {
		return fiber.NewError(fiber.StatusConflict, "Delete existing integration before connecting through Tesla.")
	}

	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
	}

	// Flag for which api version should be used
	apiVersion := constants.TeslaAPIV1
	if reqBody.Version != 0 {
		apiVersion = reqBody.Version
	}

	if reqBody.ExternalID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing externalID parameter")
	}

	v := &services.TeslaVehicle{}
	// We'll use this to kick off the job
	teslaID, err := strconv.Atoi(reqBody.ExternalID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Couldn't parse external id %q as an integer.", teslaID))
	}

	region := ""
	teslaV2CacheKey := ""
	if apiVersion == constants.TeslaAPIV2 { // If version is 2, we are using fleet api which has token stored in cache
		user, err := udc.usersClient.GetUser(c.Context(), &pb.GetUserRequest{Id: ud.UserID})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not fetch user information: %w", err.Error())
		}
		if user.EthereumAddress == nil {
			return fiber.NewError(fiber.StatusBadRequest, "missing wallet details for user")
		}
		teslaV2CacheKey = fmt.Sprintf(teslaFleetAuthCacheKey, *user.EthereumAddress)

		deviceIntReq, err := udc.getTeslaAuthFromCache(c.Context(), teslaV2CacheKey)
		if err != nil {
			udc.log.Err(err).Msg("Error occurred retrieving tesla auth from cache")
			return fiber.NewError(fiber.StatusBadRequest, "Couldn't retrieve stored credentials: "+err.Error())
		}
		reqBody.RefreshToken = deviceIntReq.RefreshToken
		reqBody.AccessToken = deviceIntReq.AccessToken
		reqBody.ExpiresIn = int(time.Until(deviceIntReq.Expiry).Seconds())
		region = deviceIntReq.Region
	}

	v, err = udc.getTeslaVehicle(c.Context(), reqBody.AccessToken, region, teslaID, apiVersion)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't retrieve vehicle from Tesla.")
	}

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.IsProduction() {
		// Probably a race condition here.
		var conflict bool
		conflict, err = models.UserDevices(
			models.UserDeviceWhere.ID.NEQ(userDeviceID), // If you want to re-register, that's okay.
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(v.VIN)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(c.Context(), tx)

		if err != nil {
			return err
		}

		if conflict {
			return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("VIN %s in use by another vehicle.", v.VIN))
		}
	}

	if err := fixTeslaDeviceDefinition(c.Context(), logger, udc.DeviceDefSvc, tx, integ, ud, v.VIN); err != nil {
		return fmt.Errorf("correcting device definition: %w", err)
	}

	encAccessToken, err := udc.cipher.Encrypt(reqBody.AccessToken)
	if err != nil {
		return err
	}

	encRefreshToken, err := udc.cipher.Encrypt(reqBody.RefreshToken)
	if err != nil {
		return err
	}

	// TODO(elffjs): Stupid to marshal this again and again.
	meta := services.UserDeviceAPIIntegrationsMetadata{
		Commands: &services.UserDeviceAPIIntegrationsMetadataCommands{
			Enabled: []string{"doors/unlock", "doors/lock", "trunk/open", "frunk/open", "charge/limit"},
		},
		TeslaAPIVersion: apiVersion,
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	taskID := ksuid.New().String()

	integration := models.UserDeviceAPIIntegration{
		UserDeviceID:    userDeviceID,
		IntegrationID:   integ.Id,
		ExternalID:      null.StringFrom(reqBody.ExternalID),
		Status:          models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:     null.StringFrom(encAccessToken),
		AccessExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(reqBody.ExpiresIn) * time.Second)),
		RefreshToken:    null.StringFrom(encRefreshToken), // Don't know when this expires.
		TaskID:          null.StringFrom(taskID),
		Metadata:        null.JSONFrom(b),
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		return err
	}

	ud.VinIdentifier = null.StringFrom(strings.ToUpper(v.VIN))
	ud.VinConfirmed = true
	_, err = ud.Update(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	if err := udc.wakeupTeslaVehicle(c.Context(), reqBody.AccessToken, region, teslaID, apiVersion); err != nil {
		logger.Err(err).Msg("Couldn't wake up Tesla.")
	}

	if udc.Settings.IsProduction() {
		tokenID := int64(0)
		if ud.R != nil && ud.R.VehicleNFT != nil {
			tokenID, _ = ud.R.VehicleNFT.TokenID.Int64()
		}
		udc.requestValuation(v.VIN, userDeviceID, tokenID)
		udc.requestInstantOffer(userDeviceID, tokenID)
	}
	if err := udc.teslaTaskService.StartPoll(v, &integration, apiVersion, region); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info().Msg("Finished Tesla device registration")

	if apiVersion == constants.TeslaAPIV2 && teslaV2CacheKey != "" {
		err = udc.redisCache.Del(c.Context(), teslaV2CacheKey).Err()
		if err != nil {
			udc.log.Err(err).Str("cacheKey", teslaV2CacheKey).Msg("error occurred deleting record from cache")
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) wakeupTeslaVehicle(ctx context.Context, token, region string, vehicleID, version int) error {
	var err error
	if version == constants.TeslaAPIV2 {
		err = udc.teslaFleetAPISvc.WakeUpVehicle(ctx, token, region, vehicleID)
	} else {
		err = udc.teslaService.WakeUpVehicle(token, vehicleID)
	}
	return err
}

func (udc *UserDevicesController) getTeslaVehicle(ctx context.Context, token, region string, vehicleID, version int) (*services.TeslaVehicle, error) {
	var vehicle *services.TeslaVehicle
	var err error
	if version == constants.TeslaAPIV2 {
		vehicle, err = udc.teslaFleetAPISvc.GetVehicle(ctx, token, region, vehicleID)
	} else {
		vehicle, err = udc.teslaService.GetVehicle(token, vehicleID)
	}

	return vehicle, err
}

func (udc *UserDevicesController) getTeslaAuthFromCache(ctx context.Context, cacheKey string) (*services.TeslaAuthCodeResponse, error) {
	encTeslaAuth, err := udc.redisCache.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("tesla authorization token has expired")
		}
		return nil, fmt.Errorf("could not retrieve Tesla credentials: %w", err)
	}
	if len(encTeslaAuth) == 0 {
		return nil, fmt.Errorf("no credential found")
	}
	decrypted, err := udc.cipher.Decrypt(encTeslaAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tesla token: %w", err)
	}

	teslaAuth := &services.TeslaAuthCodeResponse{}
	err = json.Unmarshal([]byte(decrypted), &teslaAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tesla authorization token: %w", err)
	}

	if teslaAuth.AccessToken == "" || teslaAuth.RefreshToken == "" || teslaAuth.Expiry.IsZero() {
		return nil, fmt.Errorf("missing tesla auth credentials")
	}

	return teslaAuth, nil
}

// fixTeslaDeviceDefinition tries to use the VIN provided by Tesla to correct the device definition
// used by a device.
//
// We do not attempt to create any new entries in integrations, device_definitions, or
// device_integrations. This should all be handled elsewhere for Tesla.
func fixTeslaDeviceDefinition(ctx context.Context, logger *zerolog.Logger, ddSvc services.DeviceDefinitionService, exec boil.ContextExecutor, _ *ddgrpc.Integration, ud *models.UserDevice, vin string) error {
	vinMake := "Tesla"
	vinModel := shared.VIN(vin).TeslaModel()
	vinYear := shared.VIN(vin).Year()
	mmy, err := ddSvc.FindDeviceDefinitionByMMY(ctx, vinMake, vinModel, vinYear)

	if err != nil {
		return err
	}

	if mmy.DeviceDefinitionId != ud.DeviceDefinitionID {
		logger.Warn().Msgf(
			"Device moving to new device definition from %s to %s", ud.DeviceDefinitionID, mmy.DeviceDefinitionId,
		)
		ud.DeviceDefinitionID = mmy.DeviceDefinitionId
		_, err = ud.Update(ctx, exec, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}

/** Structs for request / response **/

type UserDeviceIntegrationStatus struct {
	IntegrationID     string                 `json:"integrationId"`
	Status            string                 `json:"status"`
	CreatedAt         time.Time              `json:"createdAt"`
	ExternalID        *string                `json:"externalId"`
	UpdatedAt         time.Time              `json:"updatedAt"`
	Metadata          null.JSON              `json:"metadata" swaggertype:"string"`
	IntegrationVendor string                 `json:"integrationVendor"`
	Mint              *SyntheticDeviceStatus `json:"syntheticDevice,omitempty"`
	TokenID           *big.Int               `json:"tokenId,omitempty"`
}

// RegisterDeviceIntegrationRequest carries credentials used to connect the device to a given
// integration.
type RegisterDeviceIntegrationRequest struct {
	// Code is an OAuth authorization code. Not used in all integrations.
	Code string `json:"code"`
	// RedirectURI is the OAuth redirect URI used by the frontend. Not used in all integrations.
	RedirectURI string `json:"redirectURI"`
	// ExternalID is the only field needed for AutoPi registrations. It is the UnitID.
	ExternalID   string `json:"externalId"`
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int    `json:"expiresIn"`
	RefreshToken string `json:"refreshToken"`
	Version      int    `json:"version"`
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

type AutoPiCommandRequest struct {
	Command string `json:"command"`
}

type ManufacturerInfo struct {
	TokenID *big.Int `json:"tokenId"`
	Name    string   `json:"name"`
}

// AutoPiDeviceInfo is used to get the info about a unit
type AutoPiDeviceInfo struct {
	IsUpdated         bool      `json:"isUpdated"`
	DeviceID          string    `json:"deviceId"`
	UnitID            string    `json:"unitId"`
	DockerReleases    []int     `json:"dockerReleases"`
	HwRevision        string    `json:"hwRevision"`
	Template          int       `json:"template"`
	LastCommunication time.Time `json:"lastCommunication"`
	ReleaseVersion    string    `json:"releaseVersion"`
	ShouldUpdate      bool      `json:"shouldUpdate"`

	TokenID            *big.Int        `json:"tokenId,omitempty"`
	EthereumAddress    *common.Address `json:"ethereumAddress,omitempty"`
	OwnerAddress       *string         `json:"ownerAddress,omitempty"`
	BeneficiaryAddress *common.Address `json:"beneficiaryAddress,omitempty"`

	// Claim contains the status of the on-chain claiming meta-transaction.
	Claim *AutoPiTransactionStatus `json:"claim,omitempty"`
	// Pair contains the status of the on-chain pairing meta-transaction.
	Pair *AutoPiTransactionStatus `json:"pair,omitempty"`
	// Unpair contains the status of the on-chain unpairing meta-transaction.
	Unpair *AutoPiTransactionStatus `json:"unpair,omitempty"`

	Manufacturer *ManufacturerInfo `json:"manufacturer,omitempty"`
}

// AutoPiTransactionStatus summarizes the state of an on-chain AutoPi operation.
type AutoPiTransactionStatus struct {
	// Status is the state of the transaction performing this operation. There are only four options.
	Status string `json:"status" enums:"Unsubmitted,Submitted,Mined,Confirmed" example:"Mined"`
	// Hash is the hexidecimal transaction hash, available for any transaction at the Submitted stage or greater.
	Hash *string `json:"hash,omitempty" example:"0x28b4662f1e1b15083261a4a5077664f4003d58cb528826b7aab7fad466c28e70"`
	// CreatedAt is the timestamp of the creation of the meta-transaction.
	CreatedAt time.Time `json:"createdAt" example:"2022-10-01T09:22:21.002Z"`
	// UpdatedAt is the last time we updated the status of the transaction.
	UpdatedAt time.Time `json:"updatedAt" example:"2022-10-01T09:22:26.337Z"`
}
