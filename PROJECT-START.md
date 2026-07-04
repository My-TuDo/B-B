# 项目启动指南

## 启动 Claude Code

```bash
cd "F:/Code/web/B-B-temp"
code .    # 打开 VS Code（可选）
```

Claude Code 会自动加载 `CLAUDE.md`。

---

## 启动主 Agent

复制以下指令发给 Claude Code：

```
你作为主Agent，角色是编排者。请先读 CLAUDE.md，然后执行以下步骤：

1. 读 docs/orchestrator/backlog.md，检查是否有人的新需求
2. 读 docs/orchestrator/feature-list.md，了解 5 阶段规划
3. 告诉我当前进度，我们一起决定下一步做什么
```

---

## 常用指令

| 场景 | 指令 |
|------|------|
| 新增功能需求 | 写入 `docs/orchestrator/backlog.md` 的"待处理"区 |
| 开始下一阶段 | "开始阶段 N"（主 Agent 会生成需求文件、下派项目 Agent） |
| 查看当前进度 | "当前各阶段完成状态" |
| 新增 Skill | 写入 backlog（类型：新增Skill），或直接让主 Agent 配置文件 |
| 调整 UI 设计 | "读 docs/orchestrator/page-interaction.md，修改 XX 页面的 XX 部分" |

---

## 工作流速查

```
人的需求审批(Gate1)
  → 主Agent 下派(任务书)
    → 项目Agent 实现(交付摘要+代码)
      → QA Agent 测试(QA报告)
        → 人验收(Gate2)
          → locked → 进入下一阶段
```

---

## 分支

| 分支 | 用途 |
|------|------|
| `backend` | Go 后端代码 |
| `frontend` | Vue/Nuxt 前端代码 |

---

## 快速链接

| 文档 | 用处 |
|------|------|
| `docs/orchestrator/backlog.md` | 写入新需求 |
| `docs/orchestrator/feature-list.md` | 查看各阶段功能清单 |
| `docs/orchestrator/workflow.md` | 查看完整 Harness 工作流协议 |
| `docs/orchestrator/page-interaction.md` | 查看页面 UI 规格 |
| `docs/builder/standards.md` | 查看编码规范 |
| `docs/qa/checklist.md` | 查看 QA 检查流程 |
