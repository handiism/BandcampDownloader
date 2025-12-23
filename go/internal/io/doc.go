// Package ioutils provides file system and image processing utilities.
//
// This package contains functions for:
//   - File copying and writing
//   - Filename sanitization for cross-platform compatibility
//   - Directory creation
//   - Image resizing and format conversion
//
// # File Operations
//
//	// Copy a file
//	err := ioutils.CopyFile(ctx, "/src/file.mp3", "/dst/file.mp3")
//	
//	// Write data to file
//	err := ioutils.WriteFile(ctx, "/path/to/file.txt", []byte("content"))
//	
//	// Ensure directory exists
//	err := ioutils.EnsureDir("/path/to/new/directory")
//
// # Filename Sanitization
//
// Use SanitizeFileName to remove invalid characters from filenames:
//
//	safe := ioutils.SanitizeFileName("Song: Part 1/2") // Returns "Song_ Part 1_2"
//
// # Image Processing
//
// The ImageService handles cover art manipulation:
//
//	svc := ioutils.NewImageService()
//	
//	// Resize image to fit within 500x500
//	resized, _ := svc.ResizeImage(ctx, imageData, 500, 500)
//	
//	// Convert to JPEG
//	jpeg, _ := svc.ConvertToJPEG(ctx, pngData)
package ioutils
