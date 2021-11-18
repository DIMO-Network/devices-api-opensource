package scripts

import (
	"database/sql"
	"embed"
	"github.com/DIMO-INC/devices-api/internal/config"
	"log"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	// would doing all this just be same as `goose up` and setting env vars GOOSE_DRIVER=postgres and GOOSE_DBSTRING=DSN ?
	settings := config.LoadConfig()
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

	goose.SetBaseFS(embedMigrations)

	// test these - maybe the second one can run both sql and go migrations
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("failed to apply sql migrations: %v\n", err)
	}
	if err := goose.Run("up", db, "migrations"); err != nil {
		log.Fatalf("failed to apply go code migrations: %v\n", err)
	}
}