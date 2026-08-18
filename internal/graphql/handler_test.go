package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/user"
)

type memoryUserRepository struct {
	users map[string]*user.User
}

func (r *memoryUserRepository) FindByID(_ context.Context, id string) (*user.User, error) {
	found, ok := r.users[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	return found, nil
}

func (r *memoryUserRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	for _, found := range r.users {
		if found.Email == email {
			return found, nil
		}
	}
	return nil, user.ErrNotFound
}

func (r *memoryUserRepository) Create(_ context.Context, value *user.User) error {
	value.CreatedAt = time.Now().UTC()
	r.users[value.ID] = value
	return nil
}

type memoryRefreshTokens struct {
	tokens map[string]*auth.RefreshToken
}

func (r *memoryRefreshTokens) Create(_ context.Context, token *auth.RefreshToken) error {
	r.tokens[token.TokenHash] = token
	return nil
}

func (r *memoryRefreshTokens) FindActiveByHash(_ context.Context, hash string) (*auth.RefreshToken, error) {
	token, ok := r.tokens[hash]
	if !ok || token.RevokedAt != nil || !token.ExpiresAt.After(time.Now()) {
		return nil, auth.ErrInvalidRefreshToken
	}
	return token, nil
}

func (r *memoryRefreshTokens) Rotate(_ context.Context, hash string, replacement *auth.RefreshToken) error {
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

func TestServerRegistersAndQueriesUser(t *testing.T) {
	userRepo := &memoryUserRepository{users: make(map[string]*user.User)}
	refreshRepo := &memoryRefreshTokens{tokens: make(map[string]*auth.RefreshToken)}
	jwt, err := auth.NewJWTService(&config.Config{JWTSecret: "test-secret-with-at-least-thirty-two-bytes"})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}

	userService := user.NewService(userRepo, zap.NewNop())
	authService := auth.NewService(userService, refreshRepo, auth.NewPasswordService(), jwt, zap.NewNop())

	server := NewServer(NewResolver(user.NewResolver(userService), auth.NewResolver(authService)), zap.NewNop())

	registered := executeGraphQL(server, `mutation { register(name: "Zoravix", email: "z@fotopick.in", password: "password-123") { accessToken user { id name email } } }`)
	if registered.Code != http.StatusOK {
		t.Fatalf("register status = %d, want %d: %s", registered.Code, http.StatusOK, registered.Body.String())
	}

	var registerResponse struct {
		Data struct {
			Register struct {
				AccessToken string `json:"accessToken"`
				User        struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"user"`
			} `json:"register"`
		} `json:"data"`
	}
	if err := json.NewDecoder(registered.Body).Decode(&registerResponse); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	userID := registerResponse.Data.Register.User.ID
	if userID == "" {
		t.Fatalf("register response did not contain a user ID: %s", registered.Body.String())
	}

	found := executeGraphQL(server, `query { user(id: "`+userID+`") { id name email createdAt } }`)
	if found.Code != http.StatusOK {
		t.Fatalf("query status = %d, want %d: %s", found.Code, http.StatusOK, found.Body.String())
	}
	if !json.Valid(found.Body.Bytes()) {
		t.Fatalf("query response is not JSON: %s", found.Body.String())
	}

	// Test querying empty ID returns null user without error
	emptyQuery := executeGraphQL(server, `query { user(id: "") { id name } }`)
	if emptyQuery.Code != http.StatusOK {
		t.Fatalf("empty user id query status = %d, want %d: %s", emptyQuery.Code, http.StatusOK, emptyQuery.Body.String())
	}
	var emptyResponse struct {
		Data struct {
			User *struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
		Errors []any `json:"errors"`
	}
	if err := json.NewDecoder(emptyQuery.Body).Decode(&emptyResponse); err != nil {
		t.Fatalf("decode empty query response: %v", err)
	}
	if emptyResponse.Data.User != nil {
		t.Fatalf("expected nil user for empty id, got: %#v", emptyResponse.Data.User)
	}
	if len(emptyResponse.Errors) > 0 {
		t.Fatalf("expected no errors for empty id, got: %#v", emptyResponse.Errors)
	}
}

func executeGraphQL(server http.Handler, query string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":`+strconv.Quote(query)+`}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}
