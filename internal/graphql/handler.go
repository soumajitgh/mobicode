package graphql

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/graph/generated"
	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/user"
)

// NewServer creates the GraphQL handler with safe client error messages.
func NewServer(resolver *Resolver, log *zap.Logger) *handler.Server {
	server := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	server.SetErrorPresenter(errorPresenter(log))
	return server
}

func errorPresenter(log *zap.Logger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		if errors.Is(err, user.ErrInvalid) {
			return &gqlerror.Error{Message: "invalid user", Path: graphql.GetPath(ctx)}
		}
		if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrInvalidRefreshToken) || errors.Is(err, auth.ErrInvalidRegistration) {
			return &gqlerror.Error{Message: err.Error(), Path: graphql.GetPath(ctx)}
		}
		log.Error("graphql request failed", zap.Error(err))
		return &gqlerror.Error{Message: "internal error", Path: graphql.GetPath(ctx)}
	}
}
