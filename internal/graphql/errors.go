package graphql

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/user"
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
		if errors.Is(err, user.ErrInvalid) {
			return &gqlerror.Error{Message: "invalid user", Path: graphql.GetPath(ctx)}
		}
		if errors.Is(err, user.ErrEmailTaken) {
			return &gqlerror.Error{Message: "email already taken", Path: graphql.GetPath(ctx)}
		}
		if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrInvalidRefreshToken) || errors.Is(err, auth.ErrInvalidRegistration) {
			return &gqlerror.Error{Message: err.Error(), Path: graphql.GetPath(ctx)}
		}
		log.Error("graphql request failed", zap.Error(err))
		return &gqlerror.Error{Message: "internal error", Path: graphql.GetPath(ctx)}
	}
}
