package config

import "fmt"

// Settings contains the application config
type Settings struct {
	Environment          string `yaml:"ENVIRONMENT"`
	Port                 string `yaml:"PORT"`
	LogLevel             string `yaml:"LOG_LEVEL"`
	DBUser               string `yaml:"DB_USER"`
	DBPassword           string `yaml:"DB_PASSWORD"`
	DBPort               string `yaml:"DB_PORT"`
	DBHost               string `yaml:"DB_HOST"`
	DBName               string `yaml:"DB_NAME"`
	DBMaxOpenConnections int    `yaml:"DB_MAX_OPEN_CONNECTIONS"`
	DBMaxIdleConnections int    `yaml:"DB_MAX_IDLE_CONNECTIONS"`
	ServiceName          string `yaml:"SERVICE_NAME"`
	JwtKeySetURL         string `yaml:"JWT_KEY_SET_URL"`
	SwaggerBaseURL       string `yaml:"SWAGGER_BASE_URL"`
	TorProxyURL          string `yaml:"TOR_PROXY_URL"`
	SmartcarClientID     string `yaml:"SMARTCAR_CLIENT_ID"`
	SmartcarClientSecret string `yaml:"SMARTCAR_CLIENT_SECRET"`
	SmartcarTestMode     bool   `yaml:"SMARTCAR_TEST_MODE"`
	SmartcarWebhookID    string `yaml:"SMARTCAR_WEBHOOK_ID"`
	RedisURL             string `yaml:"REDIS_URL"`
	RedisPassword        string `yaml:"REDIS_PASSWORD"`
	RedisTLS             bool   `yaml:"REDIS_TLS"`
	IngestSmartcarURL    string `yaml:"INGEST_SMARTCAR_URL"`
	IngestSmartcarTopic  string `yaml:"INGEST_SMARTCAR_TOPIC"`
	KafkaBrokers         string `yaml:"KAFKA_BROKERS"`
	DeviceStatusTopic    string `yaml:"DEVICE_STATUS_TOPIC"`
	EventsTopic          string `yaml:"EVENTS_TOPIC"`
	ElasticSearchHost    string `yaml:"ELASTIC_SEARCH_HOST"`
	ElasticSearchToken   string `yaml:"ELASTIC_SEARCH_TOKEN"`
}

// GetWriterDSN builds the connection string to the db writer - for now same as reader
func (app *Settings) GetWriterDSN(withSearchPath bool) string {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		app.DBUser,
		app.DBPassword,
		app.DBName,
		app.DBHost,
		app.DBPort,
	)
	if withSearchPath {
		dsn = fmt.Sprintf("%s search_path=%s", dsn, app.DBName) // assumption is schema has same name as dbname
	}
	return dsn
}
