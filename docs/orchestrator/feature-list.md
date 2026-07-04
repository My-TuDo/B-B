# 产品功能清单

> **定位**：仿 B 站视频平台的全栈功能清单。按 5 阶段构建，每阶段产出独立可运行的版本。

---

## 架构决策摘要

| 决策 | 结论 |
|------|------|
| 仓库 | Monorepo，GitHub `My-TuDo/B-B` |
| 分支 | `backend` + `frontend` 双分支 |
| Go Module | `github.com/My-TuDo/B-B/backend` |
| 后端 | Go + Gin + GORM + MySQL 8.0 + Redis 7 + MinIO + RabbitMQ(阶段4) |
| 前端 | Vue 3 + Nuxt 3 + Video.js + Tailwind CSS |
| 通信 | Nuxt 直调 Go RESTful API |
| API 契约 | 前期手写，后期引入 swaggo + openapi-typescript |
| 开发环境 | Docker Compose 一键启动 |
| UI 基调 | 暗色+亮色双模、圆润柔和、中等密度、B站温度+YouTube布局 |
| 认证 | HttpOnly Cookie + Redis Token 白名单、CSRF Token(阶段2) |
| 项目结构 | 联动聚类：`platform/` `identity/` `content/` `social/` `media/` |
| Agent 模型 | 三体 Harness v2：主Agent → 项目Agent → QA Agent → 人测试锁版 |

---

## 阶段 1：核心骨架 + 安全基础

**复杂度** ★★★★
**用户价值**：能注册、能登录、能上传视频、能播放、能存草稿

### 后端

| 域 | 功能点 | 说明 |
|----|--------|------|
| 平台 | Docker Compose | MySQL 8.0 + Redis 7 + MinIO + Go(air热重载) + Nuxt(dev) |
| | 配置管理 | Viper 环境变量注入，Server/DB/Redis/MinIO/JWT 配置结构体 |
| | 数据库连接 | GORM 连接池(MaxOpenConns=100, MaxIdleConns=10)、AutoMigrate |
| | Redis连接 | go-redis/v9 初始化、连接池 |
| | 日志系统 | Zap + requestId(UUID)、请求日志中间件 |
| | 统一响应 | `{ code, message, data }`、Success()/Error()封装、统一错误码 |
| 认证 | 注册 | POST `/api/v1/auth/register`，username 3-20位，password 6-30位，bcrypt哈希 |
| | 登录 | POST `/api/v1/auth/login`，验证密码，返回 JWT → Set-Cookie(HttpOnly; Secure; SameSite=Strict) |
| | 刷新令牌 | POST `/api/v1/auth/refresh`，验证 refresh_token |
| | Token白名单 | Redis `user:{userId}:tokens` Set 存储有效Token，删除=强制下线 |
| | 当前用户 | GET `/api/v1/auth/me` |
| 用户 | 用户主页 | GET `/api/v1/users/:id`，公开信息 |
| | 编辑资料 | PUT `/api/v1/users/:id`，头像上传、昵称、签名 |
| 视频 | 上传 | POST `/api/v1/videos`，multipart/form-data，MinIO存储，SSE进度推送 |
| | 详情 | GET `/api/v1/videos/:id`，视频元信息+上传者信息 |
| | 播放地址 | GET `/api/v1/videos/:id/play-url`，MinIO预签名URL(有效期1h) |
| | 草稿 | videos.status=0(draft)，草稿列表、编辑后发布(status=1) |
| 分类 | categories表 | id/name/parent_id/sort_order，种子数据(一级分类) |
| 安全 | 限流中间件 | IP令牌桶，5次/秒，返回429 |
| | 文件校验 | MIME类型白名单(mp4/avi/mkv)、文件大小≤500MB、魔数检查 |
| | 参数校验 | Gin binding校验，username/email格式 |
| 中间件 | 执行链 | Recovery → Logger → CORS → RateLimit → Auth(路由级) |

