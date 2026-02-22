// Package monthly handles monthly feed file generation and management.
package monthly

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/grokify/signal/entry"
)

// MonthKey returns the month key for a given time (e.g., "2026-02").
func MonthKey(t time.Time) string {
	return t.Format("2006-01")
}

// SplitByMonth splits a feed's entries into monthly buckets.
func SplitByMonth(f *entry.Feed) map[string]*entry.Feed {
	buckets := make(map[string]*entry.Feed)

	for _, e := range f.Entries {
		key := MonthKey(e.Date)
		if buckets[key] == nil {
			buckets[key] = &entry.Feed{
				Generated:   f.Generated,
				Title:       f.Title,
				Description: f.Description,
				HomeURL:     f.HomeURL,
				Entries:     []entry.Entry{},
			}
		}
		buckets[key].Entries = append(buckets[key].Entries, e)
	}

	return buckets
}

// WriteMonthlyFiles writes entries to monthly JSON Feed files.
// Files are named like: prefix-2026-02.json
// Output uses JSON Feed 1.1 format (https://jsonfeed.org/version/1.1)
func WriteMonthlyFiles(f *entry.Feed, outputDir, prefix string) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	buckets := SplitByMonth(f)
	var files []string

	for month, monthFeed := range buckets {
		filename := filepath.Join(outputDir, fmt.Sprintf("%s-%s.json", prefix, month))
		// Convert to JSON Feed format and set the period
		jf := monthFeed.ToJSONFeed()
		jf.SignalPeriod = month
		if err := jf.WriteFile(filename); err != nil {
			return files, fmt.Errorf("failed to write %s: %w", filename, err)
		}
		files = append(files, filename)
	}

	sort.Strings(files)
	return files, nil
}

// Index represents an index of monthly feed files.
type Index struct {
	Generated time.Time `json:"generated"`
	Title     string    `json:"title,omitempty"`
	Files     []FileRef `json:"files"`
}

// FileRef references a monthly file.
type FileRef struct {
	Month    string `json:"month"`
	Filename string `json:"filename"`
	Count    int    `json:"count"`
}

// GenerateIndex creates an index of monthly files.
func GenerateIndex(f *entry.Feed, prefix string) *Index {
	buckets := SplitByMonth(f)

	var files []FileRef
	for month, monthFeed := range buckets {
		files = append(files, FileRef{
			Month:    month,
			Filename: fmt.Sprintf("%s-%s.json", prefix, month),
			Count:    len(monthFeed.Entries),
		})
	}

	// Sort by month, newest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].Month > files[j].Month
	})

	return &Index{
		Generated: time.Now().UTC(),
		Title:     f.Title,
		Files:     files,
	}
}

// LatestMonths returns the most recent N months of entries as a single feed.
func LatestMonths(f *entry.Feed, n int) *entry.Feed {
	buckets := SplitByMonth(f)

	// Get sorted month keys
	var months []string
	for month := range buckets {
		months = append(months, month)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(months)))

	// Limit to N months
	if n > 0 && len(months) > n {
		months = months[:n]
	}

	// Combine entries from selected months
	result := &entry.Feed{
		Generated:   f.Generated,
		Title:       f.Title,
		Description: f.Description,
		HomeURL:     f.HomeURL,
		Entries:     []entry.Entry{},
	}

	monthSet := make(map[string]bool)
	for _, m := range months {
		monthSet[m] = true
	}

	for _, e := range f.Entries {
		if monthSet[MonthKey(e.Date)] {
			result.Entries = append(result.Entries, e)
		}
	}

	result.SortByDate()
	return result
}
