package test

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/internal/database"
	"github.com/DIMO-Network/devices-api/models"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pressly/goose/v3"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func SetupDatabase(ctx context.Context, t *testing.T, migrationsDirRelPath string) (database.DbStore, *embeddedpostgres.EmbeddedPostgres) {
	dbName := "devices_api"
	// an issue here is that if the test panics, it won't kill the embedded db: lsof -i :6669, then kill it.
	edb := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V12).Port(6669).Database(dbName))
	if err := edb.Start(); err != nil {
		t.Fatal(err)
	}

	settings := config.Settings{
		LogLevel:             "info",
		DBName:               dbName,
		DBHost:               "localhost",
		DBPort:               "6669",
		DBUser:               "postgres",
		DBPassword:           "postgres",
		DBMaxOpenConnections: 2,
		DBMaxIdleConnections: 2,
		ServiceName:          "devices-api",
	}
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
	assert.Nil(t, err, "did not expect error connecting and executing query to embedded DB for schema stuff")
	goose.SetTableName("devices_api.migrations")
	if err := goose.Run("up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
		_ = edb.Stop()
		log.Fatalf("failed to apply goose migrations for test: %v\n", err)
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
	assert.NoError(t, err)

	return pdb, edb
	// if we add code migrations, import: _ "github.com/DIMO-Network/devices-api/migrations"
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

// authInjectorTestHandler injects fake jwt with sub
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

func TruncateTables(db *sql.DB, t *testing.T) {
	_, err := db.Exec(`SELECT truncate_tables();`)
	if err != nil {
		t.Fatal(err)
	}
}

/** Test Setup functions. At some point may want to move elsewhere more generic **/

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
