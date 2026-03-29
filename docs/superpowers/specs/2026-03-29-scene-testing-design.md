# 单场景测试功能设计

> **日期**: 2026-03-29
> **版本**: v1.0
> **状态**: 已确认

---

## 概述

为每个场景提供独立的视频生成、配音生成、字幕生成功能，方便用户测试单个场景的效果，无需等待整个故事处理完成。

---

## 一、功能需求

### 1.1 用户场景

用户在开发或调试过程中，经常需要：
- 测试单个场景的视频生成效果
- 测试单个场景的配音效果
- 测试单个场景的字幕显示效果
- 快速验证修改后的效果

### 1.2 功能范围

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 单场景视频生成 | 为指定场景生成视频（已有） | P0 |
| 单场景配音生成 | 为指定场景生成配音 | P0 |
| 单场景字幕生成 | 为指定场景生成字幕 | P0 |
| 场景测试面板 | 在场景卡片上显示测试按钮 | P0 |

---

## 二、API 设计

### 2.1 单场景配音生成

**端点：** `POST /api/v1/audio/generate/scene/:scene_id`

**请求体：**
```json
{
  "voice_id": "zh-CN-XiaoxiaoNeural",  // 可选，指定音色
  "overwrite": false                     // 可选，是否覆盖已有音频
}
```

**响应：**
```json
{
  "success": true,
  "scene_id": "uuid",
  "audios": [
    {
      "id": "uuid",
      "audio_type": "dialogue",
      "text_content": "今天天气真好！",
      "audio_url": "http://.../audio_1.mp3",
      "duration": 2.5
    }
  ],
  "message": "Generated 2 audio files"
}
```

**处理逻辑：**
1. 验证场景存在且用户有权限
2. 获取场景的对话和旁白内容（从 scene.dialogue 和 scene.narration）
3. 为每个对话/旁白调用 TTS 生成音频
4. 保存到 `audio_files` 表，关联 `scene_id`
5. 返回生成的音频列表

### 2.2 单场景字幕生成

**端点：** `POST /api/v1/audio/subtitles/scene/:scene_id`

**请求体：**
```json
{
  "use_existing_audio": true  // 是否使用已有音频的时间轴
}
```

**响应：**
```json
{
  "success": true,
  "scene_id": "uuid",
  "subtitles": [
    {
      "id": "uuid",
      "subtitle_type": "dialogue",
      "text": "今天天气真好！",
      "start_time": 0.0,
      "end_time": 2.5
    }
  ],
  "message": "Generated 2 subtitles"
}
```

**处理逻辑：**
1. 验证场景存在且用户有权限
2. 获取场景的音频文件
3. 根据音频时长生成字幕时间轴（从 0 开始）
4. 保存到 `subtitles` 表
5. 返回生成的字幕列表

### 2.3 单场景视频生成（已有）

**端点：** `POST /api/v1/videos/generate`

已有实现支持 `scene_id` 参数，无需修改。

---

## 三、数据模型

### 3.1 Scene 模型扩展

需要确保 Scene 模型包含对话和旁白字段：

```go
type Scene struct {
    // ... existing fields
    Dialogue   string `json:"dialogue"`    // 对话内容
    Narration  string `json:"narration"`   // 旁白内容
}
```

### 3.2 音频文件关联

`audio_files` 表已支持 `scene_id` 关联，无需修改。

### 3.3 字幕关联

`subtitles` 表已支持 `scene_id` 关联，无需修改。

---

## 四、前端设计

### 4.1 场景卡片扩展

在 `StoryDetailPage` 的场景卡片上添加测试面板：

```tsx
{story.scenes.map((scene) => (
  <div key={scene.id} className="bg-white rounded-lg shadow p-4 border">
    {/* 现有内容 */}

    {/* 测试面板 */}
    {scene.image_url && (
      <div className="mt-4 pt-4 border-t">
        <h4 className="text-sm font-medium mb-2">场景测试</h4>
        <div className="flex gap-2 flex-wrap">
          {!scene.video_url && (
            <button onClick={() => handleGenerateSceneVideo(scene.id)}>
              生成视频
            </button>
          )}
          <button onClick={() => handleGenerateSceneAudio(scene.id)}>
            生成配音
          </button>
          <button onClick={() => handleGenerateSceneSubtitle(scene.id)}>
            生成字幕
          </button>
        </div>

        {/* 测试结果预览 */}
        {sceneAudios[scene.id] && (
          <div className="mt-3">
            <audio controls src={sceneAudios[scene.id].url} />
          </div>
        )}
      </div>
    )}
  </div>
))}
```

### 4.2 状态管理

```tsx
// 场景音频状态
const [sceneAudios, setSceneAudios] = React.useState<Record<string, AudioFile[]>>({})

// 场景字幕状态
const [sceneSubtitles, setSceneSubtitles] = React.useState<Record<string, Subtitle[]>>({})

// 加载状态
const [generatingScene, setGeneratingScene] = React.useState<string | null>(null)
const [generatingType, setGeneratingType] = React.useState<'audio' | 'subtitle' | 'video' | null>(null)
```

---

## 五、实现任务

### Task 1: 后端 - 单场景配音生成 API

**文件：**
- 修改: `backend/internal/api/handler/audio_handler.go`
- 修改: `backend/internal/service/audio_service.go`
- 修改: `backend/internal/api/router/router.go`

**步骤：**
1. 添加 `GenerateSceneAudio` handler
2. 添加 `GenerateAudioForScene` service 方法
3. 注册路由 `POST /audio/generate/scene/:scene_id`

### Task 2: 后端 - 单场景字幕生成 API

**文件：**
- 修改: `backend/internal/api/handler/audio_handler.go`
- 修改: `backend/internal/service/subtitle_service.go`
- 修改: `backend/internal/api/router/router.go`

**步骤：**
1. 添加 `GenerateSceneSubtitles` handler
2. 添加 `GenerateSubtitlesForScene` service 方法
3. 注册路由 `POST /audio/subtitles/scene/:scene_id`

### Task 3: 前端 - 添加场景测试面板

**文件：**
- 修改: `frontend/src/App.tsx` (StoryDetailPage)
- 修改: `frontend/src/services/api.ts`

**步骤：**
1. 添加 `audioApi.generateSceneAudio` API 方法
2. 添加 `audioApi.generateSceneSubtitles` API 方法
3. 在场景卡片添加测试按钮
4. 添加状态管理和回调函数
5. 显示测试结果预览

---

## 六、测试验证

### 6.1 API 测试

```bash
# 单场景配音生成
curl -X POST http://localhost:8080/api/v1/audio/generate/scene/{scene_id} \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json"

# 单场景字幕生成
curl -X POST http://localhost:8080/api/v1/audio/subtitles/scene/{scene_id} \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json"
```

### 6.2 功能测试

1. 创建故事并解析
2. 生成单个场景的图片
3. 点击"生成视频"按钮，验证视频生成
4. 点击"生成配音"按钮，验证配音生成和播放
5. 点击"生成字幕"按钮，验证字幕生成

---

## 七、边界情况

| 情况 | 处理方式 |
|------|----------|
| 场景无对话也无旁白 | 返回空列表，提示"场景无内容" |
| 场景无图片 | 隐藏视频生成按钮 |
| 音频已存在且 overwrite=false | 返回已有音频，不重新生成 |
| TTS 服务不可用 | 返回错误提示 |
| 场景不属于当前用户 | 返回 404 Not Found |

---

*文档创建时间: 2026-03-29*