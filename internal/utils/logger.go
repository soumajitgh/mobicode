package utils

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates the application logger for the selected environment.
func New(environment string) (*zap.Logger, error) {
	switch environment {
	case "", "development":
		config := zap.NewDevelopmentConfig()
		config.DisableCaller = true
		config.DisableStacktrace = true
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
		config.EncoderConfig.ConsoleSeparator = "  "
		return config.Build()
	case "production":
		return zap.NewProduction()
	default:
		return nil, fmt.Errorf("invalid MOBICODE_SERVER_ENV %q (use development or production)", environment)
	}
}
