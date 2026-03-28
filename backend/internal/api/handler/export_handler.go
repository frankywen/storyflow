package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"storyflow/internal/service"
)

// ExportHandler handles export requests
type ExportHandler struct {
	exportService *service.ExportService
}

// NewExportHandler creates a new export handler
func NewExportHandler(exportService *service.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

// ExportRequest represents an export request
type ExportRequest struct {
	StoryID     string `json:"story_id" binding:"required"`
	Format      string `json:"format"`       // png, pdf
	ImageWidth  int    `json:"image_width"`  // default 1024
	Gap         int    `json:"gap"`          // default 20
	IncludeTitle bool   `json:"include_title"` // default true
}

// ExportStory handles POST /api/v1/export
func (h *ExportHandler) ExportStory(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storyID, err := uuid.Parse(req.StoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid story ID"})
		return
	}

	// Set format
	format := service.ExportFormatPNG
	if req.Format == "pdf" {
		format = service.ExportFormatPDF
	}

	// Export
	result, err := h.exportService.ExportStory(c.Request.Context(), storyID, service.ExportOptions{
		Format:       format,
		ImageWidth:   req.ImageWidth,
		Gap:          req.Gap,
		IncludeTitle: req.IncludeTitle,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set response headers
	filename := fmt.Sprintf("storyflow_%s.%s", result.ID, result.Format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", result.FileSize))

	if result.Format == "pdf" {
		c.Data(http.StatusOK, "application/pdf", result.Data)
	} else {
		c.Data(http.StatusOK, "image/png", result.Data)
	}
}