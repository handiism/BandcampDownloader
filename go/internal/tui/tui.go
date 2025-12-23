// Package tui provides a Bubble Tea terminal user interface for bandcamp-downloader.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handiism/bandcamp-downloader/internal/config"
	"github.com/handiism/bandcamp-downloader/internal/download"
)

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95E1A3"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFE66D"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8DADC"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C757D"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4ECDC4")).
			Padding(1, 2)

	albumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8B500"))
)

// State represents the current UI state.
type State int

const (
	StateInput State = iota
	StateInitializing
	StateDownloading
	StateComplete
	StateError
)

// LogEntry represents a log message in the UI.
type LogEntry struct {
	Message string
	Level   download.ProgressLevel
}

// Model is the Bubble Tea model for the TUI.
type Model struct {
	state     State
	textInput textinput.Model
	spinner   spinner.Model
	progress  progress.Model
	settings  *config.Settings
	logs      []LogEntry
	albums    []string
	err       error

	// Download context
	ctx    context.Context
	cancel context.CancelFunc

	// Download manager reference
	manager *download.Manager

	// Download progress
	totalFiles      int32
	downloadedFiles int32
	totalBytes      int64
	receivedBytes   int64

	// Options
	discography bool
	playlist    bool
	verbose     bool

	width  int
	height int
}

