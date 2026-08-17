package database

import (
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides a database and verifies that it has been initialized.
var Module = fx.Module(
	"database",
	fx.Provide(New),
	fx.Invoke(verifySchema),
	fx.Invoke(registerLifecycle),
)

func verifySchema(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	pending, err := HasPendingMigrations(sqlDB)
	if err != nil {
		return err
	}
	if pending {
		return fmt.Errorf("database schema is out of date — run `mobicode init` first")
	}
	return nil
}
