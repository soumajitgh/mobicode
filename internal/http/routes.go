package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/soumajitgh/mobicode/internal/http/handler"
)

func registerRoutes(r chi.Router) {
	r.Get("/healthz", handler.Health)
}
