package mirrorutils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// MirrorOptions holds the configuration for mirroring a website
type MirrorOptions struct {
	URL          string
	OutputDir    string
	ConvertLinks bool
	UseDynamic   bool
	RejectTypes  []string
	ExcludePaths []string
	visited      map[string]bool
	currentDepth int
	maxDepth     int
	baseHost     string // Store the base host for domain matching
}

// NewMirrorOptions creates a new MirrorOptions instance
func NewMirrorOptions(urlStr, outputDir string, convertLinks bool, rejectTypes []string, excludePaths []string) *MirrorOptions {
	// Parse the base URL to get the host
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse base URL: %v\n", err)
		return nil
	}

	return &MirrorOptions{
		URL:          urlStr,
		OutputDir:    outputDir,
		ConvertLinks: convertLinks,
		RejectTypes:  rejectTypes,
		ExcludePaths: excludePaths,
		visited:      make(map[string]bool),
		maxDepth:     5, // Maximum depth for following links
		baseHost:     baseURL.Host,
	}
}

// Mirror starts the website mirroring process
func (m *MirrorOptions) Mirror() error {
	// Create output directory
	if err := os.MkdirAll(m.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	fmt.Printf("Starting mirror of %s\n", m.URL)
	fmt.Printf("Output directory: %s\n", m.OutputDir)

	return m.processURL(m.URL)
}

// processURL downloads and processes a single URL
func (m *MirrorOptions) processURL(urlStr string) error {
	// Clean the URL by removing fragments and normalizing query parameters
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL %s: %v", urlStr, err)
	}

	// Remove fragments and query parameters for visited check
	cleanURL := *parsedURL
	cleanURL.Fragment = ""
	cleanURL.RawQuery = ""
	urlKey := cleanURL.String()

	// Check if URL has been visited
	if m.visited[urlKey] {
		return nil
	}
	m.visited[urlKey] = true

	// Check depth limit
	if m.currentDepth > m.maxDepth {
		return nil
	}

	// Only process URLs from the same domain
	if parsedURL.Host != "" && parsedURL.Host != m.baseHost {
		fmt.Printf("Skipping external domain: %s\n", urlStr)
		return nil
	}

	// Exclude the js folder
	if strings.Contains(parsedURL.Path, "/js/") {
		// fmt.Printf("Skipping js folder: %s\n", urlStr)
		return nil
	}

	// Check if path matches any exclude patterns
	for _, excludePath := range m.ExcludePaths {
		// Normalize paths by removing leading and trailing slashes
		normalizedExclude := strings.Trim(excludePath, "/")
		normalizedPath := strings.Trim(parsedURL.Path, "/")

		if strings.HasPrefix(normalizedPath, normalizedExclude) {
			fmt.Printf("Skipping excluded path: %s\n", urlStr)
			return nil
		}
	}

	// Check if file matches reject list
	filename := filepath.Base(parsedURL.Path)
	if filename == "" || filename == "/" {
		filename = "index.html"
	}
	
	// Flag to track if this file should be saved
	shouldSaveFile := true
	
	// First check full filename
	for _, rejectedType := range m.RejectTypes {
		if strings.EqualFold(filename, rejectedType) {
			fmt.Printf("Skipping rejected file: %s\n", urlStr)
			shouldSaveFile = false
			// Don't return here, we still want to process the file for links
		}
	}

	// Then check extension
	ext := strings.ToLower(filepath.Ext(parsedURL.Path))
	if ext != "" {
		// Remove leading dot if present
		ext = strings.TrimPrefix(ext, ".")
		for _, rejectedType := range m.RejectTypes {
			if strings.EqualFold(ext, rejectedType) {
				fmt.Printf("Skipping rejected file type: %s\n", urlStr)
				shouldSaveFile = false
				// Don't return here, we still want to process the file for links
			}
		}
	}

	// Skip certain file types
	if ext == "exe" || ext == "zip" || ext == "pdf" || ext == "dmg" {
		fmt.Printf("Skipping excluded file type: %s\n", urlStr)
		return nil
	}

	fmt.Printf("Downloading: %s\n", urlStr)

	// Download the URL
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers to make the request more browser-like
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status code %d", urlStr, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Prepare output path for all cases
	outputPath := filepath.Join(m.OutputDir, m.convertToLocalPath(parsedURL))
	
	// If path ends with a slash, append index.html
	if strings.HasSuffix(outputPath, "/") || outputPath == m.OutputDir {
		outputPath = filepath.Join(outputPath, "index.html")
	}

	// Ensure the file doesn't exist as a directory
	if info, err := os.Stat(outputPath); err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, "index.html")
	}

	// Save file if not rejected
	if shouldSaveFile {
		// Create directory if it doesn't exist
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(outputPath, body, 0644); err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
	}

	// Process HTML content
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		doc, err := html.Parse(bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to parse HTML: %v", err)
		}

		var processNode func(*html.Node)
		processNode = func(n *html.Node) {
			if n.Type == html.ElementNode {
				// Process attributes
				for i := 0; i < len(n.Attr); i++ {
					attr := n.Attr[i]
					switch attr.Key {
					case "href", "src":
						// Convert link to absolute URL
						absURL, err := m.resolveURL(parsedURL, attr.Val)
						if err != nil {
							fmt.Printf("Warning: Failed to resolve URL %s: %v\n", attr.Val, err)
							continue
						}

						// Skip certain resource types
						if strings.Contains(absURL.String(), "google-analytics.com") ||
							strings.Contains(absURL.String(), "analytics.js") {
							continue
						}

						// Only process URLs from the same domain
						if absURL.Host == m.baseHost {
							// Update attribute to use local path
							localPath := m.convertLinkPath(parsedURL, absURL)
							if m.ConvertLinks {
								n.Attr[i].Val = localPath
							}

							// Skip downloading if it's a fragment or query-only change
							cleanAbsURL := *absURL
							cleanAbsURL.Fragment = ""
							cleanAbsURL.RawQuery = ""
							if m.visited[cleanAbsURL.String()] {
								continue
							}

							// Download linked resource
							m.currentDepth++
							if err := m.processURL(absURL.String()); err != nil {
								fmt.Printf("Warning: Failed to process URL %s: %v\n", absURL.String(), err)
							}
							m.currentDepth--
						}
					case "style":
						// Extract URLs from inline styles
						urls := extractURLsFromCSS(attr.Val)
						for _, cssURL := range urls {
							absURL, err := m.resolveURL(parsedURL, cssURL)
							if err != nil {
								fmt.Printf("Warning: Failed to resolve URL %s: %v\n", cssURL, err)
								continue
							}

							if absURL.Host == m.baseHost {
								localPath := m.convertLinkPath(parsedURL, absURL)
								if m.ConvertLinks {
									// Replace the URL in the style attribute with the local path
									attr.Val = strings.ReplaceAll(attr.Val, fmt.Sprintf(`url('%s')`, cssURL), fmt.Sprintf(`url('%s')`, localPath))
									attr.Val = strings.ReplaceAll(attr.Val, fmt.Sprintf(`url("%s")`, cssURL), fmt.Sprintf(`url("%s")`, localPath))
									attr.Val = strings.ReplaceAll(attr.Val, fmt.Sprintf(`url(%s)`, cssURL), fmt.Sprintf(`url(%s')`, localPath))
									n.Attr[i] = attr
								}

								cleanAbsURL := *absURL
								cleanAbsURL.Fragment = ""
								cleanAbsURL.RawQuery = ""
								if m.visited[cleanAbsURL.String()] {
									continue
								}

								m.currentDepth++
								if err := m.processURL(absURL.String()); err != nil {
									fmt.Printf("Warning: Failed to process URL %s: %v\n", absURL.String(), err)
								}
								m.currentDepth--
							}
						}
					case "integrity":
						// Remove integrity attributes as they may prevent local resources from loading
						if i < len(n.Attr)-1 {
							n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
							i-- // Adjust index since we removed an element
						} else {
							n.Attr = n.Attr[:i]
						}
					}
				}

				// Process <style> tags
				if n.Data == "style" && n.FirstChild != nil {
					cssContent := n.FirstChild.Data
					urls := extractURLsFromCSS(cssContent)
					for _, cssURL := range urls {
						absURL, err := m.resolveURL(parsedURL, cssURL)
						if err != nil {
							fmt.Printf("Warning: Failed to resolve URL %s: %v\n", cssURL, err)
							continue
						}

						if absURL.Host == m.baseHost {
							localPath := m.convertLinkPath(parsedURL, absURL)
							if m.ConvertLinks {
								// Replace the URL in the style tag with the local path
								cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url('%s')`, cssURL), fmt.Sprintf(`url('%s')`, localPath))
								cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url("%s")`, cssURL), fmt.Sprintf(`url("%s")`, localPath))
								cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url(%s)`, cssURL), fmt.Sprintf(`url(%s')`, localPath))
								n.FirstChild.Data = cssContent
							}

							cleanAbsURL := *absURL
							cleanAbsURL.Fragment = ""
							cleanAbsURL.RawQuery = ""
							if m.visited[cleanAbsURL.String()] {
								continue
							}

							m.currentDepth++
							if err := m.processURL(absURL.String()); err != nil {
								fmt.Printf("Warning: Failed to process URL %s: %v\n", absURL.String(), err)
							}
							m.currentDepth--
						}
					}
				}
			}

			// Process child nodes
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				processNode(c)
			}
		}
		processNode(doc)

		// Write the updated HTML back to the file if not rejected
		if shouldSaveFile {
			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				return fmt.Errorf("failed to render HTML: %v", err)
			}

			// Write the updated HTML back to the file
			if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
				return fmt.Errorf("failed to write updated HTML: %v", err)
			}
		}
	} else if strings.Contains(contentType, "text/css") {
		// Process CSS files
		cssContent := string(body)
		urls := extractURLsFromCSS(cssContent)
		for _, cssURL := range urls {
			absURL, err := m.resolveURL(parsedURL, cssURL)
			if err != nil {
				fmt.Printf("Warning: Failed to resolve URL %s: %v\n", cssURL, err)
				continue
			}

			if absURL.Host == m.baseHost {
				localPath := m.convertLinkPath(parsedURL, absURL)
				if m.ConvertLinks {
					// Replace the URL in the CSS file with the local path
					cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url('%s')`, cssURL), fmt.Sprintf(`url('%s')`, localPath))
					cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url("%s")`, cssURL), fmt.Sprintf(`url("%s")`, localPath))
					cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url(%s)`, cssURL), fmt.Sprintf(`url(%s')`, localPath))
				}

				cleanAbsURL := *absURL
				cleanAbsURL.Fragment = ""
				cleanAbsURL.RawQuery = ""
				if m.visited[cleanAbsURL.String()] {
					continue
				}

				m.currentDepth++
				if err := m.processURL(absURL.String()); err != nil {
					fmt.Printf("Warning: Failed to process URL %s: %v\n", absURL.String(), err)
				}
				m.currentDepth--
			}
		}

		// Write the updated CSS back to the file if not rejected
		if shouldSaveFile {
			if err := os.WriteFile(outputPath, []byte(cssContent), 0644); err != nil {
				return fmt.Errorf("failed to write updated CSS: %v", err)
			}
		}
	}

	return nil
}

