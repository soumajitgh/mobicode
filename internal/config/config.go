package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// DatabaseLogLevel represents the log level for database operations.
type DatabaseLogLevel string

const (
	DatabaseLogLevelSilent DatabaseLogLevel = "silent"
	DatabaseLogLevelError  DatabaseLogLevel = "error"
	DatabaseLogLevelWarn   DatabaseLogLevel = "warn"
	DatabaseLogLevelInfo   DatabaseLogLevel = "info"
)

// Environment variable names for Mobicode configuration.
const (
	EnvServerEnv         = "MOBICODE_SERVER_ENV"
	EnvServerPort        = "MOBICODE_SERVER_PORT"
	EnvServerDevAssets   = "MOBICODE_SERVER_DEV_ASSETS"
	EnvServerPlayground  = "MOBICODE_SERVER_PLAYGROUND"
	EnvServerDataDir     = "MOBICODE_SERVER_DATA_DIR"
	EnvServerDBPath      = "MOBICODE_SERVER_DB_PATH"
	EnvServerDBLogLevel  = "MOBICODE_SERVER_DB_LOG_LEVEL"
	EnvServerSecretToken = "MOBICODE_SERVER_SECRET_TOKEN"
)

// Default configuration values.
const (
	DefaultEnvironment = "development"
	DefaultServerPort  = 8080
	DefaultDBLogLevel  = DatabaseLogLevelWarn
	DefaultDBFilename  = "mobicode.db"
)

// DefaultDataDir defines the default base directory for application data.
var DefaultDataDir = filepath.Join("tmp", "database")

// Config holds the validated application configuration.
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Settings    SettingsConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port       int
	DevAssets  bool
	Playground bool
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	DataDir  string
	Path     string
	LogLevel DatabaseLogLevel
}

// SettingsConfig holds application-wide settings.
type SettingsConfig struct {
	SecretToken string
}

// Load loads configuration from optional .env files, environment variables, applies defaults, and validates.
func Load(filenames ...string) (*Config, error) {
	if len(filenames) == 0 {
		if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) && !os.IsNotExist(err) {
			return nil, fmt.Errorf("load .env: %w", err)
		}
	} else {
		for _, file := range filenames {
			if file == "" {
				continue
			}
			if err := godotenv.Load(file); err != nil && !errors.Is(err, os.ErrNotExist) && !os.IsNotExist(err) {
				return nil, fmt.Errorf("load %s: %w", file, err)
			}
		}
	}

	cfg := &Config{}

	// Server Environment
	env := strings.TrimSpace(os.Getenv(EnvServerEnv))
	if env == "" {
		env = DefaultEnvironment
	}
	cfg.Environment = env

	// Server Port
	portStr := strings.TrimSpace(os.Getenv(EnvServerPort))
	if portStr == "" {
		cfg.Server.Port = DefaultServerPort
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("%s must be between 1 and 65535", EnvServerPort)
		}
		cfg.Server.Port = port
	}

	// Server DevAssets
	if devAssetsStr := strings.TrimSpace(os.Getenv(EnvServerDevAssets)); devAssetsStr != "" {
		if val, err := strconv.ParseBool(devAssetsStr); err == nil {
			cfg.Server.DevAssets = val
		}
	}

	// Server Playground
	if playgroundStr := strings.TrimSpace(os.Getenv(EnvServerPlayground)); playgroundStr != "" {
		if val, err := strconv.ParseBool(playgroundStr); err == nil {
			cfg.Server.Playground = val
		}
	}

	// Database DataDir & Path
	dataDir := strings.TrimSpace(os.Getenv(EnvServerDataDir))
	if dataDir == "" {
		dataDir = DefaultDataDir
	}
	cfg.Database.DataDir = dataDir

	dbPath := strings.TrimSpace(os.Getenv(EnvServerDBPath))
	if dbPath == "" {
		cfg.Database.Path = filepath.Join(cfg.Database.DataDir, DefaultDBFilename)
	} else {
		cfg.Database.Path = dbPath
		if strings.TrimSpace(os.Getenv(EnvServerDataDir)) == "" {
			cfg.Database.DataDir = filepath.Dir(dbPath)
		}
	}

	// Database LogLevel
	dbLogLevelStr := strings.ToLower(strings.TrimSpace(os.Getenv(EnvServerDBLogLevel)))
	if dbLogLevelStr == "" {
		cfg.Database.LogLevel = DefaultDBLogLevel
	} else {
		cfg.Database.LogLevel = DatabaseLogLevel(dbLogLevelStr)
	}

	// Settings SecretToken
	cfg.Settings.SecretToken = os.Getenv(EnvServerSecretToken)

	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
