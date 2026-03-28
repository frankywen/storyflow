package router

import (
	"github.com/gin-gonic/gin"
	"storyflow/internal/api/handler"
	"storyflow/internal/api/middleware"
	"storyflow/internal/repository"
	"storyflow/internal/service"
	"storyflow/pkg/ai"
)

// SetupRouter sets up all routes
func SetupRouter(
	llmProvider ai.LLMProvider,
	imageGenerator ai.ImageGenerator,
	videoGenerator ai.VideoGenerator,
	repo *repository.StoryRepository,
) *gin.Engine {
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":     "ok",
			"service":    "storyflow",
			"llm":        llmProvider.GetName(),
			"image_gen":  imageGenerator.GetName(),
			"video_gen":  videoGenerator.GetName(),
		})
	})

	// Services
	storyService := service.NewStoryService(repo, llmProvider, imageGenerator)
	exportService := service.NewExportService(repo)
	consistencyService := service.NewCharacterConsistencyService(imageGenerator, repo)

	// Handlers
	storyHandler := handler.NewStoryHandler(storyService)
	imageHandler := handler.NewImageHandler(imageGenerator, repo)
	exportHandler := handler.NewExportHandler(exportService)
	videoHandler := handler.NewVideoHandler(videoGenerator, repo)
	characterHandler := handler.NewCharacterHandler(consistencyService, repo)

	// API routes
	api := r.Group("/api/v1")
	{
		// Story routes
		stories := api.Group("/stories")
		{
			stories.POST("", storyHandler.CreateStory)
			stories.GET("", storyHandler.ListStories)
			stories.GET("/:id", storyHandler.GetStory)
			stories.DELETE("/:id", storyHandler.DeleteStory)
			stories.POST("/:id/parse", storyHandler.ParseStory)
			stories.GET("/:id/characters", storyHandler.GetCharacters)
			stories.GET("/:id/scenes", storyHandler.GetScenes)
			stories.POST("/:id/generate-references", characterHandler.GenerateAllReferenceImages)
		}

		// Character routes
		characters := api.Group("/characters")
		{
			characters.POST("/:id/reference", characterHandler.UploadReferenceImage)
			characters.DELETE("/:id/reference", characterHandler.DeleteReferenceImage)
			characters.POST("/:id/regenerate", characterHandler.RegenerateReferenceImage)
		}

		// Image routes
		images := api.Group("/images")
		{
			images.POST("/generate", imageHandler.GenerateImage)
			images.POST("/batch", imageHandler.BatchGenerate)
			images.GET("/jobs/:id", imageHandler.GetJobStatus)
			images.GET("/view", imageHandler.ViewImage)
		}

		// Video routes
		videos := api.Group("/videos")
		{
			videos.POST("/generate", videoHandler.GenerateVideo)
			videos.POST("/batch", videoHandler.GenerateAllVideos)
			videos.GET("/status/:task_id", videoHandler.GetVideoStatus)
			videos.POST("/merge", videoHandler.MergeVideos)
			videos.GET("/view", videoHandler.ViewVideo)
		}

		// Export routes
		api.POST("/export", exportHandler.ExportStory)
	}

	return r
}