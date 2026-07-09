# 阶段四交付摘要

## 构建验证
- `go build ./...` ✅ + `go vet ./...` ✅
- `pnpm build` ✅

## 新建文件（19 个）

### 后端（14 个）
| 模块 | 说明 |
|------|------|
| `pkg/rabbitmq/rabbitmq.go` | 连接 + 队列声明 + 发布/消费转码任务 |
| `internal/worker/transcode.go` | ffprobe 元数据 + ffmpeg HLS 四档转码 + 自动封面 |
| `internal/model/transcode/quality/meta/` | 3 张新表 |
| `internal/handler/transcode/` + `quality/` | 转码状态 API + 清晰度列表 API |
| `internal/handler/admin/stats.go` | 统计查询 + 用户管理 |

### 前端（5 个）
| 文件 | 说明 |
|------|------|
| `pages/admin.vue` | 占位 → 完整 Dashboard |
| `pages/video/[id].vue` | 清晰度选择器 + 转码轮询 |
| `types/index.ts` | 新 TypeScript 类型 |

### 基础设施
| 文件 | 说明 |
|------|------|
| `docker-compose.yml` | RabbitMQ 服务 (5672/15672) |
| `go.mod` | `rabbitmq/amqp091-go` |
| `pkg/config/config.go` | RabbitMQ 配置 |
| `cmd/server/main.go` | AutoMigrate 3 表 + Worker + 新路由 |

## 架构

```
上传 → videos → RabbitMQ Publish → Worker → ffprobe(meta) + ffmpeg(HLS) + ffmpeg(cover) → MinIO
```

## 降级
- RabbitMQ 连不上 → goroutine 直调
- ffmpeg 不可用 → 跳过转码 → 原始 mp4 兜底
- 单档失败 → 继续下一档
