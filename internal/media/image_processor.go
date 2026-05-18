package media

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
)

// ImageProcessor provides image format conversion, resizing, compression, and watermarking.
type ImageProcessor struct {
	maxWidth  int
	maxHeight int
	quality   int // JPEG quality 1-100
}

// NewImageProcessor creates an image processor with default settings.
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		maxWidth:  4096,
		maxHeight: 4096,
		quality:   85,
	}
}

// NewImageProcessorWithConfig creates a processor with custom limits.
func NewImageProcessorWithConfig(maxWidth, maxHeight, quality int) *ImageProcessor {
	if quality < 1 {
		quality = 85
	}
	if quality > 100 {
		quality = 100
	}
	return &ImageProcessor{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		quality:   quality,
	}
}

// ProcessResult holds the result of image processing.
type ProcessResult struct {
	Data   []byte `json:"data"`
	Format string `json:"format"` // "jpeg" or "png"
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// ToJPEG converts image bytes to JPEG format with the configured quality.
func (p *ImageProcessor) ToJPEG(data []byte) (*ProcessResult, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	resized := p.resizeImage(img, bounds.Dx(), bounds.Dy())

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: p.quality}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}

	return &ProcessResult{
		Data:   buf.Bytes(),
		Format: "jpeg",
		Width:  resized.Bounds().Dx(),
		Height: resized.Bounds().Dy(),
	}, nil
}

// ToPNG converts image bytes to PNG format.
func (p *ImageProcessor) ToPNG(data []byte) (*ProcessResult, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	resized := p.resizeImage(img, bounds.Dx(), bounds.Dy())

	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}

	return &ProcessResult{
		Data:   buf.Bytes(),
		Format: "png",
		Width:  resized.Bounds().Dx(),
		Height: resized.Bounds().Dy(),
	}, nil
}

// Resize scales an image to fit within maxWidth x maxHeight while maintaining aspect ratio.
func (p *ImageProcessor) Resize(data []byte, targetWidth, targetHeight int) (*ProcessResult, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	resized := p.resizeTo(img, targetWidth, targetHeight)

	var buf bytes.Buffer
	if format == "png" {
		if err := png.Encode(&buf, resized); err != nil {
			return nil, fmt.Errorf("encode resized png: %w", err)
		}
	} else {
		if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: p.quality}); err != nil {
			return nil, fmt.Errorf("encode resized jpeg: %w", err)
		}
	}

	return &ProcessResult{
		Data:   buf.Bytes(),
		Format: format,
		Width:  resized.Bounds().Dx(),
		Height: resized.Bounds().Dy(),
	}, nil
}

// Compress reduces file size by re-encoding with lower quality.
func (p *ImageProcessor) Compress(data []byte, quality int) (*ProcessResult, error) {
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	var buf bytes.Buffer
	if format == "png" {
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("encode compressed png: %w", err)
		}
	} else {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("encode compressed jpeg: %w", err)
		}
	}

	return &ProcessResult{
		Data:   buf.Bytes(),
		Format: format,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}, nil
}

// resizeImage scales an image down if it exceeds max dimensions, preserving aspect ratio.
func (p *ImageProcessor) resizeImage(img image.Image, w, h int) image.Image {
	return p.resizeTo(img, p.maxWidth, p.maxHeight)
}

// resizeTo scales an image to fit within the target dimensions while maintaining aspect ratio.
func (p *ImageProcessor) resizeTo(img image.Image, targetW, targetH int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= targetW && h <= targetH {
		return img
	}

	ratio := math.Min(float64(targetW)/float64(w), float64(targetH)/float64(h))
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)

	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	return p.nearestNeighborResize(img, newW, newH)
}

// nearestNeighborResize performs a fast nearest-neighbor downscale.
func (p *ImageProcessor) nearestNeighborResize(img image.Image, newW, newH int) image.Image {
	bounds := img.Bounds()
	oldW := bounds.Dx()
	oldH := bounds.Dy()

	result := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := x * oldW / newW
			srcY := y * oldH / newH
			result.Set(x, y, img.At(srcX, srcY))
		}
	}
	return result
}
