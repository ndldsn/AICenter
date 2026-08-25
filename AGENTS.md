# Agent Instruction Entry Point / Agent Constitution

> **通用多 Agent 协作规范** — 本文件是 Agent 参与任何项目的第一个入口。
>
> 本项目采用分层指令架构，本文件仅定义通用协作框架。具体项目约束由 docs/ 下的专项文档定义。

---

## 1. 总则

### 1.1 设计目标

本规范体系旨在建立一套可长期使用的 Multi-Agent 软件工程项目规范，满足：

- **通用性** — 不绑定特定语言、框架或 Agent CLI
- **可扩展性** — 未来增加 Agent、服务、仓库时无需重写体系
- **低上下文成本** — Agent 不每次读取几十个文档
- **强约束** — 关键规则是 MUST，不是建议
- **最小权限** — Agent 只获得完成任务所需的能力
- **可验证** — Agent 工作结果必须有客观证据
- **可审计** — 能知道谁、何时、为何、修改了什么
- **防冲突** — 多 Agent 同时工作不互相覆盖
- **防漂移** — 代码和文档长期保持一致
- **人类可控** — 高风险操作人类始终拥有最终控制权

### 1.2 协作媒介

Agent 之间不共享内存。唯一协作媒介：

- Git 分支 / Commit / PR / Issue
- Code Review 评论

所有沟通必须落在上述载体上。禁止通过聊天记录、临时文件或口头约定传递决策。

### 1.3 强制阅读顺序

每个 Agent 开始任务前，必须按以下顺序读取：

```
① AGENTS.md（本文件）← 你正在读
   ↓
② docs/README.md ← 文档索引 + 加载策略
   ↓
③ 任务相关 Issue / PR
   ↓
④ 根据 docs/README.md 的加载规则，读取适用的专项文档
```

### 1.4 角色声明

Agent 在第一个 Commit 或 PR 描述中必须声明：

```
Agent: <agent-id>
Role: <role>
Task: <issue-id or task-id>
```

---

## 2. Rule Precedence Model（指令优先级模型）

当不同来源的规则发生冲突时，按以下优先级裁决：

```
L0  System / Platform Rules          （平台注入，最高优先级，不可覆盖）
  ↓
L1  User Instructions                （用户当次指令）
  ↓
L2  Safety-Critical Rules            （安全红线，跨层级生效）
  ↓
L3  Project AGENTS.md                （本文件，通用协作规范）
  ↓
L4  docs/ 专项文档                   （按加载规则读取的规范）
  ↓
L5  Agent Role Definition            （.agents/*.md，角色能力边界）
  ↓
L6  Task-Specific Instructions       （Issue / PR 中的具体任务说明）
```

### 2.1 关键原则

1. **L0 / L2 不可覆盖** — 平台安全规则和全局安全红线永远生效，任何角色、任务、用户指令不能 override
2. **L1 vs L3-L6** — 用户当次指令优先于项目规范，但用户不能指令违反 L2 安全红线
3. **L4 vs L5** — 专项文档（如 coding-standards.md）优先于通用角色定义
4. **L6 最具体** — 任务级指令优先级最低，仅在上述层面无规定时生效
5. **同级冲突仲裁** — 由 Lead 裁决；无法裁决时以"更保守/更安全"为准
6. **本文件修改权** — 对本规范本身的修改，必须通过修改本文件的 PR 由 Lead 审批后生效

---

## 3. Context-Aware Documentation Loading（上下文感知文档加载）

### 3.1 Always Required（每次任务必须加载）

```
AGENTS.md                          ← 本文件
docs/README.md                     ← 文档索引和加载规则
```

### 3.2 Conditional Loading（按任务类型加载）

Agent 应根据任务关键词自动判断需要加载的文档：

| 任务关键词 | 需加载文档 |
|-----------|-----------|
| typo / docs / readme | `docs/coding-standards.md` |
| fix / bug / error | `docs/testing.md`, `docs/coding-standards.md` |
| feat / feature / implement | `docs/architecture.md`, `docs/technical-spec.md`, `docs/coding-standards.md` |
| refactor | `docs/architecture.md`, `docs/coding-standards.md`, `docs/testing.md` |
| auth / permission / token / secret | `docs/security.md`, `docs/architecture.md` |
| deploy / docker / ci-cd | `docs/deployment.md`, `docs/security.md` |
| test / coverage | `docs/testing.md` |
| git / branch / merge | `docs/git-workflow.md` |
| agent / tool / runtime | `docs/technical-spec.md`, `docs/architecture.md` |

