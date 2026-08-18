package auth

import (
	"net/http"
	"strings"
)

// Middleware reads the Authorization header, authenticates the Bearer token using Authenticator,
// and attaches a Principal to the request context if valid.
// Missing, invalid, or expired tokens do not block execution; request handling continues anonymously.
func Middleware(authenticator Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}
			principal, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Authenticate is an alias for Middleware.
func Authenticate(authenticator Authenticator) func(http.Handler) http.Handler {
	return Middleware(authenticator)
}

func bearerToken(r *http.Request) string {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
