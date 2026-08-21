package database

import (
	"path/filepath"
	"testing"

	"github.com/soumajitgh/mobicode/internal/config"
)

func TestMigratorUpCreatesSingleOwnerAuthSchemaAndIsIdempotent(t *testing.T) {
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
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('owner_identity', 'auth_replays', 'setup_sessions')").Scan(&tables).Error; err != nil {
		t.Fatalf("query schema: %v", err)
	}
	if tables != 3 {
		t.Fatalf("auth table count = %d, want 3", tables)
	}

	var legacyTables int
	if err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('users', 'refresh_tokens')").Scan(&legacyTables).Error; err != nil {
		t.Fatalf("query legacy schema: %v", err)
	}
	if legacyTables != 0 {
		t.Fatalf("legacy auth table count = %d, want 0", legacyTables)
	}
}
