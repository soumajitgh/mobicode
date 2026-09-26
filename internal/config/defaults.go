package config

import "time"

// Default configuration values.
const (
	DefaultEnvironment = "development"
	DefaultServerPort  = 8080
	DefaultDBLogLevel  = DatabaseLogLevelWarn
	DefaultDBFilename  = "mobicode.db"
	DefaultConfigName  = "config.toml"
	DefaultDataDirName = ".mobicode"
	DefaultDBDirName   = "database"
)

const PairingLifetime = 3 * time.Minute

// Environment variable names for Mobicode configuration.
const (
	EnvServerEnv         = "MOBICODE_SERVER_ENV"
	EnvServerPort        = "MOBICODE_SERVER_PORT"
	EnvServerDataDir     = "MOBICODE_SERVER_DATA_DIR"
	EnvServerDBLogLevel  = "MOBICODE_SERVER_DB_LOG_LEVEL"
	EnvServerSecretToken = "MOBICODE_SERVER_SECRET_TOKEN"
	EnvServerDevAssets   = "MOBICODE_SERVER_DEV_ASSETS"
	EnvServerPlayground  = "MOBICODE_SERVER_PLAYGROUND"
	EnvServerBaseURL     = "MOBICODE_SERVER_BASE_URL"
)
