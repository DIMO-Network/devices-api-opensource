package controllers

import (
	"io"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/aquasecurity/esquery"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
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
// @Description Get all historical data for a userDeviceID, within start and end range
// @Tags device-data
// @Produce json
// @Success 200
// @Param userDeviceID path string true "user id"
// @Param startDate query string false "startDate eg 2022-01-02"
// @Param endDate query string false "endDate eg 2022-03-01"
// @Security BearerAuth
// @Router /user/device-data/:userDeviceID/historical [get]
func (d *DeviceDataController) GetHistoricalRaw(c *fiber.Ctx) error {
	userID := getUserID(c)
	udi := c.Params("userDeviceID")
	//startDate := c.Query("start_date")
	//endDate := c.Query("end_date")
	//if endDate == "" {
	//	endDate = time.Now().Format("2006-01-02")
	//}

	// todo: cache this
	exists, err := models.UserDevices(models.UserDeviceWhere.UserID.EQ(userID), models.UserDeviceWhere.ID.EQ(udi)).Exists(c.Context(), d.DBS().Reader)
	if err != nil {
		return err
	}
	if !exists {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	// filter by udi, todo: filter by date range
	res, err := esquery.Search().
		Query(esquery.Match("subject", udi)).
		Size(50).
		Sort("data.timestamp", "desc").
		Run(d.es, d.es.Search.WithContext(c.Context()), d.es.Search.WithIndex(d.Settings.DeviceDataIndexName))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)

	c.Set("Content-type", "application/json")
	return c.Status(fiber.StatusOK).Send(b)
}

//func (d *DeviceDataController) GetTestData(c *fiber.Ctx) error {
//	udi := "22tIH4BG0vUFYpHCUyl1VrcwYwU"
//	// filter by udi, filter by date range too?
//	//res, err := esquery.Search().
//	//	Query(esquery.Bool().
//	//		Filter(esquery.Term("subject", udi))).
//	//	Run(d.es, d.es.Search.WithContext(c.Context()), d.es.Search.WithIndex(d.Settings.DeviceDataIndexName))
//	startDate := "2022-02-05"
//	endDate := "2022-02-10"
//	//				Query(esquery.Range("data.timestamp").Gte(startDate).Lte(endDate))
//
//	// todo: tweak this until get udi, start, end dates filtering
//	res, err := esquery.Search().
//		Query(esquery.Match("subject", udi)).
//		Size(50). // default is only 10, we may need to paginate
//		Sort("data.timestamp", "desc").
//		Run(d.es, d.es.Search.WithContext(c.Context()), d.es.Search.WithIndex(d.Settings.DeviceDataIndexName))
//	if err != nil {
//		return err
//	}
//	defer res.Body.Close()
//
//	return c.Status(fiber.StatusOK).SendString(res.String())
//}

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
