package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/RichardKnop/machinery/v1"
	machinery_config "github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
)

//go:generate mockgen -source task_service.go -destination mocks/task_service_mock.go

type ITaskService interface {
	StartSmartcarRegistrationTasks(userDeviceID, integrationID string) (err error)
	StartSmartcarRefresh(userDeviceID, integrationID string) (err error)
	StartSmartcarDeregistrationTasks(userDeviceID, integrationID, externalID, accessToken string) (err error)
}

type TaskService struct {
	Settings     *config.Settings
	DBS          func() *database.DBReaderWriter
	Log          *zerolog.Logger
	Machinery    *machinery.Server
	Producer     sarama.SyncProducer
	DeviceDefSvc IDeviceDefinitionService
}

const smartcarWebhookURL = "https://api.smartcar.com/v2.0/vehicles/%s/webhooks/%s"
const smartcarBatchURL = "https://api.smartcar.com/v2.0/vehicles/%s/batch"
const smartcarVINURL = "https://api.smartcar.com/v2.0/vehicles/%s/vin"

const smartcarConnectVehicleTask = "smartcar_connect_vehicle"
const smartcarGetInitialDataTask = "smartcar_get_initial_data"
const smartcarDisconnectVehicleTask = "smartcar_disconnect_vehicle"
const failIntegrationTask = "fail_integration"

const ingestSmartcarRegistrationTopic = "table.device.integration.smartcar"
const smartcarRegistrationEventType = "zone.dimo.device.integration.smartcar.register"

type batchRequest struct {
	Requests []batchRequestRequest `json:"requests"`
}

type batchRequestRequest struct {
	Path string `json:"path"`
}

var batchRequestFixed = batchRequest{
	Requests: []batchRequestRequest{
		{"/"},
		{"/battery"},
		{"/battery/capacity"},
		{"/charge"},
		{"/fuel"},
		{"/location"},
		{"/odometer"},
		{"/engine/oil"},
		{"/permissions"},
		{"/tires/pressure"},
		{"/vin"},
	},
}

func (t *TaskService) subscribeVehicle(vehicleID, accessToken string) error {
	url := fmt.Sprintf(smartcarWebhookURL, vehicleID, t.Settings.SmartcarWebhookID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to construct webhook subscription request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("SC-Unit-System", "metric")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failure making webhook subscription request: %w", err)
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook subscription request returned status code %d", resp.StatusCode)
	}
	return nil
}

func (t *TaskService) unsubscribeVehicle(vehicleID, accessToken string) (err error) {
	url := fmt.Sprintf(smartcarWebhookURL, vehicleID, t.Settings.SmartcarWebhookID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("error from Smartcar detaching vehicle %s from webhook %s, status code %d", vehicleID, t.Settings.SmartcarWebhookID, resp.StatusCode)
	}
	return
}