### 前端

| 域 | 功能点 | 说明 |
|----|--------|------|
| 初始化 | Nuxt 3 | TypeScript严格模式、pnpm、ESLint+Prettier |
| 目录结构 | assets/components/composables/layouts/middleware/pages/types | |
| 全局配置 | CSS变量 | `--color-primary/bg/surface`、`--radius-md`、`--spacing-*`、暗色主题 |
| | API客户端 | `useApi.ts`封装`$fetch`，baseURL+错误处理，自动带CSRF Token |
| | 代理 | Nuxt DevServer代理`/api/*` → Go `:8080` |
| 布局 | default.vue | AppHeader(Logo+导航+用户头像下拉) + 主内容区 + AppFooter |
| | auth.vue | 居中卡片，无导航，背景与default一致 |
| 通用组件 | AppHeader | Logo、首页/上传入口、登录态(头像下拉→个人空间/退出) |
| | LoadingSpinner | CSS动画，size prop |
| | EmptyState | 图标+提示+可选操作按钮 |
| | ErrorMessage | 错误图标+信息+重试按钮 |
| | AppToast | 右上角弹出，success/error/info，3s自动消失 |
| 路由 | `/` | 首页(default，无需认证)，阶段1占位 |
| | `/login` | 登录页(auth，guest中间件) |
| | `/register` | 注册页(auth，guest中间件) |
| | `/upload` | 上传页(default，auth中间件)，拖拽+标题+分类选择+进度条 |
| | `/video/[id]` | 视频详情+播放页(default，无需认证) |
| | `/user/[id]` | 用户主页(default，无需认证) |
| | `/drafts` | 草稿箱(default，auth中间件) |
| 路由守卫 | auth.ts | 无token → `navigateTo('/login')` |
| | guest.ts | 已登录 → `navigateTo('/')` |
| Store | userStore | token、userInfo、isLoggedIn、login()、logout()、fetchUserInfo() |

### 数据表

| 表 | 核心字段 | 索引 |
|----|---------|------|
| users | id, username(UNIQUE), email(UNIQUE), password_hash, avatar, sign, role, status, created_at | PK |
| videos | id, user_id, title, description, cover_url, file_key, file_size, duration, status(0=draft/1=published), view_count, created_at | PK, INDEX(user_id, created_at) |
| categories | id, name, parent_id, sort_order | PK |

### 交付物

```
docker compose up → 打开浏览器 http://localhost:3000 →
注册 → 登录 → 上传视频(拖拽+看进度) → 存草稿 → 编辑草稿 → 发布 →
在个人主页看到视频 → 点进去 → 播放 → 切换清晰度
```

---

## 阶段 2：内容消费 + 运营工具

**复杂度** ★★★★
**用户价值**：首页刷推荐 → 分区浏览 → 排行榜 → 搜索找视频 → UP主管理作品 → 断点续播

### 后端

| 域 | 功能点 | 说明 |
|----|--------|------|
| 首页推荐 | 热度算法 | `score = view_count*0.5 + like_count*2 + comment_count*3 - hours_elapsed*0.1`，分页返回 |
| 分区浏览 | 分类筛选 | GET `/api/v1/videos?category=1&sort=hot\|new&page=1&size=20` |
| 排行榜 | 日榜/周榜/总榜 | Redis ZSet，score=播放量+互动加权，每日凌晨更新，支持按分区 |
| 搜索 | 全文检索 | MySQL FULLTEXT INDEX(title, description)，LIKE降级 |
| | 关键词提示 | GET `/api/v1/search/suggest?prefix=xxx`，Redis ZREVRANGE |
| | 热度词 | Redis ZSet，每日凌晨统计搜索频次 |
| 标签 | tags表 + 关联 | 上传时选/建标签，tag筛选页 |
| 观看历史 | 记录进度 | POST `/api/v1/videos/:id/history`，前端每10s上报currentTime |
| | 历史列表 | GET `/api/v1/users/:id/history`，按updated_at倒序，显示进度百分比 |
| | 断点续播 | 打开视频→查询上次进度→video.currentTime跳转 |
| UP主管理 | 视频管理 | GET `/api/v1/users/:id/videos?status=draft\|published`，编辑(DELETE+PUT) |
| | 播放数据 | GET `/api/v1/videos/:id/stats`，view/like/coin/favorite/danmaku趋势(按日) |
| Admin | 视频审核 | GET `/api/v1/admin/videos?status=pending`，PUT通过/驳回+审核意见 |
| | 管理员角色 | users.role: 1=user, 2=moderator, 3=admin |
| 安全 | XSS防护 | 用户输入(content/description/sign)净化，HTML实体转义 |
| | CSRF Token | Cookie `csrf_token` + Header `X-CSRF-Token` 比对 |

