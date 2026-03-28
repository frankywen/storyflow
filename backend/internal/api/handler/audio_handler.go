package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storyflow/internal/service"
)

type AudioHandler struct {
	audioService    *service.AudioService
	subtitleService *service.SubtitleService
}

func NewAudioHandler(audioService *service.AudioService, subtitleService *service.SubtitleService) *AudioHandler {
	return &AudioHandler{
		audioService:    audioService,
		subtitleService: subtitleService,
	}
}

// GenerateAudio handles POST /api/v1/audio/generate
func (h *AudioHandler) GenerateAudio(c *gin.Context) {
	var req service.GenerateAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.audioService.GenerateAudio(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"task_id": task.ID,
		"message": "Audio generation task created",
	})
}

// GetTaskStatus handles GET /api/v1/audio/status/:task_id
func (h *AudioHandler) GetTaskStatus(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id"})
		return
	}

	task, err := h.audioService.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":           task.ID,
		"status":            task.Status,
		"progress":          task.Progress,
		"total_scenes":      task.TotalScenes,
		"completed_scenes":  task.CompletedScenes,
		"failed_scenes":     task.FailedScenes,
	})
}

// GetAudios handles GET /api/v1/audio/story/:story_id
func (h *AudioHandler) GetAudios(c *gin.Context) {
	storyIDStr := c.Param("story_id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story_id"})
		return
	}

	audios, err := h.audioService.GetAudiosByStory(c.Request.Context(), storyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audios": audios})
}

// GenerateSubtitles handles POST /api/v1/audio/subtitles/:story_id
func (h *AudioHandler) GenerateSubtitles(c *gin.Context) {
	storyIDStr := c.Param("story_id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story_id"})
		return
	}

	if err := h.subtitleService.GenerateSubtitles(c.Request.Context(), storyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subtitles generated",
	})
}

// GetSubtitles handles GET /api/v1/audio/subtitles/:story_id
func (h *AudioHandler) GetSubtitles(c *gin.Context) {
	storyIDStr := c.Param("story_id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story_id"})
		return
	}

	subtitles, err := h.subtitleService.GetSubtitlesByStory(c.Request.Context(), storyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subtitles": subtitles})
}