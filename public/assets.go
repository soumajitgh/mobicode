package public

import "embed"

// Files contains the production CSS and JavaScript served by the Go binary.
//
//go:embed css/app.css js/htmx.min.js js/shadcn-templ-*.js
var Files embed.FS
