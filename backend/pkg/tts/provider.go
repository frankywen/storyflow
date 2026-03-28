package tts

import (
	"context"
)

// Voice 音色信息
type Voice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Language    string `json:"language"`
	Description string `json:"description"`
}

// VoiceParams 音色参数
type VoiceParams struct {
	Speed  float64 `json:"speed"`  // 0.5-2.0
	Pitch  int     `json:"pitch"`  // -10 to 10
	Volume int     `json:"volume"` // 0-100
}

// AudioResult 音频生成结果
type AudioResult struct {
	AudioURL string  `json:"audio_url"`
	Duration float64 `json:"duration"`
	VoiceID  string  `json:"voice_id"`
}

// Provider TTS服务提供商接口
type Provider interface {
	// GenerateVoice 生成语音
	GenerateVoice(ctx context.Context, text string, voiceID string, params VoiceParams) (*AudioResult, error)

	// GetAvailableVoices 获取可用音色列表
	GetAvailableVoices(ctx context.Context) ([]Voice, error)

	// GetName 获取Provider名称
	GetName() string
}