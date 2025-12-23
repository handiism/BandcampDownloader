package download

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/handiism/bandcamp-downloader/internal/audio"
	"github.com/handiism/bandcamp-downloader/internal/bandcamp"
	"github.com/handiism/bandcamp-downloader/internal/config"
	"github.com/handiism/bandcamp-downloader/internal/http"
	ioutils "github.com/handiism/bandcamp-downloader/internal/io"
	"github.com/handiism/bandcamp-downloader/internal/model"
	"golang.org/x/sync/errgroup"
)

// ProgressLevel indicates the severity/type of a progress message.
type ProgressLevel int

const (
	LevelInfo ProgressLevel = iota
	LevelVerbose
	LevelWarning
	LevelError
	LevelSuccess
)

// ProgressEvent represents a download progress update.
type ProgressEvent struct {
	Message string
	Level   ProgressLevel
}

// Manager coordinates album downloads.
type Manager struct {
	settings     *config.Settings
	httpClient   *http.Client
	parser       *bandcamp.Parser
	discography  *bandcamp.Discography
	tagger       *audio.Tagger
	playlist     *audio.PlaylistCreator
	imageService *ioutils.ImageService

	albums          []*model.Album
	totalBytes      int64
	receivedBytes   int64
	totalFiles      int32
	downloadedFiles int32

	onProgress func(ProgressEvent)
	mu         sync.RWMutex
}

// NewManager creates a new download Manager.
func NewManager(settings *config.Settings, onProgress func(ProgressEvent)) *Manager {
	pathCfg := settings.ToPathConfig()
	trackCfg := settings.ToTrackConfig()

	var playlistFormat audio.PlaylistFormat
	switch settings.PlaylistFormat {
	case "pls":
		playlistFormat = audio.FormatPLS
	case "wpl":
		playlistFormat = audio.FormatWPL
	case "zpl":
		playlistFormat = audio.FormatZPL
	default:
		playlistFormat = audio.FormatM3U
	}

	return &Manager{
		settings:     settings,
		httpClient:   http.NewClient(),
		parser:       bandcamp.NewParser(pathCfg, trackCfg),
		discography:  bandcamp.NewDiscography(),
		tagger:       audio.NewTagger(audio.DefaultTagConfig()),
		playlist:     audio.NewPlaylistCreator(playlistFormat, settings.M3UExtended),
		imageService: ioutils.NewImageService(),
		onProgress:   onProgress,
	}
}

// Initialize fetches album info from the input URLs.
func (m *Manager) Initialize(ctx context.Context, inputURLs string) error {
	urls := m.parseInputURLs(inputURLs)

	var allAlbumURLs []string
	for _, inputURL := range urls {
		albumURLs, err := m.getAlbumURLs(ctx, inputURL)
		if err != nil {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Error getting albums from %s: %v", inputURL, err), Level: LevelError})
			continue
		}
		allAlbumURLs = append(allAlbumURLs, albumURLs...)
	}

	// Fetch album info
	for _, albumURL := range allAlbumURLs {
		m.progress(ProgressEvent{Message: fmt.Sprintf("Fetching album info: %s", albumURL), Level: LevelVerbose})

		html, err := m.httpClient.GetString(ctx, albumURL)
		if err != nil {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Error fetching %s: %v", albumURL, err), Level: LevelError})
			continue
		}

		album, err := m.parser.ParseAlbumPage(html)
		if err != nil {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Error parsing %s: %v", albumURL, err), Level: LevelError})
			continue
		}

		m.albums = append(m.albums, album)
		m.progress(ProgressEvent{Message: fmt.Sprintf("Found album: %s - %s (%d tracks)", album.Artist, album.Title, len(album.Tracks)), Level: LevelInfo})
	}

	// Calculate total bytes to download
	m.calculateTotals(ctx)

	return nil
}

// StartDownloads begins downloading all initialized albums.
func (m *Manager) StartDownloads(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(m.settings.MaxConcurrentAlbumsDownload)

	for _, album := range m.albums {
		album := album // capture
		g.Go(func() error {
			return m.downloadAlbum(ctx, album)
		})
	}

	return g.Wait()
}

