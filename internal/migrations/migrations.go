// Package migrations applies the embedded SQL migrations before the API starts.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/soumajitgh/mobicode/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

//go:embed sql/*.sql
var files embed.FS

func registerLifecycle(lc fx.Lifecycle, db *sql.DB, cfg config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		startupCtx, cancel := context.WithTimeout(ctx, cfg.Database.StartupTimeout)
		defer cancel()

		for {
			if err := up(startupCtx, db); err == nil {
				logger.Info("database migrations applied")
				return nil
			} else if startupCtx.Err() != nil {
				return fmt.Errorf("apply database migrations within %s: %w", cfg.Database.StartupTimeout, err)
			}

			logger.Debug("database is not ready; retrying migrations", zap.Duration("retry_in", time.Second))
			select {
			case <-startupCtx.Done():
				return fmt.Errorf("apply database migrations within %s: %w", cfg.Database.StartupTimeout, startupCtx.Err())
			case <-time.After(time.Second):
			}
		}
	}})
}

func up(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	paths, err := fs.Glob(files, "sql/*.sql")
	if err != nil {
		return fmt.Errorf("find migration files: %w", err)
	}
	sort.Strings(paths)

	for _, path := range paths {
		version := strings.TrimSuffix(strings.TrimPrefix(path, "sql/"), ".sql")
		var applied bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)`, version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}

		contents, err := files.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		for _, statement := range statements(upSection(string(contents))) {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply migration %s: %w", version, err)
			}
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}
	}
	return nil
}

func upSection(contents string) string {
	up, found := strings.CutPrefix(contents, "-- +goose Up")
	if !found {
		return contents
	}
	section, _, _ := strings.Cut(up, "-- +goose Down")
	return section
}

func statements(sql string) []string {
	var result []string
	for _, statement := range strings.Split(sql, ";") {
		if statement = strings.TrimSpace(statement); statement != "" {
			result = append(result, statement)
		}
	}
	return result
}

var Module = fx.Module("migrations", fx.Invoke(registerLifecycle))
