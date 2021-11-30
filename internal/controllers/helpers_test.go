package controllers

import (
	"context"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	_ "github.com/DIMO-INC/devices-api/migrations"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"

	"log"
	"testing"
	"time"
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
		DbName:               dbName,
		DbHost:               "localhost",
		DbPort:               "6669",
		DbUser:               "postgres",
		DbPassword:           "postgres",
		DbMaxOpenConnections: 2,
		DbMaxIdleConnections: 2,
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
}
