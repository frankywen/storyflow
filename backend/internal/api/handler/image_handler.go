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
	repo      *repository.StoryRepository
	aiFactory *service.AIServiceFactory
}

// NewImageHandler creates a new image handler
func NewImageHandler(repo *repository.StoryRepository, aiFactory *service.AIServiceFactory) *ImageHandler {
	return &ImageHandler{
		repo:      repo,
		aiFactory: aiFactory,
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
	Style    string `json:"style"`
}

// BatchGenerateRequest represents a batch generation request
type BatchGenerateRequest struct {
	StoryID        uuid.UUID `json:"story_id" binding:"required"`
	Style          string    `json:"style"`
	UseConsistency bool      `json:"use_consistency"`
}

// GenerateImage handles POST /api/v1/images/generate
func (h *ImageHandler) GenerateImage(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req GenerateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user's image generator
	imageGenerator, err := h.aiFactory.GetImageGenerator(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image generator not configured"})
		return
	}

	// Generate image
	result, err := imageGenerator.Generate(c.Request.Context(), &ai.ImageRequest{
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
	userID := c.MustGet("user_id").(uuid.UUID)

	var req BatchGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get story with scenes and characters (verify ownership)
	story, err := h.repo.GetWithRelationsForUser(c.Request.Context(), userID, req.StoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
		return
	}

	// Check if image generator is configured
	if !h.aiFactory.HasImageConfig(c.Request.Context(), userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image generator not configured"})
		return
	}

	// Create job
	job := &model.GenerationJob{
		ID:         uuid.New(),
		StoryID:    story.ID,
		Type:       "image_batch",
		Status:     "pending",
		TotalItems: len(story.Scenes),
	}
	if err := h.repo.CreateGenerationJob(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	// Start async processing
	go h.processBatchImages(userID, story, job, req.Style, req.UseConsistency)

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":     job.ID,
		"status":     "pending",
		"total":      len(story.Scenes),
		"message":    "Batch generation started",
	})
}

// processBatchImages processes batch image generation
func (h *ImageHandler) processBatchImages(userID uuid.UUID, story *model.Story, job *model.GenerationJob, style string, useConsistency bool) {
	ctx := context.Background()

	imageGenerator, err := h.aiFactory.GetImageGenerator(ctx, userID)
	if err != nil {
		log.Printf("Failed to get image generator: %v", err)
		return
	}

	consistencyService := service.NewCharacterConsistencyService(imageGenerator, h.repo)

	for _, scene := range story.Scenes {
		if scene.ImageURL != "" {
			continue // Skip already generated
		}

		job.Status = "running"
		h.repo.UpdateGenerationJob(ctx, job)

		var result *ai.ImageResult
		var err error

		if useConsistency && len(scene.CharacterIDs) > 0 {
			// Get characters for this scene
			characters, _ := h.repo.GetCharactersByStoryID(ctx, story.ID)
			result, err = consistencyService.GenerateSceneWithConsistency(ctx, &scene, characters, style)
		} else {
			result, err = imageGenerator.Generate(ctx, &ai.ImageRequest{
				Prompt: scene.ImagePrompt,
				Style:  style,
			})
		}

		if err != nil {
			log.Printf("Failed to generate image for scene %d: %v", scene.Sequence, err)
			continue
		}

		// Update scene
		scene.ImageURL = result.ImageURL
		scene.Status = "completed"
		h.repo.UpdateScene(ctx, &scene)

		job.DoneItems++
		job.Progress = int(float64(job.DoneItems) / float64(job.TotalItems) * 100)
		h.repo.UpdateGenerationJob(ctx, job)
	}

	job.Status = "completed"
	job.Progress = 100
	h.repo.UpdateGenerationJob(ctx, job)
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
		"error":       job.Error,
		"result_urls": job.ResultURLs,
	})
}

// ViewImage handles GET /api/v1/images/view
func (h *ImageHandler) ViewImage(c *gin.Context) {
	// Redirect to the image URL or serve from storage
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url parameter"})
		return
	}
	c.Redirect(http.StatusFound, url)
}