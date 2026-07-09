# 阶段四遗留文档

> 阶段 5 项目 Agent prompt 需嵌入此文件。

## 新增表（3 张）

| 表 | 说明 |
|----|------|
| `transcode_tasks` | 转码任务 UNIQUE(video_id), status:0等待/1处理中/2完成/3失败 |
| `video_qualities` | 视频质量 UNIQUE(video_id, quality), 360p/480p/720p/1080p |
| `video_metas` | 视频元数据 UNIQUE(video_id), duration/width/height/codec/bitrate |

## 新增 locked 接口

```
GET /api/v1/videos/:id/qualities  → [{quality, play_url, file_size}]   公开
GET /api/v1/videos/:id/transcode-status → {status, progress}           公开
GET /api/v1/admin/stats            → {total_users,total_videos,...}     role≥3
GET /api/v1/admin/users            → ?page=&q= → {items,total}         role≥3
PUT /api/v1/admin/users/:id/role   → {role} → 200                      role≥3
```

## 新增基础设施

| 组件 | 说明 |
|------|------|
| RabbitMQ | Docker Compose `rabbitmq:3-management` (5672/15672) |
| Go 依赖 | `github.com/rabbitmq/amqp091-go` |

## 架构

```
上传 → videos 表 → RabbitMQ Publish → Worker 消费 → ffprobe(meta) + ffmpeg(HLS) + ffmpeg(cover) → MinIO
```

## 降级

- RabbitMQ 连不上 → goroutine 直调
- ffmpeg 不可用 → 跳过转码 → 原始 mp4 兜底
- 单档失败 → 继续下一档

## 不可突破的约束

- 阶段一/二/三/四 locked 接口不可修改
- ffmpeg/ffprobe 命令通过 `exec.Command` 调用，不依赖 CGO
- 视频上传后必须触发转码（即使 RabbitMQ 不可用）
