package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})
	_ = validate.RegisterValidation("notplaceholder", func(fl validator.FieldLevel) bool {
		return !strings.Contains(fl.Field().String(), "replace-with-")
	})
	_ = validate.RegisterValidation("notrepeated", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		if len(str) == 0 {
			return true
		}
		return strings.Trim(str, string(str[0])) != ""
	})
}

// Validate checks whether the given Config conforms to Mobicode requirements using go-playground/validator.
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("validate configuration: config is nil")
	}

	err := validate.Struct(cfg)
	if err == nil {
		return nil
	}

	var valErrors validator.ValidationErrors
	if errors.As(err, &valErrors) {
		for _, fe := range valErrors {
			switch fe.Namespace() {
			case "Config.Environment":
				return fmt.Errorf("%s must be one of development, production", EnvServerEnv)
			case "Config.Server.Port":
				return fmt.Errorf("%s must be between 1 and 65535", EnvServerPort)
			case "Config.Server.DataDir":
				return fmt.Errorf("%s must not be empty", EnvServerDataDir)
			case "Config.Database.LogLevel":
				return fmt.Errorf("%s must be one of silent, error, warn, info", EnvServerDBLogLevel)
			case "Config.Settings.SecretToken":
				switch fe.Tag() {
				case "required":
					return fmt.Errorf("%s is required", EnvServerSecretToken)
				case "min":
					return fmt.Errorf("%s must be at least 32 bytes", EnvServerSecretToken)
				case "notplaceholder":
					return fmt.Errorf("%s cannot contain placeholder text", EnvServerSecretToken)
				case "notrepeated":
					return fmt.Errorf("%s cannot consist of repeated characters", EnvServerSecretToken)
				}
			case "Config.Paths.ConfigFile":
				return fmt.Errorf("ConfigFile must not be empty")
			case "Config.Paths.DatabaseDir":
				return fmt.Errorf("DatabaseDir must not be empty")
			case "Config.Paths.DatabaseFile":
				return fmt.Errorf("DatabaseFile must not be empty")
			}
		}
	}

	return fmt.Errorf("validate configuration: %w", err)
}
