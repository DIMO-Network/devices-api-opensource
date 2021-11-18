package migrations

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"os"
)

func init() {
	goose.AddMigration(upAddServiceUser, downAddServiceUser)
}

func upAddServiceUser(tx *sql.Tx) error {
	svcAccountPwd := os.Getenv("SERVICE_ACCOUNT_PASSWORD")
	if svcAccountPwd == "" {
		svcAccountPwd = "service" // default eg. for local testing
	}
	sql := fmt.Sprintf(`CREATE USER service with password '%s';
		GRANT usage ON SCHEMA devices_api to service;
	`, svcAccountPwd)

	_, err := tx.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "sql error creating user service.")
	}
	//future := `
	//	GRANT SELECT, INSERT, UPDATE, DELETE, REFERENCES ON TABLES TO service;
	//	GRANT USAGE, SELECT ON SEQUENCES TO service;
	//	GRANT EXECUTE ON FUNCTIONS TO service;
	//	GRANT USAGE ON TYPES TO service;
	//`)

	return nil
}

func downAddServiceUser(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec(`REVOKE usage ON SCHEMA devices_api FROM service;
		DROP USER service;
	`)

	if err != nil {
		return err
	}
	return nil
}
