# Wget Clone

A wget clone implementation in Go that replicates core functionalities of the original wget utility.

## Description
This project recreates the fundamental features of wget using Go. Wget is a free utility for non-interactive download of files from the Web, supporting HTTP, HTTPS, and FTP protocols.

## Current Features
- Download a file from a given URL
- Display download progress including:
  - Start time
  - Request status
  - File size
  - Progress bar with download speed
  - End time

## Usage
```bash
go run . https://example.com/file.txt
```

## Example Output
```
start at 2025-01-08 19:02:42
sending request, awaiting response... status 200 OK
content size: 56370 [~0.06MB]
saving file to: ./file.txt
55.05 KiB / 55.05 KiB [====================================] 100.00% 1.24 MiB/s 0s
Downloaded [https://example.com/file.txt]
finished at 2025-01-08 19:02:43
```
