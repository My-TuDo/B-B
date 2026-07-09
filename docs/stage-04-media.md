---
stage: 阶段四：媒体处理
status: qa-testing
created: 2026-07-08
gate1_by:
locked_by:
deps:
  - stage-01-core (locked)
  - stage-02-discovery (locked)
  - stage-03-community (locked)
  - contracts: auth/user/video/category/tag/search/history/ranking/danmaku/comment/like/coin/favorite/follow/notification (58 APIs)
---

# 阶段四：媒体处理

## 1. 功能概述

对上传的视频进行自动转码、生成多分辨率版本、提取元数据和封面。同时构建管理员后台 Dashboard，提供系统级运营数据可视化。

## 2. 前置依赖

| 依赖 | 来源 | 说明 |
|------|------|------|
| 用户系统 | 阶段一 | JWT Cookie 认证、role 字段 |
| 视频系统 | 阶段一 | 上传/播放/MinIO、videos 表 |
| 中间件 | 阶段一/三 | Auth/CSRF/RateLimit |
| ffmpeg | 系统依赖 | 转码+封面+元数据提取 |
| RabbitMQ | 新增基础设施 | 异步转码队列（3-management，端口 5672/15672） |

## 3. 数据模型

### 3.1 新增表

**transcode_tasks**
| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| id | BIGINT UNSIGNED PK | AUTO_INCREMENT | |
| video_id | BIGINT UNSIGNED | FK→videos, UNIQUE | |
| status | TINYINT | DEFAULT 0 | 0=等待 1=处理中 2=完成 3=失败 |
| progress | TINYINT UNSIGNED | DEFAULT 0 | 0-100 百分比 |
| error_msg | VARCHAR(500) | DEFAULT '' | 失败原因 |
| created_at / updated_at | DATETIME | NOT NULL | |

**video_qualities**
| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| id | BIGINT UNSIGNED PK | AUTO_INCREMENT | |
| video_id | BIGINT UNSIGNED | FK→videos, INDEX | |
| quality | VARCHAR(10) | NOT NULL | 360p/480p/720p/1080p |
| object_name | VARCHAR(500) | NOT NULL | MinIO 路径 |
| file_size | BIGINT UNSIGNED | DEFAULT 0 | |
| created_at | DATETIME | NOT NULL | |
| UNIQUE(video_id, quality) | | | |

**video_metas**
| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| id | BIGINT UNSIGNED PK | AUTO_INCREMENT | |
| video_id | BIGINT UNSIGNED | FK→videos, UNIQUE | |
| duration | FLOAT | NOT NULL | 时长(秒)—覆盖 videos.duration |
| width | INT UNSIGNED | DEFAULT 0 | 原始宽度 |
| height | INT UNSIGNED | DEFAULT 0 | 原始高度 |
| codec | VARCHAR(50) | DEFAULT '' | 编码格式 |
| bitrate | INT UNSIGNED | DEFAULT 0 | 码率(kbps) |
| created_at | DATETIME | NOT NULL | |

### 3.2 修改现有表

**videos 表新增字段**（兼容扩展，只新增可空列）：
- `cover_url` 已被阶段二使用，无需修改
- `duration` 已被阶段二前端回传填充，阶段四用 ffprobe 覆盖

## 4. 功能模块

### 4.1 自动转码 + 多分辨率

**转码流程**（视频上传成功 → 异步启动）：
1. 检查 ffmpeg 是否可用
2. 尝试转码 360P / 480P / 720P / 1080P
3. 如果源分辨率低于目标分辨率，跳过该档（如源 480P → 只转 360P + 480P）
4. ts 分片 + HLS m3u8 索引
5. 上传 MinIO `videos/{videoId}/{quality}/`
6. 写入 video_qualities 表
7. 更新 transcode_tasks status=完成

**本地降级策略**：
- ffmpeg 不可用时跳过转码，原始文件作为唯一播放源
- 用 RabbitMQ 管理转码队列：上传后生产消息 → Worker 消费执行 ffmpeg → 完成后更新状态
- RabbitMQ 环境：Docker Compose 新增服务（`rabbitmq:3-management`，端口 5672/15672）
- 转码失败不影响原始视频播放

### 4.2 HLS 播放

**当前**：直接返回 mp4 预签名 URL → `<video src="mp4">`
**之后**：返回 m3u8 索引 URL → HLS 播放
**降级**：无转码版本时退回 mp4（当前行为）

### 4.3 自动封面

**用户上传封面**（阶段二已实现）→ 优先使用
**用户未上传** → ffmpeg 截取第一帧或 10% 位置的帧作为封面

