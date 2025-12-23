package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client wraps HTTP operations with Bandcamp-specific configuration.
//
// Client provides:
//   - Configured User-Agent header for Bandcamp compatibility
//   - Timeout handling
//   - File download with progress tracking
//   - File size retrieval via HEAD requests
//
// Example usage:
//
//	client := NewClient()
//	
//	// Fetch HTML content
//	html, err := client.GetString(ctx, "https://artist.bandcamp.com/album/name")
//	
//	// Download file with progress
//	err = client.DownloadFile(ctx, mp3URL, "/path/to/file.mp3", func(written, total int64) {
//	    percent := float64(written) / float64(total) * 100
//	    fmt.Printf("%.1f%%\n", percent)
//	})
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new HTTP client configured for Bandcamp.
//
// The client is configured with:
//   - 60 second timeout
//   - "BandcampDownloader" User-Agent header
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		userAgent: "BandcampDownloader",
	}
}

// ProgressWriter wraps a writer to track download progress.
//
// Use this to monitor large downloads by providing an OnUpdate callback
// that receives the current bytes written and total expected bytes.
//
// Example:
//
//	pw := &ProgressWriter{
//	    Writer: file,
//	    Total:  contentLength,
//	    OnUpdate: func(written, total int64) {
//	        fmt.Printf("%d / %d bytes\n", written, total)
//	    },
//	}
//	io.Copy(pw, response.Body)
type ProgressWriter struct {
	// Writer is the underlying writer to write data to.
	Writer io.Writer
	
	// Total is the expected total bytes (from Content-Length header).
	Total int64
	
	// Written is the current number of bytes written.
	Written int64
	
	// OnUpdate is called after each Write with current progress.
	// Parameters are (bytesWritten, totalExpected).
	OnUpdate func(written, total int64)
}

// Write implements io.Writer, tracking progress and calling OnUpdate.
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.Written += int64(n)
	if pw.OnUpdate != nil {
		pw.OnUpdate(pw.Written, pw.Total)
	}
	return n, err
}

// Get performs a GET request and returns the response body as bytes.
//
// The request includes the configured User-Agent header.
//
// Returns an error if:
//   - The request fails
//   - The response status is not 200 OK
//   - Reading the body fails
//
// Example:
//
//	data, err := client.Get(ctx, "https://example.com/image.jpg")
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// GetString performs a GET request and returns the response body as a string.
//
// This is a convenience wrapper around Get for fetching text content like HTML.
//
// Example:
//
//	html, err := client.GetString(ctx, "https://artist.bandcamp.com/album/name")
func (c *Client) GetString(ctx context.Context, url string) (string, error) {
	body, err := c.Get(ctx, url)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetFileSize returns the size of a file at the given URL via HEAD request.
//
// This is useful for:
//   - Pre-calculating total download size
//   - Checking if a local file matches the expected size
//
// Returns an error if:
//   - The request fails
//   - The server doesn't return a Content-Length header
//
// Example:
//
//	size, err := client.GetFileSize(ctx, mp3URL)
//	fmt.Printf("File is %d bytes\n", size)
func (c *Client) GetFileSize(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.ContentLength < 0 {
		return 0, fmt.Errorf("no Content-Length header for %s", url)
	}

	return resp.ContentLength, nil
}

// DownloadFile downloads a file to the specified path with optional progress callback.
//
// The file is created (or truncated if it exists) and the content is streamed
// directly to disk, avoiding loading the entire file into memory.
//
// Parameters:
//   - ctx: Context for cancellation
//   - url: URL to download from
//   - destPath: Local file path to save to
//   - onProgress: Optional callback called with (bytesWritten, totalBytes)
//                 Pass nil to disable progress tracking
//
// Example:
//
//	err := client.DownloadFile(ctx, mp3URL, "/music/song.mp3", func(written, total int64) {
//	    if total > 0 {
//	        fmt.Printf("%.1f%%\r", float64(written)/float64(total)*100)
//	    }
//	})
func (c *Client) DownloadFile(ctx context.Context, url, destPath string, onProgress func(written, total int64)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var writer io.Writer = file
	if onProgress != nil {
		writer = &ProgressWriter{
			Writer:   file,
			Total:    resp.ContentLength,
			OnUpdate: onProgress,
		}
	}

	_, err = io.Copy(writer, resp.Body)
	return err
}

// DownloadBytes downloads a file and returns the bytes in memory.
//
// Use this for small files like cover art images. For large files like
// MP3s, use DownloadFile to stream directly to disk.
//
// Example:
//
//	imageData, err := client.DownloadBytes(ctx, artworkURL)
func (c *Client) DownloadBytes(ctx context.Context, url string) ([]byte, error) {
	return c.Get(ctx, url)
}
