package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/handiism/bandcamp-downloader/internal/model"
)

const (
	artworkURLStart = "https://f4.bcbits.com/img/a"
	artworkURLEnd   = "_0.jpg"
)

// BandcampTime is a custom time type that handles Bandcamp's date format.
type BandcampTime struct {
	time.Time
}

// UnmarshalJSON parses Bandcamp's date format: "01 Jan 2023 00:00:00 GMT"
func (bt *BandcampTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		bt.Time = time.Time{}
		return nil
	}

	// Try multiple formats
	formats := []string{
		"02 Jan 2006 15:04:05 MST",  // "01 Jan 2023 00:00:00 GMT"
		"2 Jan 2006 15:04:05 MST",   // "1 Jan 2023 00:00:00 GMT"
		time.RFC3339,                // Standard format
		"2006-01-02T15:04:05Z07:00", // ISO format
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			bt.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse date: %s", s)
}

// JSONAlbum represents the deserialized album data from Bandcamp's HTML.
type JSONAlbum struct {
	AlbumData   *JSONAlbumData `json:"current"`
	ArtID       *int64         `json:"art_id"`
	Artist      string         `json:"artist"`
	ReleaseDate *BandcampTime  `json:"album_release_date"`
	Tracks      []JSONTrack    `json:"trackinfo"`
}

// JSONAlbumData contains album metadata.
type JSONAlbumData struct {
	AlbumTitle  string        `json:"title"`
	ReleaseDate *BandcampTime `json:"release_date"`
	PublishDate *BandcampTime `json:"publish_date"`
}

// ToAlbum converts JSONAlbum to a model.Album.
func (ja *JSONAlbum) ToAlbum(pathCfg *model.PathConfig, trackCfg *model.TrackConfig) *model.Album {
	// Build artwork URL
	var artworkURL string
	if ja.ArtID != nil {
		artworkURL = fmt.Sprintf("%s%010d%s", artworkURLStart, *ja.ArtID, artworkURLEnd)
	}

	// Determine release date with fallbacks
	var releaseDate time.Time
	if ja.ReleaseDate != nil {
		releaseDate = ja.ReleaseDate.Time
	} else if ja.AlbumData != nil && ja.AlbumData.ReleaseDate != nil {
		releaseDate = ja.AlbumData.ReleaseDate.Time
	} else if ja.AlbumData != nil && ja.AlbumData.PublishDate != nil {
		releaseDate = ja.AlbumData.PublishDate.Time
	}

	title := ""
	if ja.AlbumData != nil {
		title = ja.AlbumData.AlbumTitle
	}

	album := model.NewAlbum(ja.Artist, title, artworkURL, releaseDate, pathCfg)

	// Convert tracks (skip those without files)
	// TODO: Handle multiple discs. For now, always assume disc 1.
	discNumber := 1
	for _, jt := range ja.Tracks {
		if jt.File != nil {
			track := jt.ToTrack(album, discNumber, trackCfg)
			album.Tracks = append(album.Tracks, track)
		}
	}

	return album
}
