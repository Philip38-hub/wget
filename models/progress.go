package models

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

// formatSize formats bytes into human readable format
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
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMG"[exp])
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
		width:     40, // Fixed width for the progress bar
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
	fmt.Printf("\nTotal time: %s\n", formatDuration(time.Since(p.started)))
}

// calculateSpeed calculates the current download speed using a moving average
func (p *Progress) calculateSpeed() float64 {
	now := time.Now()
	elapsed := now.Sub(p.lastTime).Seconds()
	
	// Calculate bytes transferred since last update
	bytesDiff := p.current - p.lastBytes
	
	// Calculate speed
	speed := float64(bytesDiff) / elapsed
	
	// Update last values
	p.lastBytes = p.current
	p.lastTime = now
	
	return speed
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
			formatSize(p.current),
			formatSize(int64(speed)),
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
		formatSize(p.current),
		formatSize(p.total),
		formatSize(int64(speed)),
		formatDuration(elapsed),
		eta,
	)
}
