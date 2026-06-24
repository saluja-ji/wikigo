// Package wikigo is a client for the Wikimedia API.
package wikigo

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// Client is the main entry point to interact with the Wikimedia API.
type Client struct {
	// Pages handles page content and summaries.
	Pages *PagesClient
	// Search handles search queries.
	Search *SearchClient
	// Revisions handles page edit history.
	Revisions *RevisionsClient
	// Media handles file metadata.
	Media *MediaClient

	cfg        *config
	httpClient *http.Client
}

// PagesClient handles page operations.
type PagesClient struct {
	client *Client
}

// SearchClient handles search operations.
type SearchClient struct {
	client *Client
}

// RevisionsClient handles revision history operations.
type RevisionsClient struct {
	client *Client
}

// MediaClient handles media file operations.
type MediaClient struct {
	client *Client
}

// NewClient creates a new API client.
func NewClient(opts ...Option) *Client {
	c := defaultConfig()
	for _, opt := range opts {
		opt(c)
	}

	// Set default URL if none is provided.
	if c.baseURL == "" {
		c.baseURL = "https://" + c.language + "." + c.project + ".org"
	}

	// Set up rate limiting, retries, and caching.
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

	// Set up sub-clients.
	client.Pages = &PagesClient{client: client}
	client.Search = &SearchClient{client: client}
	client.Revisions = &RevisionsClient{client: client}
	client.Media = &MediaClient{client: client}

	return client
}

// newRequest builds the HTTP request with standard headers and the correct URL.
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

	// Set headers required by Wikimedia.
	req.Header.Set("User-Agent", c.cfg.userAgent)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// do sends the request. If successful, you must close the response body.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}
