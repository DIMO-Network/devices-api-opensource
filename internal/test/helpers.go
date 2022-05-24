package test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const testDbName = "devices_api"
const testDbPort = 6669

// StartContainerDatabase starts postgres container with default test settings, and migrates the db. Caller must terminate container.
func StartContainerDatabase(ctx context.Context, t *testing.T, migrationsDirRelPath string) (database.DbStore, testcontainers.Container) {
	settings := getTestDbSettings()
	pgPort := "5432/tcp"
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", settings.DBUser, settings.DBPassword, port.Port(), settings.DBName)
	}
	cr := testcontainers.ContainerRequest{
		Image:        "postgres:12.9-alpine",
		Env:          map[string]string{"POSTGRES_USER": settings.DBUser, "POSTGRES_PASSWORD": settings.DBPassword, "POSTGRES_DB": settings.DBName},
		ExposedPorts: []string{pgPort},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
		WaitingFor:   wait.ForSQL(nat.Port(pgPort), "postgres", dbURL).Timeout(time.Second * 15),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: cr,
		Started:          true,
	})
	if err != nil {
		return handleContainerStartErr(ctx, err, pgContainer, t)
	}
	mappedPort, err := pgContainer.MappedPort(ctx, nat.Port(pgPort))
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to get container external port"), pgContainer, t)
	}
	fmt.Println("postgres container ready and running at port: ", mappedPort)
	//defer pgContainer.Terminate(ctx) // this should be done by the caller

	settings.DBPort = mappedPort.Port()
	pdb := database.NewDbConnectionFromSettings(ctx, &settings, false)
	for !pdb.IsReady() {
		time.Sleep(500 * time.Millisecond)
	}
	// can't connect to db, dsn=user=postgres password=postgres dbname=devices_api host=localhost port=49395 sslmode=disable search_path=devices_api, err=EOF
	// error happens when calling here
	_, err = pdb.DBS().Writer.Exec(`
		grant usage on schema public to public;
		grant create on schema public to public;
		CREATE SCHEMA IF NOT EXISTS devices_api;
		ALTER USER postgres SET search_path = devices_api, public;
		SET search_path = devices_api, public;
		`)
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to apply schema"), pgContainer, t)
	}
	// add truncate tables func
	_, err = pdb.DBS().Writer.Exec(`
CREATE OR REPLACE FUNCTION truncate_tables() RETURNS void AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'devices_api' and tablename != 'migrations';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
`)
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to create truncate func"), pgContainer, t)
	}

	goose.SetTableName("devices_api.migrations")
	if err := goose.Run("up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to apply goose migrations for test"), pgContainer, t)
	}

	return pdb, pgContainer
}
func handleContainerStartErr(ctx context.Context, err error, container testcontainers.Container, t *testing.T) (database.DbStore, testcontainers.Container) {
	if err != nil {
		fmt.Println("start container error: " + err.Error())
		if container != nil {
			container.Terminate(ctx) //nolint
		}
		t.Fatal(err)
	}
	return database.DbStore{}, container
}

// NewEmbedDBConfigured just returns the configured embed pg object, does not start db
func NewEmbedDBConfigured() *embeddedpostgres.EmbeddedPostgres {
	edb := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V12).Port(testDbPort).Database(testDbName))
	return edb
}

// StartAndMigrateDB used for booting up a test embed db. Migrates db schema to latest, adds function for truncating tables useful btw test runs.
func StartAndMigrateDB(ctx context.Context, migrationsDirRelPath string) (*embeddedpostgres.EmbeddedPostgres, error) {
	// an issue here is that if the test panics, it won't kill the embedded db: lsof -i :6669, then kill it.
	edb := NewEmbedDBConfigured()
	if err := edb.Start(); err != nil {
		return nil, err
	}

	settings := getTestDbSettings()
	// note the DBName will be used as the search path for the connection string
	pdb := database.NewDbConnectionFromSettings(ctx, &settings, false)
	time.Sleep(3 * time.Second) // get panic if don't have this here

	// run migrations at this point. need to do some pre-setup due to embedded db
	_, err := pdb.DBS().Writer.Exec(`
		grant usage on schema public to public;
		grant create on schema public to public;
		CREATE SCHEMA IF NOT EXISTS devices_api;
		ALTER USER postgres SET search_path = devices_api, public;
		SET search_path = devices_api, public;
		`)
	if err != nil {
		return nil, err
	}
	goose.SetTableName("devices_api.migrations")
	if err := goose.Run("up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
		_ = edb.Stop()
		return nil, errors.Wrapf(err, "failed to apply goose migrations for test")
	}
	// add truncate tables func
	_, err = pdb.DBS().Writer.Exec(`
CREATE OR REPLACE FUNCTION truncate_tables() RETURNS void AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'devices_api' and tablename != 'migrations';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
`)
	if err != nil {
		return nil, err
	}

	return edb, nil
}

// GetDBConnection gets a db connection to test embed db. Note the DBName will be used as the search path for the connection string
func GetDBConnection(ctx context.Context) database.DbStore {
	settings := getTestDbSettings()
	// note the DBName will be used as the search path for the connection string
	return database.NewDbConnectionFromSettings(ctx, &settings, false)
}

func SetupDatabase(ctx context.Context, t *testing.T, migrationsDirRelPath string) (database.DbStore, *embeddedpostgres.EmbeddedPostgres) {
	edb, err := StartAndMigrateDB(ctx, migrationsDirRelPath)
	// an issue here is that if the test panics, it won't kill the embedded db: lsof -i :6669, then kill it.
	if err != nil {
		t.Fatal(err)
	}
	pdb := GetDBConnection(ctx)

	return pdb, edb
	// if we add code migrations, import: _ "github.com/DIMO-Network/devices-api/migrations"
}

