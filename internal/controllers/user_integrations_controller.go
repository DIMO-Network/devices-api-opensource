package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GetUserDeviceIntegration godoc
// @Description  Receive status updates about a Smartcar integration
// @Tags         integrations
// @Success      200  {object}  controllers.GetUserDeviceIntegrationResponse
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/integrations/{integrationID} [get]
func (udc *UserDevicesController) GetUserDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")
	deviceExists, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
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

// DeleteUserDeviceIntegration godoc
// @Description  Remove an user device's integration
// @Tags         integrations
// @Success      204
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/integrations/{integrationID} [delete]
func (udc *UserDevicesController) DeleteUserDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

	tx, err := udc.DBS().Writer.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint

	device, err := models.UserDevices(
		models.UserDeviceWhere.UserID.EQ(userID),
		models.UserDeviceWhere.ID.EQ(userDeviceID),
		qm.Load(models.UserDeviceRels.DeviceDefinition),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no user device with ID %s", userDeviceID))
		}
		return err
	}

	// Probably don't need two queries if you're smart
	apiIntegration, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
		qm.Load(models.UserDeviceAPIIntegrationRels.Integration),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("user device %s does not have integration %s", userDeviceID, integrationID))
		}
		return err
	}

	if apiIntegration.R.Integration.Vendor == services.SmartCarVendor {
		if apiIntegration.ExternalID.Valid {
			if apiIntegration.TaskID.Valid {
				err = udc.smartcarTaskSvc.StopPoll(apiIntegration)
				if err != nil {
					return err
				}
			} else {
				err = udc.taskSvc.StartSmartcarDeregistrationTasks(userDeviceID, integrationID, apiIntegration.ExternalID.String, apiIntegration.AccessToken.String)
				if err != nil {
					return err
				}
			}
		}
	} else if apiIntegration.R.Integration.Vendor == "Tesla" {
		if apiIntegration.ExternalID.Valid {
			if err := udc.teslaTaskService.StopPoll(apiIntegration); err != nil {
				return err
			}
		}
	} else if apiIntegration.R.Integration.Vendor == services.AutoPiVendor {
		err = udc.autoPiIngestRegistrar.Deregister(apiIntegration.ExternalID.String, apiIntegration.UserDeviceID, apiIntegration.IntegrationID)
		if err != nil {
			udc.log.Err(err).Msgf("unexpected error deregistering autopi device from ingest. userDeviceID: %s", apiIntegration.UserDeviceID)
			return err
		}
	} else {
		udc.log.Warn().Msgf("Don't know how to deregister integration %s for device %s", apiIntegration.IntegrationID, userDeviceID)
	}

	_, err = apiIntegration.Delete(c.Context(), tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.delete",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    userID,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  device.R.DeviceDefinition.R.DeviceMake.Name,
				Model: device.R.DeviceDefinition.Model,
				Year:  int(device.R.DeviceDefinition.Year),
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     apiIntegration.R.Integration.ID,
				Type:   apiIntegration.R.Integration.Type,
				Style:  apiIntegration.R.Integration.Style,
				Vendor: apiIntegration.R.Integration.Vendor,
			},
		},
	})
	if err != nil {
		udc.log.Err(err).Msg("Failed to emit integration deletion")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetIntegrations godoc
// @Description  gets list of integrations we have defined
// @Tags         integrations
// @Produce      json
// @Success      200  {object}  models.Integration
// @Security     BearerAuth
// @Router       /integrations [get]
func (udc *UserDevicesController) GetIntegrations(c *fiber.Ctx) error {
	all, err := models.Integrations(qm.Limit(100)).All(c.Context(), udc.DBS().Reader)
	if err != nil {
		return errors.Wrap(err, "failed to get integrations")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"integrations": all,
	})
}

