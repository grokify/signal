package api

// Version is the current API version.
const Version = "v1"

// Config holds configuration for API generation.
type Config struct {
	// Version is the API version (e.g., "v1")
	Version string

	// OutputDir is the base output directory
	OutputDir string

	// Planet metadata
	PlanetName        string
	PlanetDescription string
	PlanetURL         string

	// Owner metadata
	OwnerName string
	OwnerURL  string

	// Generation options
	GenerateAll      bool // Generate feeds/all.json (can be large)
	GenerateSchema   bool // Generate schema.json
	GenerateAgentsMD bool // Generate AGENTS.md
	LatestMonths     int  // Number of months in feeds/latest.json
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Version:          Version,
		OutputDir:        "data",
		PlanetName:       "Orbit Feed",
		GenerateSchema:   true,
		GenerateAgentsMD: true,
		LatestMonths:     3,
	}
}
