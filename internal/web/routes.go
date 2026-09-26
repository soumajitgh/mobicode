package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/soumajitgh/mobicode/internal/web/handlers"
	"github.com/soumajitgh/mobicode/public"
)

// RegisterRoutes mounts the browser UI and its static assets.
func RegisterRoutes(r chi.Router, h *handlers.Handler, devAssets bool, auth *Auth) {
	assets := http.FileServer(http.FS(public.Files))
	if devAssets {
		assets = http.FileServer(http.Dir("./public"))
	}
	r.Handle("/assets/*", http.StripPrefix("/assets/", assets))
	auth.RegisterRoutes(r)
	r.With(auth.RequireUser).Get("/", h.Home)
	r.With(auth.RequireUser).Get("/partials/status", h.Status)
}
