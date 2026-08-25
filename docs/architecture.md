---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: Architect
related: [AGENTS.md, docs/technical-spec.md]
---

# Architecture Reference

> **架构参考** — 项目架构要点速查。
>
> 本文件提取核心架构约束和系统定位，供 Agent 快速参考。
>
> 详细技术实现参见 `docs/technical-spec.md`。

---

## 1. 项目定位

本项目是一个 **生产级 AI Server Control Plane**。

不要把项目实现成 Demo。

必须优先考虑：

```
Security > Reliability > Maintainability > Observability > Extensibility
```

---

## 2. 技术栈

| 层次 | 技术 |
|------|------|
| Frontend | React + TypeScript + Vite + Arco Design |
| Backend | Go + Gin |
| Database | SQLite (dev) / PostgreSQL (prod) |
| Real-time | WebSocket |
| AI | OpenAI Compatible API + Anthropic + Gemini + DeepSeek + Ollama |
| Infra | Docker + Docker Compose |

---

## 3. 后端分层架构

```
Handler → Service → Repository → Domain
```

**禁止**：在 Handler 中写业务逻辑。

### 3.1 模块结构

| 模块 | 路径 | 职责 |
|------|------|------|
| Auth | `internal/auth/` | JWT 认证、RBAC |
| Service | `internal/service/` | 业务逻辑 |
| Repository | `internal/repository/` | 数据访问 |
| Runtime | `internal/runtime/` | Agent Runtime（进程内组件） |
| Bridge | `internal/bridge/` | SSH Bridge |
| AI | `internal/ai/` | AI Provider 抽象层 |
| Task | `internal/task/` | 任务管理 |
| Monitor | `internal/monitor/` | 监控采集 |
| Permission | `internal/permission/` | 权限系统 |
| Websocket | `internal/websocket/` | WebSocket 处理 |

---

## 4. AI Provider 抽象

所有 AI Provider 必须通过统一 `AIProvider` interface。

**禁止**：业务代码直接调用具体 Provider。

详见 `docs/technical-spec.md` §1。

---

## 5. Agent Tool 系统

Agent 必须使用 Tool Calling。

禁止让 LLM 直接生成 Shell 并自动执行。

Shell 必须经过：

```
Tool → Permission → Risk Assessment → Approval → Execution → Verification
```

---

## 6. Server Operations

禁止浏览器直接执行 Linux command。

所有 Linux 操作必须走：

```
Frontend → API → Service → Agent → Linux
```

---

## 7. Docker 管理

优先使用 Docker SDK / Docker API。

禁止通过字符串拼接 Shell 管理 Docker。

---

## 8. 架构约束（Security Constraints）

1. 禁止浏览器直接执行命令
2. 所有操作必须走 API → Service → Agent 链路
3. 禁止通过字符串拼接 Shell 管理 Docker
4. 配置必须来自环境变量，禁止硬编码
5. 所有重要操作必须记录 Audit Log
6. 敏感值（JWT Secret 等）缺失时拒绝启动

---

## 9. 安全默认值

```
READ     = ALLOW
WRITE    = ASK
DELETE   = ASK
CRITICAL = DENY
```

高风险操作必须人工审批。

详见 `docs/security.md`。

---

## 10. 前端规范

- 优先使用 Arco Design 组件
- 不要重复实现 Arco 已提供的基础组件
- 必须保持 TypeScript strict mode
- 避免 `any` 类型
- 避免巨大组件 — 一个页面必须拆分成合理的 Components
- 业务逻辑不能全部写进 UI Component

详见 `docs/coding-standards.md`。

---

## 11. 代码修改纪律

修改之前必须：

1. 阅读相关代码
2. 理解现有架构
3. 找到真正的调用链
4. 评估影响
5. 再修改

**禁止**：为了修复一个问题而大范围重构。

---

## 12. 参考文档

- **技术实现细节**：`docs/technical-spec.md`
- **编码规范**：`docs/coding-standards.md`
- **安全规范**：`docs/security.md`
- **架构评审**：`docs/architecture-review.md`
