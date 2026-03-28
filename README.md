# StoryFlow

AI驱动的小说可视化工具 - 将文字故事转化为漫画图片和视频

## 项目定位

一站式解决小说推文内容生产，实现从文本到可视化内容的自动化流程。

```
小说文本 → 故事解析 → 角色设定 → 分镜生成 → 图片生成 → 视频输出
```

## 功能特性

### 核心功能
- 📖 **故事解析** - AI自动解析小说，提取角色、场景、对话
- 🎨 **图片生成** - 批量生成场景插画，支持多种风格
- 🎬 **视频生成** - 场景图片转视频，支持断点续传
- 🎞️ **视频合成** - 自动合并所有场景视频为完整影片
- 📤 **导出功能** - 支持PNG/PDF导出

### 角色一致性
- 角色参考图生成
- Prompt增强 + Seed锁定
- 场景-角色关联

| 功能 | 描述 | 状态 |
|------|------|------|
| 故事解析 | AI解析小说，提取角色、场景、分镜 | ✅ 已完成 |
| 多LLM支持 | 支持Claude、火山引擎、阿里云通义 | ✅ 已完成 |
| 多图片生成 | 支持ComfyUI、火山引擎、阿里万相 | ✅ 已完成 |
| 角色一致性 | 保持角色在不同场景中的外观一致 | ✅ 已完成 |
| 视频生成 | 图片转视频，支持断点续传 | ✅ 已完成 |
| 视频合成 | 自动合并为完整影片 | ✅ 已完成 |

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.21+ / Gin / GORM / PostgreSQL |
| 前端 | React 18 / TypeScript / Tailwind CSS |
| LLM | 阿里云通义千问 / 火山引擎豆包 / Claude |
| 图片生成 | 火山引擎 doubao-seedream / ComfyUI |
| 视频生成 | 火山引擎 doubao-seedance / Kling / Runway |

## AI 服务支持

### 大语言模型 (LLM) - 用于故事解析

| 提供商 | 配置值 | 模型示例 |
|--------|--------|----------|
| Claude (Anthropic) | `claude` | claude-sonnet-4-20250514 |
| 火山引擎 / 豆包 | `volcengine` | doubao-pro-32k |
| 阿里云 / 通义千问 | `alibaba` | qwen3.5-plus |

### 图片生成服务

| 提供商 | 配置值 | 说明 |
|--------|--------|------|
| ComfyUI | `comfyui` | 本地 GPU 部署，完全免费 |
| 火山引擎 / 豆包 | `volcengine` | 云端 API，按量计费 |
| 阿里云 / 通义万相 | `alibaba` | 云端 API，按量计费 |

### 视频生成服务

| 提供商 | 配置值 | 说明 | 推荐度 |
|--------|--------|------|--------|
| Kling可灵 | `kling` | 性价比高，支持5/10秒 | ⭐⭐⭐⭐⭐ |
| 火山引擎 / 豆包 | `volcengine` | 云端 API，5秒视频 | ⭐⭐⭐⭐ |
| Runway | `runway` | 国际领先，价格较高 | ⭐⭐⭐ |

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- ffmpeg (视频合成)

### 配置

1. 复制配置文件：
```bash
cp backend/.env.example backend/.env
```

2. 编辑 `.env` 配置：

```bash
# LLM 配置 (必须)
LLM_PROVIDER=alibaba
LLM_API_KEY=your-api-key
LLM_MODEL=qwen3.5-plus

# 图片生成 (必须)
IMAGE_PROVIDER=volcengine
IMAGE_API_KEY=your-api-key
IMAGE_MODEL=doubao-seedream-3-0-t2i-250415

# 视频生成 (可选)
VIDEO_PROVIDER=volcengine
VIDEO_API_KEY=your-api-key
VIDEO_MODEL=doubao-seedance-1-0-pro-250528

# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=storyflow
```

### 后端启动

```bash
# 启动数据库
docker-compose up -d postgres

# 安装依赖
cd backend
go mod download

# 启动服务
go run cmd/server/main.go
```

### 前端启动

```bash
cd frontend
npm install
npm run dev
```

### 访问地址
- 前端: http://localhost:3000
- 后端API: http://localhost:8080

## 项目结构

```
storyflow/
├── backend/
│   ├── cmd/server/          # 服务入口
│   ├── internal/
│   │   ├── api/             # API层
│   │   ├── service/         # 业务逻辑
│   │   ├── model/           # 数据模型
│   │   └── repository/      # 数据访问
│   └── pkg/ai/              # AI客户端
├── frontend/
│   └── src/
│       ├── App.tsx          # 主应用
│       ├── types/           # 类型定义
│       └── services/        # API服务
├── docs/                    # 文档
├── docker-compose.yml
└── README.md
```

## API 文档

### 故事管理
```
POST /api/v1/stories                    # 创建故事
GET  /api/v1/stories/:id                # 获取故事详情
POST /api/v1/stories/:id/parse          # AI解析故事
```

### 图片生成
```
POST /api/v1/images/batch               # 批量生成场景图片
GET  /api/v1/images/jobs/:id            # 查询生成任务状态
```

### 视频生成
```
POST /api/v1/videos/generate            # 单场景视频生成
POST /api/v1/videos/batch               # 批量视频生成 (支持断点续传)
GET  /api/v1/videos/status/:task_id     # 查询任务状态
POST /api/v1/videos/merge               # 合成完整视频
GET  /api/v1/videos/view?file=xxx       # 查看合成视频
```

### 导出
```
POST /api/v1/export                     # 导出 PNG/PDF
```

## 开发进度

详见 [docs/STATUS.md](docs/STATUS.md)

## License

MIT