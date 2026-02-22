package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/grokify/signal/entry"
)

// SignalVersion is the version of Signal.
var SignalVersion = "1.0.0"

// SignalGenerator returns the Generator metadata for Signal.
func SignalGenerator() Generator {
	return Generator{
		Name:    "Signal",
		Version: SignalVersion,
		URL:     "https://github.com/grokify/signal",
	}
}

// Generate creates the complete API structure from a feed.
func Generate(feed *entry.Feed, sources []SourceInfo, cfg Config) error {
	now := time.Now().UTC()
	baseDir := filepath.Join(cfg.OutputDir, cfg.Version)

	// Create directory structure
	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "meta"),
		filepath.Join(baseDir, "feeds"),
		filepath.Join(baseDir, "by-month"),
		filepath.Join(baseDir, "by-source"),
		filepath.Join(baseDir, "by-tag"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Analyze entries
	analysis := analyzeEntries(feed.Entries, sources)

	// Generate meta files
	if err := generateMetaFiles(baseDir, cfg, analysis, now); err != nil {
		return fmt.Errorf("failed to generate meta files: %w", err)
	}

	// Generate feeds
	if err := generateFeeds(baseDir, feed, cfg, now); err != nil {
		return fmt.Errorf("failed to generate feeds: %w", err)
	}

	// Generate by-month files
	if err := generateByMonth(baseDir, feed, now); err != nil {
		return fmt.Errorf("failed to generate by-month files: %w", err)
	}

	// Generate by-source files
	if err := generateBySource(baseDir, feed, analysis, now); err != nil {
		return fmt.Errorf("failed to generate by-source files: %w", err)
	}

	// Generate by-tag files
	if err := generateByTag(baseDir, feed, analysis, now); err != nil {
		return fmt.Errorf("failed to generate by-tag files: %w", err)
	}

	// Generate schema.json
	if cfg.GenerateSchema {
		if err := generateSchema(baseDir); err != nil {
			return fmt.Errorf("failed to generate schema: %w", err)
		}
	}

	// Generate AGENTS.md
	if cfg.GenerateAgentsMD {
		if err := generateAgentsMD(baseDir, cfg, analysis, now); err != nil {
			return fmt.Errorf("failed to generate AGENTS.md: %w", err)
		}
	}

	return nil
}

// SourceInfo contains information about a feed source from OPML.
type SourceInfo struct {
	Title       string
	Description string
	HTMLURL     string
	FeedURL     string
	Categories  []string
}

// Analysis contains analyzed data from entries.
type Analysis struct {
	TotalEntries    int
	TotalSources    int
	TotalTags       int
	OldestEntry     time.Time
	NewestEntry     time.Time
	EntriesByMonth  map[string]int
	EntriesBySource map[string]*SourceAnalysis
	EntriesByTag    map[string]int
	SourceInfo      map[string]SourceInfo
}

// SourceAnalysis contains analyzed data for a single source.
type SourceAnalysis struct {
	Title       string
	Slug        string
	Count       int
	OldestEntry time.Time
	NewestEntry time.Time
}

