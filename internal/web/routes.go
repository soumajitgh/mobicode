package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/soumajitgh/mobicode/internal/web/handlers"
	webmiddleware "github.com/soumajitgh/mobicode/internal/web/middleware"
	"github.com/soumajitgh/mobicode/public"
)

// RegisterRoutes mounts the browser UI and its static assets.
func RegisterRoutes(r chi.Router, h *handlers.Handler, devAssets bool, auth *handlers.Auth, onboarding *handlers.Onboarding) {
	assets := http.FileServer(http.FS(public.Files))
	if devAssets {
		assets = http.FileServer(http.Dir("./public"))
	}
	r.Handle("/assets/*", http.StripPrefix("/assets/", assets))
	onboarding.RegisterRoutes(r)
	auth.RegisterRoutes(r)
	requireUser := webmiddleware.RequireUser(auth.Sessions, auth.Users)
	r.With(requireUser).Get("/", h.Home)
	r.With(requireUser).Get("/partials/status", h.Status)
}
