// Package entry defines the core feed entry types for Signal output.
package entry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/grokify/signal/jsonfeed"
)

// Entry represents a single feed entry in the aggregated output.
type Entry struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	URL          string       `json:"url"`
	Author       string       `json:"author,omitempty"`
	Date         time.Time    `json:"date"`
	Feed         FeedMeta     `json:"feed"`
	Tags         []string     `json:"tags,omitempty"`
	Summary      string       `json:"summary,omitempty"`
	Content      string       `json:"content,omitempty"`
	Image        string       `json:"image,omitempty"`        // Main image URL
	ImageAlt     string       `json:"imageAlt,omitempty"`     // Alt text for image
	Source       *Source      `json:"source,omitempty"`       // Platform source metadata
	IsPriority   bool         `json:"isPriority,omitempty"`   // Hand-curated priority link
	PriorityRank int          `json:"priorityRank,omitempty"` // Ordering for priority links
	Discussions  []Discussion `json:"discussions,omitempty"`  // Links to discussions (HN, Reddit, etc.)
}

// Source represents metadata about the content source platform.
type Source struct {
	Platform string `json:"platform"`         // "linkedin", "twitter", "mastodon", etc.
	Author   string `json:"author,omitempty"` // Platform-specific author name/handle
	PostID   string `json:"postId,omitempty"` // Platform-specific post ID
}

// Discussion represents a link to a discussion forum.
type Discussion struct {
	Platform string `json:"platform"`           // "hackernews", "reddit", "lobsters", etc.
	URL      string `json:"url"`                // Full URL to the discussion
	ID       string `json:"id,omitempty"`       // Platform-specific ID (e.g., HN item ID)
	Score    int    `json:"score,omitempty"`    // Upvotes/points at time of capture
	Comments int    `json:"comments,omitempty"` // Comment count at time of capture
}

// FeedMeta contains metadata about the source feed.
type FeedMeta struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	IconURL string `json:"iconUrl,omitempty"`
}

// GenerateID creates a unique ID for an entry based on URL and date.
func GenerateID(url string, date time.Time) string {
	data := url + date.Format(time.RFC3339)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// Feed represents the complete aggregated feed output.
type Feed struct {
	Generated   time.Time `json:"generated"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	HomeURL     string    `json:"homeUrl,omitempty"`
	Entries     []Entry   `json:"entries"`
}

// NewFeed creates a new Feed with the current generation time.
func NewFeed(title, description, homeURL string) *Feed {
	return &Feed{
		Generated:   time.Now().UTC(),
		Title:       title,
		Description: description,
		HomeURL:     homeURL,
		Entries:     []Entry{},
	}
}

// AddEntry adds an entry to the feed.
func (f *Feed) AddEntry(e Entry) {
	if e.ID == "" {
		e.ID = GenerateID(e.URL, e.Date)
	}
	f.Entries = append(f.Entries, e)
}

// SortByDate sorts entries by date, newest first.
func (f *Feed) SortByDate() {
	sort.Slice(f.Entries, func(i, j int) bool {
		return f.Entries[i].Date.After(f.Entries[j].Date)
	})
}

// Deduplicate removes duplicate entries based on URL.
// When duplicates are found, it merges discussions and prefers priority entries.
func (f *Feed) Deduplicate() {
	seen := make(map[string]int) // URL -> index in unique slice
	var unique []Entry
	for _, e := range f.Entries {
		normalizedURL := strings.ToLower(strings.TrimRight(e.URL, "/"))
		if idx, exists := seen[normalizedURL]; exists {
			// Merge discussions from duplicate into existing entry
			if len(e.Discussions) > 0 {
				unique[idx].Discussions = mergeDiscussions(unique[idx].Discussions, e.Discussions)
			}
			// If duplicate is a priority entry, upgrade the existing entry
			if e.IsPriority && !unique[idx].IsPriority {
				unique[idx].IsPriority = true
				unique[idx].PriorityRank = e.PriorityRank
			}
		} else {
			seen[normalizedURL] = len(unique)
			unique = append(unique, e)
		}
	}
	f.Entries = unique
}

// mergeDiscussions combines two discussion slices, avoiding duplicates by URL.
func mergeDiscussions(existing, incoming []Discussion) []Discussion {
	seen := make(map[string]bool)
	for _, d := range existing {
		seen[d.URL] = true
	}
	result := existing
	for _, d := range incoming {
		if !seen[d.URL] {
			seen[d.URL] = true
			result = append(result, d)
		}
	}
	return result
}

// FilterByTags returns entries that match any of the given tags.
func (f *Feed) FilterByTags(tags []string) []Entry {
	if len(tags) == 0 {
		return f.Entries
	}
	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[strings.ToLower(t)] = true
	}
	var filtered []Entry
	for _, e := range f.Entries {
		for _, et := range e.Tags {
			if tagSet[strings.ToLower(et)] {
				filtered = append(filtered, e)
				break
			}
		}
	}
	return filtered
}

// WriteJSON writes the feed to a JSON file.
func (f *Feed) WriteJSON(filename string) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// ReadJSON reads a feed from a JSON file.
func ReadJSON(filename string) (*Feed, error) {
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

// ToJSONFeed converts the internal Feed to a JSON Feed 1.1 format.
func (f *Feed) ToJSONFeed() *jsonfeed.Feed {
	jf := jsonfeed.NewFeed(f.Title)
	jf.HomePageURL = f.HomeURL
	jf.Description = f.Description

	for _, e := range f.Entries {
		item := jsonfeed.Item{
			ID:              e.ID,
			URL:             e.URL,
			Title:           e.Title,
			Summary:         e.Summary,
			ContentHTML:     e.Content,
			Image:           e.Image,
			DatePublished:   e.Date.Format(time.RFC3339),
			Tags:            e.Tags,
			SignalFeedTitle: e.Feed.Title,
			SignalFeedURL:   e.Feed.URL,
			SignalPriority:  e.IsPriority,
			SignalRank:      e.PriorityRank,
		}

		if e.Author != "" {
			item.Authors = []jsonfeed.Author{{Name: e.Author}}
		}

		// Copy discussions
		for _, d := range e.Discussions {
			item.SignalDiscussions = append(item.SignalDiscussions, jsonfeed.SignalDiscussion{
				Platform: d.Platform,
				URL:      d.URL,
				ID:       d.ID,
				Score:    d.Score,
				Comments: d.Comments,
			})
		}

		// Copy source metadata
		if e.Source != nil {
			item.SignalSource = &jsonfeed.SignalSource{
				Platform: e.Source.Platform,
				Author:   e.Source.Author,
				PostID:   e.Source.PostID,
			}
		}

		jf.AddItem(item)
	}

	return jf
}

// WriteJSONFeed writes the feed in JSON Feed 1.1 format.
func (f *Feed) WriteJSONFeed(filename string) error {
	return f.ToJSONFeed().WriteFile(filename)
}