// GetProgress returns current download progress.
func (m *Manager) GetProgress() (received, total int64, filesReceived, filesTotal int32) {
	return atomic.LoadInt64(&m.receivedBytes), m.totalBytes,
		atomic.LoadInt32(&m.downloadedFiles), m.totalFiles
}

// GetAlbumNames returns the names of all initialized albums.
func (m *Manager) GetAlbumNames() []string {
	names := make([]string, len(m.albums))
	for i, album := range m.albums {
		names[i] = fmt.Sprintf("%s - %s (%d tracks)", album.Artist, album.Title, len(album.Tracks))
	}
	return names
}

func (m *Manager) parseInputURLs(input string) []string {
	lines := strings.Split(input, "\n")
	var urls []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && (strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://")) {
			urls = append(urls, line)
		}
	}
	return urls
}

func (m *Manager) getAlbumURLs(ctx context.Context, inputURL string) ([]string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return nil, err
	}

	// Check if it's already an album/track URL
	if strings.Contains(parsedURL.Path, "/album/") || strings.Contains(parsedURL.Path, "/track/") {
		return []string{inputURL}, nil
	}

	// Fetch discography
	if !m.settings.DownloadArtistDiscography {
		return []string{inputURL}, nil
	}

	musicURL := fmt.Sprintf("%s://%s/music", parsedURL.Scheme, parsedURL.Host)
	html, err := m.httpClient.GetString(ctx, musicURL)
	if err != nil {
		return nil, err
	}

	relativeURLs, err := m.discography.GetAlbumURLs(html)
	if err != nil {
		return nil, err
	}

	var absoluteURLs []string
	for _, relURL := range relativeURLs {
		absoluteURLs = append(absoluteURLs, fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, relURL))
	}

	return absoluteURLs, nil
}

func (m *Manager) calculateTotals(ctx context.Context) {
	for _, album := range m.albums {
		for _, track := range album.Tracks {
			m.totalFiles++
			size, err := m.httpClient.GetFileSize(ctx, track.Mp3URL)
			if err == nil {
				m.totalBytes += size
			}
		}
		if album.HasArtwork() {
			m.totalFiles++
			size, err := m.httpClient.GetFileSize(ctx, album.ArtworkURL)
			if err == nil {
				m.totalBytes += size
			}
		}
	}
}

