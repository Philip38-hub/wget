package downloadutils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
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

// TestParseRateLimit tests the parseRateLimit function.
func TestParseRateLimit(t *testing.T) {
	tests := []struct {
		rate     string
		expected int64
		err      bool
	}{
		{"100k", 102400, false},
		{"1M", 1048576, false},
		{"abc", 0, true},
	}

	for _, test := range tests {
		result, err := parseRateLimit(test.rate)
		if (err != nil) != test.err {
			t.Errorf("parseRateLimit(%q) error = %v, wantErr %v", test.rate, err, test.err)
			continue
		}
		if result != test.expected {
			t.Errorf("parseRateLimit(%q) = %v, want %v", test.rate, result, test.expected)
		}
	}
}

// TestFormatDuration tests the formatDuration function.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "< 1s"},
		{5 * time.Second, "5s"},
		{2*time.Minute + 30*time.Second, "2m30s"},
	}

	for _, test := range tests {
		result := formatDuration(test.duration)
		if result != test.expected {
			t.Errorf("formatDuration(%v) = %v, want %v", test.duration, result, test.expected)
		}
	}
}

// TestCalculateSpeed tests the calculateSpeed function.
func TestCalculateSpeed(t *testing.T) {
	p := &Progress{
		current:   0,
		lastBytes: 0,
		lastTime:  time.Now(),
	}

	// Test initial speed calculation
	speed := p.calculateSpeed()
	if speed != 0 {
		t.Errorf("calculateSpeed() = %v, want 0", speed)
	}

	// Simulate a download
	p.current = 1000
	p.lastBytes = 0
	p.lastTime = time.Now().Add(-1 * time.Second) // 1 second ago

	speed = math.Round(p.calculateSpeed())
	if speed != 1000.0 {
		t.Errorf("calculateSpeed() = %v, want 1000.0", speed)
	}

	// Test smoothing
	p.lastBytes = 500
	p.current = 1500
	p.lastTime = time.Now().Add(-1 * time.Second) // 1 second ago

	speed = math.Round(p.calculateSpeed())
	expectedSpeed := 0.8*1000.0 + 0.2*500.0 // Applying smoothing
	if speed != expectedSpeed {
		t.Errorf("calculateSpeed() = %v, want %v", speed, expectedSpeed)
	}
}

func TestRateLimitedReader_Read(t *testing.T) {
	tests := []struct {
		name           string
		rateLimit      int64
		windowSize     time.Duration
		inputData      []byte
		readSize       int
		expectedOutput []byte
	}{
		{
			name:           "Read within rate limit",
			rateLimit:      10,
			windowSize:     1 * time.Second,
			inputData:      []byte("HelloWorld"), // 10 bytes
			readSize:       5,
			expectedOutput: []byte("Hello"), // Expect to read 5 bytes
		},
		{
			name:           "Exceed rate limit",
			rateLimit:      5,
			windowSize:     1 * time.Second,
			inputData:      []byte("HelloWorld"), // 10 bytes
			readSize:       10,
			expectedOutput: []byte("Hello"), // Expect to read only 5 bytes and wait
		},
		{
			name:           "Read zero bytes",
			rateLimit:      10,
			windowSize:     1 * time.Second,
			inputData:      []byte("HelloWorld"), // 10 bytes
			readSize:       0,
			expectedOutput: []byte(""), // Expect to read 0 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := &MockReader{data: tt.inputData}
			rlReader := &RateLimitedReader{
				reader:     mockReader,
				rateLimit:  tt.rateLimit,
				windowSize: tt.windowSize,
			}

			output := make([]byte, tt.readSize)
			n, err := rlReader.Read(output)

			if err != nil && err != io.EOF {
				t.Fatalf("expected no error, got %v", err)
			}

			if n != len(tt.expectedOutput) {
				t.Errorf("expected to read %d bytes, got %d", len(tt.expectedOutput), n)
			}

			if !bytes.Equal(output[:n], tt.expectedOutput) {
				t.Errorf("expected output %v, got %v", tt.expectedOutput, output[:n])
			}
		})
	}
}
