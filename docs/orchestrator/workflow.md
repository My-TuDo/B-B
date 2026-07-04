# Agent 项目开发工作流 —— 三体 Harness 模型

> **定位**：本项目实操手册。定义基于 Harness Engineering 思想的 3 Agent 协作模型，覆盖从需求到合入的完整开发循环。
>
> **适用场景**：个人独立开发的中大型 Golang 项目。

---

## 一、核心原则

| 原则 | 说明 |
|------|------|
| **上下文隔离** | 每个 Agent 只拿到刚好够用的信息，不暴露全局视图 |
| **对抗验证** | 项目实现和 QA 是两个独立 Agent，QA 的职责是"破坏"，不是"确认" |
| **版本锁定** | 人测试完全通过后锁版本，未锁版前不进下一阶段，防止问题累积 |
| **文件化状态** | 所有状态、契约、报告都以文件形式持久化，对话可以中断、恢复、追溯 |
| **渐进式披露** | 每层只处理自己层级的决策，信息按以下漏斗传递： |

```
需求大盘（全量信息）
     │
     ▼
主 Agent ─── 全量视图，但只输出一个阶段的规格
     │
     ├──▶ 项目 Agent ─── 只看到本阶段需求 md（不知道其他阶段细节）
     │
     └──▶ QA Agent ─── 看到需求 md + 项目交付物（不知道主Agent的其他阶段规划）
```

---

## 二、三体 Harness 模型

### 2.1 交互模型

| Agent | 角色 | 粒度 | 回答的问题 |
|-------|------|------|-----------|
| **主 Agent** | 编排者 / 上下文分解器 | 每阶段 | "本阶段做什么"（What） |
| **项目 Agent** | 阶段内全面构建者 | 每阶段 | "怎么做"（How） |
| **QA Agent** | 破坏者 / 攻击者 | 每阶段 | "做对了吗"（Verify） |
| **人** | 最终裁决者 | 每阶段结束 | "可以锁版吗"（Decide） |

```
阶段 N 需求 md（人 + 主Agent 产出，status: approved）
     │
     ▼
主 Agent → 生成任务书 → 下派给项目 Agent
     │
     ▼
项目 Agent（本阶段所有组件完整实现）
  ├── 编码
  ├── go build + go vet（自检，必须通过）
  └── 交付
     │
     ▼
QA Agent → 压测 / 攻击 / 回归 → QA 报告
     │
     ▼
人
  ├── 读 QA 报告
  ├── 自己动手测试
  ├── 通过 → 锁版 → 更新 contracts.md → 阶段 N+1
  └── 不通过 → 测试文档写入 QA 报告（分界线以下）
                │
                ▼
          QA Agent 根据人的文档定位问题 → 回传项目 Agent
                │
                ▼
          项目 Agent 修复 → QA 再测 → 人再验
                │
                │ 仍不通过？
                ▼
          升级到主 Agent：需求 md 或任务书有问题，重新分析后重新下派
```

### 2.2 信息流协议

```
主 Agent
  输入：docs/builder/requirements/{stage}.md + docs/shared/contracts.md
  输出：任务书（HOW 视角的结构化规格，覆盖阶段内所有组件）
  不暴露：其他阶段的规划、全局架构决策

项目 Agent
  输入：任务书 + docs/shared/contracts.md
  输出：代码 diff + 自检报告（含 go build / go vet 结果）
  不暴露：需求 md 原文、主Agent 的全局规划

QA Agent
  输入：需求 md + 项目交付 diff + contracts.md
  输出：QA 报告（结构化）
  不暴露：主Agent 的全局规划
```

### 2.3 主 Agent 的任务书协议

主 Agent 将需求 md（WHAT 视角）重写为任务书（HOW 视角）：

