package bandcamp

import (
	"testing"

	"github.com/handiism/bandcamp-downloader/internal/model"
)

func TestDiscography_GetAlbumURLs(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		wantCount   int
		wantErr     bool
		wantContain string
	}{
		{
			name: "single album link",
			html: `<html><body><a href="/album/test-album">Album</a></body></html>`,
			wantCount:   1,
			wantErr:     false,
			wantContain: "/album/test-album",
		},
		{
			name: "multiple albums",
			html: `<html><body>
				<a href="/album/first-album">&quot;</a>
				<a href="/album/second-album">&quot;</a>
				<a href="/track/single-track">&quot;</a>
			</body></html>`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "duplicate albums filtered",
			html: `<html><body>
				<a href="/album/same-album">&quot;</a>
				<a href="/album/same-album">&quot;</a>
			</body></html>`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "no albums found",
			html:      `<html><body>No music here</body></html>`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "single album artist page",
			html: `<html><body>
				<div id="discography"></div>
				<a href="/album/only-album">Only Album</a>
			</body></html>`,
			wantCount:   1,
			wantErr:     false,
			wantContain: "/album/only-album",
		},
	}

	d := NewDiscography()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, err := d.GetAlbumURLs(tt.html)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(urls) != tt.wantCount {
				t.Errorf("got %d URLs, want %d", len(urls), tt.wantCount)
			}

			if tt.wantContain != "" {
				found := false
				for _, url := range urls {
					if url == tt.wantContain {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find %q in %v", tt.wantContain, urls)
				}
			}
		})
	}
}

func TestParser_ParseAlbumPage(t *testing.T) {
	// Use inline mock HTML since test files are discography pages, not album pages
	mockHTML := `<html>
	<script data-tralbum="{
		&quot;current&quot;:{&quot;title&quot;:&quot;Test Album&quot;,&quot;release_date&quot;:&quot;01 Jan 2023 00:00:00 GMT&quot;},
		&quot;artist&quot;:&quot;Test Artist&quot;,
		&quot;art_id&quot;:1234567890,
		&quot;trackinfo&quot;:[
			{&quot;track_num&quot;:1,&quot;title&quot;:&quot;First Track&quot;,&quot;duration&quot;:180.5,&quot;file&quot;:{&quot;mp3-128&quot;:&quot;https://example.com/1.mp3&quot;}},
			{&quot;track_num&quot;:2,&quot;title&quot;:&quot;Second Track&quot;,&quot;duration&quot;:200.0,&quot;file&quot;:{&quot;mp3-128&quot;:&quot;https://example.com/2.mp3&quot;}}
		]
	}"></script>
	</html>`

	pathCfg := &model.PathConfig{
		DownloadsPath:         "/tmp/test/{artist}/{album}",
		CoverArtFileNameFormat: "{album}",
		PlaylistFileNameFormat: "{album}",
		PlaylistFormat:         model.PlaylistFormatM3U,
	}
	trackCfg := &model.TrackConfig{
		FileNameFormat: "{tracknum} {title}.mp3",
	}

	parser := NewParser(pathCfg, trackCfg)
	album, err := parser.ParseAlbumPage(mockHTML)
	if err != nil {
		t.Fatalf("ParseAlbumPage failed: %v", err)
	}

	// Validate parsed data
	if album.Artist != "Test Artist" {
		t.Errorf("Artist = %q, want %q", album.Artist, "Test Artist")
	}
	if album.Title != "Test Album" {
		t.Errorf("Title = %q, want %q", album.Title, "Test Album")
	}
	if len(album.Tracks) != 2 {
		t.Errorf("Track count = %d, want 2", len(album.Tracks))
	}
	if album.Tracks[0].Title != "First Track" {
		t.Errorf("Track[0].Title = %q, want %q", album.Tracks[0].Title, "First Track")
	}

	t.Logf("Parsed album: %s - %s (%d tracks)", album.Artist, album.Title, len(album.Tracks))
}

func TestExtractAlbumData(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr bool
	}{
		{
			name: "valid data-tralbum",
			html: `<html><script data-tralbum="{&quot;current&quot;:{&quot;title&quot;:&quot;Test&quot;}}"></script></html>`,
			wantErr: false,
		},
		{
			name:    "missing data-tralbum",
			html:    `<html><body>No album data</body></html>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractAlbumData(tt.html)
			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFixJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "fix URL concatenation",
			input: `url: "http://example.bandcamp.com" + "/album/test",`,
			want:  `url: "http://example.bandcamp.com/album/test",`,
		},
		{
			name:  "no change needed",
			input: `url: "http://example.bandcamp.com/album/test",`,
			want:  `url: "http://example.bandcamp.com/album/test",`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fixJSON(tt.input)
			if got != tt.want {
				t.Errorf("fixJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}
