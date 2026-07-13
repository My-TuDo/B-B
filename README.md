# B-B

> 仿制 B 站项目介绍与技术分析 — Go + Nuxt 全栈实现

## 项目概述

B-B 是一个仿 B 站的视频分享平台，从前端 UI 到后端 API 完整实现。项目采用**渐进式开发**，从核心骨架到媒体处理，分 5 个阶段迭代，每个阶段产出可演示的完整功能。

核心技术决策围绕三个目标：**高并发可扩展**（Go + Redis + RabbitMQ）、**开发体验**（Nuxt 4 SSR + Tailwind 暗色双模）和**可学习性**（清晰的 Handler → Service → Repository 三层架构，中文注释覆盖所有文件）。

## 技术架构

```
浏览器 ──→ Nginx (443 HTTPS) ──→ Nuxt SSR (3000)  前端页面渲染
                              ──→ Go API  (8080)   后端 JSON API
                                       ├── MySQL   业务数据
                                       ├── Redis   缓存 + Token 白名单
                                       ├── MinIO   视频/图片存储
                                       └── RabbitMQ 异步转码队列
```

### 选型分析

| 技术 | 选择原因 |
|------|---------|
| **Go + Gin** | 原生并发 + 低内存，适合 I/O 密集的视频服务 |
| **GORM** | AutoMigrate 根据 struct 自动建表/更新，减少 SQL 手写 |
| **Nuxt 4 SSR** | 服务端渲染保证 SEO，同时支持 SPA 客户端导航 |
| **MinIO** | S3 兼容、Docker 单容器部署，bucket 公开策略免去 CDN |
| **Redis** | Token 白名单（支持强制下线）+ 排行榜 ZSet + 热度缓存 |
| **RabbitMQ** | 异步解耦视频转码，不可用时降级为 goroutine 直调 |

### 后端三层架构

```
Handler  ─→  HTTP 层：解析请求参数，调用 Service，返回统一 JSON
Service  ─→  业务逻辑：校验权限，编排多个 Repository，缓存/通知
Repository →  数据访问：封装 GORM 查询，单表操作
```

这种分层的核心价值：**Handler 不知道表名，Repository 不知道 HTTP，Service 不知道 Gin**。每一层只依赖下一层的接口，测试时可以独立 mock。

### 安全体系

| 层 | 机制 | 原理 |
|----|------|------|
| 认证 | JWT + Redis 白名单 | Cookie 传递 token，Redis 存储最新 token 支持服务端踢人 |
| 防 CSRF | Double-Submit Cookie | 前端读 Cookie 写 Header，恶意网站无法读取跨域 Cookie |
| 限流 | 令牌桶 + IP 计数 | 全局 100 req/s + 登录接口 IP 粒度 5 次/分钟 |
| 文件校验 | MIME + 魔数 + 大小 | 读文件头 512 字节用 `http.DetectContentType` 验真伪 |

## 快速开始

```bash
# 一键部署（自动生成 SSL 证书 + 构建 + 启动）
./deploy.sh

# 或分步启动
docker compose up -d

# 本地开发（只启动依赖，前后端单独跑）
docker compose up -d mysql redis minio rabbitmq
cd backend && go run ./cmd/server/main.go
cd frontend && pnpm install && pnpm dev
```

## 后端架构详解

### 启动流程

```
config.Load() → logger.Init() → jwt.Init() → validator.Init()
  → database.InitMySQL() → database.InitRedis() → storage.Init(MinIO)
  → middleware.InitAuth() → ws.InitHub()
  → db.AutoMigrate(14 models) → seedCategories()
  → rabbitmq.Init() → 启动 Consumer goroutine
  → gin.New() + 中间件链 → 19 模块注册路由 → r.Run(":8080")
```

### 中间件链（按序执行）

| 顺序 | 中间件 | 职责 |
|------|--------|------|
| 1 | Recovery | panic 捕获 + 堆栈日志 + 500 响应 |
| 2 | RequestID | 生成/继承 UUID，写回响应头 |
| 3 | Logger | 结构化请求日志（method/path/status/latency/requestID） |
| 4 | CORS | 跨域白名单 + OPTIONS 预检 204 直返 |
| 5 | RateLimit | 全局令牌桶 + 登录接口 IP 级限流 |
| 6 | CSRF | 写请求强制比对 Header.CSRF-Token 与 Cookie.csrf_token |
| * | AuthRequired | 路由级：JWT 解析 + Redis 白名单验证 |

### 请求生命周期（以点赞为例）

