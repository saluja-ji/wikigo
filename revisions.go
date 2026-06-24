package wikigo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/saluja-ji/wikigo/models"
)

// List returns the revision history for a page, with pagination.
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

	// Temporary struct to parse the raw API response.
	var rawResp struct {
		Revisions []models.Revision `json:"revisions"`
		Next      string            `json:"next"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, err
	}

	// Extract the pagination cursor from the next link.
	var continueToken string
	if rawResp.Next != "" {
		if parsedURL, err := url.Parse(rawResp.Next); err == nil {
			continueToken = parsedURL.Query().Get("older_than")
		}
	}

	return &models.RevisionList{
		Revisions: rawResp.Revisions,
		Continue:  continueToken,
	}, nil
}
