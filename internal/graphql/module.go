package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/user"
)

// Module provides the GraphQL handler and mounts it at /query.
var Module = fx.Module(
	"graphql",
	fx.Provide(user.NewResolver, NewResolver, NewServer),
	fx.Invoke(func(router *chi.Mux, server *handler.Server, jwt *auth.JWTService) {
		router.With(auth.Authenticate(jwt)).Handle("/query", server)
	}),
)
