package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "dimo", settings.DBUser)
	assert.Equal(t, "dimo", settings.DBPassword)
}

func Test_loadFromEnvVars(t *testing.T) {
	settings := Settings{
		Port:       "3000",
		LogLevel:   "info",
		DBUser:     "dimo",
		DBPassword: "",
		DBPort:     "5432",
		DBHost:     "localhost",
	}
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_MAX_OPEN_CONNECTIONS", "5")

	loadFromEnvVars(&settings)
	assert.NotNilf(t, settings, "expected not nil")
	assert.Equal(t, "password", settings.DBPassword)
	assert.Equal(t, 5, settings.DBMaxOpenConnections)
	assert.Equal(t, "info", settings.LogLevel)
	assert.Equal(t, "localhost", settings.DBHost)
}

func Test_loadFromEnvVars_errOnNil(t *testing.T) {
	err := loadFromEnvVars(nil)
	assert.Error(t, err, "expected error if nil settings")
}
