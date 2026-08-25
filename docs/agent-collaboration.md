---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: Lead
related: [AGENTS.md, docs/README.md]
---

# Agent Handoff Protocol

> **Agent 交接协议** — 定义 Agent 之间如何传递任务、上下文和证据。
>
> 本文件是 Multi-Agent 协作的核心协议，所有 Agent 必须遵守。

---

## 1. 协议概述

### 1.1 设计目标

- 确保任务在不同 Agent 之间传递时信息不丢失
- 确保每个 Agent 都知道当前状态、下一步行动和约束条件
- 确保交接有迹可查（Git commit / PR comment）

### 1.2 适用场景

- Orchestrator → Architect（架构设计任务）
- Architect → Dev（实现任务）
- Dev → QA（测试验证任务）
- QA → Dev（测试失败，需要修复）
- Dev → Reviewer（代码审查）
- Reviewer → Dev（需要修改）
- Reviewer → Orchestrator（审查完成）
- DevOps → Orchestrator（部署验证）

---

## 2. Handoff 消息格式

每个 Agent 在完成任务并准备交接时，必须生成以下格式的 Handoff 消息：

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

---

## 3. 标准交接场景

### 3.1 Orchestrator → Architect

```yaml
from: orchestrator
to: architect
task_id: <issue-id>

summary:
  title: "<架构设计任务>"
  objective: "设计实现方案"
  status: PENDING

deliverables: []

context:
  relevant_files: []
  relevant_docs:
    - AGENTS.md
    - docs/README.md
    - docs/architecture.md
  decisions_made: []

risks: []

next_action:
  required_by: architect
  action: "阅读相关文档，输出架构设计方案"
  constraints:
    - "遵循分层架构原则"
    - "考虑 Security > Reliability > Maintainability 优先级"

evidence: {}
```

### 3.2 Architect → Dev

```yaml
from: architect
to: developer
task_id: <issue-id>

summary:
  title: "<实现任务>"
  objective: "按架构方案实现功能"
  status: PLANNED

deliverables:
  - type: doc
    path: "docs/architecture-design.md"
    description: "架构设计方案"
    verification:
      command: "Lead review"
      status: VERIFIED
      evidence: "Lead approved in PR comment"

context:
  relevant_files:
    - "src/<module>/index.ts"
  relevant_docs:
    - AGENTS.md
    - docs/README.md
    - docs/architecture.md
  decisions_made:
    - "使用 Service 层处理业务逻辑": "遵循 Handler→Service→Repository 分层"

risks:
  - level: low
    description: "新模块需要注册到 Router"
    mitigation: "参考 existing handler 注册方式"

next_action:
  required_by: developer
  action: "按架构方案实现代码，编写测试"
  constraints:
    - "遵循 coding-standards.md"
    - "新增代码必须有测试"

evidence:
  lint: { status: NOT_APPLICABLE, output: "N/A" }
  test: { status: NOT_APPLICABLE, output: "N/A" }
  build: { status: NOT_APPLICABLE, output: "N/A" }
```

### 3.3 Dev → QA

```yaml
from: developer
to: qa
task_id: <issue-id>

summary:
  title: "<实现任务>"
  objective: "代码已实现，等待测试验证"
  status: IMPLEMENTING → REVIEWING

deliverables:
  - type: code
    path: "agent/<agent-id>/<branch-name>"
    description: "功能实现代码"
    verification:
      command: "<test-command>"
      status: VERIFIED
      evidence: "All tests passed"

context:
  relevant_files:
    - "src/<module>/index.ts"
  relevant_docs:
    - AGENTS.md
    - docs/README.md
    - docs/testing.md
  decisions_made:
    - "使用 <技术选型>": "<理由>"

risks:
  - level: medium
    description: "<潜在风险>"
    mitigation: "<缓解措施>"

next_action:
  required_by: qa
  action: "运行测试套件，验证功能正确性"
  constraints:
    - "必须运行所有相关测试"
    - "报告测试结果和覆盖率"

evidence:
  lint:
    status: VERIFIED
    output: "no issues"
  test:
    status: VERIFIED
    output: "PASS X/X"
  build:
    status: VERIFIED
    output: "build success"
```

### 3.4 QA → Dev（测试失败）

