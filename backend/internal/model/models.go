package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Story represents a story/novel input
type Story struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string    `json:"title" gorm:"not null"`
	Content     string    `json:"content" gorm:"type:text"`
	Summary     string    `json:"summary" gorm:"type:text"`
	Genre       string    `json:"genre"` // romance, suspense, fantasy, etc.
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, parsed, generating, completed
	MergedVideoURL string `json:"merged_video_url" gorm:"type:text"` // URL for merged video
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Characters []Character `json:"characters,omitempty"`
	Scenes     []Scene     `json:"scenes,omitempty"`
	Images     []Image     `json:"images,omitempty"`
}

// Character represents a character in the story
type Character struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID     uuid.UUID `json:"story_id" gorm:"type:uuid;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	// Visual attributes for consistent generation
	Gender      string `json:"gender"`
	Age         string `json:"age"`
	HairColor   string `json:"hair_color"`
	EyeColor    string `json:"eye_color"`
	BodyType    string `json:"body_type"`
	Clothing    string `json:"clothing"`
	// For image generation consistency
	ReferenceImageURL  string `json:"reference_image_url"`
	ReferenceImageID   string `json:"reference_image_id"` // ID in generation service
	Seed               int64  `json:"seed"`               // character-specific seed for consistency
	VisualPrompt       string `json:"visual_prompt" gorm:"type:text"` // detailed visual description
	PromptTemplate     string `json:"prompt_template" gorm:"type:text"`
	CreatedAt          time.Time `json:"created_at"`
}

// Scene represents a scene/panel in the story
type Scene struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID     uuid.UUID `json:"story_id" gorm:"type:uuid;not null"`
	Sequence    int       `json:"sequence" gorm:"not null"` // order in story
	Title       string    `json:"title"`
	Description string    `json:"description" gorm:"type:text"`
	// Scene details
	Location  string `json:"location"`
	TimeOfDay string `json:"time_of_day"`
	Mood      string `json:"mood"`
	// Characters in this scene (stored as UUID strings)
	CharacterIDs pq.StringArray `json:"character_ids" gorm:"type:text[]"`
	// Dialogue and narration
	Dialogue  string `json:"dialogue" gorm:"type:text"`
	Narration string `json:"narration" gorm:"type:text"`
	// Image generation
	ImagePrompt string `json:"image_prompt" gorm:"type:text"`
	ImageURL    string `json:"image_url"`
	// Video generation
	VideoURL  string    `json:"video_url"`
	Status    string    `json:"status" gorm:"default:'pending'"` // pending, generating, completed
	CreatedAt time.Time `json:"created_at"`
}

// Image represents a generated image
type Image struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID      uuid.UUID `json:"story_id" gorm:"type:uuid;not null"`
	SceneID      uuid.UUID `json:"scene_id" gorm:"type:uuid"`
	Prompt       string    `json:"prompt" gorm:"type:text"`
	ImageURL     string    `json:"image_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Seed         int64     `json:"seed"` // for reproducibility
	Model        string    `json:"model"` // sd model used
	Status       string    `json:"status" gorm:"default:'pending'"`
	CreatedAt    time.Time `json:"created_at"`
}

// GenerationJob tracks image/video generation jobs
type GenerationJob struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StoryID     uuid.UUID `json:"story_id" gorm:"type:uuid;not null"`
	Type        string    `json:"type"` // image, video, batch
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, running, completed, failed
	Progress    int       `json:"progress"` // 0-100
	TotalItems  int       `json:"total_items"`
	DoneItems   int       `json:"done_items"`
	Error       string    `json:"error"`
	ResultURLs  []string  `json:"result_urls" gorm:"type:text[]"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}