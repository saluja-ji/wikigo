package wikigo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/saluja-ji/wikigo/models"
)

// Pages searches for pages matching the query string in title or content.
// It maps to the GET /search/page endpoint using Core REST API.
func (s *SearchClient) Pages(ctx context.Context, query string, limit int) (*models.SearchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	path := "/search/page?" + params.Encode()
	req, err := s.client.newRequest(ctx, http.MethodGet, true, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResp models.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}
