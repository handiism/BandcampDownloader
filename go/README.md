# Bandcamp Downloader (Go)

A command-line tool for downloading music from Bandcamp. This is a Go port of the non-visual components from the [BandcampDownloader](https://github.com/Otiel/BandcampDownloader) C# application.

## Features

- ✅ Download albums from Bandcamp
- ✅ Download single tracks
- ✅ Download entire artist discography
- ✅ Automatic ID3 tag writing (artist, album, title, track number, year)
- ✅ Cover art embedding in MP3 files
- ✅ Cover art saved to folder
- ✅ Playlist generation (M3U, PLS, WPL, ZPL)
- ✅ Concurrent downloads with configurable limits
- ✅ Retry logic with exponential backoff
- ✅ Skip existing files

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/handiism/BandcampDownloader.git
cd BandcampDownloader/go

# Build
go build -o bandcamp-dl ./cmd/bandcamp-dl

# Or install globally
go install ./cmd/bandcamp-dl
```

### Requirements

- Go 1.21 or later

## Usage

### Basic Usage

```bash
# Download an album
./bandcamp-dl -url "https://artist.bandcamp.com/album/album-name"

# Download a single track
./bandcamp-dl -url "https://artist.bandcamp.com/track/track-name"

# Download an artist's full discography
./bandcamp-dl -url "https://artist.bandcamp.com" -discography
```

### Supported URL Formats

| Type   | Format                                        |
| ------ | --------------------------------------------- |
| Album  | `https://[artist].bandcamp.com/album/[album]` |
| Track  | `https://[artist].bandcamp.com/track/[track]` |
| Artist | `https://[artist].bandcamp.com`               |

### Command Line Options

| Flag           | Description                         | Default                             |
| -------------- | ----------------------------------- | ----------------------------------- |
| `-url`         | Bandcamp URL(s) to download         | (required)                          |
| `-output`      | Output directory                    | `~/Music/Bandcamp/{artist}/{album}` |
| `-config`      | Path to config file                 | -                                   |
| `-discography` | Download entire artist discography  | `false`                             |
| `-playlist`    | Create playlist file for each album | `false`                             |
| `-verbose`     | Show verbose output                 | `false`                             |
| `-dry-run`     | Parse URLs without downloading      | `false`                             |

### Examples

```bash
# Download to specific directory
./bandcamp-dl -url "https://artist.bandcamp.com/album/name" -output ./downloads

# Download with playlist
./bandcamp-dl -url "https://artist.bandcamp.com/album/name" -playlist

# Verbose output
./bandcamp-dl -url "https://artist.bandcamp.com/album/name" -verbose

# Dry run (preview without downloading)
./bandcamp-dl -url "https://artist.bandcamp.com/album/name" -dry-run
```

## Configuration

Create a JSON config file to customize settings:

```json
{
  "downloads_path": "/home/user/Music/Bandcamp/{artist}/{album}",
  "max_concurrent_albums": 1,
  "max_concurrent_tracks": 10,
  "file_name_format": "{tracknum} {artist} - {title}.mp3",
  "cover_art_file_name_format": "{album}",
  "save_cover_art_in_folder": true,
  "save_cover_art_in_tags": true,
  "create_playlist": false,
  "playlist_format": "m3u",
  "modify_tags": true
}
```

Use with: `./bandcamp-dl -url "..." -config ./config.json`

### Path Placeholders

Available placeholders for path/filename formats:

| Placeholder  | Description                |
| ------------ | -------------------------- |
| `{artist}`   | Artist name                |
| `{album}`    | Album title                |
| `{title}`    | Track title                |
| `{tracknum}` | Track number (zero-padded) |
| `{year}`     | Release year               |
| `{month}`    | Release month              |
| `{day}`      | Release day                |

## Project Structure

```
go/
├── cmd/
│   └── bandcamp-dl/
│       └── main.go           # CLI entry point
├── internal/
│   ├── model/
│   │   ├── album.go          # Album model with path computation
│   │   └── track.go          # Track model
│   ├── bandcamp/
│   │   ├── parser.go         # HTML parsing for album data
│   │   ├── discography.go    # Artist discography extraction
│   │   └── dto/              # JSON deserialization structs
│   ├── download/
│   │   └── manager.go        # Download orchestration
│   ├── audio/
│   │   ├── tagger.go         # ID3 tag writing
│   │   └── playlist.go       # Playlist generation
│   ├── http/
│   │   └── client.go         # HTTP client with progress
│   ├── io/
│   │   ├── file.go           # File utilities
│   │   └── image.go          # Image processing
│   └── config/
│       └── settings.go       # Configuration management
├── go.mod
└── go.sum
```

## Dependencies

- [`github.com/bogem/id3v2`](https://github.com/bogem/id3v2) - ID3 tag reading/writing
- [`golang.org/x/sync`](https://pkg.go.dev/golang.org/x/sync) - Concurrent goroutine management
- [`golang.org/x/image`](https://pkg.go.dev/golang.org/x/image) - Image processing

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific package tests
go test ./internal/bandcamp/... -v
```

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file.

## Acknowledgments

- Original [BandcampDownloader](https://github.com/Otiel/BandcampDownloader) by Otiel
- Bandcamp for providing amazing music from independent artists
