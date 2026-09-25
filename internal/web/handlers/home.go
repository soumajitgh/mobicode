package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/web/components"
	"github.com/soumajitgh/mobicode/internal/web/pages"
)

// Handler serves the browser-facing pages and HTMX fragments.
type Handler struct {
	HealthService *health.Service
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	templ.Handler(pages.Home()).ServeHTTP(w, r)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	status := h.HealthService.Status(r.Context())
	templ.Handler(components.Status(status)).ServeHTTP(w, r)
}
