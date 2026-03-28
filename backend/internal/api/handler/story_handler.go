package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"storyflow/internal/service"
)

// StoryHandler handles story-related HTTP requests
type StoryHandler struct {
	service *service.StoryService
}

// NewStoryHandler creates a new story handler
func NewStoryHandler(service *service.StoryService) *StoryHandler {
	return &StoryHandler{service: service}
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

// StoryResponse represents a story response
type StoryResponse struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content,omitempty"`
	Summary    string    `json:"summary"`
	Genre      string    `json:"genre"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

// CreateStory handles POST /api/v1/stories
func (h *StoryHandler) CreateStory(c *gin.Context) {
	var req CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	story, err := h.service.CreateStory(c.Request.Context(), req.Title, req.Content)
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	story, err := h.service.GetStoryWithRelations(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	c.JSON(http.StatusOK, story)
}

// ListStories handles GET /api/v1/stories
func (h *StoryHandler) ListStories(c *gin.Context) {
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

	stories, total, err := h.service.ListStories(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       stories,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// ParseStory handles POST /api/v1/stories/:id/parse
func (h *StoryHandler) ParseStory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	var req ParseStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Style = "manga" // default style
	}

	story, err := h.service.ParseStory(c.Request.Context(), id, req.Style)
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	characters, err := h.service.GetCharacters(c.Request.Context(), id)
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	scenes, err := h.service.GetScenes(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scenes": scenes,
	})
}

// DeleteStory handles DELETE /api/v1/stories/:id
func (h *StoryHandler) DeleteStory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	if err := h.service.DeleteStory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "story deleted"})
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}