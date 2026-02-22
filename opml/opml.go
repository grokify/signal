// Package opml provides OPML types represented in JSON for feed list management.
package opml

import (
	"encoding/json"
	"os"
	"time"
)

// OPML represents an OPML document in JSON format.
// This allows feed lists to be maintained in JSON while preserving OPML semantics.
type OPML struct {
	Version   string    `json:"version,omitempty"`
	Title     string    `json:"title,omitempty"`
	DateCreated  time.Time `json:"dateCreated,omitempty"`
	DateModified time.Time `json:"dateModified,omitempty"`
	OwnerName string    `json:"ownerName,omitempty"`
	OwnerEmail string   `json:"ownerEmail,omitempty"`
	Outlines  []Outline `json:"outlines"`
}

// Outline represents an OPML outline element, which can contain feeds or nested outlines.
type Outline struct {
	Text        string    `json:"text,omitempty"`
	Title       string    `json:"title,omitempty"`
	Type        string    `json:"type,omitempty"`        // "rss", "atom", "link", etc.
	XMLURL      string    `json:"xmlUrl,omitempty"`      // Feed URL
	HTMLURL     string    `json:"htmlUrl,omitempty"`     // Website URL
	Description string    `json:"description,omitempty"`
	Language    string    `json:"language,omitempty"`
	Categories  []string  `json:"categories,omitempty"`  // Tags/categories for filtering
	Outlines    []Outline `json:"outlines,omitempty"`    // Nested outlines (for grouping)
}

// ReadFile reads an OPML JSON file and returns the parsed OPML structure.
func ReadFile(filename string) (*OPML, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var opml OPML
	if err := json.Unmarshal(data, &opml); err != nil {
		return nil, err
	}
	return &opml, nil
}

// WriteFile writes an OPML structure to a JSON file.
func (o *OPML) WriteFile(filename string) error {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// FlattenFeeds returns all feed outlines from the OPML, flattening any nested structure.
func (o *OPML) FlattenFeeds() []Outline {
	var feeds []Outline
	var flatten func(outlines []Outline)
	flatten = func(outlines []Outline) {
		for _, outline := range outlines {
			if outline.XMLURL != "" {
				feeds = append(feeds, outline)
			}
			if len(outline.Outlines) > 0 {
				flatten(outline.Outlines)
			}
		}
	}
	flatten(o.Outlines)
	return feeds
}
