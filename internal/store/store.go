package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm/logger"

	"github.com/soumajitgh/mobicode/internal/store/repository"
)

// Config controls store initialization.
type Config struct {
	SQLitePath   string
	GORMLogLevel logger.LogLevel
}

// Store owns the database and repositories used by the application.
type Store struct {
	db       *sql.DB
	Example  repository.ExampleRepository
	Users    repository.UserRepository
	Sessions SessionStore
}

// Open connects to SQLite, applies pending migrations, then constructs repositories.
func Open(ctx context.Context, config Config) (*Store, error) {
	db, err := openSQLite(ctx, config.SQLitePath, config.GORMLogLevel)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection: %w", err)
	}
	if err := runMigrations(ctx, sqlDB); err != nil {
		return nil, fmt.Errorf("initialize database schema: %w", errors.Join(err, sqlDB.Close()))
	}
	return &Store{db: sqlDB, Example: repository.NewExample(db), Users: repository.NewUser(db), Sessions: SessionStore{DB: sqlDB}}, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}
