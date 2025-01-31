package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"wget/downloadutils"
	"wget/flagutils"
	"wget/mirrorutils"
)

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting home directory: %v", err)
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}

func main() {
	// Parse command line flags
	options, err := flagutils.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle background download (-B flag)
	if options.Background {
		if len(options.URLs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: -B flag requires exactly one URL\n")
			os.Exit(1)
		}

		// Create log file
		logFile, err := os.OpenFile("wget-log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating log file: %v\n", err)
			os.Exit(1)
		}
		defer logFile.Close()

		// Print message and start download in background
		fmt.Print("Output will be written to \"wget-log\"\n")

		// Get filename from URL
		url := options.URLs[0]
		filename := filepath.Base(url)
		if filename == "" || filename == "." || filename == "/" {
			filename = "index.html"
		}

		// Create a WaitGroup to ensure the download starts
		var wg sync.WaitGroup
		wg.Add(1)

		// Start download in background
		go func() {
			defer wg.Done()
			err := downloadutils.DownloadFileBackground(url, filename, options.RateLimit, logFile)
			if err != nil {
				fmt.Fprintf(logFile, "Error: %v\n", err)
			}
		}()

		// Wait a moment for the download to start
		time.Sleep(time.Second)
		wg.Wait()
		return
	}

	// Handle input file mode (-i flag)
	if options.InputFile != "" {
		// Read URLs from file
		file, err := os.Open(options.InputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		// Create downloads directory if it doesn't exist
		downloadsDir := "downloads"
		if options.OutputPath != "" {
			if expanded, err := expandPath(options.OutputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			} else {
				downloadsDir = expanded
			}
		}
		if err := os.MkdirAll(downloadsDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating downloads directory: %v\n", err)
			os.Exit(1)
		}

		// Create concurrent downloader
		concurrentDownloader := downloadutils.NewConcurrentDownloader(5, downloadsDir, options.RateLimit)

		// Read URLs line by line
		var urls []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" && !strings.HasPrefix(url, "#") {
				urls = append(urls, url)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}

		// Download all URLs concurrently
		results := concurrentDownloader.DownloadURLs(urls)

		// Print summary
		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
			}
		}
		fmt.Printf("\nDownload summary: %d/%d files downloaded successfully\n", successCount, len(urls))
		return
	}

	// Handle mirror mode
	if options.Mirror {
		if len(options.URLs) != 1 {
			fmt.Fprintf(os.Stderr, "Error: mirror mode requires exactly one URL\n")
			os.Exit(1)
		}

		// Set output directory
		outputDir := "mirrors"
		if options.OutputPath != "" {
			if expanded, err := expandPath(options.OutputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			} else {
				outputDir = expanded
			}
		}

		// Create mirror options
		mirrorOpts := mirrorutils.NewMirrorOptions(options.URLs[0], outputDir, options.ConvertLinks, options.RejectTypes, options.ExcludePaths)
		if mirrorOpts == nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create mirror options\n")
			os.Exit(1)
		}

		// Start mirroring
		fmt.Printf("Starting mirror of %s\n", options.URLs[0])
		fmt.Printf("Output directory: %s\n", outputDir)

		if err := mirrorOpts.Mirror(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nMirroring complete. You can use a tool like 'live-server' to view the mirrored content.\n")
		return
	}

	// Handle direct file download
	var outputPath string

	// First, expand the output path if provided
	var outputDir string
	if options.OutputPath != "" {
		expanded, err := expandPath(options.OutputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		outputDir = expanded
	} else {
		outputDir = "."
	}

	// Get filename from -O flag or URL
	var filename string
	if options.OutputFile != "" {
		filename = filepath.Base(options.OutputFile) // Only use the base name, not the full path
	} else {
		// Parse URL to get filename
		parsedURL, err := url.Parse(options.URLs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid URL: %v\n", err)
			os.Exit(1)
		}

		// Get filename from URL path
		filename = filepath.Base(parsedURL.Path)
		if filename == "" || filename == "." || filename == "/" {
			filename = "index.html"
		}

		// Clean filename (replace invalid characters)
		filename = strings.Map(func(r rune) rune {
			if strings.ContainsRune(`<>:"/\|?*`, r) {
				return '_'
			}
			return r
		}, filename)
	}

	// Combine directory and filename
	outputPath = filepath.Clean(filepath.Join(outputDir, filename))

	// Download the file
	if err := downloadutils.DownloadFile(options.URLs[0], outputPath, options.RateLimit); err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading file: %v\n", err)
		os.Exit(1)
	}
}