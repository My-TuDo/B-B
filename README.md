# B-B 项目介绍与技术分析

> 一个基于 Go + Nuxt 的全栈仿 B 站视频分享平台，侧重技术架构分析与工程实践总结。

项目效果展示：[ 视频或截图链接待补充 ]

---

## 1. 项目概览

B-B 是一个可实际运行的视频分享平台，用户可完成注册登录、上传视频、在线播放、发送弹幕、评论互动、关注 Feed 等完整业务流程。项目并非纯静态仿站——其后端通过 Go 处理视频转码（ffmpeg）、对象存储（MinIO）、消息队列（RabbitMQ）、实时通信（WebSocket），前端基于 Nuxt 4 实现 SSR 页面渲染与暗色双模，形成了一条从前端展现到后端存储的**完整业务闭环**。

核心链路：用户上传 → MinIO 存储 → RabbitMQ 异步转码 → HLS 多分辨率播放；同时提供弹幕 WebSocket 实时推送、点赞/评论消息通知、关注 Feed 信息流等社区功能。

---

## 2. 项目定位与整体架构

### 2.1 项目定位

| 组件 | 在业务中的核心职责 |
|------|-------------------|
| **MySQL 8.0** | 业务数据的持久化，存储用户、视频、评论、弹幕、收藏夹等 14 张表 |
| **Redis 7** | 三层用途：(1) JWT Token 白名单支持强制下线；(2) 热度排序/排行榜 ZSet 缓存；(3) 登录接口 IP 级别限流计数 |
| **MinIO** | 视频原始文件、HLS 分片（m3u8 + ts）、封面截图、用户头像的对象存储。Bucket 设为公开下载策略，浏览器直连获取无需 CDN |
| **RabbitMQ** | 异步视频转码任务队列。上传视频后发布消息，Worker 消费后调用 ffmpeg。不可用时降级为 goroutine 直调 |
| **WebSocket** | 弹幕实时推送。Hub 单例模式按 videoID 分房间广播，客户端无需轮询 |

### 2.2 源码观察结论

**已明确实现的部分：**
- 19 个 HTTP Handler 模块，覆盖认证/视频/弹幕/评论/点赞/收藏/关注/消息/搜索/管理后台
- 17 个 Service 层模块，处理跨 Repository 编排、权限校验、缓存穿透防护
- 7 个中间件：Recovery、RequestID、Zap Logger、CORS、令牌桶限流、CSRF、JWT+Redis 认证
- ffmpeg 转码 Worker：四档分辨率（360p~1080p）、HLS 分片、ffprobe 元数据提取、自动封面截取
- 15 个前端页面，Nuxt 4 SSR + Tailwind CSS 暗色/亮色双模，B 站风格的 RAF 时间驱动弹幕引擎
- Docker 多阶段构建 + Nginx HTTPS 反向代理 + 一键部署脚本

**尚未补完的细节（在源码中可观察到）：**
- 配置文件中明文存储默认密码（`JWT_SECRET=dev-secret-change-in-production`），仅依赖 `.env` 覆盖
- 除登录/注册接口外，其他接口无 IP 级别的细粒度限流
- WebSocket Hub 是进程内单实例，Docker Compose 多副本部署时弹幕消息无法跨实例同步
- 前端部分预留页面（`drafts.vue`、`notifications.vue`）仅有基础框架，未完全接入后端
- 缺少自动化测试（无单元测试或集成测试文件）
- RabbitMQ Consumer 恢复机制较简单，长时间断连后需手动重启服务

### 2.3 整体架构说明

| 层面 | 主要职责 | 当前实现 |
|------|---------|---------|
| **接入层** | HTTPS 终断、反向代理、静态文件服务 | Nginx (alpine)，自签名证书，`/api/*` 代理到后端 |
| **前端层** | 页面渲染、用户交互、状态管理 | Nuxt 4 + Vue 3 + Pinia + Tailwind CSS，SSR 模式 |
| **后端层** | 业务逻辑、认证鉴权、数据校验 | Go + Gin，Handler → Service → Repository 三层 |
| **中间件层** | 请求前置处理 | 7 个中间件按序链式执行 |
| **数据层** | 持久化存储、缓存、文件存储 | MySQL/GORM + Redis + MinIO |
| **任务层** | 异步视频处理 | RabbitMQ 队列 + Worker goroutine 消费 |
| **通信层** | 实时消息推送 | WebSocket Hub 按房间广播 |

