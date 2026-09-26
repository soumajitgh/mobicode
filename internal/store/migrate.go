package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pressly/goose/v3"

	"github.com/soumajitgh/mobicode/internal/store/migrations"
)

func runMigrations(ctx context.Context, db *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("configure Goose SQLite dialect: %w", err)
	}
	goose.SetBaseFS(migrations.Files)
	available, err := goose.CollectMigrations(".", 0, int64(^uint64(0)>>1))
	if err != nil && !errors.Is(err, goose.ErrNoMigrationFiles) {
		return fmt.Errorf("load embedded migrations: %w", err)
	}
	current, err := goose.EnsureDBVersionContext(ctx, db)
	if err != nil {
		return fmt.Errorf("read Goose database version: %w", err)
	}
	var latest int64
	if len(available) > 0 {
		latest = available[len(available)-1].Version
	}
	if current > latest {
		return fmt.Errorf("database version %d is newer than embedded migration version %d", current, latest)
	}
	if current > 0 {
		found := false
		for _, migration := range available {
			if migration.Version == current {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("database version %d has no matching embedded migration", current)
		}
	}
	if len(available) == 0 {
		return nil
	}
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("apply Goose migrations: %w", err)
	}
	return nil
}
