# StoryFlow 测试清单

> 更新日期: 2026-03-27

## 环境准备

### 必需配置

```env
# .env 文件
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=storyflow

# LLM (必需)
LLM_PROVIDER=claude
LLM_API_KEY=sk-ant-xxx

# 图片生成 (必需)
IMAGE_PROVIDER=comfyui
IMAGE_BASE_URL=http://localhost:8188

# 视频 (可选)
VIDEO_PROVIDER=kling
VIDEO_API_KEY=xxx
```

### 启动服务

```bash
# 1. 启动 PostgreSQL
docker run -d --name postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=storyflow -p 5432:5432 postgres:15

# 2. 启动 ComfyUI (如果使用本地)
cd /path/to/comfyui && python main.py

# 3. 启动后端
cd backend && go run cmd/server/main.go

# 4. 启动前端
cd frontend && npm run dev
```

## 功能测试

### 1. 基础流程 ✅/❌

| 步骤 | 操作 | 预期结果 | 状态 |
|------|------|----------|------|
| 1.1 | 访问首页 | 显示欢迎页面和三个入口卡片 | [ ] |
| 1.2 | 点击「创建故事」 | 跳转到新建故事页面 | [ ] |
| 1.3 | 填写标题和内容 | 表单可正常输入 | [ ] |
| 1.4 | 点击「创建故事」 | 创建成功，跳转到故事详情页 | [ ] |

### 2. 故事解析 ✅/❌

| 步骤 | 操作 | 预期结果 | 状态 |
|------|------|----------|------|
| 2.1 | 选择风格（日式漫画） | 下拉框可选 | [ ] |
| 2.2 | 点击「AI解析故事」 | 显示解析中状态 | [ ] |
| 2.3 | 等待解析完成 | 状态变为 parsed，显示角色和场景 | [ ] |
| 2.4 | 查看角色列表 | 显示角色名称、描述、性别、年龄 | [ ] |
| 2.5 | 查看场景列表 | 显示场景序号、标题、描述、位置、时间 | [ ] |

### 3. 角色一致性 ✅/❌

| 步骤 | 操作 | 预期结果 | 状态 |
|------|------|----------|------|
| 3.1 | 点击「批量生成参考图」 | 开始生成所有角色参考图 | [ ] |
| 3.2 | 查看角色卡片 | 显示生成的参考图 | [ ] |
| 3.3 | 点击「重新生成」 | 重新生成该角色参考图 | [ ] |
| 3.4 | 上传自定义参考图 | 上传成功并显示 | [ ] |
| 3.5 | 删除参考图 | 参考图被清除 | [ ] |

### 4. 图片生成 ✅/❌

| 步骤 | 操作 | 预期结果 | 状态 |
|------|------|----------|------|
| 4.1 | 点击「生成图片」 | 显示生成中状态 | [ ] |
| 4.2 | 等待生成完成 | 场景显示生成的图片 | [ ] |
| 4.3 | 查看图片质量 | 图片符合预期风格 | [ ] |
| 4.4 | 角色一致性检查 | 同一角色在不同场景中外观相似 | [ ] |

### 5. 视频生成 ✅/❌

| 步骤 | 操作 | 预期结果 | 状态 |
|------|------|----------|------|
| 5.1 | 点击「生成视频」 | 显示生成中状态和进度 | [ ] |
| 5.2 | 等待生成完成 | 场景显示视频链接 | [ ] |
| 5.3 | 点击「查看视频」 | 可播放视频 | [ ] |

### 6. 导出功能 ✅/❌

| 步骤 | 操作 | 预期结果 | 状态 |
|------|------|----------|------|
| 6.1 | 点击「PNG」按钮 | 下载 PNG 长图 | [ ] |
| 6.2 | 点击「PDF」按钮 | 下载 PDF 文件 | [ ] |
| 6.3 | 检查导出内容 | 包含标题和所有场景图片 | [ ] |

## API 测试

### 角色一致性 API

```bash
# 1. 上传参考图
curl -X POST http://localhost:8080/api/v1/characters/{id}/reference \
  -F "file=@test.png"

# 2. 删除参考图
curl -X DELETE http://localhost:8080/api/v1/characters/{id}/reference

# 3. 重新生成参考图
curl -X POST http://localhost:8080/api/v1/characters/{id}/regenerate \
  -H "Content-Type: application/json" \
  -d '{"style": "manga"}'

# 4. 批量生成参考图
curl -X POST http://localhost:8080/api/v1/stories/{id}/generate-references \
  -H "Content-Type: application/json" \
  -d '{"style": "manga"}'
```

## 常见问题排查

### 问题 1: 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
docker ps | grep postgres

# 检查连接
psql -h localhost -U postgres -d storyflow
```

### 问题 2: LLM 调用失败

- 检查 `LLM_API_KEY` 是否正确
- 检查网络是否可访问 API 端点
- 查看后端日志错误信息

### 问题 3: 图片生成失败

- ComfyUI: 检查 `http://localhost:8188` 是否可访问
- 云端服务: 检查 API Key 和额度

### 问题 4: 参考图不生效

- 确认角色有 `reference_image_url`
- 检查图片生成器是否支持 img2img
- 查看 `CharacterConsistencyService` 日志

## 测试数据

### 示例故事

```
标题：深夜的便利店

内容：
深夜两点，便利店的灯光在空旷的街道上格外醒目。

"欢迎光临~" 店员小雨有气无力地说道，眼睛盯着手机屏幕。

推门进来的是一个穿着黑色风衣的男人，帽子压得很低，看不清脸。

"那个...有热牛奶吗？" 男人的声音有些沙哑。

小雨抬起头，发现男人正在发抖。

"在里面的冷藏柜，我帮您热一下。"

"谢谢。"

雨声渐渐大了，便利店成了这座不夜城里唯一的温暖角落。
```

### 预期角色

| 角色 | 性别 | 年龄 | 特征 |
|------|------|------|------|
| 小雨 | 女 | 20s | 店员 |
| 神秘男人 | 男 | 30s | 黑色风衣 |

### 预期场景

1. 便利店外观（深夜、灯光）
2. 小雨在柜台
3. 男人推门进入
4. 对话场景
5. 温暖的结尾场景