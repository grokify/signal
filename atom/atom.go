// Package atom generates Atom feed output from aggregated entries.
package atom

import (
	"encoding/xml"
	"os"
	"time"

	"github.com/grokify/signal/entry"
)

// Feed represents an Atom feed.
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	XMLNS   string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Link    []Link   `xml:"link"`
	Updated string   `xml:"updated"`
	ID      string   `xml:"id"`
	Author  *Author  `xml:"author,omitempty"`
	Entries []Entry  `xml:"entry"`
}

// Link represents an Atom link element.
type Link struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

// Author represents an Atom author element.
type Author struct {
	Name  string `xml:"name"`
	Email string `xml:"email,omitempty"`
	URI   string `xml:"uri,omitempty"`
}

// Entry represents an Atom entry element.
type Entry struct {
	Title     string   `xml:"title"`
	Link      []Link   `xml:"link"`
	ID        string   `xml:"id"`
	Updated   string   `xml:"updated"`
	Published string   `xml:"published,omitempty"`
	Author    *Author  `xml:"author,omitempty"`
	Summary   *Content `xml:"summary,omitempty"`
	Content   *Content `xml:"content,omitempty"`
	Category  []Category `xml:"category,omitempty"`
}

// Content represents Atom content with type attribute.
type Content struct {
	Type    string `xml:"type,attr,omitempty"`
	Content string `xml:",chardata"`
}

// Category represents an Atom category element.
type Category struct {
	Term string `xml:"term,attr"`
}

// FromFeed converts an entry.Feed to an Atom Feed.
func FromFeed(f *entry.Feed, feedURL string) *Feed {
	atomFeed := &Feed{
		XMLNS:   "http://www.w3.org/2005/Atom",
		Title:   f.Title,
		Updated: f.Generated.Format(time.RFC3339),
		ID:      feedURL,
		Link: []Link{
			{Href: feedURL, Rel: "self", Type: "application/atom+xml"},
		},
	}

	if f.HomeURL != "" {
		atomFeed.Link = append(atomFeed.Link, Link{Href: f.HomeURL, Rel: "alternate", Type: "text/html"})
	}

	for _, e := range f.Entries {
		atomEntry := Entry{
			Title:     e.Title,
			ID:        "urn:signal:" + e.ID,
			Updated:   e.Date.Format(time.RFC3339),
			Published: e.Date.Format(time.RFC3339),
			Link: []Link{
				{Href: e.URL, Rel: "alternate", Type: "text/html"},
			},
		}

		if e.Author != "" {
			atomEntry.Author = &Author{Name: e.Author}
		}

		if e.Summary != "" {
			atomEntry.Summary = &Content{Type: "html", Content: e.Summary}
		}

		if e.Content != "" {
			atomEntry.Content = &Content{Type: "html", Content: e.Content}
		}

		for _, tag := range e.Tags {
			atomEntry.Category = append(atomEntry.Category, Category{Term: tag})
		}

		atomFeed.Entries = append(atomFeed.Entries, atomEntry)
	}

	return atomFeed
}

// WriteFile writes the Atom feed to a file.
func (f *Feed) WriteFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err := file.WriteString(xml.Header); err != nil {
		return err
	}

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	return encoder.Encode(f)
}

// ToXML returns the Atom feed as XML bytes.
func (f *Feed) ToXML() ([]byte, error) {
	return xml.MarshalIndent(f, "", "  ")
}
