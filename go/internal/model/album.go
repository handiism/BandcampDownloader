package model

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Album represents a Bandcamp album with its metadata and tracks.
//
// Album contains all the information needed to download and organize music files:
//   - Artist and Title for metadata and file naming
//   - ArtworkURL for downloading cover art
//   - ReleaseDate for date-based organization
//   - Computed paths for saving files locally
//
// Paths are automatically computed when creating an album via NewAlbum,
// using placeholders like {artist}, {album}, {year} etc.
//
// Example:
//
//	cfg := &PathConfig{
//	    DownloadsPath: "/music/{artist}/{album}",
//	    CoverArtFileNameFormat: "cover",
//	    PlaylistFormat: PlaylistFormatM3U,
//	}
//	album := NewAlbum("The Beatles", "Abbey Road", artURL, releaseDate, cfg)
//	// album.Path = "/music/The Beatles/Abbey Road"
type Album struct {
	// Artist is the album artist name.
	Artist string

	// Title is the album title.
	Title string

	// ArtworkURL is the URL to download the album cover art from.
	// Empty string means no artwork is available.
	ArtworkURL string

	// ReleaseDate is when the album was released.
	ReleaseDate time.Time

	// Tracks contains all tracks in this album.
	Tracks []*Track

	// Path is the computed local directory path where album files will be saved.
	// This is automatically set by NewAlbum based on PathConfig.DownloadsPath.
	Path string

	// ArtworkPath is the computed local file path for the cover art.
	// Empty if the album has no artwork.
	ArtworkPath string

	// PlaylistPath is the computed local file path for the playlist file.
	PlaylistPath string
}

// NewAlbum creates a new Album with computed paths based on settings.
//
// The pathConfig determines how file paths are constructed using placeholders:
//   - {artist} - Artist name
//   - {album} - Album title
//   - {year} - Release year (4 digits)
//   - {month} - Release month (2 digits, zero-padded)
//   - {day} - Release day (2 digits, zero-padded)
//
// Invalid filename characters are automatically replaced with underscores.
// Paths are truncated if they exceed Windows path length limits (248 for folders, 260 for files).
func NewAlbum(artist, title, artworkURL string, releaseDate time.Time, cfg *PathConfig) *Album {
	album := &Album{
		Artist:      artist,
		Title:       title,
		ArtworkURL:  artworkURL,
		ReleaseDate: releaseDate,
	}

	album.Path = album.parseFolderPath(cfg)
	album.PlaylistPath = album.parsePlaylistPath(cfg)
	album.ArtworkPath = album.parseArtworkPath(cfg)

	return album
}

// HasArtwork returns true if the album has cover art available for download.
func (a *Album) HasArtwork() bool {
	return a.ArtworkURL != ""
}

// PathConfig holds path formatting settings for albums and tracks.
//
// All path fields support placeholders that are replaced with actual values:
//   - {artist} - Artist name
//   - {album} - Album title
//   - {year}, {month}, {day} - Release date components
//
// Example configuration:
//
//	cfg := &PathConfig{
//	    DownloadsPath:         "/home/user/Music/{artist}/{album}",
//	    CoverArtFileNameFormat: "cover",
//	    PlaylistFileNameFormat: "{album}",
//	    PlaylistFormat:         PlaylistFormatM3U,
//	}
type PathConfig struct {
	// DownloadsPath is the base path template for saving albums.
	// Example: "/music/{artist}/{album}"
	DownloadsPath string

	// CoverArtFileNameFormat is the filename template for cover art (without extension).
	// Example: "cover" or "{album}"
	CoverArtFileNameFormat string

	// PlaylistFileNameFormat is the filename template for playlists (without extension).
	// Example: "{album}"
	PlaylistFileNameFormat string

	// PlaylistFormat determines the playlist file type and extension.
	PlaylistFormat PlaylistFormat
}

// PlaylistFormat represents supported playlist file formats.
type PlaylistFormat int

const (
	// PlaylistFormatM3U creates .m3u playlist files (most widely supported).
	PlaylistFormatM3U PlaylistFormat = iota

	// PlaylistFormatPLS creates .pls playlist files (used by Winamp).
	PlaylistFormatPLS

	// PlaylistFormatWPL creates .wpl playlist files (Windows Media Player).
	PlaylistFormatWPL

	// PlaylistFormatZPL creates .zpl playlist files (Zune Media Player).
	PlaylistFormatZPL
)

// Extension returns the file extension for the playlist format, including the dot.
//
// Returns:
//   - ".m3u" for PlaylistFormatM3U
//   - ".pls" for PlaylistFormatPLS
//   - ".wpl" for PlaylistFormatWPL
//   - ".zpl" for PlaylistFormatZPL
func (pf PlaylistFormat) Extension() string {
	switch pf {
	case PlaylistFormatM3U:
		return ".m3u"
	case PlaylistFormatPLS:
		return ".pls"
	case PlaylistFormatWPL:
		return ".wpl"
	case PlaylistFormatZPL:
		return ".zpl"
	default:
		return ".m3u"
	}
}

