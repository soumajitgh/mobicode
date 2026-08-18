package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/soumajitgh/mobicode/internal/config"
)

func TestMiddlewareValidTokenAttachesPrincipal(t *testing.T) {
	jwtSvc, err := NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}
	authenticator := NewJWTAuthenticator(jwtSvc)
	token, err := jwtSvc.SignAccessToken("user-999")
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	var capturedPrincipal Principal
	var capturedOK bool

	handler := Middleware(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPrincipal, capturedOK = PrincipalFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !capturedOK {
		t.Fatalf("expected principal in context")
	}
	if capturedPrincipal.UserID != "user-999" {
		t.Fatalf("principal.UserID = %q, want user-999", capturedPrincipal.UserID)
	}
}

func TestMiddlewareMissingOrInvalidTokenIsNonBlocking(t *testing.T) {
	jwtSvc, err := NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}
	authenticator := NewJWTAuthenticator(jwtSvc)

	tests := []struct {
		name   string
		header string
	}{
		{name: "no header", header: ""},
		{name: "malformed token", header: "Bearer invalid-jwt-string"},
		{name: "invalid auth scheme", header: "Basic dXNlcjpwYXNz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedOK bool
			handler := Middleware(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, capturedOK = PrincipalFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, "/query", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (non-blocking)", rec.Code, http.StatusOK)
			}
			if capturedOK {
				t.Fatalf("expected NO principal in context for %s", tt.name)
			}
		})
	}
}

func TestMiddlewareExpiredTokenIsNonBlocking(t *testing.T) {
	jwtSvc, err := NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}
	authenticator := NewJWTAuthenticator(jwtSvc)

	now := time.Now().Add(-1 * time.Hour)
	expiredClaims := Claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}
	expiredToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("test-secret-with-at-least-thirty-two-bytes"))

	var capturedOK bool
	handler := Middleware(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, capturedOK = PrincipalFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (non-blocking)", rec.Code, http.StatusOK)
	}
	if capturedOK {
		t.Fatalf("expected NO principal in context for expired token")
	}
}
