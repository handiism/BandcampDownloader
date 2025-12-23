package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/handiism/bandcamp-downloader/internal/config"
	"github.com/handiism/bandcamp-downloader/internal/download"
)

func main() {
	// Command line flags
	var (
		urlsFlag        = flag.String("url", "", "Bandcamp URL(s) to download (comma-separated or newline-separated)")
		outputFlag      = flag.String("output", "", "Output directory (overrides config)")
		configFlag      = flag.String("config", "", "Path to config file")
		discographyFlag = flag.Bool("discography", false, "Download entire artist discography")
		playlistFlag    = flag.Bool("playlist", false, "Create playlist file")
		verboseFlag     = flag.Bool("verbose", false, "Show verbose output")
		dryRunFlag      = flag.Bool("dry-run", false, "Parse URLs without downloading")
	)

	flag.Parse()

	// CLI mode - require URL
	if *urlsFlag == "" && flag.NArg() == 0 {
		fmt.Println("Bandcamp Downloader - Download music from Bandcamp")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  bandcamp-dl -url <URL> [options]")
		fmt.Println("  bandcamp-dl <URL> [options]")
		fmt.Println()
		fmt.Println("For interactive mode, use: bandcamp-tui")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load config
	settings := config.DefaultSettings()
	if *configFlag != "" {
		var err error
		settings, err = config.Load(*configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	// Apply flags
	if *outputFlag != "" {
		settings.DownloadsPath = *outputFlag + "/{artist}/{album}"
	}
	if *discographyFlag {
		settings.DownloadArtistDiscography = true
	}
	if *playlistFlag {
		settings.CreatePlaylist = true
	}

	// Get URLs
	urls := *urlsFlag
	if urls == "" && flag.NArg() > 0 {
		urls = flag.Arg(0)
	}

	// Handle interrupts
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, cancelling...")
		cancel()
	}()

	// Create manager with progress callback
	manager := download.NewManager(settings, func(event download.ProgressEvent) {
		if event.Level == download.LevelVerbose && !*verboseFlag {
			return
		}

		prefix := ""
		switch event.Level {
		case download.LevelError:
			prefix = "❌ "
		case download.LevelWarning:
			prefix = "⚠️  "
		case download.LevelSuccess:
			prefix = "✅ "
		case download.LevelInfo:
			prefix = "ℹ️  "
		default:
			prefix = "   "
		}

		fmt.Println(prefix + event.Message)
	})

	// Initialize
	fmt.Println("🎵 Bandcamp Downloader")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if err := manager.Initialize(ctx, urls); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing: %v\n", err)
		os.Exit(1)
	}

	if *dryRunFlag {
		fmt.Println("\n[Dry run - not downloading]")
		return
	}

	// Start downloads
	fmt.Println("\n📥 Starting downloads...")
	fmt.Println()

	if err := manager.StartDownloads(ctx); err != nil {
		if ctx.Err() != nil {
			fmt.Println("\nDownload cancelled.")
			os.Exit(130)
		}
		fmt.Fprintf(os.Stderr, "Error during download: %v\n", err)
		os.Exit(1)
	}

	received, total, filesReceived, filesTotal := manager.GetProgress()
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✨ Complete! Downloaded %d/%d files (%.2f MB)\n", filesReceived, filesTotal, float64(received)/1024/1024)
	if total > 0 && received < total {
		fmt.Printf("   (%.2f MB expected)\n", float64(total)/1024/1024)
	}
}
