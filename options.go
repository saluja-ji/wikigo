package wikigo

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Option configures the client.
type Option func(*config)

// config stores all settings for the client.
type config struct {
	language        string
	project         string
	baseURL         string
	timeout         time.Duration
	userAgent       string
	rateLimit       rate.Limit
	rateBurst       int
	maxRetries      int
	httpClient      *http.Client
	cacheEnabled    bool
	cacheTTL        time.Duration
	cacheMaxEntries int
}

// defaultConfig returns the default settings.
func defaultConfig() *config {
	return &config{
		language:   "en",
		project:    "wikipedia",
		timeout:    30 * time.Second,
		userAgent:  "wikigo/1.0 (https://github.com/saluja-ji/wikigo; salujapushpit3@gmail.com)",
		rateLimit:  rate.Limit(15), // 15 requests per second by default
		rateBurst:  20,
		maxRetries: 3,
		httpClient: http.DefaultClient,
	}
}

// WithLanguage sets the wiki language (like "en" or "de"). Defaults to "en".
func WithLanguage(lang string) Option {
	return func(c *config) {
		c.language = lang
	}
}

// WithProject sets the project (like "wikipedia" or "wiktionary"). Defaults to "wikipedia".
func WithProject(proj string) Option {
	return func(c *config) {
		c.project = proj
	}
}

// WithBaseURL sets a custom API URL. Useful for testing.
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.baseURL = url
	}
}

// WithTimeout sets the request timeout. Defaults to 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithUserAgent sets a custom User-Agent header required by Wikimedia.
func WithUserAgent(ua string) Option {
	return func(c *config) {
		c.userAgent = ua
	}
}

// WithRateLimit limits how many requests can be made per second.
func WithRateLimit(limit rate.Limit, burst int) Option {
	return func(c *config) {
		c.rateLimit = limit
		c.rateBurst = burst
	}
}

// WithMaxRetries sets how many times to retry failed requests (defaults to 3).
func WithMaxRetries(retries int) Option {
	return func(c *config) {
		c.maxRetries = retries
	}
}

// WithHTTPClient lets you use a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithCache turns on caching for successful GET requests to speed them up.
func WithCache(ttl time.Duration, maxEntries int) Option {
	return func(c *config) {
		c.cacheEnabled = true
		c.cacheTTL = ttl
		c.cacheMaxEntries = maxEntries
	}
}
