# B-B

> B 站风格视频平台 — Go + Nuxt 全栈项目

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.21+ / Gin / GORM / MySQL 8.0 / Redis 7 / MinIO |
| 前端 | Nuxt 3 / Vue 3 / TypeScript / Tailwind CSS |
| 基础设施 | Docker Compose |

## 快速启动

```bash
# 1. 启动全部服务
docker compose up -d

# 2. 访问
#    前端: http://localhost:3000
#    后端: http://localhost:8080
#    MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
```

## 本地开发

```bash
# 启动依赖服务
docker compose up -d mysql redis minio

# 终端 1: 后端
cd backend && go run cmd/server/main.go

# 终端 2: 前端
cd frontend && pnpm install && pnpm dev
```

## 项目结构

```
B-B/
├── backend/
│   ├── cmd/server/main.go          # 入口
│   ├── internal/
│   │   ├── handler/                # HTTP 层 (auth/user/video/category)
│   │   ├── service/                # 业务逻辑层
│   │   ├── repository/             # 数据访问层
│   │   ├── model/                  # Entity + DTO
│   │   └── middleware/             # Recovery/Logger/CORS/RateLimit/Auth
│   ├── pkg/                        # 公共包 (config/database/jwt/response/...)
│   └── migrations/                 # SQL 迁移
├── frontend/
│   ├── pages/                      # 文件路由 (7 页面)
│   ├── components/                 # 通用 + 业务组件
│   ├── composables/                # useApi/useToast/useTheme
│   ├── stores/                     # Pinia (userStore)
│   ├── layouts/                    # default / auth
│   ├── middleware/                  # auth / guest
│   └── plugins/                    # auth 恢复登录态
└── docker-compose.yml
```

## 阶段一功能

- 用户注册 / 登录 / 登出（JWT + Redis 白名单 + HttpOnly Cookie）
- 视频上传（MinIO + SSE 实时进度 + MIME/魔数/大小三重校验）
- 视频播放（MinIO 预签名 URL）
- 稿件管理（全部/草稿/已发布 Tab + 编辑/发布/删除）
- YouTube 风格侧边栏（展开 240px / 折叠 72px，分类手风琴）
- 暗色 / 亮色双模切换（CSS 变量 + localStorage 记忆）
- 7 个分类（动画/音乐/游戏/知识/生活/影视/科技）

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/logout` | 登出 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| GET | `/api/v1/auth/me` | 当前用户 |
| GET | `/api/v1/users/:id` | 用户信息 |
| PUT | `/api/v1/users/:id` | 更新用户 |
| POST | `/api/v1/videos` | 上传视频 |
| GET | `/api/v1/videos` | 视频列表 |
| GET | `/api/v1/videos/:id` | 视频详情 |
| GET | `/api/v1/videos/:id/play-url` | 播放地址 |
| PUT | `/api/v1/videos/:id` | 更新视频 |
| DELETE | `/api/v1/videos/:id` | 删除视频 |
| GET | `/api/v1/videos/users/:id/videos` | 用户视频 |
| GET | `/api/v1/categories` | 分类列表 |

## License

MIT
