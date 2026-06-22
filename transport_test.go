// Package wikigo implements the Wikimedia REST API SDK.
package wikigo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
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

func TestTransport_CacheSuccess(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("X-Custom-Header", "hello")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("cached content"))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
		WithCache(10*time.Second, 5),
	)

	ctx := context.Background()

	// First request: Cache miss, hits server
	req1, err := client.newRequest(ctx, http.MethodGet, true, "/dummy", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp1, err := client.do(req1)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	body1, err := io.ReadAll(resp1.Body)
	resp1.Body.Close()
	if err != nil {
		t.Fatalf("failed to read body 1: %v", err)
	}
	if string(body1) != "cached content" {
		t.Errorf("expected body 1 to be 'cached content', got %q", body1)
	}
	if resp1.Header.Get("X-Custom-Header") != "hello" {
		t.Errorf("expected X-Custom-Header to be 'hello', got %q", resp1.Header.Get("X-Custom-Header"))
	}

	// Second request: Cache hit, should not hit server
	req2, err := client.newRequest(ctx, http.MethodGet, true, "/dummy", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp2, err := client.do(req2)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	body2, err := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if err != nil {
		t.Fatalf("failed to read body 2: %v", err)
	}
	if string(body2) != "cached content" {
		t.Errorf("expected body 2 to be 'cached content', got %q", body2)
	}
	if resp2.Header.Get("X-Custom-Header") != "hello" {
		t.Errorf("expected X-Custom-Header to be 'hello', got %q", resp2.Header.Get("X-Custom-Header"))
	}

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected exactly 1 call to mock server, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestTransport_CacheNonGET(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("non-get content"))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
		WithCache(10*time.Second, 5),
	)

	ctx := context.Background()

	// Send POST requests
	for i := 0; i < 2; i++ {
		req, err := client.newRequest(ctx, http.MethodPost, true, "/dummy", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		resp, err := client.do(req)
		if err != nil {
			t.Fatalf("POST request failed: %v", err)
		}
		resp.Body.Close()
	}

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("expected 2 calls to mock server, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestTransport_CacheTTL(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ttl content"))
	}))
	defer server.Close()

	// Short TTL of 10ms
	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
		WithCache(10*time.Millisecond, 5),
	)

	ctx := context.Background()

	req1, _ := client.newRequest(ctx, http.MethodGet, true, "/dummy", nil)
	resp1, _ := client.do(req1)
	resp1.Body.Close()

	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)

	req2, _ := client.newRequest(ctx, http.MethodGet, true, "/dummy", nil)
	resp2, _ := client.do(req2)
	resp2.Body.Close()

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("expected 2 calls to mock server (cache expired), got %d", atomic.LoadInt32(&callCount))
	}
}

func TestTransport_CacheCapEviction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("evict content"))
	}))
	defer server.Close()

	// Cache with max capacity of 2
	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
		WithCache(10*time.Second, 2),
	)

	ctx := context.Background()

	// Add 2 entries
	req1, _ := client.newRequest(ctx, http.MethodGet, true, "/dummy1", nil)
	resp1, _ := client.do(req1)
	resp1.Body.Close()

	req2, _ := client.newRequest(ctx, http.MethodGet, true, "/dummy2", nil)
	resp2, _ := client.do(req2)
	resp2.Body.Close()

	// Add a 3rd entry which should trigger eviction of one of the previous two
	req3, _ := client.newRequest(ctx, http.MethodGet, true, "/dummy3", nil)
	resp3, _ := client.do(req3)
	resp3.Body.Close()

	// Check underlying transport/cache state if possible, or just verify cache size is <= 2
	transport := client.httpClient.Transport.(*cacheTransport)
	transport.mu.RLock()
	cacheLen := len(transport.cache)
	transport.mu.RUnlock()

	if cacheLen > 2 {
		t.Errorf("expected cache size to be capped at 2, got %d", cacheLen)
	}
}

func TestTransport_CacheConcurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("concurrent content"))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
		WithCache(10*time.Second, 10),
	)

	ctx := context.Background()
	var wg sync.WaitGroup
	workers := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			path := "/dummy"
			if workerID%2 == 0 {
				path = "/dummy2"
			}
			req, err := client.newRequest(ctx, http.MethodGet, true, path, nil)
			if err != nil {
				t.Errorf("failed to create request in worker: %v", err)
				return
			}
			resp, err := client.do(req)
			if err != nil {
				t.Errorf("request failed in worker: %v", err)
				return
			}
			resp.Body.Close()
		}(i)
	}
	wg.Wait()
}

