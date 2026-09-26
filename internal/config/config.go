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

// Config holds the validated application configuration.
type Config struct {
	Environment string         `validate:"required,oneof=development production"`
	Server      ServerConfig   `validate:"required"`
	Database    DatabaseConfig `validate:"required"`
	Settings    SettingsConfig `validate:"required"`
	Paths       PathsConfig    `validate:"required"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port            int    `validate:"required,min=1,max=65535"`
	DataDir         string `validate:"required,notblank"`
	DevAssets       bool
	Playground      bool
	BaseURL         string
	ResolvedBaseURL string
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Path     string
	LogLevel DatabaseLogLevel `validate:"required,oneof=silent error warn info"`
}

// SettingsConfig holds application-wide settings.
type SettingsConfig struct {
	SecretToken string `validate:"required,min=32,notplaceholder,notrepeated"`
}

// PathsConfig holds fully resolved file and directory paths.
type PathsConfig struct {
	ConfigFile   string `validate:"required,notblank"`
	DatabaseDir  string `validate:"required,notblank"`
	DatabaseFile string `validate:"required,notblank"`
}

// Load loads configuration following the prescribed initialization sequence.
func Load(filenames ...string) (*Config, error) {
	// 1. Load optional .env
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

	// 2. Resolve MOBICODE_SERVER_DATA_DIR, or ~/.mobicode if empty
	dataDir, err := resolveDataDir()
	if err != nil {
		return nil, err
	}

	// 3. Create data directory and database directory
	dbDir, err := initDirectories(dataDir)
	if err != nil {
		return nil, err
	}

	// 4. Create config.toml if missing and load it
	configFile, err := ensureConfigFile(dataDir)
	if err != nil {
		return nil, err
	}

	// 5. Derive paths
	paths := PathsConfig{
		ConfigFile:   configFile,
		DatabaseDir:  dbDir,
		DatabaseFile: filepath.Join(dbDir, DefaultDBFilename),
	}

	cfg := &Config{
		Paths: paths,
	}
	cfg.Server.DataDir = dataDir
	cfg.Database.Path = paths.DatabaseFile

	// 6. Server Environment
	env := strings.TrimSpace(os.Getenv(EnvServerEnv))
	if env == "" {
		env = DefaultEnvironment
	}
	cfg.Environment = env

	// 7. Server Port
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

	// 8. Server DevAssets & Playground
	cfg.Server.BaseURL = strings.TrimSpace(os.Getenv(EnvServerBaseURL))
	if cfg.Server.BaseURL != "" {
		cfg.Server.ResolvedBaseURL, err = normalizeBaseURL(cfg.Server.BaseURL)
		if err != nil {
			return nil, err
		}
	} else {
		cfg.Server.ResolvedBaseURL = discoverBaseURL(cfg.Server.Port)
	}
	if devAssetsStr := strings.TrimSpace(os.Getenv(EnvServerDevAssets)); devAssetsStr != "" {
		if val, err := strconv.ParseBool(devAssetsStr); err == nil {
			cfg.Server.DevAssets = val
		}
	}
	if playgroundStr := strings.TrimSpace(os.Getenv(EnvServerPlayground)); playgroundStr != "" {
		if val, err := strconv.ParseBool(playgroundStr); err == nil {
			cfg.Server.Playground = val
		}
	}

	// 9. Database LogLevel
	dbLogLevelStr := strings.ToLower(strings.TrimSpace(os.Getenv(EnvServerDBLogLevel)))
	if dbLogLevelStr == "" {
		cfg.Database.LogLevel = DefaultDBLogLevel
	} else {
		cfg.Database.LogLevel = DatabaseLogLevel(dbLogLevelStr)
	}

	// 10. Settings SecretToken
	cfg.Settings.SecretToken = os.Getenv(EnvServerSecretToken)

	// 11. Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