| 需求 md（WHAT） | 任务书（HOW） |
|----------------|--------------|
| 功能概述 | → 阶段内所有需要创建/修改的文件列表 |
| 前置依赖 | → 需要导入的包和接口（来自 contracts.md） |
| 数据模型 | → GORM model 定义、migration 文件名 |
| API 接口 | → handler 函数签名、路由注册位置 |
| 业务规则 | → 关键实现路径 + 错误处理清单 |
| 验收条件 | → 自测 checklist |
| 非功能约束 | → 编码规范引用 |
| 本期不做 | → stub 标记清单 |

### 2.4 与传统三体的区别

| 维度 | 旧（按模块） | 新（按阶段） |
|------|------------|------------|
| 下派粒度 | 每个组件一个实现Agent | **每阶段一个项目Agent** |
| Agent 数量 | 一个阶段可能派 N 个 | **一个阶段只派 1 个** |
| QA 节奏 | 实现后立即 2 轮对抗 | **项目Agent 完整交付后一次性压测** |
| 人的角色 | Gate 1 + Gate 2 | **需求审批 + 最终测试验收** |
| 锁版 | 模块 done | **人测试通过后锁版 → 下一阶段** |
| 失败升级 | 2轮固定超标升级 | **人测两次不通过 → 升级到主Agent** |

### 2.5 任务书模板

主 Agent 下派时，任务书中必须包含文件读取指令，确保渐进式披露：

**项目 Agent 任务书模板**：

```markdown
# 任务书：{阶段名}

## 开始前必读（按顺序）
1. docs/builder/requirements/{stage}.md          — 本阶段需求规格
2. docs/shared/contracts.md                       — 只读本章节：[列出本阶段依赖的接口]
3. docs/builder/requirements/stage-{N-1}-legacy.md — 上一阶段遗产（阶段2起必读）

## 你的工作
{具体的 HOW 视角任务拆解}

## 禁止读取
- docs/orchestrator/feature-list.md    — 你不需要知道后续阶段
- docs/qa/regression-anchors.md        — 与你无关
- docs/orchestrator/page-interaction.md — 前端部分由前端 Agent 处理（如适用）

## 交付前必做
- go build ./...    — 零错误
- go vet ./...      — 零警告
- 编写交付摘要（按 workflow.md 5.2 模板）
```

**QA Agent 任务书模板**：

```markdown
# QA 任务书：{阶段名}

## 开始前必读（按顺序）
1. docs/qa/checklist.md            — QA 执行协议
2. docs/builder/requirements/{stage}.md    — 验收条件
3. docs/qa/regression-anchors.md      — 回归锚点全量复验
4. docs/builder/standards.md        — 编码规范逐条扫描
5. docs/shared/contracts.md               — 接口一致性校验

## 执行流程
严格按 qa-checklist.md 的步骤执行，不跳过任何层。

## 禁止读取
- docs/orchestrator/feature-list.md             — 你不需要知道全局规划
```

---

## 三、需求目录结构

### 3.1 目录规范

```
docs/builder/requirements/
├── _TEMPLATE.md          # 需求 md 模板
├── stage-01-core.md
├── stage-02-discovery.md
└── ...
```

### 3.2 需求 md 模板

每份需求 md 由 **frontmatter（状态元数据）** + **正文（规格）** 组成：

```markdown
---
stage: 阶段一：基础架构+用户+视频
status: draft            # draft → approved → implementing → qa-testing → human-testing → locked / blocked
created: 2026-07-04
updated: 2026-07-04
gate1_by:               # Gate 1 审批人（status=approved 时填写）
locked_by:              # 锁版人（status=locked 时填写）
deps:
  - {依赖的阶段或contracts接口}
qa_issues: []
---

# 阶段一：基础架构+用户+视频

## 1. 功能概述
...

## 2. 前置依赖
| 依赖模块 | 依赖内容 | 接口/表 |
|---------|---------|--------|
| ... | ... | ... |

## 3. 数据模型
...

## 4. API 接口
...

## 5. 业务规则
...

## 6. 验收条件
- [ ] ...

## 7. 非功能约束
...

## 8. 本期不做
...
```

完整模板见 `docs/builder/requirements/_TEMPLATE.md`。

### 3.3 自包含原则

一份合格的需求 md 必须满足三个约束：

