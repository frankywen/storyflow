package ai

import (
	"context"
)

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	// SendMessage sends a message and returns the response
	SendMessage(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (*LLMResponse, error)

	// SendMessageWithHistory sends a message with conversation history
	SendMessageWithHistory(ctx context.Context, systemPrompt string, messages []LLMMessage, maxTokens int) (*LLMResponse, error)

	// GetName returns the provider name
	GetName() string
}

// LLMMessage represents a message in the conversation
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents a generic LLM response
type LLMResponse struct {
	ID           string
	Model        string
	Content      string
	InputTokens  int
	OutputTokens int
	StopReason   string
}

// LLMConfig holds configuration for LLM providers
type LLMConfig struct {
	Provider  string // "claude", "volcengine", "alibaba"
	APIKey    string
	Model     string
	BaseURL   string // optional, for custom endpoints
	MaxTokens int    // default max tokens
}

// NewLLMProvider creates an LLM provider based on config
func NewLLMProvider(cfg LLMConfig) LLMProvider {
	switch cfg.Provider {
	case "claude", "anthropic":
		return NewClaudeProvider(cfg)
	case "volcengine", "doubao":
		return NewVolcEngineProvider(cfg)
	case "alibaba", "qwen", "tongyi":
		return NewAlibabaProvider(cfg)
	case "deepseek", "openai", "moonshot", "zhipu":
		return NewOpenAICompatibleProvider(cfg, cfg.Provider)
	default:
		// Try OpenAI-compatible as fallback for unknown providers
		return NewOpenAICompatibleProvider(cfg, cfg.Provider)
	}
}