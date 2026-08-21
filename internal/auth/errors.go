package auth

import "errors"

var (
	// ErrUnauthenticated is returned when an operation requires authentication but no valid Principal is present.
	ErrUnauthenticated = errors.New("authentication required")

	// ErrUnauthorized is a backwards-compatible alias for ErrUnauthenticated.
	ErrUnauthorized = ErrUnauthenticated
)
