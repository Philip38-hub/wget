package downloadutils

import (
	"fmt"
	"io"
	// "strings"
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
	// lastPrint    bool
	speeds       []float64 // Slice to hold recent speeds
	maxSamples   int       // Maximum number of samples to keep
	emaSpeed     float64   // Exponential moving average of the speed
	alpha        float64   // Smoothing factor for EMA
	minSpeed     float64   // Minimum speed threshold
}

func NewProgressReader(reader io.Reader, totalSize int64, isLogging bool) *ProgressReader {
	return &ProgressReader{
		reader:       reader,
		totalSize:    totalSize,
		lastUpdate:   time.Now(),
		updatePeriod: 100 * time.Millisecond,
		isLogging:    isLogging,
		speeds:       make([]float64, 0), // Initialize the slice
		maxSamples:   20,                  // Increased to keep the last 20 speeds
		emaSpeed:     0,
		alpha:        0.4,                 // Adjusted for more responsiveness
		minSpeed:     50,                  // Minimum speed threshold in bytes per second
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.currentSize += int64(n)

	// Force a final update if this is the last read
	if err == io.EOF {
		// pr.printProgress(true)
		fmt.Println() // Add newline after progress bar
		return n, err
	}

	// Update progress if enough time has passed
	now := time.Now()
	if now.Sub(pr.lastUpdate) >= pr.updatePeriod {
		// pr.printProgress(false)
		pr.lastUpdate = now
		pr.lastBytes = pr.currentSize
	}

	return n, err
}

// func (pr *ProgressReader) printProgress(final bool) {
// 	// Calculate speed
// 	duration := time.Since(pr.lastUpdate)
// 	bytesPerSec := float64(pr.currentSize-pr.lastBytes) / duration.Seconds()
// 	if final {
// 		// For final update, calculate average speed
// 		bytesPerSec = float64(pr.currentSize) / time.Since(pr.lastUpdate).Seconds()
// 	}

// 	// Update EMA only if current speed exceeds the minimum threshold
// 	if bytesPerSec > pr.minSpeed {
// 		if pr.emaSpeed == 0 {
// 			pr.emaSpeed = bytesPerSec // Initialize EMA with the first speed
// 		} else {
// 			pr.emaSpeed = pr.alpha*bytesPerSec + (1-pr.alpha)*pr.emaSpeed // Update EMA
// 		}
// 	}


// 	// if pr.isLogging {
// 	// 	// Simple format for log files
// 	// 	fmt.Printf("%s of %s (%.1f%%) %.1f KB/s\n",
// 	// 		FormatSize(pr.currentSize),
// 	// 		FormatSize(pr.totalSize),
// 	// 		// percentage,
// 	// 		bytesPerSec/1024,
// 	// 	)
// 	// }
// }