func analyzeEntries(entries []entry.Entry, sources []SourceInfo) *Analysis {
	a := &Analysis{
		EntriesByMonth:  make(map[string]int),
		EntriesBySource: make(map[string]*SourceAnalysis),
		EntriesByTag:    make(map[string]int),
		SourceInfo:      make(map[string]SourceInfo),
	}

	// Index source info by title
	for _, s := range sources {
		a.SourceInfo[s.Title] = s
	}

	for _, e := range entries {
		a.TotalEntries++

		// Date range
		if a.OldestEntry.IsZero() || e.Date.Before(a.OldestEntry) {
			a.OldestEntry = e.Date
		}
		if a.NewestEntry.IsZero() || e.Date.After(a.NewestEntry) {
			a.NewestEntry = e.Date
		}

		// By month
		month := e.Date.Format("2006-01")
		a.EntriesByMonth[month]++

		// By source
		sourceTitle := e.Feed.Title
		if sourceTitle == "" {
			sourceTitle = "Unknown"
		}
		if a.EntriesBySource[sourceTitle] == nil {
			a.EntriesBySource[sourceTitle] = &SourceAnalysis{
				Title:       sourceTitle,
				Slug:        Slugify(sourceTitle),
				OldestEntry: e.Date,
				NewestEntry: e.Date,
			}
		}
		sa := a.EntriesBySource[sourceTitle]
		sa.Count++
		if e.Date.Before(sa.OldestEntry) {
			sa.OldestEntry = e.Date
		}
		if e.Date.After(sa.NewestEntry) {
			sa.NewestEntry = e.Date
		}

		// By tag
		for _, tag := range e.Tags {
			a.EntriesByTag[strings.ToLower(tag)]++
		}
	}

	a.TotalSources = len(a.EntriesBySource)
	a.TotalTags = len(a.EntriesByTag)

	return a
}

func generateMetaFiles(baseDir string, cfg Config, analysis *Analysis, now time.Time) error {
	metaDir := filepath.Join(baseDir, "meta")

	// about.json
	about := AboutMeta{
		Name:        cfg.PlanetName,
		Description: cfg.PlanetDescription,
		HomeURL:     cfg.PlanetURL,
		FeedURL:     fmt.Sprintf("%s/data/%s/feeds/latest.json", cfg.PlanetURL, cfg.Version),
		AtomURL:     fmt.Sprintf("%s/atom.xml", cfg.PlanetURL),
		Generated:   now,
		Generator:   SignalGenerator(),
	}
	if cfg.OwnerName != "" {
		about.Owner = &Owner{
			Name: cfg.OwnerName,
			URL:  cfg.OwnerURL,
		}
	}
	if err := writeJSON(filepath.Join(metaDir, "about.json"), about); err != nil {
		return err
	}

	// sources.json
	var sourceEntries []SourceEntry
	for title, sa := range analysis.EntriesBySource {
		se := SourceEntry{
			Slug:        sa.Slug,
			Title:       title,
			EntryCount:  sa.Count,
			LatestEntry: sa.NewestEntry,
			OldestEntry: sa.OldestEntry,
			Path:        fmt.Sprintf("/%s/by-source/%s.json", cfg.Version, sa.Slug),
		}
		if info, ok := analysis.SourceInfo[title]; ok {
			se.Description = info.Description
			se.HTMLURL = info.HTMLURL
			se.FeedURL = info.FeedURL
			se.Categories = info.Categories
		}
		sourceEntries = append(sourceEntries, se)
	}
	sort.Slice(sourceEntries, func(i, j int) bool {
		return sourceEntries[i].EntryCount > sourceEntries[j].EntryCount
	})
	sourcesMeta := SourcesMeta{
		Generated: now,
		Count:     len(sourceEntries),
		Sources:   sourceEntries,
	}
	if err := writeJSON(filepath.Join(metaDir, "sources.json"), sourcesMeta); err != nil {
		return err
	}

	// stats.json
	var monthCounts []MonthCount
	for month, count := range analysis.EntriesByMonth {
		monthCounts = append(monthCounts, MonthCount{Month: month, Count: count})
	}
	sort.Slice(monthCounts, func(i, j int) bool {
		return monthCounts[i].Month > monthCounts[j].Month
	})

	var sourceCounts []SourceCount
	for title, sa := range analysis.EntriesBySource {
		sourceCounts = append(sourceCounts, SourceCount{
			Slug:  sa.Slug,
			Title: title,
			Count: sa.Count,
		})
	}
	sort.Slice(sourceCounts, func(i, j int) bool {
		return sourceCounts[i].Count > sourceCounts[j].Count
	})

	var tagCounts []TagCount
	for tag, count := range analysis.EntriesByTag {
		tagCounts = append(tagCounts, TagCount{
			Tag:   tag,
			Slug:  Slugify(tag),
			Count: count,
		})
	}
	sort.Slice(tagCounts, func(i, j int) bool {
		return tagCounts[i].Count > tagCounts[j].Count
	})
	if len(tagCounts) > 20 {
		tagCounts = tagCounts[:20]
	}

	stats := StatsMeta{
		Generated:    now,
		TotalEntries: analysis.TotalEntries,
		TotalSources: analysis.TotalSources,
		TotalTags:    analysis.TotalTags,
		DateRange: DateRange{
			Oldest: analysis.OldestEntry,
			Newest: analysis.NewestEntry,
		},
		EntriesByMonth:  monthCounts,
		EntriesBySource: sourceCounts,
		TopTags:         tagCounts,
	}
	return writeJSON(filepath.Join(metaDir, "stats.json"), stats)
}