func (t *TaskService) vinRequest(vehicleID, accessToken string) (string, error) {
	url := fmt.Sprintf(smartcarVINURL, vehicleID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to construct VIN request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("VIN request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("VIN request returned status code %d", resp.StatusCode)
	}

	var richResp struct {
		VIN string `json:"vin"`
	}
	err = json.NewDecoder(resp.Body).Decode(&richResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse VIN response body: %w", err)
	}

	return richResp.VIN, nil
}

// batchRequest makes a batch information request to Smartcar using the given Smartcar vehicle ID.
// If this is successful, returns the raw response body.
func (t *TaskService) batchRequest(vehicleID, accessToken string) ([]byte, error) {
	url := fmt.Sprintf(smartcarBatchURL, vehicleID)

	// TODO: Lift this up, it's always the same
	requestBytes, err := json.Marshal(batchRequestFixed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request to JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to construct Smartcar batch request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("SC-Unit-System", "metric")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("batch request to Smartcar failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("batch request to Smartcar returned status code %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Smartcar batch response body: %w", err)
	}

	return respBody, nil
}

type cloudEventMessage struct {
	ID          string      `json:"id"`
	Source      string      `json:"source"`
	SpecVersion string      `json:"specversion"`
	Subject     string      `json:"subject"`
	Time        time.Time   `json:"time"`
	Type        string      `json:"type"`
	Data        interface{} `json:"data"`
}

type registrationData struct {
	DeviceID   string      `json:"deviceId"`
	ExternalID null.String `json:"externalId"`
}

func (t *TaskService) smartcarConnectVehicle(userDeviceID, integrationID string) (err error) {
	tx, err := t.DBS().Writer.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to create Smartcar registration transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	integ, err := models.UserDeviceAPIIntegrations(
		qm.Where("user_device_id = ?", userDeviceID),
		qm.And("integration_id = ?", integrationID),
		qm.Load("UserDevice")).One(context.Background(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("could not find API integration")
		}
		return fmt.Errorf("failed querying database for device's API integration: %w", err)
	}

	client := smartcar.NewClient()
	vehicleIDs, err := client.GetVehicleIDs(context.Background(), &smartcar.VehicleIDsParams{
		Access: integ.AccessToken,
	})
	if err != nil {
		return fmt.Errorf("failed request to Smartcar for vehicle IDs: %w", err)
	}
	if len(*vehicleIDs) != 1 {
		return fmt.Errorf("expected only one vehicle ID from Smartcar, but got %d", len(*vehicleIDs))
	}

	vehicleID := (*vehicleIDs)[0]
	t.Log.Info().Str("userDeviceId", userDeviceID).Str("integrationId", integrationID).Msgf("Got Smartcar vehicle ID %s", vehicleID)
	integ.ExternalID = null.StringFrom(vehicleID)
	_, err = integ.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed updating API integration with Smartcar vehicle ID: %w", err)
	}

	vin, err := t.vinRequest(vehicleID, integ.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to obtain vehicle VIN from Smartcar: %w", err)
	}

	// Prevent users from connecting a vehicle if it's already connected through another user
	// device object. Disabled outside of prod for ease of testing.
	if t.Settings.Environment == "prod" {
		// Probably a race condition here.
		var conflict bool
		conflict, err = models.UserDevices(
			models.UserDeviceWhere.ID.NEQ(userDeviceID), // If you want to re-register, that's okay.
			models.UserDeviceWhere.VinIdentifier.EQ(null.StringFrom(vin)),
			models.UserDeviceWhere.VinConfirmed.EQ(true),
		).Exists(context.Background(), tx)
		if err != nil {
			return fmt.Errorf("database error searching for existing API integration instance: %w", err)
		}

		if conflict {
			integ.Status = models.UserDeviceAPIIntegrationStatusDuplicateIntegration
			_, err = integ.Update(context.Background(), tx, boil.Whitelist("status"))
			if err != nil {
				return fmt.Errorf("database error marking API integration as duplicate: %w", err)
			}
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("database error marking API integration as duplicate: %w", err)
			}
			// This will probably get retried. That is unfortunate!
			return fmt.Errorf("VIN %s is already confirmed and attached to another device", vin)
		}
	}

	ud := integ.R.UserDevice
	ud.VinIdentifier = null.StringFrom(vin)
	ud.VinConfirmed = true
	_, err = ud.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		return fmt.Errorf("database failure adding Smartcar-confirmed VIN to user device: %w", err)
	}
	go func() {
		err := t.DeviceDefSvc.UpdateDeviceDefinitionFromNHTSA(context.Background(), ud.DeviceDefinitionID, vin)
		if err != nil {
			t.Log.Err(err).Msgf("error when trying to update deviceDefinitionID: %s from NHTSA for vin: %s", ud.DeviceDefinitionID, vin)
		}
	}()

	msg := cloudEventMessage{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + integrationID,
		Subject:     userDeviceID,
		SpecVersion: "1.0",
		Time:        time.Now(),
		Type:        smartcarRegistrationEventType,
		Data: registrationData{
			DeviceID:   userDeviceID,
			ExternalID: integ.ExternalID,
		},
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to create JSON body for Smartcar registration event: %w", err)
	}
	message := &sarama.ProducerMessage{
		Topic: ingestSmartcarRegistrationTopic,
		Key:   sarama.StringEncoder(ud.ID),
		Value: sarama.ByteEncoder(msgBytes),
	}
	_, _, err = t.Producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to emit Smartcar registration event: %w", err)
	}

	err = t.subscribeVehicle(vehicleID, integ.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to subscribe vehicle to webhook: %w", err)
	}

	integ.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
	_, err = integ.Update(context.Background(), tx, boil.Whitelist("status"))
	if err != nil {
		return fmt.Errorf("database failure setting integration status to \"pending data\": %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failure committing Smartcar registration to database: %w", err)
	}
	return nil
}