// SendAutoPiCommand godoc
// @Description Submit a raw autopi command to unit. Device must be registered with autopi before this can be used
// @Tags 		integrations
// @Accept 		json
// @Param 		AutoPiCommandRequest body controllers.AutoPiCommandRequest true "raw autopi command"
// @Success 	200
// @Security     BearerAuth
// @Router 		/user/devices/:userDeviceID/autopi/command [post]
func (udc *UserDevicesController) SendAutoPiCommand(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	req := new(AutoPiCommandRequest)
	err := c.BodyParser(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unable to parse body json")
	}

	logger := udc.log.With().
		Str("userId", userID).
		Str("userDeviceId", userDeviceID).
		Str("handler", "SendAutoPiCommand").
		Str("autopi cmd", req.Command).
		Logger()
	logger.Info().Msg("Attempting to send autopi raw command")

	udai, md, err := services.FindUserDeviceAutoPiIntegration(c.Context(), udc.DBS().Writer, userDeviceID, userID)
	if err != nil {
		logger.Err(err).Msg("error finding user device autopi integration")
		return err
	}
	// call autopi
	commandResponse, err := udc.autoPiSvc.CommandRaw(udai.ExternalID.String, req.Command, true)
	if err != nil {
		logger.Err(err).Msg("autopi returned error when calling raw command")
		return errors.Wrapf(err, "autopi returned error when calling raw command: %s", req.Command)
	}
	// add job to existing jobs
	md.AutoPiCommandJobs = append(md.AutoPiCommandJobs, services.UserDeviceAPIIntegrationJob{
		CommandJobID: commandResponse.Jid,
		CommandState: "sent",
		CommandRaw:   req.Command,
		LastUpdated:  time.Now().UTC(),
	})
	// update db
	err = udai.Metadata.Marshal(md)
	if err != nil {
		logger.Err(err).Msg("unable to marshal metadata back into database")
		return err
	}
	_, err = udai.Update(c.Context(), udc.DBS().Writer, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Metadata,
		models.UserDeviceAPIIntegrationColumns.UpdatedAt))
	if err != nil {
		logger.Err(err).Msg("failed to update database")
		return err
	}

	return c.Status(fiber.StatusOK).JSON(commandResponse)
}

// GetAutoPiCommandStatus godoc
// @Description gets the status of an autopi raw command by jobID
// @Tags 		integrations
// @Produce     json
// @Param       jobID        path  string  true  "job id, from autopi"
// @Success     200  {object}  services.UserDeviceAPIIntegrationJob
// @Security     BearerAuth
// @Router 		/user/devices/:userDeviceID/autopi/command/:jobID [get]
func (udc *UserDevicesController) GetAutoPiCommandStatus(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	jobID := c.Params("jobID")

	_, md, err := services.FindUserDeviceAutoPiIntegration(c.Context(), udc.DBS().Writer, userDeviceID, userID)
	if err != nil {
		return err
	}
	if md != nil {
		for _, job := range md.AutoPiCommandJobs {
			if job.CommandJobID == jobID {
				return c.Status(fiber.StatusOK).JSON(job)
			}
		}
	}

	return c.Status(fiber.StatusBadRequest).SendString("no job found with provided jobID")
}

// GetAutoPiUnitInfo godoc
// @Description gets the information about the autopi by the unitId
// @Tags 		integrations
// @Produce     json
// @Param       unitID        path  string  true  "autopi unit id"
// @Success     200
// @Security     BearerAuth
// @Router 		/autopi/unit/:unitID [get]
func (udc *UserDevicesController) GetAutoPiUnitInfo(c *fiber.Ctx) error {
	unitID := c.Params("unitID")
	if len(unitID) == 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	userID := getUserID(c)
	// check if unitId has already been assigned to a different user - don't allow querying in this case
	udai, err := models.UserDeviceAPIIntegrations(qm.Where("metadata ->> 'auto_pi_unit_id' = $1", unitID),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice)).
		One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if udai != nil {
		if udai.R.UserDevice.UserID != userID {
			return c.SendStatus(fiber.StatusForbidden)
		}
	}

	unit, err := udc.autoPiSvc.GetDeviceByUnitID(unitID)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"isUpdated":         unit.IsUpdated,
		"deviceId":          unit.ID,
		"unitId":            unit.UnitID,
		"dockerReleases":    unit.DockerReleases,
		"hwRevision":        unit.HwRevision,
		"template":          unit.Template,
		"lastCommunication": unit.LastCommunication,
		"releaseVersion":    unit.Release.Version,
	})
}

