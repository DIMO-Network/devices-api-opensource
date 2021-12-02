package migrations

import (
	"database/sql"
	"os"
	"strings"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upBaseSchema, downBaseSchema)
}

func upBaseSchema(tx *sql.Tx) error {
	// do not do anything if in cloud since DB in aws will already have this stuff setup
	dbHost := os.Getenv("DB_HOST")
	if strings.Contains(dbHost, "amazonaws") {
		return nil
	}

	sql := `
		REVOKE CREATE ON schema public FROM public; -- public schema isolation
		CREATE SCHEMA IF NOT EXISTS devices_api;
	`

	_, err := tx.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func downBaseSchema(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec(`
DROP SCHEMA devices_api CASCADE;
GRANT CREATE, USAGE ON schema public TO public;
	`)

	if err != nil {
		return err
	}
	return nil
}
