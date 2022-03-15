package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/DIMO-Network/devices-api/docs"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/kafka"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/zflogger"
	"github.com/Jeffail/benthos/v3/lib/util/hash/murmur2"
	"github.com/Shopify/sarama"
	"github.com/ansrivas/fiberprometheus/v2"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	_ "go.uber.org/automaxprocs"
)

// @title                       DIMO Devices API
// @version                     1.0
// @BasePath                    /v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Str("git-sha1", gitSha1).
		Logger()

	config.SetupMachineryLogging(&logger)

	settings, err := config.LoadConfig("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load settings")
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		logger.Fatal().Err(err).Msgf("could not parse LOG_LEVEL: %s", settings.LogLevel)
	}
	zerolog.SetGlobalLevel(level)

	pdb := database.NewDbConnectionFromSettings(ctx, settings, true)
	// check db ready, this is not ideal btw, the db connection handler would be nicer if it did this.
	totalTime := 0
	for !pdb.IsReady() {
		if totalTime > 30 {
			logger.Fatal().Msg("could not connect to postgres after 30 seconds")
		}
		time.Sleep(time.Second)
		totalTime++
	}

	producer, err := createKafkaProducer(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not initialize Kafka producer, terminating")
	}

	// todo: use flag or other package to handle args
	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {
	case "migrate":
		command := "up"
		if len(os.Args) > 2 {
			command = os.Args[2]
			if command == "down-to" || command == "up-to" {
				command = command + " " + os.Args[3]
			}
		}
		migrateDatabase(logger, settings, command)
	case "generate-events":
		eventService := services.NewEventService(&logger, settings, producer)
		generateEvents(logger, settings, pdb, eventService)
	case "smartcar-sync":
		syncSmartCarCompatibility(ctx, logger, settings, pdb)
	case "create-tesla-integrations":
		if err := createTeslaIntegrations(ctx, pdb, &logger); err != nil {
			logger.Fatal().Err(err).Msg("Failed to create Tesla integrations")
		}
	case "edmunds-vehicles-sync":
		logger.Info().Msgf("Loading edmunds vehicles for device definitions and merging MMYs")
		err = loadEdmundsDeviceDefinitions(ctx, &logger, settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("error trying to sync edmunds")
		}
	case "parkers-vehicles-sync":
		err = loadParkersDeviceDefinitions(ctx, &logger, settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error syncing with Parkers")
		}
	case "smartcar-compatibility":
		logger.Info().Msg("starting smartcar compatibility equalizer check to set smartcar compat forwards")
		err = smartCarForwardCompatibility(ctx, logger, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("error trying to run smartcar forwards compatibility")
		}
	case "edmunds-images":
		overwrite := false
		if len(os.Args) > 2 {
			overwrite = os.Args[2] == "--overwrite"
		}
		logger.Info().Msgf("Loading edmunds images for device definitions with overwrite: %v", overwrite)
		loadEdmundsImages(ctx, logger, settings, pdb, overwrite)
	case "remake-smartcar-topic":
		err = remakeSmartcarTopic(ctx, &logger, settings, pdb, producer)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
		}
	case "remake-fence-topic":
		err = remakeFenceTopic(&logger, settings, pdb, producer)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
		}
	case "migrate-smartcar-webhooks":
		if len(os.Args[1:]) != 2 {
			logger.Fatal().Msgf("Expected two arguments, but got %d", len(os.Args[1:]))
		}
		oldWebhookID := os.Args[2]
		err = migrateSmartcarWebhooks(ctx, &logger, settings, pdb, oldWebhookID)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running Smartcar webhook migration")
		}
	case "refresh-smartcar-tokens":
		err = refreshSmartcarTokens(ctx, &logger, settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running Smartcar webhook migration")
		}
	case "search-sync-dds":
		logger.Info().Msg("loading device definitions from our DB to elastic cluster")
		err := loadElasticDevices(ctx, &logger, settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("error syncing with elastic")
		}
	default:
		startPrometheus(logger)
		eventService := services.NewEventService(&logger, settings, producer)
		startDeviceStatusConsumer(logger, settings, pdb, eventService)
		startWebAPI(logger, settings, pdb, eventService, producer)
	}
}

