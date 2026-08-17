package main

import (
	"mobicode/apps/server/internal/app"

	"go.uber.org/fx"
)

// main starts the composed API application.
func main() {
	fx.New(app.Module).Run()
}