**典型请求路径（以「上传视频并播放」为例）：**

```
1. 前端 POST /api/v1/videos (multipart/form-data)
2. Nginx 反代到 Go 后端
3. 中间件链: Recovery → RequestID → Logger → CORS → RateLimit → CSRF → AuthRequired
4. Handler 校验文件 (MIME + 魔数 + 大小) → SSE 流式进度推送
5. Service 上传到 MinIO → 写入 videos 表 (status=0 草稿)
6. RabbitMQ 发布转码任务 (或 goroutine 降级直调)
7. Worker 消费: ffprobe 提取元数据 → ffmpeg 四档转码 → HLS 上传 → 自动封面截图
8. 更新 videos 表 (status=1 已发布) + 写入 video_qualities 表
9. 用户请求 GET /api/v1/videos/:id/play-url → 返回 HLS m3u8 地址
10. 前端 hls.js 加载 m3u8 播放
```

---

## 3. 已实现功能总览

### 核心业务模块

| 模块 | 核心功能 | 已接入页面 | 备注 |
|------|---------|-----------|------|
| 认证 | 注册/登录/登出/Refresh/GetMe | login, register | JWT HttpOnly Cookie + Redis 白名单 |
| 视频 | 上传(SSE进度)/播放(HLS+mp4)/编辑/删除/清晰度切换 | index, video/[id], upload | 支持 360p~1080p |
| 弹幕 | 发送/历史/WebSocket 实时推送 | video/[id] | RAF 时间驱动，跟随视频进度 |
| 评论 | 楼中楼/排序/点赞 | video/[id] | 两级嵌套，Redis 记录点赞 |
| 三连 | 点赞(切换式)/投币(5枚/天)/收藏(多夹) | video/[id] | 互动计数实时更新 |
| 关注 | 关注/取关/Feed 信息流 | feed, user/[id] | Feed 拉取关注者视频 |
| 消息 | 评论回复/点赞/关注通知 | notifications | 预留页面，后端 API 已就绪 |
| 搜索 | FULLTEXT 搜索+LIKE 降级 | search | 视频标题/简介全文检索 |
| 历史 | 断点续播/时间轴展示 | history | 时间轴分组+进度条 |
| 排行榜 | 日/周/总榜 | ranking | Redis ZSet 按 views 排序 |
| 管理后台 | 统计/用户管理/角色编辑/系统配置 | admin | role≥3 可访问 |
| 创作中心 | 统计/稿件管理(编辑/删除) | creator | 已发布稿件可删除编辑 |
| 个人中心 | 双栏布局/头像悬停换头像/收藏夹封面 | user/[id] | 阶段六完整重写 |

### 前端页面清单（15 个）

| 已完整实现 | 预留框架 |
|-----------|---------|
| index(首页), video/[id](详情), login, register, upload, user/[id], history, ranking, search, feed, admin, creator | drafts, notifications |

---

## 4. 各核心功能的技术实现方式

### 4.1 视频上传与转码

#### 功能作用
用户通过网页上传 MP4 视频，系统异步转为 4 档 HLS 清晰度（360p/480p/720p/1080p），支持在线播放。

#### 前端表现
上传页拖拽选择文件 → 进度条实时显示（SSE 推送） → 上传成功后自动跳转视频页。

