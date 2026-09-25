package migrations

import "embed"

// Files contains this package and any versioned SQL migrations added here later.
// Embedding the directory lets this initial phase have no SQL migration files.
//
//go:embed *
var Files embed.FS