### 3.3 加载规则

1. **最小加载** — 只加载与任务相关的文档，不读取全部
2. **逐级深入** — 先读 `docs/README.md` 了解索引，再按需加载专项文档
3. **记录加载** — 在 PR 描述中列出已阅读的文档（让 Reviewer 可追溯）

### 3.4 防止过度加载

- 单个任务的加载文档不超过 5 个
- 如果任务涉及多个领域，先完成核心文档加载，再按需补充
- 禁止将"学习项目背景"作为加载全部文档的理由

---

## 4. Agent Roles & Capability Model（角色与能力模型）

### 4.1 角色定义

| 角色 | 职责 | 写权限 | Git 权限 | 执行权限 |
|------|------|--------|----------|----------|
| **Lead** | 任务拆分、分支管理、PR 合并、争议裁决 | 配置、文档 | 全量 | L0-L2 |
| **Dev** | 实现功能、修复缺陷 | 源码、测试 | feature 分支 | L0-L1 |
| **Reviewer** | 代码审查、质量把关 | 仅评论 | 只读 | 无 |
| **QA** | 测试编写与运行、验证 | 测试代码 | test 分支 | L0-L1 (test) |
| **Docs** | 文档维护、规范更新 | 文档 | docs 分支 | 无 |

> **一个 Agent 可承担多角色，但一次任务只能以一个主角色行动。**

### 4.2 Capability Matrix（能力矩阵）

| 能力 | Lead | Dev | Reviewer | QA | Docs |
|------|------|-----|----------|-----|------|
| read | ✅ | ✅ | ✅ | ✅ | ✅ |
| search | ✅ | ✅ | ✅ | ✅ | ✅ |
| analyze | ✅ | ✅ | ✅ | ✅ | ✅ |
| plan | ✅ | ✅ | — | — | — |
| edit | ✅ | ✅ | — | ✅ (tests) | ✅ (docs) |
| execute | ✅ (L0-L2) | ✅ (L0-L1) | — | ✅ (test) | — |
| test | ✅ | ✅ | — | ✅ | — |
| git | ✅ (full) | ✅ (feature) | — | ✅ (test) | ✅ (docs) |
| deploy | ✅ (L3+) | — | — | — | — |
| approve | ✅ | — | ✅ | — | — |
| delegate | ✅ | — | — | — | — |

### 4.3 Approval Boundary（审批边界）

| 操作类型 | 审批要求 |
|----------|----------|
| 读取源码 | 无需审批 |
| 修改源码 | 自主（L1） |
| 安装依赖 | 自主（L2），需在 PR 说明 |
| 数据库迁移 | Lead 审批（L3） |
| 部署到 Staging | Lead 审批（L3） |
| 部署到 Production | Human + Lead 双审批（L4） |
| 删除生产数据 | Human 审批（L4） |
| 修改本规范文件 | Lead 审批 |

### 4.4 Escalation Mechanism（升级机制）

当 Agent 遇到以下情况时，必须升级：

| 场景 | 升级对象 |
|------|----------|
| 超过 2 小时无法推进 | Lead |
| 发现安全漏洞 | Lead + 相关角色 |
| 发现架构违规 | Lead + Architect |
| 与其他 Agent 冲突无法解决 | Lead 裁决 |
| 需要 L3+ 操作权限 | Lead 审批 |

---

## 5. Risk Levels（风险等级）

| 等级 | 含义 | 允许操作 | 审批要求 |
|------|------|----------|----------|
| **L0 — Safe** | 只读、无副作用 | read, search | 无需审批 |
| **L1 — Low Risk** | 修改非生产代码 | edit source, run tests | 自动通过 |
| **L2 — Medium Risk** | 安装依赖、修改配置 | execute dev commands | Dev 自主 |
| **L3 — High Risk** | 数据库迁移、部署 | migrate, deploy staging | Lead 审批 |
| **L4 — Critical** | 生产操作、数据删除 | deploy prod, delete data | Human + Lead 双审批 |

### 5.1 常见操作的风险分级参考