#### 后端实现
1. Handler 接收 `multipart/form-data`，读取文件头 512 字节通过 `http.DetectContentType` 验证 MIME 真伪
2. 自定义 `progressReader` 包装 `io.Reader`，每次 `Read` 回调进度百分比
3. Gin `c.Stream()` 以 SSE 格式推送进度 (`data: {"progress":65}\n\n`)
4. MinIO `PutObject` 上传原始视频 → 写入 `videos` 表 (status=0 草稿)
5. RabbitMQ `PublishTranscodeTask(videoID)` → Worker 消费：
   - ffprobe 提取 duration/width/height/codec/bitrate → 写入 `video_metas` 表
   - ffmpeg 四档 scale → HLS (m3u8 + ts 分片) → 上传到 MinIO
   - 写入 `video_qualities` 表（quality + play_url + file_size）
   - 若 `cover_url` 为空，ffmpeg 截取第一帧 → 上传为封面
6. 更新 `videos.status=1`（已发布）

#### 使用的技术
`multipart/form-data`、`io.Reader` 包装、SSE (Server-Sent Events)、MinIO SDK、RabbitMQ (amqp091-go)、ffmpeg/ffprobe 命令行调用，goroutine 降级策略。

#### 当前实现的局限
- 转码队列无优先级和死信队列，失败后仅标记 status=3
- Worker 单点运行，无横向扩展或任务分片
- 不支持非 MP4 格式（如 MOV、WebM）
- HLS 分片固定 10 秒，不支持自适应比特率

---

### 4.2 JWT + Redis 白名单认证

#### 功能作用
用户登录后获取 JWT Token，后续请求自动携带 Cookie 认证。支持服务端强制下线。

#### 后端实现
1. 登录成功 → `jwt.GenerateToken()` 签发 HS256 Token（7 天有效期）
2. `rdb.Set("auth:token:{userId}", token, 7*24*time.Hour)` 写入 Redis 白名单
3. `c.SetCookie("token", token, ..., HttpOnly: true)` 写入浏览器
4. 后续请求 → AuthRequired 中间件：
   - 从 Cookie 或 `Authorization: Bearer` 提取 Token
   - `jwt.ParseToken()` 验证签名+过期
   - `rdb.Get("auth:token:{userId}")` 比对白名单（支持踢人）
   - `c.Set("userId", ...)` 注入 Gin Context

#### 当前实现的局限
- 没有 Refresh Token 的平滑续期机制
- Token 过期后用户需重新登录
- 无多设备登录管理

---

### 4.3 弹幕系统（Bilibili 风格）

#### 功能作用
用户在视频任意时间点发送弹幕，弹幕跟随视频进度滚动显示，支持回退/快进时弹幕同步。

#### 前端表现
弹幕从视频右边缘外侧飞入，RAF 循环 60fps 驱动位置更新。暂停时冻结，回退时回退。

#### 后端实现
1. REST API：`POST /api/v1/videos/:id/danmaku` 写入 `danmakus` 表
2. WebSocket：`ws://host/api/v1/ws/danmaku/:video_id` 连接 Hub
3. Hub 单例按 `videoID` 分房间，新弹幕通过 goroutine 广播到房间内所有连接
4. 前端 `requestAnimationFrame` 循环计算位置：`translateX = cw - (cw+elWidth) * (currentTime - playTime) / 7s`

#### 使用的技术
WebSocket (gorilla/websocket)、Hub-Room 广播模式、requestAnimationFrame、视频 `currentTime` 属性驱动。

#### 当前实现的局限
- WebSocket Hub 是进程内单实例，多 Docker 副本弹幕不同步
- 弹幕无敏感词过滤机制

---

### 4.4 热度算法与 Redis 缓存

#### 功能作用
首页"推荐"标签展示热度排序的视频，排行榜按日/周/总展示高播放量视频。

#### 后端实现
1. 热度公式：`score = views×0.5 + likes×2 + comments×3 - (hours since publish)×0.1`
2. 查询 `videos` 表计算 score → 写入 Redis ZSet `videos:hot`，TTL 10 分钟
3. 首页请求优先从 Redis ZSet `ZREVRANGE` 获取，未命中则回源 MySQL → 写入缓存
4. 排行榜使用独立 ZSet：`videos:rank:day/week/total`，按 views 排序

#### 当前实现的局限
- 缓存更新时机依赖 TTL 过期，非实时
- 大量视频时 ZSet 排序在 CPU 密集场景可能有性能瓶颈

