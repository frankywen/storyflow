package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"sort"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"storyflow/internal/model"
	"storyflow/internal/repository"
)

// ExportService handles image export and composition
type ExportService struct {
	repo       *repository.StoryRepository
	httpClient *http.Client
}

// NewExportService creates a new export service
func NewExportService(repo *repository.StoryRepository) *ExportService {
	return &ExportService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 0, // No timeout for image downloads
		},
	}
}

// ExportFormat represents export format
type ExportFormat string

const (
	ExportFormatPNG ExportFormat = "png"
	ExportFormatPDF ExportFormat = "pdf"
)

// ExportOptions represents export options
type ExportOptions struct {
	Format       ExportFormat `json:"format"`
	ImageWidth   int          `json:"image_width"`   // Width of each image
	Gap          int          `json:"gap"`           // Gap between images
	IncludeTitle bool         `json:"include_title"` // Include story title
	Columns      int          `json:"columns"`       // Number of columns (0 = vertical strip)
}

// ExportResult represents export result
type ExportResult struct {
	ID       string `json:"id"`
	Format   string `json:"format"`
	Data     []byte `json:"-"` // Binary data
	FileSize int    `json:"file_size"`
}

// ExportStory exports a story as combined image
func (s *ExportService) ExportStory(ctx context.Context, storyID uuid.UUID, opts ExportOptions) (*ExportResult, error) {
	// Get story with scenes
	story, err := s.repo.GetWithRelations(ctx, storyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get story: %w", err)
	}

	if len(story.Scenes) == 0 {
		return nil, fmt.Errorf("no scenes to export")
	}

	// Sort scenes by sequence
	sort.Slice(story.Scenes, func(i, j int) bool {
		return story.Scenes[i].Sequence < story.Scenes[j].Sequence
	})

	// Filter scenes with images
	var scenesWithImages []model.Scene
	for _, scene := range story.Scenes {
		if scene.ImageURL != "" {
			scenesWithImages = append(scenesWithImages, scene)
		}
	}

	if len(scenesWithImages) == 0 {
		return nil, fmt.Errorf("no images to export")
	}

	// Download all images
	images, err := s.downloadImages(ctx, scenesWithImages)
	if err != nil {
		return nil, fmt.Errorf("failed to download images: %w", err)
	}

	// Set defaults
	if opts.ImageWidth == 0 {
		opts.ImageWidth = 1024
	}
	if opts.Gap == 0 {
		opts.Gap = 20
	}

	var result *ExportResult

	switch opts.Format {
	case ExportFormatPDF:
		result, err = s.exportPDF(story, images, opts)
	default:
		result, err = s.exportPNG(story, images, opts)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// downloadImages downloads images from URLs in parallel
func (s *ExportService) downloadImages(ctx context.Context, scenes []model.Scene) ([]image.Image, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	images := make([]image.Image, len(scenes))
	errs := make([]error, len(scenes))

	for i, scene := range scenes {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				errs[idx] = err
				return
			}

			resp, err := s.httpClient.Do(req)
			if err != nil {
				errs[idx] = err
				return
			}
			defer resp.Body.Close()

			img, _, err := image.Decode(resp.Body)
			if err != nil {
				errs[idx] = err
				return
			}

			mu.Lock()
			images[idx] = img
			mu.Unlock()
		}(i, scene.ImageURL)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("failed to download image %d: %w", i, err)
		}
	}

	return images, nil
}

// exportPNG exports as vertical PNG strip
func (s *ExportService) exportPNG(story *model.Story, images []image.Image, opts ExportOptions) (*ExportResult, error) {
	// Resize images to target width
	resizedImages := make([]image.Image, len(images))
	for i, img := range images {
		resizedImages[i] = imaging.Resize(img, opts.ImageWidth, 0, imaging.Lanczos)
	}

	// Calculate total height
	totalHeight := 0
	for _, img := range resizedImages {
		totalHeight += img.Bounds().Dy()
	}
	totalHeight += opts.Gap * (len(resizedImages) - 1)

	// Create combined image
	combined := imaging.New(opts.ImageWidth, totalHeight, image.White)
	yOffset := 0

	for _, img := range resizedImages {
		combined = imaging.Paste(combined, img, image.Pt(0, yOffset))
		yOffset += img.Bounds().Dy() + opts.Gap
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, combined); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return &ExportResult{
		ID:       uuid.New().String(),
		Format:   "png",
		Data:     buf.Bytes(),
		FileSize: buf.Len(),
	}, nil
}

// exportPDF exports as PDF document
func (s *ExportService) exportPDF(story *model.Story, images []image.Image, opts ExportOptions) (*ExportResult, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add title page if requested
	if opts.IncludeTitle {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 24)
		pdf.Cell(0, 100, story.Title)
		pdf.Ln(20)
		pdf.SetFont("Arial", "", 14)
		if story.Summary != "" {
			pdf.MultiCell(0, 10, story.Summary, "", "", false)
		}
	}

	// Add images as pages
	for i, img := range images {
		pdf.AddPage()

		// Save image to temp file for PDF
		imgPath := fmt.Sprintf("/tmp/storyflow_%d.png", i)
		if err := imaging.Save(img, imgPath); err != nil {
			continue
		}

		// Get page dimensions
		pageWidth, pageHeight := pdf.GetPageSize()

		// Add image to fit page
		pdf.Image(imgPath, 10, 10, pageWidth-20, pageHeight-20, false, "", 0, "")
	}

	// Output to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return &ExportResult{
		ID:       uuid.New().String(),
		Format:   "pdf",
		Data:     buf.Bytes(),
		FileSize: buf.Len(),
	}, nil
}