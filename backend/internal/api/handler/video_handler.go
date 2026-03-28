package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/internal/service"
	"storyflow/pkg/ai"
)

// VideoHandler handles video generation requests
type VideoHandler struct {
	repo         *repository.StoryRepository
	aiFactory    *service.AIServiceFactory
	mergeService *service.VideoMergeService
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(repo *repository.StoryRepository, aiFactory *service.AIServiceFactory) *VideoHandler {
	return &VideoHandler{
		repo:         repo,
		aiFactory:    aiFactory,
		mergeService: service.NewVideoMergeService(repo, "./uploads/videos"),
	}
}

// VideoGenerateRequest represents a video generation request
type VideoGenerateRequest struct {
	StoryID     string  `json:"story_id" binding:"required"`
	SceneID     string  `json:"scene_id"`
	Duration    float64 `json:"duration"`
	Prompt      string  `json:"prompt"`
	MotionLevel string  `json:"motion_level"`
}

// VideoTaskResponse represents video task response
type VideoTaskResponse struct {
	TaskID   string `json:"task_id"`
	SceneID  string `json:"scene_id,omitempty"`
	Status   string `json:"status"`
	VideoURL string `json:"video_url,omitempty"`
	Progress int    `json:"progress"`
	Message  string `json:"message,omitempty"`
}

// GenerateVideo handles POST /api/v1/videos/generate
func (h *VideoHandler) GenerateVideo(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req VideoGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storyID, err := uuid.Parse(req.StoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Get story with scenes (verify ownership)
	story, err := h.repo.GetWithRelationsForUser(c.Request.Context(), userID, storyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	// Get video generator
	videoGenerator, err := h.aiFactory.GetVideoGenerator(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video generator not configured"})
		return
	}

	// Find the scene to generate video for
	var targetScene *model.Scene
	if req.SceneID != "" {
		sceneID, _ := uuid.Parse(req.SceneID)
		for i := range story.Scenes {
			if story.Scenes[i].ID == sceneID && story.Scenes[i].ImageURL != "" {
				targetScene = &story.Scenes[i]
				break
			}
		}
	} else {
		for i := range story.Scenes {
			if story.Scenes[i].ImageURL != "" {
				targetScene = &story.Scenes[i]
				break
			}
		}
	}

	if targetScene == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no scene with image found"})
		return
	}

	// Generate video
	result, err := videoGenerator.GenerateFromImage(c.Request.Context(), &ai.VideoRequest{
		ImageURL:    targetScene.ImageURL,
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		MotionLevel: req.MotionLevel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start async processing
	sceneID := targetScene.ID
	go h.processSingleVideo(userID, result.TaskID, sceneID)

	c.JSON(http.StatusAccepted, VideoTaskResponse{
		TaskID:  result.TaskID,
		SceneID: sceneID.String(),
		Status:  result.Status,
		Message: "Video generation started",
	})
}

// processSingleVideo waits for video completion and updates scene
func (h *VideoHandler) processSingleVideo(userID uuid.UUID, taskID string, sceneID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	videoGenerator, err := h.aiFactory.GetVideoGenerator(ctx, userID)
	if err != nil {
		fmt.Printf("Error getting video generator: %v\n", err)
		return
	}

	for i := 0; i < 120; i++ {
		time.Sleep(5 * time.Second)

		result, err := videoGenerator.GetTaskStatus(ctx, taskID)
		if err != nil {
			fmt.Printf("Error checking video status: %v\n", err)
			continue
		}

		if result.Status == "completed" {
			scene, err := h.repo.GetScene(ctx, sceneID)
			if err != nil {
				fmt.Printf("Error getting scene: %v\n", err)
				return
			}
			scene.VideoURL = result.VideoURL
			if err := h.repo.UpdateScene(ctx, scene); err != nil {
				fmt.Printf("Error updating scene: %v\n", err)
			}
			fmt.Printf("Video completed for scene %s\n", sceneID)
			return
		}

		if result.Status == "failed" {
			fmt.Printf("Video generation failed for scene %s: %s\n", sceneID, result.Error)
			return
		}
	}
}

// GenerateAllVideos handles POST /api/v1/videos/batch
func (h *VideoHandler) GenerateAllVideos(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req VideoGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storyID, err := uuid.Parse(req.StoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	story, err := h.repo.GetWithRelationsForUser(c.Request.Context(), userID, storyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	// Filter scenes with images but no video
	var scenesToGenerate []model.Scene
	var scenesWithVideo int
	for _, scene := range story.Scenes {
		if scene.ImageURL != "" {
			if scene.VideoURL != "" {
				scenesWithVideo++
			} else {
				scenesToGenerate = append(scenesToGenerate, scene)
			}
		}
	}

	if len(scenesToGenerate) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":            "所有场景已有视频",
			"total_scenes":       len(story.Scenes),
			"scenes_with_video":  scenesWithVideo,
		})
		return
	}

	// Start async batch video generation
	go h.processBatchVideos(userID, story.ID, scenesToGenerate, req)

	c.JSON(http.StatusAccepted, gin.H{
		"message":           "视频生成任务已启动",
		"total_to_generate": len(scenesToGenerate),
		"already_has_video": scenesWithVideo,
		"total_scenes":      len(story.Scenes),
	})
}

// processBatchVideos processes batch video generation
func (h *VideoHandler) processBatchVideos(userID uuid.UUID, storyID uuid.UUID, scenes []model.Scene, req VideoGenerateRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	videoGenerator, err := h.aiFactory.GetVideoGenerator(ctx, userID)
	if err != nil {
		fmt.Printf("[Video Batch] Error getting video generator: %v\n", err)
		return
	}

	for _, scene := range scenes {
		fmt.Printf("[Video Batch] Generating video for scene %d...\n", scene.Sequence)

		result, err := videoGenerator.GenerateFromImage(ctx, &ai.VideoRequest{
			ImageURL:    scene.ImageURL,
			Prompt:      req.Prompt,
			Duration:    req.Duration,
			MotionLevel: req.MotionLevel,
		})

		if err != nil {
			fmt.Printf("[Video Batch] ERROR at scene %d: %v\n", scene.Sequence, err)
			return
		}

		var videoURL string
		for i := 0; i < 60; i++ {
			time.Sleep(5 * time.Second)
			status, err := videoGenerator.GetTaskStatus(ctx, result.TaskID)
			if err != nil {
				continue
			}

			if status.Status == "completed" {
				videoURL = status.VideoURL
				break
			}

			if status.Status == "failed" {
				fmt.Printf("[Video Batch] Video failed for scene %d: %s\n", scene.Sequence, status.Error)
				return
			}
		}

		if videoURL == "" {
			fmt.Printf("[Video Batch] Timeout for scene %d\n", scene.Sequence)
			return
		}

		scene.VideoURL = videoURL
		if err := h.repo.UpdateScene(ctx, &scene); err != nil {
			fmt.Printf("[Video Batch] Error updating scene %d: %v\n", scene.Sequence, err)
			return
		}

		fmt.Printf("[Video Batch] Scene %d video completed\n", scene.Sequence)
	}

	fmt.Printf("[Video Batch] All videos completed for story %s\n", storyID)
}

// GetVideoStatus handles GET /api/v1/videos/status/:task_id
func (h *VideoHandler) GetVideoStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	taskID := c.Param("task_id")

	videoGenerator, err := h.aiFactory.GetVideoGenerator(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video generator not configured"})
		return
	}

	result, err := videoGenerator.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, VideoTaskResponse{
		TaskID:   result.TaskID,
		Status:   result.Status,
		VideoURL: result.VideoURL,
		Progress: result.Progress,
		Message:  result.Error,
	})
}

// MergeVideosRequest represents a video merge request
type MergeVideosRequest struct {
	StoryID           string  `json:"story_id" binding:"required"`
	Transition        string  `json:"transition"`
	TransitionDuration float64 `json:"transition_duration"`
	AddAudio          bool    `json:"add_audio"`
}

// MergeVideos handles POST /api/v1/videos/merge
func (h *VideoHandler) MergeVideos(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req MergeVideosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storyID, err := uuid.Parse(req.StoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Verify ownership
	_, err = h.repo.GetByUserAndID(c.Request.Context(), userID, storyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	options := service.MergeOptions{
		Transition:          req.Transition,
		TransitionDuration:  req.TransitionDuration,
		AddAudio:            req.AddAudio,
	}

	result, err := h.mergeService.MergeStoryVideos(c.Request.Context(), storyID, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"output_url":     result.OutputURL,
		"total_scenes":   result.TotalScenes,
		"total_duration": result.TotalDuration,
		"message":        fmt.Sprintf("成功合并 %d 个场景视频，总时长 %.1f 秒", result.TotalScenes, result.TotalDuration),
	})
}

// ViewVideo handles GET /api/v1/videos/view
func (h *VideoHandler) ViewVideo(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file parameter required"})
		return
	}

	filePath := filepath.Join("./uploads/videos", filename)
	if !filepath.IsLocal(filePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.FileAttachment(filePath, filename)
}