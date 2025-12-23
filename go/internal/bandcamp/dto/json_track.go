package dto

import (
	"strings"

	"github.com/handiism/bandcamp-downloader/internal/model"
)

// JSONTrack represents a track from Bandcamp's JSON data.
type JSONTrack struct {
	Duration float64      `json:"duration"`
	File     *JSONMp3File `json:"file"`
	Lyrics   string       `json:"lyrics"`
	Number   *int         `json:"track_num"`
	Title    string       `json:"title"`
}

// JSONMp3File represents the MP3 file info.
type JSONMp3File struct {
	URL string `json:"mp3-128"`
}

// ToTrack converts JSONTrack to a model.Track.
func (jt *JSONTrack) ToTrack(album *model.Album, cfg *model.TrackConfig) *model.Track {
	// Fix URL if it starts with "//"
	mp3URL := jt.File.URL
	if strings.HasPrefix(mp3URL, "//") {
		mp3URL = "http:" + mp3URL
	}

	// Default track number to 1 for single tracks
	number := 1
	if jt.Number != nil {
		number = *jt.Number
	}

	return model.NewTrack(album, number, jt.Title, jt.Duration, jt.Lyrics, mp3URL, cfg)
}