| 操作 | 风险等级 |
|------|----------|
| 读取源码 | L0 |
| 搜索文件 | L0 |
| 修改源码 | L1 |
| 运行测试 | L1 |
| 格式化代码 | L1 |
| 安装开发依赖 | L2 |
| 修改配置文件 | L2 |
| 数据库迁移 | L3 |
| 部署到 Staging | L3 |
| 部署到 Production | L4 |
| 删除生产数据 | L4 |
| 修改安全配置 | L4 |

---

## 6. Task State Machine（任务状态机）

```
PENDING
  ↓ (Lead 分配 / Agent claim)
ANALYZING
  ↓ (Agent 完成分析)
PLANNED
  ↓ (Agent 开始实现)
IMPLEMENTING
  ↓ (实现完成，提交 PR)
REVIEWING
  ↓ (Reviewer APPROVE)
APPROVED
  ↓ (Lead 合并)
COMPLETED
```

**异常状态：**

```
BLOCKED        → 卡在某步骤 > 2 小时，必须在 Issue 中说明
FAILED         → 测试/CI 失败，回到 IMPLEMENTING
NEEDS_REVISION → Reviewer REQUEST CHANGES，回到 IMPLEMENTING
CANCELLED      → Lead 或用户终止
```

**状态转换规则：**

| 从状态 | 到状态 | 条件 | 负责 Agent |
|--------|--------|------|-----------|
| PENDING | ANALYZING | Issue 被认领 | Dev |
| ANALYZING | PLANNED | 分析完成，方案明确 | Dev |
| PLANNED | IMPLEMENTING | 方案确认，开始编码 | Dev |
| IMPLEMENTING | REVIEWING | PR 创建，CI 通过 | Dev → Reviewer |
| REVIEWING | NEEDS_REVISION | Reviewer 提出修改意见 | Reviewer |
| REVIEWING | APPROVED | Reviewer APPROVE | Reviewer |
| NEEDS_REVISION | REVIEWING | Dev 修复后重新请求 review | Dev → Reviewer |
| APPROVED | COMPLETED | Lead 合并 PR | Lead |
| 任何状态 | BLOCKED | 阻塞 > 2 小时 | 当前 Agent |
| 任何状态 | FAILED | CI/测试失败 | CI/QA |
| 任何状态 | CANCELLED | Lead 或用户终止 | Lead |

---

## 7. Evidence-Based Agent Workflow（基于证据的工作流）

Agent 不能只说"已完成"，必须提供可验证证据。

### 7.1 验证状态枚举

```
VERIFIED     — 已通过验证
NOT_RUN      — 未执行（需说明原因）
FAILED       — 执行失败（需附错误信息）
BLOCKED      — 被阻塞（需说明阻塞原因）
NOT_APPLICABLE — 不适用于当前任务
```

### 7.2 验证清单模板

Agent 在 PR 描述或最终报告中必须填写：

```yaml
validation:
  lint:
    command: "<lint命令>"
    status: VERIFIED | NOT_RUN | FAILED | BLOCKED | NOT_APPLICABLE
    evidence: "<输出摘要>"
  
  test:
    command: "<test命令>"
    status: VERIFIED | NOT_RUN | FAILED | BLOCKED | NOT_APPLICABLE
    evidence: "<通过率 / 失败用例>"
  
  build:
    command: "<build命令>"
    status: VERIFIED | NOT_RUN | FAILED | BLOCKED | NOT_APPLICABLE
    evidence: "<编译结果>"
  
  manual_check:
    - description: "<检查项>"
      status: VERIFIED | NOT_APPLICABLE
      evidence: "<说明>"
```

### 7.3 验证要求

- **VERIFIED** — 必须附上命令输出或测试结果
- **NOT_RUN** — 必须说明为什么未执行
- **FAILED** — 必须附上错误信息和分析
- **BLOCKED** — 必须说明阻塞原因和预计解决时间
- **NOT_APPLICABLE** — 必须说明为什么不适用于当前任务

---

## 8. Agent Security Boundary（Agent 安全边界）

### 8.1 安全红线（L2 级别，不可覆盖）

**永久禁止：**

