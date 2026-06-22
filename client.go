// Package wikigo provides a production-quality, type-safe Go SDK for the Wikimedia REST API.
// It wraps pages, search, revisions, and media APIs with context propagation,
// structured error handling, rate limiting, and exponential backoff retry logic.
package wikigo

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// Client is the primary entrypoint for the Wikimedia REST API SDK.
// It organizes API methods into resource-scoped sub-clients.
type Client struct {
	// Pages provides operations related to reading page content and summaries.
	Pages *PagesClient
	// Search provides endpoints to query the wiki index.
	Search *SearchClient
	// Revisions provides endpoints to inspect page history revisions.
	Revisions *RevisionsClient
	// Media provides operations to retrieve media file details.
	Media *MediaClient

	cfg        *config
	httpClient *http.Client
}

// PagesClient groups API operations related to pages.
type PagesClient struct {
	client *Client
}

// SearchClient groups API operations related to search.
type SearchClient struct {
	client *Client
}

// RevisionsClient groups API operations related to revisions history.
type RevisionsClient struct {
	client *Client
}

// MediaClient groups API operations related to media files.
type MediaClient struct {
	client *Client
}

// NewClient constructs a new wikigo client using the provided functional options.
func NewClient(opts ...Option) *Client {
	c := defaultConfig()
	for _, opt := range opts {
		opt(c)
	}

	// Build default baseURL if not overridden
	if c.baseURL == "" {
		c.baseURL = "https://" + c.language + "." + c.project + ".org"
	}

	// Wrap transport of target http.Client with our custom retry and rate limiting transport
	var transport http.RoundTripper = newRetryTransport(c.httpClient.Transport, c.rateLimit, c.rateBurst, c.maxRetries)
	if c.cacheEnabled {
		transport = newCacheTransport(transport, c.cacheTTL, c.cacheMaxEntries)
	}

	client := &Client{
		cfg: c,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   c.timeout,
		},
	}

	// Initialize sub-clients
	client.Pages = &PagesClient{client: client}
	client.Search = &SearchClient{client: client}
	client.Revisions = &RevisionsClient{client: client}
	client.Media = &MediaClient{client: client}

	return client
}

// newRequest constructs an HTTP request, injects standard headers (like User-Agent),
// sets the request context, and resolves the correct full API URL.
func (c *Client) newRequest(ctx context.Context, method string, useCoreAPI bool, path string, body io.Reader) (*http.Request, error) {
	baseURL := strings.TrimSuffix(c.cfg.baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var urlStr string
	if useCoreAPI {
		urlStr = baseURL + "/w/rest.php/v1" + path
	} else {
		urlStr = baseURL + "/api/rest_v1" + path
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}

	// Wikimedia API policy requires a custom user agent to prevent request blocks.
	req.Header.Set("User-Agent", c.cfg.userAgent)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// do executes the request and expects the response to have been checked/closed in retryTransport
// if it was an error. Therefore, if err == nil, caller MUST close the response body.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}