流程：
1. 转码时同步执行 `ffmpeg -i input -ss 1 -vframes 1 cover.jpg`
2. 上传 MinIO `{userId}/cover_{videoId}_auto.jpg`
3. 如果 videos.cover_url 为空，更新为自动封面路径

### 4.4 ffprobe 元数据提取

转码前执行：
```
ffprobe -v quiet -print_format json -show_format -show_streams input.mp4
```
提取：duration / width / height / codec / bitrate
写入 video_metas 表 + 更新 videos.duration

### 4.5 清晰度切换

**前端**：
- 播放器加清晰度选择按钮（自动 / 360P / 480P / 720P / 1080P）
- 切换时保持播放进度（保存 currentTime → 换源 → 恢复）
- 读取 video_qualities 表获取可用清晰度列表
- HLS 自适应：使用 hls.js 或原生 HLS 支持

**后端**：
- `GET /api/v1/videos/:id/qualities` → `[{quality, play_url, file_size}]`
- play_url 为 MinIO 预签名 URL

### 4.6 转码状态

- `GET /api/v1/videos/:id/transcode-status` → `{status, progress}`（公开）
- 上传成功后前端轮询或 SSE 推送转码进度
- 视频页面显示："转码中 (45%)" / "360P 已就绪" / "1080P 已就绪"

### 4.7 Admin Dashboard

**后端**（需要管理员 role≥3）：
- `GET /api/v1/admin/stats` → `{total_users, total_videos, total_views, total_comments, total_danmaku, today_new_users, today_new_videos}`
- `GET /api/v1/admin/users` → 用户列表 + 搜索 + 分页 + 角色管理
- `GET /api/v1/admin/videos` → 已有（审核队列），扩展支持全部视频管理
- `GET /api/v1/admin/system` → 系统配置（阶段 5 激活，阶段 4 占位）

**前端**：
- `pages/admin.vue` → 当前是占位页，改造为完整 Dashboard
- 统计卡片：用户数/视频数/播放量/评论数
- 折线图：每日新增用户/视频（用简单 SVG 或 ECharts）
- 用户管理表格：列表 + 搜索 + role 编辑

## 5. 非功能约束

- 转码不阻塞上传（goroutine 异步）
- 只转码新的上传（已存在视频不回溯转码）
- ffmpeg 路径：`ffmpeg` / `ffprobe` 在 PATH 中，或可配置
- 清晰度保留原始宽高比
- Admin Dashboard 仅 role≥3 可见
- 不修改阶段一/二/三 locked 接口签名

## 6. 验收条件

### 转码
- [ ] 上传视频 → 异步启动转码 → transcode_tasks 记录创建
- [ ] 转码完成 → video_qualities 写入多分辨率记录
- [ ] 转码失败 → transcode_tasks.status=3 + error_msg
- [ ] ffmpeg 不可用时优雅降级

### 播放
- [ ] `GET /api/v1/videos/:id/qualities` → 返回可用清晰度
- [ ] 前端清晰度切换 → 保持播放进度
- [ ] HLS m3u8 生成 + MinIO 上传

### 封面
- [ ] 未上传封面 → 自动截取第一帧
- [ ] 已上传封面 → 不覆盖

### 元数据
- [ ] ffprobe 提取 duration/width/height/codec
- [ ] 写入 video_metas + 更新 videos.duration

### Admin Dashboard
- [ ] 统计卡片数据正确
- [ ] 用户管理表格
- [ ] 仅 role≥3 可访问

### 编译
- [ ] `go build ./... && go vet ./...` 零错误零警告
- [ ] `pnpm build` 无报错

## 7. 本期不做
- 视频水印/滤镜
- 直播转码
- 旧视频回溯转码

## 8. 已知技术债

| 项 | 说明 |
|----|------|
| ffmpeg 依赖 | 需要系统安装 ffmpeg，Docker 环境可预装 |
| HLS 播放器 | 需替换原生 video 为 Video.js 或 hls.js |
| 转码性能 | RabbitMQ Worker Pool 可横向扩展 |
| ECharts | Admin Dashboard 图表可动态导入或阶段 5 完善 |

## 基础设施新增

### RabbitMQ
Docker Compose 新增 `rabbitmq:3-management` 服务（端口 5672/15672）。

Go 依赖：`github.com/rabbitmq/amqp091-go`

**架构**：
```
上传 → videos 表 → RabbitMQ 生产消息 → Worker 消费 → ffmpeg 转码 → video_qualities 表
```

**降级策略**：
- ffmpeg 不可用 → Worker 跳过 → 原始 mp4 兜底播放
- RabbitMQ 连不上 → 降级为 goroutine 直调（不丢消息）
- 转码失败 → transcode_tasks.status=3 + error_msg → 原始 mp4 仍可播放
- 用户感知：上传成功即播放原始文件，转码在后台异步进行
