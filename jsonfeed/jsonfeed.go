// Package jsonfeed implements the JSON Feed 1.1 specification.
// See https://jsonfeed.org/version/1.1 for the full specification.
package jsonfeed

import (
	"encoding/json"
	"os"
	"time"
)

const (
	// Version is the JSON Feed version URL.
	Version = "https://jsonfeed.org/version/1.1"
)

// Feed represents a JSON Feed 1.1 feed.
type Feed struct {
	Version     string   `json:"version"`
	Title       string   `json:"title"`
	HomePageURL string   `json:"home_page_url,omitempty"`
	FeedURL     string   `json:"feed_url,omitempty"`
	Description string   `json:"description,omitempty"`
	UserComment string   `json:"user_comment,omitempty"`
	NextURL     string   `json:"next_url,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Favicon     string   `json:"favicon,omitempty"`
	Authors     []Author `json:"authors,omitempty"`
	Language    string   `json:"language,omitempty"`
	Expired     bool     `json:"expired,omitempty"`
	Items       []Item   `json:"items"`

	// Signal extensions (prefixed with underscore per JSON Feed spec)
	SignalGenerated string `json:"_signal_generated,omitempty"`
	SignalPeriod    string `json:"_signal_period,omitempty"` // e.g., "2026-02" for monthly files
}

// Author represents a JSON Feed author.
type Author struct {
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

// Item represents a JSON Feed item.
type Item struct {
	ID            string       `json:"id"`
	URL           string       `json:"url,omitempty"`
	ExternalURL   string       `json:"external_url,omitempty"`
	Title         string       `json:"title,omitempty"`
	ContentHTML   string       `json:"content_html,omitempty"`
	ContentText   string       `json:"content_text,omitempty"`
	Summary       string       `json:"summary,omitempty"`
	Image         string       `json:"image,omitempty"`
	BannerImage   string       `json:"banner_image,omitempty"`
	DatePublished string       `json:"date_published,omitempty"`
	DateModified  string       `json:"date_modified,omitempty"`
	Authors       []Author     `json:"authors,omitempty"`
	Tags          []string     `json:"tags,omitempty"`
	Language      string       `json:"language,omitempty"`
	Attachments   []Attachment `json:"attachments,omitempty"`

	// Signal extensions
	SignalFeedTitle   string              `json:"_signal_feed_title,omitempty"`
	SignalFeedURL     string              `json:"_signal_feed_url,omitempty"`
	SignalPriority    bool                `json:"_signal_priority,omitempty"`
	SignalRank        int                 `json:"_signal_rank,omitempty"`
	SignalDiscussions []SignalDiscussion  `json:"_signal_discussions,omitempty"`
	SignalSource      *SignalSource       `json:"_signal_source,omitempty"`
}

// SignalSource represents metadata about the content source platform.
type SignalSource struct {
	Platform string `json:"platform"`         // "linkedin", "twitter", "mastodon", etc.
	Author   string `json:"author,omitempty"` // Platform-specific author name/handle
	PostID   string `json:"postId,omitempty"` // Platform-specific post ID
}

// SignalDiscussion represents a link to a discussion forum.
type SignalDiscussion struct {
	Platform string `json:"platform"`           // "hackernews", "reddit", "lobsters", etc.
	URL      string `json:"url"`                // Full URL to the discussion
	ID       string `json:"id,omitempty"`       // Platform-specific ID
	Score    int    `json:"score,omitempty"`    // Upvotes/points at time of capture
	Comments int    `json:"comments,omitempty"` // Comment count at time of capture
}

// Attachment represents a JSON Feed attachment.
type Attachment struct {
	URL               string `json:"url"`
	MIMEType          string `json:"mime_type"`
	Title             string `json:"title,omitempty"`
	SizeInBytes       int64  `json:"size_in_bytes,omitempty"`
	DurationInSeconds int    `json:"duration_in_seconds,omitempty"`
}

// NewFeed creates a new JSON Feed with the required fields.
func NewFeed(title string) *Feed {
	return &Feed{
		Version:        Version,
		Title:          title,
		Items:          []Item{},
		SignalGenerated: time.Now().UTC().Format(time.RFC3339),
	}
}

// AddItem adds an item to the feed.
func (f *Feed) AddItem(item Item) {
	f.Items = append(f.Items, item)
}

// WriteFile writes the feed to a JSON file.
func (f *Feed) WriteFile(filename string) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// ReadFile reads a feed from a JSON file.
func ReadFile(filename string) (*Feed, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var feed Feed
	if err := json.Unmarshal(data, &feed); err != nil {
		return nil, err
	}
	return &feed, nil
}

// ToJSON returns the feed as indented JSON bytes.
func (f *Feed) ToJSON() ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}
