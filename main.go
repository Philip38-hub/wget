package main

import (
	"fmt"
	"os"
	"runtime"
	"wget/downloadutils"
	"wget/flagutils"
)

func main() {
	// Store original stdout
	originalStdout := os.Stdout

	// Parse command line flags
	opts, err := flagutils.ParseFlags()
	if err != nil {
		fmt.Fprintf(originalStdout, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle background download
	if opts.Background {
		logFile, err := flagutils.HandleBackground(opts)
		if err != nil {
			fmt.Fprintf(originalStdout, "Error: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			logFile.Close()
			flagutils.RestoreStdout(originalStdout)
		}()
	}

	// Handle multiple URLs from input file
	if opts.InputFile != "" {
		urls, err := flagutils.ReadURLsFromFile(opts.InputFile)
		if err != nil {
			fmt.Printf("Error reading URLs: %v\n", err)
			os.Exit(1)
		}

		// Use concurrent downloader
		workers := runtime.NumCPU() // Use number of CPUs as worker count
		downloader := downloadutils.NewConcurrentDownloader(workers, opts)
		
		fmt.Printf("Starting download of %d files using %d workers\n", len(urls), workers)
		results := downloader.DownloadURLs(urls)

		// Print summary
		successful := 0
		for _, result := range results {
			if result.Success {
				successful++
			}
		}
		fmt.Printf("\nDownload complete: %d successful, %d failed\n", 
			successful, len(results)-successful)
		
		if successful != len(results) {
			os.Exit(1)
		}
		return
	}

	// Single URL download
	if err := downloadutils.DownloadFile(opts.URL, opts); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}