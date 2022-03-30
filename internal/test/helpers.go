package test

import (
	"context"
	"database/sql"
	"fmt"
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
	_, err := db.Exec(`SELECT truncate_tables();`)
	if err != nil {
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

func SetupCreateUserDevice(t *testing.T, testUserID string, dd *models.DeviceDefinition, pdb database.DbStore) models.UserDevice {
	ud := models.UserDevice{
		ID:                 ksuid.New().String(),
		UserID:             testUserID,
		DeviceDefinitionID: dd.ID,
		CountryCode:        null.StringFrom("USA"),
	}
	err := ud.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err)
	return ud
}

func SetupCreateDeviceIntegration(t *testing.T, dd *models.DeviceDefinition, integration models.Integration, ud models.UserDevice, pdb database.DbStore) {
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

func SetupCreateAutoPiIntegration(t *testing.T, templateID int, pdb database.DbStore) models.Integration {
	integration := models.Integration{
		ID:       ksuid.New().String(),
		Vendor:   "AutoPi",
		Type:     "API",
		Style:    models.IntegrationStyleAddon,
		Metadata: null.JSONFrom([]byte(fmt.Sprintf(`{"auto_pi_default_template_id": %d }`, templateID))),
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return integration
}
