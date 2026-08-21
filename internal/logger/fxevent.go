package logger

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/config"
)

// NewFxLogger suppresses noisy dependency-injection events during development.
// Production retains structured Fx lifecycle logs for startup diagnostics.
func NewFxLogger(cfg *config.Config, log *zap.Logger) fxevent.Logger {
	if cfg.Env == "development" {
		return fxevent.NopLogger
	}
	return &fxevent.ZapLogger{Logger: log}
}
