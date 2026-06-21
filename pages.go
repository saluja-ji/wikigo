// Package wikigo implements the Wikimedia REST API SDK.
package wikigo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/saluja-ji/wikigo/models"
)

// Get retrieves metadata for a specific wiki page by its title using the Core REST API.
// It maps to the GET /page/{title}/bare endpoint.
func (p *PagesClient) Get(ctx context.Context, title string) (*models.Page, error) {
	escapedTitle := sanitizeTitle(title)
	req, err := p.client.newRequest(ctx, http.MethodGet, true, "/page/"+escapedTitle+"/bare", nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var page models.Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}

	return &page, nil
}

// GetSummary retrieves a structured page summary for previews using the legacy REST API.
// It maps to the GET /page/summary/{title} endpoint.
func (p *PagesClient) GetSummary(ctx context.Context, title string) (*models.Summary, error) {
	escapedTitle := sanitizeTitle(title)
	req, err := p.client.newRequest(ctx, http.MethodGet, false, "/page/summary/"+escapedTitle, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var summary models.Summary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

// sanitizeTitle replaces spaces with underscores and applies URL path escaping.
func sanitizeTitle(title string) string {
	t := strings.ReplaceAll(title, " ", "_")
	return url.PathEscape(t)
}
