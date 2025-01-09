package downloadutils

import (
	"io"
	"time"
)

// RateLimitedReader wraps an io.Reader to limit its read speed
type RateLimitedReader struct {
	reader     io.ReadCloser
	rateLimit  int64 // bytes per second
	lastRead   time.Time
	bytesRead  int64
	timeWindow time.Duration
}

// NewRateLimitedReader creates a new rate-limited reader
func NewRateLimitedReader(reader io.ReadCloser, rateLimit int64) *RateLimitedReader {
	return &RateLimitedReader{
		reader:     reader,
		rateLimit:  rateLimit,
		lastRead:   time.Now(),
		timeWindow: time.Second, // Reset counter every second
	}
}

func (r *RateLimitedReader) Read(p []byte) (int, error) {
	now := time.Now()
	
	// Reset counter if we're in a new time window
	if now.Sub(r.lastRead) >= r.timeWindow {
		r.bytesRead = 0
		r.lastRead = now
	}

	// Calculate how many bytes we can read in this request
	remainingQuota := r.rateLimit - r.bytesRead
	if remainingQuota <= 0 {
		// Sleep until the next time window
		time.Sleep(r.timeWindow - now.Sub(r.lastRead))
		r.bytesRead = 0
		r.lastRead = time.Now()
		remainingQuota = r.rateLimit
	}

	// Limit the read size to respect the rate limit
	if int64(len(p)) > remainingQuota {
		p = p[:remainingQuota]
	}

	n, err := r.reader.Read(p)
	r.bytesRead += int64(n)
	return n, err
}

// Close implements io.Closer
func (r *RateLimitedReader) Close() error {
	return r.reader.Close()
}
