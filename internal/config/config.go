// Package config loads application settings from environment variables.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"go.uber.org/fx"
)

// Config contains API settings.
type Config struct {
	Port          string
	Env           string
	DatabasePath  string
	PublicBaseURL string
	DevNsec       string
}

// New constructs configuration from environment values.
func New() (*Config, error) {
	return Load("")
}

// Load constructs configuration from environment values, optionally overridden by a .env file.
func Load(path string) (*Config, error) {
	if path == "" {
		if _, err := os.Stat(".env"); err == nil {
			path = ".env"
		}
	}
	values, err := readFile(path)
	if err != nil {
		return nil, err
	}
	value := func(name string) string {
		if fromFile, ok := values[name]; ok {
			return fromFile
		}
		return os.Getenv(name)
	}

	port := value("PORT")
	if port == "" {
		port = "8080"
	}
	env := value("ENV")
	if env == "" {
		env = value("APP_ENV")
	}
	if env == "" {
		env = "development"
	}
	databasePath := value("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "data/app.db"
	}
	publicBaseURL := strings.TrimRight(strings.TrimSpace(value("PUBLIC_BASE_URL")), "/")
	devNsec := strings.TrimSpace(value("DEV_NSEC"))
	if publicBaseURL == "" {
		publicBaseURL = "http://localhost:" + port
	}
	parsedPublicURL, err := url.Parse(publicBaseURL)
	if err != nil || parsedPublicURL.Scheme == "" || parsedPublicURL.Host == "" || (parsedPublicURL.Path != "" && parsedPublicURL.Path != "/") || parsedPublicURL.RawQuery != "" || parsedPublicURL.Fragment != "" {
		return nil, fmt.Errorf("PUBLIC_BASE_URL must be an absolute origin without query or fragment")
	}
	if env == "production" && parsedPublicURL.Scheme != "https" {
		return nil, fmt.Errorf("PUBLIC_BASE_URL must use https in production")
	}
	if devNsec != "" && env != "development" {
		return nil, fmt.Errorf("DEV_NSEC may only be used in development")
	}

	return &Config{
		Port:          port,
		Env:           env,
		DatabasePath:  databasePath,
		PublicBaseURL: publicBaseURL,
		DevNsec:       devNsec,
	}, nil
}

func readFile(path string) (map[string]string, error) {
	values := make(map[string]string)
	if path == "" {
		return values, nil
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	for lineNumber, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		name, raw, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("parse config file line %d", lineNumber+1)
		}
		value := strings.TrimSpace(raw)
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			parsed, err := strconv.Unquote(value)
			if err != nil {
				return nil, fmt.Errorf("parse config file line %d: %w", lineNumber+1, err)
			}
			value = parsed
		}
		values[strings.TrimSpace(name)] = value
	}
	return values, nil
}

var Module = fx.Module("config", fx.Provide(New))
