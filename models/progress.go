package models

import (
	"io"
)

// ProgressReader wraps an io.Reader to track download progress
type ProgressReader struct {
	Reader io.Reader
}
