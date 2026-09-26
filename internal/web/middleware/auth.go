package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

type userKey struct{}

func CurrentUser(ctx context.Context) *model.User {
	user, _ := ctx.Value(userKey{}).(*model.User)
	return user
}

func RequireUser(sessions *scs.SessionManager, users repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := sessions.GetInt(r.Context(), "user_id")
			if id > 0 {
				user, err := users.FindByID(r.Context(), uint(id))
				if err == nil && user.SessionVersion == sessions.GetInt64(r.Context(), "session_version") {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
					return
				}
				_ = sessions.Destroy(r.Context())
			}
			nextPath := r.URL.RequestURI()
			if !safeNext(nextPath) {
				nextPath = "/"
			}
			http.Redirect(w, r, "/auth/login?next="+url.QueryEscape(nextPath), http.StatusSeeOther)
		})
	}
}

func safeNext(next string) bool {
	return strings.HasPrefix(next, "/") && !strings.HasPrefix(next, "//") && !strings.ContainsAny(next, "\\\r\n")
}

func RedirectTarget(next string) string {
	if safeNext(next) {
		return next
	}
	return "/"
}
