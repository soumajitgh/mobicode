package auth_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/store"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := auth.Hash("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.Verify("correct-horse-battery", hash) {
		t.Fatal("correct password was rejected")
	}
	if auth.Verify("wrong-password", hash) {
		t.Fatal("wrong password was accepted")
	}
	if auth.Verify("password", "not-a-password-hash") {
		t.Fatal("malformed hash was accepted")
	}
}

func TestAuthenticationRules(t *testing.T) {
	ctx := context.Background()
	persistence, err := store.Open(ctx, store.Config{
		SQLitePath:   filepath.Join(t.TempDir(), "auth.db"),
		GORMLogLevel: logger.Silent,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := persistence.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	service := auth.Service{
		Users:         persistence.Users,
		RecoveryToken: "a-long-recovery-token-of-at-least-32-bytes",
	}

	if _, err := service.CreateInitialUser(ctx, "invalid", "validpassword123"); !errors.Is(err, auth.ErrInvalidInput) {
		t.Fatalf("invalid initial user: %v", err)
	}
	user, err := service.CreateInitialUser(ctx, " Admin@Example.com ", "validpassword123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "admin@example.com" {
		t.Fatalf("normalized email = %q", user.Email)
	}
	if _, err := service.CreateInitialUser(ctx, "other@example.com", "validpassword123"); !errors.Is(err, auth.ErrInitialUserAlreadyExists) {
		t.Fatalf("second initial user: %v", err)
	}

	if _, err := service.Login(ctx, user.Email, "wrong-password"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	if _, err := service.Login(ctx, " ADMIN@EXAMPLE.COM ", "validpassword123"); err != nil {
		t.Fatalf("normalized login: %v", err)
	}

	if err := service.ResetPassword(ctx, user.Email, "replacement-password", "wrong-token"); !errors.Is(err, auth.ErrInvalidRecoveryToken) {
		t.Fatalf("invalid recovery token: %v", err)
	}
	if err := service.ResetPassword(ctx, user.Email, "replacement-password", service.RecoveryToken); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(ctx, user.Email, "validpassword123"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("old password after reset: %v", err)
	}
	if _, err := service.Login(ctx, user.Email, "replacement-password"); err != nil {
		t.Fatalf("new password after reset: %v", err)
	}

	updated, err := persistence.Users.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.SessionVersion != user.SessionVersion+1 {
		t.Fatalf("session version = %d, want %d", updated.SessionVersion, user.SessionVersion+1)
	}
}
