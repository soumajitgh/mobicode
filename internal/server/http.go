package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/soumajitgh/mobicode/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// New builds the configured HTTP server.
func New(cfg config.Config, router *gin.Engine) *http.Server {
	return &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), Handler: router, ReadTimeout: cfg.Server.ReadTimeout, WriteTimeout: cfg.Server.WriteTimeout, IdleTimeout: cfg.Server.IdleTimeout}
}

// registerLifecycle starts and gracefully stops the HTTP server.
func registerLifecycle(lc fx.Lifecycle, srv *http.Server, cfg config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{OnStart: func(context.Context) error {
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("http server failed", zap.Error(err))
			}
		}()
		return nil
	}, OnStop: func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}})
}

var Module = fx.Module("server", fx.Provide(New), fx.Invoke(registerLifecycle))
