package middleware

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Apply installs middleware shared by all API routes.
func Apply(r chi.Router, log *zap.Logger) {
	r.Use(chimiddleware.RequestID)
	r.Use(Logging(log))
	r.Use(chimiddleware.Recoverer)
}
