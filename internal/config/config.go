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
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Settings    SettingsConfig
	Paths       PathsConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port       int
	DataDir    string
	DevAssets  bool
	Playground bool
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Path     string
	LogLevel DatabaseLogLevel
}

// SettingsConfig holds application-wide settings.
type SettingsConfig struct {
	SecretToken string
}

// PathsConfig holds fully resolved file and directory paths.
type PathsConfig struct {
	ConfigFile   string
	DatabaseDir  string
	DatabaseFile string
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
