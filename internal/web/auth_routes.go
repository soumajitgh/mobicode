package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *Auth) RegisterRoutes(r chi.Router) {
	r.Get("/auth/register", func(w http.ResponseWriter, req *http.Request) {
		a.form(w, req, "Register", "/auth/register", "", true, 200)
	})
	r.Get("/auth/login", func(w http.ResponseWriter, req *http.Request) {
		a.form(w, req, "Log in", "/auth/login", "", false, 200)
	})
	r.Get("/auth/reset-password", func(w http.ResponseWriter, req *http.Request) {
		a.form(w, req, "Reset password", "/auth/reset-password", "", true, 200)
	})
	r.Post("/auth/register", a.submit("register"))
	r.Post("/auth/login", a.submit("login"))
	r.Post("/auth/reset-password", a.submit("reset"))
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
