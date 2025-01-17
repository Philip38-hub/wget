package downloadutils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"wget/models"
)

// parseRateLimit parses rate limit string (e.g., "100k", "1M") into bytes per second
func parseRateLimit(rate string) (int64, error) {
	if rate == "" {
		return 0, nil
	}

	rate = strings.ToLower(rate)
	var multiplier int64 = 1

	// Get the unit (k, m, g) and multiply accordingly
	unit := rate[len(rate)-1:]
	switch unit {
	case "k":
		multiplier = 1024
		rate = rate[:len(rate)-1]
	case "m":
		multiplier = 1024 * 1024
		rate = rate[:len(rate)-1]
	case "g":
		multiplier = 1024 * 1024 * 1024
		rate = rate[:len(rate)-1]
	}

	// Parse the numeric part
	value, err := strconv.ParseInt(rate, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit format: %v", err)
	}

	return value * multiplier, nil
}

// DownloadFile downloads a file from the given URL and saves it to the specified path
func DownloadFile(url, outputPath string, rateLimit string) error {
	return downloadFileWithProgress(url, outputPath, rateLimit, true, os.Stdout)
}

// DownloadFileSilent downloads a file without progress output (for concurrent downloads)
func DownloadFileSilent(url, outputPath string, rateLimit string) error {
	return downloadFileWithProgress(url, outputPath, rateLimit, false, os.Stdout)
}

// DownloadFileBackground downloads a file and writes progress to a log file
func DownloadFileBackground(url, outputPath string, rateLimit string, logFile *os.File) error {
	return downloadFileWithProgress(url, outputPath, rateLimit, true, logFile)
}

// downloadFileWithProgress is the internal download function that can toggle progress display
func downloadFileWithProgress(url, outputPath string, rateLimit string, showProgress bool, output io.Writer) error {
	if showProgress {
		startTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(output, "start at %s\n", startTime)
	}

	// Create HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if showProgress {
		fmt.Fprintf(output, "sending request, awaiting response... status %s\n", resp.Status)
	}

	// Get file size
	size := resp.ContentLength
	if showProgress {
		fmt.Fprintf(output, "content size: %d [~%.2fMB]\n", size, float64(size)/(1024*1024))
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if showProgress {
		fmt.Fprintf(output, "saving file to: ./%s\n", filepath.Base(outputPath))
	}

	// Create the file
	out, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Set up rate limiting if specified
	var reader io.Reader = resp.Body
	if rateLimit != "" {
		rateLimitBytes, err := parseRateLimit(rateLimit)
		if err != nil {
			return fmt.Errorf("failed to parse rate limit: %v", err)
		}
		if rateLimitBytes > 0 {
			reader = models.NewRateLimitedReader(resp.Body, rateLimitBytes)
		}
	}

	// Initialize progress tracking if needed
	var progress *models.Progress
	if showProgress && output == os.Stdout {
		progress = models.NewProgress(size)
		progress.Start()
		reader = io.TeeReader(reader, progress)
	}

	// Copy the response body to the file
	_, err = io.Copy(out, reader)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	if showProgress {
		if progress != nil {
			progress.Stop()
		}
		fmt.Fprintf(output, "Downloaded [%s]\n", url)
		endTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(output, "finished at %s\n", endTime)
	}

	return nil
}
