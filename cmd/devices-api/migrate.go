package main

import (
	"database/sql"

	"github.com/DIMO-INC/devices-api/internal/config"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

func migrateDatabase(logger zerolog.Logger, settings *config.Settings, command string) {
	var db *sql.DB
	// setup database
	db, err := sql.Open("postgres", settings.GetWriterDSN(false))
	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal().Msgf("goose: failed to close DB: %v\n", err)
		}
	}()
	if err != nil {
		logger.Fatal().Msgf("failed to open db connection: %v\n", err)
	}
	if err = db.Ping(); err != nil {
		logger.Fatal().Msgf("failed to ping db: %v\n", err)
	}
	// set default
	if command == "" {
		command = "up"
	}
	if err := goose.Run(command, db, "migrations"); err != nil {
		logger.Fatal().Msgf("failed to apply go code migrations: %v\n", err)
	}
	// if we add any code migrations import _ "github.com/DIMO-INC/devices-api/migrations" // migrations won't work without this
}
