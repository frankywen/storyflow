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

// ComfyUIClient handles communication with ComfyUI API
type ComfyUIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewComfyUIClient creates a new ComfyUI client
func NewComfyUIClient(baseURL string) *ComfyUIClient {
	return &ComfyUIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for image generation
		},
	}
}

// WorkflowRequest represents a ComfyUI workflow request
type WorkflowRequest struct {
	Workflow map[string]interface{} `json:"workflow"`
	ClientID string                 `json:"client_id"`
}

// QueueResponse represents the response from queueing a prompt
type QueueResponse struct {
	PromptID string `json:"prompt_id"`
	Number   int    `json:"number"`
	NodeErrors interface{} `json:"node_errors,omitempty"`
}

// HistoryResponse represents the execution history
type HistoryResponse struct {
	Outputs map[string]OutputInfo `json:"outputs"`
}

// OutputInfo contains output file information
type OutputInfo struct {
	Images []ImageInfo `json:"images,omitempty"`
}

// ImageInfo contains image file information
type ImageInfo struct {
	Filename string `json:"filename"`
	Subfolder string `json:"subfolder"`
	Type     string `json:"type"`
}

// QueuePrompt queues a workflow for execution
func (c *ComfyUIClient) QueuePrompt(ctx context.Context, workflow map[string]interface{}, clientID string) (*QueueResponse, error) {
	reqBody := map[string]interface{}{
		"prompt":    workflow,
		"client_id": clientID,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/prompt", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetHistory retrieves execution history for a prompt
func (c *ComfyUIClient) GetHistory(ctx context.Context, promptID string) (*HistoryResponse, error) {
	url := fmt.Sprintf("%s/history/%s", c.baseURL, promptID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetImage retrieves an image from ComfyUI
func (c *ComfyUIClient) GetImage(ctx context.Context, filename, subfolder, imgType string) ([]byte, error) {
	url := fmt.Sprintf("%s/view?filename=%s&subfolder=%s&type=%s",
		c.baseURL, filename, subfolder, imgType)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// WaitForCompletion waits for a prompt to complete and returns the outputs
func (c *ComfyUIClient) WaitForCompletion(ctx context.Context, promptID string, timeout time.Duration) (*HistoryResponse, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		history, err := c.GetHistory(ctx, promptID)
		if err != nil {
			return nil, err
		}

		if len(history.Outputs) > 0 {
			return history, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
			// Continue polling
		}
	}

	return nil, fmt.Errorf("timeout waiting for prompt completion")
}

// UploadImage uploads an image to ComfyUI (for ControlNet, IP-Adapter, etc.)
func (c *ComfyUIClient) UploadImage(ctx context.Context, imageData []byte, filename string, overwrite bool) (string, error) {
	// TODO: Implement image upload
	return "", fmt.Errorf("not implemented")
}