// GetIsAutoPiOnline godoc
// @Description gets whether the autopi is online right now, if already paired with a user, makes sure user has access. returns json with {"online": true/false}
// @Tags 		integrations
// @Produce     json
// @Param       unitID        path  string  true  "autopi unit id"
// @Success     200
// @Security    BearerAuth
// @Router 		/autopi/unit/is-online/:unitID [get]
func (udc *UserDevicesController) GetIsAutoPiOnline(c *fiber.Ctx) error {
	unitID := c.Params("unitID")
	if len(unitID) == 0 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	userID := getUserID(c)
	deviceID := ""
	// check if unitId has already been assigned to a different user - don't allow querying in this case
	udai, err := models.UserDeviceAPIIntegrations(qm.Where("metadata ->> 'auto_pi_unit_id' = $1", unitID),
		qm.Load(models.UserDeviceAPIIntegrationRels.UserDevice)).
		One(c.Context(), udc.DBS().Reader)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if udai != nil {
		if udai.R.UserDevice.UserID != userID {
			return c.SendStatus(fiber.StatusForbidden)
		}
		deviceID = udai.ExternalID.String
	}
	if len(deviceID) == 0 {
		unit, err := udc.autoPiSvc.GetDeviceByUnitID(unitID)
		if err != nil {
			return err
		}
		deviceID = unit.ID
	}
	// send command without webhook since we'll just query the jobid
	commandResponse, err := udc.autoPiSvc.CommandRaw(deviceID, "test.ping", false)
	if err != nil {
		return err
	}
	// for loop with wait timer of 1 second at begining that calls autopi get job id
	backoffSchedule := []time.Duration{
		1 * time.Second,
	}
	online := false
	for _, backoff := range backoffSchedule {
		time.Sleep(backoff)
		body, err := udc.autoPiSvc.GetCommandStatus(deviceID, commandResponse.Jid)
		// if device is offline, we will get err because it will most likely return 504
		if err != nil {
			continue // try again if don't get a response
		}
		if strings.Contains(string(body), "true") {
			online = true
			break
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"online": online,
	})
}

// RegisterDeviceIntegration godoc
// @Description Submit credentials for registering a device with a given integration.
// @Tags 		integrations
// @Accept 		json
// @Param 		userDeviceIntegrationRegistration body controllers.RegisterDeviceIntegrationRequest true "Integration credentials"
// @Success 	204
// @Security     BearerAuth
// @Router 		/user/devices/:userDeviceID/integrations/:integrationID [post]
func (udc *UserDevicesController) RegisterDeviceIntegration(c *fiber.Ctx) error {
	userID := getUserID(c)
	userDeviceID := c.Params("userDeviceID")
	integrationID := c.Params("integrationID")

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
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(qm.Rels(models.UserDeviceRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
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

	countryRecord := services.FindCountry(ud.CountryCode.String)
	if countryRecord == nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("can't find compatibility region for country %s", ud.CountryCode.String))
	}

	deviceInteg, err := models.DeviceIntegrations(
		models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(ud.DeviceDefinitionID),
		models.DeviceIntegrationWhere.IntegrationID.EQ(integrationID),
		models.DeviceIntegrationWhere.Region.EQ(countryRecord.Region),
		qm.Load(models.DeviceIntegrationRels.Integration),
	).One(c.Context(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// if request is for an autopi integration, create device_integration on the fly
			deviceInteg, err = createDeviceIntegrationIfAutoPi(c.Context(), integrationID, ud.DeviceDefinitionID, countryRecord.Region, tx)
			if err != nil {
				logger.Err(err).Msg("failed to create autopi device_integration on the fly.")
				return err
			}
			if deviceInteg == nil {
				logger.Warn().Msg("Attempted to register a device integration that didn't exist")
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("could not find device integration for device definition %s, integration %s and country %s", ud.DeviceDefinitionID, integrationID, ud.CountryCode.String))
			}
		} else {
			logger.Err(err).Msg("Unexpected database error searching for device integration")
			return err
		}
	}

	if exists, err := models.UserDeviceAPIIntegrationExists(c.Context(), tx, userDeviceID, integrationID); err != nil {
		logger.Err(err).Msg("Unexpected database error looking for existing instance of integration")
		return err
	} else if exists {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("device %s already has a registration with integration %s, please delete that first", userDeviceID, integrationID))
	}

	// In anticipation of a bunch more of these. Maybe move to a real internal integration registry.
	// The per-integration handler is responsible for handling the fiber context and committing the
	// transaction.
	switch deviceInteg.R.Integration.Vendor {
	case services.SmartCarVendor:
		return udc.registerSmartcarIntegration(c, &logger, tx, deviceInteg.R.Integration, ud)
	case "Tesla":
		return udc.registerDeviceTesla(c, &logger, tx, userDeviceID, deviceInteg.R.Integration, ud)
	case services.AutoPiVendor:
		return udc.registerAutoPiUnit(c, &logger, tx, ud, deviceInteg.R.Integration)
	default:
		logger.Error().Msg("Attempted to register an unsupported integration")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("unsupported integration %s", integrationID))
	}
}

