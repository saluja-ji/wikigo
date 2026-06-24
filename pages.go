package wikigo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/saluja-ji/wikigo/models"
)

// Get retrieves a page's metadata by its title.
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

// GetSummary gets a short summary of a page for previews.
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

// sanitizeTitle formats the title for use in the URL.
func sanitizeTitle(title string) string {
	t := strings.ReplaceAll(title, " ", "_")
	return url.PathEscape(t)
}
