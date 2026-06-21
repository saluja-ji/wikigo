// Package errors defines standard sentinel errors and structured error types
// for the wikigo Wikimedia REST API client.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors representing common HTTP response states.
var (
	// ErrNotFound is returned when a resource is not found (HTTP 404).
	ErrNotFound = errors.New("resource not found")

	// ErrRateLimited is returned when the rate limit has been exceeded (HTTP 429).
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrUnauthorized is returned when client credentials or permissions are invalid (HTTP 401 or 403).
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrBadRequest is returned when the request is malformed (HTTP 400).
	ErrBadRequest = errors.New("bad request")

	// ErrServerError is returned for general internal server errors (HTTP 5xx).
	ErrServerError = errors.New("wikimedia server error")
)

// WikiError wraps API errors and preserves the HTTP status code.
type WikiError struct {
	// StatusCode is the HTTP status code returned by the API response.
	StatusCode int
	// Err is the underlying sentinel or base error.
	Err error
	// Message is an optional descriptive error message returned by the server.
	Message string
}

// Error returns a formatted string representation of the WikiError.
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
