// Package config loads and validates application settings from environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

type Config struct {
	Environment string `env:"SHARED_ENV" envDefault:"development"`
	Server      ServerConfig
	Database    DatabaseConfig
	Log         LogConfig
}

type ServerConfig struct {
	Host              string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port              int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout       time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout      time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout       time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout   time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	MaxBodyBytes      int64         `env:"SERVER_MAX_BODY_BYTES" envDefault:"1048576"`
	GraphQLComplexity int           `env:"SERVER_GRAPHQL_COMPLEXITY" envDefault:"250"`
}

type DatabaseConfig struct {
	URL            string        `env:"SERVER_DATABASE_URL" envDefault:"http://localhost:9081"`
	StartupTimeout time.Duration `env:"SERVER_DATABASE_STARTUP_TIMEOUT" envDefault:"30s"`
	MaxOpenConns   int           `env:"SERVER_DATABASE_MAX_OPEN_CONNS" envDefault:"1"`
	MaxIdleConns   int           `env:"SERVER_DATABASE_MAX_IDLE_CONNS" envDefault:"1"`
}

type LogConfig struct {
	Level       string `env:"SERVER_LOG_LEVEL" envDefault:"info"`
	Development bool   `env:"SERVER_LOG_DEVELOPMENT" envDefault:"true"`
}

// Load reads and validates environment configuration.
func Load() (Config, error) {
	if err := loadRootEnv(); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// loadRootEnv loads a repository-root .env file when one is present.
func loadRootEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				return fmt.Errorf("load %s: %w", envPath, err)
			}
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect %s: %w", envPath, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

// Validate checks configuration values needed at startup.
func (c Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	if c.Database.URL == "" {
		return fmt.Errorf("SERVER_DATABASE_URL is required")
	}
	if c.Database.StartupTimeout <= 0 {
		return fmt.Errorf("SERVER_DATABASE_STARTUP_TIMEOUT must be positive")
	}
	if c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("database connection limits are invalid")
	}
	if c.Server.MaxBodyBytes < 1 {
		return fmt.Errorf("SERVER_MAX_BODY_BYTES must be positive")
	}
	if c.Server.GraphQLComplexity < 1 {
		return fmt.Errorf("SERVER_GRAPHQL_COMPLEXITY must be positive")
	}
	return nil
}

var Module = fx.Module("config", fx.Provide(Load))
