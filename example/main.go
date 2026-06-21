// Package main provides a runnable example of using the wikigo SDK
// to interact with the live Wikimedia REST API.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/saluja-ji/wikigo"
	wikierrors "github.com/saluja-ji/wikigo/errors"
	"golang.org/x/time/rate"
)

func main() {
	// 1. Initialize the client using Functional Options
	// We configure a rate limiter of 2 requests per second (burst 2) to easily observe rate limiting in action,
	// and a custom User-Agent header (required by Wikimedia API policies).
	fmt.Println("Initializing wikigo client...")
	client := wikigo.NewClient(
		wikigo.WithLanguage("en"),
		wikigo.WithProject("wikipedia"),
		wikigo.WithTimeout(15*time.Second),
		wikigo.WithRateLimit(rate.Limit(2), 2), // 2 requests per second
		wikigo.WithMaxRetries(3),
	)

	ctx := context.Background()

	// 2. Fetch page summary for "Earth"
	fmt.Println("\n--- 1. Fetching Page Summary for 'Earth' ---")
	summary, err := client.Pages.GetSummary(ctx, "Earth")
	if err != nil {
		handleError(err)
	} else {
		fmt.Printf("Title:       %s\n", summary.DisplayTitle)
		fmt.Printf("Description: %s\n", summary.Description)
		fmt.Printf("Revision ID: %s\n", summary.Revision)
		if summary.Thumbnail != nil {
			fmt.Printf("Thumbnail:   %s\n", summary.Thumbnail.Source)
		}
		// Print a small snippet of the extract
		if len(summary.Extract) > 150 {
			fmt.Printf("Extract:     %s...\n", summary.Extract[:150])
		} else {
			fmt.Printf("Extract:     %s\n", summary.Extract)
		}
	}

	// 3. Search for pages containing "Go (programming language)"
	fmt.Println("\n--- 2. Searching for 'Go (programming language)' ---")
	searchResp, err := client.Search.Pages(ctx, "Go (programming language)", 3)
	if err != nil {
		handleError(err)
	} else {
		fmt.Printf("Found %d search results:\n", len(searchResp.Pages))
		for i, page := range searchResp.Pages {
			fmt.Printf("  [%d] Title: %s (Page ID: %d)\n", i+1, page.Title, page.ID)
			if page.Description != nil {
				fmt.Printf("      Description: %s\n", *page.Description)
			}
		}
	}

	// 4. List revisions history for the "Earth" page
	fmt.Println("\n--- 3. Fetching Page Revision History for 'Earth' ---")
	history, err := client.Revisions.List(ctx, "Earth", 3, "")
	if err != nil {
		handleError(err)
	} else {
		fmt.Printf("Revisions list (latest 3):\n")
		for _, rev := range history.Revisions {
			userName := "Anonymous"
			if rev.User != nil {
				userName = rev.User.Name
			}
			fmt.Printf("  - Rev #%d | By: %s | Time: %s | Size: %d bytes\n",
				rev.ID, userName, rev.Timestamp.Format(time.RFC3339), rev.Size)
		}
		fmt.Printf("Continue Token (older_than): %s\n", history.Continue)
	}

	// 5. Retrieve details for a media file
	fmt.Println("\n--- 4. Fetching Media File details ---")
	// Note: commons.wikimedia.org files can also be retrieved. Here we get an image from Wikipedia.
	fileInfo, err := client.Media.GetFile(ctx, "File:The Earth seen from Apollo 17.jpg")
	if err != nil {
		handleError(err)
	} else {
		fmt.Printf("File Title:    %s\n", fileInfo.Title)
		fmt.Printf("Description:   %s\n", fileInfo.FileDescriptionURL)
		if fileInfo.Original != nil {
			fmt.Printf("Original URL:  %s\n", fileInfo.Original.URL)
			fmt.Printf("Dimensions:    %dx%d\n", fileInfo.Original.Width, fileInfo.Original.Height)
		}
	}

	// 6. Test Error Handling (fetching a non-existent page)
	fmt.Println("\n--- 5. Triggering Error Handling (404 Page Not Found) ---")
	_, err = client.Pages.Get(ctx, "ThisPageDefinitelyDoesNotExist123456789")
	if err != nil {
		if errors.Is(err, wikierrors.ErrNotFound) {
			fmt.Println("Successfully detected sentinel error: ErrNotFound!")
		}
		handleError(err)
	}
}

func handleError(err error) {
	var wikiErr *wikierrors.WikiError
	if errors.As(err, &wikiErr) {
		fmt.Fprintf(os.Stderr, "API Error [HTTP %d]: %s\n", wikiErr.StatusCode, wikiErr.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Generic Error: %v\n", err)
	}
}
