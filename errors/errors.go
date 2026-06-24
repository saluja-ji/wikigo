// Package errors defines standard error types for the wikigo client.
package errors

import (
	"errors"
	"fmt"
)

// Common API errors.
var (
	// ErrNotFound means the requested item was not found (404).
	ErrNotFound = errors.New("resource not found")

	// ErrRateLimited means requests were sent too fast (429).
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrUnauthorized means authentication or access was denied (401/403).
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrBadRequest means the request structure is invalid (400).
	ErrBadRequest = errors.New("bad request")

	// ErrServerError means the Wikimedia server failed (5xx).
	ErrServerError = errors.New("wikimedia server error")
)

// WikiError represents an API error with its HTTP status code.
type WikiError struct {
	// StatusCode is the HTTP status code from the server.
	StatusCode int
	// Err is the actual error type.
	Err error
	// Message is the detailed error description.
	Message string
}

// Error formats the error into a readable string.
func (e *WikiError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("wikimedia api error (status %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("wikimedia api error (status %d): %v", e.StatusCode, e.Err)
}

// Unwrap returns the underlying error.
func (e *WikiError) Unwrap() error {
	return e.Err
}
