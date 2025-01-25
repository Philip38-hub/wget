package downloadutils

import (
	"io"
	"sync"
	"time"
)

// RateLimitedReader wraps an io.Reader to limit its read speed
type RateLimitedReader struct {
	reader     io.ReadCloser
	rateLimit  int64 // bytes per second
	lastRead   time.Time
	bytesRead  int64
	timeWindow time.Duration
	mu         sync.Mutex
}

// NewRateLimitedReader creates a new rate-limited reader
func NewRateLimitedReader(reader io.ReadCloser, rateLimit int64) *RateLimitedReader {
	return &RateLimitedReader{
		reader:     reader,
		rateLimit:  rateLimit,
		lastRead:   time.Now(),
		bytesRead:  0,
		timeWindow: time.Second,
	}
}

func (r *RateLimitedReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRead)

	// Calculate allowed bytes for the elapsed time
	allowedBytes := int64(elapsed.Seconds() * float64(r.rateLimit))
	if allowedBytes > r.rateLimit {
		allowedBytes = r.rateLimit
	}

	// Reset counters if we're in a new time window
	if elapsed >= r.timeWindow {
		r.bytesRead = 0
		r.lastRead = now
	}

	// Calculate remaining quota
	remainingQuota := allowedBytes - r.bytesRead

	if remainingQuota <= 0 {
		// Sleep until more quota is available
		time.Sleep(r.timeWindow - elapsed)
		r.bytesRead = 0
		r.lastRead = time.Now()
		remainingQuota = r.rateLimit
	}

	// Limit the read size to the remaining quota
	readSize := int64(len(p))
	if readSize > remainingQuota {
		readSize = remainingQuota
	}

	n, err := r.reader.Read(p[:readSize])
	r.bytesRead += int64(n)

	// Sleep if read size exceeds instantaneous allowed bytes
	expectedReadDuration := time.Duration(float64(n) / float64(r.rateLimit) * float64(time.Second))
	timeTaken := time.Since(r.lastRead)
	if timeTaken < expectedReadDuration {
		time.Sleep(expectedReadDuration - timeTaken)
	}

	r.lastRead = time.Now()

	return n, err
}

// Close implements io.Closer
func (r *RateLimitedReader) Close() error {
	return r.reader.Close()
}