```yaml
from: qa
to: developer
task_id: <issue-id>

summary:
  title: "<实现任务>"
  objective: "测试失败，需要修复"
  status: FAILED

deliverables: []

context:
  relevant_files:
    - "src/<module>/index.ts"
  relevant_docs:
    - AGENTS.md
    - docs/testing.md
  decisions_made: []

risks:
  - level: high
    description: "测试失败可能意味着功能实现有 bug"
    mitigation: "详细分析失败原因，不要简单删除测试"

next_action:
  required_by: developer
  action: "分析测试失败原因，修复 bug，重新测试"
  constraints:
    - "禁止简单删除失败的测试"
    - "必须分析根本原因"
    - "修复后重新运行所有测试"

evidence:
  test:
    status: FAILED
    output: "FAIL: TestXxx (file.ts:42)\nexpected: ..., got: ..."
```

### 3.5 Dev → Reviewer

```yaml
from: developer
to: reviewer
task_id: <issue-id>

summary:
  title: "<实现任务>"
  objective: "代码已完成，请求审查"
  status: REVIEWING

deliverables:
  - type: code
    path: "agent/<agent-id>/<branch-name>"
    description: "PR 待审查"
    verification:
      command: "<ci-command>"
      status: VERIFIED
      evidence: "CI passed"

context:
  relevant_files:
    - "src/<module>/index.ts"
    - "src/<module>/index.test.ts"
  relevant_docs:
    - AGENTS.md
    - docs/README.md
    - docs/coding-standards.md
    - docs/security.md
  decisions_made:
    - "<决策>: <理由>"

risks:
  - level: low
    description: "新增 dependency 需在 PR 中说明"
    mitigation: "已在 PR 描述中列出"

next_action:
  required_by: reviewer
  action: "审查代码质量和安全性"
  constraints:
    - "检查是否符合 coding-standards.md"
    - "检查是否有安全漏洞"
    - "检查测试覆盖是否充分"

evidence:
  lint: { status: VERIFIED, output: "no issues" }
  test: { status: VERIFIED, output: "PASS X/X" }
  build: { status: VERIFIED, output: "build success" }
```

### 3.6 Reviewer → Dev（REQUEST CHANGES）

```yaml
from: reviewer
to: developer
task_id: <issue-id>

summary:
  title: "<实现任务>"
  objective: "审查发现问题，需要修改"
  status: NEEDS_REVISION

deliverables: []

context:
  relevant_files:
    - "src/<module>/index.ts:142"
  relevant_docs:
    - AGENTS.md
    - docs/coding-standards.md
  decisions_made: []

risks:
  - level: medium
    description: "代码存在安全问题"
    mitigation: "必须修复后才能合并"

next_action:
  required_by: developer
  action: "逐条回应 review 意见，修复问题后重新请求 review"
  constraints:
    - "必须逐条回应每个 COMMENT"
    - "同意则修复，不同意则说明理由"
    - "禁止无视 review 意见直接重新请求审批"

evidence:
  review_comments:
    - file: "src/<module>/index.ts"
      line: 142
      comment: "<问题描述>"
      severity: high
```

### 3.7 Reviewer → Orchestrator（APPROVE）

```yaml
from: reviewer
to: orchestrator
task_id: <issue-id>

summary:
  title: "<实现任务>"
  objective: "审查通过，可以合并"
  status: APPROVED

deliverables:
  - type: code
    path: "agent/<agent-id>/<branch-name>"
    description: "审查通过的 PR"
    verification:
      command: "CI check"
      status: VERIFIED
      evidence: "All checks passed"

context:
  relevant_files:
    - "src/<module>/index.ts"
  relevant_docs:
    - AGENTS.md
    - docs/README.md
    - docs/coding-standards.md
    - docs/security.md
  decisions_made: []

risks: []

next_action:
  required_by: orchestrator
  action: "合并 PR，删除分支"
  constraints:
    - "使用 Squash and Merge"
    - "合并后删除远端分支"

evidence:
  review:
    status: APPROVED
    output: "LGTM, all requirements met"
```

---

## 4. 交接纪律

1. **每次交接必须有 Handoff 消息** — 禁止口头交接或隐含交接
2. **Handoff 消息必须落地在 Git 载体上** — PR 描述、Issue 评论、Commit message
3. **接收方必须确认收到** — 在 Issue 中评论确认
4. **禁止跳过环节** — 必须按流程传递，不能越级
5. **BLOCKED 状态必须说明** — 阻塞超过 2 小时必须更新 Handoff 消息

---

## 5. 与 State Machine 的关系

Handoff 协议与 Task State Machine 配合使用：

```
PENDING → ANALYZING → PLANNED → IMPLEMENTING → REVIEWING → APPROVED → COMPLETED
    ↑                                                        ↓
    └──────────────────── BLOCKED / FAILED / CANCELLED ──────┘
```

每个状态转换都对应一次 Handoff（或 Handoff 更新）。
