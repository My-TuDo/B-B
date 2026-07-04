# 编码规范

> **定位**：项目 Agent 编码的硬性约束 + QA Agent 质量攻击的检查依据。
> 本文件中的每一条规范都是可检查的（QA Agent 能明确判断"遵守了"还是"违反了"）。

---

## 一、Go 后端规范

### 1.1 命名

| 元素 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写、单数、无下划线 | `user`、`video`、`danmaku` |
| 文件名 | 小写、下划线分隔 | `handler.go`、`repository.go`、`model.go` |
| 导出函数/类型 | PascalCase | `GetByID`、`VideoService` |
| 非导出函数/变量 | camelCase | `validateFile`、`maxUploadSize` |
| 接口 | 单方法用 `-er` 后缀 | `VideoRepository`（多方法接口） |
| 常量 | PascalCase（导出）或 camelCase（私有） | `MaxFileSize`、`defaultPageSize` |
| GORM Model 结构体 | `Entity` 后缀以区分 DTO | `VideoEntity`、`UserEntity` |
| DTO 结构体 | 语义化命名 | `VideoUploadReq`、`VideoDetailResp`、`UserPublicInfo` |

### 1.2 包结构

每个功能包必须包含以下标准文件（按需，非强制全部）：

```
internal/video/
├── handler.go       # HTTP 层：参数绑定、调用 service、返回 response
├── service.go       # 业务逻辑层
├── repository.go    # 数据访问接口 + GORM 实现
├── model.go         # Entity 结构体 + DTO 结构体
└── routes.go        # func RegisterRoutes(r *gin.RouterGroup)
```

| 规则 | 说明 |
|------|------|
| handler 不包含业务逻辑 | 只做参数绑定、校验、调用 service、返回响应 |
| service 不操作 HTTP | service 函数参数和返回值不包含 `gin.Context`、`http.Request` |
| repository 只操作数据库 | 不调用外部 API，不包含业务判断 |
| model.go 是本包的私有定义 | 不跨包共享 Entity，跨包只传 DTO |
| routes.go 注册本包路由 | `main.go` 只调用各包的 `RegisterRoutes()` |

### 1.3 错误处理

| 规则 | 示例 |
|------|------|
| 永远不忽略 error | `result, err := repo.Find(id)` → 必须检查 `if err != nil` |
| error 向上传递时用 `fmt.Errorf` 包装上下文 | `return fmt.Errorf("video.GetDetail: %w", err)` |
| handler 层的 error 转为统一错误码 | `response.Error(c, errcode.ErrVideoNotFound)` |
| 只用 `panic` 处理不可恢复的错误 | `panic` 仅限 `init()` 和 `main()` 启动阶段 |
| 不使用裸 `panic` | 必须通过 Recovery 中间件捕获并返回 500 + requestId |

### 1.4 Context 传递

| 规则 | 说明 |
|------|------|
| 所有跨层函数的第一个参数必须是 `context.Context` | `func (s *Service) GetDetail(ctx context.Context, id int64) (*VideoDetailResp, error)` |
| 不得使用包级变量存储请求状态 | ❌ `var currentUserId int64`，✅ `c.Set("userId", userId)` |
| gin.Context 不超过 handler 层 | service/repository 接收 `context.Context`，不接收 `*gin.Context` |

### 1.5 并发安全

| 规则 | 说明 |
|------|------|
| 共享 map 必须有 mutex 保护或使用 `sync.Map` | ❌ 多个 goroutine 同时读写普通 map |
| goroutine 必须有退出机制 | 使用 `context.Context` 传递取消信号，或使用 `chan struct{}` |
| channel 关闭由发送方负责 | 接收方不 close channel，避免 panic |
| `sync.WaitGroup` 的 `Add` 在 `go` 之前调用 | 避免 `Wait` 先于 `Add` 执行 |

### 1.6 数据库

| 规则 | 说明 |
|------|------|
| 所有查询使用参数化 | ❌ `db.Raw("SELECT * FROM users WHERE id = " + id)`，✅ `db.Where("id = ?", id)` |
| 迁移文件编号递增、不可回退 | `001_create_users.sql`、`002_create_videos.sql` |
| 每个新表必须定义索引 | 主键默认，复合索引在 model.go 的 struct tag 或 migration 中定义 |
| 连接池配置 | MaxOpenConns=100, MaxIdleConns=10, ConnMaxLifetime=1h |
| 不使用 `SELECT *` | 明确列出需要的字段 |

### 1.7 Redis

