# Wget Clone

A wget clone implementation in Go that replicates core functionalities of the original wget utility.

## Description
This project recreates the fundamental features of wget using Go. Wget is a free utility for non-interactive download of files from the Web, supporting HTTP, HTTPS, and FTP protocols.

## Project Objective
This project aims to recreate functionalities of wget using Go. The main objectives include:

- Downloading a file from a given URL, e.g., `wget https://some_url.org/file.zip`
- Saving a downloaded file under a different name.
- Specifying a directory for saving downloaded files.
- Limiting download speed.
- Supporting background downloads.
- Downloading multiple files simultaneously from a list of URLs.
- Mirroring an entire website.

## Functionalities
- Normal usage: `wget <URL>`
- Downloading with different flags:
  - `-O` for saving under a different name.
  - `-P` for specifying a save directory.
  - `--rate-limit` for setting download speed.
  - `-B` for background download with logging.
  - `-i` for downloading multiple files from a text file.
  - `--mirror` for mirroring websites with various options.

## Introduction
Wget is a free utility for non-interactive download of files from the Web. It supports HTTP, HTTPS, and FTP protocols, as well as retrieval through HTTP proxies.

To see more about wget, you can visit the manual by using the command `man wget`, or you can visit the website [here](https://www.gnu.org/software/wget/).

## Current Features
- Download a file from a given URL
- Mirror websites with:
  - Recursive downloading
  - Link conversion for offline viewing
  - Cross-origin resource handling
  - Depth control
- Multiple URL downloads from input file
- Background downloads with logging
- Progress tracking including:
  - Start time
  - Request status
  - File size
  - Progress bar with download speed
  - End time

## Usage

### Single File Download
```bash
go run . https://example.com/file.txt
```

### Mirror Website
```bash
go run . --mirror --convert-links https://example.com
```

### Download Multiple URLs
```bash
go run . -i urls.txt
```

### Background Download
```bash
go run . -B https://example.com/file.txt
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

## Viewing Mirrored Websites
After mirroring a website, you can use any static file server to view the content. For example:
- Use VS Code's Live Server extension
- Use Python's built-in server: `python -m http.server`
- Use Node.js's `live-server` package

