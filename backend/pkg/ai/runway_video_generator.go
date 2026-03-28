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

// RunwayVideoGenerator implements VideoGenerator for Runway Gen-3
// API Docs: https://docs.runwayml.com/
type RunwayVideoGenerator struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewRunwayVideoGenerator creates a new Runway video generator
func NewRunwayVideoGenerator(cfg VideoGeneratorConfig) *RunwayVideoGenerator {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.runwayml.com/v1"
	}
	timeout := 10 * time.Minute
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}
	return &RunwayVideoGenerator{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// runwayImageToVideoRequest represents Runway image-to-video request
type runwayImageToVideoRequest struct {
	Model            string `json:"model"`
	Image            string `json:"image,omitempty"`
	PromptText       string `json:"promptText,omitempty"`
	Duration         int    `json:"duration,omitempty"`
	AspectRatio      string `json:"aspectRatio,omitempty"`
	Seed             int64  `json:"seed,omitempty"`
	Watermark        bool   `json:"watermark"`
}

// runwayTaskResponse represents Runway task response
type runwayTaskResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// runwayTaskResultResponse represents Runway task result
type runwayTaskResultResponse struct {
	ID      string         `json:"id"`
	Status  string         `json:"status"`
	Progress float64       `json:"progress"`
	Result  *runwayResult  `json:"result,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// runwayResult contains video result
type runwayResult struct {
	URL string `json:"url"`
}

// GetName returns the provider name
func (g *RunwayVideoGenerator) GetName() string {
	return "runway"
}

// GenerateFromImage generates video from an image
func (g *RunwayVideoGenerator) GenerateFromImage(ctx context.Context, req *VideoRequest) (*VideoResult, error) {
	duration := 5
	if req.Duration > 0 {
		duration = int(req.Duration)
		if duration > 10 {
			duration = 10
		}
	}

	runwayReq := runwayImageToVideoRequest{
		Model:       "gen3a_turbo",
		Image:       req.ImageURL,
		PromptText:  req.Prompt,
		Duration:    duration,
		AspectRatio: "16:9",
		Watermark:   false,
	}

	body, err := json.Marshal(runwayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/image_to_video", bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result runwayTaskResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &VideoResult{
		TaskID: result.ID,
		Status: "pending",
	}, nil
}

// GenerateFromImages generates video from multiple images
func (g *RunwayVideoGenerator) GenerateFromImages(ctx context.Context, req *VideoRequest, imageUrls []string) (*VideoResult, error) {
	if len(imageUrls) == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	req.ImageURL = imageUrls[0]
	return g.GenerateFromImage(ctx, req)
}

// GetTaskStatus gets the status of a video generation task
func (g *RunwayVideoGenerator) GetTaskStatus(ctx context.Context, taskID string) (*VideoResult, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/image_to_video/"+taskID, nil)
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

	var result runwayTaskResultResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videoResult := &VideoResult{
		TaskID:   taskID,
		Status:   result.Status,
		Progress: int(result.Progress * 100),
	}

	switch result.Status {
	case "RUNNING":
		videoResult.Status = "processing"
	case "SUCCEEDED":
		videoResult.Status = "completed"
		videoResult.Progress = 100
		if result.Result != nil {
			videoResult.VideoURL = result.Result.URL
		}
	case "FAILED":
		videoResult.Status = "failed"
		videoResult.Error = result.Error
	}

	return videoResult, nil
}