| 规则 | 说明 |
|------|------|
| Key 命名：`{domain}:{entity}:{id}` | `user:123:token`、`video:456:like_count` |
| 所有 Redis 操作设置超时 context | `client.Get(ctx, key)`，ctx 有超时 |
| 不使用 `KEYS *` | 生产禁用，用 `SCAN` 替代 |

### 1.8 日志

| 规则 | 说明 |
|------|------|
| 使用结构化日志 | `zap.L().Info("video uploaded", zap.Int64("videoId", id), zap.String("requestId", reqId))` |
| 不拼接字符串打日志 | ❌ `log.Printf("user %d login", userId)` |
| 敏感信息不进日志 | password、token、email 不出现在日志中 |
| 每个请求带 requestId | 中间件生成 UUID → 注入 Context → 写入响应头 `X-Request-Id` |

### 1.9 测试

| 规则 | 说明 |
|------|------|
| 测试文件和源文件同目录 | `handler_test.go` 与 `handler.go` 同目录 |
| 使用 Table-driven tests | Go 惯用风格 |
| DAO 测试使用 sqlite 内存库 | 不连接真实 MySQL |
| Service 测试 mock repository | 使用接口 + mock 实现 |

---

## 二、Vue/Nuxt 前端规范

### 2.1 命名

| 元素 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `VideoCard.vue`、`DanmakuLayer.vue` |
| 页面文件 | kebab-case（Nuxt 文件路由要求） | `video/[id].vue` |
| 组合式函数 | `use` 前缀 | `useApi.ts`、`useAuth.ts`、`useToast.ts` |
| Store | `use{Name}Store` | `useUserStore`、`usePlayerStore` |
| Props | camelCase | `videoId: string`、`isLiked: boolean` |
| Emits | kebab-case | `@update:quality`、`@video-ended` |
| CSS class | kebab-case | `.video-card`、`.comment-list` |
| 类型/接口 | PascalCase，无 `I` 前缀 | `VideoDetail`、`UserInfo` |
| API 函数 | `get/list/create/update/delete` 前缀 | `getVideoDetail`、`listVideos` |

### 2.2 组件设计

| 规则 | 说明 |
|------|------|
| 单文件组件 `<script setup lang="ts">` | Vue 3 标准写法 |
| Props 必须声明类型 | `defineProps<{ videoId: string }>()` |
| 组件不超过 300 行 | 超过则拆分 |
| 公共组件放 `components/common/` | AppHeader、LoadingSpinner、EmptyState、ErrorMessage、Toast |
| 业务组件放 `components/{domain}/` | `components/video/VideoCard.vue`、`components/danmaku/DanmakuLayer.vue` |

### 2.3 状态管理

| 规则 | 说明 |
|------|------|
| 全局状态用 Pinia Store | 用户状态、播放器状态、弹幕连接状态 |
| 页面内状态用 `ref/reactive` | 不创建 Store |
| Store 不直接操作 DOM | Store 管理数据，组件管理渲染 |
| 跨页面状态通过 Store 共享 | 不通过路由 query 传大对象 |

### 2.4 API 调用

| 规则 | 说明 |
|------|------|
| 所有 API 调用通过 `composables/useApi.ts` | 统一 baseURL、错误处理、CSRF Token 注入 |
| 不绕过 useApi | ❌ 直接 `$fetch` 或 `axios`，✅ `useApi().get('/videos/' + id)` |
| API 函数按模块分组 | `api/video.ts`、`api/user.ts`、`api/danmaku.ts` |
| 错误统一由 intercept 处理 | 401→跳登录、5xx→Toast 错误 |

### 2.5 类型安全

| 规则 | 说明 |
|------|------|
| TypeScript 严格模式 | `strict: true` in `tsconfig.json` |
| API 响应有类型定义 | `types/video.ts` 中定义 `VideoDetailResp` 等 |
| 不滥用 `any` | 不得使用 `any` 除非处理第三方库边缘情况 |
| 后端 DTO 和前端类型命名一致 | `VideoUploadReq`、`VideoDetailResp` 前后端名字对应 |

### 2.6 CSS

| 规则 | 说明 |
|------|------|
| 使用 Tailwind CSS | 优先使用原子类，不在组件内写大量自定义 CSS |
| 全局变量在 `variables.css` | 颜色、圆角、间距、阴影统一管理 |
| 不使用 `!important` | 除非覆盖第三方库样式 |
| 暗色/亮色用 CSS 变量驱动 | 不写两套样式 |
| 组件内 `<style scoped>` | 避免样式泄漏 |

---

## 三、安全规范（QA 重点检查）

### 3.1 认证与授权

