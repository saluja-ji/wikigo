package wikigo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/saluja-ji/wikigo/models"
)

// List retrieves a segment of the revision history for a specific page.
// It maps to the GET /page/{title}/history endpoint using the Core REST API.
// It handles cursor-based pagination explicitly by returning a Continue token.
func (r *RevisionsClient) List(ctx context.Context, title string, limit int, olderThan string) (*models.RevisionList, error) {
	escapedTitle := sanitizeTitle(title)

	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if olderThan != "" {
		params.Set("older_than", olderThan)
	}

	path := "/page/" + escapedTitle + "/history"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	req, err := r.client.newRequest(ctx, http.MethodGet, true, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Internal response structure to capture the raw "next" field
	var rawResp struct {
		Revisions []models.Revision `json:"revisions"`
		Next      string            `json:"next"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, err
	}

	// Extract the "older_than" cursor from the "next" relative URL path if present
	var continueToken string
	if rawResp.Next != "" {
		// Treat the Next path as a relative URL and parse its query parameters
		if parsedURL, err := url.Parse(rawResp.Next); err == nil {
			continueToken = parsedURL.Query().Get("older_than")
		}
	}

	return &models.RevisionList{
		Revisions: rawResp.Revisions,
		Continue:  continueToken,
	}, nil
}
