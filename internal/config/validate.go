package config

import (
	"fmt"
	"strings"
)

// Validate checks whether the given Config conforms to Mobicode requirements.
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("validate configuration: config is nil")
	}

	// Environment validation
	switch cfg.Environment {
	case "development", "production":
		// valid
	default:
		return fmt.Errorf("%s must be one of development, production", EnvServerEnv)
	}

	// Server Port validation
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", EnvServerPort)
	}

	// Server DataDir validation
	if strings.TrimSpace(cfg.Server.DataDir) == "" {
		return fmt.Errorf("%s must not be empty", EnvServerDataDir)
	}

	// Database LogLevel validation
	switch cfg.Database.LogLevel {
	case DatabaseLogLevelSilent, DatabaseLogLevelError, DatabaseLogLevelWarn, DatabaseLogLevelInfo:
		// valid
	default:
		return fmt.Errorf("%s must be one of silent, error, warn, info", EnvServerDBLogLevel)
	}

	// Settings SecretToken validation
	secret := cfg.Settings.SecretToken
	if secret == "" {
		return fmt.Errorf("%s is required", EnvServerSecretToken)
	}
	if len(secret) < 32 {
		return fmt.Errorf("%s must be at least 32 bytes", EnvServerSecretToken)
	}
	if strings.Contains(secret, "replace-with-") {
		return fmt.Errorf("%s cannot contain placeholder text", EnvServerSecretToken)
	}
	if strings.Trim(secret, string(secret[0])) == "" {
		return fmt.Errorf("%s cannot consist of repeated characters", EnvServerSecretToken)
	}

	// Paths validation
	if strings.TrimSpace(cfg.Paths.ConfigFile) == "" {
		return fmt.Errorf("ConfigFile must not be empty")
	}
	if strings.TrimSpace(cfg.Paths.DatabaseDir) == "" {
		return fmt.Errorf("DatabaseDir must not be empty")
	}
	if strings.TrimSpace(cfg.Paths.DatabaseFile) == "" {
		return fmt.Errorf("DatabaseFile must not be empty")
	}

	return nil
}
