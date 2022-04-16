package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type WebhooksController struct {
	dbs      func() *database.DBReaderWriter
	settings *config.Settings
	log      *zerolog.Logger
}

func NewWebhooksController(settings *config.Settings, dbs func() *database.DBReaderWriter, log *zerolog.Logger) WebhooksController {
	return WebhooksController{
		dbs:      dbs,
		settings: settings,
		log:      log,
	}
}

// ProcessCommand handles the command webhook request
func (wc *WebhooksController) ProcessCommand(c *fiber.Ctx) error {
	logger := wc.log.With().
		Str("integration", "autopi").
		Str("handler", "webhooks.ProcessCommand").
		Logger()
	logger.Info().Msg("attempting to process webhook request")
	// grab values from webhook using gjson instead of parsing since more error-prone as they make changes
	apwJID := gjson.GetBytes(c.Body(), "jid")
	apwDeviceID := gjson.GetBytes(c.Body(), "device_id")
	apwState := gjson.GetBytes(c.Body(), "state")

	if !apwJID.Exists() || !apwDeviceID.Exists() || !apwState.Exists() {
		logger.Error().Str("payload", string(c.Body())).Msg("no jobId or deviceId found in payload")
		return fiber.NewError(fiber.StatusBadRequest, "invalid autopi webhook request payload")
	}
	// hmac signature validation
	reqSig := c.Get("X-Request-Signature")
	if !validateSignature(wc.settings.AutoPiAPIToken, string(c.Body()), reqSig) {
		logger.Error().Str("payload", string(c.Body())).Msg("invalid webhook signature")
		return fiber.NewError(fiber.StatusUnauthorized, "invalid autopi webhook signature")
	}
	autoPiInteg, err := services.GetOrCreateAutoPiIntegration(c.Context(), wc.dbs().Reader)
	if err != nil {
		return err
	}
	// we could have situation where there are multiple results, eg. if the AutoPi was moved from one car to another, so order by updated_at desc and grab first
	apiIntegration, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autoPiInteg.ID),
		models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom(apwDeviceID.String())),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).One(c.Context(), wc.dbs().Reader)
	if err != nil {
		return err
	}
	// make sure the jobId matches just in case
	foundMatch := false
	udMetadata := new(services.UserDeviceAPIIntegrationsMetadata)
	err = apiIntegration.Metadata.Unmarshal(udMetadata)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshall metadata json for autopi device id %s", apwDeviceID.String())
	}

	for i, job := range udMetadata.AutoPiCommandJobs {
		if job.CommandJobID == apwJID.String() {
			foundMatch = true
			// only update status of integration if command is the sync command that we use for template setup
			if job.CommandRaw == "state.sls pending" && strings.EqualFold(apwState.String(), "COMMAND_EXECUTED") {
				apiIntegration.Status = models.UserDeviceAPIIntegrationStatusActive
			}
			udMetadata.AutoPiCommandJobs[i].CommandState = apwState.String()
			udMetadata.AutoPiCommandJobs[i].LastUpdated = time.Now().UTC()

			err = apiIntegration.Metadata.Marshal(udMetadata)
			if err != nil {
				return errors.Wrap(err, "failed to marshal user device api integration metadata json from autopi webhook")
			}
			_, err = apiIntegration.Update(c.Context(), wc.dbs().Writer, boil.Whitelist(
				models.UserDeviceAPIIntegrationColumns.Metadata, models.UserDeviceAPIIntegrationColumns.Status,
				models.UserDeviceAPIIntegrationColumns.UpdatedAt))
			if err != nil {
				logger.Err(err).Msg("failed to save user device api changes")
				return errors.Wrap(err, "failed to save user device api changes to db from autopi webhook")
			}

			break
		}
	}
	if foundMatch {
		logger.Info().Msgf("processed webhook successfully, with autopi deviceId %s and jobId %s",
			apwDeviceID.String(), apwJID.String())
	} else {
		logger.Warn().Msgf("failed to process webhook because did not find a match for deviceId %s and jobId %s",
			apwDeviceID.String(), apwJID.String())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func validateSignature(secret, data, expectedSignature string) bool {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))
	// Write Data to it
	h.Write([]byte(data))
	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha == expectedSignature
}
