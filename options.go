package wikigo

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Option represents a configuration function for customizing the client.
type Option func(*config)

// config holds all configuration parameters for the wikigo client.
type config struct {
	language   string
	project    string
	baseURL    string
	timeout    time.Duration
	userAgent  string
	rateLimit  rate.Limit
	rateBurst  int
	maxRetries int
	httpClient *http.Client
}

// defaultConfig returns a config struct with production-ready default values.
func defaultConfig() *config {
	return &config{
		language:   "en",
		project:    "wikipedia",
		timeout:    30 * time.Second,
		userAgent:  "wikigo/1.0 (https://github.com/saluja-ji/wikigo; salujapushpit3@gmail.com)",
		rateLimit:  rate.Limit(15), // Default to 15 requests per second
		rateBurst:  20,
		maxRetries: 3,
		httpClient: http.DefaultClient,
	}
}

// WithLanguage sets the language sub-domain for the Wikimedia project (e.g., "en", "de").
// Default is "en".
func WithLanguage(lang string) Option {
	return func(c *config) {
		c.language = lang
	}
}

// WithProject sets the Wikimedia project domain (e.g., "wikipedia", "wiktionary", "wikibooks").
// Default is "wikipedia".
func WithProject(proj string) Option {
	return func(c *config) {
		c.project = proj
	}
}

// WithBaseURL overrides the constructed host URL entirely.
// Useful for custom deployments or local mock servers during tests.
// E.g., "http://127.0.0.1:8080".
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

// WithTimeout sets the timeout duration for client-wide request execution.
// Default is 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithUserAgent overrides the HTTP User-Agent header value sent on all requests.
// A descriptive User-Agent is required under the Wikimedia API usage guidelines.
func WithUserAgent(ua string) Option {
	return func(c *config) {
		c.userAgent = ua
	}
}

// WithRateLimit sets the parameters for the client rate limiter.
// rateLimit specifies the maximum requests per second, and burst specifies the token burst size.
func WithRateLimit(limit rate.Limit, burst int) Option {
	return func(c *config) {
		c.rateLimit = limit
		c.rateBurst = burst
	}
}

// WithMaxRetries sets the maximum number of retry attempts for failed requests.
// Only HTTP 429 and 503 errors are retried. Default is 3.
func WithMaxRetries(retries int) Option {
	return func(c *config) {
		c.maxRetries = retries
	}
}

// WithHTTPClient overrides the underlying http.Client used by the SDK.
// The SDK wraps the client's transport with its own retry and rate-limiting transport.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.httpClient = client
		}
	}
}
