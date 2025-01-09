package flagutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"wget/models"
)

// HandleBackground sets up background download and logging
func HandleBackground(opts *models.Options) (*os.File, error) {
	// Create or truncate the log file
	logFile, err := os.OpenFile("wget-log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	// Redirect stdout to the log file
	originalStdout := os.Stdout
	os.Stdout = logFile

	// Print initial message to terminal
	_, err = fmt.Fprintln(originalStdout, "Output will be written to \"wget-log\"")
	if err != nil {
		logFile.Close()
		return nil, err
	}

	// Set logging flag
	opts.IsLogging = true

	return logFile, nil
}

// RestoreStdout restores the original stdout
func RestoreStdout(original *os.File) {
	os.Stdout = original
}

// GetOutputPath returns the full path for saving the file
func GetOutputPath(opts *models.Options, url string) string {
	fileName := opts.OutputFile
	if fileName == "" {
		fileName = filepath.Base(url)
	}
	return filepath.Join(opts.OutputPath, fileName)
}

// ParseRateLimit converts rate limit string to bytes per second
func ParseRateLimit(rateStr string) (int64, error) {
	// Validate format using regex
	validFormat := regexp.MustCompile(`^\d+[kKmM]?$`)
	if !validFormat.MatchString(rateStr) {
		return 0, fmt.Errorf("invalid rate limit format. Must be a number followed by optional k/K/m/M suffix")
	}

	rateStr = strings.ToLower(rateStr)
	var multiplier int64 = 1

	switch {
	case strings.HasSuffix(rateStr, "k"):
		multiplier = 1024
		rateStr = strings.TrimSuffix(rateStr, "k")
	case strings.HasSuffix(rateStr, "m"):
		multiplier = 1024 * 1024
		rateStr = strings.TrimSuffix(rateStr, "m")
	}

	rate, err := strconv.ParseInt(rateStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit number: %s", rateStr)
	}

	if rate <= 0 {
		return 0, fmt.Errorf("rate limit must be greater than 0")
	}

	return rate * multiplier, nil
}

// ReadURLsFromFile reads URLs from the input file
func ReadURLsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" && !strings.HasPrefix(url, "#") {
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
