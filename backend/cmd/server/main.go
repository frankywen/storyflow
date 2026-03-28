package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"storyflow/internal/api/router"
	"storyflow/internal/repository"
	"storyflow/pkg/ai"
	"storyflow/pkg/database"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Database configuration
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	dbCfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "storyflow"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect database
	db, err := database.NewPostgres(dbCfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize repositories
	storyRepo := repository.NewStoryRepository(db)

	// Initialize LLM provider
	llmProvider := getEnv("LLM_PROVIDER", "claude")
	llmAPIKey := getEnv("LLM_API_KEY", "")
	if llmAPIKey == "" {
		llmAPIKey = getEnv("CLAUDE_API_KEY", "") // Fallback to CLAUDE_API_KEY
	}
	llmModel := getEnv("LLM_MODEL", "")
	llmBaseURL := getEnv("LLM_BASE_URL", "")

	if llmAPIKey == "" {
		log.Printf("Warning: No LLM API key configured. Set LLM_API_KEY or %s_API_KEY", llmProvider)
	}

	llmCfg := ai.LLMConfig{
		Provider: llmProvider,
		APIKey:   llmAPIKey,
		Model:    llmModel,
		BaseURL:  llmBaseURL,
	}
	llmClient := ai.NewLLMProvider(llmCfg)
	log.Printf("Using LLM provider: %s", llmClient.GetName())

	// Initialize image generator
	imageProvider := getEnv("IMAGE_PROVIDER", "comfyui")
	imageAPIKey := getEnv("IMAGE_API_KEY", "")
	imageBaseURL := getEnv("IMAGE_BASE_URL", "")
	imageModel := getEnv("IMAGE_MODEL", "")

	// Provider-specific defaults
	if imageBaseURL == "" {
		switch imageProvider {
		case "comfyui":
			imageBaseURL = getEnv("COMFYUI_URL", "http://localhost:8188")
		case "volcengine", "doubao":
			imageBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
		case "alibaba", "wanxiang":
			imageBaseURL = "https://dashscope.aliyuncs.com/api/v1"
		}
	}

	imageCfg := ai.ImageGeneratorConfig{
		Provider: imageProvider,
		APIKey:   imageAPIKey,
		BaseURL:  imageBaseURL,
		Model:    imageModel,
	}
	imageGenerator := ai.NewImageGenerator(imageCfg)
	log.Printf("Using image generator: %s", imageGenerator.GetName())

	// Initialize video generator
	videoProvider := getEnv("VIDEO_PROVIDER", "")
	videoAPIKey := getEnv("VIDEO_API_KEY", "")
	videoBaseURL := getEnv("VIDEO_BASE_URL", "")
	videoModel := getEnv("VIDEO_MODEL", "")

	videoCfg := ai.VideoGeneratorConfig{
		Provider: videoProvider,
		APIKey:   videoAPIKey,
		BaseURL:  videoBaseURL,
		Model:    videoModel,
	}
	videoGenerator := ai.NewVideoGenerator(videoCfg)
	log.Printf("Using video generator: %s", videoGenerator.GetName())

	// Setup router
	r := router.SetupRouter(llmClient, imageGenerator, videoGenerator, storyRepo)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("StoryFlow server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}