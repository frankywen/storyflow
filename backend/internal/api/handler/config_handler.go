package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/internal/service"
	"storyflow/pkg/crypto"
)

// ConfigHandler handles user configuration requests
type ConfigHandler struct {
	configRepo *repository.UserConfigRepository
	encryptor  *crypto.Encryptor
	aiFactory  *service.AIServiceFactory
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(configRepo *repository.UserConfigRepository, encryptor *crypto.Encryptor, aiFactory *service.AIServiceFactory) *ConfigHandler {
	return &ConfigHandler{
		configRepo: configRepo,
		encryptor:  encryptor,
		aiFactory:  aiFactory,
	}
}

// GetConfig handles GET /api/v1/user/config
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	config, err := h.configRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		// Return empty config if not found
		c.JSON(http.StatusOK, gin.H{
			"config": &model.UserConfig{
				UserID: userID,
			},
		})
		return
	}

	// Mask API keys
	response := h.maskConfig(config)
	c.JSON(http.StatusOK, gin.H{"config": response})
}

// UpdateConfigInput represents config update input
type UpdateConfigInput struct {
	// LLM
	LLMProvider string `json:"llm_provider"`
	LLMAPIKey   string `json:"llm_api_key"`
	LLMModel    string `json:"llm_model"`
	LLMBaseURL  string `json:"llm_base_url"`
	// Image
	ImageProvider string `json:"image_provider"`
	ImageAPIKey   string `json:"image_api_key"`
	ImageBaseURL  string `json:"image_base_url"`
	ImageModel    string `json:"image_model"`
	// Video
	VideoProvider string `json:"video_provider"`
	VideoAPIKey   string `json:"video_api_key"`
	VideoBaseURL  string `json:"video_base_url"`
	VideoModel    string `json:"video_model"`
	// Preferences
	DefaultStyle string `json:"default_style"`
}

// UpdateConfig handles PUT /api/v1/user/config
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var input UpdateConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Encrypt API keys
	llmAPIKey, _ := h.encryptor.Encrypt(input.LLMAPIKey)
	imageAPIKey, _ := h.encryptor.Encrypt(input.ImageAPIKey)
	videoAPIKey, _ := h.encryptor.Encrypt(input.VideoAPIKey)

	config := &model.UserConfig{
		UserID:        userID,
		LLMProvider:   input.LLMProvider,
		LLMAPIKey:     llmAPIKey,
		LLMModel:      input.LLMModel,
		LLMBaseURL:    input.LLMBaseURL,
		ImageProvider: input.ImageProvider,
		ImageAPIKey:   imageAPIKey,
		ImageBaseURL:  input.ImageBaseURL,
		ImageModel:    input.ImageModel,
		VideoProvider: input.VideoProvider,
		VideoAPIKey:   videoAPIKey,
		VideoBaseURL:  input.VideoBaseURL,
		VideoModel:    input.VideoModel,
		DefaultStyle:  input.DefaultStyle,
	}

	if err := h.configRepo.CreateOrUpdate(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}

	response := h.maskConfig(config)
	c.JSON(http.StatusOK, gin.H{"config": response})
}

// UpdateLLMConfig handles PUT /api/v1/user/config/llm
func (h *ConfigHandler) UpdateLLMConfig(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var input struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		Model    string `json:"model"`
		BaseURL  string `json:"base_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing config or create new
	config, err := h.configRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		config = &model.UserConfig{UserID: userID}
	}

	config.LLMProvider = input.Provider
	config.LLMAPIKey, _ = h.encryptor.Encrypt(input.APIKey)
	config.LLMModel = input.Model
	config.LLMBaseURL = input.BaseURL

	if err := h.configRepo.CreateOrUpdate(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "LLM config updated",
		"masked_key": crypto.MaskKey(input.APIKey),
	})
}

// UpdateImageConfig handles PUT /api/v1/user/config/image
func (h *ConfigHandler) UpdateImageConfig(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var input struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		BaseURL  string `json:"base_url"`
		Model    string `json:"model"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.configRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		config = &model.UserConfig{UserID: userID}
	}

	config.ImageProvider = input.Provider
	config.ImageAPIKey, _ = h.encryptor.Encrypt(input.APIKey)
	config.ImageBaseURL = input.BaseURL
	config.ImageModel = input.Model

	if err := h.configRepo.CreateOrUpdate(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image config updated",
		"masked_key": crypto.MaskKey(input.APIKey),
	})
}

// UpdateVideoConfig handles PUT /api/v1/user/config/video
func (h *ConfigHandler) UpdateVideoConfig(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var input struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		BaseURL  string `json:"base_url"`
		Model    string `json:"model"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.configRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		config = &model.UserConfig{UserID: userID}
	}

	config.VideoProvider = input.Provider
	config.VideoAPIKey, _ = h.encryptor.Encrypt(input.APIKey)
	config.VideoBaseURL = input.BaseURL
	config.VideoModel = input.Model

	if err := h.configRepo.CreateOrUpdate(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Video config updated",
		"masked_key": crypto.MaskKey(input.APIKey),
	})
}

// maskConfig masks API keys in config for response
func (h *ConfigHandler) maskConfig(config *model.UserConfig) map[string]interface{} {
	return map[string]interface{}{
		"user_id":        config.UserID,
		"llm_provider":   config.LLMProvider,
		"llm_api_key":    crypto.MaskKey(config.LLMAPIKey),
		"llm_model":      config.LLMModel,
		"llm_base_url":   config.LLMBaseURL,
		"image_provider": config.ImageProvider,
		"image_api_key":  crypto.MaskKey(config.ImageAPIKey),
		"image_base_url": config.ImageBaseURL,
		"image_model":    config.ImageModel,
		"video_provider": config.VideoProvider,
		"video_api_key":  crypto.MaskKey(config.VideoAPIKey),
		"video_base_url": config.VideoBaseURL,
		"video_model":    config.VideoModel,
		"default_style":  config.DefaultStyle,
	}
}

// ValidateAPIKeyInput represents API key validation input
type ValidateAPIKeyInput struct {
	Type     string `json:"type" binding:"required"`     // llm, image, video
	Provider string `json:"provider" binding:"required"` // claude, openai, comfyui, etc.
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
}

// ValidateAPIKey handles POST /api/v1/user/config/validate
func (h *ConfigHandler) ValidateAPIKey(c *gin.Context) {
	var input ValidateAPIKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.aiFactory.ValidateAPIKey(c.Request.Context(), service.ValidateAPIKeyRequest{
		Type:     input.Type,
		Provider: input.Provider,
		APIKey:   input.APIKey,
		BaseURL:  input.BaseURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   result.Valid,
		"message": result.Message,
	})
}