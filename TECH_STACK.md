# B-B 技术栈总览

## 前端

| 层级 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 框架 | **Nuxt** | 4.x (Nitro 2.x) | SSR/CSR 混合框架，文件路由 |
| 渲染 | **Vue 3** | 3.5+ | Composition API + `<script setup>` |
| 语言 | **TypeScript** | strict 模式 | 类型安全 |
| 样式 | **Tailwind CSS** | 3.x | 原子化 CSS |
| 包管理 | **pnpm** | 11+ | 严格依赖管理 |
| 构建 | **Vite** | 7.x | HMR + 生产构建 |
| 状态管理 | **Pinia** | 3.x | userStore |
| 路由 | **Vue Router** | 5.x | Nuxt 文件路由 |
| HTTP | **fetch** (内置) | — | useApi 封装（credentials: include） |
| 实时通信 | **WebSocket** | 原生 API | 弹幕实时推送 |
| 动画 | **CSS Animation** | — | 弹幕飘过 + UI 过渡 |
| 视频 | **原生 `<video>`** | — | 播放器（阶段 4 换 Video.js） |
| 设计规范 | **Taste-skill** | v2 | UI 品味约束（布局/动效/密度） |
| Skills | **senior-fullstack** | — | 项目 Agent 全栈规范 |
| Skills | **senior-qa** | — | QA Agent 测试规范 |

### 前端目录结构
```
frontend/
├── pages/          # 13 页面（文件路由）
├── components/     # 通用 + 业务组件（danmaku/comment/video/common）
├── composables/    # useApi / useToast / useTheme
├── stores/         # userStore (Pinia)
├── plugins/        # auth.ts（登录态恢复 + CSRF）
├── middleware/      # auth.ts / guest.ts
├── layouts/        # default.vue / auth.vue
├── types/          # TypeScript 接口
└── assets/styles/  # CSS 变量 + 全局样式
```

---

## 后端

| 层级 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 语言 | **Go** | 1.21+ | 后端主语言 |
| 框架 | **Gin** | 1.x | HTTP 路由 + 中间件 |
| ORM | **GORM** | 2.x | MySQL 操作 |
| 配置 | **Viper** | 1.x | 环境变量 + YAML |
| 日志 | **Zap** | 1.x | 结构化 JSON 日志 |
| 校验 | **go-playground/validator** | v10 | 参数校验 |
| JWT | **golang-jwt/jwt** | v5 | HS256 Token |
| 密码 | **bcrypt** | cost=12 | 密码哈希 |
| UUID | **google/uuid** | 1.x | 文件命名 |
| 实时 | **gorilla/websocket** | 1.x | 弹幕 WebSocket |

### 后端目录结构
```
backend/
├── cmd/server/main.go     # 入口：初始化 + AutoMigrate + 路由
├── internal/
│   ├── handler/           # HTTP 层 (14 模块)
│   ├── service/           # 业务逻辑层 (14 模块)
│   ├── repository/        # 数据访问层 (14 模块)
│   ├── model/             # Entity + DTO (9 模块)
│   ├── middleware/        # Recovery/RequestID/Logger/CORS/RateLimit/Auth/CSRF
│   └── ws/                # WebSocket Hub
├── pkg/                   # 公共包 (config/database/errcode/jwt/logger/response/storage/validator)
└── migrations/            # SQL 迁移
```

---

## 基础设施

| 组件 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 数据库 | **MySQL** | 8.0 | 主存储 |
| 缓存 | **Redis** | 7 | Token 白名单 + 弹幕池 + 排行榜 + 去重 |
| 对象存储 | **MinIO** | latest | 视频/封面文件存储 |
| 容器 | **Docker** | — | 开发环境一键启动 |
| 编排 | **Docker Compose** | 3.x | 五服务编排 |

### 服务端口
| 服务 | 端口 | 说明 |
|------|------|------|
| MySQL | 3307 | 外部映射（宿主机 3306 被占用） |
| Redis | 6379 | |
| MinIO API | 9000 | |
| MinIO Console | 9001 | Web 管理界面 |
| Go 后端 | 8080 | Gin HTTP Server |
| Nuxt 前端 | 3000 | Vite Dev Server |

---

## 数据表（16 张）

| 阶段 | 表 | 说明 |
|------|-----|------|
| 1 | `users` | 用户（bcrypt 密码 + role） |
| 1 | `videos` | 视频（MinIO + 草稿/发布/审核/删除状态） |
| 1 | `categories` | 分类（7 种子数据） |
| 2 | `tags` | 标签 |
| 2 | `video_tags` | 视频-标签多对多 |
| 2 | `video_history` | 观看历史（进度 + 续播） |
| 3 | `danmaku` | 弹幕（内容 + 颜色 + 位置 + 播放时间点） |
| 3 | `comments` | 评论（楼中楼 parent_id/root_id） |
| 3 | `video_likes` | 点赞（去重） |
| 3 | `video_coins` | 投币（去重 + 每日限额） |
| 3 | `favorites` | 收藏夹 |
| 3 | `favorite_items` | 收藏夹-视频关联 |
| 3 | `follows` | 关注关系 |
| 3 | `messages` | 通知（类型 + 已读/未读） |

---

## 中间件链

```
Recovery → RequestID → Logger → CORS → RateLimit → [路由] → Auth(路由级) → CSRF(路由级)
```

| 中间件 | 功能 |
|--------|------|
| Recovery | panic 兜底 + 记录 stack |
| RequestID | UUID 生成/透传 |
| Logger | Zap JSON 日志，脱敏 |
| CORS | localhost:3000, credentials: true |
| RateLimit | 全局令牌桶 100/s，登录/注册 IP 5/min |
| Auth | Cookie → JWT → Redis 白名单 |
| CSRF | X-CSRF-Token Header ↔ csrf_token Cookie |

