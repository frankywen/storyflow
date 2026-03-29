# 单场景测试功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为每个场景提供独立的视频生成、配音生成、字幕生成功能，方便用户测试单个场景的效果。

**Architecture:** 后端新增两个API端点（单场景配音生成、单场景字幕生成），前端在场景卡片上添加测试按钮组，复用现有的TTS和字幕生成逻辑。

**Tech Stack:** Go/Gin (后端), React/TypeScript (前端), Edge-TTS (配音)

---

## 文件结构

| 文件 | 操作 | 职责 |
|------|------|------|
| `backend/internal/service/audio_service.go` | 修改 | 添加单场景配音生成方法 |
| `backend/internal/service/subtitle_service.go` | 修改 | 添加单场景字幕生成方法 |
| `backend/internal/api/handler/audio_handler.go` | 修改 | 添加单场景API处理函数 |
| `backend/internal/api/router/router.go` | 修改 | 注册新路由 |
| `frontend/src/services/api.ts` | 修改 | 添加新API方法 |
| `frontend/src/App.tsx` | 修改 | 添加场景测试面板 |

---

## Task 1: 后端 - 单场景配音生成服务方法

**Files:**
- Modify: `backend/internal/service/audio_service.go`

- [ ] **Step 1: 添加 GenerateAudioForScene 方法**

在 `audio_service.go` 末尾添加：

