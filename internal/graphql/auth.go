package graphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/vektah/gqlparser/v2/ast"
)

type rootField struct {
	Object string
	Field  string
}

var publicFields = map[rootField]struct{}{
	{"Mutation", "register"}:     {},
	{"Mutation", "login"}:        {},
	{"Mutation", "refreshToken"}: {},
}

// AuthMiddleware returns a gqlgen RootFieldMiddleware that guards GraphQL root operations.
// By default, every root field requires a valid Principal in the context.
// Only operations explicitly listed in publicFields are allowed without authentication.
func AuthMiddleware() graphql.RootFieldMiddleware {
	return func(ctx context.Context, next graphql.RootResolver) graphql.Marshaler {
		rc := graphql.GetRootFieldContext(ctx)
		if rc != nil {
			objectName := rc.Object
			opCtx := graphql.GetOperationContext(ctx)
			if opCtx != nil && opCtx.Operation != nil {
				switch opCtx.Operation.Operation {
				case ast.Mutation:
					objectName = "Mutation"
				case ast.Query:
					objectName = "Query"
				case ast.Subscription:
					objectName = "Subscription"
				}
			}

			rf := rootField{
				Object: objectName,
				Field:  rc.Field.Name,
			}
			if _, isPublic := publicFields[rf]; !isPublic {
				if _, ok := auth.PrincipalFromContext(ctx); !ok {
					graphql.AddError(ctx, auth.ErrUnauthenticated)
					return graphql.Null
				}
			}
		}
		return next(ctx)
	}
}
