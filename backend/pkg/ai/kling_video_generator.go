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

// KlingVideoGenerator implements VideoGenerator for Kling AI
// API Docs: https://platform.klingai.com/docs/api/text-to-video
type KlingVideoGenerator struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewKlingVideoGenerator creates a new Kling video generator
func NewKlingVideoGenerator(cfg VideoGeneratorConfig) *KlingVideoGenerator {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.klingai.com/v1"
	}
	timeout := 10 * time.Minute
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}
	return &KlingVideoGenerator{
		apiKey: cfg.APIKey,
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// klingImageToVideoRequest represents Kling image-to-video request
type klingImageToVideoRequest struct {
	Model          string `json:"model"`
	ImageURL       string `json:"image_url"`
	Prompt         string `json:"prompt,omitempty"`
	Duration       string `json:"duration,omitempty"` // "5" or "10"
	AspectRatio    string `json:"aspect_ratio,omitempty"`
	CallbackURL    string `json:"callback_url,omitempty"`
}

// klingTaskResponse represents Kling task creation response
type klingTaskResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    klingTaskData `json:"data"`
}

// klingTaskData contains task info
type klingTaskData struct {
	TaskID     string `json:"task_id"`
	TaskStatus string `json:"task_status"`
}

// klingTaskResultResponse represents Kling task result response
type klingTaskResultResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    klingTaskResult  `json:"data"`
}

// klingTaskResult contains task result
type klingTaskResult struct {
	TaskID     string           `json:"task_id"`
	TaskStatus string           `json:"task_status"`
	TaskStatusDesc string       `json:"task_status_desc"`
	CreatedAt  int64            `json:"created_at"`
	UpdatedAt  int64            `json:"updated_at"`
	TaskResult *klingVideoResult `json:"task_result,omitempty"`
}

// klingVideoResult contains video result
type klingVideoResult struct {
	Videos []klingVideo `json:"videos"`
}

// klingVideo contains video info
type klingVideo struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Duration  string `json:"duration"`
}

// GetName returns the provider name
func (g *KlingVideoGenerator) GetName() string {
	return "kling"
}

// GenerateFromImage generates video from an image
func (g *KlingVideoGenerator) GenerateFromImage(ctx context.Context, req *VideoRequest) (*VideoResult, error) {
	// Build request
	duration := "5"
	if req.Duration > 5 {
		duration = "10"
	}

	klingReq := klingImageToVideoRequest{
		Model:       "kling-v1",
		ImageURL:    req.ImageURL,
		Prompt:      req.Prompt,
		Duration:    duration,
		AspectRatio: "16:9",
	}

	body, err := json.Marshal(klingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/videos/image2video", bytes.NewReader(body))
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

	var result klingTaskResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &VideoResult{
		TaskID: result.Data.TaskID,
		Status: "pending",
	}, nil
}

// GenerateFromImages generates video from multiple images
func (g *KlingVideoGenerator) GenerateFromImages(ctx context.Context, req *VideoRequest, imageUrls []string) (*VideoResult, error) {
	// Kling doesn't support multiple images directly, use first image
	if len(imageUrls) == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	req.ImageURL = imageUrls[0]
	return g.GenerateFromImage(ctx, req)
}

// GetTaskStatus gets the status of a video generation task
func (g *KlingVideoGenerator) GetTaskStatus(ctx context.Context, taskID string) (*VideoResult, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/videos/image2video/"+taskID, nil)
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

	var result klingTaskResultResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videoResult := &VideoResult{
		TaskID: taskID,
		Status: result.Data.TaskStatus,
	}

	// Map status
	switch result.Data.TaskStatus {
	case "processing":
		videoResult.Status = "processing"
		videoResult.Progress = 50
	case "succeed":
		videoResult.Status = "completed"
		videoResult.Progress = 100
		if result.Data.TaskResult != nil && len(result.Data.TaskResult.Videos) > 0 {
			videoResult.VideoURL = result.Data.TaskResult.Videos[0].URL
		}
	case "failed":
		videoResult.Status = "failed"
		videoResult.Error = result.Data.TaskStatusDesc
	}

	return videoResult, nil
}