// Package http provides an HTTP client configured for Bandcamp API requests.
//
// The Client in this package handles:
//   - User-Agent headers for Bandcamp compatibility
//   - File downloads with progress tracking
//   - File size retrieval via HEAD requests
//   - Timeout handling
//
// # Basic Usage
//
//	client := http.NewClient()
//	
//	// Fetch HTML page
//	html, err := client.GetString(ctx, "https://artist.bandcamp.com/album/name")
//	
//	// Download file with progress callback
//	client.DownloadFile(ctx, mp3URL, "/path/to/file.mp3", func(written, total int64) {
//	    fmt.Printf("%.1f%%\n", float64(written)/float64(total)*100)
//	})
//
// # Progress Tracking
//
// The ProgressWriter type can be used to wrap any io.Writer for progress tracking:
//
//	pw := &http.ProgressWriter{
//	    Writer:   file,
//	    Total:    contentLength,
//	    OnUpdate: func(written, total int64) { /* update UI */ },
//	}
package http
