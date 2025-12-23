// Package download provides the download orchestration logic for
// fetching albums and tracks from Bandcamp.
//
// # Manager
//
// The Manager coordinates the entire download process:
//
//  1. Parse input URLs
//  2. Fetch album information from Bandcamp
//  3. Download cover art
//  4. Download tracks concurrently
//  5. Tag MP3 files with ID3 metadata
//  6. Generate playlists (optional)
//
// # Basic Usage
//
//	manager := download.NewManager(settings, func(event download.ProgressEvent) {
//	    fmt.Println(event.Message)
//	})
//
//	err := manager.Initialize(ctx, "https://artist.bandcamp.com/album/name")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = manager.StartDownloads(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Concurrency
//
// The Manager uses configurable concurrency limits:
//   - MaxConcurrentAlbumsDownload: How many albums to download in parallel
//   - MaxConcurrentTracksDownload: How many tracks per album to download in parallel
//
// # Progress Tracking
//
// Progress is reported via a callback function that receives ProgressEvent:
//
//	type ProgressEvent struct {
//	    Message string
//	    Level   ProgressLevel // Info, Verbose, Warning, Error, Success
//	}
//
// # Retry Logic
//
// Failed downloads are automatically retried with exponential backoff,
// configurable via settings.DownloadMaxRetries and settings.DownloadRetryCooldown.
package download
