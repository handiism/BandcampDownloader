package model

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Track represents a single track within an album.
//
// Track contains metadata for one song including:
//   - Track number and title for ID3 tagging
//   - Duration for playlist generation
//   - Lyrics (if available on Bandcamp)
//   - MP3 download URL
//   - Computed local file path
//
// The file path is automatically computed when creating a track via NewTrack,
// using the album's path and the TrackConfig file name format.
//
// Example:
//
//	cfg := &TrackConfig{FileNameFormat: "{tracknum} {title}.mp3"}
//	track := NewTrack(album, 1, "Song Title", 180.5, "", mp3URL, cfg)
//	// track.Path = "/music/Artist/Album/01 Song Title.mp3"
type Track struct {
	// Album is a reference to the parent album.
	Album *Album

	// Number is the track number (1-indexed).
	Number int

	// Title is the track title.
	Title string

	// Duration is the track length in seconds.
	Duration float64

	// Lyrics contains the song lyrics, if available.
	// Empty string if no lyrics are available.
	Lyrics string

	// Mp3URL is the URL to download the MP3 file from.
	Mp3URL string

	// Path is the computed local file path where the track will be saved.
	// Includes the full path and filename with extension.
	Path string
}

// TrackConfig holds track path formatting settings.
//
// The FileNameFormat supports placeholders that are replaced with actual values:
//   - {tracknum} - Track number (2 digits, zero-padded)
//   - {title} - Track title
//   - {artist} - Artist name (from album)
//   - {album} - Album title
//   - {year}, {month}, {day} - Release date components
//
// Example:
//
//	cfg := &TrackConfig{
//	    FileNameFormat: "{tracknum} {artist} - {title}.mp3",
//	}
//	// Results in filenames like "01 The Beatles - Come Together.mp3"
type TrackConfig struct {
	// FileNameFormat is the template for track filenames.
	// Must include the file extension (typically ".mp3").
	FileNameFormat string
}

// NewTrack creates a new Track with computed path.
//
// Parameters:
//   - album: The parent album (required for path computation and metadata)
//   - number: Track number (1-indexed, used for filename and ID3 tag)
//   - title: Track title
//   - duration: Track length in seconds (used for playlists)
//   - lyrics: Song lyrics (empty string if not available)
//   - mp3URL: URL to download the MP3 from
//   - cfg: Configuration for file naming
//
// The file path is computed using the album's path and the configured filename format.
// Invalid filename characters are automatically replaced with underscores.
func NewTrack(album *Album, number int, title string, duration float64, lyrics, mp3URL string, cfg *TrackConfig) *Track {
	track := &Track{
		Album:    album,
		Number:   number,
		Title:    title,
		Duration: duration,
		Lyrics:   lyrics,
		Mp3URL:   mp3URL,
	}

	track.Path = track.parseFilePath(cfg)

	return track
}

// parseFilePath computes the full file path for this track.
func (t *Track) parseFilePath(cfg *TrackConfig) string {
	fileName := t.parseFileName(cfg)
	filePath := filepath.Join(t.Album.Path, fileName)

	// Limit total path length for Windows compatibility (MAX_PATH = 260)
	if len(filePath) >= 260 {
		ext := filepath.Ext(filePath)
		maxLen := 11 - len(ext) // Leave room for path separator and extension
		if maxLen > 0 && maxLen < len(fileName) {
			filePath = filepath.Join(t.Album.Path, fileName[:maxLen]+ext)
		}
	}

	return filePath
}

// parseFileName computes the filename from the config template.
func (t *Track) parseFileName(cfg *TrackConfig) string {
	fileName := cfg.FileNameFormat
	fileName = strings.ReplaceAll(fileName, "{year}", t.Album.ReleaseDate.Format("2006"))
	fileName = strings.ReplaceAll(fileName, "{month}", t.Album.ReleaseDate.Format("01"))
	fileName = strings.ReplaceAll(fileName, "{day}", t.Album.ReleaseDate.Format("02"))
	fileName = strings.ReplaceAll(fileName, "{album}", t.Album.Title)
	fileName = strings.ReplaceAll(fileName, "{artist}", t.Album.Artist)
	fileName = strings.ReplaceAll(fileName, "{title}", t.Title)
	fileName = strings.ReplaceAll(fileName, "{tracknum}", fmt.Sprintf("%02d", t.Number))
	return sanitizeFileName(fileName)
}
