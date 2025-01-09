package downloadutils

import (
	"fmt"
	"sync"
	"wget/models"
)

// Result represents the result of a download operation
type Result struct {
	URL     string
	Success bool
	Error   error
}

// ConcurrentDownloader manages concurrent downloads
type ConcurrentDownloader struct {
	workers   int
	opts      *models.Options
	results   chan Result
	waitGroup sync.WaitGroup
}

// NewConcurrentDownloader creates a new concurrent downloader
func NewConcurrentDownloader(workers int, opts *models.Options) *ConcurrentDownloader {
	return &ConcurrentDownloader{
		workers: workers,
		opts:    opts,
		results: make(chan Result, workers),
	}
}

// DownloadURLs downloads multiple URLs concurrently
func (cd *ConcurrentDownloader) DownloadURLs(urls []string) []Result {
	// Create a channel for URLs
	urlChan := make(chan string, len(urls))
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// Start worker pool
	for i := 0; i < cd.workers; i++ {
		cd.waitGroup.Add(1)
		go cd.worker(urlChan)
	}

	// Wait for all downloads to complete in a separate goroutine
	go func() {
		cd.waitGroup.Wait()
		close(cd.results)
	}()

	// Collect results
	var results []Result
	for result := range cd.results {
		if result.Error != nil {
			fmt.Printf("Error downloading %s: %v\n", result.URL, result.Error)
		} else {
			fmt.Printf("Successfully downloaded: %s\n", result.URL)
		}
		results = append(results, result)
	}

	return results
}

// worker processes URLs from the channel
func (cd *ConcurrentDownloader) worker(urls chan string) {
	defer cd.waitGroup.Done()

	for url := range urls {
		// Create a copy of options for this download
		opts := *cd.opts
		
		// Download the file
		err := DownloadFile(url, &opts)
		
		// Send result
		cd.results <- Result{
			URL:     url,
			Success: err == nil,
			Error:   err,
		}
	}
}