### 前端

| 页/组件 | 功能点 | 说明 |
|---------|--------|------|
| 首页 | 视频卡片网格 | VideoCard(封面/标题/UP主/播放量/时长)，悬停预览(3s片段) |
| | 无限滚动 | IntersectionObserver，触底加载下一页 |
| 分区页 | 分类Tab | 侧边栏或顶部Tab，点击切换分类，视频列表 |
| 排行榜 | 日/周/总Tab | 按分区筛选，排名数字动画 |
| 搜索页 | 搜索框 | 导航栏入口或独立页，输入时实时提示，回车搜索 |
| | 结果列表 | 排序(时间/播放量/弹幕数)，分页 |
| 视频详情 | 右侧推荐 | 相关视频列表(同分类/同标签)，小缩略图+文字 |
| | 简介折叠区 | 点击展开完整description |
| 历史页 | 历史列表 | 卡片网格+进度条百分比，"继续观看"按钮→跳转至上次进度 |
| 视频管理 | 投稿列表 | 表格视图，状态标签，编辑/删除按钮 |
| | 编辑表单 | 修改标题/description/标签/分类 |
| 审核页 | 审核队列 | 管理员可见，视频列表+通过/驳回按钮+意见输入 |
| Store | playerStore | currentVideo、currentTime、duration、isPlaying、currentQuality、volume |

### 数据表

| 表 | 核心字段 | 索引 |
|----|---------|------|
| tags | id, name, count | PK |
| video_tags | id, video_id(FK), tag_id(FK) | UNIQUE(video_id, tag_id) |
| video_history | id, user_id(FK), video_id(FK), progress, created_at, updated_at | INDEX(user_id, updated_at) |

### 缓存策略

| 数据 | TTL | 方式 |
|------|-----|------|
| 首页推荐 | 10min | Redis List/ZSet |
| 排行榜 | 10min | Redis ZSet |
| 搜索热词 | 每日凌晨更新 | Redis ZSet + 定时任务 |

### 交付物

```
首页 → 无限下拉刷视频 → 点分类Tab切换到游戏区 →
看排行榜本周第一 → 搜索"原神" → 找到视频 → 点进去播放 →
关掉 → 历史页看到50%进度 → 续播 → UP主后台看播放数据 →
Admin审核新上传的视频 → 通过
```

---

## 阶段 3：社区互动 + 通知系统

**复杂度** ★★★★★
**用户价值**：发弹幕 → 三连 → 楼中楼评论 → 关注UP主 → 动态时间线 → 消息通知 → 分享

### 后端

