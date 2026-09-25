package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/soumajitgh/mobicode/internal/web/handlers"
	"github.com/soumajitgh/mobicode/public"
)

// RegisterRoutes mounts the browser UI and its static assets.
func RegisterRoutes(r chi.Router, h *handlers.Handler, devAssets bool) {
	assets := http.FileServer(http.FS(public.Files))
	if devAssets {
		assets = http.FileServer(http.Dir("./public"))
	}
	r.Handle("/assets/*", http.StripPrefix("/assets/", assets))
	r.Get("/", h.Home)
	r.Get("/partials/status", h.Status)
}
