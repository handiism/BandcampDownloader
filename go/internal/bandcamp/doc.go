// Package bandcamp provides functionality to parse Bandcamp HTML pages
// and extract album/track information.
//
// The package handles two main use cases:
//
//  1. Parsing album/track pages to extract metadata and download URLs
//  2. Parsing artist discography pages to discover all albums
//
// # Album Page Parsing
//
// Use the Parser to extract album information from a Bandcamp album page:
//
//	parser := bandcamp.NewParser(pathConfig, trackConfig)
//	album, err := parser.ParseAlbumPage(htmlContent)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Album: %s by %s\n", album.Title, album.Artist)
//
// # Discography Extraction
//
// Use Discography to find all album URLs from an artist's music page:
//
//	disco := bandcamp.NewDiscography()
//	urls, err := disco.GetAlbumURLs(musicPageHTML)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, url := range urls {
//	    fmt.Println(url) // e.g., "/album/my-album"
//	}
//
// # Bandcamp Data Format
//
// Bandcamp embeds album data as JSON in the HTML page within a
// `data-tralbum` attribute. This package extracts and parses that JSON,
// handling Bandcamp's non-standard date format and fixing malformed JSON.
package bandcamp
