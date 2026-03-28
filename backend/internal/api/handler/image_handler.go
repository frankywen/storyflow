package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/internal/service"
	"storyflow/pkg/ai"
)

// ImageHandler handles image generation requests
type ImageHandler struct {
	imageGenerator      ai.ImageGenerator
	repo                *repository.StoryRepository
	consistencyService  *service.CharacterConsistencyService
}

// NewImageHandler creates a new image handler
func NewImageHandler(imageGenerator ai.ImageGenerator, repo *repository.StoryRepository) *ImageHandler {
	return &ImageHandler{
		imageGenerator:     imageGenerator,
		repo:               repo,
		consistencyService: service.NewCharacterConsistencyService(imageGenerator, repo),
	}
}

// GenerateImageRequest represents a single image generation request
type GenerateImageRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Negative string `json:"negative_prompt"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Seed     int64  `json:"seed"`
	Steps    int    `json:"steps"`
	Style    string `json:"style"` // manga, realistic, anime, etc.
}

// GenerateImageResponse represents the response
type GenerateImageResponse struct {
	ImageID  string `json:"image_id"`
	Status   string `json:"status"`
	ImageURL string `json:"image_url,omitempty"`
}

// BatchGenerateRequest represents a batch generation request
type BatchGenerateRequest struct {
	StoryID        uuid.UUID `json:"story_id" binding:"required"`
	Style          string    `json:"style"`
	UseConsistency bool      `json:"use_consistency"` // Enable character consistency
}

// GenerateImage handles POST /api/v1/images/generate
func (h *ImageHandler) GenerateImage(c *gin.Context) {
	var req GenerateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate image
	result, err := h.imageGenerator.Generate(c.Request.Context(), &ai.ImageRequest{
		Prompt:         req.Prompt,
		NegativePrompt: req.Negative,
		Width:          req.Width,
		Height:         req.Height,
		Seed:           req.Seed,
		Steps:          req.Steps,
		Style:          req.Style,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to generate image: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_id":  result.ID,
		"status":    "completed",
		"image_url": result.ImageURL,
	})
}

// BatchGenerate handles POST /api/v1/images/batch
func (h *ImageHandler) BatchGenerate(c *gin.Context) {
	var req BatchGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get story with scenes and characters
	story, err := h.repo.GetWithRelations(c.Request.Context(), req.StoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	if len(story.Scenes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no scenes to generate"})
		return
	}

	// Create generation job
	job := &model.GenerationJob{
		ID:         uuid.New(),
		StoryID:    req.StoryID,
		Type:       "batch_image",
		Status:     "pending",
		TotalItems: len(story.Scenes),
	}
	if err := h.repo.CreateGenerationJob(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	// Start async generation
	go h.processBatchGeneration(job, story, req.Style, req.UseConsistency)

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":     job.ID,
		"status":     job.Status,
		"total":      job.TotalItems,
		"message":    "Batch generation started",
		"status_url": fmt.Sprintf("/api/v1/images/jobs/%s", job.ID),
	})
}

// GetJobStatus handles GET /api/v1/images/jobs/:id
func (h *ImageHandler) GetJobStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	job, err := h.repo.GetGenerationJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          job.ID,
		"status":      job.Status,
		"progress":    job.Progress,
		"total":       job.TotalItems,
		"done":        job.DoneItems,
		"result_urls": job.ResultURLs,
		"error":       job.Error,
	})
}

// ViewImage handles GET /api/v1/images/view
// This is now a placeholder - for cloud providers, images are served from their URLs
func (h *ImageHandler) ViewImage(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter required"})
		return
	}

	// Redirect to the image URL
	c.Redirect(http.StatusFound, imageURL)
}

// processBatchGeneration processes batch image generation asynchronously
func (h *ImageHandler) processBatchGeneration(job *model.GenerationJob, story *model.Story, style string, useConsistency bool) {
	ctx := context.Background()
	jobID := job.ID

	// Update job status
	job.Status = "running"
	if err := h.repo.UpdateGenerationJob(ctx, job); err != nil {
		fmt.Printf("Error updating job status to running: %v\n", err)
	}

	var resultURLs []string

	// Get characters if using consistency
	var characters []model.Character
	if useConsistency {
		var err error
		characters, err = h.repo.GetCharactersByStoryID(ctx, story.ID)
		if err != nil {
			fmt.Printf("Warning: failed to get characters for consistency: %v\n", err)
		}
	}

	for _, scene := range story.Scenes {
		var result *ai.ImageResult
		var err error

		// Use consistency service if enabled and characters exist
		if useConsistency && len(characters) > 0 && len(scene.CharacterIDs) > 0 {
			result, err = h.consistencyService.GenerateSceneWithConsistency(ctx, &scene, characters, style)
		} else {
			// Standard generation
			result, err = h.imageGenerator.Generate(ctx, &ai.ImageRequest{
				Prompt: scene.ImagePrompt,
				Style:   style,
			})
		}

		if err != nil {
			job.Error = fmt.Sprintf("failed to generate scene %d: %v", scene.Sequence, err)
			job.Status = "failed"
			h.repo.UpdateGenerationJob(ctx, job)
			return
		}

		resultURLs = append(resultURLs, result.ImageURL)

		// Update scene with image URL
		scene.ImageURL = result.ImageURL
		scene.Status = "completed"
		h.repo.UpdateScene(ctx, &scene)

		// Save to database
		image := &model.Image{
			ID:       uuid.New(),
			StoryID:  story.ID,
			SceneID:  scene.ID,
			Prompt:   scene.ImagePrompt,
			ImageURL: result.ImageURL,
			Model:    h.imageGenerator.GetName(),
			Status:   "completed",
		}
		h.repo.CreateImage(ctx, image)

		// Update progress - reload job to avoid stale data
		currentJob, err := h.repo.GetGenerationJob(ctx, jobID)
		if err != nil {
			currentJob = job // fallback
		}
		currentJob.DoneItems++
		currentJob.Progress = int(float64(currentJob.DoneItems) / float64(currentJob.TotalItems) * 100)
		h.repo.UpdateGenerationJob(ctx, currentJob)
	}

	// Complete job - reload and update
	finalJob, err := h.repo.GetGenerationJob(ctx, jobID)
	if err != nil {
		finalJob = job
	}
	finalJob.Status = "completed"
	finalJob.ResultURLs = resultURLs
	finalJob.Progress = 100
	if err := h.repo.UpdateGenerationJob(ctx, finalJob); err != nil {
		fmt.Printf("Error updating job to completed: %v\n", err)
	}

	log.Printf("Job %s completed with %d images", jobID, len(resultURLs))
}