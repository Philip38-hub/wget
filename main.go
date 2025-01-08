package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProgressReader wraps an io.Reader to track download progress
type ProgressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	lastUpdate time.Time
	speed      float64
}

func NewProgressReader(reader io.Reader, total int64) *ProgressReader {
	return &ProgressReader{
		reader:     reader,
		total:      total,
		lastUpdate: time.Now(),
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.downloaded += int64(n)
		pr.updateProgress()
	}
	return n, err
}

func (pr *ProgressReader) updateProgress() {
	now := time.Now()
	duration := now.Sub(pr.lastUpdate).Seconds()
	if duration > 0.1 { // Update every 100ms
		// Calculate speed in bytes per second
		pr.speed = float64(pr.downloaded) / duration

		// Calculate percentage
		percentage := float64(pr.downloaded) * 100 / float64(pr.total)

		// Create progress bar
		width := 50
		completed := int(float64(width) * float64(pr.downloaded) / float64(pr.total))
		bar := strings.Repeat("=", completed) + strings.Repeat(" ", width-completed)

		// Format sizes
		downloadedSize := formatSize(pr.downloaded)
		totalSize := formatSize(pr.total)
		speedStr := formatSize(int64(pr.speed)) + "/s"

		// Calculate remaining time
		remaining := "0s"
		if pr.speed > 0 {
			remainingSecs := float64(pr.total-pr.downloaded) / pr.speed
			remaining = fmt.Sprintf("%.1fs", remainingSecs)
		}

		// Print progress
		fmt.Printf("\r%s / %s [%s] %.2f%% %s %s", 
			downloadedSize, totalSize, bar, percentage, speedStr, remaining)

		if pr.downloaded == pr.total {
			fmt.Println()
		}
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . <url>")
		os.Exit(1)
	}

	url := os.Args[1]
	if err := downloadFile(url); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func downloadFile(url string) error {
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
	fmt.Printf("content size: %d [~%.2fMB]\n", size, float64(size)/(1024*1024))

	// Get filename from URL
	fileName := filepath.Base(url)
	fmt.Printf("saving file to: ./%s\n", fileName)

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Create progress reader
	progressReader := NewProgressReader(resp.Body, size)

	// Copy the response body to the file with progress tracking
	_, err = io.Copy(out, progressReader)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\nDownloaded [%s]\n", url)
	fmt.Printf("finished at %s\n", endTime)

	return nil
}
