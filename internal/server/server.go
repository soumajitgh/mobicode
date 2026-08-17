// Package server owns HTTP routing and lifecycle management.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/config"
)

// NewRouter creates the shared HTTP router.
func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	useDefaultMiddleware(r)
	return r
}

// NewHTTPServer starts and gracefully stops the HTTP server with the Fx app.
func NewHTTPServer(lc fx.Lifecycle, mux *chi.Mux, cfg *config.Config, log *zap.Logger) *http.Server {
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("starting http server", zap.String("addr", srv.Addr), zap.String("env", cfg.Env))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("stopping http server")
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		},
	})

	return srv
}
