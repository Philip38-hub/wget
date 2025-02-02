package mirrorutils

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"testing"
)

// Mock ProcessUrl method for testing
func (m *MirrorOptions) processUrl(url string) error {
	// Simulate processing the URL
	return nil
}

func TestMirror(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir() + "/test_mirror"
	defer os.RemoveAll(tempDir) // Clean up after the test

	tests := []struct {
		name           string
		url            string
		outputDir      string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "Successful mirror",
			url:            "http://example.com",
			outputDir:      tempDir,
			expectError:    false,
			expectedOutput: fmt.Sprintf("Starting mirror of %s\nOutput directory: %s\nDownloading: %s\n", "http://example.com", tempDir, "http://example.com"),
		},
		{
			name:           "Invalid output directory",
			url:            "http://example.com",
			outputDir:      "", // Invalid output directory
			expectError:    true,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the MirrorOptions using the new constructor
			m := NewMirrorOptions(tt.url, tt.outputDir, false, nil, nil)

			// Create a pipe to capture stdout
			oldStdout := os.Stdout // Keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the Mirror function
			err := m.Mirror()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read the captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check for expected error
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
			}

			// If no error is expected, check the output
			if !tt.expectError {
				if got := buf.String(); got != tt.expectedOutput {
					t.Errorf("expected output %v, got %v", tt.expectedOutput, got)
				}
			}
		})
	}
}

func TestProcessUrl(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir() + "/test_process_url"
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Set up the MirrorOptions
	m := NewMirrorOptions("http://example.com", tempDir, false, nil, nil)

	// Create a pipe to capture stdout
	oldStdout := os.Stdout // Keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the ProcessUrl function
	err := m.ProcessUrl("http://example.com")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check for expected error
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check the output for expected messages
	expectedOutput := "Downloading: http://example.com\n" // Adjust this based on actual output
	if got := buf.String(); got != expectedOutput {
		t.Errorf("expected output %v, got %v", expectedOutput, got)
	}
}

func TestConvertToLocalPath(t *testing.T) {
	tests := []struct {
		name           string
		inputURL       string
		expectedOutput string
	}{
		{
			name:           "Standard file path",
			inputURL:       "http://example.com/images/photo.jpg",
			expectedOutput: "example.com/images/photo.jpg",
		},
		{
			name:           "Dynamic path",
			inputURL:       "http://example.com/api/v1/users/123",
			expectedOutput: "example.com/pages/api/v1/users/123/index.html",
		},
		{
			name:           "Default handling with no extension",
			inputURL:       "http://example.com/about",
			expectedOutput: "example.com/about/index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the URL
			parsedURL, err := url.Parse(tt.inputURL)
			if err != nil {
				t.Fatalf("failed to parse URL: %v", err)
			}

			// Set up the MirrorOptions
			m := NewMirrorOptions(parsedURL.String(), "/tmp", false, nil, nil)

			// Call the convertToLocalPath function
			got := m.convertToLocalPath(parsedURL)

			// Check the output
			if got != tt.expectedOutput {
				t.Errorf("expected output %v, got %v", tt.expectedOutput, got)
			}
		})
	}
}

func TestConvertLinkPath(t *testing.T) {
	tests := []struct {
		name           string
		baseURL        string
		refURL         string
		expectedOutput string
	}{
		{
			name:           "Absolute URL within same host",
			baseURL:        "http://example.com/folder/",
			refURL:         "http://example.com/folder/file.html",
			expectedOutput: "file.html", // Should return relative path
		},
		{
			name:           "External URL",
			baseURL:        "http://example.com/folder/",
			refURL:         "http://anotherdomain.com/file.html",
			expectedOutput: "http://anotherdomain.com/file.html", // Should return absolute URL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the base and reference URLs
			base, err := url.Parse(tt.baseURL)
			if err != nil {
				t.Fatalf("failed to parse base URL: %v", err)
			}
			ref, err := url.Parse(tt.refURL)
			if err != nil {
				t.Fatalf("failed to parse reference URL: %v", err)
			}

			// Set up the MirrorOptions
			m := NewMirrorOptions(base.String(), "/tmp", false, nil, nil)

			// Call the convertLinkPath function
			got := m.convertLinkPath(base, ref)

			// Check the output
			if got != tt.expectedOutput {
				t.Errorf("expected output %v, got %v", tt.expectedOutput, got)
			}
		})
	}
}

func TestExtractURLsFromCSS(t *testing.T) {
	tests := []struct {
		name     string
		cssInput string
		expected []string
	}{
		{
			name:     "Single URL",
			cssInput: "background: url('http://example.com/image.png');",
			expected: []string{"http://example.com/image.png"},
		},
		{
			name:     "Multiple URLs",
			cssInput: "background: url('http://example.com/image1.png'); background: url('http://example.com/image2.png');",
			expected: []string{"http://example.com/image1.png", "http://example.com/image2.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractURLsFromCSS(tt.cssInput)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
