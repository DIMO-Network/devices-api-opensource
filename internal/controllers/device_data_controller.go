package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	smartcar "github.com/smartcar/go-sdk"
	"github.com/tidwall/gjson"
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

// GetHistorical30mRaw godoc
// @Description  Get historical data for a userDeviceID, within start and end range, taking the
// @Description  latest status from every 30m bucket.
// @Tags         device-data
// @Produce      json
// @Success      200
// @Failure 	 400 "invalid start or end date"
// @Failure      404 "no device found for user with provided parameters"
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
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No device %s found for user %s", udi, userID))
	}
	req := esquery.Search().
		Query(esquery.Bool().Must(
			esquery.Term("subject", udi),
			// filter by OR latitude
			esquery.Bool().Should(
				esquery.Exists("data.odometer"),
				esquery.Exists("data.latitude"),
			),
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

	res, err := req.Run(d.es, d.es.Search.WithContext(c.Context()), d.es.Search.WithIndex(d.Settings.DeviceDataIndexName))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	c.Set("Content-type", "application/json")
	return c.Status(fiber.StatusOK).Send(body)
}

// GetHistoricalRaw godoc
// @Description  Get historical data for a userDeviceID, within start and end range, taking the
// @Description  latest status from every 30m bucket.
// @Tags         device-data
// @Produce      json
// @Success      200
// @Failure 	 400 "invalid start or end date"
// @Failure      404 "no device found for user with provided parameters"
// @Param        userDeviceID  path   string  true   "user id"
// @Param        startDate     query  string  false  "startDate eg 2022-01-02. if empty two weeks back"
// @Param        endDate       query  string  false  "endDate eg 2022-03-01. if empty today"
// @Param        signal        query  string  false  "odometer. if empty all"
// @Param        interval      query  string  false  "intervals for bucketing, 30m, 60m. if empty 30m"
// @Security     BearerAuth
// @Router       /user/device-data/{userDeviceID}/historical [get]
func (d *DeviceDataController) GetHistoricalRaw(c *fiber.Ctx) error {
	const dateLayout = "2006-01-02" // date layout support by elastic
	userID := getUserID(c)
	udi := c.Params("userDeviceID")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	signalName := c.Query("signal") // device signal to get from the data property in elastic
	interval := c.Query("interval") // bucket interval aggregation

	if startDate == "" {
		startDate = time.Now().Add(-1 * (time.Hour * 24 * 14)).Format(dateLayout)
	} else {
		_, err := time.Parse(dateLayout, startDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	if endDate == "" {
		endDate = time.Now().Format(dateLayout)
	} else {
		_, err := time.Parse(dateLayout, endDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	if stringContainsSpecialChars(signalName) {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("You cannot query for field %s as user %s", signalName, userID))
	}
	if interval == "" {
		interval = "30m"
	}

	// todo: cache this
	exists, err := models.UserDevices(models.UserDeviceWhere.UserID.EQ(userID), models.UserDeviceWhere.ID.EQ(udi)).Exists(c.Context(), d.DBS().Reader)
	if err != nil {
		return err
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No device %s found for user %s", udi, userID))
	}
	req := esquery.Search().
		Query(esquery.Bool().Must(
			esquery.Term("subject", udi),
			esquery.Bool().Should(
				esquery.Exists("data.odometer"),
				esquery.Exists("data.latitude"),
			),
			esquery.Range("data.timestamp").
				Gte(startDate).
				Lte(endDate))).
		Size(0).
		Aggs(
			esquery.CustomAgg("buckets_30m", map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          "data.timestamp",
					"fixed_interval": interval,
					"min_doc_count":  1,
				},
				"aggs": map[string]interface{}{
					"last_status": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": []map[string]string{
								{"data.timestamp": "desc"},
							},
							"_source": getFieldQuery(signalName),
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

// GetDistanceDriven godoc
// @Description  Get kilometers driven for a userDeviceID since connected (ie. since we have data available)
// @Description  if it returns 0 for distanceDriven it means we have no odometer data.
// @Tags         device-data
// @Produce      json
// @Success      200
// @Failure      404 "no device found for user with provided parameters"
// @Param        userDeviceID  path   string  true   "user device id"
// @Security     BearerAuth
// @Router       /user/device-data/{userDeviceID}/distance-driven [get]
func (d *DeviceDataController) GetDistanceDriven(c *fiber.Ctx) error {
	userID := getUserID(c)
	udi := c.Params("userDeviceID")

	exists, err := models.UserDevices(models.UserDeviceWhere.UserID.EQ(userID), models.UserDeviceWhere.ID.EQ(udi)).Exists(c.Context(), d.DBS().Reader)
	if err != nil {
		return err
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("No device %s found for user %s", udi, userID))
	}

	odoStart, err := d.queryOdometer(c.Context(), esquery.OrderAsc, udi)
	if err != nil {
		return errors.Wrap(err, "error querying odometer")
	}
	odoEnd, err := d.queryOdometer(c.Context(), esquery.OrderDesc, udi)
	if err != nil {
		return errors.Wrap(err, "error querying odometer")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"distanceDriven": odoEnd - odoStart,
		"units":          "kilometers",
	})
}

// GetUserDeviceStatus godoc
// @Description  Returns the latest status update for the device. May return 404 if the
// @Description  user does not have a device with the ID, or if no status updates have come
// @Tags         user-devices
// @Produce      json
// @Param        user_device_id  path  string  true  "user device ID"
// @Success      200 {object} controllers.DeviceSnapshot
// @Security     BearerAuth
// @Router       /user/devices/{userDeviceID}/status [get]
func (udc *UserDevicesController) GetUserDeviceStatus(c *fiber.Ctx) error {
	udi := c.Params("userDeviceID")
	userID := getUserID(c)
	userDevice, err := models.UserDevices(
		models.UserDeviceWhere.ID.EQ(udi),
		models.UserDeviceWhere.UserID.EQ(userID),
	).One(c.Context(), udc.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}
	deviceData, err := models.UserDeviceData(models.UserDeviceDatumWhere.UserDeviceID.EQ(userDevice.ID),
		qm.OrderBy("updated_at asc")).All(c.Context(), udc.DBS().Reader)
	if errors.Is(err, sql.ErrNoRows) || len(deviceData) == 0 || !deviceData[0].Data.Valid {
		return fiber.NewError(fiber.StatusNotFound, "no status updates yet")
	}
	if err != nil {
		return err
	}
	// how should we handle the errorData, if at all?
	ds := DeviceSnapshot{}
	// merging data: foreach order by updatedAt desc, only set property if it exists in json data
	for _, datum := range deviceData {
		if datum.Data.Valid {
			// note this means the date updated we sent may be inaccurate if have both smartcar and autopi
			if ds.RecordUpdatedAt == nil {
				ds.RecordCreatedAt = &datum.CreatedAt
				ds.RecordUpdatedAt = &datum.UpdatedAt
			}
			// note we are assuming json property names are same accross smartcar, tesla, autopi, AND same types eg. int / float / string
			// we could use reflection and just have single line assuming json name in struct matches what is in data
			charging := gjson.GetBytes(datum.Data.JSON, "charging")
			if charging.Exists() {
				c := charging.Bool()
				ds.Charging = &c
			}
			fuelPercentRemaining := gjson.GetBytes(datum.Data.JSON, "fuelPercentRemaining")
			if fuelPercentRemaining.Exists() {
				f := fuelPercentRemaining.Float()
				ds.FuelPercentRemaining = &f
			}
			batteryCapacity := gjson.GetBytes(datum.Data.JSON, "batteryCapacity")
			if batteryCapacity.Exists() {
				b := batteryCapacity.Int()
				ds.BatteryCapacity = &b
			}
			oilLevel := gjson.GetBytes(datum.Data.JSON, "oil")
			if oilLevel.Exists() {
				o := oilLevel.Float()
				ds.OilLevel = &o
			}
			odometer := gjson.GetBytes(datum.Data.JSON, "odometer")
			if odometer.Exists() {
				o := odometer.Float()
				ds.Odometer = &o
			}
			latitude := gjson.GetBytes(datum.Data.JSON, "latitude")
			if latitude.Exists() {
				l := latitude.Float()
				ds.Latitude = &l
			}
			longitude := gjson.GetBytes(datum.Data.JSON, "longitude")
			if longitude.Exists() {
				l := longitude.Float()
				ds.Longitude = &l
			}
			rangeG := gjson.GetBytes(datum.Data.JSON, "range")
			if rangeG.Exists() {
				r := rangeG.Float()
				ds.Range = &r
			}
			// TirePressure
			tires := gjson.GetBytes(datum.Data.JSON, "tires")
			if tires.Exists() {
				// weird thing here is in example payloads these are all ints, but the smartcar lib has as floats
				ds.TirePressure = &smartcar.TirePressure{
					FrontLeft:  tires.Get("frontLeft").Float(),
					FrontRight: tires.Get("frontRight").Float(),
					BackLeft:   tires.Get("backLeft").Float(),
					BackRight:  tires.Get("backRight").Float(),
				}
			}
		}
	}
	return c.JSON(ds)
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
		qm.Load(models.UserDeviceRels.UserDeviceData),
		qm.Load(qm.Rels(models.UserDeviceRels.UserDeviceData, models.UserDeviceDatumRels.Integration)),
	).One(c.Context(), udc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}

	for _, deviceDatum := range ud.R.UserDeviceData {
		if deviceDatum.R.Integration.Type == models.IntegrationTypeAPI && deviceDatum.R.Integration.Vendor == services.SmartCarVendor {

			nextAvailableTime := deviceDatum.UpdatedAt.Add(time.Second * time.Duration(deviceDatum.R.Integration.RefreshLimitSecs))
			if time.Now().Before(nextAvailableTime) {
				return fiber.NewError(fiber.StatusTooManyRequests, "rate limit for integration refresh hit")
			}

			udai, err := models.FindUserDeviceAPIIntegration(c.Context(), udc.DBS().Reader, deviceDatum.UserDeviceID, deviceDatum.IntegrationID)
			if err != nil {
				return err
			}
			if udai.TaskID.Valid {
				err = udc.smartcarTaskSvc.Refresh(udai)
				if err != nil {
					return err
				}
			} else {
				err = udc.taskSvc.StartSmartcarRefresh(udi, deviceDatum.IntegrationID)
				if err != nil {
					return err
				}
			}
			return c.SendStatus(204)
		}
	}
	return fiber.NewError(fiber.StatusBadRequest, "no active Smartcar integration found for this device")
}

// queryOdometer gets the first or last odometer reading depending on order - asc = first, desc = last
func (d *DeviceDataController) queryOdometer(ctx context.Context, order esquery.Order, userDeviceID string) (float64, error) {
	res, err := esquery.Search().SourceIncludes("data.odometer").
		Query(esquery.Bool().Must(
			esquery.Term("subject", userDeviceID),
			esquery.Exists("data.odometer"),
		)).
		Size(1).
		Sort("data.timestamp", order).
		Run(d.es, d.es.Search.WithContext(ctx), d.es.Search.WithIndex(d.Settings.DeviceDataIndexName))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close() // nolint
	body, _ := io.ReadAll(res.Body)
	result := gjson.GetBytes(body, "hits.hits.1._source.data.odometer")
	if result.Exists() {
		return result.Float(), nil
	}
	return 0, nil
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

func getFieldQuery(signalName string) map[string][]string {
	if signalName == "" {
		return map[string][]string{
			"excludes": {"data.errors", "data.hasErrors"},
		}
	}
	return map[string][]string{
		"includes": {"data." + signalName},
	}
}

func stringContainsSpecialChars(str string) bool {
	for _, charVariable := range str {
		if (charVariable < 'a' || charVariable > 'z') && (charVariable < 'A' || charVariable > 'Z') {
			return true
		}
	}
	return false
}

// DeviceSnapshot is the response object for device status endpoint
// https://docs.google.com/document/d/1DYzzTOR9WA6WJNoBnwpKOoxfmrVwPWNLv0x0MkjIAqY/edit#heading=h.dnp7xngl47bw
type DeviceSnapshot struct {
	Charging             *bool                  `json:"charging,omitempty"`
	FuelPercentRemaining *float64               `json:"fuelPercentRemaining,omitempty"`
	BatteryCapacity      *int64                 `json:"batteryCapacity,omitempty"`
	OilLevel             *float64               `json:"oil,omitempty"`
	Odometer             *float64               `json:"odometer,omitempty"`
	Latitude             *float64               `json:"latitude,omitempty"`
	Longitude            *float64               `json:"longitude,omitempty"`
	Range                *float64               `json:"range,omitempty"`
	RecordUpdatedAt      *time.Time             `json:"recordUpdatedAt,omitempty"`
	RecordCreatedAt      *time.Time             `json:"recordCreatedAt,omitempty"`
	TirePressure         *smartcar.TirePressure `json:"tirePressure,omitempty"`
}
