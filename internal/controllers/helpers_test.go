package controllers

import (
	"context"
	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"testing"
	"time"
)

func setupDatabase(ctx context.Context, t *testing.T) (database.DbStore, *embeddedpostgres.EmbeddedPostgres) {
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
		DbUser:               "database",
		DbPassword:           "database",
		DbMaxOpenConnections: 2,
		DbMaxIdleConnections: 2,
		ServiceName:          "devices-api",
	}
	pdb := database.NewDbConnectionFromSettings(ctx, settings)
	time.Sleep(3 * time.Second) // get panic if don't have this here

	// can run migrations at this point
	return pdb, edb
}