func (t *TaskService) smartcarDisconnectVehicle(userDeviceID, integrationID, externalID, accessToken string) (err error) {
	msg := cloudEventMessage{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + integrationID,
		Subject:     userDeviceID,
		SpecVersion: "1.0",
		Time:        time.Now(),
		Type:        smartcarRegistrationEventType,
		Data: registrationData{
			DeviceID:   userDeviceID,
			ExternalID: null.StringFromPtr(nil),
		},
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	message := &sarama.ProducerMessage{
		Topic: ingestSmartcarRegistrationTopic,
		Key:   sarama.StringEncoder(userDeviceID),
		Value: sarama.ByteEncoder(msgBytes),
	}
	_, _, err = t.Producer.SendMessage(message)
	if err != nil {
		return
	}

	err = t.unsubscribeVehicle(externalID, accessToken)
	return
}

func formatBatchAsWebhook(batchBytes []byte, vehicleID string) ([]byte, error) {
	var batch struct {
		Responses json.RawMessage `json:"responses"`
	}
	err := json.Unmarshal(batchBytes, &batch)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse batch response: %w", err)
	}

	type webhookVehicle struct {
		Data      json.RawMessage `json:"data"`
		RequestID string          `json:"requestId"`
		Timestamp time.Time       `json:"timestamp"`
		VehicleID string          `json:"vehicleId"`
	}

	hook := struct {
		EventName string `json:"eventName"`
		Mode      string `json:"mode"`
		Payload   struct {
			Vehicles []webhookVehicle `json:"vehicles"` // Will only ever have one element
		} `json:"payload"`
	}{
		EventName: "schedule",
		Mode:      "live",
	}

	hook.Payload.Vehicles = []webhookVehicle{
		{
			Data:      batch.Responses,
			RequestID: "", // Not needed
			Timestamp: time.Now(),
			VehicleID: vehicleID,
		},
	}

	hookBytes, err := json.Marshal(hook)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal webhook: %w", err)
	}
	return hookBytes, nil
}

func (t *TaskService) smartcarGetInitialData(userDeviceID, integrationID string) error {
	tx, err := t.DBS().Writer.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to acquire transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	integ, err := models.UserDeviceAPIIntegrations(
		qm.Where("user_device_id = ?", userDeviceID),
		qm.And("integration_id = ?", integrationID)).One(context.Background(), tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("could not find API integration")
		}
		return fmt.Errorf("failed querying database for device's API integration: %w", err)
	}

	// Use the refresh token if the access token is expired or about to expire. We are ignoring
	// the possiblity of the refresh token also being expired. Those last for 60 days, so it
	// shouldn't happen much.
	if time.Now().Add(5 * time.Minute).After(integ.AccessExpiresAt) {
		client := smartcar.NewClient()
		auth := client.NewAuth(&smartcar.AuthParams{
			ClientID:     t.Settings.SmartcarClientID,
			ClientSecret: t.Settings.SmartcarClientSecret,
		})

		var token *smartcar.Token
		token, err = auth.ExchangeRefreshToken(context.Background(), &smartcar.ExchangeRefreshTokenParams{
			Token: integ.RefreshToken,
		})
		if err != nil {
			return fmt.Errorf("failed exchanging refresh token with Smartcar: %w", err)
		}

		integ.AccessToken = token.Access
		integ.AccessExpiresAt = token.AccessExpiry
		integ.RefreshToken = token.Refresh
		integ.RefreshExpiresAt = token.RefreshExpiry

		_, err = integ.Update(context.Background(), tx, boil.Infer())
		if err != nil {
			return fmt.Errorf("database failure saving new Smartcar tokens: %w", err)
		}
	}

	batchBytes, err := t.batchRequest(integ.ExternalID.String, integ.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to make batch request to Smartcar: %w", err)
	}

	hookBytes, err := formatBatchAsWebhook(batchBytes, integ.ExternalID.String)
	if err != nil {
		return fmt.Errorf("failed to format Smartcar batch response as a webhook: %w", err)
	}

	req, err := http.NewRequest("POST", t.Settings.IngestSmartcarURL, bytes.NewReader(hookBytes))
	if err != nil {
		return fmt.Errorf("failed constructing request to DIMO Smartcar ingest: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	benthResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send batch data to DIMO Smartcar ingest: %w", err)
	}
	defer benthResp.Body.Close() //nolint

	if benthResp.StatusCode >= 400 {
		return fmt.Errorf("sending batch data to DIMO Smartcar ingest returned status code %d", benthResp.StatusCode)
	}

	integ.Status = models.UserDeviceAPIIntegrationStatusActive
	_, err = integ.Update(context.Background(), tx, boil.Whitelist("status"))
	if err != nil {
		return fmt.Errorf("failed to set integration's status to active: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit results of batch request: %w", err)
	}
	return nil
}

