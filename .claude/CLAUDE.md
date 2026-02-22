# Signal Development Guidelines

This file provides instructions for AI assistants working on the Signal codebase.

## Project Overview

Signal is a Go-based Planet-style blog aggregator that outputs JSON Feed 1.1 files. It's designed as a "headless planet for AI agents and humans" - generating structured, machine-readable output that both web frontends and AI agents can easily consume.

## Architecture

```
signal/
├── cmd/signal/         # CLI entry point (Cobra-based)
├── aggregator/        # Fetches and parses RSS/Atom feeds
├── api/               # Agent-friendly API structure generation
├── atom/              # Atom feed generation
├── entry/             # Internal entry types
├── jsonfeed/          # JSON Feed 1.1 specification types
├── monthly/           # Monthly file splitting and merging
├── opml/              # OPML in JSON format
├── priority/          # Hand-curated priority links
└── docs/              # PRD, TRD, and specifications
```

## Key Design Decisions

1. **JSON Feed 1.1**: All output follows the JSON Feed 1.1 specification with `_signal_*` extensions
2. **Self-contained packages**: `opml/` and `jsonfeed/` have no internal dependencies for reusability
3. **File-based API**: The `/v1/` structure provides predictable, discoverable paths
4. **Merge mode**: Historical entries are preserved even after they fall off source feeds

## Development Guidelines

### Adding Features

1. Consider how the feature affects the agent-friendly API structure
2. Document any new `_signal_*` extensions
3. Ensure backward compatibility with existing JSON Feed consumers

### Testing

```bash
# Build the CLI
go build ./cmd/signal/

# Run linting
golangci-lint run ./...

# Test the CLI
./signal aggregate --help
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Handle all errors (see root CLAUDE.md)
- Keep packages focused and self-contained

## API Generation

When `--api-version v1` is specified, Signal generates:

```
data/v1/
├── AGENTS.md          # AI agent instructions
├── schema.json        # JSON Schema for validation
├── meta/
│   ├── about.json     # Planet metadata
│   ├── sources.json   # All feed sources
│   └── stats.json     # Aggregate statistics
├── feeds/
│   └── latest.json    # Latest N months
├── by-month/
│   ├── index.json     # Month index
│   └── YYYY-MM.json   # Per-month archives
├── by-source/
│   ├── index.json     # Source index
│   └── {slug}.json    # Per-source feeds
└── by-tag/
    ├── index.json     # Tag index
    └── {slug}.json    # Per-tag feeds
```

## Related Projects

- [signal-ai](https://github.com/grokify/signal-ai) - Example implementation with React frontend
