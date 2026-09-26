package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/store"
	"github.com/soumajitgh/mobicode/internal/store/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

func TestMigrationsCreateWorkingAuthStorage(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "migrations.db")
	persistence, err := store.Open(ctx, store.Config{SQLitePath: path, GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}

	user := &model.User{Email: "owner@example.com", PasswordHash: "hash", SessionVersion: 1}
	if err := persistence.Users.Create(ctx, user); err != nil {
		t.Fatal(err)
	}
	if err := persistence.Users.Create(ctx, &model.User{Email: user.Email, PasswordHash: "other", SessionVersion: 1}); !errors.Is(err, repository.ErrDuplicateEmail) {
		t.Fatalf("duplicate email: %v", err)
	}
	if err := persistence.Sessions.Commit("token", []byte("session-data"), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := persistence.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := store.Open(ctx, store.Config{SQLitePath: path, GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := reopened.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	data, found, err := reopened.Sessions.Find("token")
	if err != nil || !found || string(data) != "session-data" {
		t.Fatalf("persisted session: data = %q, found = %t, err = %v", data, found, err)
	}
}

func TestStoreRejectsDatabaseFromNewerApplication(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "newer.db")
	persistence, err := store.Open(ctx, store.Config{SQLitePath: path, GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	if err := persistence.Close(); err != nil {
		t.Fatal(err)
	}

	database, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (2, 1)"); err != nil {
		_ = database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	newer, err := store.Open(ctx, store.Config{SQLitePath: path, GORMLogLevel: logger.Silent})
	if err == nil {
		_ = newer.Close()
		t.Fatal("expected a newer database schema to be rejected")
	}
}
