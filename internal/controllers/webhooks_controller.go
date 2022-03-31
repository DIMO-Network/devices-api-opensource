package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	// process payload
	awp := new(AutoPiWebhookPayload)
	if err := c.BodyParser(awp); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, "unable to parse webhook payload")
	}
	if awp == nil || awp.Jid == "" || awp.DeviceID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid autopi webhook request payload")
	}

	// hmac signature validation
	reqSig := c.Get("X-Request-Signature")
	if !validateSignature(wc.settings.AutoPiAPIToken, string(c.Body()), reqSig) {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid autopi webhook signature")
	}
	autoPiInteg, err := services.GetOrCreateAutoPiIntegration(c.Context(), wc.dbs().Reader)
	if err != nil {
		return err
	}
	// we could have situation where there are multiple results, eg. if the AutoPi was moved from one car to another, so order by updated_at desc and grab first
	apiIntegration, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autoPiInteg.ID),
		models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom(strconv.Itoa(awp.DeviceID))),
		qm.OrderBy("updated_at desc"), qm.Limit(1)).One(c.Context(), wc.dbs().Reader)
	if err != nil {
		return err
	}
	// make sure the jobId matches just in case
	foundMatch := false
	udMetadata := new(services.UserDeviceAPIIntegrationsMetadata)
	err = apiIntegration.Metadata.Unmarshal(udMetadata)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshall metadata json for autopi device id %d", awp.DeviceID)
	}

	for i, job := range udMetadata.AutoPiCommandJobs {
		if job.CommandJobID == awp.Jid {
			foundMatch = true
			// only update status of integration if command is the sync command that we use for template setup
			if job.CommandRaw == "state.sls pending" {
				apiIntegration.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
			}
			udMetadata.AutoPiCommandJobs[i].CommandState = awp.State
			udMetadata.AutoPiCommandJobs[i].LastUpdated = time.Now().UTC()

			err = apiIntegration.Metadata.Marshal(udMetadata)
			if err != nil {
				return errors.Wrap(err, "failed to marshal user device api integration metadata json from autopi webhook")
			}
			_, err = apiIntegration.Update(c.Context(), wc.dbs().Writer, boil.Whitelist(
				models.UserDeviceAPIIntegrationColumns.Metadata, models.UserDeviceAPIIntegrationColumns.Status,
				models.UserDeviceAPIIntegrationColumns.UpdatedAt))
			if err != nil {
				return errors.Wrap(err, "failed to save user device api changes to db from autopi webhook")
			}

			break
		}
	}
	if foundMatch {
		wc.log.Info().Msgf("processed webhook from raw autopi command, with autopi device id %d and job id %s",
			awp.DeviceID, awp.Jid)
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

// AutoPiWebhookPayload webhook payload from autopi
type AutoPiWebhookPayload struct {
	Response struct {
		Tag  string `json:"tag"`
		Data struct {
			FunArgs []interface{} `json:"fun_args"`
			Jid     string        `json:"jid"`
			Return  bool          `json:"return"`
			Retcode int           `json:"retcode"`
			Success bool          `json:"success"`
			Cmd     string        `json:"cmd"`
			Stamp   string        `json:"_stamp"`
			Fun     string        `json:"fun"`
			ID      string        `json:"id"`
		} `json:"data"`
	} `json:"response"`
	Jid      string `json:"jid"`
	State    string `json:"state"`
	Success  bool   `json:"success"`
	DeviceID int    `json:"device_id"`
}
