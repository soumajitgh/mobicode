package web

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/soumajitgh/mobicode/internal/store/model"
)

type userKey struct{}

func CurrentUser(ctx context.Context) *model.User {
	user, _ := ctx.Value(userKey{}).(*model.User)
	return user
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