/** Refactored / helper methods **/

// registerAutoPiUnit adds record to user api integrations table and calls various autoPi API endpoints to set our TemplateID
func (udc *UserDevicesController) registerAutoPiUnit(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, ud *models.UserDevice, integration *models.Integration) error {
	reqBody := new(RegisterDeviceIntegrationRequest) // we only care about the externalId here
	err := c.BodyParser(&reqBody)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unable to parse body json")
	}
	subLogger := logger.With().Str("autopi unitID", reqBody.ExternalID).Logger()

	// check if an existing integration exists for the unitID
	unitExists, err := models.UserDeviceAPIIntegrations(qm.Where("metadata ->> 'auto_pi_unit_id' = $1", reqBody.ExternalID),
		qm.And("status IN ('Pending', 'PendingFirstData', 'Active')")). // could not get sqlboiler typed qm.AndIn to work
		Exists(c.Context(), udc.DBS().Reader)
	if err != nil {
		return err
	}
	if unitExists {
		subLogger.Warn().Str("autoPiUnitID", reqBody.ExternalID).Msg("user tried pairing an already paired unitID")
		return fiber.NewError(fiber.StatusBadRequest, "autopi unitID already paired")
	}

	autoPiDevice, err := udc.autoPiSvc.GetDeviceByUnitID(reqBody.ExternalID)
	if err != nil {
		subLogger.Err(err).Msgf("failed to call autopi api to get autoPiDevice by unit id %s", reqBody.ExternalID)
		return err
	}
	subLogger = subLogger.With().
		Str("autopi deviceID", autoPiDevice.ID).
		Str("original templateID", strconv.Itoa(autoPiDevice.Template)).Logger()

	// validate necessary conditions:
	//- integration metadata contains AutoPiDefaultTemplateID
	im := new(services.IntegrationsMetadata)
	err = integration.Metadata.Unmarshal(&im)
	if err != nil {
		subLogger.Err(err).Msgf("failed to unmarshall integration metadata id %s", integration.ID)
		return err
	}
	if im.AutoPiDefaultTemplateID == 0 {
		return errors.Wrapf(err, "integration id %s does not have autopi default template id", integration.ID)
	}
	templateID := im.AutoPiDefaultTemplateID
	// determine templateID to apply
	if len(im.AutoPiPowertrainToTemplateID) > 0 {
		udMd := new(services.UserDeviceMetadata)
		err = ud.Metadata.Unmarshal(udMd)
		if err != nil {
			subLogger.Err(err).Msgf("failed to unmarshall user_device metadata id %s", ud.ID)
			return err
		}
		if udMd.PowertrainType != nil {
			if id, ok := im.AutoPiPowertrainToTemplateID[*udMd.PowertrainType]; ok {
				templateID = id
			}
		}
	}
	subLogger = subLogger.With().Str("templateID to apply", strconv.Itoa(templateID)).Logger()

	// creat the api int record, start filling it in
	udMetadata := services.UserDeviceAPIIntegrationsMetadata{
		AutoPiUnitID:          &autoPiDevice.UnitID,
		AutoPiIMEI:            &autoPiDevice.IMEI,
		AutoPiTemplateApplied: &templateID,
	}
	apiInt := models.UserDeviceAPIIntegration{
		UserDeviceID:  ud.ID,
		IntegrationID: integration.ID,
		ExternalID:    null.StringFrom(autoPiDevice.ID),
		Status:        models.UserDeviceAPIIntegrationStatusPending,
	}
	err = apiInt.Metadata.Marshal(udMetadata)
	if err != nil {
		return errors.Wrap(err, "failed to marshall user device integration metadata")
	}

	if err = apiInt.Insert(c.Context(), tx, boil.Infer()); err != nil {
		subLogger.Err(err).Msg("database error inserting new autopi integration registration")
		return err
	}
	// update integration record as failed if errors after this
	defer func() {
		if err != nil {
			subLogger.Err(err).Msg("registerAutoPiUnit failure")
			apiInt.Status = models.UserDeviceAPIIntegrationStatusFailed
			msg := err.Error()
			udMetadata.AutoPiRegistrationError = &msg
			_ = apiInt.Metadata.Marshal(udMetadata)
			_, err := apiInt.Update(c.Context(), tx,
				boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Status, models.UserDeviceAPIIntegrationColumns.UpdatedAt))
			if err != nil {
				subLogger.Err(err).Msg("database error updating autopi integration to failed")
			}
			err = tx.Commit()
			if err != nil {
				subLogger.Err(err).Msg("transaction error updating autopi integration to failed")
			}
		}
	}()
	// update the profile on autopi
	profile := services.PatchVehicleProfile{
		Year: int(ud.R.DeviceDefinition.Year),
	}
	if !ud.VinIdentifier.IsZero() {
		profile.Vin = ud.VinIdentifier.String
	}
	if !ud.Name.IsZero() {
		profile.CallName = ud.Name.String
	}
	err = udc.autoPiSvc.PatchVehicleProfile(autoPiDevice.Vehicle.ID, profile)
	if err != nil {
		subLogger.Err(err).Send()
		return errors.Wrap(err, "failed to patch autopi vehicle profile")
	}
	// update autopi to unassociate from current base template
	if autoPiDevice.Template > 0 {
		err = udc.autoPiSvc.UnassociateDeviceTemplate(autoPiDevice.ID, autoPiDevice.Template)
		if err != nil {
			subLogger.Err(err).Send()
			return errors.Wrapf(err, "failed to unassociate template %d", autoPiDevice.Template)
		}
	}
	// set our template on the autoPiDevice
	err = udc.autoPiSvc.AssociateDeviceToTemplate(autoPiDevice.ID, templateID)
	if err != nil {
		subLogger.Err(err).Send()
		return errors.Wrapf(err, "failed to associate autoPiDevice %s to template %d", autoPiDevice.ID, templateID)
	}
	// apply for next reboot
	err = udc.autoPiSvc.ApplyTemplate(autoPiDevice.ID, templateID)
	if err != nil {
		subLogger.Err(err).Send()
		return errors.Wrapf(err, "failed to apply autoPiDevice %s with template %d", autoPiDevice.ID, templateID)
	}
	// send sync command in case autoPiDevice is on at this moment (should be during initial setup)
	commandResp, err := udc.autoPiSvc.CommandSyncDevice(autoPiDevice.ID)
	if err != nil {
		subLogger.Err(err).Send()
		return errors.Wrapf(err, "failed to sync changes to autoPiDevice %s", autoPiDevice.ID)
	}

	// update database with integration status
	apiInt.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
	// add autopi job
	udMetadata.AutoPiCommandJobs = append(udMetadata.AutoPiCommandJobs, services.UserDeviceAPIIntegrationJob{
		CommandJobID: commandResp.Jid,
		CommandState: "sent",
		CommandRaw:   "state.sls pending",
		LastUpdated:  time.Now().UTC(),
	})
	err = apiInt.Metadata.Marshal(udMetadata)
	if err != nil {
		subLogger.Err(err).Send()
		return errors.Wrap(err, "failed to marshall user device integration metadata")
	}

	_, err = apiInt.Update(c.Context(), tx, boil.Whitelist(models.UserDeviceAPIIntegrationColumns.Status,
		models.UserDeviceAPIIntegrationColumns.UpdatedAt, models.UserDeviceAPIIntegrationColumns.Metadata))
	if err != nil {
		subLogger.Err(err).Send()
		return errors.Wrap(err, "failed to update integration status to PendingFirstData")
	}

	if err = tx.Commit(); err != nil {
		subLogger.Err(err).Send()
		return errors.Wrap(err, "failed to commit new autopi integration")
	}
	// send kafka message to autopi ingest registrar
	err = udc.autoPiIngestRegistrar.Register(autoPiDevice.ID, ud.ID, integration.ID)
	if err != nil {
		subLogger.Err(err).Msg("autopi ingest registrar error producing message to register")
		return err
	}
	subLogger.Info().Msg("succesfully registered autoPi integration. Now waiting on webhook for successful command.")

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) registerSmartcarIntegration(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, integ *models.Integration, ud *models.UserDevice) error {
	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	token, err := udc.smartcarClient.ExchangeCode(c.Context(), reqBody.Code, reqBody.RedirectURI)
	if err != nil {
		logger.Err(err).Msg("Error exchanging authorization code with Smartcar")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failure exchanging code with Smartcar: %s", err))
	}

	externalID, err := udc.smartcarClient.GetExternalID(c.Context(), token.Access)
	if err != nil {
		return err
	}

	vin, err := udc.smartcarClient.GetVIN(c.Context(), token.Access, externalID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error communicating with Smartcar.")
	}

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.Environment == "prod" {
		// Probably a race condition here.
		var conflict bool
		conflict, err = models.UserDevices(
			models.UserDeviceWhere.ID.NEQ(ud.ID), // If you want to re-register, or register a different integration, that's okay.
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(c.Context(), tx)
		if err != nil {
			return err
		}

		if conflict {
			return fiber.NewError(fiber.StatusConflict, "VIN already used for another device's integration")
		}
	}

	// Have to save this because it's not easy to re-load the relation if we do correct the device
	// definition.
	deviceMake := ud.R.DeviceDefinition.R.DeviceMake.Name
	year, err := udc.smartcarClient.GetYear(c.Context(), token.Access, externalID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error communicating with Smartcar.")
	}

	if err := udc.fixSmartcarDeviceYear(c.Context(), logger, tx, integ, ud, year); err != nil {
		logger.Err(err).Msg("Failed to correct Smartcar device definition year.")
	}

	perms, err := udc.smartcarClient.GetEndpoints(c.Context(), token.Access, externalID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error communicating with Smartcar.")
	}

	meta := services.UserDeviceAPIIntegrationsMetadata{SmartcarEndpoints: perms}

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
		TaskID:           null.StringFrom(taskID),
		ExternalID:       null.StringFrom(externalID),
		UserDeviceID:     ud.ID,
		IntegrationID:    integ.ID,
		Status:           models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:      null.StringFrom(encAccess),
		AccessExpiresAt:  null.TimeFrom(token.AccessExpiry),
		RefreshToken:     null.StringFrom(encRefresh),
		RefreshExpiresAt: null.TimeFrom(token.RefreshExpiry),
		Metadata:         null.JSONFrom(b),
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		logger.Err(err).Msg("Unexpected database error inserting new Smartcar integration registration")
		return err
	}

	if err := udc.smartcarTaskSvc.StartPoll(integration); err != nil {
		return err
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.create",
		Source:  "devices-api",
		Subject: ud.ID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    ud.UserID,
			Device: services.UserDeviceEventDevice{
				ID:    ud.ID,
				Make:  deviceMake,
				Model: ud.R.DeviceDefinition.Model,
				Year:  int(ud.R.DeviceDefinition.Year),
				VIN:   vin,
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     integ.ID,
				Type:   integ.Type,
				Style:  integ.Style,
				Vendor: integ.Vendor,
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Failed sending device integration creation event")
	}

	if err := tx.Commit(); err != nil {
		logger.Error().Msg("Failed to commit new integration")
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("failed to commit new integration: %s", err))
	}

	logger.Info().Msg("Finished Smartcar device registration")

	return c.SendStatus(fiber.StatusNoContent)
}

