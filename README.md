# wikigo

`wikigo` is a , type-safe Go SDK for the Wikimedia REST API. It wraps the API with proper typing, context propagation, error wrapping, robust rate limiting, and  backoff retry logic.

---

## Features

- **Functional Options:** Modular and backward-compatible client initialization (e.g. custom subdomains, timeouts, user-agents, and retries).
- **Context-Aware:** `context.Context` is passed as the first parameter to all public API calls to support cancellations, timeouts, and deadlines.
- **Resource-Scoped Sub-Clients:** Intuitive structure matching the API resource layout (`client.Pages`, `client.Search`, `client.Revisions`, `client.Media`).
- **Custom Transport (`retryTransport`):** An invisible `http.RoundTripper` layer handling:
  - **Token Bucket Rate Limiting:** Smooth request pacing using `golang.org/x/time/rate`.
  - **Exponential Backoff & Jitter:** Random ±20% jitter to prevent thundering herd problems.
  - **Selective Retries:** Retries only transient server errors (`HTTP 503`) and rate-limit blocks (`HTTP 429`), respecting the `Retry-After` header.
- **Dedicated Errors Package:** Explicit error handling with status code preservation (`WikiError` struct) and sentinel errors for easy comparisons (`errors.Is`).
- **No External Web Frameworks:** Built purely on top of Go's standard library and the official x/time package.

---

## Installation

```bash
go get github.com/saluja-ji/wikigo
```

---

## Directory Layout

```
├── errors/
│   └── errors.go       # Sentinel errors and WikiError definition
├── models/
│   └── models.go       # Strongly typed JSON representation structs
├── client.go           # Core client and sub-client definitions
├── options.go          # Configuration options (Functional Options)
├── transport.go        # Custom retry & rate limiting http.RoundTripper
├── pages.go            # Pages sub-client implementations
├── search.go           # Search sub-client implementations
├── revisions.go        # Revisions sub-client implementations
└── media.go            # Media sub-client implementations
```

---

## Getting Started

Here is a quick example of constructing the client and utilizing its sub-clients:

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/saluja-ji/wikigo"
	wikierrors "github.com/saluja-ji/wikigo/errors"
	"golang.org/x/time/rate"
)

func main() {
	// 1. Initialize client using Functional Options
	client := wikigo.NewClient(
		wikigo.WithLanguage("en"),
		wikigo.WithProject("wikipedia"),
		wikigo.WithTimeout(10*time.Second),
		wikigo.WithRateLimit(rate.Limit(10), 10), // Limit to 10 req/sec
		wikigo.WithMaxRetries(3),
	)

	ctx := context.Background()

	// 2. Fetch page summary details
	summary, err := client.Pages.GetSummary(ctx, "Earth")
	if err != nil {
		handleError(err)
		return
	}
	fmt.Printf("Summary of %s: %s\n", summary.DisplayTitle, summary.Description)

	// 3. Search for pages matching a query
	searchResp, err := client.Search.Pages(ctx, "Go (programming language)", 3)
	if err != nil {
		handleError(err)
		return
	}
	for _, page := range searchResp.Pages {
		fmt.Printf("Search Match: %s (ID: %d)\n", page.Title, page.ID)
	}
}

func handleError(err error) {
	var wikiErr *wikierrors.WikiError
	if errors.As(err, &wikiErr) {
		fmt.Printf("API Error: [HTTP %d] %v\n", wikiErr.StatusCode, wikiErr.Err)
	} else {
		fmt.Printf("Standard Error: %v\n", err)
	}
}
```

---

## Client API Reference

### `client.Pages`
- **`Get(ctx, title)`**: Retrieves basic page metadata (`*models.Page`) from `/page/{title}/bare` using Core REST API.
- **`GetSummary(ctx, title)`**: Retrieves a quick, preview-friendly summary (`*models.Summary`) from `/page/summary/{title}` using the Legacy/Summary API.

### `client.Search`
- **`Pages(ctx, query, limit)`**: Queries matching articles (`*models.SearchResponse`) containing title or content terms.

### `client.Revisions`
- **`List(ctx, title, limit, olderThan)`**: Fetches history edits list (`*models.RevisionList`). Does not auto-paginate. You must pass the returned `Continue` token cursor back into consecutive list requests to retrieve older segments.

### `client.Media`
- **`GetFile(ctx, title)`**: Retrieves file detail metadata (`*models.File`) from `/file/{title}` for media files (the title must contain the `File:` prefix).

---

## Detailed Design Details

### Rate Limiting
A token bucket rate limiter is embedded directly inside `retryTransport`. The client blocks appropriately using `Wait()` before a request goes out:
```go
if err := t.limiter.Wait(req.Context()); err != nil {
    return nil, fmt.Errorf("rate limiter wait: %w", err)
}
```
This is fully transparent to callers, ensuring you stay within Wikimedia limits without doing any custom interval spacing.

### Transparent Retries & Retry-After
If a request encounters a rate limit (`HTTP 429`) or a service outage (`HTTP 503`), the client automatically starts a retry cycle:
1. It parses and respects the `Retry-After` header if it's sent back as part of a `429` (supports both integer seconds and HTTP dates).
2. If `Retry-After` is missing, it calculates a backoff with exponential scaling:
   $$\text{backoff} = 1\text{s} \times 2^{\text{attempt}}$$
3. Applies a $\pm20\%$ randomized jitter to the backoff to prevent thundering herd locks.
4. Permanent errors like `404 Not Found`, `400 Bad Request`, or `401 Unauthorized` return immediately without any retry loops.

### Preserved Error Hierarchy
Any failure >= 400 returns a wrapped `WikiError` struct:
```go
type WikiError struct {
    StatusCode int    // Preserves original HTTP status code (e.g. 404)
    Err        error  // Standard Sentinel Error
    Message    string // Descriptive payload from the API body
}
```
You can inspect errors directly using `errors.Is` to check for specific error states:
- `errors.Is(err, wikierrors.ErrNotFound)`
- `errors.Is(err, wikierrors.ErrRateLimited)`
- `errors.Is(err, wikierrors.ErrUnauthorized)`

---

## Running Tests

All unit tests compile and run deterministically against mock servers. **No live requests** are ever executed during the tests.

To run the unit tests:
```bash
go test -v ./...
```
