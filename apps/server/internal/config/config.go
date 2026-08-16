// Package config loads and validates application settings from environment variables.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	return nil
}