| 规则 | 严重度 |
|------|--------|
| Token 只能通过 HttpOnly Cookie 传递，不在 localStorage | **critical** |
| 每个请求到达时，中间件必须从 Cookie 读取 Token → Redis 白名单验证 → 不通过返回 401 | **critical** |
| Redis 白名单删除 = 强制下线，用户修改密码时必须清除旧 Token | **critical** |
| 路由注册时必须区分公开路由和需要认证的路由组 | **major** |
| Admin 路由必须有 role 校验（middleware 或 handler 内判断 role ≥ 2） | **critical** |

### 3.2 输入校验

| 规则 | 严重度 |
|------|--------|
| 所有用户输入必须校验（格式、长度、范围）后处理 | **critical** |
| 文件上传校验顺序：MIME type → 魔数 → 大小 | **critical** |
| SQL 查询使用参数化，禁止字符串拼接 | **critical** |
| 用户输入的 HTML 内容输出时转义 | **major** |
| API 响应不暴露内部错误堆栈，只返回 requestId | **major** |

### 3.3 用户隔离（必检项）

| 规则 | 严重度 |
|------|--------|
| 不得使用包级变量存储 userId/token | **critical** |
| handler → service → repository 所有函数的第一个参数必须是 `context.Context` | **critical** |
| 数据查询必须带 userId 过滤（如 `WHERE user_id = ?`），不得查出其他用户的数据再过滤 | **critical** |
| Redis Key 必须带 `user:{userId}:` 前缀 | **major** |

### 3.4 限流与防护

| 规则 | 严重度 |
|------|--------|
| 注册/登录接口必须有限流（5次/分钟/IP） | **major** |
| 文件上传接口必须限制大小（≤500MB） | **major** |
| CSRF Token 校验（阶段2起所有 POST/PUT/DELETE 必须带 Header） | **major** |

---

## 四、Git 提交规范

| 规则 | 说明 |
|------|------|
| 分支命名 | `backend` / `frontend` 长期分支，feature 不另建分支 |
| Commit message | 中文，简洁描述做了什么 |
| 提交粒度 | 一个功能包完成一次提交 |
| 不提交的内容 | `.env`、`node_modules`、编译产物、IDE 配置 |

---

## 五、QA Agent 检查清单引用

QA Agent 执行质量攻击时，按以下顺序对照本文件：

1. **安全规范**（第三章）：每条逐一检查，critical 发现问题直接驳回
2. **Go 规范**：错误处理、Context 传递、并发安全、数据库
3. **前端规范**：组件命名、类型安全、API 调用封装
4. **Git 规范**：检查是否有不该提交的文件

对于每一条规范，QA 判定标准只有三个：**遵守 / 违反 / 不适用（本阶段未涉及）**。

---

## 六、规则演进：错误反馈闭环

规范不是写完就冻结的。编译预检和 QA 测试中发现的错误，如果现有规范没有覆盖，必须追加入规范，防止项目 Agent 再次犯同类错误。

### 6.1 反馈流程

```
编译预检 / QA 测试
     │
     ▼
发现错误
     │
     ├── 现有规范已覆盖？ → 标记为"违反"，按流程修复
     │
     └── 现有规范未覆盖？ → 追加新规则
            │
            ▼
         主Agent 判定：
         ├── Go 编码问题 → 追加到第一章对应小节
         ├── 前端编码问题 → 追加到第二章对应小节
         ├── 安全问题 → 追加到第三章对应小节
         ├── Git 问题 → 追加到第四章
         └── 跨阶段通用问题 → 追加到下方 6.2 累计规则表
```

### 6.2 累计规则（从错误中演化）

> 每发现一个规范未覆盖的错误类型，在此追加一行。

| 日期 | 来源阶段 | 错误描述 | 新增规则 | 严重度 |
|------|---------|---------|---------|--------|
| *暂无* | — | — | — | — |

### 6.3 示例（规则如何演化）

```
错误：阶段1 QA 发现视频上传接口没有校验空文件名
规范覆盖情况：编码规范 3.2 写了"所有用户输入必须校验"，但没具体到文件名
追加规则：→ 3.2 追加 "文件名不能为空、不能包含路径穿越字符(../)"

错误：阶段2 项目Agent 修复排行榜 bug 时改动了 feed/handler.go，导致首页回归
规范覆盖情况：orchestrator/workflow.md 5.2 写了"不改无关代码"，但项目Agent 越界了
追加规则：→ 6.2 追加 "修复 bug 只改本包文件，跨包修改需在修复报告中说明理由"
```

### 6.4 规则有效性审查

每完成 2 个阶段的锁版，主Agent 审查累计规则表：
- 是否有规则从未被触发？→ 可能过于宽松，考虑提升严重度
- 是否有规则被频繁违反？→ 考虑写入任务书的"本期重点检查"提示项目 Agent
