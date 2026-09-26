package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soumajitgh/mobicode/internal/config"
)

const validSecret = "a-long-valid-secret-token-of-at-least-32-bytes"

func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		config.EnvServerEnv,
		config.EnvServerPort,
		config.EnvServerDevAssets,
		config.EnvServerPlayground,
		config.EnvServerDataDir,
		config.EnvServerDBPath,
		config.EnvServerDBLogLevel,
		config.EnvServerSecretToken,
	}
	for _, key := range keys {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

func TestDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error loading defaults: %v", err)
	}

	if cfg.Environment != config.DefaultEnvironment {
		t.Errorf("expected default Environment %q, got %q", config.DefaultEnvironment, cfg.Environment)
	}
	if cfg.Server.Port != config.DefaultServerPort {
		t.Errorf("expected default Port %d, got %d", config.DefaultServerPort, cfg.Server.Port)
	}
	if cfg.Server.DevAssets != false {
		t.Errorf("expected default DevAssets false, got %v", cfg.Server.DevAssets)
	}
	if cfg.Server.Playground != false {
		t.Errorf("expected default Playground false, got %v", cfg.Server.Playground)
	}
	expectedDataDir := filepath.Join("tmp", "database")
	if cfg.Database.DataDir != expectedDataDir {
		t.Errorf("expected default DataDir %q, got %q", expectedDataDir, cfg.Database.DataDir)
	}
	expectedDBPath := filepath.Join("tmp", "database", "mobicode.db")
	if cfg.Database.Path != expectedDBPath {
		t.Errorf("expected default Path %q, got %q", expectedDBPath, cfg.Database.Path)
	}
	if cfg.Database.LogLevel != config.DatabaseLogLevelWarn {
		t.Errorf("expected default LogLevel %q, got %q", config.DatabaseLogLevelWarn, cfg.Database.LogLevel)
	}
	if cfg.Settings.SecretToken != validSecret {
		t.Errorf("expected SecretToken %q, got %q", validSecret, cfg.Settings.SecretToken)
	}
}

func TestDefaultWithCustomDataDir(t *testing.T) {
	clearEnv(t)
	t.Setenv(config.EnvServerSecretToken, validSecret)
	customDir := filepath.Join("var", "mobicode")
	t.Setenv(config.EnvServerDataDir, customDir)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Database.DataDir != customDir {
		t.Errorf("expected DataDir %q, got %q", customDir, cfg.Database.DataDir)
	}
	expectedDBPath := filepath.Join(customDir, "mobicode.db")
	if cfg.Database.Path != expectedDBPath {
		t.Errorf("expected Path %q, got %q", expectedDBPath, cfg.Database.Path)
	}
}

