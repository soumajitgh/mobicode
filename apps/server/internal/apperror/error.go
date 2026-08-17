// Package apperror defines transport-independent application errors.
package apperror

import "errors"

type Error struct {
	Code, Message string
	Cause         error
}

// Error returns the safe application error message.
func (e *Error) Error() string { return e.Message }

// Unwrap exposes the underlying error cause.
func (e *Error) Unwrap() error { return e.Cause }

// New creates an application error with an optional cause.
func New(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// Code extracts an application error code or returns internal.
func Code(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return "internal"
}

var ErrNotFound = New("not_found", "resource was not found", nil)
var ErrConflict = New("conflict", "resource conflicts with existing state", nil)
