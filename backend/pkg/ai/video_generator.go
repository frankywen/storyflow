package ai

import (
	"context"
)

// VideoGenerator defines the interface for video generation providers
type VideoGenerator interface {
	// GenerateFromImage generates video from an image
	GenerateFromImage(ctx context.Context, req *VideoRequest) (*VideoResult, error)

	// GenerateFromImages generates video from multiple images (slideshow)
	GenerateFromImages(ctx context.Context, req *VideoRequest, imageUrls []string) (*VideoResult, error)

	// GetTaskStatus gets the status of a video generation task
	GetTaskStatus(ctx context.Context, taskID string) (*VideoResult, error)

	// GetName returns the provider name
	GetName() string
}

// VideoRequest represents a video generation request
type VideoRequest struct {
	ImageURL    string            `json:"image_url,omitempty"`
	Prompt      string            `json:"prompt,omitempty"`
	Duration    float64           `json:"duration,omitempty"`    // Duration in seconds
	Resolution  string            `json:"resolution,omitempty"` // e.g., "1080p", "720p"
	FPS         int               `json:"fps,omitempty"`
	MotionLevel string            `json:"motion_level,omitempty"` // low, medium, high
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// VideoResult represents a video generation result
type VideoResult struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"` // pending, processing, completed, failed
	Progress   int    `json:"progress"`
	VideoURL   string `json:"video_url,omitempty"`
	Thumbnail  string `json:"thumbnail,omitempty"`
	Duration   float64 `json:"duration,omitempty"`
	Error      string `json:"error,omitempty"`
}

// VideoGeneratorConfig holds configuration for video generators
type VideoGeneratorConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string // Model name (e.g., "doubao-seedance-1-0-i2v-250428")
	Timeout  int
}

// NewVideoGenerator creates a video generator based on config
func NewVideoGenerator(cfg VideoGeneratorConfig) VideoGenerator {
	switch cfg.Provider {
	case "kling":
		return NewKlingVideoGenerator(cfg)
	case "runway":
		return NewRunwayVideoGenerator(cfg)
	case "volcengine", "doubao":
		return NewVolcengineVideoGenerator(cfg)
	default:
		// Return a mock generator for testing
		return NewMockVideoGenerator(cfg)
	}
}