func TestParsing(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		assertFn func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "valid port parsing",
			env: map[string]string{
				config.EnvServerPort: "3000",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Server.Port != 3000 {
					t.Errorf("expected port 3000, got %d", cfg.Server.Port)
				}
			},
		},
		{
			name: "port edge cases - min port",
			env: map[string]string{
				config.EnvServerPort: "1",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Server.Port != 1 {
					t.Errorf("expected port 1, got %d", cfg.Server.Port)
				}
			},
		},
		{
			name: "port edge cases - max port",
			env: map[string]string{
				config.EnvServerPort: "65535",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Server.Port != 65535 {
					t.Errorf("expected port 65535, got %d", cfg.Server.Port)
				}
			},
		},
		{
			name: "database log level silent",
			env: map[string]string{
				config.EnvServerDBLogLevel: "silent",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Database.LogLevel != config.DatabaseLogLevelSilent {
					t.Errorf("expected silent, got %s", cfg.Database.LogLevel)
				}
			},
		},
		{
			name: "database log level error",
			env: map[string]string{
				config.EnvServerDBLogLevel: "ERROR",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Database.LogLevel != config.DatabaseLogLevelError {
					t.Errorf("expected error, got %s", cfg.Database.LogLevel)
				}
			},
		},
		{
			name: "database log level info",
			env: map[string]string{
				config.EnvServerDBLogLevel: "Info",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Database.LogLevel != config.DatabaseLogLevelInfo {
					t.Errorf("expected info, got %s", cfg.Database.LogLevel)
				}
			},
		},
		{
			name: "production environment",
			env: map[string]string{
				config.EnvServerEnv: "production",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Environment != "production" {
					t.Errorf("expected production, got %s", cfg.Environment)
				}
			},
		},
		{
			name: "boolean flags enabled",
			env: map[string]string{
				config.EnvServerDevAssets:  "true",
				config.EnvServerPlayground: "1",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if !cfg.Server.DevAssets {
					t.Errorf("expected DevAssets true, got false")
				}
				if !cfg.Server.Playground {
					t.Errorf("expected Playground true, got false")
				}
			},
		},
		{
			name: "explicit database path",
			env: map[string]string{
				config.EnvServerDBPath: "/custom/path/app.db",
			},
			assertFn: func(t *testing.T, cfg *config.Config) {
				if cfg.Database.Path != "/custom/path/app.db" {
					t.Errorf("expected /custom/path/app.db, got %s", cfg.Database.Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(config.EnvServerSecretToken, validSecret)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := config.Load("")
			if err != nil {
				t.Fatalf("unexpected load error: %v", err)
			}
			tt.assertFn(t, cfg)
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		expectedError string
	}{
		{
			name: "missing server secret token",
			env: map[string]string{
				config.EnvServerSecretToken: "",
			},
			expectedError: "MOBICODE_SERVER_SECRET_TOKEN is required",
		},
		{
			name: "short server secret token",
			env: map[string]string{
				config.EnvServerSecretToken: "too-short-secret-token",
			},
			expectedError: "MOBICODE_SERVER_SECRET_TOKEN must be at least 32 bytes",
		},
		{
			name: "placeholder server secret token",
			env: map[string]string{
				config.EnvServerSecretToken: "replace-with-a-random-secret-of-at-least-32-bytes",
			},
			expectedError: "MOBICODE_SERVER_SECRET_TOKEN cannot contain placeholder text",
		},
		{
			name: "repeated character server secret token",
			env: map[string]string{
				config.EnvServerSecretToken: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			expectedError: "MOBICODE_SERVER_SECRET_TOKEN cannot consist of repeated characters",
		},
		{
			name: "invalid port - non-numeric",
			env: map[string]string{
				config.EnvServerPort: "not-a-number",
			},
			expectedError: "MOBICODE_SERVER_PORT must be between 1 and 65535",
		},
		{
			name: "invalid port - zero",
			env: map[string]string{
				config.EnvServerPort: "0",
			},
			expectedError: "MOBICODE_SERVER_PORT must be between 1 and 65535",
		},
		{
			name: "invalid port - negative",
			env: map[string]string{
				config.EnvServerPort: "-10",
			},
			expectedError: "MOBICODE_SERVER_PORT must be between 1 and 65535",
		},
		{
			name: "invalid port - out of range",
			env: map[string]string{
				config.EnvServerPort: "65536",
			},
			expectedError: "MOBICODE_SERVER_PORT must be between 1 and 65535",
		},
		{
			name: "invalid database log level",
			env: map[string]string{
				config.EnvServerDBLogLevel: "verbose",
			},
			expectedError: "MOBICODE_SERVER_DB_LOG_LEVEL must be one of silent, error, warn, info",
		},
		{
			name: "invalid environment",
			env: map[string]string{
				config.EnvServerEnv: "staging",
			},
			expectedError: "MOBICODE_SERVER_ENV must be one of development, production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(config.EnvServerSecretToken, validSecret)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			_, err := config.Load("")
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectedError)
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestValidateEmptyDatabasePath(t *testing.T) {
	cfg := &config.Config{
		Environment: "development",
		Server:      config.ServerConfig{Port: 8080},
		Database:    config.DatabaseConfig{Path: "", LogLevel: config.DatabaseLogLevelWarn},
		Settings:    config.SettingsConfig{SecretToken: validSecret},
	}
	err := config.Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "MOBICODE_SERVER_DB_PATH must not be empty") {
		t.Errorf("expected error for empty database path, got %v", err)
	}
}

func TestEnvironmentPrecedence(t *testing.T) {
	clearEnv(t)

	// Create a temp .env file with defaults from file
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	envContent := `MOBICODE_SERVER_ENV=production
MOBICODE_SERVER_PORT=9000
MOBICODE_SERVER_DB_LOG_LEVEL=silent
MOBICODE_SERVER_SECRET_TOKEN=file-secret-token-with-at-least-32-bytes
`
	if err := os.WriteFile(envPath, []byte(envContent), 0o600); err != nil {
		t.Fatalf("failed to write temp .env: %v", err)
	}

	// Override specific variables explicitly in the environment
	t.Setenv(config.EnvServerPort, "9999")
	t.Setenv(config.EnvServerDBLogLevel, "info")

	cfg, err := config.Load(envPath)
	if err != nil {
		t.Fatalf("failed to load config with temp .env: %v", err)
	}

	// Port and LogLevel should take the explicit process env values
	if cfg.Server.Port != 9999 {
		t.Errorf("expected process env Port 9999 to take precedence, got %d", cfg.Server.Port)
	}
	if cfg.Database.LogLevel != config.DatabaseLogLevelInfo {
		t.Errorf("expected process env LogLevel info to take precedence, got %s", cfg.Database.LogLevel)
	}

	// Environment and SecretToken should be populated from the .env file since process env wasn't set
	if cfg.Environment != "production" {
		t.Errorf("expected Environment 'production' from file, got %s", cfg.Environment)
	}
	if cfg.Settings.SecretToken != "file-secret-token-with-at-least-32-bytes" {
		t.Errorf("expected SecretToken from file, got %s", cfg.Settings.SecretToken)
	}
}
