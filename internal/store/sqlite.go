package store

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openSQLite(ctx context.Context, path string, logLevel logger.LogLevel) (*gorm.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("SQLite path is required")
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve SQLite path: %w", err)
	}
	dsn := (&url.URL{Scheme: "file", Path: filepath.ToSlash(absolutePath)}).String()
	dsn += "?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get SQLite connection: %w", err)
	}
	// A single connection keeps connection-local pragmas consistent and serializes writes.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("verify SQLite connectivity: %w", err)
	}
	return db, nil
}
