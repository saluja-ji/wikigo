// Package main is an example of using the wikigo client.
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
	// 1. Create the client. We set a low rate limit to test rate limiting.
	fmt.Println("Initializing wikigo client...")
	client := wikigo.NewClient(
		wikigo.WithLanguage("en"),
		wikigo.WithProject("wikipedia"),
		wikigo.WithTimeout(15*time.Second),
		wikigo.WithRateLimit(rate.Limit(2), 2), // 2 requests per second
		wikigo.WithMaxRetries(3),
	)

	ctx := context.Background()

	// 2. Get the summary for a page.
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
		// Print a snippet of the text.
		if len(summary.Extract) > 150 {
			fmt.Printf("Extract:     %s...\n", summary.Extract[:150])
		} else {
			fmt.Printf("Extract:     %s\n", summary.Extract)
		}
	}

	// 3. Search for pages.
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

	// 4. Get the edit history of a page.
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

	// 5. Get file details.
	fmt.Println("\n--- 4. Fetching Media File details ---")
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

	// 6. Request a non-existent page to test error handling.
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
