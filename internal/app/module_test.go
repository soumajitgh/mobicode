package app

import (
	"testing"

	"go.uber.org/fx"
)

// TestModuleValidates verifies that all application dependencies compose through Fx.
func TestModuleValidates(t *testing.T) {
	if err := fx.ValidateApp(Module); err != nil {
		t.Fatalf("validate application module: %v", err)
	}
}
