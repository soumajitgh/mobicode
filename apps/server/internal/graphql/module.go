// Package graphql builds the application's GraphQL transport.
package graphql

import (
	"mobicode/apps/server/internal/graphql/resolver"

	"go.uber.org/fx"
)

var Module = fx.Module("graphql", resolver.Module, fx.Provide(NewHandler))
