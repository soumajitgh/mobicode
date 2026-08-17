// Package graphql builds the application's GraphQL transport.
package graphql

import (
	"github.com/soumajitgh/mobicode/internal/graphql/resolver"

	"go.uber.org/fx"
)

var Module = fx.Module("graphql", resolver.Module, fx.Provide(fx.Annotate(NewHandler, fx.ResultTags(`name:"graphql"`))))
