package flagutils

import (
	"testing"
)

func TestParseRateLimit(t *testing.T) {
	tests := []struct {
		rateStr   string
		expected  int64
		expectErr bool
	}{
		{"100", 100, false},
		{"100k", 100 * 1024, false},
		{"100K", 100 * 1024, false},
		{"100m", 100 * 1024 * 1024, false},
		{"100M", 100 * 1024 * 1024, false},
		{"0", 0, true},
		{"-100", 0, true},
		{"1000x", 0, true}, // Invalid suffix
		{"abc", 0, true},   // Invalid number
		{"", 0, true},      // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.rateStr, func(t *testing.T) {
			result, err := ParseRateLimit(tt.rateStr)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

