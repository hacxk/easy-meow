// File: utils/thumbnail.go

package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// GetThumbnail generates a thumbnail from the given file path and returns the thumbnail as a byte slice
func GetThumbnail(path string) ([]byte, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "thumbnail")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Detect file type
	contentType, err := getFileContentType(file)
	if err != nil {
		return nil, err
	}

	var img image.Image

	switch {
	case isImageContentType(contentType):
		img, err = decodeImage(file, contentType)
	case isVideoContentType(contentType):
		img, err = extractVideoThumbnail(path, tempDir)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process file: %w", err)
	}

	// Determine appropriate thumbnail dimensions
	maxWidth, maxHeight := getAppropriateThumbSize(img.Bounds().Dx(), img.Bounds().Dy())

	// Resize the image
	resized := resize.Thumbnail(uint(maxWidth), uint(maxHeight), img, resize.Lanczos3)

	// Encode the resized image to JPEG
	thumbnailBuf := new(bytes.Buffer)
	err = jpeg.Encode(thumbnailBuf, resized, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return thumbnailBuf.Bytes(), nil
}

func getFileContentType(file *os.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}
	return http.DetectContentType(buffer), nil
}

func isImageContentType(contentType string) bool {
	return contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/gif"
}

func isVideoContentType(contentType string) bool {
	return contentType == "video/mp4" || contentType == "video/mpeg" || contentType == "video/quicktime"
}

func decodeImage(file io.Reader, contentType string) (image.Image, error) {
	switch contentType {
	case "image/jpeg":
		return jpeg.Decode(file)
	case "image/png":
		return png.Decode(file)
	case "image/gif":
		gifImg, err := gif.DecodeAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode GIF: %w", err)
		}
		return gifImg.Image[0], nil
	default:
		return nil, fmt.Errorf("unsupported image type: %s", contentType)
	}
}

func extractVideoThumbnail(videoPath, tempDir string) (image.Image, error) {
	thumbPath := filepath.Join(tempDir, "thumbnail.jpg")

	err := ffmpeg.Input(videoPath).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", 1)}).
		Output(thumbPath, ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		OverWriteOutput().
		Run()

	if err != nil {
		return nil, fmt.Errorf("failed to extract video thumbnail: %w", err)
	}

	img, err := imaging.Open(thumbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open thumbnail image: %w", err)
	}

	return img, nil
}

func getAppropriateThumbSize(width, height int) (int, int) {
	maxSize := 200
	if width > height {
		return maxSize, int(float64(height) * (float64(maxSize) / float64(width)))
	}
	return int(float64(width) * (float64(maxSize) / float64(height))), maxSize
}
