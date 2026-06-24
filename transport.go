package wikigo

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	wikierrors "github.com/saluja-ji/wikigo/errors"
	"golang.org/x/time/rate"
)

type cacheEntry struct {
	statusCode int
	status     string
	header     http.Header
	body       []byte
	expiresAt  time.Time
}

// cacheTransport caches successful GET requests in memory.
type cacheTransport struct {
	underlying http.RoundTripper
	ttl        time.Duration
	maxEntries int

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

// newCacheTransport creates a new cache transport.
func newCacheTransport(underlying http.RoundTripper, ttl time.Duration, maxEntries int) *cacheTransport {
	return &cacheTransport{
		underlying: underlying,
		ttl:        ttl,
		maxEntries: maxEntries,
		cache:      make(map[string]cacheEntry),
	}
}

// RoundTrip handles caching for the request.
func (c *cacheTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return c.underlying.RoundTrip(req)
	}

	key := req.URL.String()

	c.mu.RLock()
	entry, found := c.cache[key]
	c.mu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		resp := &http.Response{
			StatusCode:    entry.statusCode,
			Status:        entry.status,
			Header:        entry.header.Clone(),
			Body:          io.NopCloser(bytes.NewReader(entry.body)),
			ContentLength: int64(len(entry.body)),
			Request:       req,
		}
		return resp, nil
	}

	resp, err := c.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK && resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			c.mu.Lock()
			if len(c.cache) >= c.maxEntries {
				for k := range c.cache {
					delete(c.cache, k)
					break
				}
			}
			c.cache[key] = cacheEntry{
				statusCode: resp.StatusCode,
				status:     resp.Status,
				header:     resp.Header.Clone(),
				body:       bodyBytes,
				expiresAt:  time.Now().Add(c.ttl),
			}
			c.mu.Unlock()
		}
	}

	return resp, nil
}

// retryTransport handles rate limiting and retrying failed requests.
type retryTransport struct {
	underlying http.RoundTripper
	limiter    *rate.Limiter
	maxRetries int
}

// newRetryTransport creates a new retry transport.
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

// RoundTrip executes the request with rate limiting and retries.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Wait for rate limit token.
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	var lastErr error
	var resp *http.Response
	var lastWikiErr *wikierrors.WikiError

	// Run request, retrying if necessary.
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		// Reset body if retrying.
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("failed to get request body for retry: %w", err)
			}
			req.Body = body
		}

		// Check if context is cancelled.
		if err := req.Context().Err(); err != nil {
			return nil, err
		}

		resp, lastErr = t.underlying.RoundTrip(req)
		if lastErr != nil {
			// Return network errors immediately.
			return nil, lastErr
		}

		// Return successful response.
		if resp.StatusCode < 400 {
			return resp, nil
		}

		// Read error message from response.
		var msg string
		if resp.Body != nil {
			msgBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			msg = string(msgBytes)
		}

		// Map HTTP status to custom error.
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

		// Retry only on rate limit (429) or server error (503).
		shouldRetry := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable
		if !shouldRetry || attempt == t.maxRetries {
			return nil, wikiErr
		}

		// Calculate retry delay.
		var delay time.Duration
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if retryAfterHeader := resp.Header.Get("Retry-After"); retryAfterHeader != "" {
				delay = parseRetryAfter(retryAfterHeader)
			}
		}

		// Fallback to exponential backoff.
		if delay <= 0 {
			delay = calculateBackoff(attempt)
		}

		// Wait before retrying.
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

// parseRetryAfter parses the Retry-After header (seconds or date).
func parseRetryAfter(header string) time.Duration {
	// Parse as seconds.
	if secs, err := strconv.Atoi(header); err == nil {
		if secs > 0 {
			return time.Duration(secs) * time.Second
		}
		return 0
	}

	// Parse as HTTP date.
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

// calculateBackoff calculates wait time with exponential delay and random jitter.
func calculateBackoff(attempt int) time.Duration {
	// Avoid overflow.
	shift := attempt
	if shift > 30 {
		shift = 30
	}
	// Double the delay each time (base is 1 second).
	backoff := float64(time.Second << shift)

	// Cap maximum delay at 30 seconds.
	const maxBackoff = 30 * time.Second
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}

	// Add random variance (+/- 20%).
	jitter := 0.8 + 0.4*rand.Float64()
	return time.Duration(backoff * jitter)
}
