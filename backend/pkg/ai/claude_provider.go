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

// ClaudeProvider implements LLMProvider for Anthropic Claude
type ClaudeProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(cfg LLMConfig) *ClaudeProvider {
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	return &ClaudeProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		baseURL: cfg.BaseURL,
	}
}

// claudeRequest represents a Claude API request
type claudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []claudeMessage `json:"messages"`
	System      string          `json:"system,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// claudeMessage represents a message in Claude format
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse represents a Claude API response
type claudeResponse struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	Role       string              `json:"role"`
	Content    []claudeContentBlock `json:"content"`
	Model      string              `json:"model"`
	StopReason string              `json:"stop_reason"`
	Usage      claudeUsage         `json:"usage"`
}

// claudeContentBlock represents a content block in Claude response
type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// claudeUsage represents token usage in Claude
type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// GetName returns the provider name
func (p *ClaudeProvider) GetName() string {
	return "claude"
}

// SendMessage sends a message to Claude
func (p *ClaudeProvider) SendMessage(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (*LLMResponse, error) {
	messages := []LLMMessage{
		{Role: "user", Content: userMessage},
	}
	return p.SendMessageWithHistory(ctx, systemPrompt, messages, maxTokens)
}

// SendMessageWithHistory sends a message with conversation history
func (p *ClaudeProvider) SendMessageWithHistory(ctx context.Context, systemPrompt string, messages []LLMMessage, maxTokens int) (*LLMResponse, error) {
	// Convert to Claude format
	claudeMessages := make([]claudeMessage, len(messages))
	for i, m := range messages {
		claudeMessages[i] = claudeMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	reqBody := claudeRequest{
		Model:       p.model,
		MaxTokens:   maxTokens,
		System:      systemPrompt,
		Messages:    claudeMessages,
		Temperature: 0.7,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
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

	var result claudeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text content
	var content string
	for _, block := range result.Content {
		if block.Type == "text" {
			content = block.Text
			break
		}
	}

	return &LLMResponse{
		ID:           result.ID,
		Model:        result.Model,
		Content:      content,
		InputTokens:  result.Usage.InputTokens,
		OutputTokens: result.Usage.OutputTokens,
		StopReason:   result.StopReason,
	}, nil
}