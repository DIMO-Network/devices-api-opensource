package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/RichardKnop/machinery/v1"
	machinery_config "github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	smartcar "github.com/smartcar/go-sdk"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TaskService struct {
	Settings  *config.Settings
	DBS       func() *database.DBReaderWriter
	Log       *zerolog.Logger
	Machinery *machinery.Server
	Publisher *kafka.Publisher
}

const smartcarWebhookURL = "https://api.smartcar.com/v2.0/vehicles/%s/webhooks/%s"
const smartcarBatchURL = "https://api.smartcar.com/v2.0/vehicles/%s/batch"
const smartcarVINURL = "https://api.smartcar.com/v2.0/vehicles/%s/vin"

const smartcarConnectVehicleTask = "smartcar_connect_vehicle"
const smartcarGetInitialDataTask = "smartcar_get_initial_data"

type batchRequest struct {
	Requests []batchRequestRequest `json:"requests"`
}

type batchRequestRequest struct {
	Path string `json:"path"`
}

var batchRequestFixed = batchRequest{
	Requests: []batchRequestRequest{
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

func (t *TaskService) subscribeVehicle(vehicleID, accessToken string) (err error) {
	url := fmt.Sprintf(smartcarWebhookURL, vehicleID, t.Settings.SmartcarWebhookID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("error from Smartcar attaching vehicle %s to webhook %s, status code %d", vehicleID, t.Settings.SmartcarWebhookID, resp.StatusCode)
	}
	return
}

func (t *TaskService) vinRequest(vehicleID, accessToken string) (vin string, err error) {
	url := fmt.Sprintf(smartcarVINURL, vehicleID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("error from Smartcar requesting VIN for %s, status code %d", vehicleID, resp.StatusCode)
		return
	}
	var richResp struct {
		VIN string `json:"vin"`
	}
	err = json.NewDecoder(resp.Body).Decode(&richResp)
	vin = richResp.VIN
	return
}

func (t *TaskService) batchRequest(vehicleID, accessToken string) (response []byte, err error) {
	url := fmt.Sprintf(smartcarBatchURL, vehicleID)

	requestBytes, err := json.Marshal(batchRequestFixed)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBytes))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("error from Smartcar, status code %d", resp.StatusCode)
	}
	response, err = ioutil.ReadAll(resp.Body)
	return
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

func (t *TaskService) smartcarConnectVehicle(userDeviceID, integrationID string) (err error) {
	client := smartcar.NewClient()
	tx, err := t.DBS().Writer.BeginTx(context.Background(), nil)
	if err != nil {
		return
	}
	defer tx.Rollback() //nolint
	integ, err := models.UserDeviceAPIIntegrations(
		qm.Where("user_device_id = ?", userDeviceID),
		qm.And("integration_id = ?", integrationID),
		qm.Load("UserDevice")).One(context.Background(), tx)
	if err != nil {
		return
	}
	vehicleIDs, err := client.GetVehicleIDs(context.Background(), &smartcar.VehicleIDsParams{
		Access: integ.AccessToken,
	})
	if err != nil {
		return
	}
	if len(*vehicleIDs) != 1 {
		err = fmt.Errorf("expected only one vehicle id, but got %d", len(*vehicleIDs))
		return
	}
	vehicleID := (*vehicleIDs)[0]
	integ.ExternalID = null.StringFrom(vehicleID)
	_, err = integ.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		return
	}

	vin, err := t.vinRequest(vehicleID, integ.AccessToken)
	if err != nil {
		return
	}

	ud := integ.R.UserDevice
	ud.VinIdentifier = null.StringFrom(vin)
	_, err = ud.Update(context.Background(), tx, boil.Infer())
	if err != nil {
		return
	}

	if err != nil {
		return
	}
	msg := cloudEventMessage{
		ID:          ksuid.New().String(),
		Source:      "dimo/integration/" + integrationID,
		Subject:     ud.ID,
		SpecVersion: "1.0",
		Time:        time.Now(),
		Type:        "zone.dimo.device.integration.smartcar.register",
		Data: struct {
			DeviceID   string `json:"deviceId"`
			ExternalID string `json:"externalId"`
		}{ud.ID, integ.ExternalID.String},
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	message := message.NewMessage(msg.ID, msgBytes)
	err = t.Publisher.Publish("table.device.integration.smartcar", message)
	if err != nil {
		return
	}

	err = t.subscribeVehicle(vehicleID, integ.AccessToken)
	if err != nil {
		return
	}

	integ.Status = models.UserDeviceAPIIntegrationStatusPendingFirstData
	_, err = integ.Update(context.Background(), tx, boil.Whitelist("status"))
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func (t *TaskService) smartcarGetInitialData(userDeviceID, integrationID string) (err error) {
	db := t.DBS().Writer
	integ, err := models.UserDeviceAPIIntegrations(
		qm.Where("user_device_id = ?", userDeviceID),
		qm.And("integration_id = ?", integrationID)).One(context.Background(), db)
	if err != nil {
		return
	}

	resp, err := t.batchRequest(integ.ExternalID.String, integ.AccessToken)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", t.Settings.IngestSmartcarURL, bytes.NewReader(resp))
	req.Header.Set("Content-Type", "application/json")
	benthResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer benthResp.Body.Close()
	if benthResp.StatusCode >= 400 {
		err = fmt.Errorf("error from Benthos, status code %d", benthResp.StatusCode)
		return
	}

	integ.Status = models.UserDeviceAPIIntegrationStatusActive
	_, err = integ.Update(context.Background(), db, boil.Whitelist("status"))
	return
}

func (t *TaskService) BeginSmartcar(userDeviceID, integrationID string) (err error) {
	sig1 := tasks.Signature{
		Name: smartcarConnectVehicleTask,
		Args: []tasks.Arg{
			{Type: "string", Value: userDeviceID},
			{Type: "string", Value: integrationID},
		},
		RetryCount: 3, // Somewhat random
	}
	sig2 := tasks.Signature{
		Name: smartcarGetInitialDataTask,
		Args: []tasks.Arg{
			{Type: "string", Value: userDeviceID},
			{Type: "string", Value: integrationID},
		},
		RetryCount: 3,
	}
	chain, err := tasks.NewChain(&sig1, &sig2)
	if err != nil {
		return
	}
	_, err = t.Machinery.SendChain(chain)
	return
}

func NewTaskService(settings *config.Settings, dbs func() *database.DBReaderWriter) *TaskService {
	var redisConn string
	if settings.RedisPassword == "" {
		redisConn = fmt.Sprintf("redis://%s:%s", settings.RedisHost, settings.RedisPort)
	} else {
		redisConn = fmt.Sprintf("redis://%s@%s:%s", settings.RedisPassword, settings.RedisHost, settings.RedisPort)
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

	pub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   strings.Split(settings.KafkaBrokers, ","),
			Marshaler: kafka.DefaultMarshaler{},
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		panic(err)
	}

	t := &TaskService{
		Settings:  settings,
		DBS:       dbs,
		Machinery: server,
		Publisher: pub,
	}
	err = server.RegisterTasks(map[string]interface{}{
		smartcarConnectVehicleTask: t.smartcarConnectVehicle,
		smartcarGetInitialDataTask: t.smartcarGetInitialData,
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
