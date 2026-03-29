# StoryFlow 开发进度

> 最后更新: 2026-03-29

## 当前版本: v0.3.0

### v0.3 新增功能 ✅

#### 用户系统增强
- [x] JWT 认证（Access Token + Refresh Token）
- [x] 邮箱验证码登录
- [x] 密码重置功能
- [x] Token 黑名单机制
- [x] 多用户数据隔离
- [x] 管理员后台

#### 配音/字幕系统
- [x] TTS Provider 接口（Edge-TTS 实现）
- [x] 异步配音生成任务
- [x] 并发场景处理
- [x] 进度追踪
- [x] 字幕生成（SRT 格式）
- [x] 视频合成（配音+字幕+视频）
- [x] 音频间隔时间配置（对话/旁白间隔）🆕
- [x] 音频淡入淡出效果 🆕
- [x] 音视频同步策略（静音填充/循环/调速）🆕

#### 前端增强
- [x] 登录/注册页面
- [x] 密码找回页面
- [x] 用户配置页面
- [x] 管理后台页面
- [x] 配音配置页面

---

## 已完成功能

### 核心功能
- [x] 故事创建和管理
- [x] AI 故事解析（阿里云 qwen3.5-plus）
- [x] 角色提取和场景分镜
- [x] 场景-角色关联
- [x] 图片批量生成（火山引擎 doubao-seedream）
- [x] 图片拼接导出（PNG/PDF）

### 角色一致性模块 (v0.2.0)
- [x] 角色参考图生成
- [x] 批量生成角色参考图
- [x] 用户上传/替换参考图
- [x] 场景图片生成时使用角色一致性
- [x] Prompt 增强 + Seed 锁定策略

### 视频生成模块 (v0.2.1-v0.2.2)
- [x] 火山引擎视频生成（图生视频）
- [x] 视频断点续传（跳过已完成场景）
- [x] 视频自动合成（ffmpeg）
- [x] 合成视频持久化存储

---

## API 端点

### 认证相关
```
POST /api/v1/auth/register           # 注册
POST /api/v1/auth/login              # 登录
POST /api/v1/auth/logout             # 登出
POST /api/v1/auth/refresh            # 刷新Token
POST /api/v1/auth/send-code          # 发送验证码
POST /api/v1/auth/verify-code        # 验证验证码
POST /api/v1/auth/forgot-password    # 忘记密码
POST /api/v1/auth/reset-password     # 重置密码
GET  /api/v1/auth/me                 # 获取当前用户
```

### 用户配置
```
GET  /api/v1/user/config             # 获取配置
PUT  /api/v1/user/config             # 更新配置
PUT  /api/v1/user/config/llm         # 更新LLM配置
PUT  /api/v1/user/config/image       # 更新图片配置
PUT  /api/v1/user/config/video       # 更新视频配置
```

### 故事管理
```
POST /api/v1/stories                 # 创建故事
GET  /api/v1/stories                 # 故事列表
GET  /api/v1/stories/:id             # 获取故事详情
PUT  /api/v1/stories/:id             # 更新故事
DELETE /api/v1/stories/:id           # 删除故事
POST /api/v1/stories/:id/restore     # 恢复故事
POST /api/v1/stories/:id/parse       # AI解析故事
POST /api/v1/stories/:id/generate-references  # 批量生成角色参考图
```

### 音频/字幕
```
POST /api/v1/audio/generate          # 生成配音
GET  /api/v1/audio/status/:task_id   # 配音任务状态
GET  /api/v1/audio/story/:story_id   # 获取音频列表
POST /api/v1/audio/subtitles/:story_id  # 生成字幕
GET  /api/v1/audio/subtitles/:story_id  # 获取字幕
POST /api/v1/audio/synthesis         # 合成视频
```

### 视频
```
POST /api/v1/videos/generate         # 单场景视频生成
POST /api/v1/videos/batch            # 批量视频生成
GET  /api/v1/videos/status/:task_id  # 视频任务状态
POST /api/v1/videos/merge            # 合成完整视频
GET  /api/v1/videos/synthesis/:task_id  # 视频合成状态
```

