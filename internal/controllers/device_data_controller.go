package controllers

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/aquasecurity/esquery"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/sjson"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DeviceDataController struct {
	Settings *config.Settings
	DBS      func() *database.DBReaderWriter
	log      *zerolog.Logger
	es       *elasticsearch.Client
}

// NewDeviceDataController constructor
func NewDeviceDataController(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) DeviceDataController {
	es, err := connect(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not connect to elastic search")
	}
	return DeviceDataController{
		Settings: settings,
		DBS:      dbs,
		log:      logger,
		es:       es,
	}
}

// GetHistoricalRaw godoc
// @Description  Get all historical data for a userDeviceID, within start and end range
// @Tags         device-data
// @Produce      json
// @Success      200
// @Param        userDeviceID  path   string  true   "user id"
// @Param        startDate     query  string  false  "startDate eg 2022-01-02. if empty two weeks back"
// @Param        endDate       query  string  false  "endDate eg 2022-03-01. if empty today"
// @Security     BearerAuth
// @Router       /user/device-data/{userDeviceID}/historical [get]
func (d *DeviceDataController) GetHistoricalRaw(c *fiber.Ctx) error {
	const dateLayout = "2006-01-02" // date layout support by elastic
	userID := getUserID(c)
	udi := c.Params("userDeviceID")
	startDate := c.Query("startDate")
	if startDate == "" {
		startDate = time.Now().Add(-1 * (time.Hour * 24 * 14)).Format(dateLayout)
	} else {
		_, err := time.Parse(dateLayout, startDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	endDate := c.Query("endDate")
	if endDate == "" {
		endDate = time.Now().Format(dateLayout)
	} else {
		_, err := time.Parse(dateLayout, endDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	// todo: cache this
	exists, err := models.UserDevices(models.UserDeviceWhere.UserID.EQ(userID), models.UserDeviceWhere.ID.EQ(udi)).Exists(c.Context(), d.DBS().Reader)
	if err != nil {
		return err
	}
	if !exists {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	res, err := esquery.Search().
		Query(esquery.Bool().Must(
			esquery.Term("subject", udi),
			esquery.Exists("data.odometer"),
			esquery.Range("data.timestamp").
				Gte(startDate).
				Lte(endDate))).
		Size(50).
		Sort("data.timestamp", "desc").
		Run(d.es, d.es.Search.WithContext(c.Context()), d.es.Search.WithIndex(d.Settings.DeviceDataIndexName))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	c.Set("Content-type", "application/json")
	return c.Status(fiber.StatusOK).Send(body)
}

// GetHistorical30mRaw godoc
// @Description  Get historical data for a userDeviceID, within start and end range, taking the
// @Description  latest status from every 30m bucket.
// @Tags         device-data
// @Produce      json
// @Success      200
// @Param        userDeviceID  path   string  true   "user id"
// @Param        startDate     query  string  false  "startDate eg 2022-01-02. if empty two weeks back"
// @Param        endDate       query  string  false  "endDate eg 2022-03-01. if empty today"
// @Security     BearerAuth
// @Router       /user/device-data/{userDeviceID}/historical-30m [get]
func (d *DeviceDataController) GetHistorical30mRaw(c *fiber.Ctx) error {
	const dateLayout = "2006-01-02" // date layout support by elastic
	userID := getUserID(c)
	udi := c.Params("userDeviceID")
	startDate := c.Query("startDate")
	if startDate == "" {
		startDate = time.Now().Add(-1 * (time.Hour * 24 * 14)).Format(dateLayout)
	} else {
		_, err := time.Parse(dateLayout, startDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	endDate := c.Query("endDate")
	if endDate == "" {
		endDate = time.Now().Format(dateLayout)
	} else {
		_, err := time.Parse(dateLayout, endDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	// todo: cache this
	exists, err := models.UserDevices(models.UserDeviceWhere.UserID.EQ(userID), models.UserDeviceWhere.ID.EQ(udi)).Exists(c.Context(), d.DBS().Reader)
	if err != nil {
		return err
	}
	if !exists {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	req := esquery.Search().
		Query(esquery.Bool().Must(
			esquery.Term("subject", udi),
			esquery.Exists("data.odometer"),
			esquery.Range("data.timestamp").
				Gte(startDate).
				Lte(endDate))).
		Size(0).
		Aggs(
			esquery.CustomAgg("buckets_30m", map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          "data.timestamp",
					"fixed_interval": "30m",
					"min_doc_count":  1,
				},
				"aggs": map[string]interface{}{
					"last_status": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": []map[string]string{
								{"data.timestamp": "desc"},
							},
							"_source": map[string][]string{
								"excludes": {"data.errors", "data.hasErrors"},
							},
						},
					},
				},
			}),
		)
	reqb, err := json.Marshal(req.Map())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "We messed up")
	}
	d.log.Info().Msg(string(reqb))
	res, err := req.Run(d.es, d.es.Search.WithContext(c.Context()), d.es.Search.WithIndex(d.Settings.DeviceDataIndexName))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	c.Set("Content-type", "application/json")
	return c.Status(fiber.StatusOK).Send(body)
}

// GetUserDeviceStatus godoc
// @Description  Returns the latest status update for the device. May return 404 if the
// @Description  user does not have a device with the ID, or if no status updates have come
// @Tags         user-devices
// @Produce      json
// @Param        user_device_id  path  string  true  "user device ID"
// @Success      200
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/status [get]
func (udc *UserDevicesController) GetUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceDatum),
	).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}

	if userDevice.R.UserDeviceDatum == nil || !userDevice.R.UserDeviceDatum.Data.Valid {
		return fiber.NewError(fiber.StatusNotFound, "no status updates yet")
	}
	// date formatting defaults to encoding/json
	json, _ := sjson.Set(string(userDevice.R.UserDeviceDatum.Data.JSON), "recordUpdatedAt", userDevice.R.UserDeviceDatum.UpdatedAt)
	json, _ = sjson.Set(json, "recordCreatedAt", userDevice.R.UserDeviceDatum.CreatedAt)
	if userDevice.R.UserDeviceDatum.ErrorData.Valid {
		json, _ = sjson.Set(json, "errorData", userDevice.R.UserDeviceDatum.ErrorData)
	}

	c.Set("Content-Type", "application/json")
	return c.Send([]byte(json))
}

