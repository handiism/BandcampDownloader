// Package audio provides audio file manipulation services including
// ID3 tag writing and playlist generation.
//
// # ID3 Tagging
//
// Use the Tagger to write ID3 tags to MP3 files:
//
//	tagger := audio.NewTagger(audio.DefaultTagConfig())
//	err := tagger.SaveTags(track, album, artworkBytes)
//
// The tagger supports:
//   - Artist, Album Artist
//   - Album Title, Track Title
//   - Track Number, Year
//   - Lyrics
//   - Cover Art (embedded in MP3)
//
// # Playlist Generation
//
// Generate playlists in various formats:
//
//	creator := audio.NewPlaylistCreator(audio.FormatM3U, true) // extended M3U
//	content := creator.CreatePlaylist(album)
//	os.WriteFile("playlist.m3u", []byte(content), 0644)
//
// Supported formats:
//   - M3U (with optional extended info)
//   - PLS
//   - WPL (Windows Media Player)
//   - ZPL (Zune Media Player)
package audio
