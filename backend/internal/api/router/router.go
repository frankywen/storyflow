package router

import (
	"github.com/gin-gonic/gin"
	"storyflow/internal/api/handler"
	"storyflow/internal/api/middleware"
	"storyflow/internal/auth"
	"storyflow/internal/repository"
	"storyflow/internal/service"
	"storyflow/pkg/crypto"
)

// SetupRouter sets up all routes with multi-tenant support
func SetupRouter(
	jwtService *auth.JWTService,
	authService *service.AuthService,
	aiFactory *service.AIServiceFactory,
	encryptor *crypto.Encryptor,
	repo *repository.StoryRepository,
	userRepo *repository.UserRepository,
	userConfigRepo *repository.UserConfigRepository,
	rateLimitService *service.RateLimitService,
) *gin.Engine {
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userRepo)

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "storyflow",
		})
	})

	// Services
	exportService := service.NewExportService(repo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	configHandler := handler.NewConfigHandler(userConfigRepo, encryptor, aiFactory)
	storyHandler := handler.NewStoryHandler(repo, aiFactory)
	imageHandler := handler.NewImageHandler(repo, aiFactory)
	exportHandler := handler.NewExportHandler(exportService, repo)
	videoHandler := handler.NewVideoHandler(repo, aiFactory)
	characterHandler := handler.NewCharacterHandler(repo, aiFactory)
	adminHandler := handler.NewAdminHandler(userRepo, userConfigRepo, repo)

	// API routes
	api := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitByIP(rateLimitService)) // 10次/分钟，已在 service 初始化时设置
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/forgot-password", authHandler.RequestPasswordReset)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Me routes (protected)
		me := api.Group("/auth")
		me.Use(authMiddleware.RequireAuth())
		{
			me.GET("/me", authHandler.GetMe)
			me.POST("/logout", authHandler.Logout)
		}

		// User routes (protected)
		user := api.Group("/user")
		user.Use(authMiddleware.RequireAuth())
		{
			user.GET("/config", configHandler.GetConfig)
			user.PUT("/config", configHandler.UpdateConfig)
			user.PUT("/config/llm", configHandler.UpdateLLMConfig)
			user.PUT("/config/image", configHandler.UpdateImageConfig)
			user.PUT("/config/video", configHandler.UpdateVideoConfig)
			user.POST("/config/validate", configHandler.ValidateAPIKey)
		}

		// Admin routes (admin only)
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		admin.Use(authMiddleware.RequireAdmin())
		{
			admin.GET("/stats", adminHandler.GetUserStats)
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/:id", adminHandler.GetUser)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.POST("/users/:id/suspend", adminHandler.SuspendUser)
			admin.POST("/users/:id/activate", adminHandler.ActivateUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
		}

		// Protected API routes
		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Story routes
			stories := protected.Group("/stories")
			{
				stories.POST("", storyHandler.CreateStory)
				stories.GET("", storyHandler.ListStories)
				stories.GET("/:id", storyHandler.GetStory)
				stories.PUT("/:id", storyHandler.UpdateStory)
				stories.DELETE("/:id", storyHandler.DeleteStory)
				stories.POST("/:id/restore", storyHandler.RestoreStory)
				stories.POST("/:id/parse", storyHandler.ParseStory)
				stories.GET("/:id/characters", storyHandler.GetCharacters)
				stories.GET("/:id/scenes", storyHandler.GetScenes)
				stories.POST("/:id/generate-references", characterHandler.GenerateAllReferenceImages)
			}

			// Character routes
			characters := protected.Group("/characters")
			{
				characters.POST("/:id/reference", characterHandler.UploadReferenceImage)
				characters.DELETE("/:id/reference", characterHandler.DeleteReferenceImage)
				characters.POST("/:id/regenerate", characterHandler.RegenerateReferenceImage)
			}

			// Image routes
			images := protected.Group("/images")
			{
				images.POST("/generate", imageHandler.GenerateImage)
				images.POST("/batch", imageHandler.BatchGenerate)
				images.GET("/jobs/:id", imageHandler.GetJobStatus)
				images.GET("/view", imageHandler.ViewImage)
			}

			// Video routes
			videos := protected.Group("/videos")
			{
				videos.POST("/generate", videoHandler.GenerateVideo)
				videos.POST("/batch", videoHandler.GenerateAllVideos)
				videos.GET("/status/:task_id", videoHandler.GetVideoStatus)
				videos.POST("/merge", videoHandler.MergeVideos)
				videos.GET("/view", videoHandler.ViewVideo)
			}

			// Export routes
			protected.POST("/export", exportHandler.ExportStory)
		}
	}

	return r
}