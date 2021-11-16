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

// WriterConnectionString builds the connection string to the db writer - for now same as reader
func (app *Settings) WriterConnectionString() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable search_path=devices",
		app.DbUser,
		app.DbPassword,
		app.DbName,
		app.DbHost,
		app.DbPort,
	)
}
