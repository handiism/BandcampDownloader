package audio

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/handiism/bandcamp-downloader/internal/model"
)

// PlaylistFormat represents supported playlist file formats.
//
// Each format has different features and compatibility:
//   - M3U: Simple text format, widely supported
//   - PLS: INI-style format, used by Winamp
//   - WPL: XML format, Windows Media Player
//   - ZPL: XML format, Zune/Groove Music
type PlaylistFormat int

const (
	// FormatM3U creates .m3u files (most compatible).
	// Can be extended with EXTINF lines for duration/title info.
	FormatM3U PlaylistFormat = iota

	// FormatPLS creates .pls files (Winamp/SHOUTcast format).
	// INI-style format with file, title, and length info.
	FormatPLS

	// FormatWPL creates .wpl files (Windows Media Player).
	// XML-based SMIL format.
	FormatWPL

	// FormatZPL creates .zpl files (Zune/Groove Music).
	// XML-based SMIL format with extended metadata.
	FormatZPL
)

// PlaylistCreator generates playlist files in various formats.
//
// PlaylistCreator takes an album and generates a playlist containing
// all tracks in the album. The output is a string that can be written
// to a file.
//
// Example:
//
//	// Create M3U playlist with extended info
//	creator := NewPlaylistCreator(FormatM3U, true)
//	content := creator.CreatePlaylist(album)
//	os.WriteFile(album.PlaylistPath, []byte(content), 0644)
//
//	// Result:
//	// #EXTM3U
//	// #EXTINF:180,Artist - Song Title
//	// 01 Artist - Song Title.mp3
type PlaylistCreator struct {
	format   PlaylistFormat
	extended bool // For M3U: include EXTINF lines with duration/title
}

// NewPlaylistCreator creates a new PlaylistCreator.
//
// Parameters:
//   - format: The playlist format to generate
//   - extended: For M3U format, whether to include #EXTINF lines
//     (ignored for other formats)
func NewPlaylistCreator(format PlaylistFormat, extended bool) *PlaylistCreator {
	return &PlaylistCreator{
		format:   format,
		extended: extended,
	}
}

// CreatePlaylist generates playlist content for an album.
//
// Returns the playlist as a string, ready to be written to a file.
// Track paths in the playlist are relative (just the filename),
// assuming the playlist file is in the same directory as the tracks.
//
// Example:
//
//	content := creator.CreatePlaylist(album)
//	err := os.WriteFile("/music/Artist/Album/playlist.m3u", []byte(content), 0644)
func (p *PlaylistCreator) CreatePlaylist(album *model.Album) string {
	switch p.format {
	case FormatM3U:
		return p.createM3U(album)
	case FormatPLS:
		return p.createPLS(album)
	case FormatWPL:
		return p.createWPL(album)
	case FormatZPL:
		return p.createZPL(album)
	default:
		return p.createM3U(album)
	}
}

// createM3U generates an M3U playlist.
//
// Standard M3U format:
//
//	filename1.mp3
//	filename2.mp3
//
// Extended M3U format (when extended=true):
//
//	#EXTM3U
//	#EXTINF:180,Artist - Title
//	filename1.mp3
func (p *PlaylistCreator) createM3U(album *model.Album) string {
	var sb strings.Builder

	if p.extended {
		sb.WriteString("#EXTM3U\n")
	}

	for _, track := range album.Tracks {
		if p.extended {
			duration := int(track.Duration)
			sb.WriteString(fmt.Sprintf("#EXTINF:%d,%s - %s\n", duration, album.Artist, track.Title))
		}
		sb.WriteString(filepath.Base(track.Path) + "\n")
	}

	return sb.String()
}

// createPLS generates a PLS playlist.
//
// PLS format is an INI-style text file:
//
//	[playlist]
//	File1=filename1.mp3
//	Title1=Song Title
//	Length1=180
//	NumberOfEntries=2
//	Version=2
func (p *PlaylistCreator) createPLS(album *model.Album) string {
	var sb strings.Builder

	sb.WriteString("[playlist]\n")

	for i, track := range album.Tracks {
		idx := i + 1
		sb.WriteString(fmt.Sprintf("File%d=%s\n", idx, filepath.Base(track.Path)))
		sb.WriteString(fmt.Sprintf("Title%d=%s\n", idx, track.Title))
		sb.WriteString(fmt.Sprintf("Length%d=%d\n", idx, int(track.Duration)))
	}

	sb.WriteString(fmt.Sprintf("NumberOfEntries=%d\n", len(album.Tracks)))
	sb.WriteString("Version=2\n")

	return sb.String()
}

// createWPL generates a Windows Media Player playlist.
//
// WPL is an XML-based SMIL format used by Windows Media Player.
func (p *PlaylistCreator) createWPL(album *model.Album) string {
	var sb strings.Builder

	sb.WriteString("<?wpl version=\"1.0\"?>\n")
	sb.WriteString("<smil>\n")
	sb.WriteString("  <head>\n")
	sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", escapeXML(album.Title)))
	sb.WriteString("  </head>\n")
	sb.WriteString("  <body>\n")
	sb.WriteString("    <seq>\n")

	for _, track := range album.Tracks {
		sb.WriteString(fmt.Sprintf("      <media src=\"%s\"/>\n", escapeXML(filepath.Base(track.Path))))
	}

	sb.WriteString("    </seq>\n")
	sb.WriteString("  </body>\n")
	sb.WriteString("</smil>\n")

	return sb.String()
}

// createZPL generates a Zune/Groove Music playlist.
//
// ZPL is similar to WPL but includes additional metadata attributes
// like album title, artist, and track duration.
func (p *PlaylistCreator) createZPL(album *model.Album) string {
	var sb strings.Builder

	sb.WriteString("<?zpl version=\"2.0\"?>\n")
	sb.WriteString("<smil>\n")
	sb.WriteString("  <head>\n")
	sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", escapeXML(album.Title)))
	sb.WriteString(fmt.Sprintf("    <meta name=\"Generator\" content=\"BandcampDownloader\"/>\n"))
	sb.WriteString(fmt.Sprintf("    <meta name=\"ItemCount\" content=\"%d\"/>\n", len(album.Tracks)))
	sb.WriteString("  </head>\n")
	sb.WriteString("  <body>\n")
	sb.WriteString("    <seq>\n")

	for _, track := range album.Tracks {
		duration := time.Duration(track.Duration * float64(time.Second))
		sb.WriteString(fmt.Sprintf("      <media src=\"%s\" albumTitle=\"%s\" albumArtist=\"%s\" trackTitle=\"%s\" trackArtist=\"%s\" duration=\"%d\"/>\n",
			escapeXML(filepath.Base(track.Path)),
			escapeXML(album.Title),
			escapeXML(album.Artist),
			escapeXML(track.Title),
			escapeXML(album.Artist),
			int(duration.Milliseconds())))
	}

	sb.WriteString("    </seq>\n")
	sb.WriteString("  </body>\n")
	sb.WriteString("</smil>\n")

	return sb.String()
}

// escapeXML escapes special XML characters in a string.
//
// Replaces: & < > " '
// With:     &amp; &lt; &gt; &quot; &apos;
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
