// Package main provides the Signal CLI entry point.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grokify/mogo/fmt/progress"
	"github.com/grokify/signal/aggregator"
	"github.com/grokify/signal/api"
	"github.com/grokify/signal/atom"
	"github.com/grokify/signal/entry"
	"github.com/grokify/signal/monthly"
	"github.com/grokify/signal/opml"
	"github.com/grokify/signal/priority"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "signal",
	Short: "Signal is a Planet-style blog aggregator that outputs JSON",
	Long: `Signal aggregates RSS/Atom feeds and outputs JSON files
that can be consumed by React sites or other frontends.

It reads feeds from an OPML file (in JSON format), fetches entries,
and generates structured JSON output suitable for static site hosting.`,
	Version: version,
}

var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregate feeds and generate output",
	Long:  `Fetch all feeds from the OPML file and generate JSON output.`,
	RunE:  runAggregate,
}

var (
	opmlFile       string
	priorityFile   string
	outputDir      string
	outputFile     string
	atomFile       string
	monthlyOutput  bool
	monthlyPrefix  string
	latestMonths   int
	maxEntries     int
	maxAgeDays     int
	filterTags     []string
	feedTitle      string
	feedURL        string
	concurrency    int
	mergeExisting  bool
	verbose        bool

	// API generation flags
	apiVersion         string
	planetName         string
	planetDescription  string
	planetURL          string
	ownerName          string
	ownerURL           string
	generateAll        bool
	generateSchema     bool
	generateAgentsMD   bool
)

func init() {
	rootCmd.AddCommand(aggregateCmd)
	rootCmd.AddCommand(initCmd)

	aggregateCmd.Flags().StringVarP(&opmlFile, "opml", "o", "feeds.json", "OPML file (JSON format)")
	aggregateCmd.Flags().StringVarP(&priorityFile, "priority", "p", "", "Priority links file (JSON)")
	aggregateCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "data", "Output directory")
	aggregateCmd.Flags().StringVarP(&outputFile, "output", "f", "feeds.json", "Output JSON filename")
	aggregateCmd.Flags().StringVar(&atomFile, "atom", "", "Generate Atom feed file")
	aggregateCmd.Flags().BoolVar(&monthlyOutput, "monthly", false, "Split output into monthly files")
	aggregateCmd.Flags().StringVar(&monthlyPrefix, "monthly-prefix", "feeds", "Prefix for monthly files")
	aggregateCmd.Flags().IntVar(&latestMonths, "latest-months", 3, "Number of months in latest feed (0=all)")
	aggregateCmd.Flags().IntVar(&maxEntries, "max-entries", 50, "Max entries per feed")
	aggregateCmd.Flags().IntVar(&maxAgeDays, "max-age", 0, "Max entry age in days (0=unlimited)")
	aggregateCmd.Flags().StringSliceVar(&filterTags, "tags", nil, "Filter by tags")
	aggregateCmd.Flags().StringVar(&feedTitle, "title", "Signal Feed", "Feed title")
	aggregateCmd.Flags().StringVar(&feedURL, "url", "", "Feed URL for Atom output")
	aggregateCmd.Flags().IntVar(&concurrency, "concurrency", 10, "Concurrent feed fetches")
	aggregateCmd.Flags().BoolVar(&mergeExisting, "merge", true, "Merge with existing monthly files (preserves history)")
	aggregateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// API generation flags
	aggregateCmd.Flags().StringVar(&apiVersion, "api-version", "", "Generate agent-friendly API (e.g., 'v1')")
	aggregateCmd.Flags().StringVar(&planetName, "planet-name", "", "Planet name for API metadata")
	aggregateCmd.Flags().StringVar(&planetDescription, "planet-description", "", "Planet description")
	aggregateCmd.Flags().StringVar(&planetURL, "planet-url", "", "Planet home URL")
	aggregateCmd.Flags().StringVar(&ownerName, "owner-name", "", "Planet owner name")
	aggregateCmd.Flags().StringVar(&ownerURL, "owner-url", "", "Planet owner URL")
	aggregateCmd.Flags().BoolVar(&generateAll, "generate-all", false, "Generate feeds/all.json (can be large)")
	aggregateCmd.Flags().BoolVar(&generateSchema, "generate-schema", true, "Generate schema.json")
	aggregateCmd.Flags().BoolVar(&generateAgentsMD, "generate-agents-md", true, "Generate AGENTS.md")
}