```go
// GenerateAudioForScene generates audio for a single scene
func (s *AudioService) GenerateAudioForScene(ctx context.Context, sceneID uuid.UUID, voiceID string) ([]model.AudioFile, error) {
	// Get scene with story info
	scene, err := s.storyRepo.GetScene(ctx, sceneID)
	if err != nil {
		return nil, fmt.Errorf("scene not found: %w", err)
	}

	var audios []model.AudioFile

	// Generate audio for dialogue if exists
	if scene.Dialogue != "" {
		audio, err := s.generateSingleAudio(ctx, scene.ID, scene.StoryID, "dialogue", scene.Dialogue, voiceID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate dialogue audio: %w", err)
		}
		audios = append(audios, *audio)
	}

	// Generate audio for narration if exists
	if scene.Narration != "" {
		audio, err := s.generateSingleAudio(ctx, scene.ID, scene.StoryID, "narration", scene.Narration, voiceID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate narration audio: %w", err)
		}
		audios = append(audios, *audio)
	}

	return audios, nil
}

// generateSingleAudio generates a single audio file
func (s *AudioService) generateSingleAudio(ctx context.Context, sceneID, storyID uuid.UUID, audioType, text, voiceID string) (*model.AudioFile, error) {
	if voiceID == "" {
		voiceID = "zh-CN-XiaoxiaoNeural" // Default voice
	}

	// Generate audio using TTS provider
	result, err := s.ttsProvider.GenerateVoice(ctx, text, voiceID, tts.VoiceParams{})
	if err != nil {
		return nil, err
	}

	// Create audio file record
	audio := &model.AudioFile{
		StoryID:     storyID,
		SceneID:     sceneID,
		AudioType:   audioType,
		TextContent: text,
		AudioURL:    result.AudioURL,
		Duration:    result.Duration,
		VoiceID:     voiceID,
		Status:      "completed",
	}

	if err := s.audioRepo.CreateAudio(ctx, audio); err != nil {
		return nil, fmt.Errorf("failed to save audio: %w", err)
	}

	return audio, nil
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

---

## Task 2: 后端 - 单场景字幕生成服务方法

**Files:**
- Modify: `backend/internal/service/subtitle_service.go`

- [ ] **Step 1: 添加 GenerateSubtitlesForScene 方法**

在 `subtitle_service.go` 末尾添加：

```go
// GenerateSubtitlesForScene generates subtitles for a single scene
func (s *SubtitleService) GenerateSubtitlesForScene(ctx context.Context, sceneID uuid.UUID) ([]model.Subtitle, error) {
	// Get scene
	scene, err := s.storyRepo.GetScene(ctx, sceneID)
	if err != nil {
		return nil, fmt.Errorf("scene not found: %w", err)
	}

	// Get audio files for this scene
	audios, err := s.audioRepo.GetAudiosByScene(ctx, sceneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio files: %w", err)
	}

	if len(audios) == 0 {
		return nil, errors.New("no audio files found for this scene")
	}

	var subtitles []model.Subtitle
	currentTime := 0.0

	for _, audio := range audios {
		subtitle := model.Subtitle{
			StoryID:      scene.StoryID,
			SceneID:      sceneID,
			SubtitleType: audio.AudioType,
			Text:         audio.TextContent,
			StartTime:    currentTime,
			EndTime:      currentTime + audio.Duration,
		}

		if audio.CharacterID != uuid.Nil {
			subtitle.CharacterID = audio.CharacterID
		}

		if err := s.audioRepo.CreateSubtitle(ctx, &subtitle); err != nil {
			return nil, fmt.Errorf("failed to save subtitle: %w", err)
		}

		subtitles = append(subtitles, subtitle)
		currentTime += audio.Duration
	}

	return subtitles, nil
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

---

## Task 3: 后端 - 添加单场景API处理函数

**Files:**
- Modify: `backend/internal/api/handler/audio_handler.go`

- [ ] **Step 1: 添加 GenerateSceneAudio handler**

在 `audio_handler.go` 的 `GetSynthesisStatus` 函数后添加：

```go
// GenerateSceneAudioRequest represents the request body for scene audio generation
type GenerateSceneAudioRequest struct {
	VoiceID   string `json:"voice_id"`
	Overwrite bool   `json:"overwrite"`
}

// GenerateSceneAudio handles POST /api/v1/audio/generate/scene/:scene_id
func (h *AudioHandler) GenerateSceneAudio(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	sceneIDStr := c.Param("scene_id")
	sceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scene_id"})
		return
	}

	var req GenerateSceneAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Request body is optional, use defaults
		req = GenerateSceneAudioRequest{}
	}

	// Verify user owns the story containing this scene
	scene, err := h.storyRepo.GetScene(c.Request.Context(), sceneID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	_, err = h.storyRepo.GetWithRelationsForUser(c.Request.Context(), userID, scene.StoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	// Generate audio
	audios, err := h.audioService.GenerateAudioForScene(c.Request.Context(), sceneID, req.VoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"scene_id": sceneID,
		"audios":   audios,
		"message":  fmt.Sprintf("Generated %d audio files", len(audios)),
	})
}

// GenerateSceneSubtitles handles POST /api/v1/audio/subtitles/scene/:scene_id
func (h *AudioHandler) GenerateSceneSubtitles(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	sceneIDStr := c.Param("scene_id")
	sceneID, err := uuid.Parse(sceneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scene_id"})
		return
	}

	// Verify user owns the story containing this scene
	scene, err := h.storyRepo.GetScene(c.Request.Context(), sceneID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	_, err = h.storyRepo.GetWithRelationsForUser(c.Request.Context(), userID, scene.StoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scene not found"})
		return
	}

	// Generate subtitles
	subtitles, err := h.subtitleService.GenerateSubtitlesForScene(c.Request.Context(), sceneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"scene_id":  sceneID,
		"subtitles": subtitles,
		"message":   fmt.Sprintf("Generated %d subtitles", len(subtitles)),
	})
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

---

## Task 4: 后端 - 注册新路由

**Files:**
- Modify: `backend/internal/api/router/router.go`

- [ ] **Step 1: 添加路由**

找到 `audioGroup` 路由组，添加新路由：

```go
// Audio routes
audioGroup := api.Group("/audio")
audioGroup.POST("/generate", audioHandler.GenerateAudio)
audioGroup.POST("/generate/scene/:scene_id", audioHandler.GenerateSceneAudio)  // 新增
audioGroup.GET("/status/:task_id", audioHandler.GetTaskStatus)
audioGroup.GET("/story/:story_id", audioHandler.GetAudios)
audioGroup.POST("/subtitles/:story_id", audioHandler.GenerateSubtitles)
audioGroup.POST("/subtitles/scene/:scene_id", audioHandler.GenerateSceneSubtitles)  // 新增
audioGroup.GET("/subtitles/:story_id", audioHandler.GetSubtitles)
audioGroup.POST("/synthesis", audioHandler.SynthesizeVideo)
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

---

## Task 5: 前端 - 添加API方法

**Files:**
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: 添加单场景API方法**

在 `audioApi` 对象中添加新方法：

```typescript
// Audio APIs
export const audioApi = {
  generate: (storyId: string) =>
    api.post<{ success: boolean; task_id: string; message: string }>('/audio/generate', { story_id: storyId }),

  // 新增: 单场景配音生成
  generateSceneAudio: (sceneId: string, voiceId?: string) =>
    api.post<{ success: boolean; scene_id: string; audios: AudioFile[]; message: string }>(
      `/audio/generate/scene/${sceneId}`,
      { voice_id: voiceId }
    ),

  getStatus: (taskId: string) =>
    api.get<{
      task_id: string
      status: string
      progress: number
      total_scenes: number
      completed_scenes: number
      failed_scenes: Record<string, string>
    }>(`/audio/status/${taskId}`),

  getAudios: (storyId: string) =>
    api.get<{ audios: AudioFile[] }>(`/audio/story/${storyId}`),

  generateSubtitles: (storyId: string) =>
    api.post<{ success: boolean; message: string }>(`/audio/subtitles/${storyId}`),

  // 新增: 单场景字幕生成
  generateSceneSubtitles: (sceneId: string) =>
    api.post<{ success: boolean; scene_id: string; subtitles: Subtitle[]; message: string }>(
      `/audio/subtitles/scene/${sceneId}`
    ),

  getSubtitles: (storyId: string) =>
    api.get<{ subtitles: Subtitle[] }>(`/audio/subtitles/${storyId}`),

  synthesizeVideo: (storyId: string, options?: {
    video_url?: string
    add_audio?: boolean
    add_subtitle?: boolean
  }) =>
    api.post<{ success: boolean; task_id: string; message: string }>('/audio/synthesis', {
      story_id: storyId,
      ...options,
    }),

  getSynthesisStatus: (taskId: string) =>
    api.get<{
      task_id: string
      status: string
      progress: number
      output_url: string
      error_message: string
    }>(`/videos/synthesis/${taskId}`),
}
```

- [ ] **Step 2: 验证TypeScript编译**

Run: `cd frontend && npm run build`
Expected: 编译成功

---

## Task 6: 前端 - 添加场景测试面板

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: 添加状态变量**

在 `StoryDetailPage` 组件中，找到现有的状态变量，添加：

```tsx
// 场景测试状态
const [sceneAudios, setSceneAudios] = React.useState<Record<string, AudioFile[]>>({})
const [sceneSubtitles, setSceneSubtitles] = React.useState<Record<string, Subtitle[]>>({})
const [generatingScene, setGeneratingScene] = React.useState<string | null>(null)
const [generatingType, setGeneratingType] = React.useState<'audio' | 'subtitle' | 'video' | null>(null)
```

- [ ] **Step 2: 添加处理函数**

在 `handleUploadReference` 函数后添加：

```tsx
// 场景测试处理函数
const handleGenerateSceneAudio = async (sceneId: string) => {
  setGeneratingScene(sceneId)
  setGeneratingType('audio')
  try {
    const res = await audioApi.generateSceneAudio(sceneId)
    setSceneAudios(prev => ({ ...prev, [sceneId]: res.data.audios }))
    alert(`成功生成 ${res.data.audios.length} 个音频文件`)
  } catch (err: any) {
    alert(err.response?.data?.error || '生成配音失败')
  } finally {
    setGeneratingScene(null)
    setGeneratingType(null)
  }
}

const handleGenerateSceneSubtitle = async (sceneId: string) => {
  setGeneratingScene(sceneId)
  setGeneratingType('subtitle')
  try {
    const res = await audioApi.generateSceneSubtitles(sceneId)
    setSceneSubtitles(prev => ({ ...prev, [sceneId]: res.data.subtitles }))
    alert(`成功生成 ${res.data.subtitles.length} 个字幕`)
  } catch (err: any) {
    alert(err.response?.data?.error || '生成字幕失败')
  } finally {
    setGeneratingScene(null)
    setGeneratingType(null)
  }
}

const handleGenerateSceneVideo = async (sceneId: string) => {
  setGeneratingScene(sceneId)
  setGeneratingType('video')
  try {
    const res = await videoApi.generate(id, sceneId, { duration: 5, motion_level: 'medium' })
    alert(`视频生成任务已启动，任务ID: ${res.data.task_id}`)
    // Poll for video completion
    pollForVideos()
  } catch (err: any) {
    alert(err.response?.data?.error || '生成视频失败')
  } finally {
    setGeneratingScene(null)
    setGeneratingType(null)
  }
}
```

- [ ] **Step 3: 修改场景卡片渲染**

找到 `{story.scenes.map((scene) => (` 部分，在场景卡片的 `</div>` 结束标签前（dialogue 显示之后），添加测试面板：

```tsx
{/* 场景测试面板 */}
{scene.image_url && (
  <div className="mt-4 pt-4 border-t border-gray-100">
    <h4 className="text-sm font-medium text-gray-700 mb-2">场景测试</h4>
    <div className="flex gap-2 flex-wrap mb-3">
      {!scene.video_url && (
        <button
          onClick={() => handleGenerateSceneVideo(scene.id)}
          disabled={generatingScene === scene.id && generatingType === 'video'}
          className="px-3 py-1.5 text-sm bg-purple-500 text-white rounded hover:bg-purple-600 disabled:bg-gray-400 flex items-center gap-1"
        >
          <Video className="w-3 h-3" />
          {generatingScene === scene.id && generatingType === 'video' ? '生成中...' : '生成视频'}
        </button>
      )}
      <button
        onClick={() => handleGenerateSceneAudio(scene.id)}
        disabled={generatingScene === scene.id}
        className="px-3 py-1.5 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:bg-gray-400 flex items-center gap-1"
      >
        <Volume2 className="w-3 h-3" />
        {generatingScene === scene.id && generatingType === 'audio' ? '生成中...' : '生成配音'}
      </button>
      <button
        onClick={() => handleGenerateSceneSubtitle(scene.id)}
        disabled={generatingScene === scene.id}
        className="px-3 py-1.5 text-sm bg-green-500 text-white rounded hover:bg-green-600 disabled:bg-gray-400 flex items-center gap-1"
      >
        <Subtitles className="w-3 h-3" />
        {generatingScene === scene.id && generatingType === 'subtitle' ? '生成中...' : '生成字幕'}
      </button>
    </div>

    {/* 测试结果预览 */}
    {sceneAudios[scene.id] && sceneAudios[scene.id].length > 0 && (
      <div className="bg-gray-50 rounded p-2 mt-2">
        <p className="text-xs text-gray-600 mb-2">配音预览:</p>
        {sceneAudios[scene.id].map((audio, idx) => (
          <div key={audio.id || idx} className="flex items-center gap-2 mb-1">
            <span className="text-xs text-gray-500">
              {audio.audio_type === 'dialogue' ? '对话' : '旁白'}
            </span>
            <audio controls src={audio.audio_url} className="h-6 w-40" />
          </div>
        ))}
      </div>
    )}
  </div>
)}
```

- [ ] **Step 4: 添加必要的 import**

确保 `Subtitles` 图标已导入：

```tsx
import { FileText, Image, Settings, Trash2, RefreshCw, Download, Video, Upload, User, Film, LogOut, Shield, Edit3, Save, X, Volume2, Subtitles } from 'lucide-react'
```

- [ ] **Step 5: 验证前端编译**

Run: `cd frontend && npm run build`
Expected: 编译成功

---

## Task 7: 测试和提交

- [ ] **Step 1: 启动后端服务**

Run: `cd backend && go run cmd/server/main.go`
Expected: 服务启动成功

- [ ] **Step 2: 启动前端服务**

Run: `cd frontend && npm run dev`
Expected: 前端启动成功

- [ ] **Step 3: 手动测试**

1. 登录系统
2. 创建故事并解析
3. 生成单个场景的图片
4. 在场景卡片上点击"生成配音"按钮
5. 验证配音生成和播放
6. 点击"生成字幕"按钮
7. 验证字幕生成
8. 点击"生成视频"按钮
9. 验证视频生成

- [ ] **Step 4: 提交代码**

```bash
git add backend/internal/service/audio_service.go \
        backend/internal/service/subtitle_service.go \
        backend/internal/api/handler/audio_handler.go \
        backend/internal/api/router/router.go \
        frontend/src/services/api.ts \
        frontend/src/App.tsx

git commit -m "$(cat <<'EOF'
feat: 添加单场景测试功能

- 新增 POST /api/v1/audio/generate/scene/:scene_id 单场景配音生成
- 新增 POST /api/v1/audio/subtitles/scene/:scene_id 单场景字幕生成
- 前端场景卡片添加测试按钮组（视频/配音/字幕）
- 支持独立测试每个场景的生成效果
EOF
)"
```

---

*计划创建时间: 2026-03-29*