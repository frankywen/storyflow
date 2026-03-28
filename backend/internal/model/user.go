package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRole defines user roles
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// User represents a registered user
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"` // Never expose in JSON
	Name         string    `json:"name"`
	AvatarURL    string    `json:"avatar_url"`
	Role         UserRole  `json:"role" gorm:"type:varchar(20);default:'user'"`
	Status       string    `json:"status" gorm:"default:'active'"` // active, suspended, deleted
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  time.Time `json:"last_login_at"`

	// Relations
	Config *UserConfig `json:"config,omitempty"`
}

// UserConfig stores user's API keys and preferences
type UserConfig struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`

	// LLM Configuration
	LLMProvider string `json:"llm_provider"`           // claude, volcengine, alibaba
	LLMAPIKey   string `json:"llm_api_key,omitempty"`  // Encrypted in DB, masked in response
	LLMModel    string `json:"llm_model"`
	LLMBaseURL  string `json:"llm_base_url"`

	// Image Generation Configuration
	ImageProvider string `json:"image_provider"`
	ImageAPIKey   string `json:"image_api_key,omitempty"`
	ImageBaseURL  string `json:"image_base_url"`
	ImageModel    string `json:"image_model"`

	// Video Generation Configuration
	VideoProvider string `json:"video_provider"`
	VideoAPIKey   string `json:"video_api_key,omitempty"`
	VideoBaseURL  string `json:"video_base_url"`
	VideoModel    string `json:"video_model"`

	// Preferences
	DefaultStyle string `json:"default_style"` // manga, manhwa, realistic

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RefreshToken for JWT refresh
type RefreshToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	TokenHash string    `json:"-" gorm:"uniqueIndex;not null"` // Hashed token
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked" gorm:"default:false"`
}

// PasswordResetToken for password reset
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	TokenHash string    `json:"-" gorm:"uniqueIndex;not null"` // Hashed token
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	Used      bool      `json:"used" gorm:"default:false"`
}

// EmailVerificationCode represents an email verification code
type EmailVerificationCode struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string    `json:"email" gorm:"not null;index"`
	Code      string    `json:"-" gorm:"not null"` // 6位验证码，不暴露
	CodeType  string    `json:"code_type" gorm:"type:varchar(20);not null"` // register, reset_password
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsUsed    bool      `json:"is_used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// TokenBlacklist represents a blacklisted JWT token
type TokenBlacklist struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TokenJti  string    `json:"token_jti" gorm:"uniqueIndex;not null"` // JWT ID
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"` // Token过期时间
	CreatedAt time.Time `json:"created_at"`
}

// TableName methods
func (User) TableName() string {
	return "users"
}

func (UserConfig) TableName() string {
	return "user_configs"
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

func (EmailVerificationCode) TableName() string {
	return "email_verification_codes"
}

func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}