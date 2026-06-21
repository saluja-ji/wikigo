// Package wikigo implements the Wikimedia REST API SDK.
package wikigo

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	wikierrors "github.com/saluja-ji/wikigo/errors"
	"golang.org/x/time/rate"
)

// retryTransport wraps an underlying HTTP RoundTripper and implements transparent
// rate limiting and retry logic with exponential backoff and jitter.
type retryTransport struct {
	underlying http.RoundTripper
	limiter    *rate.Limiter
	maxRetries int
}

// newRetryTransport constructs a new retryTransport.
func newRetryTransport(underlying http.RoundTripper, limit rate.Limit, burst int, maxRetries int) *retryTransport {
	if underlying == nil {
		underlying = http.DefaultTransport
	}
	return &retryTransport{
		underlying: underlying,
		limiter:    rate.NewLimiter(limit, burst),
		maxRetries: maxRetries,
	}
}

// RoundTrip executes the request, applying rate limiting and retry logic.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Apply rate limiting. This blocks until a token is available or context is cancelled.
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	var lastErr error
	var resp *http.Response
	var lastWikiErr *wikierrors.WikiError

	// Seed rand source for jitter (thread-safe, package-level math/rand is auto-seeded since Go 1.20)
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		// Rewind request body if it exists for retry attempts
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("failed to get request body for retry: %w", err)
			}
			req.Body = body
		}

		// Ensure context is not cancelled before launching request
		if err := req.Context().Err(); err != nil {
			return nil, err
		}

		resp, lastErr = t.underlying.RoundTrip(req)
		if lastErr != nil {
			// Network/transport level errors are returned directly.
			return nil, lastErr
		}

		// Success path (non-error HTTP status)
		if resp.StatusCode < 400 {
			return resp, nil
		}

		// Read a short portion of response body to capture the API error message.
		var msg string
		if resp.Body != nil {
			msgBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			msg = string(msgBytes)
		}

		// Map HTTP status code to specific SDK sentinel errors
		var sentinelErr error
		switch resp.StatusCode {
		case http.StatusNotFound:
			sentinelErr = wikierrors.ErrNotFound
		case http.StatusTooManyRequests:
			sentinelErr = wikierrors.ErrRateLimited
		case http.StatusUnauthorized, http.StatusForbidden:
			sentinelErr = wikierrors.ErrUnauthorized
		case http.StatusBadRequest:
			sentinelErr = wikierrors.ErrBadRequest
		default:
			if resp.StatusCode >= 500 {
				sentinelErr = wikierrors.ErrServerError
			} else {
				sentinelErr = fmt.Errorf("http status error: %d", resp.StatusCode)
			}
		}

		wikiErr := &wikierrors.WikiError{
			StatusCode: resp.StatusCode,
			Err:        sentinelErr,
			Message:    msg,
		}
		lastWikiErr = wikiErr

		// Determine if we should retry.
		// Only retry on rate limiting (429) or temporary server unavailability (503).
		shouldRetry := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable
		if !shouldRetry || attempt == t.maxRetries {
			return nil, wikiErr
		}

		// Calculate backoff delay
		var delay time.Duration
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if retryAfterHeader := resp.Header.Get("Retry-After"); retryAfterHeader != "" {
				delay = parseRetryAfter(retryAfterHeader)
			}
		}

		// Use exponential backoff + jitter if delay is not set by Retry-After header
		if delay <= 0 {
			delay = calculateBackoff(attempt)
		}

		// Wait for backoff delay or context cancellation
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
		}
	}

	if lastWikiErr != nil {
		return nil, lastWikiErr
	}
	return nil, fmt.Errorf("retry limit exceeded")
}

// parseRetryAfter parses the Retry-After header value, which can be seconds or an HTTP-date.
func parseRetryAfter(header string) time.Duration {
	// Try parsing as integer seconds first
	if secs, err := strconv.Atoi(header); err == nil {
		if secs > 0 {
			return time.Duration(secs) * time.Second
		}
		return 0
	}

	// Try parsing standard HTTP-date formats
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC850,
		time.ANSIC,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, header); err == nil {
			if delay := time.Until(t); delay > 0 {
				return delay
			}
			return 0
		}
	}
	return 0
}

// calculateBackoff calculates exponential backoff with ±20% jitter.
// attempt is 0-indexed.
func calculateBackoff(attempt int) time.Duration {
	// Prevent bit shift overflow (time.Second is 10^9, so shift of 30 is safe for float64/int64)
	shift := attempt
	if shift > 30 {
		shift = 30
	}
	// base backoff: 1s * 2^attempt
	backoff := float64(time.Second << shift)

	// Cap base backoff at a reasonable maximum (e.g. 30 seconds) to prevent infinite growth
	const maxBackoff = 30 * time.Second
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}

	// ±20% jitter (ranges between 0.8 and 1.2)
	jitter := 0.8 + 0.4*rand.Float64()
	return time.Duration(backoff * jitter)
}
