package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"storyflow/internal/model"
	"storyflow/internal/repository"
)

type SubtitleService struct {
	audioRepo   *repository.AudioRepository
	subtitleDir string
}

func NewSubtitleService(audioRepo *repository.AudioRepository, subtitleDir string) *SubtitleService {
	return &SubtitleService{
		audioRepo:   audioRepo,
		subtitleDir: subtitleDir,
	}
}

// GenerateSubtitles 生成字幕
func (s *SubtitleService) GenerateSubtitles(ctx context.Context, storyID uuid.UUID) error {
	// 获取所有音频文件
	audios, err := s.audioRepo.GetAudiosByStory(ctx, storyID)
	if err != nil {
		return err
	}

	if len(audios) == 0 {
		return nil
	}

	// 删除旧字幕
	s.audioRepo.DeleteSubtitlesByStory(ctx, storyID)

	// 按时间轴生成字幕
	currentTime := 0.0
	for _, audio := range audios {
		subtitle := &model.Subtitle{
			StoryID:      storyID,
			SceneID:      audio.SceneID,
			SubtitleType: audio.AudioType,
			CharacterID:  audio.CharacterID,
			Text:         audio.TextContent,
			StartTime:    currentTime,
			EndTime:      currentTime + audio.Duration,
			StyleConfig: map[string]interface{}{
				"font_size":  24,
				"font_color": "#FFFFFF",
				"background": "#00000080",
				"position":   "bottom-center",
			},
		}

		if err := s.audioRepo.CreateSubtitle(ctx, subtitle); err != nil {
			return err
		}

		currentTime += audio.Duration
	}

	return nil
}

// GenerateSRT 生成SRT字幕文件
func (s *SubtitleService) GenerateSRT(ctx context.Context, storyID uuid.UUID) (string, error) {
	subtitles, err := s.audioRepo.GetSubtitlesByStory(ctx, storyID)
	if err != nil {
		return "", err
	}

	if len(subtitles) == 0 {
		return "", nil
	}

	// 确保目录存在
	if err := os.MkdirAll(s.subtitleDir, 0755); err != nil {
		return "", err
	}

	// 生成SRT内容
	var sb strings.Builder
	for i, sub := range subtitles {
		sb.WriteString(fmt.Sprintf("%d\n", i+1))
		sb.WriteString(fmt.Sprintf("%s --> %s\n",
			formatSRTTime(sub.StartTime),
			formatSRTTime(sub.EndTime),
		))
		sb.WriteString(sub.Text + "\n\n")
	}

	// 写入文件
	filename := fmt.Sprintf("story_%s.srt", storyID.String())
	filePath := filepath.Join(s.subtitleDir, filename)

	if err := os.WriteFile(filePath, []byte(sb.String()), 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

// GetSubtitlesByStory 获取故事字幕
func (s *SubtitleService) GetSubtitlesByStory(ctx context.Context, storyID uuid.UUID) ([]model.Subtitle, error) {
	return s.audioRepo.GetSubtitlesByStory(ctx, storyID)
}

// formatSRTTime 格式化SRT时间
func formatSRTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}