func (t *TaskService) failIntegration(errString, userDeviceID, integrationID string) (err error) {
	db := t.DBS().Writer
	integ, err := models.UserDeviceAPIIntegrations(
		models.UserDeviceAPIIntegrationWhere.UserDeviceID.EQ(userDeviceID),
		models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ(integrationID),
	).One(context.Background(), db)
	if err != nil {
		return
	}
	integ.Status = models.UserDeviceAPIIntegrationStatusFailed
	_, err = integ.Update(context.Background(), db, boil.Whitelist("status"))
	return
}

func (t *TaskService) StartSmartcarRegistrationTasks(userDeviceID, integrationID string) error {
	errSig := tasks.Signature{
		Name: failIntegrationTask,
		Args: []tasks.Arg{
			{Type: "string", Value: userDeviceID},
			{Type: "string", Value: integrationID},
		},
		RetryCount: 3, // Somewhat random
	}
	sig1 := tasks.Signature{
		Name: smartcarConnectVehicleTask,
		Args: []tasks.Arg{
			{Type: "string", Value: userDeviceID},
			{Type: "string", Value: integrationID},
		},
		RetryCount: 3, // Somewhat random
		OnError:    []*tasks.Signature{&errSig},
	}
	sig2 := tasks.Signature{
		Name: smartcarGetInitialDataTask,
		Args: []tasks.Arg{
			{Type: "string", Value: userDeviceID},
			{Type: "string", Value: integrationID},
		},
		RetryCount: 3,
		OnError:    []*tasks.Signature{&errSig}, // We might want to rethink this. Failing here isn't so bad
	}

	chain, err := tasks.NewChain(&sig1, &sig2)
	if err != nil {
		return fmt.Errorf("failed to create task chain: %w", err)
	}
	_, err = t.Machinery.SendChain(chain)
	if err != nil {
		return fmt.Errorf("failed to trigger task chain: %w", err)
	}

	return nil
}

func (t *TaskService) StartSmartcarRefresh(userDeviceID, integrationID string) (err error) {
	sig := tasks.Signature{
		Name: smartcarGetInitialDataTask, // This name probably needs to change
		Args: []tasks.Arg{
			{Type: "string", Value: userDeviceID},
			{Type: "string", Value: integrationID},
		},
		RetryCount: 3,
	}
	_, err = t.Machinery.SendTask(&sig)
	return
}

func (t *TaskService) StartSmartcarDeregistrationTasks(userDeviceID, integrationID, externalID, accessToken string) (err error) {
	sig := tasks.Signature{
		Name: smartcarDisconnectVehicleTask,
		Args: []tasks.Arg{
			{Type: "string", Value: userDeviceID},
			{Type: "string", Value: integrationID},
			{Type: "string", Value: externalID},
			{Type: "string", Value: accessToken},
		},
		RetryCount: 3, // Somewhat random
	}
	_, err = t.Machinery.SendTask(&sig)
	return
}

func NewTaskService(settings *config.Settings, dbs func() *database.DBReaderWriter, deviceDefSvc *DeviceDefinitionService) *TaskService {
	var redisConn string
	if settings.RedisPassword == "" {
		redisConn = fmt.Sprintf("redis://%s", settings.RedisURL)
	} else {
		redisConn = fmt.Sprintf("redis://%s@%s", settings.RedisPassword, settings.RedisURL)
	}

	var tlsConfig *tls.Config
	if settings.RedisTLS {
		tlsConfig = new(tls.Config)
	}

	server, err := machinery.NewServer(&machinery_config.Config{
		Broker:        redisConn,
		ResultBackend: redisConn,
		TLSConfig:     tlsConfig,
	})
	if err != nil {
		panic(err)
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(strings.Split(settings.KafkaBrokers, ","), kafkaConfig)
	if err != nil {
		panic(err)
	}

	t := &TaskService{
		Settings:     settings,
		DBS:          dbs,
		Machinery:    server,
		Producer:     producer,
		DeviceDefSvc: deviceDefSvc,
	}
	err = server.RegisterTasks(map[string]interface{}{
		smartcarConnectVehicleTask:    t.smartcarConnectVehicle,
		smartcarGetInitialDataTask:    t.smartcarGetInitialData,
		smartcarDisconnectVehicleTask: t.smartcarDisconnectVehicle,
		failIntegrationTask:           t.failIntegration,
	})
	if err != nil {
		panic(err)
	}
	worker := server.NewWorker("myworker", 0)
	go func() {
		err = worker.Launch()
		if err != nil {
			panic(err)
		}
	}()

	return t
}
