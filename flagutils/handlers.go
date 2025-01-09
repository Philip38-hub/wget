package flagutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"wget/models"
)

// HandleBackground sets up background download and logging
func HandleBackground(opts *models.Options) (*os.File, error) {
	logFile, err := os.Create("wget-log")
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	// Redirect output to log file
	fmt.Println("Output will be written to \"wget-log\"")
	return logFile, nil
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
		return 0, fmt.Errorf("invalid rate limit format: %s", rateStr)
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
