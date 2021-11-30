package config

import "fmt"

// Settings contains the application config
type Settings struct {
	Port                 string `yaml:"PORT"`
	LogLevel             string `yaml:"LOG_LEVEL"`
	DbUser               string `yaml:"DB_USER"`
	DbPassword           string `yaml:"DB_PASSWORD"`
	DbPort               string `yaml:"DB_PORT"`
	DbHost               string `yaml:"DB_HOST"`
	DbName               string `yaml:"DB_NAME"`
	DbMaxOpenConnections int    `yaml:"DB_MAX_OPEN_CONNECTIONS"`
	DbMaxIdleConnections int    `yaml:"DB_MAX_IDLE_CONNECTIONS"`
	ServiceName          string `yaml:"SERVICE_NAME"`
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
