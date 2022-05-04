package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/oauth2"
)

func stopKafkaTask(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, scTaskSvc services.SmartcarTaskService, taskID string) error {
	integ := &models.UserDeviceAPIIntegration{
		TaskID:        null.StringFrom(taskID),
		UserDeviceID:  "FAKE",
		IntegrationID: "FAKE",
	}

	if err := scTaskSvc.StopPoll(integ); err != nil {
		return err
	}

	return nil
}

func stopTaskByKey(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, taskKey string, producer sarama.SyncProducer) error {
	tt := struct {
		services.CloudEventHeaders
		Data interface{} `json:"data"`
	}{
		CloudEventHeaders: services.CloudEventHeaders{
			ID:          ksuid.New().String(),
			Source:      "dimo/integration/FAKE",
			SpecVersion: "1.0",
			Subject:     "FAKE",
			Time:        time.Now(),
			Type:        "zone.dimo.task.tesla.poll.stop",
		},
		Data: struct {
			TaskID        string `json:"taskId"`
			UserDeviceID  string `json:"userDeviceId"`
			IntegrationID string `json:"integrationId"`
		}{
			TaskID:        taskKey,
			UserDeviceID:  "FAKE",
			IntegrationID: "FAKE",
		},
	}

	ttb, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	producer.SendMessage(
		&sarama.ProducerMessage{
			Topic: settings.TaskStopTopic,
			Key:   sarama.StringEncoder(taskKey),
			Value: sarama.ByteEncoder(ttb),
		},
	)

	return nil
}

func startTeslaTask(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore, teslaService services.TeslaService, teslaTaskService services.TeslaTaskService, userDeviceID string, cipher shared.Cipher) error {
	teslaInt, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("Tesla")).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}
	udai, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(teslaInt.ID),
	).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return err
	}

	if !udai.TaskID.Valid {
		return fmt.Errorf("no existing TaskID")
	}

	if !udai.AccessExpiresAt.Valid {
		return fmt.Errorf("no existing token expiry")
	}

	var accessT string
	if udai.AccessExpiresAt.Time.Before(time.Now().Add(time.Minute)) {
		oldRefresh, err := cipher.Decrypt(udai.RefreshToken.String)
		if err != nil {
			return err
		}

		newToken, err := exchangeTeslaRefresh(oldRefresh)
		if err != nil {
			return err
		}
		logger.Info().Str("userDeviceId", userDeviceID).Msgf("Got new refresh token %s.", newToken.RefreshToken)
		accessT = newToken.AccessToken
		encAccess, err := cipher.Encrypt(newToken.AccessToken)
		if err != nil {
			return err
		}

		encRefresh, err := cipher.Encrypt(newToken.RefreshToken)
		if err != nil {
			return err
		}

		udai.AccessToken = null.StringFrom(encAccess)
		udai.RefreshToken = null.StringFrom(encRefresh)
		udai.AccessExpiresAt = null.TimeFrom(newToken.Expiry)

		if _, err := udai.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
			return err
		}
	} else {
		accessT, err = cipher.Decrypt(udai.AccessToken.String)
		if err != nil {
			return err
		}
	}

	teslaVehID, err := strconv.Atoi(udai.ExternalID.String)
	if err != nil {
		return err
	}

	teslaVeh, err := teslaService.GetVehicle(accessT, teslaVehID)
	if err != nil {
		return err
	}

	if err := teslaTaskService.StartPoll(teslaVeh, udai); err != nil {
		return err
	}

	return nil
}

func exchangeTeslaRefresh(token string) (*oauth2.Token, error) {
	reqs := struct {
		GrantType    string `json:"grant_type"`
		ClientID     string `json:"client_id"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}{
		GrantType:    "refresh_token",
		ClientID:     "ownerapi",
		RefreshToken: token,
		Scope:        "openid email offline_access",
	}

	reqb, err := json.Marshal(reqs)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://auth.tesla.com/oauth2/v3/token", bytes.NewBuffer(reqb))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d from token endpoint", resp.StatusCode)
	}

	resps := new(struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	})
	if err := json.NewDecoder(resp.Body).Decode(resps); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  resps.AccessToken,
		RefreshToken: resps.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(resps.ExpiresIn) * time.Second),
	}, nil
}
