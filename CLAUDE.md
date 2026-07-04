# B-B 仿 B 站视频平台

## 技术栈
Go + Gin + GORM + MySQL 8.0 + Redis 7 + MinIO + RabbitMQ(阶段4)
Vue 3 + Nuxt 3 + Tailwind CSS + Video.js

## 项目仓库
- GitHub: `My-TuDo/B-B`
- 分支: `backend` (Go) + `frontend` (Vue/Nuxt)

## Agent 入口规则

### 所有 Agent 首次进入本项目的操作

1. 读 `docs/README.md` 了解目录结构和当前进度
2. 按 README.md 中的"按角色导航"加载对应目录下的文件
3. 不一次性加载 `docs/` 下所有文件

### 主 Agent 每次进入项目时

1. 读 `docs/orchestrator/backlog.md` 检查是否有人的新需求
2. 如有待处理需求 → 分析影响范围 → 更新对应阶段需求文件
3. 如无 → 按 README.md 中的"按角色导航"进行阶段下派

### 主 Agent 每次下派任务时

- 给项目 Agent 的任务书中必须包含"开始前必读文件列表"（指定 `builder/` 目录下的文件）
- 给 QA Agent 的任务书中必须包含"开始前必读文件列表"（指定 `qa/` 目录下的文件）

### 所有 Agent 禁止的行为

- 禁止一次性读取 `docs/` 下全部 `.md` 文件
- 禁止在未读 `docs/README.md` 的情况下直接读取其他文件
- 禁止读取不属于自己角色的目录（项目 Agent 不读 `orchestrator/`、QA Agent 不读 `builder/requirements/` 以外的 `builder/` 文件）
- 禁止给下游 Agent 暴露全局规划或未来阶段信息

## 文档目录

| 目录 | 角色 | 内容 |
|------|------|------|
| `docs/orchestrator/` | 主Agent | 工作流协议、功能清单、页面交互 |
| `docs/builder/` | 项目Agent | 编码规范、需求文件 |
| `docs/qa/` | QA Agent | 执行协议、回归锚点 |
| `docs/shared/` | 跨Agent | 接口契约 |

## 工作流

本项目的开发遵循三体 Harness 模型，详见 `docs/orchestrator/workflow.md`。
每阶段走：人+主Agent 审批需求 → 主Agent 下派 → 项目Agent 实现 → QA Agent 测试 → 人验收锁版。
