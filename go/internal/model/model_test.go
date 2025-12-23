package model

import (
	"testing"
	"time"
)

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal-file.mp3", "normal-file.mp3"},
		{"file:with:colons.mp3", "file_with_colons.mp3"},
		{"file<with>brackets.mp3", "file_with_brackets.mp3"},
		{"file/with\\slashes.mp3", "file_with_slashes.mp3"},
		{"file|with|pipes.mp3", "file_with_pipes.mp3"},
		{"file?with*wildcards.mp3", "file_with_wildcards.mp3"},
		{"file\"with\"quotes.mp3", "file_with_quotes.mp3"},
		{"trailing dots...", "trailing dots"},
		{"multiple   spaces", "multiple spaces"},
		{"trailing spaces   ", "trailing spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFileName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFileName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAlbum_PathComputation(t *testing.T) {
	cfg := &PathConfig{
		DownloadsPath:          "/music/{artist}/{album}",
		CoverArtFileNameFormat: "{album}",
		PlaylistFileNameFormat: "{album}",
		PlaylistFormat:         PlaylistFormatM3U,
	}

	releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
	album := NewAlbum("Test Artist", "Test Album", "https://example.com/art.jpg", releaseDate, cfg)

	if album.Path != "/music/Test Artist/Test Album" {
		t.Errorf("Album.Path = %q, want %q", album.Path, "/music/Test Artist/Test Album")
	}

	// Path should contain the sanitized album name
	if album.ArtworkPath == "" {
		t.Error("ArtworkPath should not be empty when artwork URL is provided")
	}

	if album.PlaylistPath == "" {
		t.Error("PlaylistPath should not be empty")
	}
}

func TestAlbum_NoArtwork(t *testing.T) {
	cfg := &PathConfig{
		DownloadsPath:          "/music/{artist}/{album}",
		CoverArtFileNameFormat: "{album}",
		PlaylistFileNameFormat: "{album}",
		PlaylistFormat:         PlaylistFormatM3U,
	}

	releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
	album := NewAlbum("Test Artist", "Test Album", "", releaseDate, cfg)

	if album.HasArtwork() {
		t.Error("HasArtwork() should return false when ArtworkURL is empty")
	}

	if album.ArtworkPath != "" {
		t.Errorf("ArtworkPath should be empty, got %q", album.ArtworkPath)
	}
}

func TestTrack_PathComputation(t *testing.T) {
	albumCfg := &PathConfig{
		DownloadsPath:          "/music/{artist}/{album}",
		CoverArtFileNameFormat: "{album}",
		PlaylistFileNameFormat: "{album}",
		PlaylistFormat:         PlaylistFormatM3U,
	}
	trackCfg := &TrackConfig{
		FileNameFormat: "{tracknum} {title}.mp3",
	}

	releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
	album := NewAlbum("Artist", "Album", "", releaseDate, albumCfg)
	track := NewTrack(album, 1, "Track Title", 180.5, "", "http://example.com/track.mp3", trackCfg)

	expectedPath := "/music/Artist/Album/01 Track Title.mp3"
	if track.Path != expectedPath {
		t.Errorf("Track.Path = %q, want %q", track.Path, expectedPath)
	}
}

func TestPlaylistFormat_Extension(t *testing.T) {
	tests := []struct {
		format PlaylistFormat
		want   string
	}{
		{PlaylistFormatM3U, ".m3u"},
		{PlaylistFormatPLS, ".pls"},
		{PlaylistFormatWPL, ".wpl"},
		{PlaylistFormatZPL, ".zpl"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.format.Extension(); got != tt.want {
				t.Errorf("Extension() = %q, want %q", got, tt.want)
			}
		})
	}
}
