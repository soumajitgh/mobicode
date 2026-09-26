package auth

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/store"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("correct-horse-battery")
	if err != nil || !Verify("correct-horse-battery", hash) || Verify("wrong", hash) || Verify("x", "$broken") {
		t.Fatalf("hash verification: %v", err)
	}
}

func TestService(t *testing.T) {
	s, err := store.Open(context.Background(), store.Config{SQLitePath: filepath.Join(t.TempDir(), "test.db"), GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	service := Service{Users: s.Users, RecoveryToken: "a-long-recovery-token-of-at-least-32-bytes"}
	ctx := context.Background()
	if _, err := service.Register(ctx, " User@Example.com ", "123456789012", "bad"); !errors.Is(err, ErrInvalidRecoveryToken) {
		t.Fatal(err)
	}
	for _, password := range []string{"short", strings.Repeat("x", 129)} {
		if _, err := service.Register(ctx, " User@Example.com ", password, "a-long-recovery-token-of-at-least-32-bytes"); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(err)
		}
	}
	user, err := service.Register(ctx, " User@Example.com ", "123456789012", "a-long-recovery-token-of-at-least-32-bytes")
	if err != nil || user.Email != "user@example.com" {
		t.Fatalf("register: %+v %v", user, err)
	}
	if _, err := service.Register(ctx, "USER@example.com", "123456789012", "a-long-recovery-token-of-at-least-32-bytes"); !errors.Is(err, ErrDuplicateEmail) {
		t.Fatal(err)
	}
	for _, email := range []string{"missing@example.com", "user@example.com"} {
		if _, err := service.Login(ctx, email, "wrong password"); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatal(err)
		}
	}
	if err := service.ResetPassword(ctx, " USER@EXAMPLE.COM ", "anotherpassword", "bad"); !errors.Is(err, ErrInvalidRecoveryToken) {
		t.Fatal(err)
	}
	if err := service.ResetPassword(ctx, " USER@EXAMPLE.COM ", "anotherpassword", "a-long-recovery-token-of-at-least-32-bytes"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(ctx, "user@example.com", "123456789012"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal(err)
	}
	if _, err := service.Login(ctx, "USER@example.com", "anotherpassword"); err != nil {
		t.Fatal(err)
	}
	newer, err := s.Users.FindByID(ctx, user.ID)
	if err != nil || newer.SessionVersion != user.SessionVersion+1 {
		t.Fatalf("version: %+v %v", newer, err)
	}
}

func TestCreateInitialUser(t *testing.T) {
	s, err := store.Open(context.Background(), store.Config{SQLitePath: filepath.Join(t.TempDir(), "test-initial.db"), GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})
	ctx := context.Background()
	service := Service{Users: s.Users}

	// Count initially 0
	count, err := s.Users.Count(ctx)
	if err != nil || count != 0 {
		t.Fatalf("expected count 0, got %d, err: %v", count, err)
	}

	// Invalid password / email
	if _, err := service.CreateInitialUser(ctx, "invalid-email", "validpassword123"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if _, err := service.CreateInitialUser(ctx, "admin@example.com", "short"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	// Valid initial user creation
	user, err := service.CreateInitialUser(ctx, " Admin@Example.com ", "validpassword123")
	if err != nil || user.Email != "admin@example.com" {
		t.Fatalf("expected success, got user: %+v, err: %v", user, err)
	}

	// Count now 1
	count, err = s.Users.Count(ctx)
	if err != nil || count != 1 {
		t.Fatalf("expected count 1, got %d, err: %v", count, err)
	}

	// FindFirst
	first, err := s.Users.FindFirst(ctx)
	if err != nil || first.Email != "admin@example.com" {
		t.Fatalf("expected first user admin@example.com, got %+v, err: %v", first, err)
	}

	// Cannot create initial user again once user exists
	if _, err := service.CreateInitialUser(ctx, "other@example.com", "validpassword123"); !errors.Is(err, ErrInitialUserAlreadyExists) {
		t.Fatalf("expected ErrInitialUserAlreadyExists, got %v", err)
	}
}
