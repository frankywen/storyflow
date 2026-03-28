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
	videoGenerator ai.VideoGenerator
	repo           *repository.StoryRepository
	mergeService   *service.VideoMergeService
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(videoGenerator ai.VideoGenerator, repo *repository.StoryRepository) *VideoHandler {
	return &VideoHandler{
		videoGenerator: videoGenerator,
		repo:           repo,
		mergeService:   service.NewVideoMergeService(repo, "./uploads/videos"),
	}
}

// VideoGenerateRequest represents a video generation request
type VideoGenerateRequest struct {
	StoryID     string  `json:"story_id" binding:"required"`
	SceneID     string  `json:"scene_id"`      // Optional: specific scene
	Duration    float64 `json:"duration"`      // Duration in seconds
	Prompt      string  `json:"prompt"`        // Optional additional prompt
	MotionLevel string  `json:"motion_level"`  // low, medium, high
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

	// Get story with scenes
	story, err := h.repo.GetWithRelations(c.Request.Context(), storyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
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
		// Use first scene with image
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
	result, err := h.videoGenerator.GenerateFromImage(c.Request.Context(), &ai.VideoRequest{
		ImageURL:    targetScene.ImageURL,
		Prompt:      req.Prompt,
		Duration:    req.Duration,
		MotionLevel: req.MotionLevel,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start async processing to wait for completion and update scene
	sceneID := targetScene.ID
	go h.processSingleVideo(result.TaskID, sceneID)

	c.JSON(http.StatusAccepted, VideoTaskResponse{
		TaskID:  result.TaskID,
		SceneID: sceneID.String(),
		Status:  result.Status,
		Message: "Video generation started",
	})
}

// processSingleVideo waits for video completion and updates scene
func (h *VideoHandler) processSingleVideo(taskID string, sceneID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Poll for completion
	for i := 0; i < 120; i++ { // 120 * 5s = 10 min max
		time.Sleep(5 * time.Second)

		result, err := h.videoGenerator.GetTaskStatus(ctx, taskID)
		if err != nil {
			fmt.Printf("Error checking video status: %v\n", err)
			continue
		}

		if result.Status == "completed" {
			// Update scene with video URL
			scene, err := h.repo.GetScene(ctx, sceneID)
			if err != nil {
				fmt.Printf("Error getting scene: %v\n", err)
				return
			}
			scene.VideoURL = result.VideoURL
			if err := h.repo.UpdateScene(ctx, scene); err != nil {
				fmt.Printf("Error updating scene: %v\n", err)
			}
			fmt.Printf("Video completed for scene %s: %s\n", sceneID, result.VideoURL)
			return
		}

		if result.Status == "failed" {
			fmt.Printf("Video generation failed for scene %s: %s\n", sceneID, result.Error)
			return
		}
	}

	fmt.Printf("Video generation timeout for scene %s\n", sceneID)
}

// GenerateAllVideos handles POST /api/v1/videos/batch
func (h *VideoHandler) GenerateAllVideos(c *gin.Context) {
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

	// Get story with scenes
	story, err := h.repo.GetWithRelations(c.Request.Context(), storyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	// Filter scenes with images but no video (skip already generated)
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
			"message":         "所有场景已有视频",
			"total_scenes":    len(story.Scenes),
			"scenes_with_video": scenesWithVideo,
		})
		return
	}

	// Start async batch video generation
	go h.processBatchVideos(story.ID, scenesToGenerate, req)

	c.JSON(http.StatusAccepted, gin.H{
		"message":            "视频生成任务已启动",
		"total_to_generate":  len(scenesToGenerate),
		"already_has_video":  scenesWithVideo,
		"total_scenes":       len(story.Scenes),
	})
}

// GetVideoStatus handles GET /api/v1/videos/status/:task_id
func (h *VideoHandler) GetVideoStatus(c *gin.Context) {
	taskID := c.Param("task_id")

	result, err := h.videoGenerator.GetTaskStatus(c.Request.Context(), taskID)
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

// processBatchVideos processes batch video generation
// Stops on error to support resume from last successful scene
func (h *VideoHandler) processBatchVideos(storyID uuid.UUID, scenes []model.Scene, req VideoGenerateRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	for i, scene := range scenes {
		fmt.Printf("[Video Batch] Generating video for scene %d (sequence %d)...\n", i+1, scene.Sequence)

		result, err := h.videoGenerator.GenerateFromImage(ctx, &ai.VideoRequest{
			ImageURL:    scene.ImageURL,
			Prompt:      req.Prompt,
			Duration:    req.Duration,
			MotionLevel: req.MotionLevel,
		})

		if err != nil {
			// Stop on error - user can retry later
			fmt.Printf("[Video Batch] ERROR at scene %d: %v. Stopping batch.\n", scene.Sequence, err)
			return
		}

		// Wait for completion
		var videoURL string
		for j := 0; j < 60; j++ { // 60 * 5s = 5 min max per video
			time.Sleep(5 * time.Second)
			status, err := h.videoGenerator.GetTaskStatus(ctx, result.TaskID)
			if err != nil {
				fmt.Printf("[Video Batch] Error checking status: %v\n", err)
				continue
			}

			if status.Status == "completed" {
				videoURL = status.VideoURL
				break
			}

			if status.Status == "failed" {
				fmt.Printf("[Video Batch] Video generation failed for scene %d: %s. Stopping batch.\n", scene.Sequence, status.Error)
				return
			}
		}

		if videoURL == "" {
			fmt.Printf("[Video Batch] Timeout waiting for video for scene %d. Stopping batch.\n", scene.Sequence)
			return
		}

		// Update scene with video URL
		scene.VideoURL = videoURL
		if err := h.repo.UpdateScene(ctx, &scene); err != nil {
			fmt.Printf("[Video Batch] Error updating scene %d: %v. Stopping batch.\n", scene.Sequence, err)
			return
		}

		fmt.Printf("[Video Batch] Scene %d video completed: %s\n", scene.Sequence, videoURL[:min(60, len(videoURL))])
	}

	fmt.Printf("[Video Batch] All %d videos completed for story %s\n", len(scenes), storyID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MergeVideosRequest represents a video merge request
type MergeVideosRequest struct {
	StoryID           string  `json:"story_id" binding:"required"`
	Transition        string  `json:"transition"`         // none, fade, crossfade
	TransitionDuration float64 `json:"transition_duration"` // seconds
	AddAudio          bool    `json:"add_audio"`
}

// MergeVideosResponse represents a video merge response
type MergeVideosResponse struct {
	OutputURL     string  `json:"output_url"`
	TotalScenes   int     `json:"total_scenes"`
	TotalDuration float64 `json:"total_duration"`
	Message       string  `json:"message"`
}

// MergeVideos handles POST /api/v1/videos/merge
// Merges all scene videos into one final video
func (h *VideoHandler) MergeVideos(c *gin.Context) {
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

	// Merge videos
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

	c.JSON(http.StatusOK, MergeVideosResponse{
		OutputURL:     result.OutputURL,
		TotalScenes:   result.TotalScenes,
		TotalDuration: result.TotalDuration,
		Message:       fmt.Sprintf("成功合并 %d 个场景视频，总时长 %.1f 秒", result.TotalScenes, result.TotalDuration),
	})
}

// ViewVideo handles GET /api/v1/videos/view
// Serves merged video files
func (h *VideoHandler) ViewVideo(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file parameter required"})
		return
	}

	// Security check: only allow files from uploads/videos directory
	filePath := filepath.Join("./uploads/videos", filename)
	if !filepath.IsLocal(filePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.FileAttachment(filePath, filename)
}