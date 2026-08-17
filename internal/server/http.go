package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/soumajitgh/mobicode/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// New builds the configured HTTP server.
func New(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           handler,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
	}
}

// registerLifecycle starts and gracefully stops the HTTP server.
func registerLifecycle(lc fx.Lifecycle, srv *http.Server, cfg config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{OnStart: func(context.Context) error {
		listener, err := net.Listen("tcp", srv.Addr)
		if err != nil {
			return fmt.Errorf("listen on %s: %w", srv.Addr, err)
		}
		go func() {
			if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
