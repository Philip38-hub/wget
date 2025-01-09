package main

import (
	"fmt"
	"os"
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
		logFile, err := flagutils.HandleBackground()
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
		// TODO: Implement concurrent downloads
		for _, url := range urls {
			if err := downloadutils.DownloadFile(url, opts); err != nil {
				fmt.Printf("Error downloading %s: %v\n", url, err)
			}
		}
		return
	}

	// Single URL download
	if err := downloadutils.DownloadFile(opts.URL, opts); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}