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
		config.EnvMobileAutoPair,
		config.EnvServerDataDir,
		config.EnvServerDBLogLevel,
		config.EnvServerSecretToken,
	}
	for _, key := range keys {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

func TestMobileAutoPairIsDevelopmentConfig(t *testing.T) {
	clearEnv(t)
	t.Setenv(config.EnvServerDataDir, t.TempDir())
	t.Setenv(config.EnvServerSecretToken, validSecret)
	t.Setenv(config.EnvMobileAutoPair, "true")
	cfg, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Development.MobileAutoPair {
		t.Fatal("expected development auto pairing enabled")
	}
	t.Setenv(config.EnvServerEnv, "production")
	cfg, err = config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Development.MobileAutoPair {
		t.Fatal("production should ignore development auto pairing")
	}
}

func TestNoEnvVarDefaultsToUserHome(t *testing.T) {
	clearEnv(t)
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	expectedDataDir := filepath.Join(tempHome, ".mobicode")
	if cfg.Server.DataDir != expectedDataDir {
		t.Errorf("expected Server.DataDir %q, got %q", expectedDataDir, cfg.Server.DataDir)
	}

	expectedConfigFile := filepath.Join(expectedDataDir, "config.toml")
	if cfg.Paths.ConfigFile != expectedConfigFile {
		t.Errorf("expected ConfigFile %q, got %q", expectedConfigFile, cfg.Paths.ConfigFile)
	}

	expectedDatabaseDir := filepath.Join(expectedDataDir, "database")
	if cfg.Paths.DatabaseDir != expectedDatabaseDir {
		t.Errorf("expected DatabaseDir %q, got %q", expectedDatabaseDir, cfg.Paths.DatabaseDir)
	}

	expectedDatabaseFile := filepath.Join(expectedDatabaseDir, "mobicode.db")
	if cfg.Paths.DatabaseFile != expectedDatabaseFile {
		t.Errorf("expected DatabaseFile %q, got %q", expectedDatabaseFile, cfg.Paths.DatabaseFile)
	}
	if cfg.Database.Path != expectedDatabaseFile {
		t.Errorf("expected Database.Path %q, got %q", expectedDatabaseFile, cfg.Database.Path)
	}

	// Verify defaults
	if cfg.Environment != config.DefaultEnvironment {
		t.Errorf("expected default Environment %q, got %q", config.DefaultEnvironment, cfg.Environment)
	}
	if cfg.Server.Port != config.DefaultServerPort {
		t.Errorf("expected default Port %d, got %d", config.DefaultServerPort, cfg.Server.Port)
	}
	if cfg.Database.LogLevel != config.DatabaseLogLevelWarn {
		t.Errorf("expected default LogLevel %q, got %q", config.DatabaseLogLevelWarn, cfg.Database.LogLevel)
	}

	// Verify directories and files created in temp directory
	if info, err := os.Stat(expectedDataDir); err != nil || !info.IsDir() {
		t.Errorf("expected data directory to exist, err: %v", err)
	}
	if info, err := os.Stat(expectedDatabaseDir); err != nil || !info.IsDir() {
		t.Errorf("expected database directory to exist, err: %v", err)
	}
	if info, err := os.Stat(expectedConfigFile); err != nil || info.IsDir() {
		t.Errorf("expected config.toml to exist, err: %v", err)
	}
}

func TestServerDataDirOverridesDefault(t *testing.T) {
	clearEnv(t)
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	customDir := t.TempDir()
	t.Setenv(config.EnvServerDataDir, customDir)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if cfg.Server.DataDir != customDir {
		t.Errorf("expected Server.DataDir %q, got %q", customDir, cfg.Server.DataDir)
	}

	// Ensure ~/.mobicode was NOT created
	defaultDir := filepath.Join(tempHome, ".mobicode")
	if _, err := os.Stat(defaultDir); !os.IsNotExist(err) {
		t.Errorf("expected default dir %q not to exist, but it was created", defaultDir)
	}
}

func TestAbsoluteCustomPath(t *testing.T) {
	clearEnv(t)
	absPath := filepath.Join(t.TempDir(), "custom_mobicode")
	t.Setenv(config.EnvServerDataDir, absPath)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.DataDir != absPath {
		t.Errorf("expected %q, got %q", absPath, cfg.Server.DataDir)
	}
	if !filepath.IsAbs(cfg.Server.DataDir) {
		t.Errorf("expected absolute path, got %q", cfg.Server.DataDir)
	}
}

func TestRelativeCustomPath(t *testing.T) {
	clearEnv(t)
	tempBase := t.TempDir()
	relSubdir := filepath.Join(filepath.Base(tempBase), "data_rel")
	relPath := filepath.Join("tmp", relSubdir)
	t.Cleanup(func() { _ = os.RemoveAll(relPath) })

	t.Setenv(config.EnvServerDataDir, relPath)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !filepath.IsAbs(cfg.Server.DataDir) {
		t.Errorf("expected normalized absolute path, got %q", cfg.Server.DataDir)
	}
	if !strings.HasSuffix(cfg.Server.DataDir, filepath.Clean(relPath)) {
		t.Errorf("expected path to end with %q, got %q", relPath, cfg.Server.DataDir)
	}
	if cfg.Paths.ConfigFile != filepath.Join(cfg.Server.DataDir, "config.toml") {
		t.Errorf("expected ConfigFile in relative dir, got %q", cfg.Paths.ConfigFile)
	}
}

func TestDirectoryAndDatabaseDirCreation(t *testing.T) {
	clearEnv(t)
	nestedDir := filepath.Join(t.TempDir(), "nested", "level", "data")
	t.Setenv(config.EnvServerDataDir, nestedDir)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info, err := os.Stat(cfg.Server.DataDir); err != nil || !info.IsDir() {
		t.Errorf("expected data directory to be created, err: %v", err)
	}
	if info, err := os.Stat(cfg.Paths.DatabaseDir); err != nil || !info.IsDir() {
		t.Errorf("expected database directory to be created, err: %v", err)
	}
}

func TestEmptyConfigTomlCreation(t *testing.T) {
	clearEnv(t)
	customDir := t.TempDir()
	t.Setenv(config.EnvServerDataDir, customDir)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(cfg.Paths.ConfigFile)
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty config.toml, got %d bytes", len(data))
	}
}

