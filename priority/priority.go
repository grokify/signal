// Package priority handles hand-curated priority links.
package priority

import (
	"encoding/json"
	"os"
	"time"

	"github.com/grokify/signal/entry"
)

// Link represents a hand-curated priority link.
type Link struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Author      string    `json:"author,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	ContentHTML string    `json:"content_html,omitempty"` // Full article content
	Rank        int       `json:"rank,omitempty"`         // Lower = higher priority
	FeedTitle   string    `json:"feedTitle,omitempty"`
	FeedURL     string    `json:"feedUrl,omitempty"`

	// Image for visual pins (LinkedIn posts, articles with hero images)
	Image    string `json:"image,omitempty"`    // Main image URL
	ImageAlt string `json:"imageAlt,omitempty"` // Alt text for image

	// Source platform metadata (linkedin, twitter, etc.)
	Source *Source `json:"source,omitempty"`

	// Discussion links (HackerNews, Reddit, Lobsters, etc.)
	Discussions []Discussion `json:"discussions,omitempty"`
}

// Source represents metadata about the content source platform.
type Source struct {
	Platform string `json:"platform"`       // "linkedin", "twitter", "mastodon", etc.
	Author   string `json:"author,omitempty"` // Platform-specific author name/handle
	PostID   string `json:"postId,omitempty"` // Platform-specific post ID
}

// Discussion represents a link to a discussion forum.
type Discussion struct {
	Platform string `json:"platform"`          // "hackernews", "reddit", "lobsters", etc.
	URL      string `json:"url"`               // Full URL to the discussion
	ID       string `json:"id,omitempty"`      // Platform-specific ID (e.g., HN item ID)
	Score    int    `json:"score,omitempty"`   // Upvotes/points at time of capture
	Comments int    `json:"comments,omitempty"` // Comment count at time of capture
}

// Links represents a collection of priority links, possibly organized by time period.
type Links struct {
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Period      string    `json:"period,omitempty"` // e.g., "2026-02" for monthly files
	Updated     time.Time `json:"updated"`
	Links       []Link    `json:"links"`
}

// ReadFile reads priority links from a JSON file.
func ReadFile(filename string) (*Links, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var links Links
	if err := json.Unmarshal(data, &links); err != nil {
		return nil, err
	}
	return &links, nil
}

// WriteFile writes priority links to a JSON file.
func (l *Links) WriteFile(filename string) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// ToEntries converts priority links to feed entries.
func (l *Links) ToEntries() []entry.Entry {
	entries := make([]entry.Entry, len(l.Links))
	for i, link := range l.Links {
		date := link.Date
		if date.IsZero() {
			date = l.Updated
		}

		// Convert discussions
		var discussions []entry.Discussion
		for _, d := range link.Discussions {
			discussions = append(discussions, entry.Discussion{
				Platform: d.Platform,
				URL:      d.URL,
				ID:       d.ID,
				Score:    d.Score,
				Comments: d.Comments,
			})
		}

		// Convert source
		var source *entry.Source
		if link.Source != nil {
			source = &entry.Source{
				Platform: link.Source.Platform,
				Author:   link.Source.Author,
				PostID:   link.Source.PostID,
			}
		}

		entries[i] = entry.Entry{
			ID:     entry.GenerateID(link.URL, date),
			Title:  link.Title,
			URL:    link.URL,
			Author: link.Author,
			Date:   date,
			Feed: entry.FeedMeta{
				Title: link.FeedTitle,
				URL:   link.FeedURL,
			},
			Tags:         link.Tags,
			Summary:      link.Summary,
			Content:      link.ContentHTML,
			Image:        link.Image,
			ImageAlt:     link.ImageAlt,
			Source:       source,
			IsPriority:   true,
			PriorityRank: link.Rank,
			Discussions:  discussions,
		}
	}
	return entries
}
