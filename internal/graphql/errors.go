package graphql

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/auth"
)

// ErrorPresenter builds the GraphQL error presenter function.
// It maps domain errors to sanitized, structured GraphQL errors (e.g., UNAUTHENTICATED).
func ErrorPresenter(log *zap.Logger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		if errors.Is(err, auth.ErrUnauthenticated) {
			return &gqlerror.Error{
				Message:    "authentication required",
				Path:       graphql.GetPath(ctx),
				Extensions: map[string]any{"code": "UNAUTHENTICATED"},
			}
		}
		log.Error("graphql request failed", zap.Error(err))
		return &gqlerror.Error{Message: "internal error", Path: graphql.GetPath(ctx)}
	}
}
