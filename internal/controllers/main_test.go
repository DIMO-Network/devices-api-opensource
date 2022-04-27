package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/test"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

const migrationsDirRelPath = "../../migrations"

// TestMain used to centralize setup that will run for all tests in this package
func TestMain(m *testing.M) {
	edb := setup()
	code := m.Run()
	teardown(edb)
	os.Exit(code)
}

// setup runs once for all tests in this package. This allows us to only bring the DB up once.
func setup() *embeddedpostgres.EmbeddedPostgres {
	edb, err := test.StartAndMigrateDB(context.Background(), migrationsDirRelPath)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("\033[1;36m%s\033[0m", "> DB Setup completed\n")
	return edb
}

func teardown(edb *embeddedpostgres.EmbeddedPostgres) {
	// Do something here.
	err := edb.Stop()
	if err != nil {
		fmt.Printf("error stopping embedded db: %v", err)
	}
	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed")
	fmt.Printf("\n")
}
