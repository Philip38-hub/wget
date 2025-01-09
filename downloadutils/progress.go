package downloadutils

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type ProgressReader struct {
	reader       io.Reader
	totalSize    int64
	currentSize  int64
	lastUpdate   time.Time
	lastBytes    int64
	updatePeriod time.Duration
	isLogging    bool
	lastPrint    bool
}

func NewProgressReader(reader io.Reader, totalSize int64, isLogging bool) *ProgressReader {
	return &ProgressReader{
		reader:       reader,
		totalSize:    totalSize,
		lastUpdate:   time.Now(),
		updatePeriod: 100 * time.Millisecond,
		isLogging:    isLogging,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.currentSize += int64(n)

	// Force a final update if this is the last read
	if err == io.EOF {
		pr.printProgress(true)
		fmt.Println() // Add newline after progress bar
		return n, err
	}

	// Update progress if enough time has passed
	now := time.Now()
	if now.Sub(pr.lastUpdate) >= pr.updatePeriod {
		pr.printProgress(false)
		pr.lastUpdate = now
		pr.lastBytes = pr.currentSize
	}

	return n, err
}

func (pr *ProgressReader) printProgress(final bool) {
	// Calculate speed
	duration := time.Since(pr.lastUpdate)
	bytesPerSec := float64(pr.currentSize-pr.lastBytes) / duration.Seconds()
	if final {
		// For final update, calculate average speed
		bytesPerSec = float64(pr.currentSize) / time.Since(pr.lastUpdate).Seconds()
	}

	// Calculate percentage
	percentage := float64(pr.currentSize) * 100 / float64(pr.totalSize)

	if pr.isLogging {
		// Simple format for log files
		fmt.Printf("%s of %s (%.1f%%) %.1f KB/s\n",
			FormatSize(pr.currentSize),
			FormatSize(pr.totalSize),
			percentage,
			bytesPerSec/1024,
		)
	} else {
		// Interactive format for terminal
		// Calculate ETA
		var eta string
		if !final && bytesPerSec > 0 {
			remainingBytes := pr.totalSize - pr.currentSize
			remainingTime := time.Duration(float64(remainingBytes)/bytesPerSec) * time.Second
			if remainingTime > 0 {
				eta = remainingTime.Round(time.Second).String()
			}
		}

		// Create progress bar
		const barWidth = 30
		completed := int(float64(barWidth) * float64(pr.currentSize) / float64(pr.totalSize))
		if final {
			completed = barWidth // Ensure full bar on completion
		}
		bar := strings.Repeat("=", completed) + strings.Repeat(" ", barWidth-completed)

		// Format the output with ANSI codes for terminal
		status := fmt.Sprintf("\r[%s] %.1f%% %.1f KB/s",
			bar,
			percentage,
			bytesPerSec/1024,
		)
		if !final {
			status += fmt.Sprintf(" ETA %s", eta)
		}

		// Clear the line and print the status
		fmt.Printf("\033[2K%s", status)
	}
}
