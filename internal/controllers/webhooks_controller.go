package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type WebhooksController struct {
	DBS      func() *database.DBReaderWriter
	Settings *config.Settings
}

func NewWebhooksController(settings *config.Settings, dbs func() *database.DBReaderWriter) WebhooksController {
	return WebhooksController{
		DBS:      dbs,
		Settings: settings,
	}
}

// ProcessCommand handles the command webhook request
func (wc *WebhooksController) ProcessCommand(c *fiber.Ctx) error {
	// process payload
	awp := &AutoPiWebhookPayload{}
	if err := c.BodyParser(awp); err != nil {
		// Return status 400 and error message.
		return fiber.NewError(fiber.StatusBadRequest, "unable to parse webhook payload")
	}
	// hmac signature validation
	reqSig := c.Get("X-Request-Signature")
	if !validateSignature(wc.Settings.AutoPiAPIToken, string(c.Body()), reqSig) {
		return c.Status(fiber.StatusUnauthorized).SendString("invalid signature")
	}
	autoPiInteg, err := services.GetOrCreateAutoPiIntegration(c.Context(), wc.DBS().Reader)
	if err != nil {
		return err
	}
	// we could have situation where there are multiple results, eg. if the AutoPi was moved from one car to another
	apiIntegrations, err := models.UserDeviceAPIIntegrations(models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(autoPiInteg.ID),
		models.UserDeviceAPIIntegrationWhere.ExternalID.EQ(null.StringFrom(strconv.Itoa(awp.DeviceID)))).All(c.Context(), wc.DBS().Reader)
	if err != nil {
		return err
	}
	// make sure the jobId matches just in case
	foundMatch := false
	for _, ai := range apiIntegrations {
		m := new(services.UserDeviceAPIIntegrationsMetadata)
		err := ai.Metadata.Unmarshal(m)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshall metadata json for autopi device id %d", awp.DeviceID)
		}
		if *m.AutoPiSyncCommandJobID == awp.Jid {
			foundMatch = true
			ai.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
			m.AutoPiSyncCommandState = &awp.State
			err = ai.Metadata.Marshal(m)
			if err != nil {
				return errors.Wrap(err, "failed to marshal user device api integration metadata json from autopi webhook")
			}
			_, err = ai.Update(c.Context(), wc.DBS().Writer, boil.Whitelist(
				models.UserDeviceAPIIntegrationColumns.Metadata, models.UserDeviceAPIIntegrationColumns.Status,
				models.UserDeviceAPIIntegrationColumns.UpdatedAt))
			if err != nil {
				return errors.Wrap(err, "failed to save user device api changes to db from autopi webhook")
			}
		}
	}
	if !foundMatch {
		return c.Status(fiber.StatusBadRequest).SendString(
			fmt.Sprintf("could not find record with device id %d and job id %s", awp.DeviceID, awp.Jid))
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
