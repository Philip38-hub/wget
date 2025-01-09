package flagutils

import (
	"flag"
	"fmt"
	"os"
	"wget/models"
)

// ParseFlags parses command line arguments and returns Options
func ParseFlags() (*models.Options, error) {
	opts := models.NewOptions()

	// Define flags
	flag.BoolVar(&opts.Background, "B", false, "Download in background")
	flag.StringVar(&opts.OutputFile, "O", "", "Save file under this name")
	flag.StringVar(&opts.OutputPath, "P", ".", "Save files in this directory")
	flag.StringVar(&opts.RateLimit, "rate-limit", "", "Limit download speed (e.g., 200k, 2M)")
	flag.StringVar(&opts.InputFile, "i", "", "Read URLs from this file")
	flag.BoolVar(&opts.Mirror, "mirror", false, "Mirror website")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] URL\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Get the URL from remaining arguments
	args := flag.Args()
	if len(args) != 1 && opts.InputFile == "" {
		return nil, fmt.Errorf("exactly one URL required, or use -i flag for multiple URLs")
	}

	if len(args) == 1 {
		opts.URL = args[0]
	}

	// Validate options
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	return opts, nil
}

// validateOptions checks if the provided options are valid
func validateOptions(opts *models.Options) error {
	// Check if output directory exists
	if opts.OutputPath != "." {
		if _, err := os.Stat(opts.OutputPath); os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", opts.OutputPath)
		}
	}

	// Validate rate limit format if provided
	if opts.RateLimit != "" {
		if err := validateRateLimit(opts.RateLimit); err != nil {
			return err
		}
	}

	// Check input file exists if specified
	if opts.InputFile != "" {
		if _, err := os.Stat(opts.InputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", opts.InputFile)
		}
	}

	return nil
}

// validateRateLimit checks if the rate limit format is valid
func validateRateLimit(rate string) error {
	// TODO: Implement rate limit validation
	// Should accept formats like: 100k, 2M, etc.
	return nil
}
