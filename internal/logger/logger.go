package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// New creates the application logger for the selected environment.
func New(environment string) (*zap.Logger, error) {
	switch environment {
	case "", "development":
		return zap.NewDevelopment()
	case "production":
		return zap.NewProduction()
	default:
		return nil, fmt.Errorf("invalid MOBICODE_SERVER_ENV %q (use development or production)", environment)
	}
}
