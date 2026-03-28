package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"storyflow/internal/model"
)

type AudioRepository struct {
	db *gorm.DB
}

func NewAudioRepository(db *gorm.DB) *AudioRepository {
	return &AudioRepository{db: db}
}

// AudioFile methods

func (r *AudioRepository) CreateAudio(ctx context.Context, audio *model.AudioFile) error {
	return r.db.WithContext(ctx).Create(audio).Error
}

func (r *AudioRepository) GetAudiosByStory(ctx context.Context, storyID uuid.UUID) ([]model.AudioFile, error) {
	var audios []model.AudioFile
	err := r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Order("created_at ASC").
		Find(&audios).Error
	return audios, err
}

func (r *AudioRepository) GetAudiosByScene(ctx context.Context, sceneID uuid.UUID) ([]model.AudioFile, error) {
	var audios []model.AudioFile
	err := r.db.WithContext(ctx).
		Where("scene_id = ?", sceneID).
		Order("created_at ASC").
		Find(&audios).Error
	return audios, err
}

func (r *AudioRepository) DeleteAudiosByStory(ctx context.Context, storyID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Delete(&model.AudioFile{}).Error
}

// AudioGenerationTask methods

func (r *AudioRepository) CreateTask(ctx context.Context, task *model.AudioGenerationTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *AudioRepository) GetTask(ctx context.Context, taskID uuid.UUID) (*model.AudioGenerationTask, error) {
	var task model.AudioGenerationTask
	err := r.db.WithContext(ctx).First(&task, "id = ?", taskID).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AudioRepository) UpdateTask(ctx context.Context, task *model.AudioGenerationTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// CharacterVoice methods

func (r *AudioRepository) GetCharacterVoice(ctx context.Context, characterID uuid.UUID) (*model.CharacterVoice, error) {
	var voice model.CharacterVoice
	err := r.db.WithContext(ctx).First(&voice, "character_id = ?", characterID).Error
	if err != nil {
		return nil, err
	}
	return &voice, nil
}

func (r *AudioRepository) SaveCharacterVoice(ctx context.Context, voice *model.CharacterVoice) error {
	return r.db.WithContext(ctx).Save(voice).Error
}

// Subtitle methods

func (r *AudioRepository) CreateSubtitle(ctx context.Context, subtitle *model.Subtitle) error {
	return r.db.WithContext(ctx).Create(subtitle).Error
}

func (r *AudioRepository) GetSubtitlesByStory(ctx context.Context, storyID uuid.UUID) ([]model.Subtitle, error) {
	var subtitles []model.Subtitle
	err := r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Order("start_time ASC").
		Find(&subtitles).Error
	return subtitles, err
}

func (r *AudioRepository) DeleteSubtitlesByStory(ctx context.Context, storyID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Delete(&model.Subtitle{}).Error
}

// VoiceMapping methods

func (r *AudioRepository) GetVoiceMapping(ctx context.Context, standardVoiceID, provider string) (*model.VoiceMapping, error) {
	var mapping model.VoiceMapping
	err := r.db.WithContext(ctx).
		Where("standard_voice_id = ? AND provider = ?", standardVoiceID, provider).
		First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}