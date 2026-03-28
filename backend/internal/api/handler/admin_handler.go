package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storyflow/internal/model"
	"storyflow/internal/repository"
)

// AdminHandler handles admin operations
type AdminHandler struct {
	userRepo       *repository.UserRepository
	userConfigRepo *repository.UserConfigRepository
	storyRepo      *repository.StoryRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userRepo *repository.UserRepository,
	userConfigRepo *repository.UserConfigRepository,
	storyRepo *repository.StoryRepository,
) *AdminHandler {
	return &AdminHandler{
		userRepo:       userRepo,
		userConfigRepo: userConfigRepo,
		storyRepo:      storyRepo,
	}
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	role := c.Query("role")

	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	users, total, err := h.userRepo.List(c.Request.Context(), offset, pageSize, status, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	// Sanitize users (remove password hashes)
	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = sanitizeUserForAdmin(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetUser handles GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get user's config
	config, _ := h.userConfigRepo.GetByUserID(c.Request.Context(), id)

	// Get user's story count
	storyCount, _ := h.storyRepo.CountByUserID(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"user":        sanitizeUserForAdmin(user),
		"config":      config,
		"story_count": storyCount,
	})
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	Name   string         `json:"name"`
	Role   model.UserRole `json:"role"`
	Status string         `json:"status"`
}

// UpdateUser handles PUT /api/v1/admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	adminID := c.MustGet("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Prevent admin from modifying themselves (role/status)
	if id == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify your own admin account"})
		return
	}

	var input UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update fields
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Role != "" {
		if input.Role != model.RoleAdmin && input.Role != model.RoleUser {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		user.Role = input.Role
	}
	if input.Status != "" {
		if input.Status != "active" && input.Status != "suspended" && input.Status != "deleted" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		user.Status = input.Status
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user updated",
		"user":    sanitizeUserForAdmin(user),
	})
}

// SuspendUser handles POST /api/v1/admin/users/:id/suspend
func (h *AdminHandler) SuspendUser(c *gin.Context) {
	adminID := c.MustGet("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if id == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot suspend your own account"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.Status = "suspended"
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to suspend user"})
		return
	}

	// Revoke all refresh tokens
	h.userRepo.RevokeAllTokens(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{"message": "user suspended"})
}

// ActivateUser handles POST /api/v1/admin/users/:id/activate
func (h *AdminHandler) ActivateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.Status = "active"
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user activated"})
}

// DeleteUser handles DELETE /api/v1/admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminID := c.MustGet("user_id").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if id == adminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own account"})
		return
	}

	// Soft delete by setting status to deleted
	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.Status = "deleted"
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	// Revoke all tokens
	h.userRepo.RevokeAllTokens(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// GetUserStats handles GET /api/v1/admin/stats
func (h *AdminHandler) GetUserStats(c *gin.Context) {
	totalUsers, _ := h.userRepo.Count(c.Request.Context(), "", "")
	activeUsers, _ := h.userRepo.Count(c.Request.Context(), "active", "")
	suspendedUsers, _ := h.userRepo.Count(c.Request.Context(), "suspended", "")
	adminCount, _ := h.userRepo.Count(c.Request.Context(), "", "admin")

	totalStories, _ := h.storyRepo.CountAll(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"total_users":      totalUsers,
		"active_users":     activeUsers,
		"suspended_users":  suspendedUsers,
		"admin_count":      adminCount,
		"total_stories":    totalStories,
	})
}

// sanitizeUserForAdmin returns user data safe for admin viewing
func sanitizeUserForAdmin(user *model.User) map[string]interface{} {
	return map[string]interface{}{
		"id":            user.ID,
		"email":         user.Email,
		"name":          user.Name,
		"avatar_url":    user.AvatarURL,
		"role":          user.Role,
		"status":        user.Status,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
		"last_login_at": user.LastLoginAt,
	}
}