package user

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

// Module provides the user feature and registers its routes.
var Module = fx.Module(
	"user",
	fx.Provide(NewRepository, NewService, NewHandler),
	fx.Invoke(func(mux *chi.Mux, handler *Handler) {
		handler.RegisterRoutes(mux)
	}),
)