---

## API 统计

| 阶段 | 新增 | 累计 |
|------|------|------|
| 阶段一 | 15 | 15 |
| 阶段二 | 18 | 33 |
| 阶段三 | 25 | 58 |

---

## 项目亮点

### 安全设计

**双重认证体系**
- JWT HS256（7 天过期）+ Redis Token 白名单，登出即时失效
- 每请求 Cookie → JWT 解析 → Redis 验证，任一失败 401
- 阶段一即实现，阶段二叠加 CSRF Token 双重防护

**CSRF 防护**
- 所有 POST/PUT/DELETE 需 `X-CSRF-Token` Header 匹配 `csrf_token` Cookie
- 公开路由智能豁免（登录/注册）
- 前端 useApi 自动携带，开发者无感

**用户隔离**
- 所有写操作验证 `token.userId == resource.userId`
- handler → service → repository 全链路 context 传递
- Redis Key 命名空间隔离 `auth:token:{userId}`

**文件安全**
- 三重校验：MIME 白名单 → 文件头魔数（前 512 字节）→ 大小上限 500MB
- MinIO 对象名 UUID 命名，用户文件名不进入存储路径
- 预签名 URL 1 小时过期，无法猜测下载链接

**防注入 + 脱敏**
- SQL 全参数化，零字符串拼接
- Zap 日志自动脱敏 password/token/secret 字段
- 错误响应不暴露堆栈，仅返回 requestId

### 实时弹幕系统

- **WebSocket Room 模式**：gorilla/websocket Hub 按 video_id 分 room，广播仅推送到同视频观众
- **Redis ZSet 弹幕池**：每视频 500 条上限，score=play_time 精确排序
- **前端轨道分配**：5 轨道 CSS Animation GPU 加速，弹幕不碰撞
- **时间同步**：弹幕按 `play_time` 精准出现，拖拽进度条即时回显历史弹幕
- **乐观渲染**：发送者本地立即显示（临时 ID），WebSocket 回传去重替换

### 社区互动体系

**楼中楼评论**
- parent_id + root_id 双字段，支持无限层级嵌套
- 递归 CommentNode 组件，内联回复输入框
- Redis Set 去重点赞，5 分钟批量同步 MySQL

**三连系统**
- 点赞/投币/收藏均为切换态，乐观 UI 更新
- 投币 Redis Lua 原子脚本防并发，每视频限 1 次 + 每日 5 枚全局上限
- 收藏夹自定义 + 默认收藏夹自动创建

**通知系统**
- 评论回复/点赞/关注三类通知
- Redis INCR 未读计数
- 前端铃铛红点 + 通知列表

### 搜索架构

- MySQL FULLTEXT 优先 + LIKE 降级（兼容中文 + 短关键词）
- 搜索范围：标题 + 简介 + 标签名（LEFT JOIN 三表查询）
- 建议接口 UNION 标题 + 标签名
- 300ms 防抖实时建议下拉

### 排行榜 + 热度算法

- Redis ZSet 日/周/总排行，TTL 10min 防穿透
- 热度算法 `score = views × 0.5 - hours × 0.1`（likes/comments 权重阶段 3 激活）
- 首次查询/过期自动回源 MySQL 重建

### 渐进式用户体验

**暗/亮双模**：CSS 变量驱动，localStorage 记忆，一键切换全站无闪烁

**页面刷新保持登录**：`plugins/auth.ts` 启动时调用 `/auth/me` + 5s 超时保底，刷新不丢登录态

**断点续播**：5 秒间隔记录播放进度，拖动进度条即时保存，历史页一键续播

**分段式进度条**：20 段格子，按视频真实时长比例填充（播放器 loadedmetadata 回传 duration）

**Taste-skill UI 优化**：工业级排版/动效/密度约束，反 AI 模板化设计

### 全栈 Agent 协作体系

- **Harness 三体模型**：主 Agent（决策）+ 项目 Agent（编码）+ QA Agent（测试）
- **多 Skill 系统**：senior-fullstack / senior-qa / design-taste-frontend 全流程规范
- **QA 严格分层**：编译预检 → 回归锚点 → 正确性攻击 → 编码规范 → 安全扫描 → 报告
- **契约锁版**：每阶段 locked 接口签名不可变，仅追加不修改
- **Playwright 自动化**：主 Agent 可直接操作浏览器验证 UI 状态

---

## 研发规范

### 后端
- 分层: handler → service → repository（不跨层调用）
- `context.Context` 第一参数
- 错误传递: `fmt.Errorf("pkg.Func: %w", err)`
- JSON tag: snake_case
- SQL: GORM 参数化，禁字符串拼接
- Redis Key: `{domain}:{entity}:{id}`
- 用户隔离: token.userId == resource.userId

### 前端
- `<script setup lang="ts">` + strict TypeScript
- 全部 API 经 useApi（upload.vue SSE 除外）
- 401 → /login（`/api/v1/auth/*` 除外）
- Tailwind + CSS 变量 `var(--color-*)`
- 暗/亮双模: `data-theme="dark|light"`
- CSRF Token 自动携带

### Cookie
- `token`: JWT HS256, HttpOnly, SameSite=Lax, 7天
- `csrf_token`: 随机 32 字节 hex, 非 HttpOnly

---

## 开发命令

```bash
# 启动全部服务
docker compose up -d mysql redis minio

# 后端
cd backend && go run cmd/server/main.go

# 前端
cd frontend && pnpm dev

# 编译检查
cd backend && go build ./... && go vet ./...
cd frontend && pnpm build
```
