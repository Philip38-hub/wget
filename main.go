package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . <url>")
		os.Exit(1)
	}

	url := os.Args[1]
	if err := downloadFile(url); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func downloadFile(url string) error {
	startTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("start at %s\n", startTime)

	// Create HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("sending request, awaiting response... status %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Get file size
	size := resp.ContentLength
	fmt.Printf("content size: %d [~%.2fMB]\n", size, float64(size)/(1024*1024))

	// Get filename from URL
	fileName := filepath.Base(url)
	fmt.Printf("saving file to: ./%s\n", fileName)

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Copy the response body to the file
	// TODO: Implement progress bar
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	endTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\nDownloaded [%s]\n", url)
	fmt.Printf("finished at %s\n", endTime)

	return nil
}