| 约束 | 要求 | 反例 |
|------|------|------|
| **自包含** | 实现 Agent 拿到这一份文件即可开始编码，无需查其他资料 | "参考 B 模块的接口" → 应该写死接口签名或引用 contracts.md |
| **可验证** | QA Agent 的每项检查都能对应到一个具体条款 | "系统应该高性能" → 应该写"P99 ≤ 200ms" |
| **有边界** | 明确标出本期做与不做，不给实现 Agent 留自由发挥空间 | 漏标"本期不做"导致实现 Agent 过度工程 |

### 3.4 需求拆分流程（Gate 1）

```
人起草模块列表
     │
     ▼
主 Agent 检查：
  - 每个模块是否自包含？
  - 依赖关系是否形成 DAG（无循环）？
  - 验收条件是否可验证？
     │
     ▼
人审批每份需求 md ─── status: draft → approved
     │
     ▼
模块进入可下派队列
```

**Gate 1 人的决策清单**：
- [ ] 模块边界清晰？（不会做着做着越界到另一个模块）
- [ ] 本期不做的内容是否明确标出？
- [ ] 依赖模块的接口是否已确认（或已在 contracts.md 中）？
- [ ] 验收条件是否覆盖了 Happy Path + 主要 Error Case？
- [ ] 非功能约束是否合理（不过严也不过松）？

---

## 四、状态与记忆系统

### 4.1 阶段状态机

```
              ┌──────────┐
              │  draft   │  人起草
              └────┬─────┘
                   │ Gate 1 审批通过
                   ▼
              ┌──────────┐
              │ approved │  等待主Agent下派
              └────┬─────┘
                   │ 主Agent生成任务书，下派给项目Agent
                   ▼
           ┌─────────────┐
           │implementing │  项目Agent编码中
           └──────┬──────┘
                  │ 编译预检通过 → 交付
                  ▼
           ┌─────────────┐
           │ qa-testing   │  QA Agent 压测/攻击中
           └──────┬──────┘
                  │ QA 报告产出
                  ▼
           ┌──────────────┐
           │human-testing │  人读报告 + 自己动手测
           └──────┬───────┘
                  │
         ┌────────┴────────┐
         ▼                 ▼
    ┌────────┐      ┌──────────┐
    │ locked │      │ fixing   │ 人写测试文档 → QA定位 → 项目Agent修复
    └────────┘      └────┬─────┘
    锁版 → 下一阶段      │ 修复完成 → 回到 qa-testing → 人再验
                         │
                         │ 仍不通过？
                         ▼
                    ┌──────────┐
                    │ blocked  │ 升级到主Agent：需求md或任务书有问题
                    └──────────┘
```

### 4.2 状态流转规则

| 状态变更 | 触发条件 | 执行者 |
|---------|---------|--------|
| draft → approved | Gate 1 审批通过 | **人** |
| approved → implementing | 主Agent 下派任务 | 主Agent |
| implementing → qa-testing | 项目Agent 交付 + 编译预检通过 | 项目Agent |
| qa-testing → human-testing | QA 报告产出 | QA Agent |
| human-testing → locked | 人测试完全通过 | **人** |
| human-testing → fixing | 人测试发现问题 | **人** → QA → 项目Agent |
| fixing → qa-testing | 项目Agent 修复完成 | 项目Agent → QA Agent |
| human-testing → blocked | 修复后仍不通过（第二次） | 主Agent → **人** |
| approved（回退）| 需求 md 有缺陷需修改 | **人** |

### 4.3 契约记忆：contracts.md

`docs/shared/contracts.md` 存储所有 `status: locked` 的阶段对外暴露的接口契约：

```markdown
# 接口契约

> 只有 status: locked 的阶段才能写入此文件。
> 后续阶段引用此文件中的接口。

## 阶段一（locked，2026-07-04）
- `middleware.Auth()` → 将 `userId` 注入 `gin.Context`
- `GET /api/v1/users/:id` → `{ code, data: { id, nickname, avatar } }`
- `GET /api/v1/videos/:id` → `{ code, data: { id, title, ... } }`
- `pkg/storage.Upload(file io.Reader, bucket, key string) error`
```