| 域 | 功能点 | 说明 |
|----|--------|------|
| 弹幕 | WebSocket | go-socket.io Hub模式，按video_id分room，连接/断开/广播 |
| | 弹幕池 | Redis ZSet(send_time)，窗口500条，播放时按时间范围拉取 |
| | 历史弹幕 | GET `/api/v1/videos/:id/danmaku?start=0&end=300` |
| | 密度控制 | 同屏最多50条，超限按优先级(会员>普通)丢弃 |
| 评论 | 楼中楼 | parent_id: 0=一级评论 >0=回复，查一级→展开子回复 |
| | 评论点赞 | Redis Set去重 + 定时同步MySQL |
| 三连 | 点赞 | Redis Set(user_id+video_id去重)，切换状态 |
| | 投币 | 每日5枚上限(users.coins字段)，可选1或2枚 |
| | 收藏 | 创建/编辑收藏夹(favorites表)，添加到收藏夹(favorite_items) |
| | 计数同步 | Redis INCR + 每5min批量同步MySQL |
| 关注 | 关注/取关 | POST/DELETE `/api/v1/users/:id/follow`，Redis Set去重 |
| | 列表 | 粉丝列表/关注列表，分页 |
| Feed | 动态推送 | 关注UP主新视频，时间倒序，分页 |
| 通知 | 消息类型 | 评论回复/点赞/关注/视频审核通过 |
| | 未读计数 | Redis INCR，按userId统计 |
| | 消息列表 | GET `/api/v1/messages?page=1&type=all`，已读/未读标记 |
| 分享 | 分享链接 | 生成 `{base}/video/{id}` 分享链接 |
| | OG Meta | 服务端渲染title/description/cover → 社交平台预览卡片 |

### 前端

| 页/组件 | 功能点 | 说明 |
|---------|--------|------|
| DanmakuLayer | 弹幕渲染 | requestAnimationFrame驱动，CSS transform平移动画，右→左，离屏移除DOM |
| | 发送框 | 悬浮在播放器底部，选颜色(预设6色)/位置(滚动/顶部/底部)/字号 |
| | 开关 | 播放器控制栏弹幕开关按钮 |
| CommentList | 评论列表 | 一级评论+折叠子回复，@用户名高亮 |
| | 评论框 | 支持@用户，回车发送 |
| 互动按钮 | 点赞/投币/收藏 | 播放器下方横排，点赞动画(粒子效果)，投币选择弹窗，收藏夹选择弹窗 |
| 用户空间 | Tab切换 | 投稿/收藏夹/动态，关注按钮(已关注/未关注切换) |
| 关注页 | Feed时间线 | 关注UP主新视频列表，按时间倒序 |
| 通知页 | 消息列表 | 通知类型图标+文字，已读灰色/未读高亮，筛选类型 |
| 分享按钮 | 复制链接 | 弹出分享面板，复制链接+社交平台按钮 |
| Store | danmakuStore | danmakuList、isConnected、currentColor、currentType、sendDanmaku() |

### 数据表

| 表 | 核心字段 | 索引 |
|----|---------|------|
| danmaku | id, video_id(FK), user_id(FK), content, color, type(1=滚动/2=顶部/3=底部), send_time | INDEX(video_id, send_time) |
| comments | id, video_id(FK), user_id(FK), parent_id(0=一级), content, like_count | INDEX(video_id, parent_id, created_at) |
| video_likes | id, user_id(FK), video_id(FK), created_at | UNIQUE(user_id, video_id) |
| video_coins | id, user_id(FK), video_id(FK), count(1或2) | UNIQUE(user_id, video_id) |
| video_favorites | id, user_id(FK), video_id(FK), favorite_id(FK), created_at | UNIQUE(user_id, video_id, favorite_id) |
| favorites | id, user_id(FK), name, is_default | PK |
| follows | id, follower_id(FK), followee_id(FK), created_at | UNIQUE(follower_id, followee_id), INDEX(followee_id) |
| messages | id, to_user_id(FK), type, content, related_id, is_read, created_at | INDEX(to_user_id, is_read, created_at) |

### 缓存策略

| 数据 | 方式 |
|------|------|
| 点赞/播放/投币/收藏计数 | Redis INCR + 5min批量同步MySQL |
| 弹幕池 | Redis ZSet(send_time)，窗口500条，新弹幕入池，旧弹幕过期删除 |
| 未读消息数 | Redis INCR/DECR |

