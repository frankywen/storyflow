package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"storyflow/internal/model"
)

// UserRepository handles user data access
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateLastLogin updates the last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("last_login_at", time.Now()).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

// List lists users with pagination and filters
func (r *UserRepository) List(ctx context.Context, offset, limit int, status, role string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Count counts users with optional filters
func (r *UserRepository) Count(ctx context.Context, status, role string) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.User{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// RevokeAllTokens revokes all refresh tokens for a user
func (r *UserRepository) RevokeAllTokens(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}

// UserConfigRepository handles user configuration data access
type UserConfigRepository struct {
	db *gorm.DB
}

// NewUserConfigRepository creates a new user config repository
func NewUserConfigRepository(db *gorm.DB) *UserConfigRepository {
	return &UserConfigRepository{db: db}
}

// GetByUserID retrieves user config by user ID
func (r *UserConfigRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.UserConfig, error) {
	var config model.UserConfig
	err := r.db.WithContext(ctx).First(&config, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Create creates a new user config
func (r *UserConfigRepository) Create(ctx context.Context, config *model.UserConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// CreateOrUpdate creates or updates user config
func (r *UserConfigRepository) CreateOrUpdate(ctx context.Context, config *model.UserConfig) error {
	var existing model.UserConfig
	err := r.db.WithContext(ctx).First(&existing, "user_id = ?", config.UserID).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(config).Error
	}
	if err != nil {
		return err
	}
	config.ID = existing.ID
	return r.db.WithContext(ctx).Save(config).Error
}

// Update updates user config
func (r *UserConfigRepository) Update(ctx context.Context, config *model.UserConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// RefreshTokenRepository handles refresh token data access
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByTokenHash retrieves a refresh token by hash
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Revoke revokes a refresh token
func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked", true).Error
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}

// DeleteExpired deletes all expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&model.RefreshToken{}).Error
}

// PasswordResetTokenRepository handles password reset token data access
type PasswordResetTokenRepository struct {
	db *gorm.DB
}

// NewPasswordResetTokenRepository creates a new password reset token repository
func NewPasswordResetTokenRepository(db *gorm.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

// Create creates a new password reset token
func (r *PasswordResetTokenRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByTokenHash retrieves a password reset token by hash
func (r *PasswordResetTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// MarkUsed marks a password reset token as used
func (r *PasswordResetTokenRepository) MarkUsed(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Model(&model.PasswordResetToken{}).
		Where("token_hash = ?", tokenHash).
		Update("used", true).Error
}

// DeleteExpired deletes all expired password reset tokens
func (r *PasswordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&model.PasswordResetToken{}).Error
}

// DeleteAllForUser deletes all password reset tokens for a user
func (r *PasswordResetTokenRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.PasswordResetToken{}).Error
}

// CreateVerificationCode 创建验证码记录
func (r *UserRepository) CreateVerificationCode(ctx context.Context, code *model.EmailVerificationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

// GetValidVerificationCode 获取有效验证码
func (r *UserRepository) GetValidVerificationCode(ctx context.Context, email, code, codeType string) (*model.EmailVerificationCode, error) {
	var record model.EmailVerificationCode
	err := r.db.WithContext(ctx).
		Where("email = ? AND code = ? AND code_type = ?", email, code, codeType).
		Where("expires_at > ?", time.Now()).
		Where("is_used = ?", false).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetLatestVerificationCode 获取最新验证码
func (r *UserRepository) GetLatestVerificationCode(ctx context.Context, email, codeType string) (*model.EmailVerificationCode, error) {
	var record model.EmailVerificationCode
	err := r.db.WithContext(ctx).
		Where("email = ? AND code_type = ?", email, codeType).
		Where("expires_at > ?", time.Now()).
		Where("is_used = ?", false).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// MarkCodeAsUsed 标记验证码已使用
func (r *UserRepository) MarkCodeAsUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.EmailVerificationCode{}).
		Where("id = ?", id).
		Update("is_used", true).Error
}

// CanSendCode checks if a verification code can be sent (60-second interval)
// Returns false if rate limited, true if can send
// Returns true on database error (fail-open for availability)
func (r *UserRepository) CanSendCode(ctx context.Context, email string) bool {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.EmailVerificationCode{}).
		Where("email = ? AND created_at > ?", email, time.Now().Add(-60*time.Second)).
		Count(&count).Error; err != nil {
		// Log error but allow sending (fail-open for availability)
		// In production, consider logging: log.Printf("CanSendCode error: %v", err)
		return true
	}
	return count == 0
}

// CleanupExpiredCodes 清理过期验证码
func (r *UserRepository) CleanupExpiredCodes(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&model.EmailVerificationCode{}).Error
}