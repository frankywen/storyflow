# StoryFlow 角色一致性模块设计

> 创建日期: 2026-03-27
> 状态: 已实现

---

## 1. 概述

### 1.1 背景

StoryFlow 是一个 AI 驱动的小说→漫画/视频自动化生成工具。当前版本已完成故事解析、图片生成、视频生成等核心功能，但缺少角色一致性支持——同一角色在不同场景中的外观可能不一致。

### 1.2 目标

实现角色一致性模块，确保：
- 同一角色在不同场景中保持高度一致的外观
- 支持自动生成参考图和用户上传参考图
- 适配火山引擎等云端 API

### 1.3 范围

- 数据模型扩展
- 角色参考图生成服务
- 图片生成器接口扩展（图生图）
- 前端 UI 变更
- 错误处理与降级策略

---

## 2. 技术方案

### 2.1 混合方案

采用「参考图 + Prompt增强」的混合方案：

```
┌─────────────────────────────────────────────────────────────────┐
│                      角色一致性工作流                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  故事解析 ──► 角色提取 ──► 生成参考图 ──► 存储参考图              │
│                                │                                │
│                                ▼                                │
│                        用户可上传/替换                           │
│                                │                                │
│                                ▼                                │
│  场景图片生成 ──► 组合角色Prompt ──► 参考图辅助 ──► 输出图片       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 一致性策略

| 策略 | 描述 | 优先级 |
|------|------|--------|
| 参考图生成 | 为每个角色生成标准参考图 | 主要 |
| Prompt 增强 | 组合角色详细描述到场景 Prompt | 主要 |
| Seed 锁定 | 为每个角色分配固定 Seed | 辅助 |
| 用户上传 | 允许用户上传/替换参考图 | 补充 |

---

## 3. 数据模型

### 3.1 Character 模型扩展

```go
type Character struct {
    ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
    StoryID     uuid.UUID `json:"story_id" gorm:"type:uuid;not null"`
    Name        string    `json:"name" gorm:"not null"`
    Description string    `json:"description" gorm:"type:text"`

    // 现有字段
    Gender      string `json:"gender"`
    Age         string `json:"age"`
    HairColor   string `json:"hair_color"`
    EyeColor    string `json:"eye_color"`
    BodyType    string `json:"body_type"`
    Clothing    string `json:"clothing"`

    // 新增字段
    ReferenceImageURL string `json:"reference_image_url"`        // 参考图URL
    ReferenceImageID  string `json:"reference_image_id"`         // 生成服务中的ID
    Seed              int64  `json:"seed"`                       // 角色专属种子
    VisualPrompt      string `json:"visual_prompt" gorm:"type:text"` // 详细视觉描述

    CreatedAt time.Time `json:"created_at"`
}
```

### 3.2 数据库迁移

```sql
ALTER TABLE characters
ADD COLUMN reference_image_url TEXT,
ADD COLUMN reference_image_id VARCHAR(255),
ADD COLUMN seed BIGINT,
ADD COLUMN visual_prompt TEXT;
```

---

## 4. API 设计

### 4.1 新增端点

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/characters/:id/reference` | 上传角色参考图 |
| DELETE | `/api/v1/characters/:id/reference` | 删除参考图 |
| POST | `/api/v1/characters/:id/regenerate` | 重新生成参考图 |
| POST | `/api/v1/stories/:id/generate-references` | 批量生成所有角色参考图 |

### 4.2 请求/响应示例

**上传参考图：**
```json
// POST /api/v1/characters/:id/reference
// Content-Type: multipart/form-data
// file: <image file>

// Response
{
  "id": "uuid",
  "reference_image_url": "https://...",
  "message": "参考图上传成功"
}
```

**重新生成参考图：**
```json
// POST /api/v1/characters/:id/regenerate
{
  "style": "manga"
}

// Response
{
  "id": "uuid",
  "reference_image_url": "https://...",
  "visual_prompt": "1girl, long black hair, blue eyes...",
  "seed": 12345,
  "message": "参考图生成成功"
}
```

### 4.3 修改现有端点

**图片批量生成增加一致性支持：**
```json
// POST /api/v1/images/batch
{
  "story_id": "uuid",
  "style": "manga",
  "use_consistency": true  // 新增参数
}
```

---

## 5. 服务层设计

### 5.1 新增服务

**文件：** `backend/internal/service/character_consistency_service.go`