func (udc *UserDevicesController) registerDeviceTesla(c *fiber.Ctx, logger *zerolog.Logger, tx *sql.Tx, userDeviceID string, integ *models.Integration, ud *models.UserDevice) error {
	reqBody := new(RegisterDeviceIntegrationRequest)
	if err := c.BodyParser(reqBody); err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	// We'll use this to kick off the job
	teslaID, err := strconv.Atoi(reqBody.ExternalID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "externalId for Tesla must be a positive integer")
	}
	v, err := udc.teslaService.GetVehicle(reqBody.AccessToken, teslaID)
	if err != nil {
		return errorResponseHandler(c, err, fiber.StatusBadRequest)
	}

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if udc.Settings.Environment == "prod" {
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
			return fiber.NewError(fiber.StatusConflict, "VIN already used for another device's integration")
		}
	}

	if err := udc.fixTeslaDeviceDefinition(c.Context(), logger, tx, integ, ud, v.VIN); err != nil {
		return errors.Wrap(err, "Failed to fix up device definition")
	}

	encAccessToken, err := udc.cipher.Encrypt(reqBody.AccessToken)
	if err != nil {
		return opaqueInternalError
	}

	encRefreshToken, err := udc.cipher.Encrypt(reqBody.RefreshToken)
	if err != nil {
		return opaqueInternalError
	}

	taskID := ksuid.New().String()

	integration := models.UserDeviceAPIIntegration{
		UserDeviceID:    userDeviceID,
		IntegrationID:   integ.ID,
		ExternalID:      null.StringFrom(reqBody.ExternalID),
		Status:          models.UserDeviceAPIIntegrationStatusPendingFirstData,
		AccessToken:     null.StringFrom(encAccessToken),
		AccessExpiresAt: null.TimeFrom(time.Now().Add(time.Duration(reqBody.ExpiresIn) * time.Second)),
		RefreshToken:    null.StringFrom(encRefreshToken), // Don't know when this expires.
		TaskID:          null.StringFrom(taskID),
	}

	if err := integration.Insert(c.Context(), tx, boil.Infer()); err != nil {
		logger.Err(err).Msg("Unexpected database error inserting new Tesla integration registration")
		return err
	}

	ud.VinIdentifier = null.StringFrom(v.VIN)
	ud.VinConfirmed = true
	_, err = ud.Update(c.Context(), tx, boil.Infer())
	if err != nil {
		return err
	}

	err = udc.eventService.Emit(&services.Event{
		Type:    "com.dimo.zone.device.integration.create",
		Source:  "devices-api",
		Subject: userDeviceID,
		Data: services.UserDeviceIntegrationEvent{
			Timestamp: time.Now(),
			UserID:    ud.UserID,
			Device: services.UserDeviceEventDevice{
				ID:    userDeviceID,
				Make:  "Tesla", // this method is specific to Tesla so ok to hardcode
				Model: ud.R.DeviceDefinition.Model,
				Year:  int(ud.R.DeviceDefinition.Year),
				VIN:   v.VIN,
			},
			Integration: services.UserDeviceEventIntegration{
				ID:     integ.ID,
				Type:   integ.Type,
				Style:  integ.Style,
				Vendor: integ.Vendor,
			},
		},
	})
	if err != nil {
		logger.Err(err).Msg("Failed sending device integration creation event")
	}

	if err := udc.teslaService.WakeUpVehicle(reqBody.AccessToken, teslaID); err != nil {
		logger.Err(err).Msg("Failed waking up device")
	}

	if err := udc.teslaTaskService.StartPoll(v, &integration); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info().Msg("Finished Tesla device registration")

	return c.SendStatus(fiber.StatusNoContent)
}

