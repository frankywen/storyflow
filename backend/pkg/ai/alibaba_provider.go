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

// AlibabaProvider implements LLMProvider for Alibaba Bailian/Tongyi Qwen
// Supports both DashScope API and OpenAI-compatible API (for Bailian Coding)
type AlibabaProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// NewAlibabaProvider creates a new Alibaba/Tongyi provider
func NewAlibabaProvider(cfg LLMConfig) *AlibabaProvider {
	if cfg.Model == "" {
		cfg.Model = "qwen-max"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://dashscope.aliyuncs.com/api/v1"
	}
	return &AlibabaProvider{
		apiKey: cfg.APIKey,
		model:  cfg.Model,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
		baseURL: cfg.BaseURL,
	}
}

// openAIChatRequest represents OpenAI-compatible chat request
type openAIChatRequest struct {
	Model       string           `json:"model"`
	Messages    []openAIMessage  `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
}

// openAIMessage represents a message in OpenAI format
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse represents OpenAI-compatible chat response
type openAIChatResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []openAIChoice   `json:"choices"`
	Usage   openAIUsage      `json:"usage"`
	Error   *openAIError     `json:"error,omitempty"`
}

// openAIChoice represents a choice in the response
type openAIChoice struct {
	Index        int            `json:"index"`
	Message      openAIMessage  `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

// openAIUsage represents token usage
type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// openAIError represents an error response
type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// GetName returns the provider name
func (p *AlibabaProvider) GetName() string {
	return "alibaba"
}

// SendMessage sends a message to Alibaba Bailian
func (p *AlibabaProvider) SendMessage(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (*LLMResponse, error) {
	messages := []LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}
	return p.SendMessageWithHistory(ctx, "", messages, maxTokens)
}

// SendMessageWithHistory sends a message with conversation history
func (p *AlibabaProvider) SendMessageWithHistory(ctx context.Context, systemPrompt string, messages []LLMMessage, maxTokens int) (*LLMResponse, error) {
	// Build messages in OpenAI format
	openAIMessages := make([]openAIMessage, 0, len(messages)+1)

	// Add system prompt as first message if provided
	if systemPrompt != "" {
		openAIMessages = append(openAIMessages, openAIMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Add conversation messages
	for _, m := range messages {
		openAIMessages = append(openAIMessages, openAIMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Use OpenAI-compatible API format (works with Bailian Coding)
	reqBody := openAIChatRequest{
		Model:       p.model,
		Messages:    openAIMessages,
		MaxTokens:   maxTokens,
		Temperature: 0.7,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use OpenAI-compatible endpoint
	url := p.baseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
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

	var result openAIChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for error in response
	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", result.Error.Code, result.Error.Message)
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