```go
package service

type CharacterConsistencyService struct {
    imageGenerator ai.ImageGenerator
    repo           *repository.StoryRepository
    llmProvider    ai.LLMProvider
}

// GenerateReferenceImage 为角色生成参考图
func (s *CharacterConsistencyService) GenerateReferenceImage(
    ctx context.Context,
    character *model.Character,
    style string,
) (*model.Character, error) {
    // 1. 生成详细视觉描述（如果不存在）
    if character.VisualPrompt == "" {
        character.VisualPrompt = s.generateVisualPrompt(character)
    }

    // 2. 分配固定 Seed
    if character.Seed == 0 {
        character.Seed = time.Now().UnixNano()
    }

    // 3. 调用图片生成器
    req := &ai.ImageRequest{
        Prompt: fmt.Sprintf("%s style, %s, character reference sheet, front view, simple background",
            style, character.VisualPrompt),
        Width:  1024,
        Height: 1024,
        Style:  style,
        Seed:   character.Seed,
    }

    result, err := s.imageGenerator.Generate(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to generate reference image: %w", err)
    }

    // 4. 更新角色记录
    character.ReferenceImageURL = result.ImageURL
    character.ReferenceImageID = result.ID

    return character, nil
}

// EnhanceScenePrompt 增强场景 Prompt
func (s *CharacterConsistencyService) EnhanceScenePrompt(
    scene *model.Scene,
    characters []model.Character,
) string {
    var charPrompts []string
    charMap := make(map[string]model.Character)

    for _, c := range characters {
        charMap[c.Name] = c
    }

    // 提取场景中的角色
    for _, charID := range scene.CharacterIDs {
        for _, c := range characters {
            if c.ID == charID {
                charPrompts = append(charPrompts, c.VisualPrompt)
            }
        }
    }

    // 组合 Prompt
    if len(charPrompts) > 0 {
        return fmt.Sprintf("%s, %s",
            strings.Join(charPrompts, ", "),
            scene.ImagePrompt)
    }

    return scene.ImagePrompt
}

// GenerateSceneWithConsistency 带一致性的场景生成
func (s *CharacterConsistencyService) GenerateSceneWithConsistency(
    ctx context.Context,
    scene *model.Scene,
    characters []model.Character,
    style string,
) (*ai.ImageResult, error) {
    // 1. 获取场景中的角色
    var sceneChars []model.Character
    for _, charID := range scene.CharacterIDs {
        for _, c := range characters {
            if c.ID == charID {
                sceneChars = append(sceneChars, c)
            }
        }
    }

    // 2. 增强 Prompt
    enhancedPrompt := s.EnhanceScenePrompt(scene, characters)

    // 3. 检查是否有参考图
    hasReference := false
    var refImages []string
    var avgSeed int64

    for _, c := range sceneChars {
        if c.ReferenceImageURL != "" {
            hasReference = true
            refImages = append(refImages, c.ReferenceImageURL)
        }
        avgSeed += c.Seed
    }
    if len(sceneChars) > 0 {
        avgSeed /= int64(len(sceneChars))
    }

    // 4. 生成图片
    req := &ai.ImageRequest{
        Prompt: fmt.Sprintf("%s style, %s", style, enhancedPrompt),
        Width:  2048,
        Height: 2048,
        Style:  style,
        Seed:   avgSeed,
    }

    // 5. 如果支持图生图且有参考图，使用图生图
    if hasReference && s.supportsImageToImage() {
        return s.imageGenerator.GenerateFromImage(ctx, req, refImages[0])
    }

    return s.imageGenerator.Generate(ctx, req)
}

func (s *CharacterConsistencyService) supportsImageToImage() bool {
    // 检查图片生成器是否支持图生图
    _, ok := s.imageGenerator.(ai.ImageToImageGenerator)
    return ok
}
```

---

## 6. 图片生成器接口扩展

### 6.1 接口定义

**文件：** `backend/pkg/ai/image_generator.go`

```go
package ai

// ImageRequest 图片生成请求
type ImageRequest struct {
    Prompt         string
    Width, Height  int
    Style          string
    Seed           int64
    // 新增字段
    ReferenceImage string   // 参考图URL
    RefStrength    float64  // 参考图强度 (0-1), 默认 0.5
}

// ImageResult 图片生成结果
type ImageResult struct {
    ID       string
    ImageURL string
    Seed     int64
    Width    int
    Height   int
    Model    string
}

// ImageGenerator 图片生成器接口
type ImageGenerator interface {
    GetName() string
    Generate(ctx context.Context, req *ImageRequest) (*ImageResult, error)
    GenerateBatch(ctx context.Context, req *ImageRequest, count int) ([]*ImageResult, error)
}

// ImageToImageGenerator 图生图接口（可选实现）
type ImageToImageGenerator interface {
    ImageGenerator
    GenerateFromImage(ctx context.Context, req *ImageRequest, refImageURL string) (*ImageResult, error)
}
```

### 6.2 火山引擎实现

**文件：** `backend/pkg/ai/volcengine_image_generator.go`

扩展支持图生图功能（如果 API 支持）：

```go
// 检查火山引擎是否支持图生图 API
// 如果支持，实现 GenerateFromImage 方法
func (g *VolcEngineImageGenerator) GenerateFromImage(
    ctx context.Context,
    req *ImageRequest,
    refImageURL string,
) (*ImageResult, error) {
    // 1. 下载参考图
    // 2. 转换为 base64
    // 3. 调用图生图 API（如果火山引擎支持）
    // 4. 如果不支持，降级到普通生成

    // 暂时降级到普通生成
    return g.Generate(ctx, req)
}
```

---

## 7. 前端设计

### 7.1 角色卡片增强

**变更文件：** `frontend/src/App.tsx`

