package logger

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// NewFxLogger adapts Zap to Fx's lifecycle event logger.
func NewFxLogger(log *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: log}
}
