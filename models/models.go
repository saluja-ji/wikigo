// Package models defines the strongly typed data models used by the wikigo
// Wikimedia REST API client.
package models

import (
	"time"
)

// Thumbnail represents metadata about a page's or file's thumbnail image.
type Thumbnail struct {
	// Source is the URL of the thumbnail image.
	Source string `json:"source"`
	// Width is the width of the thumbnail in pixels.
	Width int `json:"width"`
	// Height is the height of the thumbnail in pixels.
	Height int `json:"height"`
}

// ImageInfo represents detailed metadata for an image (original or preferred version).
type ImageInfo struct {
	// Source is the URL of the image.
	Source string `json:"source"`
	// Width is the width of the image in pixels.
	Width int `json:"width"`
	// Height is the height of the image in pixels.
	Height int `json:"height"`
}

// NamespaceInfo represents a wiki namespace identifier and label.
type NamespaceInfo struct {
	// ID is the namespace ID (e.g., 0 for main articles).
	ID int `json:"id"`
	// Text is the namespace prefix text.
	Text string `json:"text"`
}

// TitleVariants represents normalized, canonical, and display variants of a title.
type TitleVariants struct {
	// Canonical is the canonical form of the title.
	Canonical string `json:"canonical"`
	// Normalized is the normalized form of the title.
	Normalized string `json:"normalized"`
	// Display is the display form of the title.
	Display string `json:"display"`
}

// LinkItem represents a link to a page version.
type LinkItem struct {
	// Page is the URL to the wiki page.
	Page string `json:"page"`
	// Revisions is the URL to the page history.
	Revisions string `json:"revisions"`
	// Edit is the URL to edit the page.
	Edit string `json:"edit"`
	// Talk is the URL to the talk page.
	Talk string `json:"talk"`
}

// ContentURLs represents the desktop and mobile page URLs.
type ContentURLs struct {
	// Desktop contains desktop-specific URLs.
	Desktop *LinkItem `json:"desktop,omitempty"`
	// Mobile contains mobile-specific URLs.
	Mobile *LinkItem `json:"mobile,omitempty"`
}

// RevisionInfo represents basic metadata for a page revision.
type RevisionInfo struct {
	// ID is the unique revision identifier.
	ID int64 `json:"id"`
	// Timestamp is the date and time when the revision was saved.
	Timestamp time.Time `json:"timestamp"`
}

// License represents licensing metadata.
type License struct {
	// URL is the web address of the license terms.
	URL string `json:"url"`
	// Title is the name of the license.
	Title string `json:"title"`
}

// Page represents metadata for a wiki page as returned by the bare endpoint.
type Page struct {
	// ID is the unique page identifier.
	ID int64 `json:"id"`
	// Key is the URL-friendly key of the page.
	Key string `json:"key"`
	// Title is the human-readable display title.
	Title string `json:"title"`
	// Latest holds metadata about the page's latest revision.
	Latest *RevisionInfo `json:"latest,omitempty"`
	// ContentModel is the content model type (e.g., wikitext).
	ContentModel string `json:"content_model"`
	// License contains license information for this page.
	License *License `json:"license,omitempty"`
	// HTMLURL is the endpoint URL to retrieve the Parsoid HTML of the page.
	HTMLURL string `json:"html_url"`
}

// Summary represents the page summary details used for previews.
type Summary struct {
	// Type indicates the page type (e.g., "standard").
	Type string `json:"type"`
	// Title is the normalized page title.
	Title string `json:"title"`
	// DisplayTitle is the formatted display title.
	DisplayTitle string `json:"displaytitle"`
	// Namespace details the namespace of the page.
	Namespace *NamespaceInfo `json:"namespace,omitempty"`
	// WikibaseItem is the Wikidata entity ID.
	WikibaseItem string `json:"wikibase_item"`
	// Titles contains canonical, normalized and display variants of the title.
	Titles *TitleVariants `json:"titles,omitempty"`
	// PageID is the unique page ID.
	PageID int64 `json:"pageid"`
	// Thumbnail is the lead image thumbnail metadata.
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
	// OriginalImage is the full resolution lead image metadata.
	OriginalImage *ImageInfo `json:"originalimage,omitempty"`
	// Lang is the language code of the page.
	Lang string `json:"lang"`
	// Dir is the text direction (e.g., "ltr").
	Dir string `json:"dir"`
	// Revision is the latest revision ID.
	Revision string `json:"revision"`
	// Tid is the transaction ID of the page revision.
	Tid string `json:"tid"`
	// Timestamp is the date and time of the latest edit.
	Timestamp time.Time `json:"timestamp"`
	// Description is the short Wikidata summary description.
	Description string `json:"description"`
	// Extract is a plain text summary extract of the page.
	Extract string `json:"extract"`
	// ExtractHTML is the HTML summary extract of the page.
	ExtractHTML string `json:"extract_html"`
	// ContentURLs provides external links to the page.
	ContentURLs *ContentURLs `json:"content_urls,omitempty"`
}

