package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/DIMO-Network/devices-api/docs"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/internal/kafka"
	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/DIMO-Network/shared"
	"github.com/Jeffail/benthos/v3/lib/util/hash/murmur2"
	"github.com/Shopify/sarama"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/customerio/go-customerio/v3"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
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

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load settings")
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		logger.Fatal().Err(err).Msgf("could not parse LOG_LEVEL: %s", settings.LogLevel)
	}
	zerolog.SetGlobalLevel(level)

	deps := newDependencyContainer(&settings, logger)

	pdb := database.NewDbConnectionFromSettings(ctx, &settings, true)
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
		migrateDatabase(logger, &settings, command)
	case "generate-events":
		eventService := services.NewEventService(&logger, &settings, deps.getKafkaProducer())
		generateEvents(logger, &settings, pdb, eventService)
	case "set-command-compat":
		if err := setCommandCompatibility(ctx, logger, &settings, pdb); err != nil {
			logger.Fatal().Err(err).Msg("Failed during command compatibility fill.")
		}
		logger.Info().Msg("Finished setting command compatibility.")
	case "smartcar-sync":
		syncSmartCarCompatibility(ctx, logger, &settings, pdb)
	case "create-tesla-integrations":
		if err := createTeslaIntegrations(ctx, pdb, &logger); err != nil {
			logger.Fatal().Err(err).Msg("Failed to create Tesla integrations")
		}
	case "edmunds-vehicles-sync":
		logger.Info().Msgf("Loading edmunds vehicles for device definitions and merging MMYs")
		err = loadEdmundsDeviceDefinitions(ctx, &logger, &settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("error trying to sync edmunds")
		}
	case "parkers-vehicles-sync":
		err = loadParkersDeviceDefinitions(ctx, &logger, &settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error syncing with Parkers")
		}
	case "adac-vehicles-sync":
		err = loadADACDeviceDefinitions(ctx, &logger, &settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error syncing with ADAC")
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
		loadEdmundsImages(ctx, logger, &settings, pdb, overwrite)
	case "remake-smartcar-topic":
		err = remakeSmartcarTopic(ctx, &logger, &settings, pdb, deps.getKafkaProducer())
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
		}
	case "remake-autopi-topic":
		err = remakeAutoPiTopic(ctx, &logger, &settings, pdb, deps.getKafkaProducer())
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running AutoPi Kafka re-registration")
		}
	case "remake-fence-topic":
		err = remakeFenceTopic(&logger, &settings, pdb, deps.getKafkaProducer())
		if err != nil {
			logger.Fatal().Err(err).Msg("Error running Smartcar Kafka re-registration")
		}
	case "search-sync-dds":
		logger.Info().Msg("loading device definitions from our DB to elastic cluster")
		err := loadElasticDevices(ctx, &logger, &settings, pdb)
		if err != nil {
			logger.Fatal().Err(err).Msg("error syncing with elastic")
		}
	case "populate-usa-powertrain":
		logger.Info().Msg("Populating USA powertrain data from VINs")
		nhtsaSvc := services.NewNHTSAService()
		err := populateUSAPowertrain(ctx, &logger, &settings, pdb, nhtsaSvc)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error filling in powertrain data.")
		}
	case "stop-task-by-key":
		if len(os.Args[1:]) != 2 {
			logger.Fatal().Msgf("Expected an argument, the task key.")
		}
		taskKey := os.Args[2]
		logger.Info().Msgf("Stopping task %s", taskKey)
		err := stopTaskByKey(ctx, &logger, &settings, pdb, taskKey, deps.getKafkaProducer())
		if err != nil {
			logger.Fatal().Err(err).Msg("Error stopping task.")
		}
	case "start-smartcar-from-refresh":
		if len(os.Args[1:]) != 2 {
			logger.Fatal().Msgf("Expected an argument, the device ID.")
		}
		userDeviceID := os.Args[2]
		logger.Info().Msgf("Trying to start Smartcar task for %s.", userDeviceID)
		var cipher shared.Cipher
		if settings.Environment == "dev" || settings.Environment == "prod" {
			cipher = createKMS(&settings, &logger)
		} else {
			logger.Warn().Msg("Using ROT13 encrypter. Only use this for testing!")
			cipher = new(shared.ROT13Cipher)
		}
		scClient := services.NewSmartcarClient(&settings)
		scTask := services.NewSmartcarTaskService(&settings, deps.getKafkaProducer())
		if err := startSmartcarFromRefresh(ctx, &logger, &settings, pdb, cipher, userDeviceID, scClient, scTask); err != nil {
			logger.Fatal().Err(err).Msg("Error starting Smartcar task.")
		}
		logger.Info().Msgf("Successfully started Smartcar task for %s.", userDeviceID)
	default:
		startMonitoringServer(logger)
		eventService := services.NewEventService(&logger, &settings, deps.getKafkaProducer())
		startDeviceStatusConsumer(logger, &settings, pdb, eventService)
		startCredentialConsumer(logger, &settings, pdb)
		startTaskStatusConsumer(logger, &settings, pdb)
		startMintStatusConsumer(logger, &settings, pdb)
		startWebAPI(logger, &settings, pdb, eventService, deps.getKafkaProducer(), deps.getS3ServiceClient(ctx), deps.getS3NFTServiceClient(ctx))
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

