// Package aggregator fetches and aggregates RSS/Atom feeds.
package aggregator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/grokify/signal/entry"
	"github.com/grokify/signal/opml"
	"github.com/mmcdole/gofeed"
)

// Config holds aggregator configuration.
type Config struct {
	// UserAgent for HTTP requests
	UserAgent string
	// Timeout for each feed fetch
	Timeout time.Duration
	// MaxEntries limits the number of entries per feed (0 = unlimited)
	MaxEntries int
	// MaxAge filters out entries older than this duration (0 = no limit)
	MaxAge time.Duration
	// FilterTags only includes entries matching these tags (empty = all)
	FilterTags []string
	// Concurrency controls parallel feed fetching
	Concurrency int
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		UserAgent:   "Signal/1.0 (+https://github.com/grokify/signal)",
		Timeout:     30 * time.Second,
		MaxEntries:  50,
		MaxAge:      0,
		FilterTags:  nil,
		Concurrency: 10,
	}
}

// Aggregator fetches and combines feeds.
type Aggregator struct {
	config Config
	parser *gofeed.Parser
}

// New creates a new Aggregator with the given configuration.
func New(cfg Config) *Aggregator {
	parser := gofeed.NewParser()
	parser.UserAgent = cfg.UserAgent
	return &Aggregator{
		config: cfg,
		parser: parser,
	}
}

// FetchResult holds the result of fetching a single feed.
type FetchResult struct {
	Outline opml.Outline
	Entries []entry.Entry
	Error   error
}

// FetchFeed fetches and parses a single feed.
func (a *Aggregator) FetchFeed(ctx context.Context, outline opml.Outline) FetchResult {
	result := FetchResult{Outline: outline}

	if outline.XMLURL == "" {
		result.Error = fmt.Errorf("no XML URL for feed: %s", outline.Title)
		return result
	}

	ctx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	feed, err := a.parser.ParseURLWithContext(outline.XMLURL, ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse %s: %w", outline.XMLURL, err)
		return result
	}

	feedMeta := entry.FeedMeta{
		Title: feed.Title,
		URL:   feed.Link,
	}
	if feedMeta.Title == "" {
		feedMeta.Title = outline.Title
	}
	if feedMeta.URL == "" {
		feedMeta.URL = outline.HTMLURL
	}
	if feed.Image != nil {
		feedMeta.IconURL = feed.Image.URL
	}

	cutoff := time.Time{}
	if a.config.MaxAge > 0 {
		cutoff = time.Now().Add(-a.config.MaxAge)
	}

	for i, item := range feed.Items {
		if a.config.MaxEntries > 0 && i >= a.config.MaxEntries {
			break
		}

		pubDate := time.Now()
		if item.PublishedParsed != nil {
			pubDate = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			pubDate = *item.UpdatedParsed
		}

		if !cutoff.IsZero() && pubDate.Before(cutoff) {
			continue
		}

		// Combine feed categories with outline categories
		tags := append([]string{}, outline.Categories...)
		tags = append(tags, item.Categories...)

		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}

		summary := item.Description
		content := item.Content
		if summary == "" && content != "" {
			// Use first 500 chars of content as summary
			summary = truncateHTML(content, 500)
		}

		e := entry.Entry{
			ID:      entry.GenerateID(item.Link, pubDate),
			Title:   item.Title,
			URL:     item.Link,
			Author:  author,
			Date:    pubDate,
			Feed:    feedMeta,
			Tags:    uniqueStrings(tags),
			Summary: summary,
			Content: content,
		}
		result.Entries = append(result.Entries, e)
	}

	return result
}

// ProgressFunc is called when a feed fetch completes.
// current is the number of feeds fetched so far, total is the total number.
// name is the feed title, entries is the number of entries fetched (0 if error).
type ProgressFunc func(current, total int, name string, entries int, err error)

// FetchAll fetches all feeds from an OPML and returns combined entries.
func (a *Aggregator) FetchAll(ctx context.Context, o *opml.OPML) (*entry.Feed, []error) {
	return a.FetchAllWithProgress(ctx, o, nil)
}

// FetchAllWithProgress fetches all feeds with progress reporting.
func (a *Aggregator) FetchAllWithProgress(ctx context.Context, o *opml.OPML, progress ProgressFunc) (*entry.Feed, []error) {
	feeds := o.FlattenFeeds()

	results := make(chan FetchResult, len(feeds))
	sem := make(chan struct{}, a.config.Concurrency)

	var wg sync.WaitGroup
	for _, outline := range feeds {
		wg.Add(1)
		go func(out opml.Outline) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results <- a.FetchFeed(ctx, out)
		}(outline)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	feed := entry.NewFeed(o.Title, "", "")
	var errors []error
	completed := 0
	total := len(feeds)

	for result := range results {
		completed++
		if result.Error != nil {
			errors = append(errors, result.Error)
			if progress != nil {
				progress(completed, total, result.Outline.Title, 0, result.Error)
			}
			continue
		}
		for _, e := range result.Entries {
			feed.AddEntry(e)
		}
		if progress != nil {
			progress(completed, total, result.Outline.Title, len(result.Entries), nil)
		}
	}

	feed.Deduplicate()
	feed.SortByDate()

	return feed, errors
}

// truncateHTML truncates HTML content to approximately n characters.
func truncateHTML(s string, n int) string {
	if len(s) <= n {
		return s
	}
	// Simple truncation - a proper implementation would handle HTML tags
	truncated := s[:n]
	if idx := strings.LastIndex(truncated, " "); idx > n/2 {
		truncated = truncated[:idx]
	}
	return truncated + "..."
}

// uniqueStrings returns unique strings, preserving order.
func uniqueStrings(ss []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range ss {
		lower := strings.ToLower(s)
		if !seen[lower] && s != "" {
			seen[lower] = true
			result = append(result, s)
		}
	}
	return result
}