**写入规则**：
- 只有阶段达到 `locked` 状态（人测试完全通过），其公开接口才被写入 contracts.md
- 接口变更时，contracts.md 必须同步更新，且触发 7.3 回归检查

**已锁版接口保护**：
- `locked` 状态的接口签名不可变更，后续阶段只能新增，不能修改或删除
- 如果新阶段需要修改已有接口行为，必须新增接口（如 `/api/v2/...`），旧接口保留

---

## 五、标准工作流（单阶段开发循环）

### 5.0 阶段 0：需求文件准备

```
角色：人 + 主Agent
输入：产品功能清单
输出：docs/builder/requirements/{stage}.md（status: approved）

步骤：
1. 人起草阶段需求 md（本阶段所有功能组件）
2. 主Agent 检查自包含性、依赖 DAG、验收可验证性
3. 人审批 → status: draft → approved
4. Gate 1 完成，阶段进入可下派队列
```

### 5.1 阶段 1：主Agent 下派

```
角色：主Agent
输入：docs/builder/requirements/{stage}.md + docs/shared/contracts.md
输出：任务书（给项目Agent 的 prompt）

步骤：
1. 主Agent 读取需求 md（WHAT 视角）
2. 主Agent 查询 contracts.md，获取依赖接口签名
3. 主Agent 将需求重写为任务书（HOW 视角），覆盖阶段内所有组件
4. 主Agent 更新需求 md：status: approved → implementing
```

### 5.2 阶段 2：项目Agent 实现

```
角色：项目Agent
输入：任务书 + contracts.md + stage-(N-1)-legacy.md（阶段2起）
输出：代码 diff + 编译预检报告 + 交付摘要

步骤：
1. 项目Agent 按任务书实现阶段内所有组件
2. 完成后硬性执行编译预检：
   - go build ./...       # 零错误
   - go vet ./...         # 零警告
3. 预检不通过 → 自行修复后重试（最多 3 次，超限升级到主Agent）
4. 预检通过 → 编写交付摘要（见下方模板）
5. 更新需求 md：status: implementing → qa-testing
6. 产出交给 QA Agent
```

**交付摘要模板**（项目 Agent 交付时必写）：

```markdown
# 交付摘要：{阶段名}

## 本阶段新建
- internal/xxx/          — {用途}
- migrations/004_xxx.sql — {说明}

## 本阶段修改
- internal/video/handler.go — {改了什么、为什么改}
- internal/video/model.go   — {新增的字段}

## 不可碰的边界
- pkg/jwt/ — {原因}
- internal/middleware/auth.go — {原因}
- migrations/001_*.sql ~ 003_*.sql — 已有migration不可修改

## 代码风格约定
- {本阶段采用或延续的编码模式，后续阶段应保持一致}
```

**修复约束协议**：项目 Agent 在修复 bug 时必须遵守：

| 约束 | 说明 |
|------|------|
| **最小化 diff** | 单次修复 ≤ 20 行，超过需在修复报告中说明原因 |
| **不改无关代码** | 不得修改与问题无关的文件或函数 |
| **修复报告** | 列出"我改了哪几行、为什么只改这几行就够了" |
| **保留正确逻辑** | 原有经过验证的边界处理、错误路径、性能优化不得重写 |

### 5.3 阶段 3：QA Agent 测试

```
角色：QA Agent
输入：需求 md + 代码 diff + contracts.md
输出：QA 报告（结构化）

测试维度：
  - 正确性：逐条对照验收条件，构造 Happy Path + Edge Case + Error Case
  - 质量：按 Go 专项检查清单扫描
  - 回归：已有功能不被破坏

QA Agent 更新需求 md：status: qa-testing → human-testing
```

### 5.4 阶段 4：人测试验收