// resolveURL converts a relative URL to an absolute URL
func (m *MirrorOptions) resolveURL(base *url.URL, ref string) (*url.URL, error) {
	refURL, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(refURL), nil
}

// convertToLocalPath converts a URL to a local file path
func (m *MirrorOptions) convertToLocalPath(u *url.URL) string {
	// Get the path without query parameters and fragments
	cleanPath := u.Path
	
	// Split the path into components
	parts := strings.Split(strings.TrimPrefix(cleanPath, "/"), "/")
	
	// Handle dynamic paths
	if len(parts) > 0 {
		// Convert paths with extensions to files
		if hasFileExtension(parts[len(parts)-1]) {
			return filepath.Join(u.Host, cleanPath)
		}
		
		// Handle paths that look like API endpoints or dynamic routes
		// Examples: /api/v1/users, /users/123, /repo/branch/path
		if containsNumericID(cleanPath) || containsDynamicSegments(parts) {
			// Store under a 'pages' directory to separate dynamic content
			return filepath.Join(u.Host, "pages", cleanPath, "index.html")
		}
	}
	
	// Default handling for other paths
	path := filepath.Join(u.Host, strings.TrimPrefix(cleanPath, "/"))
	
	// If path is empty or ends with a slash, append index.html
	if path == u.Host || strings.HasSuffix(path, "/") || !hasFileExtension(path) {
		path = filepath.Join(path, "index.html")
	}
	
	return path
}

