package main

import (
	"database/sql"
	"log"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"

	_ "github.com/DIMO-INC/devices-api/migrations" // migrations won't work without these two
	_ "github.com/lib/pq"
)

func migrateDatabase(logger zerolog.Logger, settings *config.Settings) {
	var db *sql.DB
	// setup database
	db, err := sql.Open("postgres", settings.GetWriterDSN(false))
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
