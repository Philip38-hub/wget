package downloadutils

import (
	"io"
	"time"
)

// RateLimitedReader wraps an io.Reader to limit its read rate
type RateLimitedReader struct {
	reader     io.ReadCloser
	rateLimit  int64 // bytes per second
	lastRead   time.Time
	bytesRead  int64
	windowSize time.Duration
}

// NewRateLimitedReader creates a new rate-limited reader
func NewRateLimitedReader(reader io.ReadCloser, rateLimit int64) *RateLimitedReader {
	return &RateLimitedReader{
		reader:     reader,
		rateLimit:  rateLimit,
		lastRead:   time.Now(),
		windowSize: time.Second,
	}
}

// Read implements io.Reader interface with rate limiting
func (r *RateLimitedReader) Read(p []byte) (n int, err error) {
	now := time.Now()
	elapsed := now.Sub(r.lastRead)

	// If we've moved to a new window, reset the counter
	if elapsed >= r.windowSize {
		r.bytesRead = 0
		r.lastRead = now
	}

	// Calculate how many bytes we can read in this window
	allowedBytes := r.rateLimit - r.bytesRead
	if allowedBytes <= 0 {
		// We've exceeded our rate limit for this window
		// Sleep until the next window
		time.Sleep(r.windowSize - elapsed)
		r.bytesRead = 0
		r.lastRead = time.Now()
		allowedBytes = r.rateLimit
	}

	// Limit the read size to respect the rate limit
	if int64(len(p)) > allowedBytes {
		p = p[:allowedBytes]
	}

	// Perform the actual read
	n, err = r.reader.Read(p)
	r.bytesRead += int64(n)

	return n, err
}

// Close implements io.Closer interface
func (r *RateLimitedReader) Close() error {
	return r.reader.Close()
}
