package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/alexedwards/scs/v2"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

type Auth struct {
	Service  *auth.Service
	Sessions *scs.SessionManager
	Users    repository.UserRepository
}

func NewAuth(service *auth.Service, sessions *scs.SessionManager, users repository.UserRepository) *Auth {
	return &Auth{Service: service, Sessions: sessions, Users: users}
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
