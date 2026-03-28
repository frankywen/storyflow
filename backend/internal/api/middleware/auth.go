package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storyflow/internal/auth"
	"storyflow/internal/model"
	"storyflow/internal/repository"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtService *auth.JWTService
	userRepo   *repository.UserRepository
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *auth.JWTService, userRepo *repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Check token blacklist
		if m.userRepo.IsTokenBlacklisted(c.Request.Context(), claims.Jti) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token已失效"})
			c.Abort()
			return
		}

		// Load user to get role
		user, err := m.userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		// Check if user is active
		if user.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is suspended or deleted"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", user.Role)
		c.Set("token_jti", claims.Jti)
		c.Set("token_exp", claims.ExpiresAt.Time)
		c.Next()
	}
}

// OptionalAuth is a middleware that optionally extracts user info if present
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

// MustGetUserID extracts user ID from context, panics if not found
func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, exists := GetUserID(c)
	if !exists {
		panic("user_id not found in context")
	}
	return userID
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) (model.UserRole, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	return role.(model.UserRole), true
}

// GetTokenJti extracts token JTI from context
func GetTokenJti(c *gin.Context) (string, bool) {
	jti, exists := c.Get("token_jti")
	if !exists {
		return "", false
	}
	return jti.(string), true
}

// GetTokenExp extracts token expiration from context
func GetTokenExp(c *gin.Context) (time.Time, bool) {
	exp, exists := c.Get("token_exp")
	if !exists {
		return time.Time{}, false
	}
	return exp.(time.Time), true
}

// RequireAdmin is a middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if role != model.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}