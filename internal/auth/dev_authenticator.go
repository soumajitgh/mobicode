package auth

import (
	"context"

	"go.uber.org/zap"
)

// DevAuthenticator is a development-only authenticator that accepts both raw user IDs and standard access JWTs.
type DevAuthenticator struct {
	jwt *JWTService
	log *zap.Logger
}

// NewDevAuthenticator constructs a new DevAuthenticator.
func NewDevAuthenticator(jwt *JWTService, log *zap.Logger) *DevAuthenticator {
	return &DevAuthenticator{jwt: jwt, log: log}
}

// Authenticate maps the bearer token to a Principal.
// If the token is a valid JWT, it extracts the UserID from claims.
// Otherwise, it interprets the raw token directly as the UserID.
func (a *DevAuthenticator) Authenticate(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrUnauthenticated
	}

	if a.jwt != nil {
		if claims, err := a.jwt.VerifyAccessToken(token); err == nil && claims.UserID != "" {
			return Principal{
				UserID:    claims.UserID,
				SessionID: claims.SessionID,
			}, nil
		}
	}

	return Principal{
		UserID:    token,
		SessionID: "dev",
	}, nil
}
