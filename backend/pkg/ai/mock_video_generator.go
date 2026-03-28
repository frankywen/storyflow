package ai

import (
	"context"
	"fmt"
	"time"
)

// MockVideoGenerator is a mock video generator for testing
type MockVideoGenerator struct {
	config VideoGeneratorConfig
}

// NewMockVideoGenerator creates a new mock video generator
func NewMockVideoGenerator(cfg VideoGeneratorConfig) *MockVideoGenerator {
	return &MockVideoGenerator{config: cfg}
}

// GetName returns the provider name
func (g *MockVideoGenerator) GetName() string {
	return "mock"
}

// GenerateFromImage generates a mock video
func (g *MockVideoGenerator) GenerateFromImage(ctx context.Context, req *VideoRequest) (*VideoResult, error) {
	return &VideoResult{
		TaskID:   fmt.Sprintf("mock-%d", time.Now().Unix()),
		Status:   "pending",
		Progress: 0,
	}, nil
}

// GenerateFromImages generates a mock video from images
func (g *MockVideoGenerator) GenerateFromImages(ctx context.Context, req *VideoRequest, imageUrls []string) (*VideoResult, error) {
	return &VideoResult{
		TaskID:   fmt.Sprintf("mock-%d", time.Now().Unix()),
		Status:   "pending",
		Progress: 0,
	}, nil
}

// GetTaskStatus returns mock task status
func (g *MockVideoGenerator) GetTaskStatus(ctx context.Context, taskID string) (*VideoResult, error) {
	return &VideoResult{
		TaskID:   taskID,
		Status:   "completed",
		Progress: 100,
		VideoURL: "https://example.com/mock-video.mp4",
		Duration: 5.0,
	}, nil
}