// User represents a wiki contributor.
type User struct {
	// ID is the unique user ID of the editor, or nil/0 for anonymous edits.
	ID int64 `json:"id"`
	// Name is the username of the contributor.
	Name string `json:"name"`
}

// Revision represents a single edit revision in history.
type Revision struct {
	// ID is the unique revision ID.
	ID int64 `json:"id"`
	// Timestamp is the date and time when the edit was saved.
	Timestamp time.Time `json:"timestamp"`
	// Minor indicates if the edit was marked as minor.
	Minor bool `json:"minor"`
	// Size is the size of the page in bytes.
	Size int64 `json:"size"`
	// Delta is the size difference compared to the previous revision in bytes.
	Delta int64 `json:"delta"`
	// Comment is the edit summary summary message.
	Comment string `json:"comment"`
	// User is the contributor who made the edit.
	User *User `json:"user,omitempty"`
}

// RevisionList represents a segment of a page's revision history.
type RevisionList struct {
	// Revisions is the list of revision edits returned.
	Revisions []Revision `json:"revisions"`
	// Continue is the pagination token indicating where to continue (corresponds to older_than revision ID).
	Continue string `json:"continue,omitempty"`
}

// SearchResult represents a single search match in the wiki index.
type SearchResult struct {
	// ID is the unique page identifier.
	ID int64 `json:"id"`
	// Key is the URL-friendly key of the matching page.
	Key string `json:"key"`
	// Title is the readable display title.
	Title string `json:"title"`
	// Excerpt is the snippet containing the highlighted match term.
	Excerpt string `json:"excerpt"`
	// MatchedTitle is the title of the redirect if matched via redirect.
	MatchedTitle *string `json:"matched_title,omitempty"`
	// Description is the Wikidata description.
	Description *string `json:"description,omitempty"`
	// Thumbnail is the page thumbnail if available.
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
}

// SearchResponse represents the search results page wrapper.
type SearchResponse struct {
	// Pages is the slice of matching search results.
	Pages []SearchResult `json:"pages"`
}

// LatestFileInfo represents upload metadata of the latest file version.
type LatestFileInfo struct {
	// Timestamp is the upload date and time.
	Timestamp time.Time `json:"timestamp"`
	// User is the editor who uploaded the file.
	User *User `json:"user,omitempty"`
}

// PreferredFileInfo represents preview metadata for the preferred file resolution.
type PreferredFileInfo struct {
	// MediaType indicates the media classification (e.g. BITMAP, VIDEO, AUDIO).
	MediaType string `json:"mediatype"`
	// Size is the size of the file in bytes.
	Size int64 `json:"size"`
	// Width is the horizontal dimension in pixels.
	Width int `json:"width"`
	// Height is the vertical dimension in pixels.
	Height int `json:"height"`
	// Duration is the length in seconds (applicable to audio/video).
	Duration *float64 `json:"duration,omitempty"`
	// URL is the download path.
	URL string `json:"url"`
}

// OriginalFileInfo represents the full resolution file details.
type OriginalFileInfo struct {
	// URL is the download path of the original full-size file.
	URL string `json:"url"`
	// Width is the horizontal dimension of the original file in pixels.
	Width int `json:"width"`
	// Height is the vertical dimension of the original file in pixels.
	Height int `json:"height"`
}

// File represents file metadata for a media resource.
type File struct {
	// Title is the standard title of the file, including the File: namespace prefix.
	Title string `json:"title"`
	// FileDescriptionURL is the URL linking to the Commons description page.
	FileDescriptionURL string `json:"file_description_url"`
	// Latest contains details of the latest upload transaction.
	Latest *LatestFileInfo `json:"latest,omitempty"`
	// Preferred holds preview metadata for the file.
	Preferred *PreferredFileInfo `json:"preferred,omitempty"`
	// Original holds original resource download URL.
	Original *OriginalFileInfo `json:"original,omitempty"`
	// Thumbnail contains metadata for the default size thumbnail.
	Thumbnail *Thumbnail `json:"thumbnail,omitempty"`
}
