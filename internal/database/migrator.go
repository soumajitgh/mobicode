package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/soumajitgh/mobicode/db/migrations"
)

// RunMigrations applies embedded migrations to the database at databasePath.
// Calling it after the schema is current is safe.
func RunMigrations(databasePath string) error {
	db, err := Open(databasePath)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database connection: %w", err)
	}
	defer sqlDB.Close()
	return runMigrations(sqlDB)
}

func runMigrations(sqlDB *sql.DB) error {

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}
	driver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}
	migrator, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// HasPendingMigrations reports whether the database version lags embedded migrations.
func HasPendingMigrations(db *sql.DB) (bool, error) {
	latest, err := latestMigrationVersion()
	if err != nil {
		return false, err
	}
	if latest == 0 {
		return false, nil
	}

	var version uint
	var dirty bool
	err = db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return true, nil
		}
		return false, fmt.Errorf("read migration state: %w", err)
	}
	if dirty {
		return false, fmt.Errorf("database migration is dirty; resolve it before serving")
	}
	return version < latest, nil
}

func latestMigrationVersion() (uint, error) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return 0, fmt.Errorf("read embedded migrations: %w", err)
	}
	var latest uint
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		prefix := strings.SplitN(filepath.Base(entry.Name()), "_", 2)[0]
		version, err := strconv.ParseUint(prefix, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse migration %q: %w", entry.Name(), err)
		}
		if uint(version) > latest {
			latest = uint(version)
		}
	}
	return latest, nil
}
