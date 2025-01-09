package models

// Options holds all command line flags and options
type Options struct {
	// Background download flag
	Background bool
	// Custom output filename
	OutputFile string
	// Custom output directory
	OutputPath string
	// Download speed limit (e.g., "200k", "2M")
	RateLimit string
	// Input file containing URLs
	InputFile string
	// Mirror website
	Mirror bool
	// URL to download (non-flag argument)
	URL string
	// Track if we're writing to a log file
	IsLogging bool
}

// NewOptions creates a new Options instance with default values
func NewOptions() *Options {
	return &Options{
		Background: false,
		OutputPath: ".",
	}
}
