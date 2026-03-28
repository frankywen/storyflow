package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"storyflow/internal/service"
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rateLimitService *service.RateLimitService, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if !rateLimitService.Check(c.Request.Context(), key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "请求过于频繁，请稍后再试",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitByIP rate limits by client IP
func RateLimitByIP(rateLimitService *service.RateLimitService) gin.HandlerFunc {
	return RateLimitMiddleware(rateLimitService, func(c *gin.Context) string {
		return "ip:" + c.ClientIP()
	})
}

// RateLimitByUser rate limits by user ID (falls back to IP if not authenticated)
func RateLimitByUser(rateLimitService *service.RateLimitService) gin.HandlerFunc {
	return RateLimitMiddleware(rateLimitService, func(c *gin.Context) string {
		userID, exists := c.Get("user_id")
		if exists {
			if userIDStr, ok := userID.(string); ok {
				return "user:" + userIDStr
			}
			// uuid.UUID type
			if uid, ok := userID.(uuid.UUID); ok {
				return "user:" + uid.String()
			}
		}
		return "ip:" + c.ClientIP()
	})
}