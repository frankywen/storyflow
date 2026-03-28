package tts

import (
	"context"
	"fmt"
	"time"
)

// MockTTSProvider Mock实现（测试用）
type MockTTSProvider struct{}

func NewMockTTSProvider() *MockTTSProvider {
	return &MockTTSProvider{}
}

func (p *MockTTSProvider) GetName() string {
	return "mock"
}

func (p *MockTTSProvider) GenerateVoice(ctx context.Context, text string, voiceID string, params VoiceParams) (*AudioResult, error) {
	// 返回模拟结果
	return &AudioResult{
		AudioURL: fmt.Sprintf("http://localhost:8080/uploads/audio/mock_%d.mp3", time.Now().Unix()),
		Duration: float64(len(text)) * 0.1, // 假设每个字符0.1秒
		VoiceID:  voiceID,
	}, nil
}

func (p *MockTTSProvider) GetAvailableVoices(ctx context.Context) ([]Voice, error) {
	return []Voice{
		{ID: "mock-male", Name: "Mock Male", Gender: "Male", Language: "zh-CN"},
		{ID: "mock-female", Name: "Mock Female", Gender: "Female", Language: "zh-CN"},
	}, nil
}