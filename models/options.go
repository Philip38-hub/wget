package models

// Options holds the command line options
type Options struct {
	// URLs to download (non-flag arguments)
	URLs          []string
	// Background download flag
	Background    bool
	// Custom output filename
	OutputFile    string
	// Custom output directory
	OutputPath    string
	// Download speed limit (e.g., "200k", "2M")
	RateLimit     string
	// Input file containing URLs
	InputFile     string
	// Track if we're writing to a log file
	IsLogging     bool
	// Mirror website
	Mirror        bool
	// List of file extensions to reject
	RejectTypes   []string // List of file extensions to reject
	// List of paths to exclude
	ExcludePaths  []string // List of paths to exclude
	// Convert links for offline viewing
	ConvertLinks  bool     // Convert links for offline viewing
	// Enable JavaScript rendering
	UseDynamic    bool     // Enable JavaScript rendering
}

// NewOptions creates a new Options instance with default values
func NewOptions() *Options {
	return &Options{
		Background: false,
		OutputPath: ".",
		URLs:      []string{},
	}
}