- 提交任何密钥、凭证、`.env`、私钥
- 强制推送到受保护分支（`--force` 仅限个人 feature 分支且需 Lead 知晓）
- 绕过 CI、跳过 hooks（`--no-verify`）
- 修改 git history 中他人的提交
- 在代码中引入未审查的第三方依赖
- 让 AI 默认 root 执行
- 绕过权限系统
- 绕过 Approval 流程
- 静默执行高风险命令
- 在代码中硬编码敏感信息

### 8.2 默认权限模型

```
READ     = ALLOW
WRITE    = ASK
DELETE   = ASK
CRITICAL = DENY
```

高风险操作必须人工审批。

### 8.3 Audit Log

所有重要操作必须记录 Audit Log。不得记录：

- Password
- API Key
- Private Key
- Token

### 8.4 Agent 安全原则

- **Agent 默认只读** — Agent 必须先观察，再分析，再计划
- **工具调用必须** — 禁止让 LLM 直接生成命令并自动执行
- **命令执行必须经过审批链** — Tool → Permission → Risk Assessment → Approval → Execution → Verification
- **禁止绕过权限系统** — 任何权限绕过尝试必须触发安全警报

---

## 9. Git Work Isolation Strategy（Git 隔离策略）

### 9.1 分支模型

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

### 9.2 分支命名规则

```
agent/<agent-id>/<type>-<task>-<short-desc>
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

### 9.3 分支所有权

- 一个分支同一时间只属于一个 Agent
- 禁止直接推送到 `main` / `develop`
- 禁止在别人的分支上直接提交（Review 建议通过 PR 评论传达）
- 分支创建后 48 小时无活动，Lead 可声明接管或删除

### 9.4 冲突处理

- 合并或 rebase 前必须先拉取目标分支最新代码
- 冲突双方 Agent 无法自行解决时，由 Lead 裁决
- 裁决原则：以更接近 main 的分支为准，后提交者负责适配
- 解决冲突后必须重新运行测试再更新 PR

### 9.5 合并策略

- 统一使用 **Squash and Merge**
- 例外：跨分支同步使用普通 merge
- PR 合并后必须删除远端分支

---

## 10. Commit & PR Standards（提交与 PR 规范）

### 10.1 Commit Message 格式

```
<type>(<scope>): <subject>

[optional body]

Agent: <agent-id>
Refs: <issue-id>
```

### 10.2 Type 枚举

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

### 10.3 Commit 纪律

- 一个 Commit 只做一件事
- 单个 Commit 的 diff 超过 400 行时必须拆分
- 禁止混合：功能 + 格式化、重构 + 修复
- 禁止提交 `.env`、Secrets、密钥、构建产物
- Commit 前必须 `git status` 与 `git diff --staged` 自查

### 10.4 PR 规范

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

## 已读文档
- AGENTS.md
- docs/README.md
- docs/<relevant-doc>.md

## 自测结果
- 已运行的测试及结果
- 未覆盖的风险点

## 关联
Refs #<issue-id>
```

### 10.5 PR 约束

- 目标分支只能是 `develop` 或 `main`（依据仓库配置）
- diff 超过 1000 行的 PR 必须拆分成多个 PR
- PR 必须通过 CI（构建 + 测试 + lint）才能请求 Review
- 禁止 merge 冲突未解决的 PR

---

## 11. Review Process（审查流程）

```
提交 PR → CI 通过 → 至少 1 名 Reviewer 审查 → Lead 审批 → 合并
```

### 11.1 Reviewer 规则

- 24 小时内必须响应（自动化 Agent 应立即响应）
- Review 结论只能是以下三种之一：
  - `APPROVE`：可以合并
  - `REQUEST CHANGES`：列出必须修改的点
  - `COMMENT`：非阻塞建议
- Review 意见必须具体：指到文件和行号，说明问题和建议

### 11.2 Dev 处理规则

- 收到 `REQUEST CHANGES` 后，逐条回应
- 同意则修复，不同意必须给出理由
- 禁止无视 review 意见直接重新请求审批

---

## 12. Quality Gates（质量门禁）

PR 合并前必须满足：

1. CI 全绿（编译通过、测试通过、lint 通过）
2. 新增代码有对应测试
3. 无硬编码密钥、密码、Token
4. 遵循仓库现有架构分层，不为图方便破坏边界
5. 无调试代码残留（TODO 必须有关联 Issue）
6. 至少 1 个 approve

---

