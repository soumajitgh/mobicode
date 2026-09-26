package web

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"

	"github.com/soumajitgh/mobicode/internal/http/middleware"
)

func (a *Auth) RegisterRoutes(r chi.Router) {
	// A shared store keeps a separate bucket for each endpoint and client IP.
	limiter := middleware.NewKeyedRateLimiter(rate.Every(6*time.Second), 10, 2*time.Minute)
	limitAuth := middleware.RateLimit(limiter, func(r *http.Request) string {
		return r.Method + ":" + r.URL.Path + ":" + middleware.ClientIP(r)
	})

	r.Get("/auth/register", func(w http.ResponseWriter, req *http.Request) {
		a.form(w, req, "Register", "/auth/register", "", true, 200)
	})
	r.Get("/auth/login", func(w http.ResponseWriter, req *http.Request) {
		a.form(w, req, "Log in", "/auth/login", "", false, 200)
	})
	r.Get("/auth/reset-password", func(w http.ResponseWriter, req *http.Request) {
		a.form(w, req, "Reset password", "/auth/reset-password", "", true, 200)
	})
	r.With(limitAuth).Post("/auth/register", a.submit("register"))
	r.With(limitAuth).Post("/auth/login", a.submit("login"))
	r.With(limitAuth).Post("/auth/reset-password", a.submit("reset"))
	r.With(a.RequireUser).Post("/auth/logout", func(w http.ResponseWriter, req *http.Request) {
		if !a.checkCSRF(req) {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}
		if err := a.Sessions.Destroy(req.Context()); err != nil {
			http.Error(w, "session error", 500)
			return
		}
		http.Redirect(w, req, "/auth/login", http.StatusSeeOther)
	})
}
