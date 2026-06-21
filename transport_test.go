// Package wikigo implements the Wikimedia REST API SDK.
package wikigo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	wikierrors "github.com/saluja-ji/wikigo/errors"
	"golang.org/x/time/rate"
)

func TestTransport_RateLimiter(t *testing.T) {
	// Setup a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Configure transport with a strict rate limiter (e.g., 2 requests per second, burst 1)
	// This means a request can only occur every 500ms.
	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Limit(2), 1),
	)

	// Send 3 requests and measure elapsed time
	start := time.Now()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		req, err := client.newRequest(ctx, http.MethodGet, true, "/dummy", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		resp, err := client.do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}

	elapsed := time.Since(start)
	// 3 requests with rate limit of 2/sec:
	// Req 1: immediate (uses burst token)
	// Req 2: waits 500ms
	// Req 3: waits 500ms
	// Total wait time should be at least 950ms (giving 50ms buffer).
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected rate limiter to delay requests, total time elapsed: %v", elapsed)
	}
}

func TestTransport_Retry503Success(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithMaxRetries(2),
		WithRateLimit(rate.Inf, 1), // Disable rate limit for test speed
	)

	req, err := client.newRequest(context.Background(), http.MethodGet, true, "/dummy", nil)
	if err != nil {
		t.Fatalf("newRequest failed: %v", err)
	}

	resp, err := client.do(req)
	if err != nil {
		t.Fatalf("expected call to succeed after retries, got error: %v", err)
	}
	defer resp.Body.Close()

	finalCount := atomic.LoadInt32(&callCount)
	if finalCount != 3 {
		t.Errorf("expected exactly 3 request attempts, got %d", finalCount)
	}
}

func TestTransport_RetryExceeded(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusServiceUnavailable) // 503
		w.Write([]byte("server unavailable body"))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithMaxRetries(2),
		WithRateLimit(rate.Inf, 1),
	)

	req, err := client.newRequest(context.Background(), http.MethodGet, true, "/dummy", nil)
	if err != nil {
		t.Fatalf("newRequest failed: %v", err)
	}

	resp, err := client.do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected request to fail when retries are exhausted, but it succeeded")
	}

	var wikiErr *wikierrors.WikiError
	if !errors.As(err, &wikiErr) {
		t.Fatalf("expected a WikiError, got %T: %v", err, err)
	}

	if wikiErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", wikiErr.StatusCode)
	}

	if !errors.Is(err, wikierrors.ErrServerError) {
		t.Errorf("expected wrapped ErrServerError, got: %v", err)
	}

	finalCount := atomic.LoadInt32(&callCount)
	if finalCount != 3 { // 1 initial request + 2 retries = 3 total attempts
		t.Errorf("expected exactly 3 attempts, got %d", finalCount)
	}
}

func TestTransport_RetryAfterHeader(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			w.Header().Set("Retry-After", "1") // Ask client to wait 1 second
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithMaxRetries(1),
		WithRateLimit(rate.Inf, 1),
	)

	start := time.Now()
	req, err := client.newRequest(context.Background(), http.MethodGet, true, "/dummy", nil)
	if err != nil {
		t.Fatalf("newRequest failed: %v", err)
	}

	resp, err := client.do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)
	// We expect the client to have waited around 1s before the retry.
	// Give a small buffer of 50ms.
	if elapsed < 950*time.Millisecond {
		t.Errorf("expected Retry-After to wait at least 1s, waited %v", elapsed)
	}

	finalCount := atomic.LoadInt32(&callCount)
	if finalCount != 2 {
		t.Errorf("expected 2 attempts, got %d", finalCount)
	}
}

func TestTransport_NoRetryPermanentError(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusNotFound) // 404 is a permanent error
		w.Write([]byte("not found message"))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithMaxRetries(3),
		WithRateLimit(rate.Inf, 1),
	)

	req, err := client.newRequest(context.Background(), http.MethodGet, true, "/dummy", nil)
	if err != nil {
		t.Fatalf("newRequest failed: %v", err)
	}

	resp, err := client.do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected request to fail with 404")
	}

	if !errors.Is(err, wikierrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound error, got: %v", err)
	}

	finalCount := atomic.LoadInt32(&callCount)
	if finalCount != 1 {
		t.Errorf("expected exactly 1 attempt for permanent error, got %d", finalCount)
	}
}

func TestTransport_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithMaxRetries(3),
		WithRateLimit(rate.Inf, 1),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	req, err := client.newRequest(ctx, http.MethodGet, true, "/dummy", nil)
	if err != nil {
		t.Fatalf("newRequest failed: %v", err)
	}

	_, err = client.do(req)
	if err == nil {
		t.Fatal("expected request to fail with cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error to be context.Canceled, got: %v", err)
	}
}
