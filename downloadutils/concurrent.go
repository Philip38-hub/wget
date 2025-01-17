package downloadutils

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

// Result represents the result of a download
type Result struct {
	URL        string
	Success    bool
	Error      error
	Size       int64
	OutputPath string
}

// ConcurrentDownloader manages concurrent downloads
type ConcurrentDownloader struct {
	concurrency int
	outputPath  string
	rateLimit   string
}

// NewConcurrentDownloader creates a new ConcurrentDownloader instance
func NewConcurrentDownloader(concurrency int, outputPath string, rateLimit string) *ConcurrentDownloader {
	return &ConcurrentDownloader{
		concurrency: concurrency,
		outputPath:  outputPath,
		rateLimit:   rateLimit,
	}
}

// getContentSizes fetches the content sizes for all URLs
func (d *ConcurrentDownloader) getContentSizes(urls []string) []int64 {
	var sizes []int64
	for _, urlStr := range urls {
		resp, err := http.Head(urlStr)
		if err != nil {
			sizes = append(sizes, 0)
			continue
		}
		defer resp.Body.Close()
		sizes = append(sizes, resp.ContentLength)
	}
	return sizes
}

// DownloadURLs downloads multiple URLs concurrently
func (d *ConcurrentDownloader) DownloadURLs(urls []string) []Result {
	// Get content sizes first
	sizes := d.getContentSizes(urls)
	fmt.Printf("content size: [")
	for i, size := range sizes {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(size)
	}
	fmt.Printf("]\n")

	// Create channels for URLs and results
	urlChan := make(chan string, len(urls))
	results := make(chan Result, len(urls))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < d.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for urlStr := range urlChan {
				// Parse URL to get filename
				parsedURL, err := url.Parse(urlStr)
				if err != nil {
					results <- Result{
						URL:     urlStr,
						Success: false,
						Error:   fmt.Errorf("invalid URL: %v", err),
					}
					continue
				}

				// Get filename from URL
				fileName := filepath.Base(parsedURL.Path)
				if fileName == "" || fileName == "." || fileName == "/" {
					fileName = "index.html"
				}

				// Clean filename
				fileName = strings.Map(func(r rune) rune {
					if strings.ContainsRune(`<>:"/\|?*`, r) {
						return '_'
					}
					return r
				}, fileName)

				// Create output path
				outputPath := filepath.Join(d.outputPath, fileName)

				// Download the file
				err = DownloadFileSilent(urlStr, outputPath, d.rateLimit)
				result := Result{
					URL:        urlStr,
					Success:    err == nil,
					Error:      err,
					OutputPath: outputPath,
				}
				results <- result

				if err == nil {
					fmt.Printf("finished %s\n", fileName)
				}
			}
		}()
	}

	// Send URLs to workers
	for _, urlStr := range urls {
		urlChan <- urlStr
	}
	close(urlChan)

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var resultsList []Result
	var successfulURLs []string
	for result := range results {
		if result.Error != nil {
			fmt.Printf("Error downloading %s: %v\n", result.URL, result.Error)
		} else {
			successfulURLs = append(successfulURLs, result.URL)
		}
		resultsList = append(resultsList, result)
	}

	// Print final summary
	fmt.Printf("\nDownload finished: [%s]\n", strings.Join(successfulURLs, " "))

	return resultsList
}