func createKMS(settings *config.Settings, logger *zerolog.Logger) shared.Cipher {
	// Need AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY to be set.
	// TODO(elffjs): Can we let the SDK grab the region too?
	awscfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(settings.AWSRegion))
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't create AWS config.")
	}

	return &shared.KMSCipher{
		KeyID:  settings.KMSKeyID,
		Client: kms.NewFromConfig(awscfg),
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

func startDeviceStatusConsumer(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore, eventService services.EventService) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
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
		logger.Fatal().Err(err).Msg("Could not start device status update consumer")
	}
	ingestSvc := services.NewDeviceStatusIngestService(pdb.DBS, &logger, eventService)
	consumer.Start(context.Background(), ingestSvc.ProcessDeviceStatusMessages)

	logger.Info().Msg("Device status update consumer started")
}

func startCredentialConsumer(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.TaskCredentialTopic,
		GroupID:         "user-devicesYY",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start credential update consumer")
	}
	credService := services.NewCredentialListener(pdb.DBS, &logger)
	consumer.Start(context.Background(), credService.ProcessCredentialsMessages)

	logger.Info().Msg("Credential update consumer started")
}

func startTaskStatusConsumer(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.TaskStatusTopic,
		GroupID:         "user-devices",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start credential update consumer")
	}
	cio := customerio.NewTrackClient(
		settings.CIOSiteID,
		settings.CIOApiKey,
		customerio.WithRegion(customerio.RegionUS),
	)
	taskStatusService := services.NewTaskStatusListener(pdb.DBS, &logger, cio)
	consumer.Start(context.Background(), taskStatusService.ProcessTaskUpdates)

	logger.Info().Msg("Task status consumer started")
}

func startMintStatusConsumer(logger zerolog.Logger, settings *config.Settings, pdb database.DbStore) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = sarama.V2_8_1_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	cfg := &kafka.Config{
		ClusterConfig:   clusterConfig,
		BrokerAddresses: strings.Split(settings.KafkaBrokers, ","),
		Topic:           settings.NFTOutputTopic,
		GroupID:         "user-devices",
		MaxInFlight:     int64(5),
	}
	consumer, err := kafka.NewConsumer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not start credential update consumer")
	}

	nftListenService := services.NewNFTListener(pdb.DBS, &logger)
	consumer.Start(context.Background(), nftListenService.ProcessMintStatus)

	logger.Info().Msg("NFT mint status consumer started")
}

func startMonitoringServer(logger zerolog.Logger) {
	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	monApp.Put("/loglevel", changeLogLevel)

	go func() {
		// TODO(elffjs): Make the port a setting.
		if err := monApp.Listen(":8888"); err != nil {
			logger.Fatal().Err(err).Str("port", "8888").Msg("Failed to start monitoring web server.")
		}
	}()

	logger.Info().Str("port", "8888").Msg("Started monitoring web server.")
}

// dependencyContainer way to hold different dependencies we need for our app. We could put all our deps and follow this pattern for everything.
type dependencyContainer struct {
	kafkaProducer      sarama.SyncProducer
	settings           *config.Settings
	logger             *zerolog.Logger
	s3ServiceClient    *s3.Client
	s3NFTServiceClient *s3.Client
}

func newDependencyContainer(settings *config.Settings, logger zerolog.Logger) dependencyContainer {
	return dependencyContainer{
		settings: settings,
		logger:   &logger,
	}
}

// getKafkaProducer instantiates a new kafka producer if not already set in our container and returns
func (dc *dependencyContainer) getKafkaProducer() sarama.SyncProducer {
	if dc.kafkaProducer == nil {
		p, err := createKafkaProducer(dc.settings)
		if err != nil {
			dc.logger.Fatal().Err(err).Msg("Could not initialize Kafka producer, terminating")
		}
		dc.kafkaProducer = p
	}
	return dc.kafkaProducer
}

// getS3ServiceClient instantiates a new default config and then a new s3 services client if not already set. Takes context in, although it could likely use a context from container passed in on instantiation
func (dc *dependencyContainer) getS3ServiceClient(ctx context.Context) *s3.Client {
	if dc.s3ServiceClient == nil {

		cfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(dc.settings.AWSRegion),
			// Comment the below out if not using localhost
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {

					if dc.settings.Environment == "local" {
						return aws.Endpoint{PartitionID: "aws", URL: dc.settings.DocumentsAWSEndpoint, SigningRegion: dc.settings.AWSRegion}, nil // The SigningRegion key was what's was missing! D'oh.
					}

					// returning EndpointNotFoundError will allow the service to fallback to its default resolution
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				})))

		if err != nil {
			dc.logger.Fatal().Err(err).Msg("Could not load aws config, terminating")
		}

		dc.s3ServiceClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.Region = dc.settings.AWSRegion
			o.Credentials = credentials.NewStaticCredentialsProvider(dc.settings.DocumentsAWSAccessKeyID, dc.settings.DocumentsAWSSecretsAccessKey, "")
		})
	}
	return dc.s3ServiceClient
}

func (dc *dependencyContainer) getS3NFTServiceClient(ctx context.Context) *s3.Client {
	if dc.s3NFTServiceClient == nil {

		cfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(dc.settings.AWSRegion),
			// Comment the below out if not using localhost
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {

					if dc.settings.Environment == "local" {
						return aws.Endpoint{PartitionID: "aws", URL: dc.settings.DocumentsAWSEndpoint, SigningRegion: dc.settings.AWSRegion}, nil // The SigningRegion key was what's was missing! D'oh.
					}

					// returning EndpointNotFoundError will allow the service to fallback to its default resolution
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				})))

		if err != nil {
			dc.logger.Fatal().Err(err).Msg("Could not load aws config, terminating")
		}

		dc.s3NFTServiceClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.Region = dc.settings.AWSRegion
			o.Credentials = credentials.NewStaticCredentialsProvider(dc.settings.NFTAWSAccessKeyID, dc.settings.NFTAWSSecretsAccessKey, "")
		})
	}
	return dc.s3NFTServiceClient
}
