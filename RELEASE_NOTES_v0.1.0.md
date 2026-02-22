# Signal v0.1.0 Release Notes

**Release Date:** 2026-02-22

Signal is a Go-based Planet-style blog aggregator that outputs JSON Feed 1.1 files. It's designed as a "headless planet for AI agents and humans" - generating structured, machine-readable output that both web frontends and AI agents can easily consume.

## Highlights

- **Planet-style blog aggregator** with JSON Feed 1.1 output designed for AI agents and human readers
- **Agent-friendly API structure** with predictable URLs, schema.json, and AGENTS.md for discoverability
- **Priority links** for hand-curated content with discussion links (HackerNews, Reddit, Lobsters) and image pins

## Features

### JSON Feed 1.1 Output

- Full JSON Feed 1.1 specification support
- Signal extensions for feed metadata (`_signal_feed_title`, `_signal_feed_url`)
- Priority markers (`_signal_priority`, `_signal_rank`)
- Discussion links (`_signal_discussions`) for HackerNews, Reddit, Lobsters
- Source platform metadata (`_signal_source`) for LinkedIn, Twitter, etc.

### Feed Aggregation

- Concurrent RSS/Atom feed fetching with configurable concurrency
- OPML in JSON format for version-controllable feed lists
- Progress reporting during feed fetching
- Tag-based filtering

### Priority Links

- Hand-curated links that appear prominently in feeds
- Support for discussion links (HackerNews, Reddit, Lobsters) with score and comment counts
- Image pins for visual content (LinkedIn posts, articles with hero images)
- Source platform metadata for social media posts

### Monthly Archives

- Split output into monthly files to avoid ever-growing files
- Merge mode preserves historical entries even after they fall off source feeds
- Index file for discovering available months

### Agent-Friendly API

- Structured file-based API with predictable paths (`/v1/by-source/{slug}.json`)
- Auto-generated `schema.json` for validation
- Auto-generated `AGENTS.md` with AI agent instructions
- Organized endpoints: by-month, by-source, by-tag

### Output Formats

- JSON Feed 1.1 (primary output)
- Atom feed for RSS reader compatibility
- Monthly archive files with index

## Installation

```bash
go install github.com/grokify/signal/cmd/signal@latest
```

## Quick Start

```bash
# Aggregate feeds and generate JSON Feed output
signal aggregate \
  --opml feeds.json \
  --priority priority.json \
  --output-dir data \
  --monthly \
  --atom atom.xml \
  -v
```

## Documentation

- [README](https://github.com/grokify/signal/blob/main/README.md) - Feature overview and usage
- [PRD](https://github.com/grokify/signal/blob/main/docs/PRD.md) - Product Requirements Document
- [TRD](https://github.com/grokify/signal/blob/main/docs/TRD.md) - Technical Requirements Document

## Example Implementation

See [planet-ai](https://github.com/grokify/planet-ai) for a complete React frontend example using Signal.

## License

MIT License
