// Package graphql provides the GraphQL HTTP transport.
package graphql

import (
	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/user"
)

// Resolver composes feature resolvers for gqlgen's executable schema.
type Resolver struct {
	user *user.Resolver
	auth *auth.Resolver
}

func NewResolver(userResolver *user.Resolver, authResolver *auth.Resolver) *Resolver {
	return &Resolver{user: userResolver, auth: authResolver}
}
