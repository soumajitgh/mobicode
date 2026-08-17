package health

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

// Module provides health checks and registers their routes.
var Module = fx.Module(
	"health",
	fx.Provide(NewHandler),
	fx.Invoke(func(mux *chi.Mux, handler *Handler) {
		mux.Get("/health", handler.Check)
	}),
)
