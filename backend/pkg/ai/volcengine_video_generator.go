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

// VolcengineVideoGenerator implements VideoGenerator for VolcEngine/Doubao
// API Docs: https://www.volcengine.com/docs/6791/1347772
type VolcengineVideoGenerator struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// NewVolcengineVideoGenerator creates a new VolcEngine video generator
func NewVolcengineVideoGenerator(cfg VideoGeneratorConfig) *VolcengineVideoGenerator {
	if cfg.Model == "" {
		cfg.Model = "doubao-seedance-1-0-i2v-250428" // 默认图生视频模型
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	timeout := 10 * time.Minute
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}
	return &VolcengineVideoGenerator{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: cfg.BaseURL,
	}
}

// volcengineVideoRequest represents a VolcEngine video generation request
type volcengineVideoRequest struct {
	Model     string                 `json:"model"`
	Prompt    string                 `json:"prompt,omitempty"`
	ImageURL  string                 `json:"image_url,omitempty"`
	ImageURLs []string               `json:"image_urls,omitempty"`
	Duration  float64                `json:"duration,omitempty"`
	Size      string                 `json:"size,omitempty"`
	Extra     map[string]interface{} `json:"-,omitempty"` // 扩展字段
}

// volcengineVideoResponse represents a VolcEngine video generation response
type volcengineVideoResponse struct {
	ID      string                   `json:"id,omitempty"`
	Created int64                    `json:"created,omitempty"`
	Data    []volcengineVideoData    `json:"data,omitempty"`
	Error   *volcengineVideoError    `json:"error,omitempty"`
}

// volcengineVideoData represents video data in response
type volcengineVideoData struct {
	TaskID   string  `json:"task_id,omitempty"`
	Status   string  `json:"status,omitempty"`
	VideoURL string  `json:"url,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

// volcengineVideoError represents an error response
type volcengineVideoError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// volcengineTaskStatusResponse represents task status response
type volcengineTaskStatusResponse struct {
	ID        string                        `json:"id"`
	Model     string                        `json:"model"`
	Status    string                        `json:"status"`
	Content   *volcengineTaskContent        `json:"content,omitempty"`
	Error     *volcengineVideoError         `json:"error,omitempty"`
	CreatedAt int64                         `json:"created_at"`
	UpdatedAt int64                         `json:"updated_at"`
	Duration  int                           `json:"duration,omitempty"`
}

// volcengineTaskContent contains the generated content
type volcengineTaskContent struct {
	VideoURL string `json:"video_url,omitempty"`
}

// GetName returns the provider name
func (g *VolcengineVideoGenerator) GetName() string {
	return "volcengine"
}

// GenerateFromImage generates video from an image (图生视频)
func (g *VolcengineVideoGenerator) GenerateFromImage(ctx context.Context, req *VideoRequest) (*VideoResult, error) {
	// Build request - Volcengine contents API format
	// Format: content is an array with image_url object
	content := []map[string]interface{}{
		{
			"type": "image_url",
			"image_url": map[string]string{
				"url": req.ImageURL,
			},
		},
	}

	videoReq := map[string]interface{}{
		"model":   g.model,
		"content": content,
	}

	// Add optional parameters
	if req.Duration > 0 {
		videoReq["duration"] = req.Duration
	}
	if req.Resolution != "" {
		videoReq["resolution"] = req.Resolution
	}
	if req.FPS > 0 {
		videoReq["framespersecond"] = req.FPS
	}

	body, err := json.Marshal(videoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Printf("Volcengine Video Request: %s\n", string(body))

	// Use base URL directly
	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL, bytes.NewReader(body))
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

	fmt.Printf("Volcengine Video API Response: status=%d, body=%s\n", resp.StatusCode, string(respBody)[:min(500, len(respBody))])

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response - format: {"id": "cgt-xxx"}
	var result struct {
		ID    string                   `json:"id"`
		Error *volcengineVideoError    `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", result.Error.Code, result.Error.Message)
	}

	return &VideoResult{
		TaskID: result.ID,
		Status: "pending",
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateFromImages generates video from multiple images
func (g *VolcengineVideoGenerator) GenerateFromImages(ctx context.Context, req *VideoRequest, imageUrls []string) (*VideoResult, error) {
	if len(imageUrls) == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	// Build content array with all images
	content := make([]map[string]interface{}, len(imageUrls))
	for i, url := range imageUrls {
		content[i] = map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": url,
			},
		}
	}

	videoReq := map[string]interface{}{
		"model":   g.model,
		"content": content,
	}

	body, err := json.Marshal(videoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL, bytes.NewReader(body))
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

	var result struct {
		ID    string                `json:"id"`
		Error *volcengineVideoError `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", result.Error.Code, result.Error.Message)
	}

	return &VideoResult{
		TaskID: result.ID,
		Status: "pending",
	}, nil
}

// GetTaskStatus gets the status of a video generation task
func (g *VolcengineVideoGenerator) GetTaskStatus(ctx context.Context, taskID string) (*VideoResult, error) {
	// Use base URL with task ID appended
	statusURL := g.baseURL + "/" + taskID

	httpReq, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	var result volcengineTaskStatusResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videoResult := &VideoResult{
		TaskID: taskID,
		Status: result.Status,
	}

	// Map status
	switch result.Status {
	case "pending", "queued":
		videoResult.Status = "pending"
		videoResult.Progress = 0
	case "processing", "running":
		videoResult.Status = "processing"
		videoResult.Progress = 50
	case "succeeded", "completed":
		videoResult.Status = "completed"
		videoResult.Progress = 100
		if result.Content != nil {
			videoResult.VideoURL = result.Content.VideoURL
		}
		if result.Duration > 0 {
			videoResult.Duration = float64(result.Duration)
		}
	case "failed", "error":
		videoResult.Status = "failed"
		if result.Error != nil {
			videoResult.Error = result.Error.Message
		}
	}

	return videoResult, nil
}