---

## 5. 前端与后端技术栈说明

### 前端技术栈

| 技术 | 当前用途 |
|------|---------|
| Nuxt 4 | SSR 服务端渲染框架，文件路由自动注册 |
| Vue 3 + TypeScript | 组件化开发，Composition API |
| Tailwind CSS | 原子化 CSS，CSS 变量暗色/亮色双模 |
| Pinia | 状态管理（userStore, playerStore, danmakuStore） |
| Video.js + hls.js | 视频播放器 + HLS 流加载 |
| WebSocket | 弹幕实时通信 |
| Fredoka + Inter 字体 | Google Fonts 排版系统 |

**代码风格评价**：使用 Nuxt Composition API + `<script setup>` 语法，组件细粒度拆分合理。CSS 变量体系统一管理暗色/亮色，避免硬编码色值。P0 页面交互细节投入较多（Header 毛玻璃、VideoCard 悬停缩放、侧边栏缓动过渡），P2 页面（drafts、notifications）仅维持基础框架。

### 后端技术栈

| 技术 | 当前用途 |
|------|---------|
| Go 1.25 + Gin | HTTP 框架、中间件链、参数绑定 |
| GORM | ORM，AutoMigrate 自动建表 |
| MySQL 8.0 | 业务数据持久化 |
| Redis 7 | Token 白名单、热度缓存、排行榜、限流计数 |
| MinIO | 视频/图片对象存储 |
| RabbitMQ | 转码任务队列 |
| gorilla/websocket | 弹幕 WebSocket |
| zap | 结构化 JSON 日志 |
| viper | YAML + 环境变量配置 |
| golang.org/x/time/rate | 令牌桶限流 |

**代码风格评价**：严格的三层架构（Handler → Service → Repository），每一层通过 `fmt.Errorf("包.方法: %w", err)` 保留错误链。Service 返回自定义 `*Error` 类型，Handler 通过 `errors.As` 区分业务错误和内部错误。中间件链完全自实现（未使用 Gin 默认 Logger/Recovery），对 RequestID 注入、zap 集成、敏感字段脱敏做了工程化处理。全部 Go 文件已加中文注释。

---

## 6. 项目目录与分层职责

```
B-B/
├── backend/
│   ├── cmd/server/main.go               # 启动入口：初始化→建表→注册路由
│   ├── internal/
│   │   ├── handler/     (19 modules)    # HTTP 层
│   │   ├── service/     (17 modules)    # 业务逻辑层
│   │   ├── repository/  (17 modules)    # 数据访问层
│   │   ├── model/       (14 modules)    # GORM Entity + DTO
│   │   ├── middleware/  (7 files)       # 中间件链
│   │   ├── worker/      (2 files)       # ffmpeg 转码 Worker
│   │   └── ws/          (1 file)        # WebSocket Hub
│   ├── pkg/                             # 公共基础设施
│   └── migrations/                      # SQL 迁移脚本
└── frontend/
    ├── pages/            (15 pages)     # 文件路由
    ├── components/        (20+ comps)   # 通用+业务组件
    ├── composables/                     # useApi/useToast/useTheme
    ├── stores/                          # Pinia stores
    ├── layouts/                         # default/auth
    └── middleware/                      # auth/guest
```

| 层 | 职责 | 不做什么 |
|----|------|---------|
| Handler | 解析 HTTP 参数，调用 Service，返回统一 JSON | 不写 SQL，不判断业务规则 |
| Service | 业务规则校验，跨 Repository 编排，缓存/通知 | 不解析 HTTP，不写裸 SQL |
| Repository | 单表 GORM 查询，返回实体或 DTO | 不做业务判断，不调其他 Repository |
| Model | 定义表结构（gorm tag）+ 请求/响应 DTO（json tag） | 不含任何逻辑 |

---

## 7. 当前项目的亮点

**业务闭环完整**——从上传到转码到 HLS 播放，从弹幕发送到 WebSocket 实时推送，从点赞评论到消息通知，核心链路均已跑通，不是静态演示项目。

