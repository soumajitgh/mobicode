// Package graphql provides the GraphQL HTTP transport.
package graphql

import (
	"github.com/soumajitgh/mobicode/internal/auth"
)

// Resolver composes feature resolvers for gqlgen's executable schema.
type Resolver struct {
	owner *auth.OwnerService
}

func NewResolver(owner *auth.OwnerService) *Resolver {
	return &Resolver{owner: owner}
}
