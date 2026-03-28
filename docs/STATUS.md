# StoryFlow 开发进度

> 最后更新: 2026-03-28

## 当前版本: v0.2.2

### 已完成功能

#### 核心功能
- [x] 故事创建和管理
- [x] AI 故事解析（阿里云 qwen3.5-plus）
- [x] 角色提取和场景分镜
- [x] 场景-角色关联
- [x] 图片批量生成（火山引擎 doubao-seedream）
- [x] 图片拼接导出（PNG/PDF）

#### 角色一致性模块 (v0.2.0)
- [x] 角色参考图生成
- [x] 批量生成角色参考图
- [x] 用户上传/替换参考图
- [x] 场景图片生成时使用角色一致性
- [x] Prompt 增强 + Seed 锁定策略
- [x] 场景-角色关联存储

#### 视频生成模块 (v0.2.1-v0.2.2)
- [x] 火山引擎视频生成（图生视频）
- [x] 视频断点续传（跳过已完成场景）
- [x] 视频自动合成（ffmpeg）
- [x] 合成视频持久化存储

### 视频生成支持

| Provider | 状态 | 说明 |
|----------|------|------|
| Kling (可灵) | ✅ 已实现 | 快手 AI 视频生成 |
| Runway Gen-3 | ✅ 已实现 | Runway 视频生成 |
| Volcengine (豆包) | ✅ 已验证 | 图生视频，API端点已确认 |
| Mock | ✅ 已实现 | 测试用 |

### 测试状态

| 功能 | 状态 | 备注 |
|------|------|------|
| 创建故事 | ✅ | API 正常 |
| AI解析 | ✅ | 阿里云 LLM |
| 角色参考图 | ✅ | 火山引擎 2048x2048 |
| 场景-角色关联 | ✅ | 已修复存储问题 |
| 场景图片 | ✅ | 批量生成正常 |
| 场景排序 | ✅ | 已修复顺序问题 |
| 任务状态 | ✅ | 已修复更新问题 |
| 导出 PNG | ✅ | 正常 |
| 导出 PDF | 待测试 | - |
| 视频生成 | ✅ | 火山引擎图生视频已验证 |
| 视频断点续传 | ✅ | 跳过已完成场景 |
| 视频合成 | ✅ | ffmpeg 合并成功 |
| 合成视频持久化 | ✅ | 刷新后仍可查看 |

### 本次修复

| 问题 | 解决方案 |
|------|----------|
| 火山引擎图片尺寸限制 | 改用 2048x2048 |
| 任务状态未更新 | 重载 job 后更新 |
| 场景-角色关联未存储 | 使用 pq.StringArray 类型 |
| 场景顺序错误 | Preload 添加 ORDER BY |

### 视频生成配置说明

```env
# Kling (推荐)
VIDEO_PROVIDER=kling
VIDEO_API_KEY=your-kling-api-key

# Runway
VIDEO_PROVIDER=runway
VIDEO_API_KEY=your-runway-api-key

# 火山引擎 (待验证)
VIDEO_PROVIDER=volcengine
VIDEO_API_KEY=your-volcengine-api-key
VIDEO_MODEL=doubao-seedance-1-0-i2v-250428
```

### 火山引擎视频生成 API 验证 (v0.2.1)

**API 端点：**
- 创建任务: `POST https://ark.cn-beijing.volces.com/api/v3/contents/generations/tasks`
- 查询状态: `GET https://ark.cn-beijing.volces.com/api/v3/contents/generations/tasks/{task_id}`

**请求格式：**
```json
{
  "model": "doubao-seedance-1-0-pro-250528",
  "content": [
    {
      "type": "image_url",
      "image_url": {"url": "https://your-image-url.jpg"}
    }
  ]
}
```

**响应格式：**
```json
{
  "id": "cgt-20260327161505-fch79",
  "status": "succeeded",
  "content": {"video_url": "https://..."},
  "resolution": "1080p",
  "duration": 5,
  "framespersecond": 24
}
```

**测试结果：** ✅ 成功生成 5秒 1080p 视频

**API 端到端测试：**
```bash
# 生成视频
POST /api/v1/videos/generate
Response: {"task_id": "cgt-xxx", "status": "pending"}

# 查询状态 (约10秒后完成)
GET /api/v1/videos/status/cgt-xxx
Response: {"status": "completed", "video_url": "https://...", "progress": 100}
```

### 待处理事项

- [ ] 用户申请 Kling API Key
- [ ] 角色一致性外观优化（用户反馈"外观有些不一样"）

### API 端点

```
POST /api/v1/stories                    # 创建故事
GET  /api/v1/stories/:id                # 获取故事详情
POST /api/v1/stories/:id/parse          # AI解析故事
POST /api/v1/stories/:id/generate-references  # 批量生成角色参考图

POST /api/v1/images/batch               # 批量生成场景图片
GET  /api/v1/images/jobs/:id            # 查询生成任务状态

POST /api/v1/characters/:id/regenerate  # 重新生成参考图
POST /api/v1/characters/:id/reference   # 上传参考图

POST /api/v1/videos/generate            # 单场景视频生成
POST /api/v1/videos/batch               # 批量视频生成（支持断点续传）
GET  /api/v1/videos/status/:task_id     # 查询视频任务状态
POST /api/v1/videos/merge               # 合成完整视频
GET  /api/v1/videos/view?file=xxx.mp4   # 查看合成视频

POST /api/v1/export                     # 导出 PNG/PDF
```

### 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.21+ / Gin / GORM |
| 前端 | React 18 / TypeScript / Tailwind |
| 数据库 | PostgreSQL 15 |
| LLM | 阿里云 qwen3.5-plus |
| 图片生成 | 火山引擎 doubao-seedream |

### 部署信息

```bash
# 启动服务
cd backend && ./server

# 启动前端
cd frontend && npm run dev

# 访问地址
后端: http://localhost:8080
前端: http://localhost:3001
```