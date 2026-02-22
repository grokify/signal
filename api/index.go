package api

import (
	"time"
)

// MonthIndex lists all available monthly archives.
type MonthIndex struct {
	Generated time.Time  `json:"generated"`
	Count     int        `json:"count"`
	Months    []MonthRef `json:"months"`
}

// MonthRef references a monthly archive file.
type MonthRef struct {
	Month string `json:"month"`
	Count int    `json:"count"`
	Path  string `json:"path"`
}

// SourceIndex lists all available source feeds.
type SourceIndex struct {
	Generated time.Time   `json:"generated"`
	Count     int         `json:"count"`
	Sources   []SourceRef `json:"sources"`
}

// SourceRef references a source feed file.
type SourceRef struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Count int    `json:"count"`
	Path  string `json:"path"`
}

// TagIndex lists all available tag feeds.
type TagIndex struct {
	Generated time.Time `json:"generated"`
	Count     int       `json:"count"`
	Tags      []TagRef  `json:"tags"`
}

// TagRef references a tag feed file.
type TagRef struct {
	Tag   string `json:"tag"`
	Slug  string `json:"slug"`
	Count int    `json:"count"`
	Path  string `json:"path"`
}
