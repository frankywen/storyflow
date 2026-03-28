package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"storyflow/internal/model"
)

// StoryRepository handles database operations for stories
type StoryRepository struct {
	db *gorm.DB
}

// NewStoryRepository creates a new story repository
func NewStoryRepository(db *gorm.DB) *StoryRepository {
	return &StoryRepository{db: db}
}

// Create creates a new story
func (r *StoryRepository) Create(ctx context.Context, story *model.Story) error {
	return r.db.WithContext(ctx).Create(story).Error
}

// GetByID retrieves a story by ID (without user check - internal use only)
func (r *StoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Story, error) {
	var story model.Story
	err := r.db.WithContext(ctx).First(&story, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// GetByUserAndID retrieves a story by ID for a specific user
func (r *StoryRepository) GetByUserAndID(ctx context.Context, userID, id uuid.UUID) (*model.Story, error) {
	var story model.Story
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&story).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// GetWithRelations retrieves a story with all related data (without user check)
func (r *StoryRepository) GetWithRelations(ctx context.Context, id uuid.UUID) (*model.Story, error) {
	var story model.Story
	err := r.db.WithContext(ctx).
		Preload("Characters").
		Preload("Scenes", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Preload("Images").
		First(&story, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// GetWithRelationsForUser retrieves a story with all related data for a specific user
func (r *StoryRepository) GetWithRelationsForUser(ctx context.Context, userID, id uuid.UUID) (*model.Story, error) {
	var story model.Story
	err := r.db.WithContext(ctx).
		Preload("Characters").
		Preload("Scenes", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Preload("Images").
		Where("id = ? AND user_id = ?", id, userID).
		First(&story).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// List retrieves stories with pagination (all stories - admin use)
func (r *StoryRepository) List(ctx context.Context, offset, limit int) ([]model.Story, int64, error) {
	var stories []model.Story
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.Story{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&stories).Error
	if err != nil {
		return nil, 0, err
	}

	return stories, total, nil
}

// ListByUser retrieves stories for a specific user with pagination
func (r *StoryRepository) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Story, int64, error) {
	var stories []model.Story
	var total int64

	baseQuery := r.db.WithContext(ctx).Model(&model.Story{}).Where("user_id = ?", userID)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&stories).Error
	if err != nil {
		return nil, 0, err
	}

	return stories, total, nil
}

// Update updates a story
func (r *StoryRepository) Update(ctx context.Context, story *model.Story) error {
	return r.db.WithContext(ctx).Save(story).Error
}

// Delete deletes a story
func (r *StoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Story{}, "id = ?", id).Error
}

// DeleteForUser deletes a story for a specific user
func (r *StoryRepository) DeleteForUser(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Story{}).Error
}

// CreateCharacter creates a new character
func (r *StoryRepository) CreateCharacter(ctx context.Context, character *model.Character) error {
	return r.db.WithContext(ctx).Create(character).Error
}

// GetCharactersByStoryID retrieves characters for a story
func (r *StoryRepository) GetCharactersByStoryID(ctx context.Context, storyID uuid.UUID) ([]model.Character, error) {
	var characters []model.Character
	err := r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Find(&characters).Error
	return characters, err
}

// GetCharacterByID retrieves a character by ID
func (r *StoryRepository) GetCharacterByID(ctx context.Context, id uuid.UUID) (*model.Character, error) {
	var character model.Character
	err := r.db.WithContext(ctx).First(&character, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &character, nil
}

// UpdateCharacter updates a character
func (r *StoryRepository) UpdateCharacter(ctx context.Context, character *model.Character) error {
	return r.db.WithContext(ctx).Save(character).Error
}

// CreateScene creates a new scene
func (r *StoryRepository) CreateScene(ctx context.Context, scene *model.Scene) error {
	return r.db.WithContext(ctx).Create(scene).Error
}

// GetScenesByStoryID retrieves scenes for a story
func (r *StoryRepository) GetScenesByStoryID(ctx context.Context, storyID uuid.UUID) ([]model.Scene, error) {
	var scenes []model.Scene
	err := r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Order("sequence ASC").
		Find(&scenes).Error
	return scenes, err
}

// GetScene retrieves a single scene by ID
func (r *StoryRepository) GetScene(ctx context.Context, sceneID uuid.UUID) (*model.Scene, error) {
	var scene model.Scene
	err := r.db.WithContext(ctx).First(&scene, "id = ?", sceneID).Error
	return &scene, err
}

// UpdateScene updates a scene
func (r *StoryRepository) UpdateScene(ctx context.Context, scene *model.Scene) error {
	return r.db.WithContext(ctx).Save(scene).Error
}

// CreateImage creates a new image
func (r *StoryRepository) CreateImage(ctx context.Context, image *model.Image) error {
	return r.db.WithContext(ctx).Create(image).Error
}

// GetImagesByStoryID retrieves images for a story
func (r *StoryRepository) GetImagesByStoryID(ctx context.Context, storyID uuid.UUID) ([]model.Image, error) {
	var images []model.Image
	err := r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Find(&images).Error
	return images, err
}

// CreateGenerationJob creates a new generation job
func (r *StoryRepository) CreateGenerationJob(ctx context.Context, job *model.GenerationJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

// GetGenerationJob retrieves a generation job by ID
func (r *StoryRepository) GetGenerationJob(ctx context.Context, id uuid.UUID) (*model.GenerationJob, error) {
	var job model.GenerationJob
	err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateGenerationJob updates a generation job
func (r *StoryRepository) UpdateGenerationJob(ctx context.Context, job *model.GenerationJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

// CountByUserID counts stories for a specific user
func (r *StoryRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Story{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountAll counts all stories
func (r *StoryRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Story{}).Count(&count).Error
	return count, err
}

// ListByUserWithFilters lists stories with filtering support
func (r *StoryRepository) ListByUserWithFilters(ctx context.Context, userID uuid.UUID, offset, limit int, status, search string) ([]model.Story, int64, error) {
	var stories []model.Story
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Story{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Query list
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&stories).Error

	return stories, total, err
}

// SoftDeleteStory soft deletes a story
func (r *StoryRepository) SoftDeleteStory(ctx context.Context, userID, storyID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Story{}).
		Where("id = ? AND user_id = ?", storyID, userID).
		Update("deleted_at", time.Now()).Error
}

// RestoreStory restores a soft-deleted story
func (r *StoryRepository) RestoreStory(ctx context.Context, userID, storyID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Story{}).
		Where("id = ? AND user_id = ?", storyID, userID).
		Update("deleted_at", nil).Error
}

// UpdateCoverURL updates the cover URL of a story
func (r *StoryRepository) UpdateCoverURL(ctx context.Context, storyID uuid.UUID, coverURL string) error {
	return r.db.WithContext(ctx).
		Model(&model.Story{}).
		Where("id = ?", storyID).
		Update("cover_url", coverURL).Error
}