func generateFeeds(baseDir string, feed *entry.Feed, cfg Config, now time.Time) error {
	feedsDir := filepath.Join(baseDir, "feeds")

	// latest.json - use existing ToJSONFeed conversion
	latestFeed := filterLatestMonths(feed, cfg.LatestMonths)
	jf := latestFeed.ToJSONFeed()
	jf.Title = cfg.PlanetName
	return jf.WriteFile(filepath.Join(feedsDir, "latest.json"))
}

func filterLatestMonths(feed *entry.Feed, months int) *entry.Feed {
	if months <= 0 {
		return feed
	}

	// Find cutoff date
	cutoff := time.Now().AddDate(0, -months, 0)

	filtered := &entry.Feed{
		Generated:   feed.Generated,
		Title:       feed.Title,
		Description: feed.Description,
		HomeURL:     feed.HomeURL,
	}

	for _, e := range feed.Entries {
		if e.Date.After(cutoff) {
			filtered.Entries = append(filtered.Entries, e)
		}
	}

	return filtered
}

func generateByMonth(baseDir string, feed *entry.Feed, now time.Time) error {
	byMonthDir := filepath.Join(baseDir, "by-month")

	// Group entries by month
	byMonth := make(map[string][]entry.Entry)
	for _, e := range feed.Entries {
		month := e.Date.Format("2006-01")
		byMonth[month] = append(byMonth[month], e)
	}

	// Generate index
	var monthRefs []MonthRef
	for month, entries := range byMonth {
		monthRefs = append(monthRefs, MonthRef{
			Month: month,
			Count: len(entries),
			Path:  fmt.Sprintf("/v1/by-month/%s.json", month),
		})

		// Generate month file
		monthFeed := &entry.Feed{
			Generated: feed.Generated,
			Title:     feed.Title,
			Entries:   entries,
		}
		jf := monthFeed.ToJSONFeed()
		jf.SignalPeriod = month
		if err := jf.WriteFile(filepath.Join(byMonthDir, month+".json")); err != nil {
			return err
		}
	}

	sort.Slice(monthRefs, func(i, j int) bool {
		return monthRefs[i].Month > monthRefs[j].Month
	})

	index := MonthIndex{
		Generated: now,
		Count:     len(monthRefs),
		Months:    monthRefs,
	}
	return writeJSON(filepath.Join(byMonthDir, "index.json"), index)
}

