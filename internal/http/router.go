package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/soumajitgh/mobicode/internal/http/middleware"
)

// NewRouter builds the API's HTTP handler.
func NewRouter() http.Handler {
	r := chi.NewRouter()
	middleware.Apply(r)
	registerRoutes(r)
	return r
}
