package audio

import (
	"fmt"
	"os"

	"github.com/bogem/id3v2"
	"github.com/handiism/bandcamp-downloader/internal/model"
)

// TagEditAction defines how to handle individual ID3 tags.
//
// Each tag field can be configured independently to determine whether
// it should be modified, cleared, or left unchanged.
type TagEditAction int

const (
	// TagEmpty clears the tag value (sets to empty string).
	TagEmpty TagEditAction = iota

	// TagModify updates the tag with the value from Bandcamp.
	TagModify

	// TagDoNotModify leaves the existing tag value unchanged.
	TagDoNotModify
)

// TagConfig holds tagging configuration for each ID3 field.
//
// This allows fine-grained control over which tags are modified
// when processing downloaded MP3 files.
//
// Example:
//
//	cfg := &TagConfig{
//	    ModifyTags:  true,
//	    Artist:      TagModify,      // Update artist from Bandcamp
//	    Album:       TagModify,      // Update album from Bandcamp
//	    TrackTitle:  TagModify,      // Update title from Bandcamp
//	    Year:        TagModify,      // Update year from release date
//	    Lyrics:      TagModify,      // Add lyrics if available
//	    Comments:    TagEmpty,       // Clear any existing comments
//	    AlbumArtist: TagDoNotModify, // Keep existing album artist
//	}
type TagConfig struct {
	// ModifyTags is a master switch. If false, no string tags are modified.
	ModifyTags bool

	// Artist controls the TPE1 (Lead artist) frame.
	Artist TagEditAction

	// AlbumArtist controls the TPE2 (Album artist) frame.
	AlbumArtist TagEditAction

	// Album controls the TALB (Album title) frame.
	Album TagEditAction

	// Year controls the TYER (Year) frame.
	Year TagEditAction

	// Date controls the TDRC (Recording time) frame (ID3v2.4).
	Date TagEditAction

	// TrackNumber controls the TRCK (Track number) frame.
	TrackNumber TagEditAction

	// DiscNumber controls the TPOS (Part of a set) frame.
	DiscNumber TagEditAction

	// TrackTitle controls the TIT2 (Title) frame.
	TrackTitle TagEditAction

	// Lyrics controls the USLT (Unsynchronized lyrics) frame.
	Lyrics TagEditAction

	// Comments controls the COMM (Comments) frame.
	Comments TagEditAction
}

// DefaultTagConfig returns the default tag configuration.
//
// By default, all tags except comments are set to TagModify,
// which updates them with Bandcamp data. Comments are cleared.
func DefaultTagConfig() *TagConfig {
	return &TagConfig{
		ModifyTags:  true,
		Artist:      TagModify,
		AlbumArtist: TagModify,
		Album:       TagModify,
		Year:        TagModify,
		Date:        TagModify,
		TrackNumber: TagModify,
		DiscNumber:  TagModify,
		TrackTitle:  TagModify,
		Lyrics:      TagModify,
		Comments:    TagEmpty,
	}
}

// Tagger writes ID3 tags to MP3 files.
//
// Tagger uses the id3v2 library to modify MP3 file metadata including:
//   - Artist, Album Artist, Album, Title
//   - Track Number, Year
//   - Lyrics (unsynchronized)
//   - Cover Art (attached picture)
//
// Example:
//
//	tagger := NewTagger(DefaultTagConfig())
//
//	// After downloading track
//	err := tagger.SaveTags(track, album, artworkBytes)
//	if err != nil {
//	    log.Printf("Failed to tag %s: %v", track.Path, err)
//	}
type Tagger struct {
	config *TagConfig
}

// NewTagger creates a new Tagger with the given configuration.
//
// If config is nil, DefaultTagConfig() is used.
func NewTagger(config *TagConfig) *Tagger {
	if config == nil {
		config = DefaultTagConfig()
	}
	return &Tagger{config: config}
}

