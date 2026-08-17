package observability

import (
	"context"
	"fmt"
	"strings"

	"mobicode/apps/server/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewLogger builds the application's configured Zap logger.
func NewLogger(cfg config.Config) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		return nil, fmt.Errorf("parse LOG_LEVEL: %w", err)
	}
	loggerConfig := zap.NewProductionConfig()
	if cfg.Log.Development {
		loggerConfig = zap.NewDevelopmentConfig()
	}
	loggerConfig.Level = level
	return loggerConfig.Build()
}

// registerLifecycle flushes logs during application shutdown.
func registerLifecycle(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		err := logger.Sync()
		if err != nil && !strings.Contains(err.Error(), "invalid argument") {
			return err
		}
		return nil
	}})
}

var Module = fx.Module("observability", fx.Provide(NewLogger), fx.Invoke(registerLifecycle))
