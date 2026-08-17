package main

import (
	"github.com/soumajitgh/mobicode/internal/app"

	"go.uber.org/fx"
)

// main starts the composed API application.
func main() {
	fx.New(app.Module).Run()
}
