// Package media provides AI-powered media generation interfaces and adapters.
// It defines the ImageGenerator and VideoGenerator interfaces along with
// shared types used by all media generation adapters (image, video, audio, subtitle).
package media

import (
	"context"
	"time"
)

// MediaType represents the type of generated media.
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeSubtitle MediaType = "subtitle"
)

// GenerationStatus represents the current status of a media generation job.
type GenerationStatus string

const (
	StatusPending    GenerationStatus = "pending"
	StatusProcessing GenerationStatus = "processing"
	StatusCompleted  GenerationStatus = "completed"
	StatusFailed     GenerationStatus = "failed"
)

// GenerationResult holds the result of a media generation operation.
type GenerationResult struct {
	ID        string           `json:"id"`
	MediaType MediaType        `json:"media_type"`
	Status    GenerationStatus `json:"status"`
	Data      []byte           `json:"data,omitempty"`
	URL       string           `json:"url,omitempty"`
	Error     string           `json:"error,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// ImageGenerator defines the interface for AI image generation services.
type ImageGenerator interface {
	// Generate creates an image from a text prompt.
	// Returns the image bytes (e.g., PNG/JPEG) or an error.
	Generate(ctx context.Context, prompt string, opts map[string]interface{}) ([]byte, error)
}

// VideoGenerator defines the interface for AI video generation services.
type VideoGenerator interface {
	// GenerateVideo submits a text-to-video generation job and returns a job ID.
	GenerateVideo(ctx context.Context, prompt string, opts map[string]interface{}) (string, error)
}

// StatusChecker defines the interface for checking generation job status.
type StatusChecker interface {
	// CheckStatus queries the current status of a generation job.
	CheckStatus(ctx context.Context, jobID string) (*GenerationResult, error)
}
