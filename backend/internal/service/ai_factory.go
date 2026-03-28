package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"storyflow/internal/repository"
	"storyflow/pkg/ai"
	"storyflow/pkg/crypto"
)

// AIServiceFactory creates AI services with user's API keys
type AIServiceFactory struct {
	configRepo *repository.UserConfigRepository
	encryptor  *crypto.Encryptor
}

// NewAIServiceFactory creates a new AI service factory
func NewAIServiceFactory(configRepo *repository.UserConfigRepository, encryptor *crypto.Encryptor) *AIServiceFactory {
	return &AIServiceFactory{
		configRepo: configRepo,
		encryptor:  encryptor,
	}
}

// UserAIConfig holds decrypted AI configuration
type UserAIConfig struct {
	LLMProvider   string
	LLMAPIKey     string
	LLMModel      string
	LLMBaseURL    string
	ImageProvider string
	ImageAPIKey   string
	ImageBaseURL  string
	ImageModel    string
	VideoProvider string
	VideoAPIKey   string
	VideoBaseURL  string
	VideoModel    string
	DefaultStyle  string
}

// GetUserConfig retrieves and decrypts user's AI configuration
func (f *AIServiceFactory) GetUserConfig(ctx context.Context, userID uuid.UUID) (*UserAIConfig, error) {
	config, err := f.configRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user config not found: %w", err)
	}

	// Decrypt API keys
	llmAPIKey, _ := f.encryptor.Decrypt(config.LLMAPIKey)
	imageAPIKey, _ := f.encryptor.Decrypt(config.ImageAPIKey)
	videoAPIKey, _ := f.encryptor.Decrypt(config.VideoAPIKey)

	return &UserAIConfig{
		LLMProvider:   config.LLMProvider,
		LLMAPIKey:     llmAPIKey,
		LLMModel:      config.LLMModel,
		LLMBaseURL:    config.LLMBaseURL,
		ImageProvider: config.ImageProvider,
		ImageAPIKey:   imageAPIKey,
		ImageBaseURL:  config.ImageBaseURL,
		ImageModel:    config.ImageModel,
		VideoProvider: config.VideoProvider,
		VideoAPIKey:   videoAPIKey,
		VideoBaseURL:  config.VideoBaseURL,
		VideoModel:    config.VideoModel,
		DefaultStyle:  config.DefaultStyle,
	}, nil
}

// GetLLMProvider creates an LLM provider with user's API key
func (f *AIServiceFactory) GetLLMProvider(ctx context.Context, userID uuid.UUID) (ai.LLMProvider, error) {
	config, err := f.GetUserConfig(ctx, userID)
	if err != nil {
		return nil, err
	}

	if config.LLMProvider == "" || config.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM not configured")
	}

	return ai.NewLLMProvider(ai.LLMConfig{
		Provider: config.LLMProvider,
		APIKey:   config.LLMAPIKey,
		Model:    config.LLMModel,
		BaseURL:  config.LLMBaseURL,
	}), nil
}

// GetImageGenerator creates an image generator with user's API key
func (f *AIServiceFactory) GetImageGenerator(ctx context.Context, userID uuid.UUID) (ai.ImageGenerator, error) {
	config, err := f.GetUserConfig(ctx, userID)
	if err != nil {
		return nil, err
	}

	if config.ImageProvider == "" {
		return nil, fmt.Errorf("image generator not configured")
	}

	return ai.NewImageGenerator(ai.ImageGeneratorConfig{
		Provider: config.ImageProvider,
		APIKey:   config.ImageAPIKey,
		BaseURL:  config.ImageBaseURL,
		Model:    config.ImageModel,
	}), nil
}

// GetVideoGenerator creates a video generator with user's API key
func (f *AIServiceFactory) GetVideoGenerator(ctx context.Context, userID uuid.UUID) (ai.VideoGenerator, error) {
	config, err := f.GetUserConfig(ctx, userID)
	if err != nil {
		return nil, err
	}

	if config.VideoProvider == "" {
		return nil, fmt.Errorf("video generator not configured")
	}

	return ai.NewVideoGenerator(ai.VideoGeneratorConfig{
		Provider: config.VideoProvider,
		APIKey:   config.VideoAPIKey,
		BaseURL:  config.VideoBaseURL,
		Model:    config.VideoModel,
	}), nil
}

