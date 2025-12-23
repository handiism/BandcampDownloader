package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/handiism/bandcamp-downloader/internal/model"
)

// Settings holds all configuration options.
type Settings struct {
	// Download settings
	DownloadsPath               string  `json:"downloads_path"`
	MaxConcurrentAlbumsDownload int     `json:"max_concurrent_albums"`
	MaxConcurrentTracksDownload int     `json:"max_concurrent_tracks"`
	DownloadMaxRetries          int     `json:"download_max_retries"`
	DownloadRetryCooldown       float64 `json:"download_retry_cooldown"`
	DownloadRetryExponent       float64 `json:"download_retry_exponent"`
	AllowedFileSizeDifference   float64 `json:"allowed_file_size_difference"`
	DownloadArtistDiscography   bool    `json:"download_artist_discography"`

	// File naming
	FileNameFormat         string `json:"file_name_format"`
	CoverArtFileNameFormat string `json:"cover_art_file_name_format"`
	PlaylistFileNameFormat string `json:"playlist_file_name_format"`

	// Cover art settings
	SaveCoverArtInFolder    bool `json:"save_cover_art_in_folder"`
	SaveCoverArtInTags      bool `json:"save_cover_art_in_tags"`
	CoverArtInFolderResize  bool `json:"cover_art_in_folder_resize"`
	CoverArtInFolderMaxSize int  `json:"cover_art_in_folder_max_size"`
	CoverArtInTagsResize    bool `json:"cover_art_in_tags_resize"`
	CoverArtInTagsMaxSize   int  `json:"cover_art_in_tags_max_size"`
	ConvertCoverArtToJPG    bool `json:"convert_cover_art_to_jpg"`

	// Playlist settings
	CreatePlaylist bool   `json:"create_playlist"`
	PlaylistFormat string `json:"playlist_format"` // m3u, pls, wpl, zpl
	M3UExtended    bool   `json:"m3u_extended"`

	// Tag settings
	ModifyTags bool `json:"modify_tags"`

	// Proxy settings
	ProxyType    string `json:"proxy_type"` // none, system, manual
	ProxyAddress string `json:"proxy_address"`
	ProxyPort    int    `json:"proxy_port"`
}

// DefaultSettings returns settings with default values.
func DefaultSettings() *Settings {
	homeDir, _ := os.UserHomeDir()
	return &Settings{
		DownloadsPath:               filepath.Join(homeDir, "Music", "Bandcamp", "{artist}", "{album}"),
		MaxConcurrentAlbumsDownload: 1,
		MaxConcurrentTracksDownload: 10,
		DownloadMaxRetries:          7,
		DownloadRetryCooldown:       0.2,
		DownloadRetryExponent:       4.0,
		AllowedFileSizeDifference:   0.05,
		DownloadArtistDiscography:   false,

		FileNameFormat:         "{tracknum} {artist} - {title}.mp3",
		CoverArtFileNameFormat: "{album}",
		PlaylistFileNameFormat: "{album}",

		SaveCoverArtInFolder:    false,
		SaveCoverArtInTags:      true,
		CoverArtInFolderResize:  false,
		CoverArtInFolderMaxSize: 1000,
		CoverArtInTagsResize:    true,
		CoverArtInTagsMaxSize:   1000,
		ConvertCoverArtToJPG:    true,

		CreatePlaylist: false,
		PlaylistFormat: "m3u",
		M3UExtended:    true,

		ModifyTags: true,

		ProxyType: "system",
	}
}

// Load reads settings from a JSON file.
func Load(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return nil, err
	}

	settings := DefaultSettings()
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// Save writes settings to a JSON file.
func (s *Settings) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ToPathConfig converts settings to PathConfig.
func (s *Settings) ToPathConfig() *model.PathConfig {
	var pf model.PlaylistFormat
	switch s.PlaylistFormat {
	case "m3u":
		pf = model.PlaylistFormatM3U
	case "pls":
		pf = model.PlaylistFormatPLS
	case "wpl":
		pf = model.PlaylistFormatWPL
	case "zpl":
		pf = model.PlaylistFormatZPL
	default:
		pf = model.PlaylistFormatM3U
	}

	return &model.PathConfig{
		DownloadsPath:          s.DownloadsPath,
		CoverArtFileNameFormat: s.CoverArtFileNameFormat,
		PlaylistFileNameFormat: s.PlaylistFileNameFormat,
		PlaylistFormat:         pf,
	}
}

// ToTrackConfig converts settings to TrackConfig.
func (s *Settings) ToTrackConfig() *model.TrackConfig {
	return &model.TrackConfig{
		FileNameFormat: s.FileNameFormat,
	}
}
