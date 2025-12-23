// Package ioutils provides file system utilities for the bandcamp-downloader.
//
// This package contains functions for:
//   - File copying
//   - File writing
//   - Filename sanitization
//   - Directory creation
//
// All functions that accept a context.Context respect cancellation,
// though file operations themselves may not be interruptible.
package ioutils

import (
	"context"
	"io"
	"os"
	"regexp"
	"strings"
)

// CopyFile copies a file from source to destination.
//
// The destination file is created with mode 0644 if it doesn't exist,
// or truncated if it does. The source file must exist and be readable.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused but reserved for future use)
//   - src: Source file path (must exist)
//   - dst: Destination file path (will be created/overwritten)
//
// Returns an error if:
//   - Source file cannot be opened
//   - Destination file cannot be created
//   - Copy operation fails
//
// Example:
//
//	err := CopyFile(ctx, "/path/to/source.mp3", "/path/to/dest.mp3")
func CopyFile(ctx context.Context, src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// WriteFile writes data to a file, creating it if necessary.
//
// The file is created with mode 0644. If the file already exists,
// it is truncated before writing.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused but reserved for future use)
//   - path: File path to write to
//   - data: Bytes to write
//
// Example:
//
//	playlistContent := []byte("#EXTM3U\n...")
//	err := WriteFile(ctx, "/music/playlist.m3u", playlistContent)
func WriteFile(ctx context.Context, path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// SanitizeFileName removes or replaces characters that are invalid in file/folder names.
//
// This function ensures filenames are valid across different operating systems,
// particularly Windows which has the most restrictive naming rules.
//
// The following transformations are applied:
//   - Invalid characters (<>:"/\|?* and control chars 0x00-0x1f) → underscore
//   - Trailing dots → removed (Windows limitation)
//   - Multiple whitespace → single space
//   - Trailing whitespace → removed
//
// Example:
//
//	SanitizeFileName("Song: Part 1/2")     // Returns "Song_ Part 1_2"
//	SanitizeFileName("Track...")           // Returns "Track"
//	SanitizeFileName("Name   with  spaces") // Returns "Name with spaces"
func SanitizeFileName(name string) string {
	// Replace invalid path/file characters with underscore
	// Characters: < > : " / \ | ? * and control characters (0x00-0x1f)
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	name = invalidChars.ReplaceAllString(name, "_")

	// Remove trailing dots (Windows doesn't allow filenames ending with dots)
	name = regexp.MustCompile(`\.+$`).ReplaceAllString(name, "")

	// Replace multiple whitespace with single space for cleaner names
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	// Remove trailing whitespace
	name = strings.TrimRight(name, " ")

	return name
}

// EnsureDir creates a directory and all parent directories if they don't exist.
//
// Directories are created with mode 0755 (rwxr-xr-x).
// If the directory already exists, no error is returned.
//
// Example:
//
//	err := EnsureDir("/music/Artist/Album")
//	// Creates /music, /music/Artist, and /music/Artist/Album if needed
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
