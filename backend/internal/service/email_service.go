package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"storyflow/internal/model"
	"storyflow/internal/repository"
)

type EmailService struct {
	repo       *repository.UserRepository
	codeExpire time.Duration
}

func NewEmailService(repo *repository.UserRepository) *EmailService {
	return &EmailService{
		repo:       repo,
		codeExpire: 5 * time.Minute,
	}
}

// GenerateCode 生成6位验证码
func (s *EmailService) GenerateCode() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备方案
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return hex.EncodeToString(b)[:6]
}

// SendVerificationCode 发送验证码
func (s *EmailService) SendVerificationCode(ctx context.Context, email string, codeType string) error {
	// 检查发送频率限制
	if !s.repo.CanSendCode(ctx, email) {
		return errors.New("验证码发送过于频繁，请60秒后再试")
	}

	// 生成验证码
	code := s.GenerateCode()

	// 存储验证码
	verificationCode := &model.EmailVerificationCode{
		Email:     email,
		Code:      code,
		CodeType:  codeType,
		ExpiresAt: time.Now().Add(s.codeExpire),
	}

	if err := s.repo.CreateVerificationCode(ctx, verificationCode); err != nil {
		return err
	}

	// TODO: 集成邮件发送服务
	return nil
}

// VerifyCode 验证验证码
func (s *EmailService) VerifyCode(ctx context.Context, email string, code string, codeType string) error {
	record, err := s.repo.GetValidVerificationCode(ctx, email, code, codeType)
	if err != nil {
		return errors.New("验证码错误")
	}

	if time.Now().After(record.ExpiresAt) {
		return errors.New("验证码已过期")
	}

	if record.IsUsed {
		return errors.New("验证码已使用")
	}

	return s.repo.MarkCodeAsUsed(ctx, record.ID)
}

// GetCodeForDevelopment 开发环境获取验证码
func (s *EmailService) GetCodeForDevelopment(ctx context.Context, email string, codeType string) (string, error) {
	record, err := s.repo.GetLatestVerificationCode(ctx, email, codeType)
	if err != nil {
		return "", err
	}
	return record.Code, nil
}