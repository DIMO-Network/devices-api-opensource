package config

// LoadConfig fills in all the values in the Settings from local yml file (for dev) and env vars (for deployments)
func LoadConfig() *Settings {
	// todo: load this from local yml file and env vars
	settings := Settings{
		Port:                 "3000",
		LogLevel:             "info",
		DbUser:               "dimo",
		DbPassword:           "dimo",
		DbPort:               "5432",
		DbHost:               "localhost",
		DbName:               "devices_api",
		DbMaxOpenConnections: 5,
		DbMaxIdleConnections: 5,
		ServiceName:          "devices-api",
	}
	return &settings
}
