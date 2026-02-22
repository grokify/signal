# Signal

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

Signal is a Go-based Planet-style blog aggregator that outputs [JSON Feed 1.1](https://jsonfeed.org/version/1.1) files. It's designed to be used in CI/CD pipelines (like GitHub Actions) to automatically aggregate RSS/Atom feeds and generate static JSON files that any frontend can consume.

## Features

- 📄 **JSON Feed 1.1 Output** - Standard format with Signal extensions for feed metadata
- 📂 **OPML in JSON** - Maintain feed lists in JSON while preserving OPML semantics
- 📅 **Monthly Archives** - Split output into monthly files to avoid ever-growing files
- 🔄 **Merge Mode** - Preserves historical entries even after they fall off source feeds
- ⭐ **Priority Links** - Hand-curated links that appear at the top of feeds
- ⚛️ **Atom Generation** - Optional Atom feed output for RSS readers
- 🏷️ **Tag Filtering** - Filter entries by tags/keywords
- ⚡ **Concurrent Fetching** - Fast parallel feed fetching
- 🤖 **GitHub Actions Ready** - Built for automated updates in CI/CD

## Installation

```bash
go install github.com/grokify/signal/cmd/signal@latest
```

## Quick Start

```bash
# Initialize a new project with sample files
signal init

# Aggregate feeds and generate JSON Feed output
signal aggregate

# With all options
signal aggregate \
  --opml feeds.json \
  --priority priority.json \
  --output-dir data \
  --output feeds.json \
  --monthly \
  --latest-months 3 \
  --atom atom.xml \
  --title "My Feed" \
  -v
```

## Configuration

### Feed List (feeds.json)

Signal uses OPML represented in JSON format. This allows you to maintain your feed list in a structured, version-controllable format:

```json
{
  "version": "2.0",
  "title": "My Feed Collection",
  "outlines": [
    {
      "text": "Technology",
      "title": "Technology",
      "outlines": [
        {
          "text": "Go Blog",
          "title": "Go Blog",
          "type": "rss",
          "xmlUrl": "https://go.dev/blog/feed.atom",
          "htmlUrl": "https://go.dev/blog",
          "categories": ["Go", "Programming"]
        },
        {
          "text": "fast.ai",
          "title": "fast.ai",
          "type": "rss",
          "xmlUrl": "https://www.fast.ai/index.xml",
          "htmlUrl": "https://www.fast.ai",
          "categories": ["AI", "Machine Learning"]
        }
      ]
    }
  ]
}
```

### Priority Links (priority.json)

Hand-curated links that always appear at the top of feeds:

```json
{
  "title": "Curated Links",
  "description": "Hand-picked priority content",
  "updated": "2026-02-16T00:00:00Z",
  "links": [
    {
      "title": "Important Article",
      "url": "https://example.com/article",
      "author": "Author Name",
      "date": "2026-02-16T00:00:00Z",
      "tags": ["Featured"],
      "summary": "A hand-picked important article.",
      "rank": 1
    }
  ]
}
```

## Output Format

All output uses the [JSON Feed 1.1](https://jsonfeed.org/version/1.1) specification with Signal extensions (prefixed with `_signal_`).

### Main Feed (data/feeds.json)

```json
{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "My Feed",
  "home_page_url": "https://example.com",
  "_signal_generated": "2026-02-16T12:00:00Z",
  "items": [
    {
      "id": "abc123",
      "title": "Article Title",
      "url": "https://example.com/article",
      "date_published": "2026-02-16T10:00:00Z",
      "authors": [{"name": "Author Name"}],
      "tags": ["Go", "Programming"],
      "summary": "Article summary...",
      "content_html": "<p>Full content...</p>",
      "_signal_feed_title": "Source Blog",
      "_signal_feed_url": "https://example.com",
      "_signal_priority": false
    }
  ]
}
```

### Monthly Files

When using `--monthly`, entries are split by publication month:

```
data/
├── feeds.json           # Latest N months combined
├── feeds-2026-02.json   # February 2026 entries
├── feeds-2026-01.json   # January 2026 entries
├── feeds-2025-12.json   # December 2025 entries
├── index.json           # Index of all monthly files
└── atom.xml             # Atom feed (optional)
```

Each monthly file includes `_signal_period` in the feed metadata:

```json
{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "My Feed",
  "_signal_generated": "2026-02-16T12:00:00Z",
  "_signal_period": "2026-02",
  "items": [...]
}
```

### Index File (data/index.json)

```json
{
  "generated": "2026-02-16T12:00:00Z",
  "title": "My Feed",
  "files": [
    {"month": "2026-02", "filename": "feeds-2026-02.json", "count": 15},
    {"month": "2026-01", "filename": "feeds-2026-01.json", "count": 23},
    {"month": "2025-12", "filename": "feeds-2025-12.json", "count": 18}
  ]
}
```

## CLI Reference

```
signal aggregate [flags]

Flags:
  -o, --opml string           OPML file in JSON format (default "feeds.json")
  -p, --priority string       Priority links file (JSON)
  -d, --output-dir string     Output directory (default "data")
  -f, --output string         Output filename (default "feeds.json")
      --atom string           Generate Atom feed file
      --monthly               Split into monthly files
      --monthly-prefix string Prefix for monthly files (default "feeds")
      --latest-months int     Months in latest feed (default 3)
      --merge                 Merge with existing files (default true)
      --max-entries int       Max entries per feed (default 50)
      --max-age int           Max entry age in days (0 = unlimited)
      --tags strings          Filter by tags
      --title string          Feed title (default "Signal Feed")
      --url string            Feed URL for Atom output
      --concurrency int       Concurrent fetches (default 10)
  -v, --verbose               Verbose output

API Generation Flags:
      --api-version string    Generate agent-friendly API (e.g., "v1")
      --planet-name string    Planet name for API metadata
      --planet-description string  Planet description
      --planet-url string     Planet home URL
      --owner-name string     Planet owner name
      --owner-url string      Planet owner URL
      --generate-all          Generate feeds/all.json (can be large)
      --generate-schema       Generate schema.json (default true)
      --generate-agents-md    Generate AGENTS.md (default true)
```

## Agent-Friendly API

Signal can generate a structured, file-based API designed for both AI agents and human developers. Enable it with `--api-version v1`:

```bash
signal aggregate --api-version v1 --planet-name "My Planet" --planet-url "https://example.com"
```

### API Structure

```
data/v1/
├── AGENTS.md              # AI agent instructions
├── schema.json            # JSON Schema for validation
├── meta/
│   ├── about.json         # Planet metadata
│   ├── sources.json       # All feed sources with counts
│   └── stats.json         # Aggregate statistics
├── feeds/
│   └── latest.json        # Latest N months (JSON Feed 1.1)
├── by-month/
│   ├── index.json         # List of all months
│   └── 2026-02.json       # Entries for February 2026
├── by-source/
│   ├── index.json         # List of all sources
│   └── go-blog.json       # Entries from Go Blog
└── by-tag/
    ├── index.json         # List of all tags
    └── programming.json   # Entries tagged "programming"
```

### Why Agent-Friendly?

- **Predictable URLs**: `/v1/by-source/{slug}.json` - no API calls needed to discover paths
- **Self-describing**: AGENTS.md and schema.json explain the structure
- **Standard format**: JSON Feed 1.1 with documented extensions
- **Stateless**: Pure static files, no authentication required
- **Discoverable**: Index files list all available resources

## GitHub Actions

Create a separate repository for your site that uses Signal:

```
my-planet-site/
├── feeds.json              # Your feed list
├── priority.json           # Curated links (optional)
├── frontend/               # Your frontend (React, Vue, etc.)
├── data/                   # Generated output
└── .github/workflows/
    └── update.yml
```

### Workflow Example

```yaml
name: Update Feeds

on:
  schedule:
    - cron: '0 * * * *'  # Every hour
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install Signal
        run: go install github.com/grokify/signal/cmd/signal@latest

      - name: Update feeds
        run: |
          signal aggregate \
            --monthly \
            --title "My Planet" \
            --atom atom.xml \
            -v

      - name: Commit changes
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add data/
          git diff --staged --quiet || git commit -m "Update feeds"
          git push
```

## Building a Frontend

Signal outputs standard JSON Feed that any frontend can consume. See [planet-ai](https://github.com/grokify/planet-ai) for a complete React example.

### Fetching Data (JavaScript)

```javascript
// Fetch latest entries
const feed = await fetch('/data/feeds.json').then(r => r.json());

// Access entries
feed.items.forEach(item => {
  console.log(item.title, item.url, item.date_published);
  console.log('Source:', item._signal_feed_title);
});

// Load monthly archives
const index = await fetch('/data/index.json').then(r => r.json());
for (const file of index.files) {
  const monthly = await fetch(`/data/${file.filename}`).then(r => r.json());
  console.log(`${file.month}: ${monthly.items.length} entries`);
}
```

### React Example

```jsx
function FeedList() {
  const [items, setItems] = useState([]);

  useEffect(() => {
    fetch('/data/feeds.json')
      .then(r => r.json())
      .then(feed => setItems(feed.items));
  }, []);

  return (
    <div>
      {items.map(item => (
        <article key={item.id}>
          <h2><a href={item.url}>{item.title}</a></h2>
          <p>{item.summary}</p>
          <small>{item._signal_feed_title}</small>
        </article>
      ))}
    </div>
  );
}
```

## Architecture

```
┌─────────────────┐     ┌─────────────────┐
│   feeds.json    │     │  priority.json  │
│  (OPML in JSON) │     │ (curated links) │
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     ▼
              ┌──────────────┐
              │  Signal CLI  │
              └──────┬───────┘
                     │
      ┌──────────────┼──────────────┐
      ▼              ▼              ▼
┌───────────┐  ┌───────────┐  ┌───────────┐
│feeds.json │  │ monthly/  │  │ atom.xml  │
│ (latest)  │  │  *.json   │  │           │
└───────────┘  └───────────┘  └───────────┘
      │              │              │
      └──────────────┴──────────────┘
                     │
                     ▼
           ┌─────────────────┐
           │  Your Frontend  │
           │ (React, Vue, …) │
           └─────────────────┘
```

## Packages

| Package | Description |
|---------|-------------|
| `cmd/signal` | CLI application |
| `aggregator` | Fetches and parses RSS/Atom feeds |
| `api` | Agent-friendly API structure generation |
| `atom` | Generates Atom feed output |
| `entry` | Internal entry types and JSON Feed conversion |
| `jsonfeed` | JSON Feed 1.1 specification types |
| `monthly` | Monthly file splitting, merging, and indexing |
| `opml` | OPML in JSON format |
| `priority` | Hand-curated priority links |

## License

MIT License

 [go-ci-svg]: https://github.com/grokify/signal/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/signal/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/signal/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/signal/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/grokify/signal/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/grokify/signal/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/signal
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/signal
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/signal
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/signal
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/signal/blob/main/LICENSE
