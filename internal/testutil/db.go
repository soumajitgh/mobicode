// Package testutil contains small, reusable infrastructure for backend tests.
package testutil

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/soumajitgh/mobicode/internal/database"
)

// NewDB opens an isolated SQLite database and applies the production migrations.
func NewDB(t *testing.T) *gorm.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(path)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	require.NoError(t, database.RunMigrations(path))
	return db
}
