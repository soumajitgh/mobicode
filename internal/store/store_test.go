package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/soumajitgh/mobicode/internal/store/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStoreLifecycleAndExampleRepository(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "store.db")
	s, err := Open(ctx, Config{SQLitePath: path, GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	var foreignKeys, busyTimeout int
	if err := s.db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 || busyTimeout != 5000 {
		t.Fatalf("SQLite pragmas: foreign_keys=%d busy_timeout=%d", foreignKeys, busyTimeout)
	}

	var migrationTable string
	if err := s.db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'goose_db_version'").Scan(&migrationTable); err != nil {
		t.Fatalf("migration table: %v", err)
	}
	var exampleTableCount int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'examples'").Scan(&exampleTableCount); err != nil {
		t.Fatal(err)
	}
	if exampleTableCount != 0 {
		t.Fatal("example table was created without a migration")
	}

	// This table exists only in the test so the repository flow can be exercised.
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE examples (id integer PRIMARY KEY, name text NOT NULL, created_at datetime, updated_at datetime)`); err != nil {
		t.Fatal(err)
	}
	example := &model.Example{Name: "first"}
	if err := s.Example.Create(ctx, example); err != nil {
		t.Fatal(err)
	}
	found, err := s.Example.FindByID(ctx, example.ID)
	if err != nil || found.Name != "first" {
		t.Fatalf("find: %+v, %v", found, err)
	}
	example.Name = "second"
	if err := s.Example.Update(ctx, example); err != nil {
		t.Fatal(err)
	}
	list, err := s.Example.List(ctx)
	if err != nil || len(list) != 1 || list[0].Name != "second" {
		t.Fatalf("list: %+v, %v", list, err)
	}
	if err := s.Example.Delete(ctx, example.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Example.FindByID(ctx, example.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted record lookup: %v", err)
	}
}

func TestOpenRejectsNewerSchema(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "store.db")
	s, err := Open(ctx, Config{SQLitePath: path, GORMLogLevel: logger.Silent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, 1)"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(ctx, Config{SQLitePath: path, GORMLogLevel: logger.Silent}); err == nil {
		t.Fatal("expected startup to reject a database newer than embedded migrations")
	}
}
