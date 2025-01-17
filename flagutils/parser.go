package flagutils

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"wget/models"
)

// ParseFlags parses command line arguments and returns Options
func ParseFlags() (*models.Options, error) {
	// Create a new FlagSet to avoid global state
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts := models.NewOptions()

	// Define flags
	fs.BoolVar(&opts.Background, "B", false, "Go to background after startup")
	fs.StringVar(&opts.OutputFile, "O", "", "Write documents to FILE")
	fs.StringVar(&opts.OutputPath, "P", "", "Save files to PATH")
	fs.StringVar(&opts.RateLimit, "rate-limit", "", "Limit the download speed to rate (e.g., 100k or 1M)")
	fs.StringVar(&opts.InputFile, "i", "", "Read URLs from file")
	fs.BoolVar(&opts.Mirror, "mirror", false, "Mirror website")

	// Mirror-related flags
	var rejectListShort, rejectListLong string
	fs.StringVar(&rejectListShort, "R", "", "Reject file types (comma-separated list)")
	fs.StringVar(&rejectListLong, "reject", "", "Reject file types (comma-separated list)")

	var excludeListShort, excludeListLong string
	fs.StringVar(&excludeListShort, "X", "", "Exclude directories (comma-separated list)")
	fs.StringVar(&excludeListLong, "exclude", "", "Exclude directories (comma-separated list)")

	fs.BoolVar(&opts.ConvertLinks, "convert-links", false, "Convert links for offline viewing")
	fs.BoolVar(&opts.UseDynamic, "dynamic", true, "Enable JavaScript rendering")

	// Custom usage message
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] URL\n\nOptions:\n", os.Args[0])
		fs.PrintDefaults()
	}

	// Parse flags, but skip the program name
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// Get URLs from remaining arguments
	args := fs.Args()
	if len(args) < 1 && opts.InputFile == "" {
		return nil, fmt.Errorf("no URL specified")
	}

	// Store URLs
	opts.URLs = args

	// Process reject lists (combine short and long options)
	rejectTypes := []string{}
	if rejectListShort != "" {
		rejectTypes = append(rejectTypes, strings.Split(rejectListShort, ",")...)
	}
	if rejectListLong != "" {
		rejectTypes = append(rejectTypes, strings.Split(rejectListLong, ",")...)
	}
	for i := range rejectTypes {
		rejectTypes[i] = strings.TrimSpace(rejectTypes[i])
	}
	opts.RejectTypes = rejectTypes

	// Process exclude lists (combine short and long options)
	excludePaths := []string{}
	if excludeListShort != "" {
		excludePaths = append(excludePaths, strings.Split(excludeListShort, ",")...)
	}
	if excludeListLong != "" {
		excludePaths = append(excludePaths, strings.Split(excludeListLong, ",")...)
	}
	for i := range excludePaths {
		excludePaths[i] = strings.TrimSpace(excludePaths[i])
	}
	opts.ExcludePaths = excludePaths

	return opts, nil
}
