package auth

import "errors"

var (
	// ErrUnauthenticated is returned when an operation requires authentication but no valid Principal is present.
	ErrUnauthenticated = errors.New("authentication required")

	// ErrInvalidCredentials is returned when login fails due to bad email or password.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidRefreshToken is returned when a refresh token is expired, revoked, or invalid.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrInvalidRegistration is returned when registration input validation fails.
	ErrInvalidRegistration = errors.New("invalid registration")

	// ErrUnauthorized is a backwards-compatible alias for ErrUnauthenticated.
	ErrUnauthorized = ErrUnauthenticated
)
