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

// OpenAICompatibleProvider implements LLMProvider for OpenAI-compatible APIs
// Works with: OpenAI, DeepSeek, Moonshot, Zhipu, and other OpenAI-compatible providers
type OpenAICompatibleProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
	provider   string
}

// NewOpenAICompatibleProvider creates a new OpenAI-compatible provider
func NewOpenAICompatibleProvider(cfg LLMConfig, providerName string) *OpenAICompatibleProvider {
	if cfg.Model == "" {
		switch providerName {
		case "deepseek":
			cfg.Model = "deepseek-chat"
		case "openai":
			cfg.Model = "gpt-4o"
		case "moonshot":
			cfg.Model = "moonshot-v1-8k"
		case "zhipu":
			cfg.Model = "glm-4"
		default:
			cfg.Model = "gpt-4o"
		}
	}

	// Strip provider prefix from model name if present (e.g., "deepseek/deepseek-chat" -> "deepseek-chat")
	cfg.Model = stripProviderPrefix(cfg.Model, providerName)

	if cfg.BaseURL == "" {
		switch providerName {
		case "deepseek":
			cfg.BaseURL = "https://api.deepseek.com/v1"
		case "openai":
			cfg.BaseURL = "https://api.openai.com/v1"
		case "moonshot":
			cfg.BaseURL = "https://api.moonshot.cn/v1"
		case "zhipu":
			cfg.BaseURL = "https://open.bigmodel.cn/api/paas/v4"
		default:
			cfg.BaseURL = "https://api.openai.com/v1"
		}
	}

	return &OpenAICompatibleProvider{
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		provider: providerName,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
		baseURL: cfg.BaseURL,
	}
}

// stripProviderPrefix removes provider prefix from model name
// e.g., "deepseek/deepseek-chat" -> "deepseek-chat"
func stripProviderPrefix(model, provider string) string {
	prefix := provider + "/"
	if len(model) > len(prefix) && model[:len(prefix)] == prefix {
		return model[len(prefix):]
	}
	return model
}

// GetName returns the provider name
func (p *OpenAICompatibleProvider) GetName() string {
	return p.provider
}

// SendMessage sends a message to the API
func (p *OpenAICompatibleProvider) SendMessage(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (*LLMResponse, error) {
	messages := []LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}
	return p.SendMessageWithHistory(ctx, "", messages, maxTokens)
}

// SendMessageWithHistory sends a message with conversation history
func (p *OpenAICompatibleProvider) SendMessageWithHistory(ctx context.Context, systemPrompt string, messages []LLMMessage, maxTokens int) (*LLMResponse, error) {
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
		if m.Role != "" && m.Content != "" {
			openAIMessages = append(openAIMessages, openAIMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	// Build request
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
	// Handle base URLs that may or may not include /v1
	baseURL := p.baseURL
	if !endsWithV1(baseURL) {
		baseURL = baseURL + "/v1"
	}
	url := baseURL + "/chat/completions"

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

// endsWithV1 checks if URL ends with /v1 or /v1/
func endsWithV1(url string) bool {
	return len(url) >= 3 && (url[len(url)-3:] == "/v1" || len(url) >= 4 && url[len(url)-4:] == "/v1/")
}