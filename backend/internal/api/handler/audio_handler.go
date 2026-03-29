package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storyflow/internal/repository"
	"storyflow/internal/service"
)

type AudioHandler struct {
	audioService           *service.AudioService
	subtitleService        *service.SubtitleService
	videoSynthesisService  *service.VideoSynthesisService
	storyRepo              *repository.StoryRepository
}

func NewAudioHandler(
	audioService *service.AudioService,
	subtitleService *service.SubtitleService,
	videoSynthesisService *service.VideoSynthesisService,
	storyRepo *repository.StoryRepository,
) *AudioHandler {
	return &AudioHandler{
		audioService:          audioService,
		subtitleService:       subtitleService,
		videoSynthesisService: videoSynthesisService,
		storyRepo:             storyRepo,
	}
}

// verifyStoryOwnership verifies the user owns the story
func (h *AudioHandler) verifyStoryOwnership(c *gin.Context, storyID uuid.UUID) bool {
	userID := c.MustGet("user_id").(uuid.UUID)
	_, err := h.storyRepo.GetWithRelationsForUser(c.Request.Context(), userID, storyID)
	return err == nil
}

// GenerateAudio handles POST /api/v1/audio/generate
func (h *AudioHandler) GenerateAudio(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req service.GenerateAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user owns the story
	if _, err := h.storyRepo.GetWithRelationsForUser(c.Request.Context(), userID, req.StoryID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	task, err := h.audioService.GenerateAudio(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Verify user owns the story
	if !h.verifyStoryOwnership(c, storyID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Verify user owns the story
	if !h.verifyStoryOwnership(c, storyID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Verify user owns the story
	if !h.verifyStoryOwnership(c, storyID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	subtitles, err := h.subtitleService.GetSubtitlesByStory(c.Request.Context(), storyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subtitles": subtitles})
}

// SynthesizeVideoRequest represents the request body for video synthesis
type SynthesizeVideoRequest struct {
	StoryID     string `json:"story_id" binding:"required"`
	VideoURL    string `json:"video_url"`
	AddAudio    bool   `json:"add_audio"`
	AddSubtitle bool   `json:"add_subtitle"`
}

// GenerateSceneAudioRequest represents the request body for scene audio generation
type GenerateSceneAudioRequest struct {
	VoiceID   string `json:"voice_id"`
	Overwrite bool   `json:"overwrite"`
}

// SynthesizeVideo handles POST /api/v1/audio/synthesis
func (h *AudioHandler) SynthesizeVideo(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req SynthesizeVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse story ID
	storyID, err := uuid.Parse(req.StoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story_id"})
		return
	}

	// Verify user owns the story
	if _, err := h.storyRepo.GetWithRelationsForUser(c.Request.Context(), userID, storyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	// Create synthesis request
	synthesisReq := service.SynthesizeVideoRequest{
		StoryID:     storyID,
		VideoURL:    req.VideoURL,
		AddAudio:    req.AddAudio,
		AddSubtitle: req.AddSubtitle,
	}

	// Start video synthesis task
	task, err := h.videoSynthesisService.SynthesizeVideo(c.Request.Context(), synthesisReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"task_id": task.ID,
		"message": "Video synthesis task created",
	})
}

// GetSynthesisStatus handles GET /api/v1/videos/synthesis/:task_id
func (h *AudioHandler) GetSynthesisStatus(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id"})
		return
	}

	task, err := h.videoSynthesisService.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":       task.ID,
		"status":        task.Status,
		"progress":      task.Progress,
		"output_url":    task.OutputURL,
		"error_message": task.ErrorMessage,
	})
}

// GenerateSceneAudio handles POST /api/v1/audio/generate/scene/:scene_id
func (h *AudioHandler) GenerateSceneAudio(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	sceneIDStr := c.Param("scene_id")
	sceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scene_id"})
		return
	}

	var req GenerateSceneAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Request body is optional, use defaults
		req = GenerateSceneAudioRequest{}
	}

	// Verify user owns the story containing this scene
	scene, err := h.storyRepo.GetScene(c.Request.Context(), sceneID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	_, err = h.storyRepo.GetWithRelationsForUser(c.Request.Context(), userID, scene.StoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	// Generate audio
	audios, err := h.audioService.GenerateAudioForScene(c.Request.Context(), sceneID, req.VoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"scene_id": sceneID,
		"audios":   audios,
		"message":  fmt.Sprintf("Generated %d audio files", len(audios)),
	})
}

// GenerateSceneSubtitles handles POST /api/v1/audio/subtitles/scene/:scene_id
func (h *AudioHandler) GenerateSceneSubtitles(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	sceneIDStr := c.Param("scene_id")
	sceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scene_id"})
		return
	}

	// Verify user owns the story containing this scene
	scene, err := h.storyRepo.GetScene(c.Request.Context(), sceneID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	_, err = h.storyRepo.GetWithRelationsForUser(c.Request.Context(), userID, scene.StoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	// Generate subtitles
	subtitles, err := h.subtitleService.GenerateSubtitlesForScene(c.Request.Context(), sceneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"scene_id":  sceneID,
		"subtitles": subtitles,
		"message":   fmt.Sprintf("Generated %d subtitles", len(subtitles)),
	})
}