package gophermart

import (
	"flag"
	"os"
)

const (
	_serverAddress  = ":8080"
	_accrualAddress = ":8081"
	_databaseURI    = ""
	_logLevel       = "info"
)

type Config struct {
	ServerAddress  string
	AccrualAddress string
	DatabaseUri    string
	LogLevel       string
}

func NewConfig() *Config {
	serverAddress := flag.String("a", _serverAddress, "Server address")
	accrualAddress := flag.String("r", _accrualAddress, "Accrual address")
	databaseUri := flag.String("d", _databaseURI, "Database URI")
	logLevel := flag.String("l", _logLevel, "Log level")

	flag.Parse()

	return &Config{
		ServerAddress:  getEnv("RUN_ADDRESS", *serverAddress),
		AccrualAddress: getEnv("ACCRUAL_SYSTEM_ADDRESS", *accrualAddress),
		DatabaseUri:    getEnv("DATABASE_URI", *databaseUri),
		LogLevel:       getEnv("LOG_LEVEL", *logLevel),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