// convertLinkPath converts a URL to a relative link path for use in HTML
func (m *MirrorOptions) convertLinkPath(base, ref *url.URL) string {
	// If the reference URL is absolute (starts with a protocol), keep it as is
	if ref.Scheme != "" || ref.Host != "" {
		if ref.Host != base.Host {
			// External link, keep it as is
			return ref.String()
		}
	}

	// Get the local path for the reference URL
	localPath := m.convertToLocalPath(ref)

	// Get the local path for the base URL
	basePath := m.convertToLocalPath(base)
	baseDir := filepath.Dir(basePath)

	// Calculate the relative path from base to ref
	rel, err := filepath.Rel(baseDir, localPath)
	if err != nil {
		// Fallback to absolute path if relative path calculation fails
		return "/" + localPath
	}

	// Convert Windows backslashes to forward slashes for URLs
	return strings.ReplaceAll(rel, "\\", "/")
}

// hasFileExtension checks if a path has a file extension
func hasFileExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext != "" && !strings.Contains(ext, "/")
}

// containsNumericID checks if a path contains numeric IDs
// Examples: /users/123, /posts/456/comments
func containsNumericID(path string) bool {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if isNumeric(part) {
			return true
		}
	}
	return false
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// containsDynamicSegments checks if path parts look like dynamic segments
// Examples: /blob/master/file.txt, /tree/main/src, /api/v1/users
func containsDynamicSegments(parts []string) bool {
	// Common dynamic path patterns
	dynamicPatterns := []string{
		"api",
		"v1", "v2", "v3",  // API versions
		"blob", "tree",    // Repository patterns
		"branch", "tag",
		"commit", "pull",
		"issues", "wiki",
		"raw", "edit",
	}
	
	for _, part := range parts {
		for _, pattern := range dynamicPatterns {
			if strings.EqualFold(part, pattern) {
				return true
			}
		}
	}
	
	return false
}

// extractURLsFromCSS extracts URLs from CSS content
func extractURLsFromCSS(css string) []string {
	var urls []string
	// Match url(...) patterns in CSS
	urlPattern := regexp.MustCompile(`url\(['"]?([^'"\)]+)['"]?\)`)
	matches := urlPattern.FindAllStringSubmatch(css, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			urls = append(urls, match[1])
		}
	}
	return urls
}

// Hardcoded URL for mirroring
const hardcodedURL = "https://trypap.com/"

// Start the mirroring process
func StartMirroring() {
	options := NewMirrorOptions(hardcodedURL, "output_directory", true, nil, nil)
	if err := options.Mirror(); err != nil {
		fmt.Printf("Error during mirroring: %v\n", err)
	}
}