```
POST /api/v1/videos/7/like
  → Recovery → RequestID → Logger → CORS → RateLimit → CSRF
  → AuthRequired（JWT 验证 + Redis 白名单）
  → handler.ToggleLike
    → c.Param("id") → 7，middleware.GetUserID(c) → 4
    → svc.ToggleLike(ctx, 4, 7)
      → repo.Exists(ctx, 4, 7)        → SELECT COUNT(*) FROM video_likes
      → repo.Delete/Create             → INSERT/DELETE
      → videoRepo.IncrementLikeCount   → UPDATE videos SET likes = likes ± 1
      → messageSvc.NotifyAuthor        → INSERT notifications
    → response.Success(c, {liked, count})
  → 返回 {"code":200, "message":"成功", "data":{...}, "request_id":"uuid"}
```

## 核心功能实现分析

### 视频上传与转码

视频上传采用 **SSE 流式进度**：自定义 `progressReader` 包装 `io.Reader`，每次读块回调进度百分比，通过 `c.Stream()` 推给前端。

上传后触发转码：优先 RabbitMQ 异步队列，连不上则 goroutine 直调。转码 Worker 调用 ffmpeg 输出 360p/480p/720p/1080p 四档 HLS（m3u8 + ts 分片），同时 ffprobe 提取元数据写入 `video_metas` 表。

### 弹幕系统（Bilibili 风格）

弹幕引擎采用 **requestAnimationFrame 循环 + 视频时间驱动**，每条弹幕位置由 `(currentTime - play_time) / 4s` 实时计算。暂停/回退/快进时，弹幕位置同步跟随视频进度，不再使用 CSS 固定时长的 animation。

WebSocket Hub 按 videoID 分房间广播，REST API 拉取历史弹幕。

### 热度算法与缓存

首页推荐使用加权公式：`views×0.5 + likes×2 + comments×3 - hours×0.1`，通过 Redis ZSet 排序，TTL 10 分钟。排行榜按 views 用 ZSet 维护日/周/总三个维度。

### 缩略图与封面

自动封面由转码 Worker 在检测到 `cover_url` 为空时调用 ffmpeg 截取第一帧（`-ss 00:00:01 -vframes 1`）。已上传自定义封面的视频不覆盖。

## API 接口

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/logout` | 登出 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| GET | `/api/v1/auth/me` | 当前用户信息 |

### 用户
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/users/:id` | 用户信息 |
| PUT | `/api/v1/users/:id` | 更新用户 |
| POST | `/api/v1/users/:id/avatar` | 上传头像 |
| POST | `/api/v1/users/:id/follow` | 关注/取关 |

### 视频
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/videos` | 上传视频（SSE 进度） |
| GET | `/api/v1/videos` | 视频列表（分页+分类） |
| GET | `/api/v1/videos/:id` | 视频详情 |
| GET | `/api/v1/videos/:id/play-url` | 播放地址 |
| GET | `/api/v1/videos/:id/qualities` | 清晰度列表 |
| GET | `/api/v1/videos/:id/transcode-status` | 转码状态 |
| GET | `/api/v1/videos/hot` | 首页推荐 |
| GET | `/api/v1/videos/ranking` | 排行榜 |
| PUT | `/api/v1/videos/:id` | 更新视频 |
| DELETE | `/api/v1/videos/:id` | 删除视频 |

### 社区
| 方法 | 路径 | 说明 |
|------|------|------|
| WS | `/api/v1/ws/danmaku/:video_id` | 弹幕 WebSocket |
| GET/POST | `/api/v1/videos/:id/danmaku` | 弹幕历史/发送 |
| GET/POST | `/api/v1/videos/:id/comments` | 评论列表/创建 |
| POST | `/api/v1/videos/:id/like` | 点赞/取消 |
| POST | `/api/v1/videos/:id/coin` | 投币 |
| GET/POST | `/api/v1/favorites` | 收藏夹 |
| GET | `/api/v1/history` | 观看历史 |
| GET | `/api/v1/feed` | 关注 Feed |

### 管理员（role ≥ 3）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/stats` | 统计面板 |
| GET | `/api/v1/admin/users` | 用户管理 |
| PUT | `/api/v1/admin/users/:id/role` | 修改角色 |
| GET | `/api/v1/admin/system` | 系统配置 |

## 基础设施配置

| 服务 | 用途 | 内部端口 | 外部端口 |
|------|------|---------|---------|
| nginx | HTTPS 反向代理 | 80/443 | 80/443 |
| backend | Go API 服务 | 8080 | — |
| frontend | Nuxt SSR | 3000 | — |
| mysql | 业务数据 | 3306 | 3307 |
| redis | 缓存 + Token | 6379 | 6379 |
| minio | 对象存储 | 9000/9001 | 9001 |
| rabbitmq | 转码队列 | 5672/15672 | 5672/15672 |

## 部署与运维

```bash
./deploy.sh                    # 一键部署（构建 + 启动 + 健康检查）
./backup.sh                    # 备份 MySQL + Redis RDB + MinIO
./restore-mysql.sh <file>     # 恢复 MySQL
docker compose logs -f backend # 查看后端日志
```

## License

MIT
