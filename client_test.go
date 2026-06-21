// Package wikigo implements the Wikimedia REST API SDK.
package wikigo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	wikierrors "github.com/saluja-ji/wikigo/errors"
	"golang.org/x/time/rate"
)

func TestPagesClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/w/rest.php/v1/page/Albert_Einstein/bare" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": 736,
			"key": "Albert_Einstein",
			"title": "Albert Einstein",
			"latest": {
				"id": 123456789,
				"timestamp": "2026-06-21T08:00:00Z"
			},
			"content_model": "wikitext",
			"html_url": "https://en.wikipedia.org/v1/page/Albert_Einstein/html"
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
	)

	page, err := client.Pages.Get(context.Background(), "Albert Einstein")
	if err != nil {
		t.Fatalf("Pages.Get failed: %v", err)
	}

	if page.ID != 736 {
		t.Errorf("expected page.ID = 736, got %d", page.ID)
	}
	if page.Title != "Albert Einstein" {
		t.Errorf("expected Title = Albert Einstein, got %s", page.Title)
	}
	if page.Latest == nil || page.Latest.ID != 123456789 {
		t.Errorf("expected Latest.ID = 123456789, got %v", page.Latest)
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2026-06-21T08:00:00Z")
	if !page.Latest.Timestamp.Equal(expectedTime) {
		t.Errorf("expected Latest.Timestamp = %v, got %v", expectedTime, page.Latest.Timestamp)
	}
}

func TestPagesClient_GetSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/rest_v1/page/summary/Albert_Einstein" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"type": "standard",
			"title": "Albert_Einstein",
			"displaytitle": "Albert Einstein",
			"pageid": 736,
			"lang": "en",
			"description": "German theoretical physicist",
			"extract": "Albert Einstein was a German theoretical physicist...",
			"extract_html": "<p>Albert Einstein was...</p>",
			"thumbnail": {
				"source": "https://upload.wikimedia.org/Einstein.jpg",
				"width": 320,
				"height": 427
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
	)

	summary, err := client.Pages.GetSummary(context.Background(), "Albert Einstein")
	if err != nil {
		t.Fatalf("Pages.GetSummary failed: %v", err)
	}

	if summary.PageID != 736 {
		t.Errorf("expected PageID = 736, got %d", summary.PageID)
	}
	if summary.Description != "German theoretical physicist" {
		t.Errorf("expected description check, got %s", summary.Description)
	}
	if summary.Thumbnail == nil || summary.Thumbnail.Source != "https://upload.wikimedia.org/Einstein.jpg" {
		t.Errorf("expected valid thumbnail source, got %v", summary.Thumbnail)
	}
}

func TestSearchClient_Pages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/w/rest.php/v1/search/page" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		q := r.URL.Query()
		if q.Get("q") != "jupiter" || q.Get("limit") != "2" {
			t.Errorf("unexpected query parameters: %v", q)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"pages": [
				{
					"id": 16024,
					"key": "Jupiter",
					"title": "Jupiter",
					"excerpt": "Jupiter is the fifth planet from the Sun",
					"description": "fifth planet from the Sun"
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
	)

	resp, err := client.Search.Pages(context.Background(), "jupiter", 2)
	if err != nil {
		t.Fatalf("Search.Pages failed: %v", err)
	}

	if len(resp.Pages) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Pages))
	}
	if resp.Pages[0].ID != 16024 {
		t.Errorf("expected page ID = 16024, got %d", resp.Pages[0].ID)
	}
	if resp.Pages[0].Title != "Jupiter" {
		t.Errorf("expected Title = Jupiter, got %s", resp.Pages[0].Title)
	}
}

func TestRevisionsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/w/rest.php/v1/page/Jupiter/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		q := r.URL.Query()
		if q.Get("limit") != "5" || q.Get("older_than") != "999999" {
			t.Errorf("unexpected parameters: %v", q)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"revisions": [
				{
					"id": 12345,
					"timestamp": "2026-06-21T07:30:00Z",
					"minor": true,
					"size": 4500,
					"delta": 10,
					"comment": "grammar fix",
					"user": {
						"id": 101,
						"name": "WikiEditor"
					}
				}
			],
			"next": "/w/rest.php/v1/page/Jupiter/history?older_than=12345&limit=5"
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
	)

	history, err := client.Revisions.List(context.Background(), "Jupiter", 5, "999999")
	if err != nil {
		t.Fatalf("Revisions.List failed: %v", err)
	}

	if len(history.Revisions) != 1 {
		t.Errorf("expected 1 revision, got %d", len(history.Revisions))
	}
	rev := history.Revisions[0]
	if rev.ID != 12345 {
		t.Errorf("expected revision ID = 12345, got %d", rev.ID)
	}
	if rev.User == nil || rev.User.Name != "WikiEditor" {
		t.Errorf("expected user Name = WikiEditor, got %v", rev.User)
	}
	if history.Continue != "12345" {
		t.Errorf("expected Continue token = 12345, got %s", history.Continue)
	}
}

func TestMediaClient_GetFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/w/rest.php/v1/file/File:The_Blue_Marble.jpg" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"title": "File:The_Blue_Marble.jpg",
			"file_description_url": "https://commons.wikimedia.org/wiki/File:The_Blue_Marble.jpg",
			"latest": {
				"timestamp": "2026-06-21T06:00:00Z",
				"user": {
					"id": 200,
					"name": "UploaderUser"
				}
			},
			"original": {
				"url": "https://upload.wikimedia.org/The_Blue_Marble.jpg",
				"width": 3000,
				"height": 3000
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
	)

	file, err := client.Media.GetFile(context.Background(), "File:The Blue Marble.jpg")
	if err != nil {
		t.Fatalf("Media.GetFile failed: %v", err)
	}

	if file.Title != "File:The_Blue_Marble.jpg" {
		t.Errorf("expected Title = File:The_Blue_Marble.jpg, got %s", file.Title)
	}
	if file.Latest == nil || file.Latest.User == nil || file.Latest.User.Name != "UploaderUser" {
		t.Errorf("expected uploader UploaderUser, got %v", file.Latest)
	}
	if file.Original == nil || file.Original.Width != 3000 {
		t.Errorf("expected width = 3000, got %v", file.Original)
	}
}

func TestClient_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found details"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(rate.Inf, 1),
	)

	_, err := client.Pages.Get(context.Background(), "MissingPage")
	if err == nil {
		t.Fatal("expected call to fail with 404")
	}

	if !errors.Is(err, wikierrors.ErrNotFound) {
		t.Errorf("expected error to be ErrNotFound, got: %v", err)
	}

	var wikiErr *wikierrors.WikiError
	if !errors.As(err, &wikiErr) {
		t.Fatalf("expected a WikiError, got %T: %v", err, err)
	}
	if wikiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", wikiErr.StatusCode)
	}
	if wikiErr.Message != `{"error": "not found details"}` {
		t.Errorf("expected message from response, got: %s", wikiErr.Message)
	}
}