func runAggregate(cmd *cobra.Command, args []string) error {
	// Read OPML
	if verbose {
		fmt.Printf("Reading OPML from %s\n", opmlFile)
	}
	o, err := opml.ReadFile(opmlFile)
	if err != nil {
		return fmt.Errorf("failed to read OPML: %w", err)
	}

	feeds := o.FlattenFeeds()
	if verbose {
		fmt.Printf("Found %d feeds\n", len(feeds))
	}

	// Configure aggregator
	cfg := aggregator.Config{
		UserAgent:   "Signal/1.0 (+https://github.com/grokify/signal)",
		Timeout:     30 * time.Second,
		MaxEntries:  maxEntries,
		Concurrency: concurrency,
		FilterTags:  filterTags,
	}
	if maxAgeDays > 0 {
		cfg.MaxAge = time.Duration(maxAgeDays) * 24 * time.Hour
	}

	// Fetch feeds
	agg := aggregator.New(cfg)
	ctx := context.Background()

	var feed *entry.Feed
	var fetchErrors []error

	if verbose {
		fmt.Println("Fetching feeds...")
		// Use progress bar for verbose mode
		renderer := progress.NewSingleStageRenderer(os.Stdout).
			WithBarWidth(30).
			WithTextWidth(40)

		var allErrors []error
		feed, allErrors = agg.FetchAllWithProgress(ctx, o, func(current, total int, name string, entries int, err error) {
			if err != nil {
				renderer.Update(current, total, fmt.Sprintf("%s (error)", name))
			} else {
				renderer.Update(current, total, fmt.Sprintf("%s (%d entries)", name, entries))
			}
		})
		fetchErrors = allErrors
		renderer.Done("")

		fmt.Printf("Fetched %d entries from %d feeds\n", len(feed.Entries), len(feeds))
		if len(fetchErrors) > 0 {
			fmt.Printf("Encountered %d errors:\n", len(fetchErrors))
			for _, e := range fetchErrors {
				fmt.Printf("  - %v\n", e)
			}
		}
	} else {
		feed, fetchErrors = agg.FetchAll(ctx, o)
	}
	feed.Title = feedTitle
	_ = fetchErrors // errors already printed in verbose mode

	// Add priority links
	if priorityFile != "" {
		if verbose {
			fmt.Printf("Reading priority links from %s\n", priorityFile)
		}
		pLinks, err := priority.ReadFile(priorityFile)
		if err != nil {
			return fmt.Errorf("failed to read priority file: %w", err)
		}
		for _, e := range pLinks.ToEntries() {
			feed.AddEntry(e)
		}
		if verbose {
			fmt.Printf("Added %d priority links\n", len(pLinks.Links))
		}
	}

	// Always deduplicate and sort
	feed.Deduplicate()
	feed.SortByDate()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Merge with existing entries if enabled
	if mergeExisting && monthlyOutput {
		existing, err := monthly.LoadExistingEntries(outputDir, monthlyPrefix)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: could not load existing entries: %v\n", err)
			}
		} else if len(existing) > 0 {
			if verbose {
				fmt.Printf("Loaded %d existing entries from monthly files\n", len(existing))
			}
			merged := monthly.MergeEntries(existing, feed.Entries)
			feed.Entries = merged
			feed.Deduplicate()
			feed.SortByDate()
			if verbose {
				fmt.Printf("After merge: %d total entries\n", len(feed.Entries))
			}
		}
	}

	// Write output
	if monthlyOutput {
		// Write monthly files
		files, err := monthly.WriteMonthlyFiles(feed, outputDir, monthlyPrefix)
		if err != nil {
			return fmt.Errorf("failed to write monthly files: %w", err)
		}
		if verbose {
			fmt.Printf("Wrote %d monthly files\n", len(files))
		}

		// Write index
		index := monthly.GenerateIndex(feed, monthlyPrefix)
		indexPath := filepath.Join(outputDir, "index.json")
		indexData, _ := json.MarshalIndent(index, "", "  ")
		if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
			return fmt.Errorf("failed to write index: %w", err)
		}
		if verbose {
			fmt.Printf("Wrote index to %s\n", indexPath)
		}

		// Write latest feed in JSON Feed format
		if latestMonths > 0 {
			latestFeed := monthly.LatestMonths(feed, latestMonths)
			latestPath := filepath.Join(outputDir, outputFile)
			if err := latestFeed.WriteJSONFeed(latestPath); err != nil {
				return fmt.Errorf("failed to write latest feed: %w", err)
			}
			if verbose {
				fmt.Printf("Wrote latest %d months to %s\n", latestMonths, latestPath)
			}
		}
	} else {
		// Write single file in JSON Feed format
		outputPath := filepath.Join(outputDir, outputFile)
		if err := feed.WriteJSONFeed(outputPath); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		if verbose {
			fmt.Printf("Wrote %d entries to %s\n", len(feed.Entries), outputPath)
		}
	}

	// Generate Atom feed
	if atomFile != "" {
		atomFeed := atom.FromFeed(feed, feedURL)
		atomPath := filepath.Join(outputDir, atomFile)
		if err := atomFeed.WriteFile(atomPath); err != nil {
			return fmt.Errorf("failed to write Atom feed: %w", err)
		}
		if verbose {
			fmt.Printf("Wrote Atom feed to %s\n", atomPath)
		}
	}

	// Generate agent-friendly API structure
	if apiVersion != "" {
		if verbose {
			fmt.Printf("Generating API %s structure...\n", apiVersion)
		}

		// Use feed title as planet name if not specified
		pName := planetName
		if pName == "" {
			pName = feedTitle
		}

		// Convert OPML feeds to SourceInfo
		var sources []api.SourceInfo
		for _, f := range feeds {
			sources = append(sources, api.SourceInfo{
				Title:       f.Title,
				Description: f.Description,
				HTMLURL:     f.HTMLURL,
				FeedURL:     f.XMLURL,
				Categories:  f.Categories,
			})
		}

		cfg := api.Config{
			Version:           apiVersion,
			OutputDir:         outputDir,
			PlanetName:        pName,
			PlanetDescription: planetDescription,
			PlanetURL:         planetURL,
			OwnerName:         ownerName,
			OwnerURL:          ownerURL,
			GenerateAll:       generateAll,
			GenerateSchema:    generateSchema,
			GenerateAgentsMD:  generateAgentsMD,
			LatestMonths:      latestMonths,
		}

		if err := api.Generate(feed, sources, cfg); err != nil {
			return fmt.Errorf("failed to generate API: %w", err)
		}
		if verbose {
			fmt.Printf("Generated API %s structure in %s\n", apiVersion, outputDir)
		}
	}

	fmt.Printf("Generated feed with %d entries\n", len(feed.Entries))
	return nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Signal project",
	Long:  `Create sample configuration files for a new Signal project.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// Create sample OPML
	sampleOPML := &opml.OPML{
		Version:      "2.0",
		Title:        "My Feed Collection",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		OwnerName:    "Your Name",
		Outlines: []opml.Outline{
			{
				Text:  "Technology",
				Title: "Technology",
				Outlines: []opml.Outline{
					{
						Text:       "Go Blog",
						Title:      "Go Blog",
						Type:       "rss",
						XMLURL:     "https://go.dev/blog/feed.atom",
						HTMLURL:    "https://go.dev/blog",
						Categories: []string{"Go", "Programming"},
					},
					{
						Text:       "Hacker News",
						Title:      "Hacker News",
						Type:       "rss",
						XMLURL:     "https://news.ycombinator.com/rss",
						HTMLURL:    "https://news.ycombinator.com",
						Categories: []string{"Tech", "News"},
					},
				},
			},
		},
	}

	if err := sampleOPML.WriteFile("feeds.json"); err != nil {
		return fmt.Errorf("failed to write feeds.json: %w", err)
	}
	fmt.Println("Created feeds.json")

	// Create sample priority links
	samplePriority := &priority.Links{
		Title:       "Curated Links",
		Description: "Hand-picked priority content",
		Updated:     time.Now(),
		Links: []priority.Link{
			{
				Title:   "Example Priority Link",
				URL:     "https://example.com/important-article",
				Author:  "Author Name",
				Date:    time.Now(),
				Tags:    []string{"Featured", "Important"},
				Summary: "This is a hand-curated priority link that will appear at the top.",
				Rank:    1,
			},
		},
	}

	if err := samplePriority.WriteFile("priority.json"); err != nil {
		return fmt.Errorf("failed to write priority.json: %w", err)
	}
	fmt.Println("Created priority.json")

	// Create data directory
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	fmt.Println("Created data/ directory")

	fmt.Println("\nSignal project initialized!")
	fmt.Println("Run 'signal aggregate' to fetch feeds and generate JSON output.")
	return nil
}
