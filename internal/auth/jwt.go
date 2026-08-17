package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/soumajitgh/mobicode/internal/config"
)

const accessTokenLifetime = 15 * time.Minute

// Claims identifies the authenticated user.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTService signs and verifies short-lived access tokens.
type JWTService struct {
	secret []byte
}

func NewJWTService(cfg *config.Config) (*JWTService, error) {
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	return &JWTService{secret: []byte(cfg.JWTSecret)}, nil
}

func (s *JWTService) Sign(userID string) (string, error) {
	now := time.Now()
	claims := Claims{UserID: userID, RegisteredClaims: jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenLifetime)),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *JWTService) Verify(token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %q", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid || claims.UserID == "" {
		return nil, ErrUnauthorized
	}
	return claims, nil
}
