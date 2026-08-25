# Multi-Agent Collaboration Spec

通用多 Agent 协作规范（基于 GitHub 分支协作）。

适用于任意仓库、任意数量的 AI Agent（以下称 Agent）并行开发。

## 1. 总则

每个 Agent 在开始任务前必须完整阅读本文件。

阅读顺序（强制）：

```
AGENTS.md（本文件，协作规范）
→ PROJECT_SPEC.md（项目级技术规范，如存在）
→ 仓库 README / CONTRIBUTING.md
→ 任务相关 Issue
```

如果仓库根目录存在 PROJECT_SPEC.md，它定义了本项目的技术栈、架构分层、安全默认值等项目级约束，优先级高于 Agent 的通用习惯，每个 Agent 在写任何代码前必须先阅读并遵守。

Agent 之间不共享内存，唯一协作媒介是：

- Git 分支
- Commit
- Pull Request
- Issue
- Code Review

所有沟通必须落在上述载体上。

禁止 Agent 之间通过聊天记录、临时文件或口头约定传递决策。

## 2. 角色定义

每个 Agent 必须声明一个角色。角色决定权限边界。

| 角色 | 职责 | 权限 |
|------|------|------|
| Lead | 任务拆分、分支管理、合并 PR | 创建/合并分支、审批 PR |
| Dev | 实现功能、修复缺陷 | 创建 feature 分支、提交 PR |
| Reviewer | 代码审查 | 在 PR 上评论、approve / request changes |
| QA | 测试与验证 | 运行测试、在 PR 上标注测试结果 |
| Docs | 文档维护 | 修改 docs/、README、CHANGELOG |

一个 Agent 可以承担多个角色，但一次任务中只能以一个主角色行动。

### 2.1 角色声明

Agent 在第一个 Commit 或 PR 描述中必须声明：

```
Agent: <agent-id>
Role: <role>
Task: <issue-id or task-id>
```

## 3. 分支策略

### 3.1 分支模型

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

### 3.2 分支命名规则

```
agent/<agent-id>/<type>-<issue-or-task>-<short-desc>
```

示例：

```
agent/a1/feat-42-user-login
agent/a2/fix-57-token-expiry
agent/a3/refactor-auth-module
```

规则：

- 全小写
- 单词用连字符 `-` 分隔
- 必须包含 issue 编号或任务编号
- 禁止使用 `tmp`、`test123`、`wip` 等无意义名称

### 3.3 分支所有权

- 一个分支同一时间只属于一个 Agent。
- 禁止直接推送到 `main` / `develop`。
- 禁止在别人的分支上直接提交（Review 建议通过 PR 评论传达）。
- 分支创建后 48 小时无活动，Lead 可声明接管或删除。

### 3.4 分支生命周期

```
创建分支 → 开发 → 推送 → 创建 PR → Review → 合并 → 删除分支
```

PR 合并后必须删除远端分支，避免分支堆积。

## 4. 提交规范

### 4.1 Commit Message 格式

```
<type>(<scope>): <subject>

[optional body]

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

- 一个 Commit 只做一件事。
- 单个 Commit 的 diff 超过 400 行时必须拆分。
- 禁止混合：功能 + 格式化、重构 + 修复。
- 禁止提交 `.env`、Secrets、密钥、构建产物、`node_modules`。
- Commit 前必须 `git status` 与 `git diff --staged` 自查。

## 5. Pull Request 规范

### 5.1 创建 PR

PR 标题格式：

```
[type] <issue-id>: <summary>
```

PR 描述必须包含：

```markdown
## Agent
<agent-id> / <role>

## 变更内容
- 改了什么
- 为什么改

## 影响范围
- 涉及模块
- 是否影响 API / 数据库 / 配置

## 自测结果
- 已运行的测试及结果
- 未覆盖的风险点

