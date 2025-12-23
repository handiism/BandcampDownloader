package bandcamp

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/handiism/bandcamp-downloader/internal/bandcamp/dto"
	"github.com/handiism/bandcamp-downloader/internal/model"
)

// Parser extracts album information from Bandcamp HTML pages.
//
// Bandcamp embeds album data as JSON within the HTML page in a data-tralbum
// attribute. The Parser extracts this JSON, fixes any malformed content,
// and deserializes it into an Album model.
//
// Example usage:
//
//	parser := NewParser(pathConfig, trackConfig)
//	
//	resp, _ := http.Get("https://artist.bandcamp.com/album/name")
//	html, _ := io.ReadAll(resp.Body)
//	
//	album, err := parser.ParseAlbumPage(string(html))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	fmt.Printf("Album: %s by %s\n", album.Title, album.Artist)
//	for _, track := range album.Tracks {
//	    fmt.Printf("  %d. %s\n", track.Number, track.Title)
//	}
type Parser struct {
	pathConfig  *model.PathConfig
	trackConfig *model.TrackConfig
}

// NewParser creates a new Parser with the given configuration.
//
// The pathConfig and trackConfig are used to compute file paths for the
// parsed albums and tracks. These configs determine where files will be
// saved and how they will be named.
//
// Parameters:
//   - pathCfg: Configuration for album folder paths, cover art, and playlists
//   - trackCfg: Configuration for track file naming
func NewParser(pathCfg *model.PathConfig, trackCfg *model.TrackConfig) *Parser {
	return &Parser{
		pathConfig:  pathCfg,
		trackConfig: trackCfg,
	}
}

// ParseAlbumPage extracts album info from a Bandcamp album or track page HTML.
//
// This method performs the following steps:
//  1. Extracts the data-tralbum JSON from the HTML
//  2. Fixes malformed JSON (e.g., URL concatenation issues)
//  3. Deserializes JSON into album/track data
//  4. Extracts lyrics from HTML elements (if available)
//  5. Computes file paths based on configuration
//
// The HTML should be the full page source from a Bandcamp URL like:
//   - https://artist.bandcamp.com/album/album-name
//   - https://artist.bandcamp.com/track/track-name
//
// Returns an error if:
//   - The data-tralbum attribute cannot be found
//   - The JSON is malformed and cannot be parsed
//
// Example:
//
//	album, err := parser.ParseAlbumPage(htmlContent)
//	if err != nil {
//	    return fmt.Errorf("failed to parse album: %w", err)
//	}
func (p *Parser) ParseAlbumPage(htmlContent string) (*model.Album, error) {
	// Extract the data-tralbum JSON
	albumData, err := extractAlbumData(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve album data: %w", err)
	}

	// Fix malformed JSON
	albumData = fixJSON(albumData)

	// Deserialize JSON
	var jsonAlbum dto.JSONAlbum
	if err := json.Unmarshal([]byte(albumData), &jsonAlbum); err != nil {
		return nil, fmt.Errorf("failed to parse album JSON: %w", err)
	}

	album := jsonAlbum.ToAlbum(p.pathConfig, p.trackConfig)

	// Extract lyrics from HTML
	p.extractLyrics(htmlContent, album)

	return album, nil
}

// extractAlbumData extracts the data-tralbum JSON string from HTML.
//
// Bandcamp embeds album data in the HTML like this:
//
//	<script ... data-tralbum="{...JSON...}">
//
// This function finds and extracts that JSON, then HTML-unescapes it
// (since the JSON is embedded in an HTML attribute, characters like
// quotes are escaped as &quot;).
func extractAlbumData(htmlContent string) (string, error) {
	const startString = `data-tralbum="{`
	const stopString = `}"`

	startIndex := strings.Index(htmlContent, startString)
	if startIndex == -1 {
		return "", fmt.Errorf("could not find album data in HTML")
	}

	startIndex += len(startString) - 1 // Include the opening brace
	remaining := htmlContent[startIndex:]

	endIndex := strings.Index(remaining, stopString)
	if endIndex == -1 {
		return "", fmt.Errorf("could not find end of album data")
	}

	albumData := remaining[:endIndex+1]
	return html.UnescapeString(albumData), nil
}

// fixJSON fixes malformed JSON from Bandcamp pages.
//
// Some Bandcamp pages have JavaScript-style URL concatenation in the JSON:
//
//	url: "http://example.bandcamp.com" + "/album/name",
//
// This is not valid JSON, so we fix it by removing the concatenation:
//
//	url: "http://example.bandcamp.com/album/name",
func fixJSON(albumData string) string {
	// Fix: url: "http://..." + "/album/..."
	// Remove the " + " concatenation
	re := regexp.MustCompile(`(url: ".+)" \+ "(.+",)`)
	return re.ReplaceAllString(albumData, "${1}${2}")
}

// extractLyrics extracts lyrics from the HTML and updates track Lyrics fields.
//
// Bandcamp displays lyrics in elements with IDs like "lyrics_row_1", "lyrics_row_2", etc.
// This method finds these elements and extracts the text content, stripping HTML tags.
func (p *Parser) extractLyrics(htmlContent string, album *model.Album) {
	for _, track := range album.Tracks {
		lyricsID := fmt.Sprintf(`id="lyrics_row_%d"`, track.Number)
		startIdx := strings.Index(htmlContent, lyricsID)
		if startIdx == -1 {
			continue
		}

		// Find the lyrics content within the element
		remaining := htmlContent[startIdx:]
		
		// Look for the lyrics text between tags
		contentStart := strings.Index(remaining, ">")
		if contentStart == -1 {
			continue
		}

		// Simple extraction - find text content
		contentEnd := strings.Index(remaining[contentStart:], "</div>")
		if contentEnd == -1 {
			continue
		}

		lyricsHTML := remaining[contentStart+1 : contentStart+contentEnd]
		// Strip HTML tags and clean up
		tagRegex := regexp.MustCompile(`<[^>]*>`)
		lyrics := tagRegex.ReplaceAllString(lyricsHTML, "")
		track.Lyrics = strings.TrimSpace(html.UnescapeString(lyrics))
	}
}