## 13. Agent Behavior Guidelines（Agent 行为准则）

### 13.1 标准流程

```
读 AGENTS.md → 读 docs/README.md
→ 按加载规则读适用文档 → 检查现状（代码 + 分支）
→ 认领任务 → 创建分支 → 实现 → 自测
→ 提 PR → 响应 Review → 合并 → 删除分支 → 更新 Issue
```

### 13.2 核心原则

- **先读后写**：修改前必须阅读相关代码，理解调用链
- **最小改动**：只为完成任务而修改，不顺手重构
- **透明决策**：所有设计取舍写入 PR 描述，不留隐性决定
- **及时同步**：每日至少推送一次进度（Commit 或 Issue 评论）
- **失败即报**：卡住超过 2 小时必须在 Issue 中说明阻塞点，不要静默死磕

### 13.3 代码修改纪律

修改代码前必须：

1. 阅读相关代码
2. 理解现有架构
3. 找到真正的调用链
4. 评估影响范围
5. 再修改

禁止：为了修复一个问题而大范围重构。

---

## 14. Agent Handoff Protocol（Agent 交接协议）

### 14.1 协议概述

每次 Agent 完成任务并准备交接时，必须生成 Handoff 消息，确保：

- 任务信息不丢失
- 接收方知道当前状态、下一步行动和约束
- 交接有迹可查（Git commit / PR comment / Issue 评论）

### 14.2 Handoff 消息格式

```yaml
handoff:
  from: <agent-id>
  to: <next-agent-id>
  task_id: <issue-id or task-id>
  timestamp: "YYYY-MM-DDTHH:MM:SSZ"
  
  summary:
    title: "<任务标题>"
    objective: "<完成的目标>"
    status: COMPLETED | BLOCKED | FAILED
  
  deliverables:
    - type: code | doc | test | config
      path: "<文件路径或分支名>"
      description: "<描述>"
      verification:
        command: "<验证命令>"
        status: VERIFIED | NOT_RUN | FAILED
        evidence: "<输出摘要>"
  
  context:
    relevant_files:
      - "<文件路径>"
    relevant_docs:
      - "<文档路径>"
    decisions_made:
      - "<决策1>: 理由"
      - "<决策2>: 理由"
  
  risks:
    - level: low | medium | high | critical
      description: "<风险描述>"
      mitigation: "<缓解措施>"
  
  next_action:
    required_by: <next-agent-id>
    action: "<下一步需要做什么>"
    constraints:
      - "<约束条件1>"
      - "<约束条件2>"
  
  evidence:
    lint:
      status: VERIFIED | NOT_RUN | FAILED | BLOCKED | NOT_APPLICABLE
      output: "<命令输出>"
    test:
      status: VERIFIED | NOT_RUN | FAILED | BLOCKED | NOT_APPLICABLE
      output: "<测试输出>"
    build:
      status: VERIFIED | NOT_RUN | FAILED | BLOCKED | NOT_APPLICABLE
      output: "<编译输出>"
```

### 14.3 交接纪律

1. 每次交接必须有 Handoff 消息
2. Handoff 消息必须落地在 Git 载体上（PR 描述、Issue 评论、Commit message）
3. 接收方必须确认收到
4. 禁止跳过环节
5. BLOCKED 状态必须说明阻塞原因和预计解决时间

---

## 15. Documentation Governance（文档治理）

### 15.1 文档分层