func createKafkaProducer(settings *config.Settings) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_1_0
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewCustomPartitioner(
		sarama.WithAbsFirst(),
		sarama.WithCustomHashFunction(murmur2.New32),
	)
	p, err := sarama.NewSyncProducer(strings.Split(settings.KafkaBrokers, ","), config)
	if err != nil {
		return nil, fmt.Errorf("failed to construct producer with broker list %s: %w", settings.KafkaBrokers, err)
	}
	return p, nil
}

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, eventService services.EventService, producer sarama.SyncProducer) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return ErrorHandler(c, err, logger)
		},
		DisableStartupMessage: true,
		ReadBufferSize:        16000,
	})

	var encrypter services.Encrypter
	if settings.Environment == "dev" || settings.Environment == "prod" {
		encrypter = createKMS(settings, &logger)
	} else {
		logger.Warn().Msg("Using ROT13 encrypter. Only use this for testing!")
		encrypter = new(services.ROT13Encrypter)
	}

	nhtsaSvc := services.NewNHTSAService()
	ddSvc := services.NewDeviceDefinitionService(settings.TorProxyURL, pdb.DBS, &logger, nhtsaSvc)
	smartCarSvc := services.NewSmartCarService(pdb.DBS, logger)
	smartcarClient := services.NewSmartcarClient(settings)
	teslaTaskService := services.NewTeslaTaskService(settings, producer)
	teslaSvc := services.NewTeslaService(settings)
	taskSvc := services.NewTaskService(settings, pdb.DBS, ddSvc, eventService, &logger, producer, &smartCarSvc)
	deviceControllers := controllers.NewDevicesController(settings, pdb.DBS, &logger, nhtsaSvc, ddSvc)
	userDeviceControllers := controllers.NewUserDevicesController(settings, pdb.DBS, &logger, ddSvc, taskSvc, eventService, smartcarClient, teslaSvc, teslaTaskService, encrypter)
	geofenceController := controllers.NewGeofencesController(settings, pdb.DBS, &logger, producer)
	deviceDataController := controllers.NewDeviceDataController(settings, pdb.DBS, &logger)

	prometheus := fiberprometheus.New("devices-api")
	app.Use(prometheus.Middleware)

	app.Use(recover.New(recover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: nil,
	}))
	//cors
	app.Use(cors.New())
	// request logging
	app.Use(zflogger.New(logger, nil))
	//cache
	cacheHandler := cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   1 * time.Minute,
		CacheControl: true,
	})
	// application routes
	app.Get("/", healthCheck)
	app.Put("/loglevel", changeLogLevel)

	v1 := app.Group("/v1")
	sc := swagger.Config{ // custom
		// Expand ("list") or Collapse ("none") tag groups by default
		//DocExpansion: "list",
	}
	v1.Get("/swagger/*", swagger.New(sc))
	// Device Definitions
	v1.Get("/device-definitions/all", cacheHandler, deviceControllers.GetAllDeviceMakeModelYears)
	v1.Get("/device-definitions/:id", deviceControllers.GetDeviceDefinitionByID)
	v1.Get("/device-definitions/:id/integrations", deviceControllers.GetIntegrationsByID)
	v1.Get("/device-definitions", deviceControllers.GetDeviceDefinitionByMMY)

	// secured paths
	keyRefreshInterval := time.Hour
	keyRefreshUnknownKID := true
	jwtAuth := jwtware.New(jwtware.Config{
		KeySetURL:            settings.JwtKeySetURL,
		KeyRefreshInterval:   &keyRefreshInterval,
		KeyRefreshUnknownKID: &keyRefreshUnknownKID,
	})
	v1Auth := app.Group("/v1", jwtAuth)
	// user's devices
	v1Auth.Get("/user/devices/me", userDeviceControllers.GetUserDevices)
	v1Auth.Post("/user/devices", userDeviceControllers.RegisterDeviceForUser)
	v1Auth.Delete("/user/devices/:userDeviceID", userDeviceControllers.DeleteUserDevice)
	v1Auth.Patch("/user/devices/:userDeviceID/vin", userDeviceControllers.UpdateVIN)
	v1Auth.Patch("/user/devices/:userDeviceID/name", userDeviceControllers.UpdateName)
	v1Auth.Patch("/user/devices/:userDeviceID/country-code", userDeviceControllers.UpdateCountryCode)
	v1Auth.Get("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceControllers.GetUserDeviceIntegration)
	v1Auth.Delete("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceControllers.DeleteUserDeviceIntegration)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceControllers.RegisterDeviceIntegration)
	v1Auth.Get("/user/devices/:userDeviceID/status", userDeviceControllers.GetUserDeviceStatus)
	v1Auth.Post("/user/devices/:userDeviceID/commands/refresh", userDeviceControllers.RefreshUserDeviceStatus)

	// geofence
	v1Auth.Post("/user/geofences", geofenceController.Create)
	v1Auth.Get("/user/geofences", geofenceController.GetAll)
	v1Auth.Delete("/user/geofences/:geofenceID", geofenceController.Delete)
	v1Auth.Put("/user/geofences/:geofenceID", geofenceController.Update)

	// elastic device data
	v1Auth.Get("/user/device-data/:userDeviceID/historical", deviceDataController.GetHistoricalRaw)
	v1Auth.Get("/user/device-data/:userDeviceID/historical-30m", deviceDataController.GetHistorical30mRaw)
	v1Auth.Get("/user/device-data/:userDeviceID/distance-driven", deviceDataController.GetDistanceDriven)

	// admin / internal operations paths
	// v1.Post("/admin/user/:user_id/devices", userDeviceControllers.AdminRegisterUserDevice)

	logger.Info().Msg("Server started on port " + settings.Port)
	// Start Server from a different go routine
	go func() {
		if err := app.Listen(":" + settings.Port); err != nil {
			logger.Fatal().Err(err)
		}
	}()
	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-c                                             // This blocks the main thread until an interrupt is received
	logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
	_ = app.Shutdown()
	_ = pdb.DBS().Writer.Close()
	_ = pdb.DBS().Reader.Close()
	_ = producer.Close()
}

func healthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}

