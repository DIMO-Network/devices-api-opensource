package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
)

func setupDatabase(ctx context.Context, t *testing.T, migrationsDirRelPath string) (database.DbStore, *embeddedpostgres.EmbeddedPostgres) {
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
	pdb := database.NewDbConnectionFromSettings(ctx, &settings)
	time.Sleep(3 * time.Second) // get panic if don't have this here

	// run migrations at this point. need to do some pre-setup due to embedded db
	_, err := pdb.DBS().Writer.Exec(`
		grant usage on schema public to public;
		grant create on schema public to public;
		CREATE SCHEMA IF NOT EXISTS devices_api;
		SET search_path = devices_api, public;
		ALTER USER postgres SET search_path = devices_api, public;
		`)
	assert.Nil(t, err, "did not expect error connecting and executing query to DB")
	if err := goose.Run("up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
		_ = edb.Stop()
		log.Fatalf("failed to apply go code migrations: %v\n", err)
	}

	return pdb, edb
	// if we add code migrations, import: _ "github.com/DIMO-INC/devices-api/migrations"
}

func buildRequest(method, url, body string) *http.Request {
	req, _ := http.NewRequest(
		method,
		url,
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	return req
}

// authInjectorTestHandler injects fake jwt with sub
func authInjectorTestHandler(userID string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"nbf": time.Now().Unix(),
		})

		c.Locals("user", token)
		return c.Next()
	}
}