| 层级 | 文件 | 性质 | 修改权限 |
|------|------|------|----------|
| L1 | AGENTS.md | 通用规范 | Lead |
| L2 | docs/*.md | 项目专项文档 | 各角色按职责维护 |
| L3 | .agents/*.md | 角色定义 | Lead |

### 15.2 文档版本标记

每个文档头部应包含：

```markdown
---
status: active | deprecated | superseded
version: 1.x
last_reviewed: YYYY-MM-DD
owner: <role>
related: [<相关文档>]
---
```

| 状态 | 含义 |
|------|------|
| `active` | 当前有效 |
| `deprecated` | 已弃用，仍可参考 |
| `superseded` | 已被新文档替代 |

### 15.3 防漂移机制

- 架构变更时，必须同步更新 `docs/architecture.md`
- 技术栈变更时，必须同步更新 `docs/technical-spec.md`
- Lead 在合并涉及架构/规范的 PR 时，必须验证相关文档已更新
- 定期（每月）Review 文档 with-code alignment

### 15.4 文档维护责任

| 文档 | 维护者 | 更新触发条件 |
|------|--------|-------------|
| AGENTS.md | Lead | 规范变更时 |
| docs/architecture.md | Architect | 架构变更时 |
| docs/technical-spec.md | Dev | 技术栈变更时 |
| docs/coding-standards.md | Lead | 编码规范变更时 |
| docs/testing.md | QA | 测试策略变更时 |
| docs/git-workflow.md | Lead | Git 策略变更时 |
| docs/security.md | Lead | 安全策略变更时 |
| docs/deployment.md | DevOps | 部署方式变更时 |
| docs/agent-collaboration.md | Lead | 协作机制变更时 |

---

## 16. Multi-Agent Collaboration（多 Agent 协作）

### 16.1 任务认领

- 任务以 Issue 承载，每个 Issue 对应一个 Agent 分支
- Agent 开始前在 Issue 下评论：`@<agent-id> claim`
- 已被认领的 Issue 禁止其他 Agent 重复认领
- 认领后 48 小时无 Commit 视为放弃，Lead 可释放任务

### 16.2 任务拆分原则

Lead 拆分任务时必须保证：

- 任务之间低耦合，可并行开发
- 接口先行：先合并接口定义/契约（API schema、类型定义），再并行实现
- 单任务预期 diff 不超过 1000 行

### 16.3 并行协作纪律

- 修改公共文件（接口定义、共享配置）前必须在 Issue 中声明
- 公共文件的修改必须优先合并，其他 Agent 及时 rebase
- 不确定归属的文件，先在 Issue 中询问 Lead

---

## 17. Dispute Resolution（争议解决）

优先级从高到低：

1. 本文件（AGENTS.md）
2. docs/ 专项文档（按冲突类型选择适用的）
3. 仓库 CONTRIBUTING.md / README
4. Lead 裁决
5. 代码现状（已合并的实现即为既定事实，除非 Lead 判定需要重构）

---

## 18. Quick Checklist（快速检查清单）

Agent 在提交 PR 前自查：

```
[ ] 已阅读 AGENTS.md
[ ] 已按 docs/README.md 加载适用文档
[ ] 已声明 Agent / Role / Task
[ ] 分支命名符合 agent/<id>/<type>-<task>-<desc>
[ ] Commit message 格式正确
[ ] diff 已自查，无调试代码、无密钥
[ ] 本地测试通过
[ ] 验证清单已填写（validation section）
[ ] PR 描述完整
[ ] 已 rebase 目标分支最新代码
```

---

## 19. File Relationships（文件关系）

```
AGENTS.md（本文件 — 通用规范入口）
  ├── 引用 docs/README.md（文档索引 + 加载策略）
  │     ├── docs/architecture.md        ← 系统架构 + 技术栈 + 架构约束
  │     ├── docs/technical-spec.md      ← 技术实现细节（接口、配置、协议）
  │     ├── docs/coding-standards.md    ← 编码规范
  │     ├── docs/testing.md             ← 测试规范
  │     ├── docs/git-workflow.md        ← Git 工作流
  │     ├── docs/security.md            ← 安全规范
  │     ├── docs/deployment.md          ← 部署规范
  │     └── docs/agent-collaboration.md ← Agent 交接协议
  └── 引用 .agents/<role>.md（角色能力定义）
        ├── .agents/orchestrator.md
        ├── .agents/developer.md
        ├── .agents/reviewer.md
        ├── .agents/qa.md
        ├── .agents/docs.md
        └── .agents/devops.md
```

本文件是通用规范入口，不重复包含详细项目规范。详细规范在各对应文件中维护。

---

## 20. Vendor Neutrality（供应商中立性）

本规范体系设计为 Vendor-Neutral：

- 不绑定特定 Agent CLI（Claude Code、Codex、Gemini CLI 等均可使用）
- 不绑定特定编程语言或框架
- 不绑定特定 CI/CD 平台
- 使用纯 Markdown + YAML Frontmatter，兼容所有 Agent 工具

各 Agent CLI 可能有自己的扩展机制（如 settings.json、hook 配置），这些属于平台层（L0），不影响本规范的通用性。
