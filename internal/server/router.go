package server

import (
	"github.com/go-chi/chi/v5"
)

// NewRouter creates the shared HTTP router with default middleware.
func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	useDefaultMiddleware(r)
	return r
}
