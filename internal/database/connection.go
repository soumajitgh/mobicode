// Package database provides the application's SQLite connection.
package database

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/config"
)

// New opens the SQLite database and configures it for its single-process use.
func New(cfg *config.Config) (*gorm.DB, error) {
	return Open(cfg.DatabasePath)
}

// Open creates a SQLite connection for the supplied path.
func Open(databasePath string) (*gorm.DB, error) {
	if err := ensureParentDirectory(databasePath); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	return db, nil
}

func registerLifecycle(lc fx.Lifecycle, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database connection: %w", err)
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return sqlDB.Close() }})
	return nil
}