```
角色：人
输入：QA 报告 + 代码 diff
输出：锁版 or 修复指令

步骤：
1. 人读 QA 报告
2. 人自己动手测试（本地 Docker 环境）
3. 确认功能、质量、回归均通过 → 锁版
   - 更新需求 md：status: human-testing → locked
   - 主Agent 更新 contracts.md（提取本阶段公开接口）
   - 主Agent 基于项目Agent的交付摘要生成 stage-N-legacy.md
   - 进入阶段 N+1
4. 发现问题 → 写入测试文档（同一份 QA 报告，分界线以下）
   - 更新需求 md：status: human-testing → fixing
   - 交给 QA Agent 定位 → 回传项目 Agent 修复
```

### 5.5 修复循环与升级

```
项目 Agent 修复 → QA 再测 → 人再验
  │
  ├── 通过 → locked → 下一阶段
  │
  ├── 修复引入回归（回归锚点失败或原有功能被破坏）
  │     → 项目Agent 回退本次修复 diff
  │     → 重新提交最小化修复
  │     → 连续两次引入回归 → 升级到主Agent（任务书需重新评估）
  │
  └── 仍不通过 → 升级到主Agent
                    │
                    主Agent 分析根因：
                    ├── 需求 md 有缺陷 → 回退 draft，人修订
                    ├── 任务书有歧义 → 重写任务书，重新下派
                    └── 任务拆分不合理 → 人 + 主Agent 重新规划

---

## 六、QA 测试协议

### 6.1 正确性攻击

对需求 md 的**每条验收条件**，构造三个攻击场景：

| 攻击类型 | 说明 | 示例（视频上传） |
|---------|------|-----------------|
| **Happy Path** | 正常输入，期待正常输出 | 登录用户上传 50MB mp4 → 返回 video_id |
| **Edge Case** | 边界值输入 | 文件恰好 500MB、标题恰好 80 字符 |
| **Error Case** | 异常/恶意输入 | 未登录上传、超 500MB、非 mp4 格式 |

### 6.2 质量攻击

Go 专项固定检查清单：

| 类别 | 检查项 | 严重度 |
|------|--------|--------|
| **nil safety** | 所有指针解引用前是否判空？map 是否初始化？ | critical |
| **并发安全** | 共享变量是否有 mutex 保护？goroutine 是否有退出机制？ | critical |
| **资源泄漏** | file/conn/response body 是否在 defer 中关闭？ | major |
| **SQL 注入** | 是否使用参数化查询？有无字符串拼接 SQL？ | critical |
| **错误处理** | 所有 error 返回值是否被检查？panic 是否仅在 init 中使用？ | major |
| **事务一致性** | 多表写入是否在同一个事务中？有回滚处理？ | major |
| **整数安全** | 溢出检查？类型转换是否安全？ | minor |
| **日志安全** | 是否记录了敏感信息（token/password）？ | minor |

### 6.3 回归测试与锚点

**回归锚点**：阶段锁版时，QA 将已验证通过的关键行为记录为锚点，写入 `docs/qa/regression-anchors.md`：

```markdown
# 回归锚点

