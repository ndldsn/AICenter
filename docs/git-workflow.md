---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: Lead
related: [AGENTS.md]
---

# Git Workflow

> **Git 工作流规范** — 项目的 Git 分支、提交、PR 策略。
>
> 适用对象：所有 Agent。

---

## 1. 分支模型

```
main            受保护，只接受 PR 合并，始终可发布
develop         集成分支（可选，小项目可省略）
feature/*       功能开发
fix/*           缺陷修复
refactor/*      重构
docs/*          文档
test/*          测试补充
agent/<id>/<task>  Agent 专属任务分支（推荐）
```

---

## 2. 分支命名规则

```
agent/<agent-id>/<type>-<task>-<short-desc>
```

示例：

```
agent/a1/feat-42-user-login
agent/a2/fix-57-token-expiry
agent/a3/refactor-auth-module
```

**规则**：

- 全小写
- 单词用连字符 `-` 分隔
- 必须包含 issue 编号或任务编号
- 禁止使用 `tmp`、`test123`、`wip` 等无意义名称

---

## 3. 分支所有权

- 一个分支同一时间只属于一个 Agent
- 禁止直接推送到 `main` / `develop`
- 禁止在别人的分支上直接提交（Review 建议通过 PR 评论传达）
- 分支创建后 48 小时无活动，Lead 可声明接管或删除

---

## 4. Commit 规范

### 4.1 格式

```
<type>(<scope>): <subject>

Agent: <agent-id>
Refs: <issue-id>
```

### 4.2 Type 枚举

```
feat:      新功能
fix:       缺陷修复
refactor:  重构（不改行为）
docs:      文档
test:      测试
chore:     构建/工具/依赖
perf:      性能优化
style:     格式调整（不改逻辑）
revert:    回滚
```

### 4.3 提交纪律

- 一个 Commit 只做一件事
- 单个 Commit 的 diff 超过 400 行时必须拆分
- 禁止混合：功能 + 格式化、重构 + 修复
- 禁止提交 `.env`、Secrets、密钥、构建产物
- Commit 前必须 `git status` 与 `git diff --staged` 自查

---

## 5. PR 规范

### 5.1 创建 PR

PR 标题格式：

```
[type] <issue-id>: <summary>
```

### 5.2 PR 约束

- 目标分支只能是 `develop` 或 `main`（依据仓库配置）
- diff 超过 1000 行的 PR 必须拆分成多个 PR
- PR 必须通过 CI（构建 + 测试 + lint）才能请求 Review
- 禁止 merge 冲突未解决的 PR

### 5.3 合并方式

- 统一使用 **Squash and Merge**
- 例外：跨分支同步使用普通 merge
- PR 合并后必须删除远端分支

---

## 6. 冲突处理

- 合并或 rebase 前必须先拉取目标分支最新代码
- 冲突双方 Agent 无法自行解决时，由 Lead 裁决
- 裁决原则：以更接近 main 的分支为准，后提交者负责适配
- 解决冲突后必须重新运行测试再更新 PR

---

## 7. 禁止的 Git 操作

- 禁止强制推送到 `main` / `develop`（`--force` 仅限个人 feature 分支且需 Lead 知晓）
- 禁止绕过 hooks（`--no-verify`）
- 禁止修改 git history 中他人的提交
- 禁止直接推送到受保护分支
