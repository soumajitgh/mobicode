// Package migrations exposes the embedded database migration files.
package migrations

import "embed"

// FS contains all SQL migrations.
//
//go:embed *.sql
var FS embed.FS
