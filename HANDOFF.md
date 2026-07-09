# B-B 项目交接文档

## 项目概述

B 站风格视频平台，Go + Gin + Nuxt 3 全栈项目。四个阶段全部完成。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.21+ / Gin / GORM / MySQL 8.0 / Redis 7 / MinIO / RabbitMQ |
| 前端 | Nuxt 4 / Vue 3 / TypeScript / Tailwind CSS / pnpm |
| 基础设施 | Docker Compose（MySQL + Redis + MinIO + RabbitMQ） |

## 当前进度

| 阶段 | 状态 |
|------|------|
| 阶段一：核心骨架 | 🔒 locked |
| 阶段二：内容消费 | 🔒 locked |
| 阶段三：社区互动 | 🔒 locked |
| 阶段四：媒体处理 | 🔍 qa-testing（待验收） |

## 快速启动

```bash
docker compose up -d mysql redis minio rabbitmq
cd backend && go run cmd/server/main.go
cd frontend && pnpm dev
```

## 关键目录

```
docs/
├── README.md              ← 入口索引
├── blueprints.md           ← 建设蓝图
├── rules.md                ← 规则手册 + 累计规则
├── harness.md              ← Harness 三体 Agent 协议
├── stage-01-core.md        ← 阶段一需求（locked）
├── stage-02-discovery.md   ← 阶段二需求（locked）
├── stage-03-community.md   ← 阶段三需求（locked）
├── stage-04-media.md       ← 阶段四需求（qa-testing）
├── stage-0*-legacy.md      ← 各阶段遗留文档
├── stage-0*-delivery-summary.md ← 各阶段交付摘要
└── requirements/
    ├── contracts.md        ← 58 个 locked 接口契约
    └── regression-anchors.md ← 回归锚点
```

## 阶段四待验收项

- [ ] 上传视频 → 转码任务创建
- [ ] 清晰度选择器 + 切换
- [ ] HLS m3u8 生成
- [ ] 自动封面（无用户封面时截帧）
- [ ] Admin Dashboard（统计卡片+用户表格）
- [ ] RabbitMQ 降级（连不上时 goroutine 直调）
- [ ] ffmpeg 降级（不可用时跳过不阻塞播放）

## 研发规范

### 后端
- handler → service → repository 三层分离
- ctx 第一参数，SQL 参数化，错误 %w 传递
- JSON snake_case
- 58 个 locked 接口不可修改签名

### 前端
- `<script setup lang="ts">` + strict TypeScript
- API 经 useApi（upload SSE 除外）
- Tailwind + CSS 变量 `var(--color-*)`
- 暗/亮双模切换

## 已安装 Skills

- `design-taste-frontend` — UI 品味优化（v2）
- `senior-fullstack` — 项目 Agent 全栈规范
- `senior-qa` — QA Agent 测试规范

## Harness 流程

```
主 Agent 读 docs/ → Gate 1 审批需求 → 下派项目 Agent → 读 diff 写摘要
→ 下派 QA Agent → 分析报告 → 人验收 → 锁版 → 写 legacy + 更新 contracts
```

主 Agent 只读 docs/，不直接改代码。代码修改都通过 Agent 下派。
