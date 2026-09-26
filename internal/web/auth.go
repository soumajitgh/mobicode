package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
	"github.com/soumajitgh/mobicode/internal/web/pages"
)

type userKey struct{}

func CurrentUser(ctx context.Context) *model.User {
	user, _ := ctx.Value(userKey{}).(*model.User)
	return user
}

type Auth struct {
	Service  *auth.Service
	Sessions *scs.SessionManager
	Users    repository.UserRepository
	limiter  *rateLimiter
}

func NewAuth(service *auth.Service, sessions *scs.SessionManager, users repository.UserRepository) *Auth {
	return &Auth{Service: service, Sessions: sessions, Users: users, limiter: &rateLimiter{entries: make(map[string]*rateEntry)}}
}

func (a *Auth) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := a.Sessions.GetInt(r.Context(), "user_id")
		if id > 0 {
			user, err := a.Users.FindByID(r.Context(), uint(id))
			if err == nil && user.SessionVersion == a.Sessions.GetInt64(r.Context(), "session_version") {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
				return
			}
			_ = a.Sessions.Destroy(r.Context())
		}
		nextPath := r.URL.RequestURI()
		if !safeNext(nextPath) {
			nextPath = "/"
		}
		http.Redirect(w, r, "/auth/login?next="+url.QueryEscape(nextPath), http.StatusSeeOther)
	})
}

func safeNext(next string) bool {
	return strings.HasPrefix(next, "/") && !strings.HasPrefix(next, "//") && !strings.ContainsAny(next, "\\\r\n")
}

func redirectTarget(next string) string {
	if safeNext(next) {
		return next
	}
	return "/"
}

func (a *Auth) csrf(ctx context.Context) string {
	token := a.Sessions.GetString(ctx, "csrf_token")
	if token == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return ""
		}
		token = base64.RawURLEncoding.EncodeToString(b)
		a.Sessions.Put(ctx, "csrf_token", token)
	}
	return token
}

func (a *Auth) checkCSRF(r *http.Request) bool {
	expected := a.Sessions.GetString(r.Context(), "csrf_token")
	supplied := r.PostFormValue("csrf_token")
	return expected != "" && subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) == 1
}

func (a *Auth) form(w http.ResponseWriter, r *http.Request, title, action, message string, recovery bool, status int) {
	csrf := a.csrf(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	templ.Handler(pages.AuthForm(title, action, redirectTarget(r.FormValue("next")), csrf, message, recovery)).ServeHTTP(w, r)
}

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