// RefreshUserDeviceStatus godoc
// @Description  Starts the process of refreshing device status from Smartcar
// @Tags         user-devices
// @Param        user_device_id  path  string  true  "user device ID"
// @Success      204
// @Failure      429  "rate limit hit for integration"
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/commands/refresh [post]
func (udc *UserDevicesController) RefreshUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	// We could probably do a smarter join here, but it's unclear to me how to handle that
	// in SQLBoiler.
	ud, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations),
		qm.Load(models.UserDeviceRels.UserDeviceDatum),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationRels.Integration)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}
	// note: the UserDeviceDatum is not tied to the integration table

	for _, devInteg := range ud.R.UserDeviceAPIIntegrations {
		if devInteg.R.Integration.Type == models.IntegrationTypeAPI && devInteg.R.Integration.Vendor == services.SmartCarVendor && devInteg.Status == models.UserDeviceAPIIntegrationStatusActive {
			if ud.R.UserDeviceDatum != nil {
				nextAvailableTime := ud.R.UserDeviceDatum.UpdatedAt.Add(time.Second * time.Duration(devInteg.R.Integration.RefreshLimitSecs))
				if time.Now().Before(nextAvailableTime) {
					return fiber.NewError(fiber.StatusTooManyRequests, "rate limit for integration refresh hit")
				}
			}
			err = udc.taskSvc.StartSmartcarRefresh(udi, devInteg.R.Integration.ID)
			if err != nil {
				return err
			}
			return c.SendStatus(204)
		}
	}
	return fiber.NewError(fiber.StatusBadRequest, "no active Smartcar integration found for this device")
}

func connect(settings *config.Settings) (*elasticsearch.Client, error) {
	// maybe refactor some of this into elasticsearchservice

	if settings.ElasticSearchAnalyticsUsername == "" {
		// we're connecting to local instance at localhost:9200
		return elasticsearch.NewDefaultClient()
	}

	return elasticsearch.NewClient(elasticsearch.Config{
		Addresses:            []string{settings.ElasticSearchAnalyticsHost},
		Username:             settings.ElasticSearchAnalyticsUsername,
		Password:             settings.ElasticSearchAnalyticsPassword,
		EnableRetryOnTimeout: true,
		MaxRetries:           5,
	})
}
