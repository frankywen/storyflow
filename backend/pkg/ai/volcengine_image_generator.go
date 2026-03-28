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

// VolcEngineImageGenerator implements ImageGenerator for VolcEngine/Doubao
// API Docs: https://www.volcengine.com/docs/6791/1347778
type VolcEngineImageGenerator struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// NewVolcEngineImageGenerator creates a new VolcEngine image generator
func NewVolcEngineImageGenerator(cfg ImageGeneratorConfig) *VolcEngineImageGenerator {
	if cfg.Model == "" {
		cfg.Model = "doubao-seedream-3-0-t2i-250415"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	timeout := 5 * time.Minute
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}
	return &VolcEngineImageGenerator{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: cfg.BaseURL,
	}
}

// volcengineImgRequest represents a VolcEngine image generation request (OpenAI-compatible)
type volcengineImgRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

// volcengineImgResponse represents a VolcEngine image generation response
type volcengineImgResponse struct {
	Created int64               `json:"created"`
	Data    []volcengineImgData `json:"data"`
	Error   *volcengineImgError `json:"error,omitempty"`
}

// volcengineImgData represents image data in response
type volcengineImgData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// volcengineImgError represents an error response
type volcengineImgError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// GetName returns the provider name
func (g *VolcEngineImageGenerator) GetName() string {
	return "volcengine"
}

// Generate generates a single image from prompt
func (g *VolcEngineImageGenerator) Generate(ctx context.Context, req *ImageRequest) (*ImageResult, error) {
	// Set defaults - VolcEngine requires at least 3686400 pixels (e.g., 2048x2048)
	if req.Width == 0 {
		req.Width = 2048
	}
	if req.Height == 0 {
		req.Height = 2048
	}

	// Build style prompt
	stylePrompt := req.Prompt
	if req.Style != "" {
		stylePrompt = fmt.Sprintf("%s style, %s, high quality, detailed", req.Style, req.Prompt)
	}

	// Build size string
	size := fmt.Sprintf("%dx%d", req.Width, req.Height)

	// Build request - OpenAI-compatible format
	imgReq := volcengineImgRequest{
		Model:  g.model,
		Prompt: stylePrompt,
		N:      1,
		Size:   size,
	}

	body, err := json.Marshal(imgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/images/generations", bytes.NewReader(body))
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

	var result volcengineImgResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

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
func (g *VolcEngineImageGenerator) GenerateBatch(ctx context.Context, req *ImageRequest, count int) ([]*ImageResult, error) {
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