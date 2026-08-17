// Package database owns SQLite and GORM setup.
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/soumajitgh/mobicode/internal/config"

	"github.com/tursodatabase/libsql-client-go/libsql"
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

// New opens and configures the remote libSQL database connection.
func New(cfg config.Config, logger *zap.Logger) (Result, error) {
	connector, err := libsql.NewConnector(cfg.Database.URL)
	if err != nil {
		return Result{}, fmt.Errorf("create libsql connector: %w", err)
	}
	sqlDB := sql.OpenDB(connector)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		_ = sqlDB.Close()
		return Result{}, fmt.Errorf("open gorm database: %w", err)
	}
	logger.Info("database initialized", zap.String("url", cfg.Database.URL))
	return Result{DB: db, SQLDB: sqlDB}, nil
}

// registerLifecycle closes the database during application shutdown.
func registerLifecycle(lc fx.Lifecycle, db *sql.DB) {
	lc.Append(fx.Hook{OnStop: func(ctx context.Context) error { return db.Close() }})
}

var Module = fx.Module("database", fx.Provide(New), fx.Invoke(registerLifecycle))
