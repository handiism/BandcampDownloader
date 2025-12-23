package bandcamp

import (
	"errors"
	"regexp"
	"strings"
)

// ErrNoAlbumFound is returned when no album or track URLs can be found on a page.
//
// This typically occurs when:
//   - The URL is not a valid Bandcamp artist/music page
//   - The artist has no published albums or tracks
//   - The HTML structure has changed unexpectedly
var ErrNoAlbumFound = errors.New("no album found on page")

// Discography extracts album and track URLs from Bandcamp artist pages.
//
// When given an artist's music page HTML (e.g., from https://artist.bandcamp.com/music),
// Discography can find all album and track URLs listed on that page.
//
// Discography handles two cases:
//  1. Normal music pages with multiple albums listed
//  2. Single-album artists where the music page redirects to the album page
//
// Example usage:
//
//	disco := NewDiscography()
//
//	resp, _ := http.Get("https://artist.bandcamp.com/music")
//	html, _ := io.ReadAll(resp.Body)
//
//	urls, err := disco.GetAlbumURLs(string(html))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, url := range urls {
//	    fullURL := "https://artist.bandcamp.com" + url
//	    fmt.Println(fullURL)
//	}
type Discography struct{}

// NewDiscography creates a new Discography service.
func NewDiscography() *Discography {
	return &Discography{}
}

// GetAlbumURLs extracts all album and track URLs from a Bandcamp music page.
//
// The returned URLs are relative paths like:
//   - /album/my-album
//   - /track/my-track
//
// These should be combined with the artist's base URL to form full URLs.
//
// The method handles two cases:
//  1. Normal music pages: Scans for all /album/ and /track/ links
//  2. Single-album artists: Detects redirect to album page and extracts that URL
//
// Duplicate URLs are automatically filtered out.
//
// Returns ErrNoAlbumFound if no album or track URLs can be found.
//
// Example:
//
//	urls, err := disco.GetAlbumURLs(musicPageHTML)
//	if errors.Is(err, ErrNoAlbumFound) {
//	    fmt.Println("Artist has no published music")
//	    return
//	}
func (d *Discography) GetAlbumURLs(musicPageHTML string) ([]string, error) {
	if d.isSingleAlbumArtist(musicPageHTML) {
		albumURL, err := d.getSingleAlbumURL(musicPageHTML)
		if err != nil {
			return nil, err
		}
		return []string{albumURL}, nil
	}

	// Match URLs like: /album/name" or /album/name&quot;
	re := regexp.MustCompile(`(?P<url>/(album|track)/.+?)("|&quot;)`)
	matches := re.FindAllStringSubmatch(musicPageHTML, -1)
	if len(matches) == 0 {
		return nil, ErrNoAlbumFound
	}

	// Collect unique URLs using a map
	urlSet := make(map[string]struct{})
	for _, match := range matches {
		if len(match) > 1 {
			urlSet[match[1]] = struct{}{}
		}
	}

	// Convert map keys to slice
	urls := make([]string, 0, len(urlSet))
	for url := range urlSet {
		urls = append(urls, url)
	}

	return urls, nil
}

// isSingleAlbumArtist checks if the page is an album page rather than a music listing.
//
// When an artist has only one album, Bandcamp often redirects their /music page
// to their album page. We detect this by looking for the "discography" div,
// which is only present on album pages, not music listing pages.
func (d *Discography) isSingleAlbumArtist(html string) bool {
	return strings.Contains(html, `div id="discography"`)
}

// getSingleAlbumURL extracts the album URL from a single-album artist's page.
//
// This is called when isSingleAlbumArtist returns true, indicating we're on
// an album page rather than a music listing. We extract the album URL from
// the page's links.
//
// Returns ErrNoAlbumFound if no album URL or multiple album URLs are found.
func (d *Discography) getSingleAlbumURL(html string) (string, error) {
	re := regexp.MustCompile(`href="(?P<url>/album/.+?)"`)
	matches := re.FindAllStringSubmatch(html, -1)

	// Collect unique URLs
	urlSet := make(map[string]struct{})
	for _, match := range matches {
		if len(match) > 1 {
			urlSet[match[1]] = struct{}{}
		}
	}

	switch len(urlSet) {
	case 0:
		return "", ErrNoAlbumFound
	case 1:
		for url := range urlSet {
			return url, nil
		}
	}

	return "", errors.New("found multiple album URLs, expected exactly one")
}
