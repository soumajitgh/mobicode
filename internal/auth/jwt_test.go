package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/soumajitgh/mobicode/internal/config"
)

func TestJWTServiceSignAndVerify(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"}
	svc, err := NewJWTService(cfg)
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}

	tokenStr, err := svc.SignAccessToken("user-123")
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	claims, err := svc.VerifyAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("verify access token: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("claims.UserID = %q, want %q", claims.UserID, "user-123")
	}
}

func TestJWTServiceShortSecretError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "short"}
	_, err := NewJWTService(cfg)
	if err == nil {
		t.Fatalf("expected error for short JWT secret, got nil")
	}
}

func TestJWTServiceVerifyInvalidTokens(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"}
	svc, err := NewJWTService(cfg)
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}

	// Malformed token
	if _, err := svc.VerifyAccessToken("invalid-token-string"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for malformed token, got: %v", err)
	}

	// Signed with wrong secret
	otherSvc, _ := NewJWTService(&config.Config{JWTSecret: "other-secret-with-at-least-thirty-two-bytes"})
	otherToken, _ := otherSvc.SignAccessToken("user-123")
	if _, err := svc.VerifyAccessToken(otherToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for wrong secret, got: %v", err)
	}

	// Expired token
	now := time.Now().Add(-1 * time.Hour)
	expiredClaims := Claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}
	expiredToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte(cfg.JWTSecret))
	if _, err := svc.VerifyAccessToken(expiredToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for expired token, got: %v", err)
	}
}
