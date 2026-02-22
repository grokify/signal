// Package api provides the agent-friendly API structure generation.
package api

import (
	"regexp"
	"strings"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9-]`)
	multipleHyphens = regexp.MustCompile(`-+`)
)

// Slugify converts a string to a URL-safe slug.
// Examples:
//   - "fast.ai" → "fastai"
//   - "Peter Steinberger" → "peter-steinberger"
//   - "Steve Yegge" → "steve-yegge"
//   - "Machine Learning" → "machine-learning"
func Slugify(s string) string {
	// Lowercase
	s = strings.ToLower(s)
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric except hyphens
	s = nonAlphanumeric.ReplaceAllString(s, "")
	// Collapse multiple hyphens
	s = multipleHyphens.ReplaceAllString(s, "-")
	// Trim hyphens from ends
	s = strings.Trim(s, "-")
	return s
}
