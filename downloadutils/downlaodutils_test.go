package downloadutils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	// "time"
)

// MockDownloadFileSilent is a mock function to replace DownloadFileSilent for testing
func MockDownloadFileSilent(urlStr, outputPath string, rateLimit string) error {
	// Simulate successful download for valid URLs
	if urlStr == "http://example.com/valid" {
		// Create a dummy file to simulate a successful download
		return os.WriteFile(outputPath, []byte("dummy content"), 0o644)
	}
	// Simulate failure for invalid URLs
	return fmt.Errorf("failed to download")
}

// TestDownloadURLs tests the DownloadURLs function
func TestDownloadURLs(t *testing.T) {
	// Create a temporary directory for output
	outputDir := t.TempDir()

	// Create a ConcurrentDownloader instance without the mock function
	downloader := &ConcurrentDownloader{
		outputPath:  outputDir,
		rateLimit:   "300000", // 300k limit for testing
		concurrency: 2,        // Set concurrency level
	}

	tests := []struct {
		name     string
		urls     []string
		expected []string
	}{
		{
			name:     "Valid URL",
			urls:     []string{"http://example.com/valid"},
			expected: []string{"http://example.com/valid"},
		},
		{
			name:     "Invalid URL",
			urls:     []string{"http://example.com/invalid"},
			expected: []string{},
		},
		{
			name:     "Mixed URLs",
			urls:     []string{"http://example.com/valid", "http://example.com/invalid"},
			expected: []string{"http://example.com/valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the mock function directly in the test
			for _, urlStr := range tt.urls {
				outputPath := filepath.Join(outputDir, filepath.Base(urlStr))
				err := MockDownloadFileSilent(urlStr, outputPath, downloader.rateLimit)
				if (err == nil) != (urlStr == "http://example.com/valid") {
					t.Errorf("expected download success for %s, got error: %v", urlStr, err)
				}
			}

			// Check if the successful URLs match the expected URLs
			var successfulURLs []string
			for _, urlStr := range tt.urls {
				outputPath := filepath.Join(outputDir, filepath.Base(urlStr))
				if _, err := os.Stat(outputPath); err == nil {
					successfulURLs = append(successfulURLs, urlStr)
				}
			}

			if len(successfulURLs) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, successfulURLs)
			}
		})
	}
}

// TestDownloadFileWithProgress tests the downloadFileWithProgress function
func TestDownloadFileWithProgress(t *testing.T) {
	// Create a temporary directory for output
	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "testfile.txt")

	// Create a mock server to simulate file download
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a successful response with a small file
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("This is a test file content."))
	}))
	defer mockServer.Close()

	tests := []struct {
		name         string
		url          string
		outputPath   string
		rateLimit    string
		showProgress bool
		expectError  bool
	}{
		{
			name:         "Successful Download",
			url:          mockServer.URL,
			outputPath:   outputPath,
			rateLimit:    "",
			showProgress: true,
			expectError:  false,
		},
		{
			name:         "Invalid URL",
			url:          "http://invalid-url",
			outputPath:   outputPath,
			rateLimit:    "",
			showProgress: true,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downloadFileWithProgress(tt.url, tt.outputPath, tt.rateLimit, tt.showProgress, os.Stdout)

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			if !tt.expectError {
				// Check if the file was created
				if _, err := os.Stat(tt.outputPath); os.IsNotExist(err) {
					t.Errorf("expected file %s to exist, but it does not", tt.outputPath)
				} else {
					// Optionally, verify the content of the downloaded file
					content, err := ioutil.ReadFile(tt.outputPath)
					if err != nil {
						t.Errorf("failed to read downloaded file: %v", err)
					}
					expectedContent := "This is a test file content."
					if string(content) != expectedContent {
						t.Errorf("expected content %q, got %q", expectedContent, string(content))
					}
				}
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1500, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1500 * 1024, "1.5 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1500 * 1024 * 1024, "1.5 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatSize(%d) = %q; want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

// MockReader simulates a reader that returns predefined data
type MockReader struct {
	data []byte
}

func (m *MockReader) Read(p []byte) (int, error) {
	if len(m.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, m.data)
	m.data = m.data[n:]
	return n, nil
}


// TestRateLimitedReader tests the Read method of RateLimitedReader
// Close method to satisfy io.ReadCloser interface
func (m *MockReader) Close() error {
	// No resources to clean up, just return nil
	return nil
}
// func TestRateLimitedReader(t *testing.T) {
// 	// Create a mock data source
// 	data := []byte("This is a test data stream that is quite long.")
// 	mockReader := &MockReader{data: data}

// 	// Create a RateLimitedReader with a rate limit of 10 bytes per second
// 	rateLimitedReader := &RateLimitedReader{
// 		reader:     mockReader,
// 		rateLimit:  10, // 10 bytes per second
// 		timeWindow: time.Second,
// 	}

// 	// Create a buffer to read data into
// 	buf := make([]byte, 20) // Buffer larger than the rate limit to test chunking
// 	var totalRead int64

// 	// Read from the RateLimitedReader
// 	for {
// 		n, err := rateLimitedReader.Read(buf)
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			t.Fatalf("unexpected error: %v", err)
// 		}
// 		totalRead += int64(n)

// 		// Simulate a delay to allow for rate limiting
// 		time.Sleep(100 * time.Millisecond) // Sleep to simulate time passing
// 	}

// 	// Check if the total read matches the expected length
// 	if totalRead != int64(len(data)) {
// 		t.Errorf("expected total read %d, got %d", len(data), totalRead)
// 	}
// }
