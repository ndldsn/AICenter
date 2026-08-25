---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: Lead
related: [AGENTS.md]
---

# docs/ 文档索引

> **文档路由器** — 告诉 Agent 有哪些规范、每个规范负责什么、哪些任务需要读取哪些规范。
>
> 本文件是 Agent 读取 AGENTS.md 后的第二步。

---

## 文档列表

| 文档 | 类型 | 强制加载 | 适用任务 | Owner |
|------|------|----------|----------|-------|
| `architecture.md` | 技术 | ❌ | 架构变更、新功能设计 | Architect |
| `technical-spec.md` | 技术 | ❌ | 涉及技术栈变更、API 设计 | Dev |
| `coding-standards.md` | 规范 | ❌ | 所有代码修改 | Lead |
| `testing.md` | 规范 | ❌ | 测试相关、验证相关 | QA |
| `git-workflow.md` | 规范 | ❌ | Git 操作、分支管理 | Lead |
| `security.md` | 安全 | ⚠️ | 认证、权限、安全相关 | Lead |
| `deployment.md` | 运维 | ❌ | Docker、CI/CD、部署 | DevOps |
| `agent-collaboration.md` | 协作 | ❌ | 多 Agent 协作、任务交接 | Lead |

---

## 强制加载规则（Always Required）

以下文档**每次任务必须加载**（已由 AGENTS.md 强制要求）：

1. **AGENTS.md** — 通用协作规范入口
2. **docs/README.md** — 本文件（文档索引）

---

## 条件加载规则（Conditional）

根据任务关键词，Agent 应自动加载以下文档：

### 任务分类与文档映射

```
┌─────────────────────────────────────────────────────────────┐
│                     任务关键词 → 文档                        │
├─────────────────────────────────────────────────────────────┤
│  typo / docs / readme    → coding-standards.md              │
│                                                             │
│  fix / bug / error       → testing.md                       │
│                           → coding-standards.md             │
│                                                             │
│  feat / feature / impl   → architecture.md                  │
│                           → technical-spec.md               │
│                           → coding-standards.md             │
│                                                             │
│  refactor                → architecture.md                  │
│                           → coding-standards.md             │
│                           → testing.md                      │
│                                                             │
│  auth / permission /     → security.md                      │
│  token / secret          → architecture.md                  │
│                                                             │
│  deploy / docker / ci    → deployment.md                    │
│                           → security.md                     │
│                                                             │
│  test / coverage         → testing.md                       │
│                                                             │
│  git / branch / merge    → git-workflow.md                  │
│                                                             │
│  agent / tool / runtime  → technical-spec.md                │
│                           → architecture.md                 │
│                           → agent-collaboration.md          │
└─────────────────────────────────────────────────────────────┘
```

---

## 文档优先级

当多个文档存在冲突时：

1. `security.md` > 其他文档（安全优先）
2. `architecture.md` > `coding-standards.md`（架构优先）
3. `AGENTS.md` > `docs/*.md`（总规范 > 专项规范）

---

## 文档状态

每个文档头部应包含元数据：

```markdown
---
status: active | deprecated | superseded
version: 1.x
last_reviewed: YYYY-MM-DD
owner: <role>
related: [<相关文档>]
---
```

- `active` — 当前有效
- `deprecated` — 已弃用，但仍可参考
- `superseded` — 已被新文档替代，应阅读新文档

---

## 文档维护责任

| 文档 | 维护者 | 更新触发条件 |
|------|--------|-------------|
| `architecture.md` | Architect | 架构变更时 |
| `technical-spec.md` | Dev | 技术栈/依赖变更时 |
| `coding-standards.md` | Lead | 编码规范变更时 |
| `testing.md` | QA | 测试策略变更时 |
| `git-workflow.md` | Lead | Git 策略变更时 |
| `security.md` | Lead | 安全策略变更时 |
| `deployment.md` | DevOps | 部署方式变更时 |
| `agent-collaboration.md` | Lead | 协作机制变更时 |

**规则**：涉及架构/规范的代码变更 PR，必须同步更新相关文档，否则 Lead 拒绝合并。
