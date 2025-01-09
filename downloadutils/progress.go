package downloadutils

import (
	"fmt"
	"io"
	"strings"
	"time"
	"wget/models"
)

type progressReader struct {
	reader *models.ProgressReader
}

func NewProgressReader(reader io.Reader, total int64) *models.ProgressReader {
	return &models.ProgressReader{
		Reader:     reader,
		Total:      total,
		LastUpdate: time.Now(),
	}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Reader.Read(p)
	if n > 0 {
		pr.reader.Downloaded += int64(n)
		updateProgress(pr.reader)
	}
	return n, err
}

func updateProgress(pr *models.ProgressReader) {
	now := time.Now()
	duration := now.Sub(pr.LastUpdate).Seconds()
	if duration > 0.1 { // Update every 100ms
		// Calculate speed in bytes per second
		pr.Speed = float64(pr.Downloaded) / duration

		// Calculate percentage
		percentage := float64(pr.Downloaded) * 100 / float64(pr.Total)

		// Create progress bar
		width := 50
		completed := int(float64(width) * float64(pr.Downloaded) / float64(pr.Total))
		bar := strings.Repeat("=", completed) + strings.Repeat(" ", width-completed)

		// Format sizes
		downloadedSize := FormatSize(pr.Downloaded)
		totalSize := FormatSize(pr.Total)
		speedStr := FormatSize(int64(pr.Speed)) + "/s"

		// Calculate remaining time
		remaining := "0s"
		if pr.Speed > 0 {
			remainingSecs := float64(pr.Total-pr.Downloaded) / pr.Speed
			remaining = fmt.Sprintf("%.1fs", remainingSecs)
		}

		// Print progress
		fmt.Printf("\r%s / %s [%s] %.2f%% %s %s", 
			downloadedSize, totalSize, bar, percentage, speedStr, remaining)

		if pr.Downloaded == pr.Total {
			fmt.Println()
		}

		pr.LastUpdate = now
	}
}
