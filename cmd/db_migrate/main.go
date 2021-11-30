package main

import (
	"database/sql"
	"log"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/pressly/goose/v3"

	_ "github.com/DIMO-INC/devices-api/migrations"
	_ "github.com/lib/pq"
)

func main() {
	// would doing all this just be same as `goose up` and setting env vars GOOSE_DRIVER=postgres and GOOSE_DBSTRING=DSN ?
	settings, err := config.LoadConfig("settings.yaml")
	if err != nil {
		log.Fatalf("failed to load settings: %v\n", err)
	}
	var db *sql.DB
	// setup database
	db, err = sql.Open("postgres", settings.GetWriterDSN(false))
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to open db connection: %v\n", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v\n", err)
	}

	if err := goose.Run("up", db, "migrations"); err != nil {
		log.Fatalf("failed to apply go code migrations: %v\n", err)
	}
}
