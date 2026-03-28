package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSONB is a helper type for JSONB fields in PostgreSQL
type JSONB map[string]interface{}

// Value implements driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// GormDataType implements the GormDataTypeInterface
func (JSONB) GormDataType() string {
	return "jsonb"
}

// CharacterVoice 角色音色配置
type CharacterVoice struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CharacterID   uuid.UUID `json:"character_id" gorm:"type:uuid;uniqueIndex;not null"`
	VoiceID       string    `json:"voice_id" gorm:"not null"`
	VoiceName     string    `json:"voice_name"`
	VoiceProvider string    `json:"voice_provider"`
	VoiceParams   JSONB     `json:"voice_params" gorm:"type:jsonb"`
	IsCustom      bool      `json:"is_custom" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AudioFile 音频文件
type AudioFile struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID     uuid.UUID `json:"story_id" gorm:"type:uuid;not null;index"`
	SceneID     uuid.UUID `json:"scene_id" gorm:"type:uuid;index"`
	CharacterID uuid.UUID `json:"character_id" gorm:"type:uuid"`
	AudioType   string    `json:"audio_type" gorm:"not null"` // dialogue, narration
	TextContent string    `json:"text_content" gorm:"not null"`
	AudioURL    string    `json:"audio_url" gorm:"not null"`
	Duration    float64   `json:"duration" gorm:"not null"`
	VoiceID     string    `json:"voice_id"`
	Status      string    `json:"status" gorm:"default:'completed'"`
	CreatedAt   time.Time `json:"created_at"`
}

// Subtitle 字幕
type Subtitle struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID      uuid.UUID `json:"story_id" gorm:"type:uuid;not null;index"`
	SceneID      uuid.UUID `json:"scene_id" gorm:"type:uuid;index"`
	SubtitleType string    `json:"subtitle_type" gorm:"not null"` // dialogue, narration
	CharacterID  uuid.UUID `json:"character_id" gorm:"type:uuid"`
	Text         string    `json:"text" gorm:"not null"`
	StartTime    float64   `json:"start_time" gorm:"not null"`
	EndTime      float64   `json:"end_time" gorm:"not null"`
	StyleConfig  JSONB     `json:"style_config" gorm:"type:jsonb"`
	CreatedAt    time.Time `json:"created_at"`
}

// AudioGenerationTask 配音生成任务
type AudioGenerationTask struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID         uuid.UUID  `json:"story_id" gorm:"type:uuid;not null;index"`
	Status          string     `json:"status" gorm:"default:'pending'"`
	TotalScenes     int        `json:"total_scenes"`
	CompletedScenes int        `json:"completed_scenes"`
	Progress        int        `json:"progress"`
	FailedScenes    JSONB      `json:"failed_scenes" gorm:"type:jsonb"`
	ErrorMessage    string     `json:"error_message" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

// VideoSynthesisTask 视频合成任务
type VideoSynthesisTask struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID      uuid.UUID  `json:"story_id" gorm:"type:uuid;not null;index"`
	Status       string     `json:"status" gorm:"default:'pending'"`
	Progress     int        `json:"progress"`
	OutputURL    string     `json:"output_url"`
	ErrorMessage string     `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

// VoiceMapping 音色映射
type VoiceMapping struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StandardVoiceID  string    `json:"standard_voice_id" gorm:"not null"`
	Provider         string    `json:"provider" gorm:"not null"`
	ProviderVoiceID  string    `json:"provider_voice_id" gorm:"not null"`
	VoiceName        string    `json:"voice_name"`
}

// TableName methods
func (CharacterVoice) TableName() string     { return "character_voices" }
func (AudioFile) TableName() string          { return "audio_files" }
func (Subtitle) TableName() string            { return "subtitles" }
func (AudioGenerationTask) TableName() string { return "audio_generation_tasks" }
func (VideoSynthesisTask) TableName() string  { return "video_synthesis_tasks" }
func (VoiceMapping) TableName() string        { return "voice_mappings" }