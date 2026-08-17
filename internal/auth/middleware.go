package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey struct{}

// UserIDFromContext returns the authenticated user ID when one is present.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKey{}).(string)
	return userID, ok
}

// Authenticate verifies a supplied bearer token and adds its user ID to the request context.
// Requests without credentials are allowed so public GraphQL operations remain reachable.
func Authenticate(jwt *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := jwt.Verify(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, claims.UserID)))
		})
	}
}

// RequireAuth rejects requests that do not have a valid bearer token.
func RequireAuth(jwt *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			claims, err := jwt.Verify(token)
			if token == "" || err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, claims.UserID)))
		})
	}
}

func bearerToken(r *http.Request) string {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
