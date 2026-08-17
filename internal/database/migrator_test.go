package database

import (
	"path/filepath"
	"testing"

	"github.com/soumajitgh/mobicode/internal/config"
)

func TestMigratorUpCreatesUsersAndIsIdempotent(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "app.db")
	db, err := New(&config.Config{DatabasePath: databasePath})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get SQL database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	pending, err := HasPendingMigrations(sqlDB)
	if err != nil {
		t.Fatalf("check pending migrations: %v", err)
	}
	if !pending {
		t.Fatal("fresh database unexpectedly has no pending migrations")
	}
	if err := RunMigrations(databasePath); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	if err := RunMigrations(databasePath); err != nil {
		t.Fatalf("reapply migrations: %v", err)
	}
	pending, err = HasPendingMigrations(sqlDB)
	if err != nil {
		t.Fatalf("recheck pending migrations: %v", err)
	}
	if pending {
		t.Fatal("migrated database still has pending migrations")
	}

	var tables int
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('users', 'refresh_tokens')").Scan(&tables).Error; err != nil {
		t.Fatalf("query schema: %v", err)
	}
	if tables != 2 {
		t.Fatalf("auth table count = %d, want 2", tables)
	}

	var passwordHashColumns int
	if err := db.Raw("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = 'password_hash'").Scan(&passwordHashColumns).Error; err != nil {
		t.Fatalf("query users columns: %v", err)
	}
	if passwordHashColumns != 1 {
		t.Fatalf("password_hash column count = %d, want 1", passwordHashColumns)
	}
}
