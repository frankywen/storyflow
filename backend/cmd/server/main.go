package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"storyflow/internal/api/router"
	"storyflow/internal/auth"
	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/internal/service"
	"storyflow/pkg/crypto"
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

	// Auto migrate
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserConfig{},
		&model.RefreshToken{},
		&model.PasswordResetToken{},
		&model.EmailVerificationCode{},
		&model.TokenBlacklist{},
		&model.Story{},
		&model.Character{},
		&model.Scene{},
		&model.Image{},
		&model.GenerationJob{},
	); err != nil {
		log.Fatal("Failed to auto migrate:", err)
	}

	// Initialize repositories
	storyRepo := repository.NewStoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	userConfigRepo := repository.NewUserConfigRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)

	// Seed initial admin user if no admin exists
	seedAdminUser(userRepo)

	// Initialize encryption service
	encryptionSecret := getEnv("ENCRYPTION_SECRET", "storyflow-default-encryption-key-32b")
	encryptor := crypto.NewEncryptor(encryptionSecret)

	// Initialize JWT service
	jwtConfig := auth.DefaultJWTConfig()
	jwtConfig.AccessTokenSecret = getEnv("JWT_ACCESS_SECRET", "storyflow-access-secret-key")
	jwtConfig.RefreshTokenSecret = getEnv("JWT_REFRESH_SECRET", "storyflow-refresh-secret-key")
	jwtService := auth.NewJWTService(jwtConfig)

	// Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, resetTokenRepo, jwtService, encryptor)
	aiFactory := service.NewAIServiceFactory(userConfigRepo, encryptor)
	emailService := service.NewEmailService(userRepo)
	rateLimitService := service.NewRateLimitService(60, time.Minute)

	// Setup router
	r := router.SetupRouter(
		jwtService,
		authService,
		emailService,
		aiFactory,
		encryptor,
		storyRepo,
		userRepo,
		userConfigRepo,
		rateLimitService,
	)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("StoryFlow server starting on port %s (multi-tenant mode)", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// seedAdminUser creates an initial admin user if no admin exists
func seedAdminUser(userRepo *repository.UserRepository) {
	ctx := context.Background()

	// Check if any admin exists
	adminCount, err := userRepo.Count(ctx, "", "admin")
	if err != nil {
		log.Printf("Failed to check admin count: %v", err)
		return
	}

	if adminCount > 0 {
		log.Printf("Admin user already exists (count: %d)", adminCount)
		return
	}

	// Get admin credentials from environment
	adminEmail := getEnv("ADMIN_EMAIL", "admin@storyflow.local")
	adminPassword := getEnv("ADMIN_PASSWORD", "admin123456")
	adminName := getEnv("ADMIN_NAME", "Admin")

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return
	}

	// Create admin user
	admin := &model.User{
		Email:        adminEmail,
		Name:         adminName,
		PasswordHash: string(hashedPassword),
		Role:         model.RoleAdmin,
		Status:       "active",
	}

	if err := userRepo.Create(ctx, admin); err != nil {
		log.Printf("Failed to create admin user: %v", err)
		return
	}

	log.Printf("Created initial admin user: %s (password: %s)", adminEmail, adminPassword)
	log.Println("Please change the admin password immediately after first login!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}