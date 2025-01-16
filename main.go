package main

import (
	"fmt"
	"os"
	"path/filepath"
	"wget/flagutils"
	"wget/mirrorutils"
)

func main() {
	// Parse command line flags
	opts, err := flagutils.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle mirror mode
	if opts.Mirror {
		// Set output directory
		outputDir := filepath.Join(".", "mirrors")
		if opts.OutputPath != "" {
			outputDir = opts.OutputPath
		}

		// Create mirror options
		mirrorOpts := mirrorutils.NewMirrorOptions(opts.URL, outputDir, opts.ConvertLinks, opts.RejectTypes, opts.ExcludePaths)
		if mirrorOpts == nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create mirror options\n")
			os.Exit(1)
		}

		// Start mirroring
		fmt.Printf("Starting mirror of %s\n", opts.URL)
		fmt.Printf("Output directory: %s\n", outputDir)

		if err := mirrorOpts.Mirror(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nMirroring complete. You can use a tool like 'live-server' to view the mirrored content.\n")
		return
	}

	fmt.Fprintf(os.Stderr, "Error: No URL specified\n")
	os.Exit(1)
}