### 交付物

```
打开视频 → 弹幕飘过 → 发一条红色滚动弹幕 → 点赞 → 投2枚硬币 →
收藏到"学习资料"收藏夹 → 看评论 → 回复@某人 → 
关注UP主 → 关注页出现他的新视频 → 收到评论回复通知 →
分享视频链接到微信 → 预览卡片显示封面+标题
```

---

## 阶段 4：媒体处理

**复杂度** ★★★★★
**用户价值**：上传后自动转码 → 多分辨率切换 → HLS不卡顿 → 自动封面

### 后端

| 域 | 功能点 | 说明 |
|----|--------|------|
| 转码引擎 | ffmpeg封装 | Go exec.Command，分辨率配置: 360P/480P/720P/1080P |
| 异步队列 | RabbitMQ | 视频上传→发转码消息→消费者执行ffmpeg→回调更新状态 |
| 状态管理 | transcode_status | pending→processing→completed/failed，前端轮询或回调通知 |
| 多分辨率 | video_qualities | 每分辨率一条记录(url/size/duration)，播放时切换 |
| HLS | 分片(.ts) + m3u8 | ffmpeg输出HLS格式，边下边播，任意拖动进度 |
| 自动封面 | ffmpeg截图 | `ffmpeg -i input -ss 00:00:05 -vframes 1 output.jpg`，自动设为video.cover_url |
| 视频元数据 | ffprobe | 提取duration/width/height/bitrate/fps/codec → video_metas |
| Admin | 数据Dashboard | 注册用户数/日活/视频上传量/转码成功率/播放总量趋势图(ECharts) |
| | 用户管理 | 用户列表、搜索、状态管理(封禁/解封) |
| | 系统配置 | 转码开关、上传大小限制、推荐算法权重、分类管理 |

### 前端

| 页/组件 | 功能点 | 说明 |
|---------|--------|------|
| 播放器 | 清晰度切换 | 1080P/720P/480P/360P 下拉菜单，切换时保持当前进度 |
| | HLS自适应 | 网速变化自动切换清晰度 |
| 上传页 | 转码状态 | 上传完成后显示"转码中...预计X分钟"→"转码完成" |
| | 自动封面 | 上传完成后展示ffmpeg截取的封面图，可换帧或自定义上传 |
| Admin | 统计面板 | ECharts折线图(日活/上传量)、饼图(分辨率占比)、数字卡片(总用户/总视频) |
| | 用户管理 | 用户列表表格、搜索、封禁/解封操作 |
| | 系统配置 | 表单页：转码开关、上传大小滑块、推荐算法权重调节 |

### 数据表

| 表 | 核心字段 | 索引 |
|----|---------|------|
| video_qualities | id, video_id(FK), quality(360P/480P/720P/1080P), url, size, duration | INDEX(video_id, quality) |
| video_metas | id, video_id(FK), duration, width, height, bitrate, fps, codec | UNIQUE(video_id) |
| transcode_tasks | id, video_id(FK), quality, status(pending/processing/completed/failed), progress, error_msg | INDEX(video_id) |

### 交付物

```
上传4K视频 → 上传成功 → "转码中..." → 5分钟后转码完成 →
播放页清晰度菜单出现720P/1080P/4K → 切换到1080P →
任意拖动进度条 → 秒加载 → Admin面板显示今日转码成功率95% →
封面图自动生成(视频第5秒画面)
```

---

## 阶段 5：部署运维

**复杂度** ★★★
**用户价值**：HTTPS域名访问 → 一键部署 → 服务监控 → 数据不丢

### 后端

