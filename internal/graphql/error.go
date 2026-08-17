package graphql

import (
	"context"
	"errors"

	"github.com/soumajitgh/mobicode/internal/apperror"
	"github.com/soumajitgh/mobicode/internal/requestctx"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// PresentError converts application errors into safe GraphQL errors.
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	code, message := graphQLErrorDetails(err)
	graphQLError := &gqlerror.Error{Message: message, Extensions: map[string]any{"code": code}}
	if requestID := requestctx.RequestID(ctx); requestID != "" {
		graphQLError.Extensions["request_id"] = requestID
	}
	return graphQLError
}

// graphQLErrorDetails maps application errors to public GraphQL details.
func graphQLErrorDetails(err error) (string, string) {
	var applicationError *apperror.Error
	if !errors.As(err, &applicationError) {
		return "INTERNAL_SERVER_ERROR", "internal server error"
	}

	switch applicationError.Code {
	case "validation":
		return "BAD_USER_INPUT", applicationError.Message
	case "not_found":
		return "NOT_FOUND", applicationError.Message
	case "conflict":
		return "CONFLICT", applicationError.Message
	default:
		return "INTERNAL_SERVER_ERROR", "internal server error"
	}
}
