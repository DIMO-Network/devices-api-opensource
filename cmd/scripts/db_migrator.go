package scripts

import (
	"database/sql"
	"embed"
	"github.com/DIMO-INC/devices-api/internal/config"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	settings := config.Settings{} // todo: pull in settings
	var db *sql.DB
	// setup database
	db, err := sql.Open("postgres", settings.GetWriterDSN())
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	// todo: how can we replace strings in migrations, eg. create user with password 'replaceme'
	// could we mix .go and .sql migrations?
	goose.SetBaseFS(embedMigrations)

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}

}