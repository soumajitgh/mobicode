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

func TestServerCreatesAndQueriesUser(t *testing.T) {
	repo := &memoryUserRepository{users: make(map[string]*user.User)}
	service := user.NewService(repo, zap.NewNop())
	server := NewServer(&Resolver{user: user.NewResolver(service)}, zap.NewNop())

	created := executeGraphQL(server, `mutation { createUser(name: "Zoravix", email: "z@fotopick.in") { id name email } }`)
	if created.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d: %s", created.Code, http.StatusOK, created.Body.String())
	}
	var createResponse struct {
		Data struct {
			CreateUser struct {
				ID string `json:"id"`
			} `json:"createUser"`
		} `json:"data"`
	}
	if err := json.NewDecoder(created.Body).Decode(&createResponse); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResponse.Data.CreateUser.ID == "" {
		t.Fatalf("create response did not contain a user ID: %s", created.Body.String())
	}

	found := executeGraphQL(server, `query { user(id: "`+createResponse.Data.CreateUser.ID+`") { id name email createdAt } }`)
	if found.Code != http.StatusOK {
		t.Fatalf("query status = %d, want %d: %s", found.Code, http.StatusOK, found.Body.String())
	}
	if !json.Valid(found.Body.Bytes()) {
		t.Fatalf("query response is not JSON: %s", found.Body.String())
	}
}

func executeGraphQL(server http.Handler, query string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":`+strconv.Quote(query)+`}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}
