package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storyflow/internal/repository"
	"storyflow/internal/service"
)

// StoryHandler handles story-related HTTP requests
type StoryHandler struct {
	repo      *repository.StoryRepository
	aiFactory *service.AIServiceFactory
}

// NewStoryHandler creates a new story handler
func NewStoryHandler(repo *repository.StoryRepository, aiFactory *service.AIServiceFactory) *StoryHandler {
	return &StoryHandler{repo: repo, aiFactory: aiFactory}
}

// CreateStoryRequest represents the request body for creating a story
type CreateStoryRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// ParseStoryRequest represents the request body for parsing a story
type ParseStoryRequest struct {
	Style string `json:"style"` // manga, manhwa, western_comic, realistic, anime
}

// CreateStory handles POST /api/v1/stories
func (h *StoryHandler) CreateStory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storyService := service.NewStoryService(h.repo, h.aiFactory)
	story, err := storyService.CreateStory(c.Request.Context(), userID, req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         story.ID,
		"title":      story.Title,
		"status":     story.Status,
		"created_at": story.CreatedAt,
	})
}

// GetStory handles GET /api/v1/stories/:id
func (h *StoryHandler) GetStory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	story, err := h.repo.GetWithRelationsForUser(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	c.JSON(http.StatusOK, story)
}

// ListStories handles GET /api/v1/stories (with filters)
func (h *StoryHandler) ListStories(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// Filter parameters
	status := c.Query("status")
	search := c.Query("search")

	offset := (page - 1) * pageSize
	stories, total, err := h.repo.ListByUserWithFilters(c.Request.Context(), userID, offset, pageSize, status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      stories,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ParseStory handles POST /api/v1/stories/:id/parse
func (h *StoryHandler) ParseStory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	var req ParseStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Style = "manga" // default style
	}

	storyService := service.NewStoryService(h.repo, h.aiFactory)
	story, err := storyService.ParseStory(c.Request.Context(), userID, id, req.Style)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         story.ID,
		"title":      story.Title,
		"summary":    story.Summary,
		"genre":      story.Genre,
		"status":     story.Status,
		"characters": story.Characters,
		"scenes":     story.Scenes,
	})
}

// GetCharacters handles GET /api/v1/stories/:id/characters
func (h *StoryHandler) GetCharacters(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Verify story belongs to user
	_, err = h.repo.GetByUserAndID(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	characters, err := h.repo.GetCharactersByStoryID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"characters": characters,
	})
}

// GetScenes handles GET /api/v1/stories/:id/scenes
func (h *StoryHandler) GetScenes(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Verify story belongs to user
	_, err = h.repo.GetByUserAndID(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	scenes, err := h.repo.GetScenesByStoryID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scenes": scenes,
	})
}

// DeleteStory handles DELETE /api/v1/stories/:id (soft delete)
func (h *StoryHandler) DeleteStory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	if err := h.repo.SoftDeleteStory(c.Request.Context(), userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "story deleted"})
}

// RestoreStory handles POST /api/v1/stories/:id/restore
func (h *StoryHandler) RestoreStory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	if err := h.repo.RestoreStory(c.Request.Context(), userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "story restored"})
}

// UpdateStoryRequest represents the request body for updating a story
type UpdateStoryRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// UpdateStory handles PUT /api/v1/stories/:id
func (h *StoryHandler) UpdateStory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	var req UpdateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing story
	story, err := h.repo.GetByUserAndID(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	// Update fields
	if req.Title != "" {
		story.Title = req.Title
	}
	if req.Content != "" {
		story.Content = req.Content
		// Reset status to draft when content changes, so user can re-parse
		if story.Status != "draft" {
			story.Status = "draft"
		}
	}

	if err := h.repo.Update(c.Request.Context(), story); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated story with relations
	updatedStory, _ := h.repo.GetWithRelationsForUser(c.Request.Context(), userID, id)
	c.JSON(http.StatusOK, updatedStory)
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}