// Package database owns SQLite and GORM setup.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"mobicode/apps/server/internal/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Result struct {
	fx.Out
	DB    *gorm.DB
	SQLDB *sql.DB
}

// New opens and configures the SQLite database connection.
func New(cfg config.Config, logger *zap.Logger) (Result, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0o750); err != nil {
		return Result{}, fmt.Errorf("create database directory: %w", err)
	}
	path := url.PathEscape(cfg.Database.Path)
	dsn := "file:" + path + "?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=" + strconv.FormatInt(cfg.Database.BusyTimeout.Milliseconds(), 10)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return Result{}, fmt.Errorf("open sqlite database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return Result{}, fmt.Errorf("get sql database: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	logger.Info("database initialized", zap.String("path", cfg.Database.Path))
	return Result{DB: db, SQLDB: sqlDB}, nil
}

// registerLifecycle closes the database during application shutdown.
func registerLifecycle(lc fx.Lifecycle, db *sql.DB) {
	lc.Append(fx.Hook{OnStop: func(ctx context.Context) error { return db.Close() }})
}

var Module = fx.Module("database", fx.Provide(New), fx.Invoke(registerLifecycle))