## 关联
Refs #<issue-id>
```

### 5.2 PR 约束

- 目标分支只能是 `develop` 或 `main`（依据仓库配置）。
- diff 超过 1000 行的 PR 必须拆分成多个 PR。
- PR 必须通过 CI（构建 + 测试 + lint）才能请求 Review。
- 禁止 merge 冲突未解决的 PR。

### 5.3 Review 流程

```
提交 PR → CI 通过 → 至少 1 名 Reviewer 审查 → Lead 审批 → 合并
```

Reviewer 规则：

- 24 小时内必须响应（自动化 Agent 应立即响应）。
- Review 结论只能是以下三种之一：
  - `APPROVE`：可以合并
  - `REQUEST CHANGES`：列出必须修改的点
  - `COMMENT`：非阻塞建议
- Review 意见必须具体：指到文件和行号，说明问题和建议。

Dev 处理规则：

- 收到 `REQUEST CHANGES` 后，逐条回应。
- 同意则修复，不同意必须给出理由。
- 禁止无视 review 意见直接重新请求审批。

### 5.4 合并方式

统一使用 Squash and Merge，保持 main 历史整洁。

例外：跨分支同步（如 develop → main）使用普通 merge。

## 6. 冲突处理

- 合并或 rebase 前必须先拉取目标分支最新代码。
- 冲突双方 Agent 无法自行解决时，由 Lead 裁决。
- 裁决原则：以更接近 main 的分支为准，后提交者负责适配。
- 解决冲突后必须重新运行测试再更新 PR。

## 7. 任务分配与并行

### 7.1 任务认领

- 任务以 Issue 承载，每个 Issue 对应一个 Agent 分支。
- Agent 开始前在 Issue 下评论：`@<agent-id> claim`。
- 已被认领的 Issue 禁止其他 Agent 重复认领。
- 认领后 48 小时无 Commit 视为放弃，Lead 可释放任务。

### 7.2 任务拆分原则

Lead 拆分任务时必须保证：

- 任务之间低耦合，可并行开发。
- 接口先行：先合并接口定义/契约（API schema、类型定义），再并行实现。
- 单任务预期 diff 不超过 1000 行。

### 7.3 并行协作纪律

- 修改公共文件（接口定义、共享配置）前必须在 Issue 中声明。
- 公共文件的修改必须优先合并，其他 Agent 及时 rebase。
- 不确定归属的文件，先在 Issue 中询问 Lead。

## 8. 质量门禁

PR 合并前必须满足：

1. CI 全绿（编译通过、测试通过、lint 通过）
2. 新增代码有对应测试
3. 无硬编码密钥、密码、Token
4. 遵循仓库现有架构分层，不为图方便破坏边界
5. 无调试代码残留（console.log、fmt.Println；TODO 必须有关联 Issue）
6. 至少 1 个 approve

## 9. 安全红线

禁止：

- 提交任何密钥、凭证、`.env`、私钥
- 强制推送到 `main` / `develop`（`--force` 仅限个人 feature 分支且需 Lead 知晓）
- 绕过 CI、跳过 hooks（`--no-verify`）
- 修改 git history 中他人的提交
- 在代码中引入未审查的第三方依赖（新依赖必须在 PR 中说明理由）

## 10. Agent 行为准则

每次任务的标准流程：

```
读 AGENTS.md → 读 PROJECT_SPEC.md（如存在）→ 读 Issue → 检查现状（代码 + 分支）→ 认领
→ 创建分支 → 实现 → 自测 → 提 PR → 响应 Review → 合并 → 删除分支 → 更新 Issue
```

原则：

- 先读后写：修改前必须阅读相关代码，理解调用链。
- 最小改动：只为完成任务而修改，不顺手重构。
- 透明决策：所有设计取舍写入 PR 描述，不留隐性决定。
- 及时同步：每日至少推送一次进度（Commit 或 Issue 评论）。
- 失败即报：卡住超过 2 小时必须在 Issue 中说明阻塞点，不要静默死磕。

## 11. 争议解决

优先级从高到低：

1. 本文件
2. 项目级技术规范 PROJECT_SPEC.md（仅限项目技术约束）
3. 仓库 CONTRIBUTING.md / README
4. Lead 裁决
5. 代码现状（已合并的实现即为既定事实，除非 Lead 判定需要重构）

任何对本规范本身的修改，必须通过修改本文件的 PR 由 Lead 审批后生效。

## 12. 快速检查清单

Agent 在提交 PR 前自查：

```
[ ] 已阅读 AGENTS.md 和 PROJECT_SPEC.md（如存在）
[ ] 分支命名符合 agent/<id>/<type>-<issue>-<desc>
[ ] Commit message 符合 type(scope): subject 格式
[ ] diff 已自查，无调试代码、无密钥
[ ] 本地测试通过
[ ] CI 全绿
[ ] PR 描述完整（Agent / 变更 / 影响 / 自测 / 关联）
[ ] 已 rebase 目标分支最新代码，无冲突
```