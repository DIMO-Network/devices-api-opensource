package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DIMO-Network/devices-api/internal/api"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/controllers"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"
	pb "github.com/DIMO-Network/shared/api/devices"
	"github.com/DIMO-Network/zflogger"
	"github.com/Shopify/sarama"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func startWebAPI(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, eventService services.EventService, producer sarama.SyncProducer, s3ServiceClient *s3.Client, s3NFTServiceClient *s3.Client) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return ErrorHandler(c, err, logger, settings.Environment)
		},
		DisableStartupMessage: true,
		ReadBufferSize:        16000,
		BodyLimit:             10 * 1024 * 1024,
	})

	var cipher shared.Cipher
	if settings.Environment == "dev" || settings.Environment == "prod" {
		cipher = createKMS(settings, &logger)
	} else {
		logger.Warn().Msg("Using ROT13 encrypter. Only use this for testing!")
		cipher = new(shared.ROT13Cipher)
	}

	// services
	nhtsaSvc := services.NewNHTSAService()
	ddSvc := services.NewDeviceDefinitionService(settings.TorProxyURL, pdb.DBS, &logger, nhtsaSvc)
	scTaskSvc := services.NewSmartcarTaskService(settings, producer)
	smartcarClient := services.NewSmartcarClient(settings)
	teslaTaskService := services.NewTeslaTaskService(settings, producer)
	teslaSvc := services.NewTeslaService(settings)
	autoPiSvc := services.NewAutoPiAPIService(settings, pdb.DBS)
	autoPiIngest := services.NewIngestRegistrar(services.AutoPi, producer)
	autoPiTaskService := services.NewAutoPiTaskService(settings, autoPiSvc, logger)
	// controllers
	deviceControllers := controllers.NewDevicesController(settings, pdb.DBS, &logger, nhtsaSvc, ddSvc)
	userDeviceController := controllers.NewUserDevicesController(settings, pdb.DBS, &logger, ddSvc, eventService, smartcarClient, scTaskSvc, teslaSvc, teslaTaskService, cipher, autoPiSvc, services.NewNHTSAService(), autoPiIngest, autoPiTaskService, producer, s3NFTServiceClient)
	geofenceController := controllers.NewGeofencesController(settings, pdb.DBS, &logger, producer)
	webhooksController := controllers.NewWebhooksController(settings, pdb.DBS, &logger, autoPiSvc)
	documentsController := controllers.NewDocumentsController(settings, s3ServiceClient, pdb.DBS)
	// commenting this out b/c the library includes the path in the metrics which saturates prometheus queries - need to fork / make our own
	//prometheus := fiberprometheus.New("devices-api")
	//app.Use(prometheus.Middleware)

	app.Use(fiberrecover.New(fiberrecover.Config{
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

	v1 := app.Group("/v1")
	sc := swagger.Config{ // custom
		// Expand ("list") or Collapse ("none") tag groups by default
		//DocExpansion: "list",
	}
	v1.Get("/swagger/*", swagger.New(sc))
	// Device Definitions
	v1.Get("/device-definitions/all", cacheHandler, deviceControllers.GetAllDeviceMakeModelYears)
	v1.Get("/device-definitions/:id", deviceControllers.GetDeviceDefinitionByID)
	v1.Get("/device-definitions/:id/integrations", deviceControllers.GetDeviceIntegrationsByID)
	v1.Get("/device-definitions", deviceControllers.GetDeviceDefinitionByMMY)

	nftController := controllers.NewNFTController(settings, pdb.DBS, &logger, s3NFTServiceClient)
	v1.Get("/nfts/:tokenID", nftController.GetNFTMetadata)
	v1.Get("/nfts/:tokenID/image", nftController.GetNFTImage)

	// webhooks, performs signature validation
	v1.Post(services.AutoPiWebhookPath, webhooksController.ProcessCommand)

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
	v1Auth.Get("/user/devices/me", userDeviceController.GetUserDevices)
	v1Auth.Post("/user/devices", userDeviceController.RegisterDeviceForUser)
	v1Auth.Delete("/user/devices/:userDeviceID", userDeviceController.DeleteUserDevice)
	v1Auth.Patch("/user/devices/:userDeviceID/vin", userDeviceController.UpdateVIN)
	v1Auth.Patch("/user/devices/:userDeviceID/name", userDeviceController.UpdateName)
	v1Auth.Patch("/user/devices/:userDeviceID/country-code", userDeviceController.UpdateCountryCode)
	v1Auth.Patch("/user/devices/:userDeviceID/image", userDeviceController.UpdateImage)
	// device integrations
	v1Auth.Get("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceController.GetUserDeviceIntegration)
	v1Auth.Delete("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceController.DeleteUserDeviceIntegration)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID", userDeviceController.RegisterDeviceIntegration)
	v1Auth.Get("/user/devices/:userDeviceID/status", userDeviceController.GetUserDeviceStatus)
	v1Auth.Post("/user/devices/:userDeviceID/commands/refresh", userDeviceController.RefreshUserDeviceStatus)

	// Device commands.
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/doors/unlock", userDeviceController.UnlockDoors)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/doors/lock", userDeviceController.LockDoors)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/trunk/open", userDeviceController.OpenTrunk)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/frunk/open", userDeviceController.OpenFrunk)
	v1Auth.Post("/user/devices/:userDeviceID/integrations/:integrationID/commands/charge/limit", userDeviceController.SetChargeLimit)

	// Device NFT.
	v1Auth.Get("/user/devices/:userDeviceID/commands/mint", userDeviceController.GetMintDataToSign)
	v1Auth.Post("/user/devices/:userDeviceID/commands/mint", userDeviceController.MintDevice)

	v1Auth.Get("/integrations", userDeviceController.GetIntegrations)
	// autopi specific
	v1Auth.Post("/user/devices/:userDeviceID/autopi/command", userDeviceController.SendAutoPiCommand)
	v1Auth.Get("/user/devices/:userDeviceID/autopi/command/:jobID", userDeviceController.GetAutoPiCommandStatus)
	v1Auth.Get("/autopi/unit/:unitID", userDeviceController.GetAutoPiUnitInfo)
	v1Auth.Get("/autopi/unit/:unitID/is-online", userDeviceController.GetIsAutoPiOnline)
	// delete below line once confirmed no active apps using it.
	v1Auth.Get("/autopi/unit/is-online/:unitID", userDeviceController.GetIsAutoPiOnline) // this one is deprecated
	v1Auth.Post("/autopi/unit/:unitID/update", userDeviceController.StartAutoPiUpdateTask)
	v1Auth.Get("/autopi/task/:taskID", userDeviceController.GetAutoPiTask)

	// geofence
	v1Auth.Post("/user/geofences", geofenceController.Create)
	v1Auth.Get("/user/geofences", geofenceController.GetAll)
	v1Auth.Delete("/user/geofences/:geofenceID", geofenceController.Delete)
	v1Auth.Put("/user/geofences/:geofenceID", geofenceController.Update)

	// documents
	v1Auth.Get("/documents", documentsController.GetDocuments)
	v1Auth.Get("/documents/:id", documentsController.GetDocumentByID)
	v1Auth.Post("/documents", documentsController.PostDocument)
	v1Auth.Delete("/documents/:id", documentsController.DeleteDocument)
	v1Auth.Get("/documents/:id/download", documentsController.DownloadDocument)

	go startGRPCServer(settings, pdb.DBS, &logger)

	logger.Info().Msg("Server started on port " + settings.Port)
	// Start Server from a different go routine
	go func() {
		if err := app.Listen(":" + settings.Port); err != nil {
			logger.Fatal().Err(err)
		}
	}()
	// start task consumer for autopi
	ctx := context.Background()
	autoPiTaskService.StartConsumer(ctx)

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-c                                             // This blocks the main thread until an interrupt is received
	logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
	_ = ctx.Done()
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

func startGRPCServer(settings *config.Settings, dbs func() *database.DBReaderWriter, logger *zerolog.Logger) {
	lis, err := net.Listen("tcp", ":"+settings.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't listen on gRPC port %s", settings.GRPCPort)
	}

	logger.Info().Msgf("Starting gRPC server on port %s", settings.GRPCPort)
	server := grpc.NewServer()
	pb.RegisterIntegrationServiceServer(server, api.NewIntegrationService(dbs))
	pb.RegisterUserDeviceServiceServer(server, api.NewUserDeviceService(dbs, logger))

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}

// ErrorHandler custom handler to log recovered errors using our logger and return json instead of string
func ErrorHandler(c *fiber.Ctx, err error, logger zerolog.Logger, environment string) error {
	code := fiber.StatusInternalServerError // Default 500 statuscode

	e, fiberTypeErr := err.(*fiber.Error)
	if fiberTypeErr {
		// Override status code if fiber.Error type
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	logger.Err(err).Msg("caught an error")
	// return an opaque error if we're in a higher level environment and we haven't specified an fiber type err.
	if !fiberTypeErr && environment == "prod" {
		err = fiber.NewError(fiber.StatusInternalServerError, "Internal error")
	}

	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": err.Error(),
	})
}
