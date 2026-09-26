package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolveDataDir determines the Mobicode data directory from environment or user home.
func resolveDataDir() (string, error) {
	raw := strings.TrimSpace(os.Getenv(EnvServerDataDir))
	if raw != "" {
		abs, err := filepath.Abs(filepath.Clean(raw))
		if err != nil {
			return "", fmt.Errorf("resolve configured data directory: %w", err)
		}
		return abs, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve default data directory: %w", err)
	}

	abs, err := filepath.Abs(filepath.Clean(filepath.Join(home, DefaultDataDirName)))
	if err != nil {
		return "", fmt.Errorf("resolve default data directory: %w", err)
	}
	return abs, nil
}

// initDirectories creates the data directory and database directory.
func initDirectories(dataDir string) (string, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("create mobicode data directory: %w", err)
	}

	dbDir := filepath.Join(dataDir, DefaultDBDirName)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return "", fmt.Errorf("create database directory: %w", err)
	}

	return dbDir, nil
}

// ensureConfigFile creates config.toml if missing and loads it.
func ensureConfigFile(dataDir string) (string, error) {
	configFile := filepath.Join(dataDir, DefaultConfigName)

	f, err := os.OpenFile(configFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err == nil {
		_ = f.Close()
	} else if !errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("create config file: %w", err)
	}

	if _, err := os.ReadFile(configFile); err != nil {
		return "", fmt.Errorf("load config file: %w", err)
	}

	return configFile, nil
}
