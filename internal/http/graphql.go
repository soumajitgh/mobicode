package http

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	appgraphql "github.com/soumajitgh/mobicode/internal/graphql"
)

func newGraphQLHandler(resolver *appgraphql.Resolver) *handler.Server {
	srv := handler.New(appgraphql.NewExecutableSchema(appgraphql.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	return srv
}
