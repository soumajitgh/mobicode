package web

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/a-h/templ"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/web/pages"
)

func (a *Auth) form(w http.ResponseWriter, r *http.Request, title, action, message string, recovery bool, status int) {
	csrf := a.csrf(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	templ.Handler(pages.AuthForm(title, action, redirectTarget(r.FormValue("next")), csrf, message, recovery)).ServeHTTP(w, r)
}

func (a *Auth) submit(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			host = r.RemoteAddr
		}
		if !a.limiter.Allow(kind + ":" + host) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		if !a.checkCSRF(r) {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}
		email, password, token := r.PostFormValue("email"), r.PostFormValue("password"), r.PostFormValue("recovery_token")
		var user *model.User
		var err error
		switch kind {
		case "register":
			user, err = a.Service.Register(r.Context(), email, password, token)
		case "login":
			user, err = a.Service.Login(r.Context(), email, password)
		case "reset":
			err = a.Service.ResetPassword(r.Context(), email, password, token)
		}
		if err != nil {
			message := "Invalid input"
			if errors.Is(err, auth.ErrInvalidCredentials) {
				message = "Invalid email or password"
			}
			if errors.Is(err, auth.ErrDuplicateEmail) {
				message = "Email already registered"
			}
			if errors.Is(err, auth.ErrInvalidRecoveryToken) {
				message = "Invalid recovery token"
			}
			if !errors.Is(err, auth.ErrInvalidCredentials) && !errors.Is(err, auth.ErrDuplicateEmail) && !errors.Is(err, auth.ErrInvalidRecoveryToken) && !errors.Is(err, auth.ErrInvalidInput) {
				http.Error(w, "internal error", 500)
				return
			}
			title := map[string]string{"register": "Register", "login": "Log in", "reset": "Reset password"}[kind]
			a.form(w, r, title, "/auth/"+map[string]string{"register": "register", "login": "login", "reset": "reset-password"}[kind], message, kind != "login", http.StatusUnprocessableEntity)
			return
		}
		if kind == "reset" {
			_ = a.Sessions.Destroy(r.Context())
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
		if err := a.Sessions.RenewToken(r.Context()); err != nil {
			http.Error(w, "session error", 500)
			return
		}
		a.Sessions.Put(r.Context(), "user_id", int(user.ID))
		a.Sessions.Put(r.Context(), "session_version", user.SessionVersion)
		http.Redirect(w, r, redirectTarget(r.PostFormValue("next")), http.StatusSeeOther)
	}
}

type rateEntry struct {
	count int
	start time.Time
}
type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
}

func (l *rateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e := l.entries[key]
	if e == nil || now.Sub(e.start) > time.Minute {
		e = &rateEntry{start: now}
		l.entries[key] = e
	}
	e.count++
	return e.count <= 10
}
