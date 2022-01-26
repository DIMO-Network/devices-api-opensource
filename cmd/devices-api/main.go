package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/DIMO-INC/devices-api/docs"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/controllers"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/internal/kafka"
	"github.com/DIMO-INC/devices-api/internal/services"
	"github.com/DIMO-Network/zflogger"
	"github.com/Shopify/sarama"
	"github.com/ansrivas/fiberprometheus/v2"
	swagger "github.com/arsmn/fiber-swagger/v2"
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

// @title     DIMO Devices API
// @version   2.0
// @BasePath  /v1

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
		eventService := services.NewEventService(&logger, settings)
		generateEvents(logger, settings, pdb, eventService)
	case "seed-smartcar":
		loadSmartCarData(ctx, logger, settings, pdb)
	case "edmunds-vehicles-sync":
		mergeMMYMatch := false
		if len(os.Args) > 2 {
			mergeMMYMatch = os.Args[2] == "--mergemmy"
		}
		logger.Info().Msgf("Loading edmunds vehicles for device definitions with merge MMY match: %v", mergeMMYMatch)
		err = loadEdmundsDeviceDefinitions(ctx, &logger, settings, pdb, mergeMMYMatch)
		if err != nil {
			logger.Fatal().Err(err).Msg("error trying to sync edmunds")
		}
	case "edmunds-cli-migrator":
		logger.Info().Msg("starting edmunds CLI migration tool. Recommend having your DB view open.")
		err = mergeEdmundsDefinitions(ctx, &logger, settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("error trying to run migrator tool")
		}
	case "edmunds-images":
		overwrite := false
		if len(os.Args) > 2 {
			overwrite = os.Args[2] == "--overwrite"
		}
		logger.Info().Msgf("Loading edmunds images for device definitions with overwrite: %v", overwrite)
		loadEdmundsImages(ctx, logger, settings, pdb, overwrite)
	case "remake-smartcar-topic":
		err = remakeSmartcarTopic(&logger, settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
		}
	default:
		startPrometheus(logger)
		eventService := services.NewEventService(&logger, settings)
		startDeviceStatusConsumer(logger, settings, pdb, eventService)
		startWebAPI(logger, settings, pdb, eventService)
	}
}

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, eventService services.EventService) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return ErrorHandler(c, err, logger)
		},
		DisableStartupMessage: true,
		ReadBufferSize:        16000,
	})
	nhtsaSvc := services.NewNHTSAService()
	ddSvc := services.NewDeviceDefinitionService(settings, pdb.DBS, &logger, nhtsaSvc)
	taskSvc := services.NewTaskService(settings, pdb.DBS, ddSvc, eventService, &logger)
	deviceControllers := controllers.NewDevicesController(settings, pdb.DBS, &logger, nhtsaSvc, ddSvc)
	userDeviceControllers := controllers.NewUserDevicesController(settings, pdb.DBS, &logger, ddSvc, taskSvc, eventService)

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
	v1.Get("/user/devices/me", jwtAuth, userDeviceControllers.GetUserDevices)
	v1.Post("/user/devices", jwtAuth, userDeviceControllers.RegisterDeviceForUser)
	v1.Delete("/user/devices/:userDeviceID", jwtAuth, userDeviceControllers.DeleteUserDevice)
	v1.Patch("/user/devices/:userDeviceID/vin", jwtAuth, userDeviceControllers.UpdateVIN)
	v1.Patch("/user/devices/:userDeviceID/name", jwtAuth, userDeviceControllers.UpdateName)
	v1.Patch("/user/devices/:userDeviceID/country-code", jwtAuth, userDeviceControllers.UpdateCountryCode)
	v1.Get("/user/devices/:userDeviceID/integrations/:integrationID", jwtAuth, userDeviceControllers.GetUserDeviceIntegration)
	v1.Delete("/user/devices/:userDeviceID/integrations/:integrationID", jwtAuth, userDeviceControllers.DeleteUserDeviceIntegration)
	v1.Post("/user/devices/:userDeviceID/integrations/:integrationID", jwtAuth, userDeviceControllers.RegisterSmartcarIntegration)
	v1.Get("/user/devices/:userDeviceID/status", jwtAuth, userDeviceControllers.GetUserDeviceStatus)
	v1.Post("/user/devices/:userDeviceID/commands/refresh", jwtAuth, userDeviceControllers.RefreshUserDeviceStatus)

	// admin / internal operations paths
	// v1.Post("/admin/user/:user_id/devices", userDeviceControllers.AdminRegisterUserDevice)

	// swagger - note could add auth middleware so it is not open
	sc := swagger.Config{ // custom
		// Expand ("list") or Collapse ("none") tag groups by default
		DocExpansion: "list",
	}
	v1.Get("/swagger/*", swagger.New(sc))

	logger.Info().Msg("Server started on port " + settings.Port)
	// Start Server
	if err := app.Listen(":" + settings.Port); err != nil {
		logger.Fatal().Err(err)
	}
}

// healthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func healthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
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
	logger.Err(err).Msg("caught a panic")

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
