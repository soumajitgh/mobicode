package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/user"
)

type memoryUsers struct {
	byID    map[string]*user.User
	byEmail map[string]*user.User
}

func (r *memoryUsers) FindByID(_ context.Context, id string) (*user.User, error) {
	account, ok := r.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return account, nil
}

func (r *memoryUsers) FindByEmail(_ context.Context, email string) (*user.User, error) {
	account, ok := r.byEmail[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	return account, nil
}

func (r *memoryUsers) Create(_ context.Context, account *user.User) error {
	r.byID[account.ID] = account
	r.byEmail[account.Email] = account
	return nil
}

type memoryRefreshTokens struct{ tokens map[string]*RefreshToken }

func (r *memoryRefreshTokens) Create(_ context.Context, token *RefreshToken) error {
	r.tokens[token.TokenHash] = token
	return nil
}

func (r *memoryRefreshTokens) FindActiveByHash(_ context.Context, hash string) (*RefreshToken, error) {
	token, ok := r.tokens[hash]
	if !ok || token.RevokedAt != nil || !token.ExpiresAt.After(time.Now()) {
		return nil, ErrInvalidRefreshToken
	}
	return token, nil
}

func (r *memoryRefreshTokens) Rotate(_ context.Context, hash string, replacement *RefreshToken) error {
	token, err := r.FindActiveByHash(context.Background(), hash)
	if err != nil {
		return err
	}
	now := time.Now()
	token.RevokedAt = &now
	r.tokens[replacement.TokenHash] = replacement
	return nil
}

func (r *memoryRefreshTokens) RevokeByHash(_ context.Context, hash string) error {
	token, err := r.FindActiveByHash(context.Background(), hash)
	if err != nil {
		return err
	}
	now := time.Now()
	token.RevokedAt = &now
	return nil
}

func TestServiceRegisterLoginRefreshLogout(t *testing.T) {
	users := &memoryUsers{byID: map[string]*user.User{}, byEmail: map[string]*user.User{}}
	refresh := &memoryRefreshTokens{tokens: map[string]*RefreshToken{}}
	jwt, err := NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}
	userService := user.NewService(users, zap.NewNop())
	service := NewService(userService, refresh, NewPasswordService(), jwt, zap.NewNop())

	registered, err := service.Register(context.Background(), "Zoravix", "z@fotopick.in", "password-123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	claims, err := jwt.Verify(registered.AccessToken)
	if err != nil || claims.UserID != registered.User.ID {
		t.Fatalf("registered access token claims = %#v, %v", claims, err)
	}

	loggedIn, err := service.Login(context.Background(), "z@fotopick.in", "password-123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if _, err := service.Login(context.Background(), "z@fotopick.in", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password error = %v", err)
	}

	rotated, err := service.Refresh(context.Background(), loggedIn.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if _, err := service.Refresh(context.Background(), loggedIn.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("reused refresh token error = %v", err)
	}
	if err := service.Logout(context.Background(), rotated.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := service.Refresh(context.Background(), rotated.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("revoked refresh token error = %v", err)
	}
}

func TestAuthenticateAddsUserToContext(t *testing.T) {
	jwt, err := NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}
	token, err := jwt.Sign("user-1")
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Authenticate(jwt)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok || userID != "user-1" {
			t.Fatalf("context user ID = %q, %v", userID, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
