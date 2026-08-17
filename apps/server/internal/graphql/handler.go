package graphql

import (
	"context"
	"errors"
	"net/http"

	"mobicode/apps/server/internal/config"
	"mobicode/apps/server/internal/graphql/generated"
	"mobicode/apps/server/internal/graphql/resolver"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"go.uber.org/zap"
)

// NewHandler constructs the GraphQL HTTP handler and its safety controls.
func NewHandler(cfg config.Config, resolvers *resolver.Resolver, logger *zap.Logger) http.Handler {
	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolvers}))
	server.Use(extension.FixedComplexityLimit(cfg.Server.GraphQLComplexity))
	server.SetErrorPresenter(PresentError)
	server.SetRecoverFunc(func(ctx context.Context, recovered any) error {
		logger.Error("graphql panic recovered", zap.Any("panic", recovered))
		return errors.New("internal server error")
	})
	return server
}
