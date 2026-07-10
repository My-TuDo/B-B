# B-B

> B 站风格视频平台 — Go + Nuxt 全栈项目（仿 B 站）

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.23+ / Gin / GORM / MySQL 8.0 / Redis 7 / MinIO / RabbitMQ |
| 前端 | Nuxt 4 / Vue 3 / TypeScript / Tailwind CSS / Video.js / hls.js |
| 基础设施 | Docker Compose / Nginx / ffmpeg / Let's Encrypt(可选) |

## 快速启动（Docker 一键部署）

```bash
# 一键部署（SSL 证书自动生成 + 构建 + 启动）
./deploy.sh

# 访问
#    前端: https://localhost
#    MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
#    RabbitMQ: http://localhost:15672 (guest/guest)

# 单独启动（不运行部署脚本）
docker compose up -d
```

## 本地开发

```bash
# 启动依赖服务
docker compose up -d mysql redis minio rabbitmq

# 终端 1: 后端
cd backend && go run ./cmd/server/main.go

# 终端 2: 前端
cd frontend && pnpm install && pnpm dev
```

> 前端开发时通过 Nuxt 代理 `/api/*` → `localhost:8080`，MinIO 通过 Docker DNS `minio:9000` 访问

## 一键部署

```bash
./deploy.sh
```

脚本会：
1. 检查 Docker / Docker Compose 环境
2. 生成自签名 SSL 证书（首次运行）
3. `docker compose build` 构建镜像
4. `docker compose up -d` 启动全部服务
5. 等待后端健康检查通过

入口网关：
- **80** → **443** HTTPS 301 跳转
- **443** → 前端页面（Nuxt SSR） / API 反向代理

## 数据备份与恢复

```bash
# 备份（MySQL + Redis + MinIO）
./scripts/backup.sh

# 恢复 MySQL
./scripts/restore-mysql.sh ./backups/mysql/daily/bb_20260710_030000.sql.gz
```

备份数据保存在 `./backups/`，MySQL 保留最近 7 天。

## 项目结构

```
B-B/
├── backend/
│   ├── cmd/server/main.go          # 入口
│   ├── internal/
│   │   ├── handler/                # HTTP 层 (12 模块)
│   │   ├── service/                # 业务逻辑层
│   │   ├── repository/             # 数据访问层
│   │   ├── model/                  # Entity + DTO
│   │   ├── middleware/             # Recovery/Logger/CORS/RateLimit/Auth/CSRF
│   │   └── worker/                 # 转码 Worker (ffmpeg)
│   ├── pkg/                        # 公共包 (config/database/jwt/response/storage/rabbitmq)
│   └── migrations/                 # SQL 迁移
├── frontend/
│   ├── pages/                      # 文件路由 (12+ 页面)
│   ├── components/                 # 通用 + 业务组件
│   ├── composables/                # useApi/useToast/useTheme
│   ├── stores/                     # Pinia (userStore / playerStore / danmakuStore)
│   ├── layouts/                    # default / auth
│   └── middleware/                 # auth / guest
├── nginx/
│   ├── Dockerfile                  # nginx + SSL 证书
│   ├── nginx.conf                  # HTTPS 反向代理配置
│   └── ssl/generate.sh             # 自签名证书生成
├── scripts/
│   ├── backup.sh                   # 数据备份脚本
│   └── restore-mysql.sh            # MySQL 恢复脚本
├── deploy.sh                       # 一键部署脚本
└── docker-compose.yml              # 全部 8 个服务
```

## 功能清单

### 阶段一 — 核心骨架
- 用户注册 / 登录 / 登出（JWT + Redis 白名单 + HttpOnly Cookie）
- 视频上传（MinIO + SSE 实时进度 + MIME/魔数/大小三重校验）
- 视频播放（MinIO 预签名 URL）+ 分类浏览
- 稿件管理（草稿/已发布/已删除 Tab + 编辑/删除）
- 暗色 / 亮色双模切换（CSS 变量 + localStorage）
- 侧边栏（展开 240px / 折叠 72px）+ 7 个分类
- 限流（5 次/s）、文件 MIME+魔数校验（≤500MB）

### 阶段二 — 内容消费
- 首页推荐（热度算法 `views×0.5+likes×2+comments×3-hours×0.1`，Redis 10min 缓存）
- 排行榜（日/周/总，Redis ZSet）
- 全文搜索（MySQL FULLTEXT → LIKE 降级）+ 关键词提示
- 分区 Tab + 分页
- 标签系统（多对多）
- 观看历史 + 断点续播
- UP 主创作中心（视频管理 + 播放数据）
- Admin 审核队列（role≥2/3）+ CSRF Token

### 阶段三 — 社区互动
- 弹幕系统（WebSocket 实时 + REST 历史 + Bilibili 风格 RAF 时间驱动）
- 评论（楼中楼两层 + 排序 + Redis 点赞）
- 三连（点赞/投币限 5 枚/天/收藏多收藏夹）
- 关注/粉丝 + Feed 信息流
- 消息通知（评论回复/点赞/关注）
- 用户空间（视频/收藏/动态）
- 头像上传（JPEG/PNG/WebP ≤2MB，MinIO 存储+直接 URL）
- 分享链接

