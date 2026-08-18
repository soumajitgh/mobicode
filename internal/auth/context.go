package auth

import "context"

// Principal represents the authenticated caller identity.
type Principal struct {
	UserID    string
	SessionID string
}

type contextKey struct{}

// WithPrincipal returns a new context containing the given Principal.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// PrincipalFromContext extracts the Principal from context if present.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	return p, ok
}