// fixTeslaDeviceDefinition tries to use the VIN provided by Tesla to correct the device definition
// used by a device.
//
// We do not attempt to create any new entries in integrations, device_definitions, or
// device_integrations. This should all be handled elsewhere for Tesla.
func (udc *UserDevicesController) fixTeslaDeviceDefinition(ctx context.Context, logger *zerolog.Logger, exec boil.ContextExecutor, integ *models.Integration, ud *models.UserDevice, vin string) error {
	vinMake := "Tesla"
	vinModel := shared.VIN(vin).TeslaModel()
	vinYear := shared.VIN(vin).Year()

	dd := ud.R.DeviceDefinition

	if dd.R.DeviceMake.Name != "Tesla" || dd.Model != vinModel || int(dd.Year) != vinYear {
		logger.Warn().Msgf(
			"Device was attached to %s, %s, %d but should be %s, %s, %d",
			dd.R.DeviceMake.Name, dd.Model, dd.Year,
			vinMake, vinModel, vinYear,
		)

		region := ""
		if countryRecord := services.FindCountry(ud.CountryCode.String); countryRecord != nil {
			region = countryRecord.Region
		}

		newDD, err := models.DeviceDefinitions(
			qm.InnerJoin(models.TableNames.DeviceMakes+" on "+models.DeviceMakeTableColumns.ID+" = "+models.DeviceDefinitionTableColumns.DeviceMakeID),
			models.DeviceMakeWhere.Name.EQ(vinMake),
			models.DeviceDefinitionWhere.Model.EQ(vinModel),
			models.DeviceDefinitionWhere.Year.EQ(int16(vinYear)),
			qm.Load(
				models.DeviceDefinitionRels.DeviceIntegrations,
				models.DeviceIntegrationWhere.IntegrationID.EQ(integ.ID),
				models.DeviceIntegrationWhere.Region.EQ(region),
			),
		).One(ctx, exec)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("no device definition %s, %s, %d", vinMake, vinModel, vinYear)
			}
			return fmt.Errorf("database error: %w", err)
		}

		if len(newDD.R.DeviceIntegrations) == 0 {
			return fmt.Errorf("correct device definition %s has no integration %s for country %s", newDD.ID, integ.ID, ud.CountryCode.String)
		}

		if err := ud.SetDeviceDefinition(ctx, exec, false, newDD); err != nil {
			return fmt.Errorf("failed switching device definition to %s: %w", newDD.ID, err)
		}
	}

	return nil
}

