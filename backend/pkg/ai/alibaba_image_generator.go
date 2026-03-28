package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AlibabaImageGenerator implements ImageGenerator for Alibaba Wanxiang (通义万相)
// Supports OpenAI-compatible API for Bailian Coding
type AlibabaImageGenerator struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// NewAlibabaImageGenerator creates a new Alibaba image generator
func NewAlibabaImageGenerator(cfg ImageGeneratorConfig) *AlibabaImageGenerator {
	if cfg.Model == "" {
		cfg.Model = "wanx-v1"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://dashscope.aliyuncs.com/api/v1"
	}
	timeout := 5 * time.Minute
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}
	return &AlibabaImageGenerator{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: cfg.BaseURL,
	}
}

// openAIImageRequest represents OpenAI-compatible image generation request
type openAIImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

// openAIImageResponse represents OpenAI-compatible image generation response
type openAIImageResponse struct {
	Created int64                `json:"created"`
	Data    []openAIImageData    `json:"data"`
	Error   *openAIImageError    `json:"error,omitempty"`
}

// openAIImageData represents image data in response
type openAIImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// openAIImageError represents an error response
type openAIImageError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// GetName returns the provider name
func (g *AlibabaImageGenerator) GetName() string {
	return "alibaba"
}

// Generate generates a single image from prompt
func (g *AlibabaImageGenerator) Generate(ctx context.Context, req *ImageRequest) (*ImageResult, error) {
	// Set defaults
	if req.Width == 0 {
		req.Width = 1024
	}
	if req.Height == 0 {
		req.Height = 1024
	}

	// Build style prompt
	stylePrompt := req.Prompt
	if req.Style != "" {
		stylePrompt = fmt.Sprintf("%s style, %s, high quality, detailed", req.Style, req.Prompt)
	}

	// Build size string
	size := fmt.Sprintf("%dx%d", req.Width, req.Height)

	// Build request using OpenAI-compatible format
	imgReq := openAIImageRequest{
		Model:  g.model,
		Prompt: stylePrompt,
		N:      1,
		Size:   size,
	}

	body, err := json.Marshal(imgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use OpenAI-compatible endpoint
	url := g.baseURL + "/images/generations"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result openAIImageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for error
	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", result.Error.Code, result.Error.Message)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	imgData := result.Data[0]
	return &ImageResult{
		ID:       fmt.Sprintf("%d", result.Created),
		ImageURL: imgData.URL,
		Seed:     req.Seed,
		Width:    req.Width,
		Height:   req.Height,
		Model:    g.model,
	}, nil
}

// GenerateBatch generates multiple images
func (g *AlibabaImageGenerator) GenerateBatch(ctx context.Context, req *ImageRequest, count int) ([]*ImageResult, error) {
	results := make([]*ImageResult, count)
	for i := 0; i < count; i++ {
		result, err := g.Generate(ctx, req)
		if err != nil {
			return results, err
		}
		results[i] = result
	}
	return results, nil
}