package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type memoryRepository struct {
	users map[string]*User
}

func (r *memoryRepository) FindByID(_ context.Context, id string) (*User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

func (r *memoryRepository) FindByEmail(_ context.Context, email string) (*User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryRepository) Create(_ context.Context, user *User) error {
	r.users[user.ID] = user
	return nil
}

func TestHandlerCreatesAndGetsUser(t *testing.T) {
	repo := &memoryRepository{users: make(map[string]*User)}
	handler := NewHandler(NewService(repo, zap.NewNop()))
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	create := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Zoravix","email":"z@fotopick.in"}`))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	router.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", created.Code, http.StatusCreated, created.Body.String())
	}

	var user User
	if err := json.NewDecoder(created.Body).Decode(&user); err != nil {
		t.Fatalf("decode created user: %v", err)
	}
	if user.ID == "" {
		t.Fatal("created user has no ID")
	}

	found := httptest.NewRecorder()
	router.ServeHTTP(found, httptest.NewRequest(http.MethodGet, "/users/"+user.ID, nil))
	if found.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d: %s", found.Code, http.StatusOK, found.Body.String())
	}
}
