package ioutils

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	_ "image/png" // PNG decoder registration

	"golang.org/x/image/draw"
)

// ImageService provides image processing operations for cover art.
//
// ImageService is used to:
//   - Resize images to fit maximum dimensions (for embedding in MP3 or saving)
//   - Convert images to JPEG format (for better compatibility)
//
// Example usage:
//
//	svc := NewImageService()
//
//	// Download cover art
//	imageData, _ := downloadCoverArt(url)
//
//	// Resize to max 500x500 and convert to JPEG
//	resized, _ := svc.ResizeImage(ctx, imageData, 500, 500)
//	jpeg, _ := svc.ConvertToJPEG(ctx, resized)
type ImageService struct{}

// NewImageService creates a new ImageService.
func NewImageService() *ImageService {
	return &ImageService{}
}

// ResizeImage resizes an image to fit within the specified maximum dimensions.
//
// The aspect ratio is preserved. If the image is already smaller than the
// maximum dimensions, it will still be processed (re-encoded as JPEG).
//
// Parameters:
//   - ctx: Context for cancellation (currently unused)
//   - data: Original image data (JPEG, PNG, etc.)
//   - maxWidth: Maximum width in pixels
//   - maxHeight: Maximum height in pixels
//
// Returns the resized image as JPEG-encoded bytes.
//
// The Catmull-Rom algorithm is used for high-quality resizing.
//
// Example:
//
//	// Resize to fit within 1000x1000, maintaining aspect ratio
//	resized, err := svc.ResizeImage(ctx, imageData, 1000, 1000)
//	// A 1500x1000 image becomes 1000x667
//	// A 800x600 image remains 800x600 (but re-encoded)
func (s *ImageService) ResizeImage(ctx context.Context, data []byte, maxWidth, maxHeight int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	if width > maxWidth || height > maxHeight {
		ratio := float64(width) / float64(height)
		if float64(maxWidth)/float64(maxHeight) > ratio {
			// Height is the limiting factor
			width = int(float64(maxHeight) * ratio)
			height = maxHeight
		} else {
			// Width is the limiting factor
			height = int(float64(maxWidth) / ratio)
			width = maxWidth
		}
	}

	// Create new image with calculated dimensions
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Use Catmull-Rom for high-quality scaling
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	// Encode to JPEG with high quality
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ConvertToJPEG converts an image to JPEG format.
//
// This is useful for:
//   - Ensuring consistent format for ID3 cover art embedding
//   - Reducing file size compared to PNG
//   - Better compatibility with older players
//
// Parameters:
//   - ctx: Context for cancellation (currently unused)
//   - data: Original image data (JPEG, PNG, GIF, etc.)
//
// Returns the image as JPEG-encoded bytes with 90% quality.
//
// Note: If the input is already JPEG, it will be re-encoded, which may
// slightly change file size but ensures consistent encoding.
//
// Example:
//
//	pngData, _ := downloadImage("cover.png")
//	jpegData, err := svc.ConvertToJPEG(ctx, pngData)
func (s *ImageService) ConvertToJPEG(ctx context.Context, data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
