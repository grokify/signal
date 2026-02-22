package monthly

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/grokify/signal/entry"
	"github.com/grokify/signal/jsonfeed"
)

// LoadExistingEntries loads all entries from existing monthly files in a directory.
// This allows merging new entries with historical data.
func LoadExistingEntries(dir, prefix string) ([]entry.Entry, error) {
	var entries []entry.Entry

	pattern := filepath.Join(dir, prefix+"-*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// Skip if not a monthly file (e.g., skip index.json)
		base := filepath.Base(file)
		if !strings.HasPrefix(base, prefix+"-") {
			continue
		}

		jf, err := jsonfeed.ReadFile(file)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		for _, item := range jf.Items {
			e := itemToEntry(item)
			entries = append(entries, e)
		}
	}

	return entries, nil
}

// itemToEntry converts a JSON Feed item back to an internal Entry.
func itemToEntry(item jsonfeed.Item) entry.Entry {
	e := entry.Entry{
		ID:      item.ID,
		URL:     item.URL,
		Title:   item.Title,
		Summary: item.Summary,
		Content: item.ContentHTML,
		Tags:    item.Tags,
		Feed: entry.FeedMeta{
			Title: item.SignalFeedTitle,
			URL:   item.SignalFeedURL,
		},
		IsPriority:   item.SignalPriority,
		PriorityRank: item.SignalRank,
	}

	if len(item.Authors) > 0 {
		e.Author = item.Authors[0].Name
	}

	// Parse date
	if item.DatePublished != "" {
		if t, err := time.Parse(time.RFC3339, item.DatePublished); err == nil {
			e.Date = t
		}
	}

	return e
}

// MergeEntries merges new entries with existing entries, deduplicating by URL.
// New entries take precedence over existing entries with the same URL.
func MergeEntries(existing, new []entry.Entry) []entry.Entry {
	// Build map of existing entries by normalized URL
	byURL := make(map[string]entry.Entry)
	for _, e := range existing {
		key := normalizeURL(e.URL)
		byURL[key] = e
	}

	// Add/update with new entries
	for _, e := range new {
		key := normalizeURL(e.URL)
		byURL[key] = e
	}

	// Convert back to slice
	result := make([]entry.Entry, 0, len(byURL))
	for _, e := range byURL {
		result = append(result, e)
	}

	return result
}

func normalizeURL(u string) string {
	return strings.ToLower(strings.TrimRight(u, "/"))
}
