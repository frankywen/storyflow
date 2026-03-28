package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"storyflow/internal/repository"
	"storyflow/internal/service"
)

// CharacterHandler handles character-related HTTP requests
type CharacterHandler struct {
	consistencyService *service.CharacterConsistencyService
	repo               *repository.StoryRepository
}

// NewCharacterHandler creates a new character handler
func NewCharacterHandler(
	consistencyService *service.CharacterConsistencyService,
	repo *repository.StoryRepository,
) *CharacterHandler {
	return &CharacterHandler{
		consistencyService: consistencyService,
		repo:               repo,
	}
}

// RegenerateReferenceRequest represents the request for regenerating reference image
type RegenerateReferenceRequest struct {
	Style string `json:"style"` // manga, realistic, anime, etc.
}

// UploadReferenceImage handles POST /api/v1/characters/:id/reference
func (h *CharacterHandler) UploadReferenceImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	// Get character
	character, err := h.repo.GetCharacterByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type, only jpg, jpeg, png, webp are allowed"})
		return
	}

	// Create upload directory if not exists
	uploadDir := "uploads/references"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d%s", id.String(), time.Now().Unix(), ext)
	filepath := filepath.Join(uploadDir, filename)

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file"})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Update character with reference image URL
	// In production, this would be a cloud storage URL
	imageURL := fmt.Sprintf("/api/v1/images/view?url=/uploads/references/%s", filename)
	character.ReferenceImageURL = imageURL
	character.ReferenceImageID = filename

	if err := h.repo.UpdateCharacter(c.Request.Context(), character); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update character"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                  character.ID,
		"reference_image_url": character.ReferenceImageURL,
		"message":             "参考图上传成功",
	})
}

// DeleteReferenceImage handles DELETE /api/v1/characters/:id/reference
func (h *CharacterHandler) DeleteReferenceImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	// Get character
	character, err := h.repo.GetCharacterByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	// Delete file if exists
	if character.ReferenceImageID != "" {
		filepath := filepath.Join("uploads/references", character.ReferenceImageID)
		os.Remove(filepath) // Ignore error if file doesn't exist
	}

	// Clear reference image
	character.ReferenceImageURL = ""
	character.ReferenceImageID = ""

	if err := h.repo.UpdateCharacter(c.Request.Context(), character); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update character"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      character.ID,
		"message": "参考图已删除",
	})
}

// RegenerateReferenceImage handles POST /api/v1/characters/:id/regenerate
func (h *CharacterHandler) RegenerateReferenceImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character ID"})
		return
	}

	// Get character
	character, err := h.repo.GetCharacterByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}

	// Parse request
	var req RegenerateReferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Style = "manga" // default style
	}

	// Generate reference image
	updatedChar, err := h.consistencyService.GenerateReferenceImage(
		c.Request.Context(),
		character,
		req.Style,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "CHAR_REF_GEN_FAILED",
			"message": fmt.Sprintf("参考图生成失败: %v", err),
		})
		return
	}

	// Save to database
	if err := h.repo.UpdateCharacter(c.Request.Context(), updatedChar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update character"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                  updatedChar.ID,
		"reference_image_url": updatedChar.ReferenceImageURL,
		"visual_prompt":       updatedChar.VisualPrompt,
		"seed":                updatedChar.Seed,
		"message":             "参考图生成成功",
	})
}

// GenerateAllReferencesRequest represents the request for batch reference generation
type GenerateAllReferencesRequest struct {
	Style string `json:"style"`
}

// GenerateAllReferenceImages handles POST /api/v1/stories/:id/generate-references
func (h *CharacterHandler) GenerateAllReferenceImages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Check story exists and is parsed
	story, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	if story.Status != "parsed" && story.Status != "generating" && story.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "STORY_NOT_PARSED",
			"message": "故事未解析，无法生成参考图",
		})
		return
	}

	// Parse request
	var req GenerateAllReferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Style = "manga" // default style
	}

	// Generate all reference images
	updatedChars, err := h.consistencyService.GenerateAllReferenceImages(
		c.Request.Context(),
		id,
		req.Style,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "CHAR_REF_GEN_FAILED",
			"message": fmt.Sprintf("批量生成参考图失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"story_id":    id,
		"characters":  updatedChars,
		"total":       len(updatedChars),
		"message":     "角色参考图批量生成完成",
	})
}