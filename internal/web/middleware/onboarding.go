package middleware

import (
	"context"
	"net/http"
	"strings"
)

type OnboardingState interface {
	HasUsers(context.Context) (bool, error)
	IsWizardInProgress(*http.Request) bool
}

func GateOnboarding(state OnboardingState) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Always allow static assets and health check
			if strings.HasPrefix(path, "/assets/") || path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			hasUsers, err := state.HasUsers(r.Context())
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			isOnboardingPath := path == "/onboarding" || strings.HasPrefix(path, "/onboarding/")

			if !hasUsers || state.IsWizardInProgress(r) {
				if isOnboardingPath {
					next.ServeHTTP(w, r)
					return
				}
				if strings.HasPrefix(path, "/mobile") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte(`{"errors":[{"message":"server onboarding incomplete"}]}`))
					return
				}
				http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
				return
			}

			if isOnboardingPath {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