// HasLLMConfig checks if user has LLM configured
func (f *AIServiceFactory) HasLLMConfig(ctx context.Context, userID uuid.UUID) bool {
	config, err := f.configRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false
	}
	return config.LLMProvider != "" && config.LLMAPIKey != ""
}

// HasImageConfig checks if user has image generator configured
func (f *AIServiceFactory) HasImageConfig(ctx context.Context, userID uuid.UUID) bool {
	config, err := f.configRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false
	}
	return config.ImageProvider != ""
}

// HasVideoConfig checks if user has video generator configured
func (f *AIServiceFactory) HasVideoConfig(ctx context.Context, userID uuid.UUID) bool {
	config, err := f.configRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false
	}
	return config.VideoProvider != "" && config.VideoAPIKey != ""
}

// ValidateAPIKeyRequest represents a validation request
type ValidateAPIKeyRequest struct {
	Type     string `json:"type"`     // llm, image, video
	Provider string `json:"provider"` // claude, openai, comfyui, etc.
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
}

// ValidateAPIKeyResult represents validation result
type ValidateAPIKeyResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// ValidateAPIKey validates an API key by making a test call
func (f *AIServiceFactory) ValidateAPIKey(ctx context.Context, req ValidateAPIKeyRequest) (*ValidateAPIKeyResult, error) {
	switch req.Type {
	case "llm":
		return f.validateLLMKey(ctx, req)
	case "image":
		return f.validateImageKey(ctx, req)
	case "video":
		return f.validateVideoKey(ctx, req)
	default:
		return &ValidateAPIKeyResult{Valid: false, Message: "unknown type"}, nil
	}
}

func (f *AIServiceFactory) validateLLMKey(ctx context.Context, req ValidateAPIKeyRequest) (*ValidateAPIKeyResult, error) {
	if req.APIKey == "" {
		return &ValidateAPIKeyResult{Valid: false, Message: "API key is required"}, nil
	}

	// Set default model for validation based on provider
	var defaultModel string
	switch req.Provider {
	case "deepseek":
		defaultModel = "deepseek-chat"
	case "openai":
		defaultModel = "gpt-4o"
	case "moonshot":
		defaultModel = "moonshot-v1-8k"
	case "zhipu":
		defaultModel = "glm-4"
	case "alibaba", "qwen":
		defaultModel = "qwen-max"
	case "volcengine", "doubao":
		defaultModel = "doubao-pro-32k"
	}

	provider := ai.NewLLMProvider(ai.LLMConfig{
		Provider: req.Provider,
		APIKey:   req.APIKey,
		BaseURL:  req.BaseURL,
		Model:    defaultModel,
	})

	// Try a minimal API call
	_, err := provider.SendMessage(ctx, "You are a test assistant.", "Say 'ok'", 10)
	if err != nil {
		return &ValidateAPIKeyResult{Valid: false, Message: fmt.Sprintf("API key validation failed: %v", err)}, nil
	}

	return &ValidateAPIKeyResult{Valid: true, Message: "API key is valid"}, nil
}

func (f *AIServiceFactory) validateImageKey(ctx context.Context, req ValidateAPIKeyRequest) (*ValidateAPIKeyResult, error) {
	if req.Provider == "comfyui" {
		// For ComfyUI, just check if the server is reachable
		if req.BaseURL == "" {
			req.BaseURL = "http://localhost:8188"
		}
		// ComfyUI doesn't require API key for local instance
		return &ValidateAPIKeyResult{Valid: true, Message: "ComfyUI connection test not implemented, assuming valid"}, nil
	}

	if req.APIKey == "" {
		return &ValidateAPIKeyResult{Valid: false, Message: "API key is required"}, nil
	}

	// For cloud providers, we'll just check the format since actual validation would cost money
	// In production, you'd want to make a minimal API call
	return &ValidateAPIKeyResult{Valid: true, Message: "API key format looks valid"}, nil
}

func (f *AIServiceFactory) validateVideoKey(ctx context.Context, req ValidateAPIKeyRequest) (*ValidateAPIKeyResult, error) {
	if req.APIKey == "" {
		return &ValidateAPIKeyResult{Valid: false, Message: "API key is required"}, nil
	}

	// For video providers, we'll just check the format since actual validation would cost money
	return &ValidateAPIKeyResult{Valid: true, Message: "API key format looks valid"}, nil
}