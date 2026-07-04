# Skills 调用协议

> **定位**：定义三个 Agent 可使用的 Skill 及其调用条件。所有 Agent 遵守此协议。本文件在 `shared/` 下，三个 Agent 均可读取。

---

## 一、项目Agent 可用 Skills

### 1.1 编译预检（每阶段必须）

| 调用点 | Skill/命令 | 说明 |
|--------|-----------|------|
| 编码完成后 | `go build ./...` | Bash 执行，零错误 |
| 编码完成后 | `go vet ./...` | Bash 执行，零警告 |
| 编码完成后 | `go test -race ./...` | Bash 执行，零 race（如有测试文件） |

### 1.2 运行验证

| 场景 | Skill | 说明 |
|------|-------|------|
| 需要验证功能可用 | `run` skill | 启动 Docker Compose，确认 API 可访问 |
| 需要验证前端页面 | `run` skill | 启动前端 Dev Server，浏览器确认 |

---

## 二、QA Agent 可用 Skills

### 2.1 质量扫描

| 调用点 | Skill | 参数 | 说明 |
|--------|-------|------|------|
| 编码规范扫描（第4层） | `code-review` | `--effort medium`，维度：correctness + security | 聚焦正确性和安全，关闭 style 和 simplify |
| 安全专项（第5层） | `code-review` | `--effort high`，维度：security | 深度安全审查 |
| 修复验证 | `simplify` | 对修复 diff 做简化检查 | 确保修复没有引入冗余代码 |

### 2.2 端到端验证

| 调用点 | Skill | 说明 |
|--------|-------|------|
| 正确性攻击（第3层） | `verify` skill | 逐条对照验收条件，启动应用，驱动 API 端到端验证 |
| 回归锚点复验（第2层） | `verify` skill | 执行 regression-anchors.md 中的所有锚点 API 调用 |
| 启动环境 | `run` skill | 启动 Docker Compose，确保服务可用 |

### 2.3 代码扫描

| 调用点 | 命令 | 说明 |
|--------|------|------|
| SQL 注入扫描 | `grep`（代码模式匹配）| 搜索字符串拼接 SQL |
| 敏感信息扫描 | `grep`（代码模式匹配）| 搜索硬编码密钥、token 入日志 |

---

## 三、主Agent 可用 Skills

| 调用点 | Skill | 说明 |
|--------|-------|------|
| 需求审批前 | 读 `backlog.md` | 检查人的新需求 |
| 锁版后 | `Write` + `Edit` | 更新 contracts.md 和 stage-{N}-legacy.md |
| 下派任务 | `Agent` 工具 | 派项目Agent / QA Agent |

---

## 四、项目自定义 Skills

所有自定义 Skill 定义在 `.claude/settings.json` 中，通过 `/skill-name` 或 Agent 的 Bash 调用执行。

### 4.1 Go 后端 Skills

| Skill | 用途 | 使用者 |
|-------|------|--------|
| `go-init` | 初始化 Go module + 依赖 | 项目Agent |
| `go-dep` | 安装/更新 Go 依赖 | 项目Agent |
| `go-qa` | build + vet + race 聚合检查 | 项目Agent、QA Agent |
| `go-lint` | golangci-lint 风格检查 | 项目Agent、QA Agent |
| `go-errcheck` | 未处理 error 返回值检查 | QA Agent |
| `go-run` | 启动 Go 后端（air 热重载） | 项目Agent、人 |

### 4.2 前端 Skills

| Skill | 用途 | 使用者 |
|-------|------|--------|
| `frontend-init` | `pnpm install` 安装前端依赖 | 项目Agent |
| `frontend-dev` | `pnpm dev` 启动前端 Dev Server | 项目Agent、人 |
| `frontend-build` | `pnpm build` 生产构建 | 项目Agent |
| `frontend-lint` | ESLint 代码检查 | 项目Agent、QA Agent |

### 4.3 运维 Skills

| Skill | 用途 | 使用者 |
|-------|------|--------|
| `docker-up` | Docker Compose 启动全栈 | 项目Agent、QA Agent、人 |
| `docker-down` | Docker Compose 停止 | 项目Agent、人 |
| `docker-reset` | 清理数据卷 + 重建 | 项目Agent、人 |
| `docker-logs` | 查看容器日志 | 项目Agent、QA Agent、人 |
| `dev-all` | 一键启动基础设施（MySQL+Redis+MinIO） | 项目Agent、QA Agent |

### 4.4 API 测试 Skills

| Skill | 用途 | 使用者 |
|-------|------|--------|
| `api-test` | `curl GET` + JSON 格式化 | QA Agent、人 |
| `api-post` | `curl POST` + JSON body | QA Agent、人 |

---

## 五、调用约束

| 规则 | 说明 |
|------|------|
| 项目Agent 不得调用 `code-review` | 不审查自己写的代码 |
| QA Agent 不得调用 `simplify` 在首次扫描 | 只在修复验证阶段使用 |
| QA Agent 不得跳过回归锚点 | `verify` 首先执行锚点列表 |
| 所有 Agent 不得修改不属于自己职责的文件 | 权限边界见 agent-workflow.md 8.3 |
| 所有 Agent 不得自行修改 `.claude/settings.json` | 新增 Skill/权限必须写入 `backlog.md` 或口头提案，由人审批后手动配置 |

---

## 六、Skill 缺口处理协议

Agent 发现当前缺少完成工作所需的 Skill 时：

```
发现 Skill 缺口
     │
     ├── 内置 Skill 可替代？ → 使用内置 Skill
     │
     └── 需要自定义 Skill → 
            │
            ├── 汇总缺口信息（需要什么能力、为什么现有 Skill 不够）
            ├── 生成配置建议（settings.json 片段）
            ├── 写入 docs/orchestrator/backlog.md（类型：新增Skill）
            │   或
            └── 口头告知人："我需要 X 权限来完成 Y，请在 settings.json 中加..."
            
人审批 → 手动修改 .claude/settings.json → Agent 验证 Skill 可用 → 继续工作
```

**禁止行为**：
- Agent 不得直接 `Write` 或 `Edit` `.claude/settings.json`
- Agent 不得在未经人确认的情况下执行需要新增权限的操作
- Agent 不得绕过 Skill 协议直接调用底层工具（除非协议明确允许）