## 阶段1
- [ ] 用户注册：POST /register → 200 + JWT
- [ ] 未登录访问 /users/me → 401
- [ ] 上传 50MB mp4 → 200 + video_id
- [ ] 获取播放地址 → 200 + 有效预签名 URL
```

后续阶段每次 QA 测试时，锚点必须全部复验。**新修复如果导致锚点失败 → QA 直接驳回，不进入人测试。**

| 检查项 | 方式 |
|--------|------|
| 回归锚点全量复验 | 逐条执行锚点列表，任一失败 → 驳回 |
| 数据库迁移 | 确认新增 migration 不破坏已有表结构 |

### 6.4 QA + 人测试报告模板

同一份文件，分界线之上为 QA 产出，分界线之下为人追加：

```
┌─────────────────────────────────────────────────────┐
│  测试报告：{阶段名}                                   │
│  日期：{YYYY-MM-DD}                                  │
├─────────────────────────────────────────────────────┤
│  结论：[流程通过 / 需要修复]                          │
├─────────────────────────────────────────────────────┤
│  ▶ QA 测试结果：                                     │
│    - 正确性：Happy Path N/N | Edge N/N | Error N/N  │
│    - 质量：P0: N | P1: N | P2: N                    │
│    - 回归：已锁版接口 N/N 无损                       │
├─────────────────────────────────────────────────────┤
│  ▶ QA 发现问题：                                     │
│    1. [P0] ...                                      │
│    2. [P1] ...                                      │
│    3. [P2] ...                                      │
├─────────────────────────────────────────────────────┤
│  ▶ QA 已验证通过：                                   │
│    1. ...                                           │
│    2. ...                                           │
├─────────────────────────────────────────────────────┤
│  ▶ QA 详细日志：（折叠）                             │
│  ...                                                │
│                                                     │
│ ─────────────── QA / 人 分界线 ───────────────────  │
│                                                     │
│  ▶ 人测试结果：                                      │
│    结论：[通过 / 发现以下问题]                        │
│    测试环境：本地 Docker Compose                     │
│                                                     │
│  ▶ 人发现问题：                                      │
│    1. {问题描述}                                     │
│       复现步骤：                                     │
│       预期行为：                                     │
│       实际行为：                                     │
│    2. ...                                           │
│                                                     │
│  ▶ 人验证通过的功能：                                │
│    1. ...                                           │
│    2. ...                                           │
└─────────────────────────────────────────────────────┘
```

---

## 七、升级与恢复路径

### 7.1 编译预检失败

```
项目Agent 编码 → 编译预检失败
     │
     ├── 第 1 次失败：项目Agent 自行修复，重试
     ├── 第 2 次失败：项目Agent 自行修复，重试
     ├── 第 3 次失败：升级到主Agent
     │     ├── 任务书有歧义 → 主Agent 重写任务书，重新下派
     │     └── 复杂度太高 → 主Agent 建议调整阶段拆分，人决策
     └── 超过 3 次：强制升级到人
```

### 7.2 人测试不通过

```
人测试发现问题 → 写入测试文档 → QA定位 → 项目Agent修复 → QA再测 → 人再验
     │
     ├── 通过 → locked → 下一阶段
     │
     └── 仍不通过（第二次）→ 升级到主Agent
              │
              主Agent 分析根因：
              ├── 需求 md 有缺陷 → 回退 draft，人修订 → 重新 Gate 1
              ├── 任务书有歧义 → 重写任务书，重新下派项目Agent
              └── 阶段拆分不合理 → 人 + 主Agent 重新规划
```

### 7.3 修复引入回归

```
项目Agent 修复 bug → QA 复测
     │
     ├── 回归锚点全部通过 → 进入人测试
     │
     └── 回归锚点失败 → QA 驳回
            │
            项目Agent 回退本次修复 diff
            │
            重新提交最小化修复（≤20行，不改无关代码）
            │
            ├── 修复成功 → 继续流程
            │
            └── 再次引入回归 → 升级到主Agent
                  → 主Agent 重新评估任务书（复杂度过高？需拆分修复？）
                  → 人介入决策
```

### 7.4 已锁版阶段被新阶段破坏

```
阶段 N+1 开发中，QA 回归测试发现阶段 N 的回归锚点失败
     │
     ▼
QA 在报告中标注"回归问题"（指出哪个锚点失败）
     │
     ▼
项目Agent 修复（不得修改阶段 N 已锁定的接口签名，只修 bug）
     │
     ▼