// parseFolderPath computes the album folder path from the config template.
func (a *Album) parseFolderPath(cfg *PathConfig) string {
	path := cfg.DownloadsPath
	path = strings.ReplaceAll(path, "{year}", sanitizeFileName(a.ReleaseDate.Format("2006")))
	path = strings.ReplaceAll(path, "{month}", sanitizeFileName(a.ReleaseDate.Format("01")))
	path = strings.ReplaceAll(path, "{day}", sanitizeFileName(a.ReleaseDate.Format("02")))
	path = strings.ReplaceAll(path, "{artist}", sanitizeFileName(a.Artist))
	path = strings.ReplaceAll(path, "{album}", sanitizeFileName(a.Title))

	// Limit path length for cross-platform compatibility (Windows MAX_PATH)
	if len(path) >= 248 {
		path = path[:247]
	}

	return path
}

// parsePlaylistPath computes the full playlist file path.
func (a *Album) parsePlaylistPath(cfg *PathConfig) string {
	fileName := a.parsePlaylistFileName(cfg)
	ext := cfg.PlaylistFormat.Extension()
	filePath := filepath.Join(a.Path, fileName+ext)

	// Limit total path length for Windows compatibility
	if len(filePath) >= 260 {
		maxLen := 11 - len(ext)
		if maxLen > 0 && maxLen < len(fileName) {
			filePath = filepath.Join(a.Path, fileName[:maxLen]+ext)
		}
	}

	return filePath
}

// parsePlaylistFileName computes the playlist filename from the config template.
func (a *Album) parsePlaylistFileName(cfg *PathConfig) string {
	fileName := cfg.PlaylistFileNameFormat
	fileName = strings.ReplaceAll(fileName, "{year}", a.ReleaseDate.Format("2006"))
	fileName = strings.ReplaceAll(fileName, "{month}", a.ReleaseDate.Format("01"))
	fileName = strings.ReplaceAll(fileName, "{day}", a.ReleaseDate.Format("02"))
	fileName = strings.ReplaceAll(fileName, "{album}", a.Title)
	fileName = strings.ReplaceAll(fileName, "{artist}", a.Artist)
	return sanitizeFileName(fileName)
}

// parseArtworkPath computes the full cover art file path.
func (a *Album) parseArtworkPath(cfg *PathConfig) string {
	if !a.HasArtwork() {
		return ""
	}

	ext := filepath.Ext(a.ArtworkURL)
	fileName := a.parseCoverArtFileName(cfg)
	artworkPath := filepath.Join(a.Path, fileName+ext)

	// Limit total path length for Windows compatibility
	if len(artworkPath) >= 260 {
		maxLen := 11 - len(ext)
		if maxLen > 0 && maxLen < len(fileName) {
			artworkPath = filepath.Join(a.Path, fileName[:maxLen]+ext)
		}
	}

	return artworkPath
}

// parseCoverArtFileName computes the cover art filename from the config template.
func (a *Album) parseCoverArtFileName(cfg *PathConfig) string {
	fileName := cfg.CoverArtFileNameFormat
	fileName = strings.ReplaceAll(fileName, "{year}", a.ReleaseDate.Format("2006"))
	fileName = strings.ReplaceAll(fileName, "{month}", a.ReleaseDate.Format("01"))
	fileName = strings.ReplaceAll(fileName, "{day}", a.ReleaseDate.Format("02"))
	fileName = strings.ReplaceAll(fileName, "{album}", a.Title)
	fileName = strings.ReplaceAll(fileName, "{artist}", a.Artist)
	return sanitizeFileName(fileName)
}

// sanitizeFileName removes or replaces characters that are invalid in file/folder names.
//
// The following transformations are applied:
//   - Invalid characters (<>:"/\|?* and control chars) are replaced with underscore
//   - Trailing dots are removed (Windows limitation)
//   - Multiple whitespace is collapsed to single space
//   - Trailing whitespace is removed
//
// Example:
//
//	sanitizeFileName("Song: Part 1/2") // Returns "Song_ Part 1_2"
func sanitizeFileName(name string) string {
	// Replace invalid path/file characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	name = invalidChars.ReplaceAllString(name, "_")

	// Remove trailing dots
	name = regexp.MustCompile(`\.+$`).ReplaceAllString(name, "")

	// Replace multiple whitespace with single space
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	// Remove trailing whitespace
	name = strings.TrimRight(name, " ")

	return name
}
