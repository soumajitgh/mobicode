package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/alexedwards/scs/v2"

	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/web/components"
	"github.com/soumajitgh/mobicode/internal/web/pages"
)

// Handler serves the browser-facing pages and HTMX fragments.
type Handler struct {
	HealthService *health.Service
	Sessions      *scs.SessionManager
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	csrf := ""
	if h.Sessions != nil {
		csrf = h.Sessions.GetString(r.Context(), "csrf_token")
	}
	templ.Handler(pages.Home(csrf)).ServeHTTP(w, r)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	status := h.HealthService.Status(r.Context())
	templ.Handler(components.Status(status)).ServeHTTP(w, r)
}