func TestExistingConfigTomlPreserved(t *testing.T) {
	clearEnv(t)
	customDir := t.TempDir()
	configFile := filepath.Join(customDir, "config.toml")
	expectedContent := "# custom configuration\n[custom]\nsetting = true\n"
	if err := os.WriteFile(configFile, []byte(expectedContent), 0o644); err != nil {
		t.Fatalf("failed to seed config.toml: %v", err)
	}

	t.Setenv(config.EnvServerDataDir, customDir)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(cfg.Paths.ConfigFile)
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}
	if string(data) != expectedContent {
		t.Errorf("expected existing config.toml to be preserved, got %q", string(data))
	}
}

func TestMalformedConfigTomlReturnsError(t *testing.T) {
	clearEnv(t)
	customDir := t.TempDir()
	configFile := filepath.Join(customDir, "config.toml")
	if err := os.WriteFile(configFile, []byte("invalid = ["), 0o644); err != nil {
		t.Fatalf("failed to seed malformed config.toml: %v", err)
	}

	t.Setenv(config.EnvServerDataDir, customDir)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	_, err := config.Load("")
	if err == nil {
		t.Fatal("expected error for malformed config.toml, got nil")
	}
	if !strings.Contains(err.Error(), "parse config file") {
		t.Errorf("expected error containing 'parse config file', got %v", err)
	}
}

func TestDatabasePathDerivesFromCustomDataDirectory(t *testing.T) {
	clearEnv(t)
	customDir := t.TempDir()
	t.Setenv(config.EnvServerDataDir, customDir)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDBFile := filepath.Join(customDir, "database", "mobicode.db")
	if cfg.Paths.DatabaseFile != expectedDBFile {
		t.Errorf("expected DatabaseFile %q, got %q", expectedDBFile, cfg.Paths.DatabaseFile)
	}
	if cfg.Database.Path != expectedDBFile {
		t.Errorf("expected Database.Path %q, got %q", expectedDBFile, cfg.Database.Path)
	}
}

func TestInvalidUnwritableDataDirectoryReturnsError(t *testing.T) {
	clearEnv(t)
	tempFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tempFile, []byte("blocker"), 0o600); err != nil {
		t.Fatalf("failed to write blocker file: %v", err)
	}

	// Pointing to a path beneath an existing file cannot be created with os.MkdirAll
	unwritableDir := filepath.Join(tempFile, "sub_directory")
	t.Setenv(config.EnvServerDataDir, unwritableDir)
	t.Setenv(config.EnvServerSecretToken, validSecret)

	_, err := config.Load("")
	if err == nil {
		t.Fatal("expected error for unwritable data directory, got nil")
	}
	if !strings.Contains(err.Error(), "create mobicode data directory") {
		t.Errorf("expected error containing 'create mobicode data directory', got %v", err)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("HOME", t.TempDir())
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
			t.Setenv("HOME", t.TempDir())
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

func TestValidateEmptyDataDir(t *testing.T) {
	cfg := &config.Config{
		Environment: "development",
		Server:      config.ServerConfig{Port: 8080, DataDir: ""},
		Database:    config.DatabaseConfig{LogLevel: config.DatabaseLogLevelWarn},
		Paths: config.PathsConfig{
			ConfigFile:   "/path/config.toml",
			DatabaseDir:  "/path/database",
			DatabaseFile: "/path/database/mobicode.db",
		},
		Settings: config.SettingsConfig{SecretToken: validSecret},
	}
	err := config.Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "MOBICODE_SERVER_DATA_DIR must not be empty") {
		t.Errorf("expected error for empty DataDir, got %v", err)
	}
}

func TestEnvironmentPrecedence(t *testing.T) {
	clearEnv(t)
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

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

	t.Setenv(config.EnvServerPort, "9999")
	t.Setenv(config.EnvServerDBLogLevel, "info")

	cfg, err := config.Load(envPath)
	if err != nil {
		t.Fatalf("failed to load config with temp .env: %v", err)
	}

	if cfg.Server.Port != 9999 {
		t.Errorf("expected process env Port 9999 to take precedence, got %d", cfg.Server.Port)
	}
	if cfg.Database.LogLevel != config.DatabaseLogLevelInfo {
		t.Errorf("expected process env LogLevel info to take precedence, got %s", cfg.Database.LogLevel)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected Environment 'production' from file, got %s", cfg.Environment)
	}
	if cfg.Settings.SecretToken != "file-secret-token-with-at-least-32-bytes" {
		t.Errorf("expected SecretToken from file, got %s", cfg.Settings.SecretToken)
	}
}