// SaveTags writes ID3 tags to the track's MP3 file.
//
// This method:
//  1. Opens the existing MP3 file (or creates empty tags if none exist)
//  2. Updates string tags based on TagConfig settings
//  3. Embeds cover art if artwork bytes are provided
//  4. Saves the modified tags to the file
//
// Parameters:
//   - track: The track being tagged (provides title, lyrics, file path)
//   - album: The album (provides artist, title, release date)
//   - artwork: JPEG image bytes for cover art (nil to skip artwork)
//
// Returns an error if the file cannot be opened or saved.
//
// Example:
//
//	tagger := NewTagger(DefaultTagConfig())
//	err := tagger.SaveTags(track, album, jpegBytes)
func (t *Tagger) SaveTags(track *model.Track, album *model.Album, artwork []byte) error {
	tag, err := id3v2.Open(track.Path, id3v2.Options{Parse: true})
	if err != nil {
		// If file doesn't have tags, create new
		if os.IsNotExist(err) {
			tag = id3v2.NewEmptyTag()
		} else {
			return err
		}
	}
	defer tag.Close()

	if t.config.ModifyTags {
		t.updateStringTags(tag, track, album)
	}

	if artwork != nil {
		t.updateArtwork(tag, artwork)
	}

	return tag.Save()
}

// updateStringTags updates text-based ID3 frames based on configuration.
func (t *Tagger) updateStringTags(tag *id3v2.Tag, track *model.Track, album *model.Album) {
	// Artist (TPE1)
	switch t.config.Artist {
	case TagEmpty:
		tag.SetArtist("")
	case TagModify:
		tag.SetArtist(album.Artist)
	}

	// Album (TALB)
	switch t.config.Album {
	case TagEmpty:
		tag.SetAlbum("")
	case TagModify:
		tag.SetAlbum(album.Title)
	}

	// Year (TYER) - ID3v2.3
	switch t.config.Year {
	case TagEmpty:
		tag.DeleteFrames("TYER")
	case TagModify:
		tag.AddTextFrame("TYER", id3v2.EncodingUTF8, album.ReleaseDate.Format("2006"))
	}

	// Date (TDRC) - ID3v2.4
	switch t.config.Date {
	case TagEmpty:
		tag.DeleteFrames("TDRC")
	case TagModify:
		tag.AddTextFrame("TDRC", id3v2.EncodingUTF8, album.ReleaseDate.Format("2006-01-02"))
	}

	// Track Number (TRCK)
	switch t.config.TrackNumber {
	case TagEmpty:
		tag.DeleteFrames("TRCK")
	case TagModify:
		tag.AddTextFrame("TRCK", id3v2.EncodingUTF8, fmt.Sprintf("%d", track.Number))
	}

	// Disc Number (TPOS)
	switch t.config.DiscNumber {
	case TagEmpty:
		tag.DeleteFrames("TPOS")
	case TagModify:
		if track.DiscNumber > 0 {
			tag.AddTextFrame("TPOS", id3v2.EncodingUTF8, fmt.Sprintf("%d", track.DiscNumber))
		}
	}

	// Track Title (TIT2)
	switch t.config.TrackTitle {
	case TagEmpty:
		tag.SetTitle("")
	case TagModify:
		tag.SetTitle(track.Title)
	}

	// Album Artist (TPE2)
	switch t.config.AlbumArtist {
	case TagEmpty:
		tag.DeleteFrames("TPE2")
	case TagModify:
		tag.AddTextFrame("TPE2", id3v2.EncodingUTF8, album.Artist)
	}

	// Lyrics (USLT)
	switch t.config.Lyrics {
	case TagEmpty:
		tag.DeleteFrames(tag.CommonID("Unsynchronised lyrics/text transcription"))
	case TagModify:
		if track.Lyrics != "" {
			uslf := id3v2.UnsynchronisedLyricsFrame{
				Encoding:          id3v2.EncodingUTF8,
				Language:          "eng",
				ContentDescriptor: "",
				Lyrics:            track.Lyrics,
			}
			tag.AddUnsynchronisedLyricsFrame(uslf)
		}
	}

	// Genre - always clear as Bandcamp doesn't provide genre info
	tag.SetGenre("")
}

// updateArtwork embeds cover art as an attached picture frame.
func (t *Tagger) updateArtwork(tag *id3v2.Tag, artwork []byte) {
	// Remove any existing cover pictures
	tag.DeleteFrames(tag.CommonID("Attached picture"))

	// Add new artwork as front cover (APIC frame)
	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     artwork,
	}
	tag.AddAttachedPicture(pic)
}
