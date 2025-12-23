// Package model defines the core data structures used throughout
// the bandcamp-downloader application.
//
// # Album
//
// Album represents a Bandcamp album with metadata and computed file paths:
//
//	album := model.NewAlbum("Artist", "Title", artworkURL, releaseDate, pathConfig)
//	fmt.Println(album.Path)        // Where to save the album
//	fmt.Println(album.ArtworkPath) // Where to save cover art
//
// # Track
//
// Track represents a single track within an album:
//
//	track := model.NewTrack(album, 1, "Song Title", 180.5, "", mp3URL, trackConfig)
//	fmt.Println(track.Path) // Full path where track will be saved
//
// # Path Configuration
//
// PathConfig controls how album/track paths are computed using placeholders:
//
//	cfg := &model.PathConfig{
//	    DownloadsPath:         "/music/{artist}/{album}",
//	    CoverArtFileNameFormat: "{album}",
//	    PlaylistFileNameFormat: "{album}",
//	    PlaylistFormat:         model.PlaylistFormatM3U,
//	}
//
// Available placeholders: {artist}, {album}, {title}, {tracknum}, {year}, {month}, {day}
package model