func generateBySource(baseDir string, feed *entry.Feed, analysis *Analysis, now time.Time) error {
	bySourceDir := filepath.Join(baseDir, "by-source")

	// Group entries by source
	bySource := make(map[string][]entry.Entry)
	for _, e := range feed.Entries {
		title := e.Feed.Title
		if title == "" {
			title = "Unknown"
		}
		bySource[title] = append(bySource[title], e)
	}

	// Generate index
	var sourceRefs []SourceRef
	for title, entries := range bySource {
		slug := Slugify(title)
		sourceRefs = append(sourceRefs, SourceRef{
			Slug:  slug,
			Title: title,
			Count: len(entries),
			Path:  fmt.Sprintf("/v1/by-source/%s.json", slug),
		})

		// Generate source file
		sourceFeed := &entry.Feed{
			Generated: feed.Generated,
			Title:     title,
			Entries:   entries,
		}
		jf := sourceFeed.ToJSONFeed()
		if err := jf.WriteFile(filepath.Join(bySourceDir, slug+".json")); err != nil {
			return err
		}
	}

	sort.Slice(sourceRefs, func(i, j int) bool {
		return sourceRefs[i].Count > sourceRefs[j].Count
	})

	index := SourceIndex{
		Generated: now,
		Count:     len(sourceRefs),
		Sources:   sourceRefs,
	}
	return writeJSON(filepath.Join(bySourceDir, "index.json"), index)
}

func generateByTag(baseDir string, feed *entry.Feed, analysis *Analysis, now time.Time) error {
	byTagDir := filepath.Join(baseDir, "by-tag")

	// Group entries by tag (lowercase)
	byTag := make(map[string][]entry.Entry)
	tagTitles := make(map[string]string) // lowercase -> original case

	for _, e := range feed.Entries {
		for _, tag := range e.Tags {
			lower := strings.ToLower(tag)
			byTag[lower] = append(byTag[lower], e)
			if _, ok := tagTitles[lower]; !ok {
				tagTitles[lower] = tag
			}
		}
	}

	// Generate index
	var tagRefs []TagRef
	for lower, entries := range byTag {
		slug := Slugify(lower)
		tagRefs = append(tagRefs, TagRef{
			Tag:   tagTitles[lower],
			Slug:  slug,
			Count: len(entries),
			Path:  fmt.Sprintf("/v1/by-tag/%s.json", slug),
		})

		// Generate tag file
		tagFeed := &entry.Feed{
			Generated: feed.Generated,
			Title:     fmt.Sprintf("Tag: %s", tagTitles[lower]),
			Entries:   entries,
		}
		jf := tagFeed.ToJSONFeed()
		if err := jf.WriteFile(filepath.Join(byTagDir, slug+".json")); err != nil {
			return err
		}
	}

	sort.Slice(tagRefs, func(i, j int) bool {
		return tagRefs[i].Count > tagRefs[j].Count
	})

	index := TagIndex{
		Generated: now,
		Count:     len(tagRefs),
		Tags:      tagRefs,
	}
	return writeJSON(filepath.Join(byTagDir, "index.json"), index)
}

func generateSchema(baseDir string) error {
	schema := map[string]interface{}{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"title":       "Signal API Schema",
		"description": "JSON Schema for Signal planet API",
		"$defs": map[string]interface{}{
			"entry": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":             map[string]string{"type": "string"},
					"url":            map[string]string{"type": "string", "format": "uri"},
					"title":          map[string]string{"type": "string"},
					"date_published": map[string]string{"type": "string", "format": "date-time"},
					"summary":        map[string]string{"type": "string"},
					"content_html":   map[string]string{"type": "string"},
					"authors": map[string]interface{}{
						"type":  "array",
						"items": map[string]string{"$ref": "#/$defs/author"},
					},
					"tags": map[string]interface{}{
						"type":  "array",
						"items": map[string]string{"type": "string"},
					},
					"_signal_feed_title": map[string]string{"type": "string"},
					"_signal_feed_url":   map[string]string{"type": "string", "format": "uri"},
					"_signal_priority":   map[string]string{"type": "boolean"},
				},
				"required": []string{"id"},
			},
			"author": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]string{"type": "string"},
					"url":  map[string]string{"type": "string", "format": "uri"},
				},
			},
			"feed": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version":          map[string]string{"type": "string"},
					"title":            map[string]string{"type": "string"},
					"home_page_url":    map[string]string{"type": "string", "format": "uri"},
					"_signal_generated": map[string]string{"type": "string", "format": "date-time"},
					"_signal_period":    map[string]string{"type": "string"},
					"items": map[string]interface{}{
						"type":  "array",
						"items": map[string]string{"$ref": "#/$defs/entry"},
					},
				},
				"required": []string{"version", "items"},
			},
		},
	}
	return writeJSON(filepath.Join(baseDir, "schema.json"), schema)
}

