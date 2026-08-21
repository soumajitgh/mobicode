package setup

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

// Module mounts the one-time browser pairing flow.
var Module = fx.Module(
	"setup",
	fx.Provide(NewRepository, NewService, NewHandler),
	fx.Invoke(func(router *chi.Mux, handler *Handler) {
		router.Mount("/setup", handler.Router())
	}),
	fx.Invoke(func(_ *http.Server) {}),
)
