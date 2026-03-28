package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"storyflow/internal/model"
	"storyflow/internal/repository"
	"storyflow/pkg/tts"
)

type AudioService struct {
	audioRepo     *repository.AudioRepository
	storyRepo     *repository.StoryRepository
	ttsProvider   tts.Provider
	activeTasks   int32
	maxConcurrent int
	outputDir     string
	audioBaseURL  string
}

func NewAudioService(
	audioRepo *repository.AudioRepository,
	storyRepo *repository.StoryRepository,
	ttsProvider tts.Provider,
	outputDir string,
	audioBaseURL string,
) *AudioService {
	return &AudioService{
		audioRepo:     audioRepo,
		storyRepo:     storyRepo,
		ttsProvider:   ttsProvider,
		maxConcurrent: 3,
		outputDir:     outputDir,
		audioBaseURL:  audioBaseURL,
	}
}

// GenerateAudioRequest 配音生成请求
type GenerateAudioRequest struct {
	StoryID uuid.UUID `json:"story_id"`
}

// GenerateAudio 生成配音（异步）
func (s *AudioService) GenerateAudio(ctx context.Context, req GenerateAudioRequest) (*model.AudioGenerationTask, error) {
	// 检查并发限制
	if atomic.LoadInt32(&s.activeTasks) >= 5 {
		return nil, errors.New("system busy, please try again later")
	}

	// 创建任务
	task := &model.AudioGenerationTask{
		StoryID: req.StoryID,
		Status:  "pending",
	}
	if err := s.audioRepo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	// 异步执行
	go s.processAudioGeneration(context.Background(), task)

	return task, nil
}

func (s *AudioService) processAudioGeneration(ctx context.Context, task *model.AudioGenerationTask) {
	atomic.AddInt32(&s.activeTasks, 1)
	defer atomic.AddInt32(&s.activeTasks, -1)

	// 更新状态
	task.Status = "processing"
	s.audioRepo.UpdateTask(ctx, task)

	// 获取故事和场景
	story, err := s.storyRepo.GetWithRelations(ctx, task.StoryID)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = "Failed to get story"
		s.audioRepo.UpdateTask(ctx, task)
		return
	}

	task.TotalScenes = len(story.Scenes)
	s.audioRepo.UpdateTask(ctx, task)

	// 并发生成音频
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.maxConcurrent)
	var failedScenes []map[string]interface{}
	var mu sync.Mutex

	for i, scene := range story.Scenes {
		wg.Add(1)
		go func(idx int, sc model.Scene) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := s.generateSceneAudio(ctx, task.StoryID, sc); err != nil {
				mu.Lock()
				failedScenes = append(failedScenes, map[string]interface{}{
					"scene_id": sc.ID,
					"error":    err.Error(),
				})
				mu.Unlock()
			}

			task.CompletedScenes = idx + 1
			task.Progress = int(float64(task.CompletedScenes) / float64(task.TotalScenes) * 100)
			s.audioRepo.UpdateTask(ctx, task)
		}(i, scene)
	}

	wg.Wait()

	// 更新最终状态
	if len(failedScenes) > 0 {
		task.FailedScenes = model.JSONB{}
		for _, fs := range failedScenes {
			sceneID, _ := fs["scene_id"].(uuid.UUID)
			task.FailedScenes[sceneID.String()] = fs["error"]
		}
	}
	task.Status = "completed"
	now := time.Now()
	task.CompletedAt = &now
	task.Progress = 100
	s.audioRepo.UpdateTask(ctx, task)
}

func (s *AudioService) generateSceneAudio(ctx context.Context, storyID uuid.UUID, scene model.Scene) error {
	// 默认音色
	defaultVoiceID := "zh-CN-XiaoxiaoNeural"

	// 生成对话音频
	if scene.Dialogue != "" {
		result, err := s.ttsProvider.GenerateVoice(ctx, scene.Dialogue, defaultVoiceID, tts.VoiceParams{})
		if err != nil {
			return fmt.Errorf("dialogue TTS failed: %w", err)
		}

		s.audioRepo.CreateAudio(ctx, &model.AudioFile{
			StoryID:     storyID,
			SceneID:     scene.ID,
			AudioType:   "dialogue",
			TextContent: scene.Dialogue,
			AudioURL:    result.AudioURL,
			Duration:    result.Duration,
			VoiceID:     result.VoiceID,
			Status:      "completed",
		})
	}

	// 生成旁白音频
	if scene.Narration != "" {
		result, err := s.ttsProvider.GenerateVoice(ctx, scene.Narration, defaultVoiceID, tts.VoiceParams{})
		if err != nil {
			return fmt.Errorf("narration TTS failed: %w", err)
		}

		s.audioRepo.CreateAudio(ctx, &model.AudioFile{
			StoryID:     storyID,
			SceneID:     scene.ID,
			AudioType:   "narration",
			TextContent: scene.Narration,
			AudioURL:    result.AudioURL,
			Duration:    result.Duration,
			VoiceID:     result.VoiceID,
			Status:      "completed",
		})
	}

	return nil
}

// GetTaskStatus 获取任务状态
func (s *AudioService) GetTaskStatus(ctx context.Context, taskID uuid.UUID) (*model.AudioGenerationTask, error) {
	return s.audioRepo.GetTask(ctx, taskID)
}

// GetAudiosByStory 获取故事的所有音频
func (s *AudioService) GetAudiosByStory(ctx context.Context, storyID uuid.UUID) ([]model.AudioFile, error) {
	return s.audioRepo.GetAudiosByStory(ctx, storyID)
}