// NewModel creates a new TUI model.
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "https://artist.bandcamp.com/album/name"
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 50

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		state:     StateInput,
		textInput: ti,
		spinner:   sp,
		progress:  prog,
		settings:  config.DefaultSettings(),
		logs:      make([]LogEntry, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

// Message types
type (
	// ProgressMsg is sent when download progress updates.
	ProgressMsg struct {
		Event download.ProgressEvent
	}

	// InitDoneMsg is sent when initialization completes.
	InitDoneMsg struct {
		Albums  []string
		Manager *download.Manager
		Err     error
	}

	// DownloadStartMsg triggers the actual download after init.
	DownloadStartMsg struct{}

	// DownloadDoneMsg is sent when all downloads complete.
	DownloadDoneMsg struct {
		Received int64
		Total    int64
		Files    int32
		TotalF   int32
		Err      error
	}

	// TickMsg is for periodic progress updates.
	TickMsg struct{}
)

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		if m.progress.Width < 20 {
			m.progress.Width = 20
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancel()
			return m, tea.Quit

		case "esc":
			if m.state == StateInput {
				return m, tea.Quit
			}
			if m.state == StateDownloading || m.state == StateInitializing {
				m.cancel()
				m.state = StateError
				m.err = fmt.Errorf("cancelled by user")
			}

		case "enter":
			if m.state == StateInput && m.textInput.Value() != "" {
				m.state = StateInitializing
				return m, tea.Batch(m.initializeDownload(), m.spinner.Tick)
			}

		case "d":
			if m.state == StateInput {
				m.discography = !m.discography
			}

		case "p":
			if m.state == StateInput {
				m.playlist = !m.playlist
			}

		case "v":
			if m.state == StateInput {
				m.verbose = !m.verbose
			}

		case "q":
			if m.state == StateComplete || m.state == StateError {
				return m, tea.Quit
			}

		case "r":
			if m.state == StateComplete || m.state == StateError {
				// Reset for new download
				m.state = StateInput
				m.logs = nil
				m.albums = nil
				m.err = nil
				m.downloadedFiles = 0
				m.totalFiles = 0
				m.receivedBytes = 0
				m.totalBytes = 0
				m.manager = nil
				m.ctx, m.cancel = context.WithCancel(context.Background())
				m.textInput.SetValue("")
				m.textInput.Focus()
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case ProgressMsg:
		// Filter verbose messages if not in verbose mode
		if msg.Event.Level == download.LevelVerbose && !m.verbose {
			return m, nil
		}
		m.logs = append(m.logs, LogEntry{
			Message: msg.Event.Message,
			Level:   msg.Event.Level,
		})
		// Keep only last 10 logs
		if len(m.logs) > 10 {
			m.logs = m.logs[len(m.logs)-10:]
		}

	case InitDoneMsg:
		if msg.Err != nil {
			m.state = StateError
			m.err = msg.Err
		} else {
			m.albums = msg.Albums
			m.manager = msg.Manager
			m.state = StateDownloading
			// Start the actual download and tick for progress updates
			cmds = append(cmds, m.startDownload(), m.tickProgress())
		}

	case DownloadDoneMsg:
		m.receivedBytes = msg.Received
		m.totalBytes = msg.Total
		m.downloadedFiles = msg.Files
		m.totalFiles = msg.TotalF
		if msg.Err != nil && m.ctx.Err() == nil {
			m.state = StateError
			m.err = msg.Err
		} else if m.ctx.Err() != nil {
			m.state = StateError
			m.err = fmt.Errorf("cancelled by user")
		} else {
			m.state = StateComplete
		}

	case TickMsg:
		// Update progress from manager
		if m.manager != nil && m.state == StateDownloading {
			received, total, files, totalFiles := m.manager.GetProgress()
			m.receivedBytes = received
			m.totalBytes = total
			m.downloadedFiles = files
			m.totalFiles = totalFiles

			// Calculate percentage and animate progress bar
			var percent float64
			if totalFiles > 0 {
				percent = float64(files) / float64(totalFiles)
			}
			progressCmd := m.progress.SetPercent(percent)
			cmds = append(cmds, progressCmd, m.tickProgress())
		}

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	// Update text input
	if m.state == StateInput {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// tickProgress returns a command to tick progress updates.
func (m Model) tickProgress() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}

// View renders the UI.
func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("🎵 Bandcamp Downloader"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Download music from Bandcamp"))
	b.WriteString("\n\n")

	switch m.state {
	case StateInput:
		b.WriteString(m.viewInput())
	case StateInitializing:
		b.WriteString(m.viewInitializing())
	case StateDownloading:
		b.WriteString(m.viewDownloading())
	case StateComplete:
		b.WriteString(m.viewComplete())
	case StateError:
		b.WriteString(m.viewError())
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(m.getHelpText()))

	return b.String()
}

func (m Model) viewInput() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("Enter Bandcamp URL:"))
	b.WriteString("\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Options
	discographyCheck := "[ ]"
	if m.discography {
		discographyCheck = "[×]"
	}
	playlistCheck := "[ ]"
	if m.playlist {
		playlistCheck = "[×]"
	}
	verboseCheck := "[ ]"
	if m.verbose {
		verboseCheck = "[×]"
	}

	b.WriteString(infoStyle.Render("Options:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s Download discography (d)\n", discographyCheck))
	b.WriteString(fmt.Sprintf("  %s Create playlist (p)\n", playlistCheck))
	b.WriteString(fmt.Sprintf("  %s Verbose/debug output (v)\n", verboseCheck))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("Download path: %s", m.settings.DownloadsPath)))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewInitializing() string {
	var b strings.Builder

	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(subtitleStyle.Render("Fetching album info..."))
	b.WriteString("\n\n")

	// Show logs
	b.WriteString(m.renderLogs())

	return b.String()
}

func (m Model) viewDownloading() string {
	var b strings.Builder

	// Albums found
	if len(m.albums) > 0 {
		b.WriteString(successStyle.Render(fmt.Sprintf("Found %d album(s):", len(m.albums))))
		b.WriteString("\n")
		for _, album := range m.albums {
			b.WriteString(albumStyle.Render(fmt.Sprintf("  ♪ %s", album)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Progress bar
	var percent float64
	if m.totalFiles > 0 {
		percent = float64(m.downloadedFiles) / float64(m.totalFiles)
	}
	b.WriteString(m.progress.ViewAs(percent))
	b.WriteString("\n")

	b.WriteString(infoStyle.Render(fmt.Sprintf(
		"Files: %d/%d | Downloaded: %.2f MB",
		m.downloadedFiles,
		m.totalFiles,
		float64(m.receivedBytes)/1024/1024,
	)))
	b.WriteString("\n\n")

	// Logs
	b.WriteString(m.renderLogs())

	return b.String()
}

func (m Model) viewComplete() string {
	var b strings.Builder

	box := boxStyle.Render(fmt.Sprintf(
		"✨ Download Complete!\n\n"+
			"Albums: %d\n"+
			"Files: %d\n"+
			"Size: %.2f MB",
		len(m.albums),
		m.downloadedFiles,
		float64(m.receivedBytes)/1024/1024,
	))
	b.WriteString(box)

	return b.String()
}

func (m Model) viewError() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render("❌ Error occurred:"))
	b.WriteString("\n\n")
	if m.err != nil {
		b.WriteString(fmt.Sprintf("  %s", m.err.Error()))
	}

	return b.String()
}

func (m Model) renderLogs() string {
	var b strings.Builder

	for _, log := range m.logs {
		var style lipgloss.Style
		prefix := "•"
		switch log.Level {
		case download.LevelError:
			style = errorStyle
			prefix = "✗"
		case download.LevelWarning:
			style = warningStyle
			prefix = "!"
		case download.LevelSuccess:
			style = successStyle
			prefix = "✓"
		case download.LevelInfo:
			style = infoStyle
			prefix = "›"
		default:
			style = dimStyle
		}
		b.WriteString(style.Render(prefix + " " + log.Message))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) getHelpText() string {
	switch m.state {
	case StateInput:
		return "enter: start • d: discography • p: playlist • v: verbose • esc: quit"
	case StateInitializing, StateDownloading:
		return "esc: cancel"
	case StateComplete, StateError:
		return "r: new download • q: quit"
	}
	return ""
}

// initializeDownload fetches album info and creates the manager.
func (m *Model) initializeDownload() tea.Cmd {
	return func() tea.Msg {
		url := m.textInput.Value()

		// Apply options
		settings := config.DefaultSettings()
		if m.discography {
			settings.DownloadArtistDiscography = true
		}
		if m.playlist {
			settings.CreatePlaylist = true
		}

		var albumNames []string

		// Create manager with progress callback
		manager := download.NewManager(settings, func(event download.ProgressEvent) {
			// Progress events are collected but not sent directly
			// The TUI polls for progress via TickMsg
		})

		// Initialize - this fetches album info
		if err := manager.Initialize(m.ctx, url); err != nil {
			return InitDoneMsg{Err: err}
		}

		// Get album info for display
		albumNames = manager.GetAlbumNames()

		return InitDoneMsg{
			Albums:  albumNames,
			Manager: manager,
			Err:     nil,
		}
	}
}

// startDownload starts the actual download in background.
func (m *Model) startDownload() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return DownloadDoneMsg{Err: fmt.Errorf("no manager")}
		}

		err := m.manager.StartDownloads(m.ctx)
		received, total, files, totalFiles := m.manager.GetProgress()

		return DownloadDoneMsg{
			Received: received,
			Total:    total,
			Files:    files,
			TotalF:   totalFiles,
			Err:      err,
		}
	}
}

// Run starts the TUI application.
func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