### 角色管理
```
POST /api/v1/characters/:id/reference     # 上传参考图
DELETE /api/v1/characters/:id/reference   # 删除参考图
POST /api/v1/characters/:id/regenerate    # 重新生成参考图
```

### 导出
```
POST /api/v1/export                  # 导出 PNG/PDF
```

### 管理后台
```
GET  /api/v1/admin/stats             # 系统统计
GET  /api/v1/admin/users             # 用户列表
GET  /api/v1/admin/users/:id         # 用户详情
PUT  /api/v1/admin/users/:id         # 更新用户
POST /api/v1/admin/users/:id/suspend # 暂停用户
POST /api/v1/admin/users/:id/activate # 激活用户
DELETE /api/v1/admin/users/:id       # 删除用户
```

---

## 视频生成支持

| Provider | 状态 | 说明 |
|----------|------|------|
| Kling (可灵) | ✅ 已实现 | 快手 AI 视频生成 |
| Runway Gen-3 | ✅ 已实现 | Runway 视频生成 |
| Volcengine (豆包) | ✅ 已验证 | 图生视频 |
| Mock | ✅ 已实现 | 测试用 |

## TTS 支持场景

| Provider | 状态 | 说明 |
|----------|------|------|
| Edge-TTS | ✅ 已实现 | 免费，微软在线TTS |
| Mock | ✅ 已实现 | 测试用 |

---

## 环境变量配置

```env
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=storyflow

# JWT
JWT_ACCESS_SECRET=your-access-secret
JWT_REFRESH_SECRET=your-refresh-secret

# TTS 配置
TTS_OUTPUT_DIR=./uploads/audio
AUDIO_BASE_URL=http://localhost:8080/uploads/audio

# 字幕配置
SUBTITLE_DIR=./uploads/subtitles

# 视频合成配置
SYNTHESIS_OUTPUT_DIR=./uploads/synthesis
SYNTHESIS_BASE_URL=http://localhost:8080/uploads/synthesis

# 音频间隔时间配置 (v0.3 新增)
AUDIO_GAP_BETWEEN_DIALOGUES=0.3    # 对话间隔(秒)
AUDIO_GAP_BEFORE_NARRATION=0.5     # 旁白前间隔(秒)
AUDIO_GAP_AFTER_NARRATION=0.3      # 旁白后间隔(秒)
AUDIO_GAP_BETWEEN_CHARACTERS=0.5   # 不同角色对话间隔(秒)

# 音频淡入淡出效果配置 (v0.3 新增)
AUDIO_FADE_IN_DURATION=0.3         # 淡入时长(秒)
AUDIO_FADE_OUT_DURATION=0.3        # 淡出时长(秒)

# 音视频同步策略配置 (v0.3 新增)
AUDIO_SYNC_MODE=silence            # 同步模式: silence(静音填充)/loop(循环)/speed(调速)

# 服务端口
PORT=8080
```

---

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.21+ / Gin / GORM |
| 前端 | React 18 / TypeScript / Tailwind |
| 数据库 | PostgreSQL 15 |
| LLM | 阿里云 qwen3.5-plus |
| 图片生成 | 火山引擎 doubao-seedream |
| TTS | Edge-TTS (免费) |
| 视频合成 | ffmpeg |

---

## 部署信息

```bash
# 启动后端
cd backend && go run cmd/server/main.go

# 启动前端
cd frontend && npm run dev

# 访问地址
后端: http://localhost:8080
前端: http://localhost:5173
```

---

## 使用流程

```
1. 输入故事 → 2. AI解析 → 3. 生成图片 → 4. 配音字幕 → 5. 视频生成 → 6. 导出
```

---

## 待处理事项

- [ ] 角色一致性外观优化
- [ ] 更多 TTS Provider 支持（火山引擎、阿里云等）
- [ ] 字幕样式自定义
- [ ] 字幕换行规则优化（10.6）
- [ ] 缓存失效策略细化（10.7）
- [ ] 多设备并发操作（10.8）
- [ ] 错误提示文案优化（10.9）
- [ ] TTS重试策略细化（10.10）
- [ ] 并发参数调优（10.4）