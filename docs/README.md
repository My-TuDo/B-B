# Docs 目录

> **此文件是 `docs/` 的入口索引。Agent 进入本目录时，先读此文件，按角色定位到对应子目录。禁止一次性加载全部文件。**

---

## 目录结构

```
docs/
├── README.md                        ← 你在这里
├── orchestrator/                    ← 主Agent
│   ├── workflow.md                  # 三体Harness工作流协议
│   ├── feature-list.md              # 5阶段功能清单+技术栈+架构
│   ├── page-interaction.md          # 页面交互规格
│   └── backlog.md                   # 需求后门（人→主Agent）
├── builder/                         ← 项目Agent
│   ├── standards.md                 # 编码规范+规则演化
│   └── requirements/                # 阶段需求文件
│       └── _TEMPLATE.md
├── qa/                              ← QA Agent
│   ├── checklist.md                 # QA执行协议(6层)
│   └── regression-anchors.md        # 回归锚点
└── shared/                          ← 跨Agent共享
    ├── contracts.md                 # 跨阶段接口契约
    └── skills.md                    # Skills 调用协议
```

---

## 按角色导航

### 主 Agent（编排者）

| 文件 | 何时读 |
|------|--------|
| `orchestrator/workflow.md` | 首次进入项目时必读 |
| `orchestrator/feature-list.md` | 生成任务书前，查阶段范围和依赖 |
| `orchestrator/page-interaction.md` | 生成前端相关的任务书时查阅 |
| `shared/contracts.md` | 生成任务书前，查已有接口契约 |
| `builder/requirements/{stage}.md` | 审批需求时（Gate 1） |

### 项目 Agent（构建者）

| 文件 | 何时读 |
|------|--------|
| `builder/requirements/{stage}.md` | 开始编码前，唯一必读的需求文件 |
| `builder/standards.md` | 编码时参考，提交前自检 |
| `shared/contracts.md` | 编码前，只读本阶段依赖的接口部分 |

### QA Agent（测试者）

| 文件 | 何时读 |
|------|--------|
| `qa/checklist.md` | 开始测试前必读（执行协议） |
| `builder/requirements/{stage}.md` | 读取验收条件，构造攻击场景 |
| `qa/regression-anchors.md` | 第2层回归检查必读 |
| `builder/standards.md` | 第4层编码规范扫描必读 |
| `shared/contracts.md` | 接口一致性校验 |

---

## 文档建设进度

| 步骤 | 文件 | 状态 |
|------|------|------|
| ① Agent 工作流 | `orchestrator/workflow.md` | ✅ |
| ② 产品功能清单 | `orchestrator/feature-list.md` | ✅ |
| ③ 页面交互说明 | `orchestrator/page-interaction.md` | ✅ |
| ④ 技术选型 | `orchestrator/feature-list.md` 中覆盖 | — |
| ⑤ 架构设计 | `orchestrator/feature-list.md` 中覆盖 | — |
| ⑥ 编码规范 | `builder/standards.md` | ✅ |
| ⑦ QA 检查清单 | `qa/checklist.md` | ✅ |

## 实施进度

| 阶段 | 需求文件 | 状态 |
|------|---------|------|
| 1 | `builder/requirements/stage-01-core.md` | ⬜ 待建 |
| 2 | `builder/requirements/stage-02-discovery.md` | ⬜ 待建 |
| 3 | `builder/requirements/stage-03-social.md` | ⬜ 待建 |
| 4 | `builder/requirements/stage-04-media.md` | ⬜ 待建 |
| 5 | `builder/requirements/stage-05-deploy.md` | ⬜ 待建 |