QA 全量回归锚点验证 → 人确认阶段 N 功能恢复 → 阶段 N+1 继续
```

### 7.5 需求 md 冻结规则

- 需求 md 在 `approved` 之后**原则上冻结**
- 阶段进行中（implementing / qa-testing / human-testing）如需修改：
  1. status 回退到 `draft`
  2. 人在需求 md 中标注修改内容和原因
  3. 重新走 Gate 1 审批 → 全流程
- 已 `locked` 的阶段需求 md **不可修改**

---

## 八、工具系统

### 8.1 Skills 调用点

| 步骤 | 调用点 | Skill | 说明 |
|------|--------|-------|------|
| 5.2 项目Agent | 编码完成后 | `go build` + `go vet` | 硬性阻断 |
| 5.3 QA Agent | 正确性 + 回归 | `verify`（端到端验收驱动） | 内置 Skill |
| 5.3 QA Agent | 质量扫描 | `code-review`（correctness + security，关闭 style） | 内置 Skill |
| 5.4 人测试 | QA 问题定位 | `code-review`（聚焦人发现的问题点） | 内置 Skill |

### 8.2 项目自定义 Skill 预留

```json
// docs/go-qa.json（后续阶段定义）
{
  "name": "go-qa",
  "description": "Go 专项质量检查聚合命令",
  "steps": [
    "go vet ./...",
    "go test -race ./...",
    "errcheck ./..."
  ]
}
```

### 8.3 工具调用权限边界

| Agent | 可调用的工具 | 不可调用的工具 |
|-------|------------|--------------|
| 主Agent | 文件读写（docs/）、Agent 下发 | 代码修改（internal/、pkg/、cmd/） |
| 项目Agent | 文件读写（全项目）、Bash（go build/vet/test） | Agent 下发 |
| QA Agent | 文件读取（全项目）、Bash（go test -race）、Skill 调用、文件写入（测试报告） | Agent 下发、业务代码修改 |

---

## 九、反模式

### 9.1 什么时候不走 3 Agent 流程

| 场景 | 理由 |
|------|------|
| 修改配置文件（`config.yaml`） | 无业务逻辑变更 |
| 更新文档（`docs/`） | 无代码变更 |
| 修复拼写错误/注释 | 无行为变更 |
| 依赖版本升级（无 API 变更） | 仅 go.mod 变更，跑一次 `go build` 即可 |

判断标准：**改动不改变外部行为，跳过 Harness。**

### 9.2 不要在 approved 后修改需求 md 而不重走流程

需求 md 是 Agent 工作的"锚"。如果必须修改，回退 status 到 `draft`，重新走 Gate 1。

### 9.3 不要让 QA 验证未通过编译的代码

编译预检门是硬性阻断。未通过编译的代码进入 QA 等于浪费测试轮次。

### 9.4 不要跳过 Gate 1

未经人确认的需求 md，一个歧义会在实现和 QA 中被放大，修复成本远高于审批成本。

### 9.5 不要在人测试通过前开始下一阶段

人测试没通过就进入下一阶段，问题会跨阶段累积，修复成本指数增长。**版本锁定是硬约束。**

### 9.6 人测试不通过两次后不要再循环

两次人测不通过说明不是修 bug 能解决的，是需求 md 或任务书在根上有问题。升级到主 Agent，不要继续在修复循环里消耗。

---

## 十、快速上手

### 10.1 新建阶段操作清单

```
□ 1. 确定本阶段范围和边界
□ 2. 复制 docs/builder/requirements/_TEMPLATE.md → docs/requirements/{阶段名}.md
□ 3. 填写模板（功能概述、依赖、数据模型、API、业务规则、验收条件、非功能约束、本期不做）
□ 4. 主Agent 检查自包含性和依赖 DAG
□ 5. 人审批 → status: draft → approved（Gate 1）
□ 6. 主Agent 生成任务书 → 下派项目Agent
□ 7. 项目Agent 实现 → 编译预检 → 交付
□ 8. QA Agent 压测 → 产出 QA 报告
□ 9. 人测试验收 → 通过则锁版 / 不通过则写入测试文档
□ 10. 修复循环（如需）→ 人再验 → locked
□ 11. 主Agent 更新 contracts.md
```

### 10.2 需求 md 完整模板

见 `docs/builder/requirements/_TEMPLATE.md`。

---

> **对于个人开发者**：三体 Harness 的目标是用最少的 Agent 数量（3 个）实现最大化的质量保障。每个阶段产出独立可运行的版本，版本锁定后再进入下一阶段，让复杂度阶梯式增长。
