package auth

import "context"

// JWTAuthenticator verifies standard access JWTs using JWTService.
type JWTAuthenticator struct {
	jwt *JWTService
}

// NewJWTAuthenticator constructs a new JWTAuthenticator.
func NewJWTAuthenticator(jwt *JWTService) *JWTAuthenticator {
	return &JWTAuthenticator{jwt: jwt}
}

// Authenticate verifies the token as a JWT and maps claims to a Principal.
func (a *JWTAuthenticator) Authenticate(ctx context.Context, token string) (Principal, error) {
	claims, err := a.jwt.VerifyAccessToken(token)
	if err != nil {
		return Principal{}, err
	}

	return Principal{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
	}, nil
}
