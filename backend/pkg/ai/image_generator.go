package ai

import (
	"context"
)

// ImageGenerator defines the interface for image generation providers
type ImageGenerator interface {
	// Generate generates a single image from prompt
	Generate(ctx context.Context, req *ImageRequest) (*ImageResult, error)

	// GenerateBatch generates multiple images
	GenerateBatch(ctx context.Context, req *ImageRequest, count int) ([]*ImageResult, error)

	// GetName returns the provider name
	GetName() string
}

// ImageToImageGenerator defines the interface for image-to-image generation
// This is an optional extension to ImageGenerator
type ImageToImageGenerator interface {
	ImageGenerator
	// GenerateFromImage generates an image based on a reference image
	GenerateFromImage(ctx context.Context, req *ImageRequest, refImageURL string) (*ImageResult, error)
}

// ImageRequest represents an image generation request
type ImageRequest struct {
	Prompt         string            `json:"prompt"`
	NegativePrompt string            `json:"negative_prompt,omitempty"`
	Width          int               `json:"width,omitempty"`
	Height         int               `json:"height,omitempty"`
	Seed           int64             `json:"seed,omitempty"`
	Steps          int               `json:"steps,omitempty"`
	Style          string            `json:"style,omitempty"` // manga, realistic, anime, etc.
	// For character consistency
	ReferenceImage string  `json:"reference_image,omitempty"` // Reference image URL for img2img
	RefStrength    float64 `json:"ref_strength,omitempty"`    // Reference strength (0-1), default 0.5
	Extra          map[string]interface{} `json:"extra,omitempty"` // Provider-specific options
}

// ImageResult represents an image generation result
type ImageResult struct {
	ID        string   `json:"id"`
	ImageURL  string   `json:"image_url"`
	Seed      int64    `json:"seed"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Model     string   `json:"model"`
	ImageData []byte   `json:"-"` // Raw image data if available
}

// ImageGeneratorConfig holds configuration for image generators
type ImageGeneratorConfig struct {
	Provider string // "comfyui", "volcengine", "alibaba"
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  int // timeout in seconds
}

// NewImageGenerator creates an image generator based on config
func NewImageGenerator(cfg ImageGeneratorConfig) ImageGenerator {
	switch cfg.Provider {
	case "comfyui":
		return NewComfyUIImageGenerator(cfg)
	case "volcengine", "doubao":
		return NewVolcEngineImageGenerator(cfg)
	case "alibaba", "wanxiang":
		return NewAlibabaImageGenerator(cfg)
	default:
		return NewComfyUIImageGenerator(cfg)
	}
}