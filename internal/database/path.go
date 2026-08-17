package database

import (
	"fmt"
	"os"
	"path/filepath"
)

func ensureParentDirectory(databasePath string) error {
	dir := filepath.Dir(databasePath)
	if dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	return nil
}
