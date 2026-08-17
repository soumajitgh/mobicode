// Package requestctx owns typed access to request-scoped transport metadata.
package requestctx

import "context"

type requestIDKey struct{}

// WithRequestID returns a context carrying the HTTP request identifier.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID returns the request identifier, or an empty string when absent.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