func (m *Manager) downloadAlbum(ctx context.Context, album *model.Album) error {
	// Create directory
	if err := os.MkdirAll(album.Path, 0755); err != nil {
		m.progress(ProgressEvent{Message: fmt.Sprintf("Error creating directory: %v", err), Level: LevelError})
		return err
	}

	var artwork []byte

	// Download artwork
	if (m.settings.SaveCoverArtInTags || m.settings.SaveCoverArtInFolder) && album.HasArtwork() {
		var err error
		artwork, err = m.downloadArtwork(ctx, album)
		if err != nil {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Error downloading artwork for %s: %v", album.Title, err), Level: LevelWarning})
		}
	}

	// Download tracks
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(m.settings.MaxConcurrentTracksDownload)

	var successCount int32
	for _, track := range album.Tracks {
		track := track // capture
		g.Go(func() error {
			if err := m.downloadTrack(ctx, track, album, artwork); err != nil {
				m.progress(ProgressEvent{Message: fmt.Sprintf("Error downloading %s: %v", track.Title, err), Level: LevelError})
				return nil // Continue with other tracks
			}
			atomic.AddInt32(&successCount, 1)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Create playlist
	if m.settings.CreatePlaylist {
		content := m.playlist.CreatePlaylist(album)
		if err := os.WriteFile(album.PlaylistPath, []byte(content), 0644); err != nil {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Error creating playlist: %v", err), Level: LevelWarning})
		} else {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Created playlist for %s", album.Title), Level: LevelSuccess})
		}
	}

	if int(successCount) == len(album.Tracks) {
		m.progress(ProgressEvent{Message: fmt.Sprintf("Successfully downloaded album: %s", album.Title), Level: LevelSuccess})
	} else {
		m.progress(ProgressEvent{Message: fmt.Sprintf("Finished %s, some tracks failed", album.Title), Level: LevelWarning})
	}

	return nil
}

func (m *Manager) downloadArtwork(ctx context.Context, album *model.Album) ([]byte, error) {
	var artwork []byte
	var err error

	for tries := 0; tries < m.settings.DownloadMaxRetries; tries++ {
		artwork, err = m.httpClient.DownloadBytes(ctx, album.ArtworkURL)
		if err == nil {
			break
		}
		m.waitForRetry(ctx, tries)
	}

	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&m.downloadedFiles, 1)

	// Save to folder if requested
	if m.settings.SaveCoverArtInFolder {
		artworkToSave := artwork

		if m.settings.CoverArtInFolderResize {
			artworkToSave, _ = m.imageService.ResizeImage(ctx, artworkToSave, m.settings.CoverArtInFolderMaxSize, m.settings.CoverArtInFolderMaxSize)
		}

		if m.settings.ConvertCoverArtToJPG {
			artworkToSave, _ = m.imageService.ConvertToJPEG(ctx, artworkToSave)
		}

		if err := os.WriteFile(album.ArtworkPath, artworkToSave, 0644); err != nil {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Error saving artwork: %v", err), Level: LevelWarning})
		}
	}

	// Prepare for tags
	if m.settings.SaveCoverArtInTags {
		if m.settings.CoverArtInTagsResize {
			artwork, _ = m.imageService.ResizeImage(ctx, artwork, m.settings.CoverArtInTagsMaxSize, m.settings.CoverArtInTagsMaxSize)
		}
		if m.settings.ConvertCoverArtToJPG {
			artwork, _ = m.imageService.ConvertToJPEG(ctx, artwork)
		}
	}

	m.progress(ProgressEvent{Message: fmt.Sprintf("Downloaded artwork for %s", album.Title), Level: LevelVerbose})
	return artwork, nil
}

func (m *Manager) downloadTrack(ctx context.Context, track *model.Track, album *model.Album, artwork []byte) error {
	// Check if file already exists with acceptable size
	if info, err := os.Stat(track.Path); err == nil {
		expectedSize, _ := m.httpClient.GetFileSize(ctx, track.Mp3URL)
		diff := m.settings.AllowedFileSizeDifference
		if expectedSize > 0 {
			sizeDiff := float64(info.Size()-expectedSize) / float64(expectedSize)
			if math.Abs(sizeDiff) <= diff {
				m.progress(ProgressEvent{Message: fmt.Sprintf("Skipping existing: %s", filepath.Base(track.Path)), Level: LevelVerbose})
				atomic.AddInt32(&m.downloadedFiles, 1)
				return nil
			}
		}
	}

	var err error
	for tries := 0; tries < m.settings.DownloadMaxRetries; tries++ {
		err = m.httpClient.DownloadFile(ctx, track.Mp3URL, track.Path, func(written, total int64) {
			// Progress tracking could be added here
		})
		if err == nil {
			break
		}
		m.progress(ProgressEvent{Message: fmt.Sprintf("Retry %d/%d for %s", tries+1, m.settings.DownloadMaxRetries, track.Title), Level: LevelWarning})
		m.waitForRetry(ctx, tries)
	}

	if err != nil {
		return err
	}

	atomic.AddInt32(&m.downloadedFiles, 1)

	// Tag the file
	if m.settings.ModifyTags || (m.settings.SaveCoverArtInTags && artwork != nil) {
		if err := m.tagger.SaveTags(track, album, artwork); err != nil {
			m.progress(ProgressEvent{Message: fmt.Sprintf("Error tagging %s: %v", track.Title, err), Level: LevelWarning})
		}
	}

	m.progress(ProgressEvent{Message: fmt.Sprintf("Downloaded: %s", filepath.Base(track.Path)), Level: LevelVerbose})
	return nil
}

func (m *Manager) waitForRetry(ctx context.Context, tries int) {
	cooldown := m.settings.DownloadRetryCooldown * math.Pow(m.settings.DownloadRetryExponent, float64(tries))
	select {
	case <-ctx.Done():
	case <-time.After(time.Duration(cooldown * float64(time.Second))):
	}
}

func (m *Manager) progress(event ProgressEvent) {
	if m.onProgress != nil {
		m.onProgress(event)
	}
}
