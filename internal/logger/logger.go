// Package logger provides application logging.
package logger

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/soumajitgh/mobicode/internal/config"
)

// New builds an environment-aware Zap logger and flushes it at shutdown.
func New(lc fx.Lifecycle, cfg *config.Config) (*zap.Logger, error) {
	var (
		log *zap.Logger
		err error
	)
	if cfg.Env == "production" {
		log, err = zap.NewProduction()
	} else {
		developmentConfig := zap.NewDevelopmentConfig()
		developmentConfig.EncoderConfig.TimeKey = "time"
		developmentConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
		developmentConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		log, err = developmentConfig.Build(zap.WithCaller(false))
	}
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			// stdout/stderr can report an unsupported sync error; it must not fail shutdown.
			_ = log.Sync()
			return nil
		},
	})

	return log, nil
}
