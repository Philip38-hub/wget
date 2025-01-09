package downloadutils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"wget/flagutils"
	"wget/models"
)

// DownloadFile downloads a file from the given URL and saves it to disk
func DownloadFile(url string, opts *models.Options) error {
	startTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("start at %s\n", startTime)

	// Create HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("sending request, awaiting response... status %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Get file size
	size := resp.ContentLength
	if size <= 0 {
		return fmt.Errorf("invalid or unknown file size: %d", size)
	}
	fmt.Printf("content size: %d [~%.2fMB]\n", size, float64(size)/(1024*1024))

	// Get output path
	fileName := opts.OutputFile
	if fileName == "" {
		fileName = filepath.Base(url)
	}
	fullPath := filepath.Join(opts.OutputPath, fileName)
	fmt.Printf("saving file to: %s\n", fullPath)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the file
	out, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Set up the download chain: response body -> rate limiter -> progress tracker -> file
	reader := resp.Body

	// Apply rate limiting if specified
	if opts.RateLimit != "" {
		rateLimit, err := flagutils.ParseRateLimit(opts.RateLimit)
		if err != nil {
			return fmt.Errorf("invalid rate limit: %v", err)
		}
		reader = NewRateLimitedReader(reader, rateLimit)
		fmt.Printf("Rate limit set to: %s/s\n", FormatSize(rateLimit))
	}

	// Create progress reader
	progressReader := NewProgressReader(reader, size, opts.IsLogging)

	// Copy the response body to the file with progress tracking
	_, err = io.Copy(out, io.TeeReader(progressReader, out))
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\nDownloaded [%s]\n", url)
	fmt.Printf("finished at %s\n", endTime)

	return nil
}