// fixSmartcarDeviceYear tries to use the MMY provided by Smartcar to at least correct the year of
// the device definition used by the device.
//
// We do not attempt to create any new entries in integrations, device_definitions, or
// device_integrations. This seems too dangerous to me.
func (udc *UserDevicesController) fixSmartcarDeviceYear(ctx context.Context, logger *zerolog.Logger, exec boil.ContextExecutor, integ *models.Integration, ud *models.UserDevice, year int) error {
	dd := ud.R.DeviceDefinition

	if int(dd.Year) != year {
		logger.Warn().Msgf("Device was attached to year %d but should be %d.", dd.Year, year)
		region := ""
		if countryRecord := services.FindCountry(ud.CountryCode.String); countryRecord != nil {
			region = countryRecord.Region
		}

		newDD, err := models.DeviceDefinitions(
			models.DeviceDefinitionWhere.DeviceMakeID.EQ(dd.DeviceMakeID),
			models.DeviceDefinitionWhere.Model.EQ(dd.Model),
			models.DeviceDefinitionWhere.Year.EQ(int16(year)),
			qm.Load(
				models.DeviceDefinitionRels.DeviceIntegrations,
				models.DeviceIntegrationWhere.IntegrationID.EQ(integ.ID),
				models.DeviceIntegrationWhere.Region.EQ(region),
			),
		).One(ctx, exec)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("no device definition %s, %s, %d", dd.R.DeviceMake.Name, dd.Model, year)
			}
			return fmt.Errorf("database error: %w", err)
		}

		if len(newDD.R.DeviceIntegrations) == 0 {
			return fmt.Errorf("correct device definition %s has no integration %s for region %s", newDD.ID, integ.ID, region)
		}

		if err := ud.SetDeviceDefinition(ctx, exec, false, newDD); err != nil {
			return fmt.Errorf("failed switching device definition to %s: %w", newDD.ID, err)
		}
	}

	return nil
}

// createDeviceIntegrationIfAutoPi will create a device_integration on the fly if the integrationID belongs to AutoPi.
// returns deviceIntegration including integration relationship
func createDeviceIntegrationIfAutoPi(ctx context.Context, integrationID, deviceDefinitionID, region string, exec boil.ContextExecutor) (*models.DeviceIntegration, error) {
	autoPiInteg, err := services.GetOrCreateAutoPiIntegration(ctx, exec)
	if err != nil {
		return nil, err
	}
	if autoPiInteg.ID == integrationID {
		// create device integ on the fly
		di := models.DeviceIntegration{
			DeviceDefinitionID: deviceDefinitionID,
			IntegrationID:      integrationID,
			Region:             region,
		}
		err = di.Insert(ctx, exec, boil.Infer())
		if err != nil {
			return nil, err
		}
		di.R = di.R.NewStruct()
		di.R.Integration = autoPiInteg
		return &di, nil
	}
	return nil, nil
}

/** Structs for request / response **/

type UserDeviceIntegrationStatus struct {
	IntegrationID string    `json:"integrationId"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	ExternalID    *string   `json:"externalId"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Metadata      null.JSON `json:"metadata" swaggertype:"string"`
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
