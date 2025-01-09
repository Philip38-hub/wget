package models

import (
	"io"
	"time"
)

// ProgressReader wraps an io.Reader to track download progress
type ProgressReader struct {
	Reader     io.Reader
	Total      int64
	Downloaded int64
	LastUpdate time.Time
	Speed      float64
}

// Read implements io.Reader interface
func (pr *ProgressReader) Read(p []byte) (int, error) {
	return pr.Reader.Read(p)
}
