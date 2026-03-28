package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"storyflow/internal/auth"
	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/pkg/crypto"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo         *repository.UserRepository
	tokenRepo        *repository.RefreshTokenRepository
	resetTokenRepo   *repository.PasswordResetTokenRepository
	jwtService       *auth.JWTService
	encryptor        *crypto.Encryptor
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.RefreshTokenRepository,
	resetTokenRepo *repository.PasswordResetTokenRepository,
	jwtService *auth.JWTService,
	encryptor *crypto.Encryptor,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		resetTokenRepo: resetTokenRepo,
		jwtService:     jwtService,
		encryptor:      encryptor,
	}
}

// RegisterInput represents registration input
type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

// LoginInput represents login input
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*model.User, *auth.TokenPair, error) {
	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, nil, errors.New("email already registered")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	// Create user
	user := &model.User{
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		Name:         input.Name,
		Status:       "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	// Store refresh token
	refreshToken := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: s.jwtService.HashToken(tokens.RefreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.tokenRepo.Create(ctx, refreshToken)

	return user, tokens, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*model.User, *auth.TokenPair, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	// Check status
	if user.Status != "active" {
		return nil, nil, errors.New("account is not active")
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	// Store refresh token
	refreshToken := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: s.jwtService.HashToken(tokens.RefreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.tokenRepo.Create(ctx, refreshToken)

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	return user, tokens, nil
}

// RefreshToken refreshes an access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// Validate refresh token
	userID, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if token is revoked
	tokenHash := s.jwtService.HashToken(refreshToken)
	storedToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || storedToken.Revoked {
		return nil, errors.New("refresh token is revoked")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Revoke old token
	s.tokenRepo.Revoke(ctx, tokenHash)

	// Generate new tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	newRefreshToken := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: s.jwtService.HashToken(tokens.RefreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.tokenRepo.Create(ctx, newRefreshToken)

	return tokens, nil
}

// Logout logs out a user by revoking their refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := s.jwtService.HashToken(refreshToken)
	return s.tokenRepo.Revoke(ctx, tokenHash)
}

// LogoutWithAccessToken logs out a user by revoking refresh token and blacklisting access token
func (s *AuthService) LogoutWithAccessToken(ctx context.Context, refreshToken string, accessTokenJti string, userID uuid.UUID, accessTokenExp time.Time) error {
	// Revoke refresh token
	if refreshToken != "" {
		tokenHash := s.jwtService.HashToken(refreshToken)
		if err := s.tokenRepo.Revoke(ctx, tokenHash); err != nil {
			// Log but continue - access token blacklisting is more important
			// In production: log.Printf("failed to revoke refresh token: %v", err)
		}
	}

	// Add access token to blacklist
	if accessTokenJti != "" {
		if err := s.userRepo.AddTokenToBlacklist(ctx, accessTokenJti, userID, accessTokenExp); err != nil {
			return err
		}
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// RequestPasswordResetInput represents password reset request input
type RequestPasswordResetInput struct {
	Email string `json:"email" binding:"required,email"`
}

// RequestPasswordReset generates a password reset token for a user
// Returns the reset token (in production, this would be sent via email)
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not
		return "", nil
	}

	// Delete any existing reset tokens for this user
	s.resetTokenRepo.DeleteAllForUser(ctx, user.ID)

	// Generate reset token
	resetToken := generateSecureToken()

	// Store hashed token
	tokenRecord := &model.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: s.jwtService.HashToken(resetToken),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.resetTokenRepo.Create(ctx, tokenRecord); err != nil {
		return "", err
	}

	// In production, send email with reset link
	// For now, return the token directly (for testing/development)
	return resetToken, nil
}

// ResetPasswordInput represents password reset input
type ResetPasswordInput struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	// Validate token format
	if len(token) < 32 {
		return errors.New("invalid reset token")
	}

	// Get token record
	tokenHash := s.jwtService.HashToken(token)
	tokenRecord, err := s.resetTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if token is expired or used
	if tokenRecord.ExpiresAt.Before(time.Now()) {
		return errors.New("reset token has expired")
	}
	if tokenRecord.Used {
		return errors.New("reset token has already been used")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, tokenRecord.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	user.PasswordHash = string(passwordHash)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Mark token as used
	s.resetTokenRepo.MarkUsed(ctx, tokenHash)

	// Revoke all refresh tokens (force re-login)
	s.tokenRepo.RevokeAllForUser(ctx, user.ID)

	return nil
}

// generateSecureToken generates a secure random token
func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}