// getTestDbSettings builds test db config.Settings object
func getTestDbSettings() config.Settings {
	settings := config.Settings{
		LogLevel:             "info",
		DBName:               testDbName,
		DBHost:               "localhost",
		DBPort:               "6669",
		DBUser:               "postgres",
		DBPassword:           "postgres",
		DBMaxOpenConnections: 2,
		DBMaxIdleConnections: 2,
		ServiceName:          "devices-api",
	}
	return settings
}

func BuildRequest(method, url, body string) *http.Request {
	req, _ := http.NewRequest(
		method,
		url,
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	return req
}

// AuthInjectorTestHandler injects fake jwt with sub
func AuthInjectorTestHandler(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"nbf": time.Now().Unix(),
		})

		c.Locals("user", token)
		return c.Next()
	}
}

// TruncateTables truncates tables for the test db, useful to run as teardown at end of each DB dependent test.
func TruncateTables(db *sql.DB, t *testing.T) {
	fmt.Println("db exec: truncate_tables()")
	_, err := db.Exec(`SELECT truncate_tables();`)
	if err != nil {
		fmt.Println("truncating tables failed.")
		t.Fatal(err)
	}
}

/** Test Setup functions. At some point may want to move elsewhere more generic **/

func Logger() *zerolog.Logger {
	l := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "devices-api").
		Logger()
	return &l
}

func SetupCreateUserDevice(t *testing.T, testUserID string, dd *models.DeviceDefinition, powertrain *string, pdb database.DbStore) models.UserDevice {
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             testUserID,
		DeviceDefinitionID: dd.ID,
		CountryCode:        null.StringFrom("USA"),
	}
	if powertrain != nil {
		ud.Metadata = null.JSONFrom([]byte(fmt.Sprintf(`{"powertrainType": "%s"}`, *powertrain)))
	}
	err := ud.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return ud
}

func SetupCreateDeviceIntegration(t *testing.T, dd *models.DeviceDefinition, integration models.Integration, pdb database.DbStore) {
	di := models.DeviceIntegration{
		DeviceDefinitionID: dd.ID,
		IntegrationID:      integration.ID,
		Region:             "Americas",
	}
	err := di.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
}

func SetupCreateDeviceDefinition(t *testing.T, dm models.DeviceMake, model string, year int, pdb database.DbStore) *models.DeviceDefinition {
	dd := &models.DeviceDefinition{
		ID:           ksuid.New().String(),
		DeviceMakeID: dm.ID,
		Model:        model,
		Year:         int16(year),
		Verified:     true,
	}
	err := dd.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return dd
}

func SetupCreateMake(t *testing.T, mk string, pdb database.DbStore) models.DeviceMake {
	dm := models.DeviceMake{
		ID:   ksuid.New().String(),
		Name: mk,
	}
	err := dm.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "no db error expected")
	return dm
}

func SetupCreateSmartCarIntegration(t *testing.T, pdb database.DbStore) models.Integration {
	integration := models.Integration{
		ID:     ksuid.New().String(),
		Type:   models.IntegrationTypeAPI,
		Style:  models.IntegrationStyleWebhook,
		Vendor: "SmartCar",
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return integration
}

func SetupCreateAutoPiIntegration(t *testing.T, templateID int, evTemplateID *int, pdb database.DbStore) models.Integration {
	integration := models.Integration{
		ID:       ksuid.New().String(),
		Vendor:   "AutoPi",
		Type:     "Hardware",
		Style:    models.IntegrationStyleAddon,
		Metadata: null.JSONFrom([]byte(fmt.Sprintf(`{"autoPiDefaultTemplateId": %d }`, templateID))),
	}
	if evTemplateID != nil {
		integration.Metadata = null.JSONFrom([]byte(fmt.Sprintf(`{"autoPiDefaultTemplateId": %d,
			"autoPiPowertrainToTemplateId":{"BEV": %d}}`, templateID, *evTemplateID)))
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return integration
}

// SetupCreateUserDeviceAPIIntegration status set to Active, autoPiUnitId is optional
func SetupCreateUserDeviceAPIIntegration(t *testing.T, autoPiUnitID, externalID, userDeviceID, integrationID string, pdb database.DbStore) models.UserDeviceAPIIntegration {
	udapiInt := models.UserDeviceAPIIntegration{
		UserDeviceID:  userDeviceID,
		IntegrationID: integrationID,
		Status:        models.UserDeviceAPIIntegrationStatusActive,
		ExternalID:    null.StringFrom(externalID),
	}
	if autoPiUnitID != "" {
		md := fmt.Sprintf(`{"auto_pi_unit_id": "%s"}`, autoPiUnitID)
		_ = udapiInt.Metadata.UnmarshalJSON([]byte(md))
	}
	err := udapiInt.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return udapiInt
}

func SetupCreateAutoPiJob(t *testing.T, jobID, deviceID, cmd, userDeviceID string, pdb database.DbStore) *models.AutopiJob {
	autopiJob := models.AutopiJob{
		ID:             jobID,
		AutopiDeviceID: deviceID,
		Command:        cmd,
		State:          "sent",
		UserDeviceID:   null.StringFrom(userDeviceID),
	}
	err := autopiJob.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return &autopiJob
}
