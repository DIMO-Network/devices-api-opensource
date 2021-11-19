package controllers

import (
	"context"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/postgres"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"testing"
	"time"
)

func setupDatabase(ctx context.Context, t *testing.T) (postgres.DbStore, *embeddedpostgres.EmbeddedPostgres) {
	dbName := "devices_api"
	// an issue here is that if the test panics, it won't kill the embedded db: lsof -i :6669, then kill it.
	database := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V12).Port(6669).Database(dbName))
	if err := database.Start(); err != nil {
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
	pdb := postgres.NewDbStore(ctx, settings)
	time.Sleep(3 * time.Second) // get panic if don't have this here

	// can run migrations at this point

	return pdb, database
}
