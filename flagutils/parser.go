package flagutils

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	flag.BoolVar(&opts.Mirror, "mirror", false, "Mirror the website")

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
