package config

import "fmt"

// Settings contains the application config
type Settings struct {
	Port                 string
	LogLevel             string
	DbUser               string
	DbPassword           string
	DbPort               string
	DbHost               string
	DbName               string
	DbMaxOpenConnections int
	DbMaxIdleConnections int
	ServiceName          string
}

// GetWriterDSN builds the connection string to the db writer - for now same as reader
func (app *Settings) GetWriterDSN(withSearchPath bool) string {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		app.DbUser,
		app.DbPassword,
		app.DbName,
		app.DbHost,
		app.DbPort,
	)
	if withSearchPath {
		dsn = fmt.Sprintf("%s search_path=%s", dsn, app.DbName) // assumption is schema has same name as dbname
	}
	return dsn
}

