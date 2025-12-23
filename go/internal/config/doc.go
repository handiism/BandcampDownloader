// Package config provides configuration management for bandcamp-downloader.
//
// This package handles:
//   - Loading and saving settings from JSON files
//   - Default configuration values
//   - Conversion to PathConfig and TrackConfig for other packages
//
// # Default Settings
//
// Use DefaultSettings() to get sensible defaults:
//
//	settings := config.DefaultSettings()
//	// Downloads to ~/Music/Bandcamp/{artist}/{album}
//	// Concurrent downloads enabled
//	// ID3 tagging enabled
//
// # Loading from File
//
//	settings, err := config.Load("/path/to/config.json")
//	if err != nil {
//	    // Uses defaults if file doesn't exist
//	}
//
// # Saving Settings
//
//	settings.DownloadsPath = "/custom/path/{artist}/{album}"
//	err := settings.Save("/path/to/config.json")
//
// # Configuration Options
//
// Settings includes options for:
//   - Download paths and file naming
//   - Concurrent download limits
//   - Retry behavior
//   - Cover art handling
//   - Playlist generation
//   - ID3 tag modification
//   - Proxy configuration
package config
