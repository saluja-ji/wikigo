// Package models defines data models for the wikigo client.
package models

import (
	"time"
)

// Thumbnail represents a page's thumbnail image.
type Thumbnail struct {
	// Source is the image URL.
	Source string `json:"source"`
	// Width in pixels.
	Width int `json:"width"`
	// Height in pixels.
	Height int `json:"height"`
}

// ImageInfo represents metadata for an image.
type ImageInfo struct {
	// Source is the image URL.
	Source string `json:"source"`
	// Width in pixels.
	Width int `json:"width"`
	// Height in pixels.
	Height int `json:"height"`
}

// NamespaceInfo holds wiki namespace details.
type NamespaceInfo struct {
	// ID is the namespace ID (e.g. 0 for articles).
	ID int `json:"id"`
	// Text is the namespace name.
	Text string `json:"text"`
}

// TitleVariants holds different versions of a page title.
type TitleVariants struct {
	// Canonical form.
	Canonical string `json:"canonical"`
	// Normalized form.
	Normalized string `json:"normalized"`
	// Display form.
	Display string `json:"display"`
}

// LinkItem holds URLs for different actions on a page.
type LinkItem struct {
	// Page URL.
	Page string `json:"page"`
	// History URL.
	Revisions string `json:"revisions"`
	// Edit URL.
	Edit string `json:"edit"`
	// Talk URL.
	Talk string `json:"talk"`
}

// ContentURLs has links to desktop and mobile versions.
type ContentURLs struct {
	// Desktop URLs.
	Desktop *LinkItem `json:"desktop,omitempty"`
	// Mobile URLs.
	Mobile *LinkItem `json:"mobile,omitempty"`
}

// RevisionInfo is revision metadata.
type RevisionInfo struct {
	// ID is the revision identifier.
	ID int64 `json:"id"`
	// Timestamp of the edit.
	Timestamp time.Time `json:"timestamp"`
}

// License is licensing metadata.
type License struct {
	// URL of the license.
	URL string `json:"url"`
	// Title of the license.
	Title string `json:"title"`
}

// Page is standard wiki page info.
type Page struct {
	// ID of the page.
	ID int64 `json:"id"`
	// Key is the URL name.
	Key string `json:"key"`
	// Title is the readable name.
	Title string `json:"title"`
	// Latest is revision info.
	Latest *RevisionInfo `json:"latest,omitempty"`
	// ContentModel (e.g., "wikitext").
	ContentModel string `json:"content_model"`
	// License info.
	License *License `json:"license,omitempty"`
	// HTMLURL is the page HTML source URL.
	HTMLURL string `json:"html_url"`
}

// Summary is a short overview of a page.
type Summary struct {
	// Type of page.
	Type string `json:"type"`
	// Title of page.
	Title string `json:"title"`
	// DisplayTitle formatted for display.
	DisplayTitle string `json:"displaytitle"`
	// Namespace details.
	Namespace *NamespaceInfo `json:"namespace,omitempty"`
	// WikibaseItem is the Wikidata ID.
	WikibaseItem string `json:"wikibase_item"`
	// Titles holds title variants.
	Titles *TitleVariants `json:"titles,omitempty"`
	// PageID is the unique page ID.
	PageID int64 `json:"pageid"`
	// Thumbnail image.
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
	// OriginalImage full version.
	OriginalImage *ImageInfo `json:"originalimage,omitempty"`
	// Lang is the language code.
	Lang string `json:"lang"`
	// Dir is text direction (ltr/rtl).
	Dir string `json:"dir"`
	// Revision ID.
	Revision string `json:"revision"`
	// Tid is the transaction ID.
	Tid string `json:"tid"`
	// Timestamp of last edit.
	Timestamp time.Time `json:"timestamp"`
	// Description is the short Wikidata summary.
	Description string `json:"description"`
	// Extract is the text summary.
	Extract string `json:"extract"`
	// ExtractHTML is the HTML summary.
	ExtractHTML string `json:"extract_html"`
	// ContentURLs contains page links.
	ContentURLs *ContentURLs `json:"content_urls,omitempty"`
}

// User is a wiki editor.
type User struct {
	// ID of the user (0 if anonymous).
	ID int64 `json:"id"`
	// Name of the user.
	Name string `json:"name"`
}

// Revision represents a single page edit.
type Revision struct {
	// ID of the edit.
	ID int64 `json:"id"`
	// Timestamp of the edit.
	Timestamp time.Time `json:"timestamp"`
	// Minor indicates if it was a minor edit.
	Minor bool `json:"minor"`
	// Size of the page in bytes.
	Size int64 `json:"size"`
	// Delta is the size change in bytes.
	Delta int64 `json:"delta"`
	// Comment is the edit description.
	Comment string `json:"comment"`
	// User who made the edit.
	User *User `json:"user,omitempty"`
}

// RevisionList is a list of revisions.
type RevisionList struct {
	// Revisions list.
	Revisions []Revision `json:"revisions"`
	// Continue token for next page.
	Continue string `json:"continue,omitempty"`
}

// SearchResult is a search match.
type SearchResult struct {
	// ID of the page.
	ID int64 `json:"id"`
	// Key is the URL name.
	Key string `json:"key"`
	// Title of the page.
	Title string `json:"title"`
	// Excerpt with matching terms.
	Excerpt string `json:"excerpt"`
	// MatchedTitle (if redirected).
	MatchedTitle *string `json:"matched_title,omitempty"`
	// Description from Wikidata.
	Description *string `json:"description,omitempty"`
	// Thumbnail image.
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
}

// SearchResponse holds search results.
type SearchResponse struct {
	// Pages matches.
	Pages []SearchResult `json:"pages"`
}

// LatestFileInfo details the latest upload.
type LatestFileInfo struct {
	// Timestamp of upload.
	Timestamp time.Time `json:"timestamp"`
	// User who uploaded.
	User *User `json:"user,omitempty"`
}

// PreferredFileInfo is preview info for a file.
type PreferredFileInfo struct {
	// MediaType classification (image/video/etc).
	MediaType string `json:"mediatype"`
	// Size in bytes.
	Size int64 `json:"size"`
	// Width in pixels.
	Width int `json:"width"`
	// Height in pixels.
	Height int `json:"height"`
	// Duration in seconds (video/audio).
	Duration *float64 `json:"duration,omitempty"`
	// URL to download.
	URL string `json:"url"`
}

// OriginalFileInfo is original file info.
type OriginalFileInfo struct {
	// URL to download.
	URL string `json:"url"`
	// Width in pixels.
	Width int `json:"width"`
	// Height in pixels.
	Height int `json:"height"`
}

// File represents media metadata.
type File struct {
	// Title of the file.
	Title string `json:"title"`
	// FileDescriptionURL is the Commons page.
	FileDescriptionURL string `json:"file_description_url"`
	// Latest upload details.
	Latest *LatestFileInfo `json:"latest,omitempty"`
	// Preferred preview details.
	Preferred *PreferredFileInfo `json:"preferred,omitempty"`
	// Original file details.
	Original *OriginalFileInfo `json:"original,omitempty"`
	// Thumbnail details.
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
}