func generateAgentsMD(baseDir string, cfg Config, analysis *Analysis, now time.Time) error {
	content := fmt.Sprintf(`# %s - Agent API Reference

## Overview

This is a file-based API for **%s**, generated by [Signal](https://github.com/grokify/signal).

All data is static JSON following the [JSON Feed 1.1](https://jsonfeed.org/version/1.1) specification with Signal extensions.

## Quick Start

| Task | Path |
|------|------|
| Latest entries | ` + "`/v1/feeds/latest.json`" + ` |
| All sources | ` + "`/v1/meta/sources.json`" + ` |
| Statistics | ` + "`/v1/meta/stats.json`" + ` |
| Schema | ` + "`/v1/schema.json`" + ` |
| Entries by source | ` + "`/v1/by-source/{slug}.json`" + ` |
| Entries by month | ` + "`/v1/by-month/{YYYY-MM}.json`" + ` |
| Entries by tag | ` + "`/v1/by-tag/{tag}.json`" + ` |

## Statistics

- **Total Entries**: %d
- **Total Sources**: %d
- **Total Tags**: %d
- **Date Range**: %s to %s

## Available Sources

| Source | Entries | Path |
|--------|---------|------|
`, cfg.PlanetName, cfg.PlanetName, analysis.TotalEntries, analysis.TotalSources, analysis.TotalTags,
		analysis.OldestEntry.Format("2006-01-02"), analysis.NewestEntry.Format("2006-01-02"))

	// Add sources table
	for title, sa := range analysis.EntriesBySource {
		content += fmt.Sprintf("| %s | %d | `/%s/by-source/%s.json` |\n",
			title, sa.Count, cfg.Version, sa.Slug)
	}

	content += `
## Navigation

1. Start with ` + "`/v1/meta/about.json`" + ` for planet metadata
2. Use ` + "`/v1/meta/sources.json`" + ` to list all sources
3. Use ` + "`/v1/meta/stats.json`" + ` for aggregate statistics
4. Use index files (` + "`index.json`" + `) to discover available paths
5. Construct paths directly: ` + "`/v1/by-source/{slug}.json`" + `

## Entry Structure

Each entry in a feed follows JSON Feed 1.1 with Signal extensions:

` + "```json" + `
{
  "id": "abc123",
  "url": "https://example.com/article",
  "title": "Article Title",
  "date_published": "2026-02-16T10:00:00Z",
  "summary": "Article summary...",
  "content_html": "<p>Full content...</p>",
  "authors": [{"name": "Author Name"}],
  "tags": ["AI", "Programming"],
  "_signal_feed_title": "Source Blog",
  "_signal_feed_url": "https://example.com",
  "_signal_priority": false
}
` + "```" + `

## Orbit Extensions

Fields prefixed with ` + "`_signal_`" + ` are Orbit-specific:

| Field | Description |
|-------|-------------|
| ` + "`_orbit_generated`" + ` | When the feed was generated |
| ` + "`_orbit_period`" + ` | Month period for monthly archives (e.g., "2026-02") |
| ` + "`_orbit_feed_title`" + ` | Title of the source feed |
| ` + "`_orbit_feed_url`" + ` | URL of the source feed |
| ` + "`_orbit_priority`" + ` | Whether this is a hand-curated priority entry |

---

`
	content += fmt.Sprintf("Generated: %s\nGenerator: Signal %s\n", now.Format(time.RFC3339), SignalVersion)

	return os.WriteFile(filepath.Join(baseDir, "AGENTS.md"), []byte(content), 0644)
}

func writeJSON(filename string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
