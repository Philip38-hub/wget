package flagutils

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"wget/models"
)

// ParseFlags parses command line arguments and returns Options
func ParseFlags() (*models.Options, error) {
	opts := &models.Options{}

	// Define flags
	flag.BoolVar(&opts.Background, "B", false, "Go to background after startup")
	flag.StringVar(&opts.OutputFile, "O", "", "Write documents to FILE")
	flag.StringVar(&opts.OutputPath, "P", "", "Save files to PATH")
	flag.StringVar(&opts.RateLimit, "rate-limit", "", "Limit the download speed to rate (e.g., 100k or 1M)")
	flag.StringVar(&opts.InputFile, "i", "", "Read URLs from file")
	flag.BoolVar(&opts.Mirror, "mirror", false, "Mirror website")

	// Mirror-related flags
	var rejectListShort, rejectListLong string
	flag.StringVar(&rejectListShort, "R", "", "Reject file types (comma-separated list)")
	flag.StringVar(&rejectListLong, "reject", "", "Reject file types (comma-separated list)")

	var excludeListShort, excludeListLong string
	flag.StringVar(&excludeListShort, "X", "", "Exclude directories (comma-separated list)")
	flag.StringVar(&excludeListLong, "exclude", "", "Exclude directories (comma-separated list)")

	flag.BoolVar(&opts.ConvertLinks, "convert-links", false, "Convert links for offline viewing")
	flag.BoolVar(&opts.UseDynamic, "dynamic", true, "Enable JavaScript rendering")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] URL\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Parse flags
	flag.Parse()

	// Get URL from remaining arguments
	args := flag.Args()
	if len(args) < 1 && opts.InputFile == "" {
		return nil, fmt.Errorf("no URL specified")
	}
	if len(args) > 0 {
		opts.URL = args[0]
	}

	// Process reject list
	rejectList := rejectListShort
	if rejectListLong != "" {
		rejectList = rejectListLong
	}
	if rejectList != "" {
		opts.RejectTypes = strings.Split(rejectList, ",")
		// Clean up the reject types
		for i := range opts.RejectTypes {
			opts.RejectTypes[i] = strings.TrimSpace(opts.RejectTypes[i])
			// Remove leading dot if present
			opts.RejectTypes[i] = strings.TrimPrefix(opts.RejectTypes[i], ".")
		}
	}

	// Process exclude list
	excludeList := excludeListShort
	if excludeListLong != "" {
		excludeList = excludeListLong
	}
	if excludeList != "" {
		opts.ExcludePaths = strings.Split(excludeList, ",")
		// Clean up the exclude paths
		for i := range opts.ExcludePaths {
			opts.ExcludePaths[i] = strings.TrimSpace(opts.ExcludePaths[i])
			// Remove leading and trailing slashes
			opts.ExcludePaths[i] = strings.Trim(opts.ExcludePaths[i], "/")
		}
	}

	// Validate output path
	if opts.OutputPath != "" {
		absPath, err := filepath.Abs(opts.OutputPath)
		if err != nil {
			return nil, fmt.Errorf("invalid output path: %v", err)
		}
		opts.OutputPath = absPath
	} else {
		// Use current directory if no path specified
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %v", err)
		}
		opts.OutputPath = currentDir
	}

	// Validate input file exists if specified
	if opts.InputFile != "" {
		if _, err := os.Stat(opts.InputFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("input file does not exist: %s", opts.InputFile)
		}
	}

	return opts, nil
}