### 阶段四 — 媒体处理
- 自动转码（360p/480p/720p/1080p，RabbitMQ 异步队列）
- HLS 分片 + m3u8 索引
- ffprobe 元数据提取（duration/width/height/codec/bitrate）
- 自动封面截取（ffmpeg 第一帧，不覆盖用户上传）
- 清晰度切换（前端保持播放进度）
- 降级策略（RabbitMQ 不可用→goroutine 直调 / ffmpeg 不可用→原始 mp4 兜底）
- Admin Dashboard（统计卡片 + 用户管理 + 角色编辑 + 系统配置）

### 阶段五 — 部署运维
- Docker 多阶段构建（后端 < 100MB）
- Nginx 反向代理（HTTPS + HTTP→HTTPS 跳转 + 安全头 + Gzip）
- 自签名 SSL 证书（dev/demo，可选 Let's Encrypt）
- Admin 系统配置页（Go 版本 / 运行时间 / 数据库连接状态）
- gin-prometheus `/metrics` 端点暴露
- 数据备份脚本（MySQL dump + Redis RDB + MinIO mirror）
- 一键部署脚本 `deploy.sh`

## 基础设施

| 服务 | 内部端口 | 外部端口 | 说明 |
|------|---------|---------|------|
| nginx | 80/443 | 80/443 | 反向代理入口（HTTPS） |
| backend | 8080 | — | Go API |
| frontend | 3000 | — | Nuxt SSR |
| mysql | 3306 | 3307 | 数据库 |
| redis | 6379 | 6379 | 缓存 + Token 白名单 |
| minio | 9000/9001 | 9001 | 对象存储（API/Console） |
| rabbitmq | 5672/15672 | 5672/15672 | 转码消息队列 |

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
| POST | `/api/v1/videos` | 上传视频（+可选封面，SSE 进度） |
| GET | `/api/v1/videos` | 视频列表（分页+分类） |
| GET | `/api/v1/videos/:id` | 视频详情 |
| GET | `/api/v1/videos/:id/play-url` | 播放地址 |
| GET | `/api/v1/videos/:id/qualities` | 清晰度列表 |
| GET | `/api/v1/videos/:id/transcode-status` | 转码状态 |
| GET | `/api/v1/videos/hot` | 首页推荐 |
| GET | `/api/v1/videos/ranking` | 排行榜 |
| PUT | `/api/v1/videos/:id` | 更新视频 |
| DELETE | `/api/v1/videos/:id` | 删除视频 |
| GET | `/api/v1/videos/users/:id/videos` | 用户视频列表 |

### 标签 / 分类 / 搜索
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/categories` | 分类列表 |
| GET/POST | `/api/v1/tags` | 标签列表 / 创建 |
| POST | `/api/v1/videos/:id/tags` | 设置视频标签 |
| GET | `/api/v1/search` | 搜索 |
| GET | `/api/v1/search/suggestions` | 搜索建议 |

### 社区
| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/videos/:id/danmaku` | 弹幕列表/发送 |
| WS | `/api/v1/ws/danmaku/:video_id` | 弹幕 WebSocket |
| GET/POST | `/api/v1/videos/:id/comments` | 评论列表/创建 |
| POST | `/api/v1/comments/:id/like` | 评论点赞/取消 |
| POST | `/api/v1/videos/:id/like` | 视频点赞/取消 |
| POST | `/api/v1/videos/:id/coin` | 投币 |
| GET/POST | `/api/v1/favorites` | 收藏夹列表/创建 |
| POST | `/api/v1/favorites/:id/items` | 收藏/取消视频 |
| GET | `/api/v1/history` | 观看历史 |
| POST | `/api/v1/history` | 记录进度 |
| GET | `/api/v1/feed` | 关注 Feed |
| GET | `/api/v1/notifications` | 消息通知 |

### 创作中心
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/creator/videos` | 我的视频 |
| GET | `/api/v1/creator/stats` | 创作数据 |

### 管理员（role≥3）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/stats` | 统计面板 |
| GET | `/api/v1/admin/users` | 用户管理 |
| PUT | `/api/v1/admin/users/:id/role` | 修改角色 |
| GET | `/api/v1/admin/videos` | 视频审核 |
| PUT | `/api/v1/admin/videos/:id/review` | 审核操作 |
| GET | `/api/v1/admin/system` | 系统配置信息 |

### 安全
- `X-CSRF-Token` Header 保护（POST/PUT/DELETE）
- JWT HttpOnly Cookie + Redis 白名单
- 路由级限流（5 次/s）
- 文件上传三重校验（MIME → 魔数 → 大小）
- HTTPS + 安全头（HSTS/X-Content-Type-Options/X-Frame-Options）

## 开发阶段

| 阶段 | 说明 | 状态 |
|------|------|------|
| 阶段一 | 核心骨架（注册/登录/上传/播放） | 🔒 locked |
| 阶段二 | 内容消费（首页/排行榜/搜索/历史） | 🔒 locked |
| 阶段三 | 社区互动（弹幕/评论/三连/关注/通知） | 🔒 locked |
| 阶段四 | 媒体处理（转码/HLS/封面/Admin Dashboard） | 🔒 locked |
| 阶段五 | 部署运维（HTTPS/Nginx/备份/一键部署） | ✅ implementing |

## License

MIT
