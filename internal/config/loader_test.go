package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_loadFromYaml(t *testing.T) {
	var data = `
PORT: 3000
LOG_LEVEL: info
DB_USER: dimo
DB_PASSWORD: dimo
`
	settings, err := loadFromYaml([]byte(data))
	assert.NoError(t, err, "no error expected")
	assert.NotNilf(t, settings, "settings not expected to be nil")
	assert.Equal(t, "3000", settings.Port)
	assert.Equal(t, "info", settings.LogLevel)
	assert.Equal(t, "dimo", settings.DbUser)
	assert.Equal(t, "dimo", settings.DbPassword)
}

func Test_loadFromEnvVars(t *testing.T) {
	settings := Settings{
		Port:                 "3000",
		LogLevel:             "info",
		DbUser:               "dimo",
		DbPassword: "",
		DbPort:               "5432",
		DbHost:               "localhost",
	}
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_MAX_OPEN_CONNECTIONS", "5")

	loadFromEnvVars(&settings)
	assert.NotNilf(t, settings, "expected not nil")
	assert.Equal(t, "password", settings.DbPassword)
	assert.Equal(t, 5, settings.DbMaxOpenConnections)
	assert.Equal(t, "info", settings.LogLevel)
	assert.Equal(t, "localhost", settings.DbHost)
}