| 域 | 功能点 | 说明 |
|----|--------|------|
| Docker | 多阶段构建 | Go: `golang:1.22-alpine build` → `alpine:3.19 run`(二进制<20MB) |
| | | 前端: `node:20-alpine build` → Nginx静态托管 |
| Nginx | 反向代理 | `/` → Nuxt静态文件，`/api/` → Go `:8080`，`/uploads/` → MinIO |
| | HTTPS | Let's Encrypt证书，Certbot自动续期 |
| | 优化 | Gzip压缩、静态资源强缓存(`Cache-Control: max-age=31536000`) |
| 监控 | Prometheus | gin-prometheus中间件自动采集HTTP指标 |
| | Grafana | QPS/延迟P50 P99/错误率/Goroutine数/GC频率 仪表盘 |
| | 业务指标 | 自定义：注册数、上传数、转码成功率、弹幕数/分钟 |
| 告警 | 规则 | 错误率>1%(持续2min)、服务不可达、CPU>80%、内存>85%、磁盘>90% |
| | 通知 | 钉钉/企业微信Webhook |
| 备份 | MySQL | 每日凌晨3:00 mysqldump全量，保留7天，binlog实时增量 |
| | MinIO | `mc mirror` 备份到备用MinIO实例 |
| | Redis | RDB(30min快照) + AOF(每秒fsync) |
| Admin | 系统配置 | 转码开关、上传大小限制、推荐算法参数在线修改(无需重启) |

### 运维

| 项 | 内容 |
|----|------|
| docker-compose.yml | 全服务编排: mysql/redis/minio/rabbitmq/backend/frontend/nginx/prometheus/grafana |
| docker-compose.prod.yml | 生产配置：资源限制(cpu/memory)、重启策略(always)、日志驱动(json-file+rotate) |
| .env.production | 生产环境变量(JWT_SECRET DB_PASSWORD MINIO_KEY 等，不进git) |
| Makefile | `make up-dev` / `make up-prod` / `make backup` / `make restore` / `make logs` |
| 健康检查 | 所有容器 `healthcheck`，depends_on 条件为 `service_healthy` |
| 日志 | Docker json-file → logrotate(10MB×3文件)，Zap JSON格式输出 |

### 交付物

```
git push → 服务器 docker compose up -d →
https://你的域名.com 访问 → Let's Encrypt自动HTTPS →
Grafana仪表盘显示QPS/延迟/在线用户数 →
磁盘90% → 钉钉收到告警 → MySQL自动备份到 /backup/ →
服务器重启 → 所有服务自动恢复
```

---

## 全阶段汇总

### 数据表总计（18 张）

| 阶段 | 表 |
|------|----|
| 1 | users、videos、categories |
| 2 | tags、video_tags、video_history |
| 3 | danmaku、comments、video_likes、video_coins、video_favorites、favorites、favorite_items、follows、messages |
| 4 | video_qualities、video_metas、transcode_tasks |
| 5 | — |

### 缓存策略总计

| 阶段 | 数据 | 策略 |
|------|------|------|
| 1 | Token白名单 | Redis Set |
| 2 | 首页推荐、排行榜 | Redis List/ZSet 10min |
| 2 | 搜索热词 | Redis ZSet 每日更新 |
| 3 | 点赞/播放/投币/收藏计数 | Redis INCR + 5min同步MySQL |
| 3 | 弹幕池 | Redis ZSet 500条窗口 |
| 3 | 评论点赞、关注去重 | Redis Set |
| 3 | 未读消息数 | Redis INCR/DECR |

### 安全措施总计

| 阶段 | 措施 |
|------|------|
| 1 | 限流(5次/秒)、文件MIME校验、参数校验、bcrypt密码、HttpOnly Cookie |
| 2 | XSS防护(输入净化)、CSRF Token |
| 4 | HTTPS、Docker非root用户运行 |

### 前端 Store 总计

| Store | 阶段 | 职责 |
|-------|------|------|
| userStore | 1 | userInfo、isLoggedIn、login/logout/fetchUserInfo |
| playerStore | 2 | currentVideo/time/duration/isPlaying/quality/volume |
| danmakuStore | 3 | danmakuList、isConnected、color/type、sendDanmaku |