func createKMS(settings *config.Settings, logger *zerolog.Logger) services.Encrypter {
	// Need AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY to be set.
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(settings.AWSRegion),
	}))

	return &services.KMSEncrypter{
		KeyID:  settings.KMSKeyID,
		Client: kms.New(sess),
	}
}

func changeLogLevel(c *fiber.Ctx) error {
	payload := struct {
		LogLevel string `json:"logLevel"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return err
	}
	level, err := zerolog.ParseLevel(payload.LogLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(level)
	return c.Status(fiber.StatusOK).SendString("log level set to: " + level.String())
}

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger zerolog.Logger) error {
	code := fiber.StatusInternalServerError // Default 500 statuscode

	if e, ok := err.(*fiber.Error); ok {
		// Override status code if fiber.Error type
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	logger.Err(err).Msg("caught an error")

	return c.Status(code).JSON(fiber.Map{
		"error": true,
		"msg":   err.Error(),
	})
}

func startDeviceStatusConsumer(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, eventService services.EventService) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_6_0_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.DeviceStatusTopic,
		GroupID:         "user-devices",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not start consumer")
	}
	ingestSvc := services.NewIngestService(pdb.DBS, &logger, eventService)
	consumer.Start(context.Background(), ingestSvc.ProcessDeviceStatusMessages)

	logger.Info().Msg("kafka consumer started")
}

func startPrometheus(logger zerolog.Logger) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":8888", nil)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not start consumer")
		}
	}()
	logger.Info().Msg("prometheus metrics at :8888/metrics")
}
