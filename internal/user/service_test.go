package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

type mockRepo struct {
	users map[string]*User
}

func (m *mockRepo) FindByID(_ context.Context, id string) (*User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockRepo) Create(_ context.Context, user *User) error {
	m.users[user.ID] = user
	return nil
}

func TestGetUser(t *testing.T) {
	existingUser := &User{
		ID:        "user-123",
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
	}

	repo := &mockRepo{
		users: map[string]*User{
			"user-123": existingUser,
		},
	}
	service := NewService(repo, zap.NewNop())

	t.Run("empty string id returns ErrInvalid without repo lookup", func(t *testing.T) {
		u, err := service.GetUser(context.Background(), "")
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected ErrInvalid, got %v", err)
		}
		if u != nil {
			t.Fatalf("expected nil user, got %v", u)
		}
	})

	t.Run("whitespace string id returns ErrInvalid", func(t *testing.T) {
		u, err := service.GetUser(context.Background(), "   ")
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected ErrInvalid, got %v", err)
		}
		if u != nil {
			t.Fatalf("expected nil user, got %v", u)
		}
	})

	t.Run("non-existent id returns ErrNotFound", func(t *testing.T) {
		u, err := service.GetUser(context.Background(), "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
		if u != nil {
			t.Fatalf("expected nil user, got %v", u)
		}
	})

	t.Run("valid id returns user", func(t *testing.T) {
		u, err := service.GetUser(context.Background(), "user-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u == nil || u.ID != "user-123" {
			t.Fatalf("expected user-123, got %v", u)
		}
	})
}
