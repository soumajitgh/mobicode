package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/graphql/generated"
)

// NewServer creates the GraphQL handler configured with its ErrorPresenter.
func NewServer(resolver *Resolver, log *zap.Logger) *handler.Server {
	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	server.SetErrorPresenter(ErrorPresenter(log))
	return server
}
