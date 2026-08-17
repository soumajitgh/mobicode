// Package config loads and validates application settings from environment variables.
package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"
)

type Config struct {
	Environment string `env:"APP_ENV" envDefault:"development"`
	Server      ServerConfig
	Database    DatabaseConfig
	Log         LogConfig
}

type ServerConfig struct {
	Host            string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	MaxBodyBytes    int64         `env:"HTTP_MAX_BODY_BYTES" envDefault:"1048576"`
}

type DatabaseConfig struct {
	Path         string        `env:"DATABASE_PATH" envDefault:"./data/app.db"`
	BusyTimeout  time.Duration `env:"DATABASE_BUSY_TIMEOUT" envDefault:"5s"`
	MaxOpenConns int           `env:"DATABASE_MAX_OPEN_CONNS" envDefault:"1"`
	MaxIdleConns int           `env:"DATABASE_MAX_IDLE_CONNS" envDefault:"1"`
}

type LogConfig struct {
	Level       string `env:"LOG_LEVEL" envDefault:"info"`
	Development bool   `env:"LOG_DEVELOPMENT" envDefault:"true"`
}

// Load reads and validates environment configuration.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks configuration values needed at startup.
func (c Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}
	if c.Database.Path == "" {
		return fmt.Errorf("DATABASE_PATH is required")
	}
	if c.Database.BusyTimeout <= 0 {
		return fmt.Errorf("DATABASE_BUSY_TIMEOUT must be positive")
	}
	if c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("database connection limits are invalid")
	}
	if c.Server.MaxBodyBytes < 1 {
		return fmt.Errorf("HTTP_MAX_BODY_BYTES must be positive")
	}
	if filepath.Dir(c.Database.Path) == "" {
		return fmt.Errorf("DATABASE_PATH must include a directory")
	}
	return nil
}

var Module = fx.Module("config", fx.Provide(Load))
