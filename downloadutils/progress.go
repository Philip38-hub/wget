package downloadutils

import (
	"fmt"
	"strings"
	"time"
)

// Progress tracks download progress
type Progress struct {
	total     int64
	current   int64
	lastPrint time.Time
	started   time.Time
	width     int
	lastBytes int64    // Track bytes since last update
	lastTime  time.Time // Track time since last update
}

// formatDuration formats duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "< 1s"
	}
	
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// NewProgress creates a new Progress instance
func NewProgress(total int64) *Progress {
	now := time.Now()
	return &Progress{
		total:     total,
		lastPrint: now,
		started:   now,
		lastTime:  now,
		width:     50, // Fixed width for the progress bar
	}
}

// Write implements io.Writer to track progress
func (p *Progress) Write(b []byte) (n int, err error) {
	n = len(b)
	p.current += int64(n)

	// Update progress every 100ms
	if time.Since(p.lastPrint) >= 100*time.Millisecond {
		p.printProgress()
		p.lastPrint = time.Now()
	}

	return n, nil
}

// Start begins progress tracking
func (p *Progress) Start() {
	now := time.Now()
	p.started = now
	p.lastTime = now
	p.printProgress()
}

// Stop ends progress tracking
func (p *Progress) Stop() {
	p.printProgress()
	// fmt.Printf("\nTotal time: %s\n", formatDuration(time.Since(p.started)))
}

// calculateSpeed calculates the current download speed using a moving average
func (p *Progress) calculateSpeed() float64 {
	now := time.Now()
	elapsed := now.Sub(p.lastTime).Seconds()

	if elapsed == 0 {
		return 0 // Avoid division by zero
	}

	bytesDiff := p.current - p.lastBytes
	speed := float64(bytesDiff) / elapsed

	// Use a moving average to smooth speed fluctuations
	const smoothingFactor = 0.8
	if p.lastBytes == 0 {
		return speed
	}
	return smoothingFactor*speed + (1-smoothingFactor)*float64(p.lastBytes)/elapsed
}


// printProgress prints the current progress
func (p *Progress) printProgress() {
	var bar string

	// Calculate speed using moving average
	speed := p.calculateSpeed()
	elapsed := time.Since(p.started)

	// Handle unknown total size
	if p.total <= 0 {
		// Show a moving cursor for unknown size
		pos := (p.current / 1024) % int64(p.width)
		left := strings.Repeat("-", int(pos))
		mid := ">"
		right := strings.Repeat("-", p.width-int(pos)-1)
		bar = left + mid + right

		fmt.Printf("\r[%s] %s @ %s/s Time: %s", 
			bar,
			FormatSize(p.current),
			FormatSize(int64(speed)),
			formatDuration(elapsed),
		)
		return
	}

	// Calculate progress for known size
	percent := float64(p.current) / float64(p.total) * 100
	completed := int(float64(p.width) * float64(p.current) / float64(p.total))
	if completed > p.width {
		completed = p.width
	}

	bar = strings.Repeat("=", completed) + strings.Repeat("-", p.width-completed)

	// Calculate ETA
	var eta string
	if speed > 0 {
		remainingBytes := p.total - p.current
		remainingTime := time.Duration(float64(remainingBytes)/speed) * time.Second
		eta = formatDuration(remainingTime)
	} else {
		eta = "Unknown"
	}

	// Print progress
	fmt.Printf("\r[%s] %.1f%% %s/%s @ %s/s Time: %s ETA: %s",
		bar,
		percent,
		FormatSize(p.current),
		FormatSize(p.total),
		FormatSize(int64(speed)),
		formatDuration(elapsed),
		eta,
	)
}
