package audio

import (
	"strings"
	"testing"
	"time"

	"github.com/handiism/bandcamp-downloader/internal/model"
)

func TestPlaylistCreator_M3U(t *testing.T) {
	album := createTestAlbum()
	creator := NewPlaylistCreator(FormatM3U, false)

	content := creator.CreatePlaylist(album)

	// Check basic format
	if !strings.Contains(content, "track1.mp3") {
		t.Error("M3U should contain track filename")
	}
}

func TestPlaylistCreator_M3UExtended(t *testing.T) {
	album := createTestAlbum()
	creator := NewPlaylistCreator(FormatM3U, true)

	content := creator.CreatePlaylist(album)

	if !strings.HasPrefix(content, "#EXTM3U") {
		t.Error("Extended M3U should start with #EXTM3U")
	}
	if !strings.Contains(content, "#EXTINF:") {
		t.Error("Extended M3U should contain #EXTINF")
	}
}

func TestPlaylistCreator_PLS(t *testing.T) {
	album := createTestAlbum()
	creator := NewPlaylistCreator(FormatPLS, false)

	content := creator.CreatePlaylist(album)

	if !strings.HasPrefix(content, "[playlist]") {
		t.Error("PLS should start with [playlist]")
	}
	if !strings.Contains(content, "File1=") {
		t.Error("PLS should contain File1=")
	}
	if !strings.Contains(content, "NumberOfEntries=") {
		t.Error("PLS should contain NumberOfEntries")
	}
}

func TestPlaylistCreator_WPL(t *testing.T) {
	album := createTestAlbum()
	creator := NewPlaylistCreator(FormatWPL, false)

	content := creator.CreatePlaylist(album)

	if !strings.Contains(content, "<?wpl") {
		t.Error("WPL should contain XML declaration")
	}
	if !strings.Contains(content, "<smil>") {
		t.Error("WPL should contain smil element")
	}
	if !strings.Contains(content, "<media src=") {
		t.Error("WPL should contain media elements")
	}
}

func TestPlaylistCreator_ZPL(t *testing.T) {
	album := createTestAlbum()
	creator := NewPlaylistCreator(FormatZPL, false)

	content := creator.CreatePlaylist(album)

	if !strings.Contains(content, "<?zpl") {
		t.Error("ZPL should contain XML declaration")
	}
	if !strings.Contains(content, "albumTitle=") {
		t.Error("ZPL should contain albumTitle attribute")
	}
}

func TestPlaylistCreator_XMLEscape(t *testing.T) {
	albumCfg := &model.PathConfig{
		DownloadsPath:          "/music",
		CoverArtFileNameFormat: "{album}",
		PlaylistFileNameFormat: "{album}",
	}
	trackCfg := &model.TrackConfig{
		FileNameFormat: "{title}.mp3",
	}

	album := model.NewAlbum("Artist & Co", "Album <Special>", "", time.Now(), albumCfg)
	track := model.NewTrack(album, 1, 1, "Track & \"Quote\"", 180, "", "http://example.com", trackCfg)
	album.Tracks = append(album.Tracks, track)

	creator := NewPlaylistCreator(FormatWPL, false)
	content := creator.CreatePlaylist(album)

	if strings.Contains(content, "&") && !strings.Contains(content, "&amp;") {
		t.Error("WPL should escape & as &amp;")
	}
	if strings.Contains(content, "<Special>") {
		t.Error("WPL should escape < and >")
	}
}

func createTestAlbum() *model.Album {
	albumCfg := &model.PathConfig{
		DownloadsPath:          "/music/{artist}/{album}",
		CoverArtFileNameFormat: "{album}",
		PlaylistFileNameFormat: "{album}",
	}
	trackCfg := &model.TrackConfig{
		FileNameFormat: "{title}.mp3",
	}

	album := model.NewAlbum("Test Artist", "Test Album", "", time.Now(), albumCfg)

	track1 := model.NewTrack(album, 1, 1, "track1", 180, "", "http://example.com/1.mp3", trackCfg)
	track2 := model.NewTrack(album, 1, 2, "track2", 200, "", "http://example.com/2.mp3", trackCfg)

	album.Tracks = []*model.Track{track1, track2}

	return album
}
