package auth

import (
	"errors"
	"net/http"
)

// RequireOwner validates a NIP-98 proof and attaches the configured owner principal.
// GraphQL is intentionally all-private: invalid authentication stops execution at HTTP.
func RequireOwner(verifier *NIP98Verifier, owner *OwnerService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, err := owner.Owner(r.Context())
			if err != nil {
				if errors.Is(err, ErrSetupRequired) {
					http.Error(w, "server setup required", http.StatusServiceUnavailable)
					return
				}
				http.Error(w, "authentication unavailable", http.StatusInternalServerError)
				return
			}
			proof, err := verifier.Verify(r)
			if err != nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
			if proof.PublicKey != identity.PublicKey {
				http.Error(w, "authenticated key is not authorized", http.StatusForbidden)
				return
			}
			if err := owner.ClaimReplay(r.Context(), proof.ID, proof.ExpiresAt); err != nil {
				if errors.Is(err, ErrReplay) {
					http.Error(w, "authentication proof already used", http.StatusUnauthorized)
					return
				}
				http.Error(w, "authentication unavailable", http.StatusInternalServerError)
				return
			}
			ctx := WithPrincipal(r.Context(), Principal{PublicKey: proof.PublicKey})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
