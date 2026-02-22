package api

import (
	"time"
)

// AboutMeta contains metadata about the planet.
type AboutMeta struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	HomeURL     string     `json:"home_url,omitempty"`
	FeedURL     string     `json:"feed_url,omitempty"`
	AtomURL     string     `json:"atom_url,omitempty"`
	Owner       *Owner     `json:"owner,omitempty"`
	Generated   time.Time  `json:"generated"`
	Generator   Generator  `json:"generator"`
}

// Owner contains information about the planet owner.
type Owner struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Generator contains information about the software that generated the output.
type Generator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

// SourcesMeta contains metadata about all feed sources.
type SourcesMeta struct {
	Generated time.Time     `json:"generated"`
	Count     int           `json:"count"`
	Sources   []SourceEntry `json:"sources"`
}

// SourceEntry contains metadata about a single feed source.
type SourceEntry struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	HTMLURL     string    `json:"html_url,omitempty"`
	FeedURL     string    `json:"feed_url,omitempty"`
	Categories  []string  `json:"categories,omitempty"`
	EntryCount  int       `json:"entry_count"`
	LatestEntry time.Time `json:"latest_entry,omitempty"`
	OldestEntry time.Time `json:"oldest_entry,omitempty"`
	Path        string    `json:"path"`
}

// StatsMeta contains aggregate statistics about the planet.
type StatsMeta struct {
	Generated       time.Time     `json:"generated"`
	TotalEntries    int           `json:"total_entries"`
	TotalSources    int           `json:"total_sources"`
	TotalTags       int           `json:"total_tags"`
	DateRange       DateRange     `json:"date_range"`
	EntriesByMonth  []MonthCount  `json:"entries_by_month"`
	EntriesBySource []SourceCount `json:"entries_by_source"`
	TopTags         []TagCount    `json:"top_tags"`
}

// DateRange represents a range of dates.
type DateRange struct {
	Oldest time.Time `json:"oldest"`
	Newest time.Time `json:"newest"`
}

// MonthCount represents entry count for a month.
type MonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// SourceCount represents entry count for a source.
type SourceCount struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Count int    `json:"count"`
}

// TagCount represents entry count for a tag.
type TagCount struct {
	Tag   string `json:"tag"`
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}
