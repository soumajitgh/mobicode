package auth

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/config"
)

// Authenticator validates a raw bearer token string and returns an authenticated Principal.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (Principal, error)
}

// NewAuthenticator constructs an Authenticator implementation based on application configuration.
// If DevAuthEnabled is true, it verifies that environment is "development" before constructing a DevAuthenticator.
func NewAuthenticator(cfg *config.Config, jwt *JWTService, log *zap.Logger) (Authenticator, error) {
	if cfg.DevAuthEnabled {
		if cfg.Env != "development" {
			return nil, fmt.Errorf("DEV_AUTH_ENABLED may only be used in development environment (current: %s)", cfg.Env)
		}

		log.Warn("WARN dev authentication enabled; bearer tokens are interpreted as user IDs")
		return NewDevAuthenticator(jwt, log), nil
	}

	return &JWTAuthenticator{jwt: jwt}, nil
}