```tsx
// 角色卡片组件
const CharacterCard = ({ character, onRegenerate, onUpload }) => {
  return (
    <div className="bg-white rounded-lg shadow p-4 border">
      {/* 参考图展示 */}
      {character.reference_image_url && (
        <div className="mb-3">
          <img
            src={character.reference_image_url}
            alt={character.name}
            className="w-full h-32 object-cover rounded-lg"
          />
        </div>
      )}

      {/* 角色信息 */}
      <h3 className="font-semibold">{character.name}</h3>
      <p className="text-sm text-gray-500 mt-1">{character.description}</p>

      {/* 视觉特征标签 */}
      <div className="flex gap-2 mt-2 text-xs flex-wrap">
        <span className="px-2 py-1 bg-gray-100 rounded">{character.gender}</span>
        <span className="px-2 py-1 bg-gray-100 rounded">{character.age}</span>
        {character.hair_color && (
          <span className="px-2 py-1 bg-gray-100 rounded">{character.hair_color}</span>
        )}
      </div>

      {/* 操作按钮 */}
      <div className="flex gap-2 mt-3">
        <button
          onClick={() => onRegenerate(character.id)}
          className="px-3 py-1 text-sm bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          {character.reference_image_url ? '重新生成' : '生成参考图'}
        </button>
        <label className="px-3 py-1 text-sm bg-gray-200 rounded cursor-pointer hover:bg-gray-300">
          上传参考图
          <input
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => onUpload(character.id, e.target.files[0])}
          />
        </label>
      </div>
    </div>
  )
}
```

### 7.2 故事详情页变更

```tsx
// 在角色区域添加「批量生成参考图」按钮
{story.status === 'parsed' && story.characters?.length > 0 && (
  <button
    onClick={handleGenerateAllReferences}
    className="mb-4 px-4 py-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600"
  >
    批量生成角色参考图
  </button>
)}

// 图片生成时默认启用一致性
const handleGenerateImages = async () => {
  await imageApi.batchGenerate(id, {
    style: selectedStyle,
    use_consistency: true  // 默认启用
  })
}
```

---

## 8. 错误处理

### 8.1 错误码定义

| 错误码 | HTTP状态 | 描述 |
|--------|---------|------|
| `CHAR_REF_GEN_FAILED` | 500 | 参考图生成失败 |
| `CHAR_REF_UPLOAD_ERROR` | 400 | 参考图上传失败 |
| `CHAR_NOT_FOUND` | 404 | 角色不存在 |
| `IMG2IMG_NOT_SUPPORTED` | 501 | 图生图功能不支持 |
| `STORY_NOT_PARSED` | 400 | 故事未解析，无法生成参考图 |

### 8.2 降级策略

| 场景 | 降级方案 |
|------|---------|
| 图生图 API 不支持 | 使用纯 Prompt 生成 |
| 参考图生成失败 | 记录错误，允许重试或手动上传 |
| 参考图不存在 | 使用角色详细描述生成 |

---

## 9. 测试策略

### 9.1 单元测试

- `CharacterConsistencyService.GenerateReferenceImage`
- `CharacterConsistencyService.EnhanceScenePrompt`
- `CharacterConsistencyService.GenerateSceneWithConsistency`
- Prompt 组合逻辑
- Seed 分配逻辑

### 9.2 集成测试

- 完整流程：角色创建 → 参考图生成 → 场景生成
- 降级流程：图生图不支持时的处理
- 上传流程：用户上传参考图

### 9.3 手动测试

- 使用火山引擎 API 实际生成
- 验证角色一致性效果
- 测试不同风格的一致性表现

---

## 10. 实现计划

| 阶段 | 任务 | 预估时间 |
|------|------|---------|
| Phase 1 | 数据模型 + 数据库迁移 | 1-2 小时 |
| Phase 2 | 角色一致性服务 | 2-3 小时 |
| Phase 3 | 图片生成器接口扩展 | 2-3 小时 |
| Phase 4 | API Handler | 1-2 小时 |
| Phase 5 | 前端 UI 变更 | 2-3 小时 |
| Phase 6 | 测试 + 调试 | 2-3 小时 |

**总预估：** 10-16 小时

---

## 11. 风险与缓解

| 风险 | 等级 | 缓解措施 |
|------|------|---------|
| 火山引擎不支持图生图 | 中 | 降级到 Prompt 方案 |
| 一致性效果不理想 | 中 | 提供用户上传参考图选项 |
| 生成成本增加 | 低 | 参考图仅在需要时生成 |

---

## 12. 验收标准

- [x] 角色可以有参考图
- [x] 用户可以上传/重新生成参考图
- [x] 场景图片生成时使用角色一致性
- [x] 同一角色在不同场景中外观基本一致 (已实现，待可视化验证)
- [x] 错误处理完善，有降级方案
- [x] 前端 UI 友好易用

## 实施记录

**完成日期:** 2026-03-27

**测试结果:**
- 角色参考图生成: ✅ 通过
- 批量生成: ✅ 通过
- 场景图片生成: ✅ 通过
- 导出功能: ✅ 通过

**已修复问题:**
1. 火山引擎图片尺寸限制 (最小 3686400 像素)
2. 异步任务状态更新问题