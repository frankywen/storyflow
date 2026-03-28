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

// VolcEngineProvider implements LLMProvider for VolcEngine/Doubao
// VolcEngine uses OpenAI-compatible API
// API Docs: https://www.volcengine.com/docs/82379/1298454
type VolcEngineProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// NewVolcEngineProvider creates a new VolcEngine provider
func NewVolcEngineProvider(cfg LLMConfig) *VolcEngineProvider {
	if cfg.Model == "" {
		cfg.Model = "doubao-pro-32k" // Default model
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	return &VolcEngineProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		baseURL: cfg.BaseURL,
	}
}

// volcengineRequest represents a VolcEngine API request (OpenAI-compatible)
type volcengineRequest struct {
	Model       string                 `json:"model"`
	Messages    []volcengineMessage    `json:"messages"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream"`
}

// volcengineMessage represents a message in VolcEngine format
type volcengineMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// volcengineResponse represents a VolcEngine API response
type volcengineResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []volcengineChoice     `json:"choices"`
	Usage   volcengineUsage        `json:"usage"`
}

// volcengineChoice represents a choice in the response
type volcengineChoice struct {
	Index        int                `json:"index"`
	Message      volcengineMessage  `json:"message"`
	FinishReason string             `json:"finish_reason"`
}

// volcengineUsage represents token usage
type volcengineUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GetName returns the provider name
func (p *VolcEngineProvider) GetName() string {
	return "volcengine"
}

// SendMessage sends a message to VolcEngine
func (p *VolcEngineProvider) SendMessage(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (*LLMResponse, error) {
	messages := []LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}
	return p.SendMessageWithHistory(ctx, "", messages, maxTokens)
}

// SendMessageWithHistory sends a message with conversation history
func (p *VolcEngineProvider) SendMessageWithHistory(ctx context.Context, systemPrompt string, messages []LLMMessage, maxTokens int) (*LLMResponse, error) {
	// Convert to VolcEngine format
	volcMessages := make([]volcengineMessage, 0, len(messages)+1)

	// Add system prompt as first message if provided
	if systemPrompt != "" {
		volcMessages = append(volcMessages, volcengineMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Add conversation messages
	for _, m := range messages {
		volcMessages = append(volcMessages, volcengineMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	reqBody := volcengineRequest{
		Model:       p.model,
		Messages:    volcMessages,
		MaxTokens:   maxTokens,
		Temperature: 0.7,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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

	var result volcengineResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := result.Choices[0]

	return &LLMResponse{
		ID:           result.ID,
		Model:        result.Model,
		Content:      choice.Message.Content,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		StopReason:   choice.FinishReason,
	}, nil
}