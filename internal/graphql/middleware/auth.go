package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/soumajitgh/mobicode/internal/pairing"
	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

type identityKey struct{}

type Identity struct {
	WebUser       *model.User
	MobileUser    *repository.MobileIdentity
	InvalidBearer bool
	AuthError     error
}

func FromContext(ctx context.Context) Identity {
	v, _ := ctx.Value(identityKey{}).(Identity)
	return v
}

func Authenticate(sessions *scs.SessionManager, users repository.UserRepository, mobile repository.MobileRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var identity Identity
			if header := r.Header.Get("Authorization"); header != "" {
				parts := strings.Fields(header)
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
					identity.InvalidBearer = true
				} else {
					user, err := mobile.FindMobileIdentity(r.Context(), pairing.HashToken(parts[1]))
					if err == nil {
						identity.MobileUser = user
					} else if err == sql.ErrNoRows {
						identity.InvalidBearer = true
					} else {
						identity.AuthError = err
					}
				}
			} else if id := sessions.GetInt(r.Context(), "user_id"); id > 0 {
				user, err := users.FindByID(r.Context(), uint(id))
				if err == nil && user.SessionVersion == sessions.GetInt64(r.Context(), "session_version") {
					identity.WebUser = user
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), identityKey{}, identity)))
		})
	}
}
