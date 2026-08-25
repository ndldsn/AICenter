# AICenter Development Rules

你正在开发 AICenter。

AICenter 是一个生产级 AI Server Control Plane。

不要把项目实现成 Demo。

必须优先考虑：

Security
Reliability
Maintainability
Observability
Extensibility

## Architecture

Frontend:
React + TypeScript + Arco Design

Backend:
Go

Database:
SQLite development
PostgreSQL production

Communication:
REST + WebSocket

## Frontend Rules

优先使用 Arco Design。

不要重复实现 Arco 已经提供的基础组件。

必须保持：

TypeScript strict mode。

避免：

any

避免：

巨大组件。

一个页面必须拆分成合理的 Components。

业务逻辑不能全部写进 UI Component。

## Backend Rules

Backend 必须采用分层架构：

Handler
Service
Repository
Domain

不要把业务逻辑写在 HTTP Handler。

## Server Operations

禁止浏览器直接执行 Linux command。

所有 Linux 操作必须：

Frontend
↓
API
↓
Service
↓
Agent
↓
Linux

## Docker

优先 Docker SDK / Docker API。

不要通过字符串拼接 Shell 管理 Docker。

## AI

所有 AI Provider 必须通过统一 AIProvider interface。

禁止业务代码直接调用具体 Provider。

## Agent

Agent 必须使用 Tool Calling。

禁止让 LLM 直接生成 Shell 并自动执行。

Shell 必须经过：

Tool
Permission
Risk Assessment
Approval
Execution
Verification

## Security

默认：

READ = ALLOW

WRITE = ASK

DELETE = ASK

CRITICAL = DENY

高风险操作必须人工审批。

## Audit

所有重要操作必须 Audit Log。

不得记录：

Password
API Key
Private Key
Token

## AI Agent

Agent 默认只读。

Agent 必须先观察。

不要为了完成任务而直接修改系统。

流程：

Inspect
Analyze
Plan
Approve
Execute
Verify

## 修改代码

修改之前：

1. 阅读相关代码
2. 理解现有架构
3. 找到真正的调用链
4. 评估影响
5. 再修改

不要为了修复一个问题而大范围重构。

## 测试

每个重要模块必须有测试。

修改以后运行相关测试。

如果测试失败：

分析原因。

不要简单删除测试。

## Git

每个阶段完成后：

git status

检查 diff。

不要提交：

.env
Secrets
Credentials
Private Keys

Commit message 使用：

feat:
fix:
refactor:
docs:
test:
chore:

## 不允许

不要：

硬编码 API Key

不要：

硬编码服务器密码

不要：

让 AI 默认 root

不要：

绕过权限系统

不要：

绕过 Approval

不要：

静默执行高风险命令

不要：

把所有逻辑写进一个文件

不要：

为了方便而破坏架构

## 工作方式

每次开始任务：

先检查现状。

然后分析。

然后制定实现方案。

然后修改。

然后测试。

最后汇报：

Changed
Tested
Issues
Next Step
