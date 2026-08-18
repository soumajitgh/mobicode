package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/soumajitgh/mobicode/internal/user"
)

// Module provides the GraphQL handler and mounts it on chi router with HTTP auth middleware.
var Module = fx.Module(
	"graphql",
	fx.Provide(user.NewResolver, NewResolver, NewServer),
	fx.Invoke(func(router *chi.Mux, server *handler.Server, authMiddleware func(http.Handler) http.Handler) {
		router.With(authMiddleware).Handle("/query", server)
		router.With(authMiddleware).Handle("/graphql", server)
	}),
)
