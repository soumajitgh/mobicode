package auth

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/config"
)

func TestNewAuthenticatorSelection(t *testing.T) {
	jwtSvc, err := NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}

	// 1. DevAuthEnabled=false -> JWTAuthenticator
	cfgProd := &config.Config{Env: "production", DevAuthEnabled: false}
	authProd, err := NewAuthenticator(cfgProd, jwtSvc, zap.NewNop())
	if err != nil {
		t.Fatalf("expected no error for prod config, got: %v", err)
	}
	if _, ok := authProd.(*JWTAuthenticator); !ok {
		t.Fatalf("expected *JWTAuthenticator for prod config, got %T", authProd)
	}

	// 2. DevAuthEnabled=true in development -> DevAuthenticator
	cfgDev := &config.Config{Env: "development", DevAuthEnabled: true}
	authDev, err := NewAuthenticator(cfgDev, jwtSvc, zap.NewNop())
	if err != nil {
		t.Fatalf("expected no error for dev config, got: %v", err)
	}
	if _, ok := authDev.(*DevAuthenticator); !ok {
		t.Fatalf("expected *DevAuthenticator for dev config, got %T", authDev)
	}

	// 3. Safety rail 1: DevAuthEnabled=true in production -> error
	cfgFail := &config.Config{Env: "production", DevAuthEnabled: true}
	_, err = NewAuthenticator(cfgFail, jwtSvc, zap.NewNop())
	if err == nil {
		t.Fatalf("expected error when DEV_AUTH_ENABLED=true in production")
	}
}

func TestDevAuthenticator(t *testing.T) {
	jwtSvc, _ := NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	devAuth := NewDevAuthenticator(jwtSvc, zap.NewNop())

	// Empty token -> error
	_, err := devAuth.Authenticate(context.Background(), "")
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated for empty token, got: %v", err)
	}

	// Raw token -> Principal with UserID=token
	p, err := devAuth.Authenticate(context.Background(), "user-uuid-1234")
	if err != nil {
		t.Fatalf("dev authenticate error: %v", err)
	}
	if p.UserID != "user-uuid-1234" || p.SessionID != "dev" {
		t.Fatalf("got principal %+v, want UserID=user-uuid-1234, SessionID=dev", p)
	}

	// JWT token -> Principal with UserID=claims.UserID
	tokenStr, _ := jwtSvc.SignAccessToken("user-jwt-5678")
	pJWT, err := devAuth.Authenticate(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("dev authenticate with JWT error: %v", err)
	}
	if pJWT.UserID != "user-jwt-5678" {
		t.Fatalf("got principal UserID=%q, want user-jwt-5678", pJWT.UserID)
	}
}