**技术选型贴合视频场景**——Go 的原生并发模型天然适合 I/O 密集的视频服务；MinIO 的 S3 兼容性 + 公开 bucket 策略节省了 CDN 成本；RabbitMQ 的异步解耦避免了上传接口阻塞在转码上。

**工程化程度较高**——中间件链全部自实现且有序配合（RequestID 贯穿 Recovery/Logger）；统一响应格式 `{code, message, data, request_id}` 前端一套解析逻辑全覆盖；错误包装链 `fmt.Errorf("包.方法: %w", err)` 明确错误来源。

**降级策略有心设计**——RabbitMQ 不可用时自动降级为 goroutine 直调转码；ffmpeg 不可用时保留原始 MP4 播放；Redis 不可用时限流自动放行。

**代码可读性投入较大**——全部 Go 文件中文注释、模块化拆分合理、Handler 方法不超过 40 行、Repository 方法通常仅一条 GORM 链式调用。

---

## 8. 当前观察到的不足与改进建议

| 类别 | 问题 | 建议 |
|------|------|------|
| **安全性** | `config.go` 中明文默认密码 (`dev-secret-change-in-production`, `minioadmin`, `bb_password`) | 迁移到 Vault 或至少通过 `os.Getenv` 强制从环境变量读取，无默认值 |
| **安全性** | 弹幕/评论无敏感词过滤 | 接入 DFA 敏感词匹配库 |
| **可靠性** | RabbitMQ Consumer 断连后无自动重连 | 增加 channel/connection 关闭监听 + 指数退避重试 |
| **可靠性** | 转码 Worker 单点运行 | 引入分布式任务调度或至少多 Worker 竞争消费 |
| **扩展性** | WebSocket Hub 进程内单实例 | 改用 Redis Pub/Sub 做跨实例消息同步，支持多副本 |
| **扩展性** | 热度排序每次计算所有视频的 score | 超过 10 万视频时应改为定时任务预计算 + 增量更新 |
| **数据保护** | 视频删除为软删除 (status=3)，但 MinIO 文件未删除 | 增加 TTL 清理策略或标记文件待删除 |
| **测试** | 无单元测试或集成测试 | 至少为 Service 层增加 mock Repository 的 Table-Driven 测试 |
| **监控** | 仅有 gin-prometheus `/metrics` 端点 | 增加转码队列积压告警、慢查询日志、错误率面板 |
| **部署** | `deploy.sh` 生成的自签名证书浏览器不信任 | 可选集成 certbot 或提供 acme.sh 配置说明 |

---

## 9. 适合展示或答辩时的总结性描述

> B-B 是一个基于 Go + Nuxt 全栈实现的仿 B 站视频平台。项目从前端 SSR 页面渲染到后端三层架构、从 JWT+Redis 双重认证到 CSRF 防护、从 ffmpeg 四档 HLS 转码到 RabbitMQ 异步任务队列，跑通了视频平台的完整业务闭环。代码严格分层——Handler 管 HTTP、Service 管逻辑、Repository 管数据——每一层通过统一错误包装链可追踪。弹幕系统用 RAF 时间驱动替代 CSS 动画，支持视频回退时弹幕同步跟随。工程化上自实现 7 个中间件、统一 JSON 响应格式、敏感日志字段脱敏，RabbitMQ 不可用时 goroutine 降级保障核心功能可用。源码全部中文注释，适合作为 Go 全栈项目的学习参考。

---

## 10. 结语

B-B 是一个以“可学习的全栈实践”为目标的视频平台项目。它没有选择用 ORM 自动生成代码，也没有引入过度抽象——每一行 Handler、Service、Repository 的职责都可以在对应文件中一眼看明白。如果你是 Go 后端开发者，可以从 `cmd/server/main.go` 的启动流程开始读；如果你是前端开发者，可以从 `pages/index.vue` 的 VideoCard 组件开始调试。希望这份代码和技术分析对你有帮助。

---

*最后更新：2026-07-13 · 项目状态：阶段六（UI 优化）已完成*
