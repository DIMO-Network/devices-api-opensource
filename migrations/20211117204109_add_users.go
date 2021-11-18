package migrations

import (
	"database/sql"
	"fmt"
	"github.com/pressly/goose/v3"
	"os"
)

func init() {
	goose.AddMigration(upAddUsers, downAddUsers)
}

func upAddUsers(tx *sql.Tx) error {
	svcAccountPwd := os.Getenv("SERVICE_ACCOUNT_PASSWORD")
	if svcAccountPwd == "" {
		svcAccountPwd = "service" // default eg. for local testing
	}
	_, err := tx.Exec(fmt.Sprintf("CREATE USER service with password %s;", svcAccountPwd))
	if err != nil {
		return err
	}
	_, err = tx.Exec(`GRANT usage ON SCHEMA devices_api to service;
		GRANT usage ON SCHEMA devices_api to service;
		GRANT SELECT, INSERT, UPDATE, DELETE, REFERENCES ON TABLES TO service;
		GRANT USAGE, SELECT ON SEQUENCES TO service;
		GRANT EXECUTE ON FUNCTIONS TO service;
		GRANT USAGE ON TYPES TO service;
	`)

	if err != nil {
		return err
	}
	return nil
}

func downAddUsers(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec(`REVOKE usage ON SCHEMA devices_api FROM service;
		DROP USER service;
	`)

	if err != nil {
		return err
	}
	return nil
}
