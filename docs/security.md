---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: Lead
related: [AGENTS.md, docs/architecture.md]
---

# Security Standards

> **安全规范** — 项目的安全策略和操作边界。
>
> 适用对象：所有 Agent。本文件优先级高于其他专项文档。

---

## 1. 安全默认值

```
READ     = ALLOW
WRITE    = ASK
DELETE   = ASK
CRITICAL = DENY
```

高风险操作必须人工审批。

---

## 2. 安全红线（L2，不可覆盖）

**永久禁止：**

- 提交任何密钥、凭证、`.env`、私钥
- 硬编码 API Key
- 硬编码服务器密码
- 让 AI 默认 root 执行
- 绕过权限系统
- 绕过 Approval 流程
- 静默执行高风险命令
- 在代码中引入未审查的第三方依赖

---

## 3. 敏感信息处理

### 3.1 禁止记录的信息

Audit Log 中不得记录：

- Password
- API Key
- Private Key
- Token
- Credit Card Number
- SSN

### 3.2 配置管理

- 敏感配置必须来自环境变量
- 使用 `.env.example` 提供配置模板（不含真实值）
- 必要的安全配置缺失时拒绝启动

---

## 4. 操作风险等级

| 等级 | 操作示例 | 审批要求 |
|------|----------|----------|
| L0 — Safe | 读取源码、搜索文件 | 无需审批 |
| L1 — Low Risk | 修改源码、运行测试 | 自动通过 |
| L2 — Medium Risk | 安装依赖、修改配置 | Dev 自主 |
| L3 — High Risk | 数据库迁移、部署 Staging | Lead 审批 |
| L4 — Critical | 部署 Production、删除生产数据 | Human + Lead 双审批 |

---

## 5. 权限模型

### 5.1 Agent 权限矩阵

| 操作 | Lead | Dev | Reviewer | QA | Docs |
|------|------|-----|----------|-----|------|
| read source | ✅ | ✅ | ✅ | ✅ | ✅ |
| edit source | ✅ | ✅ | — | — | — |
| execute command | ✅ (L0-L2) | ✅ (L0-L1) | — | ✅ (test) | — |
| deploy | ✅ (L3+) | — | — | — | — |
| manage users | ✅ | — | — | — | — |
| access production | ✅ (L4) | — | — | — | — |

### 5.2 审批流程

```
低风险操作（L0-L1）→ 自动通过
中风险操作（L2）   → Dev 自主决策
高风险操作（L3）   → Lead 审批
critical 操作（L4）→ Human + Lead 双审批
```

---

## 6. 审计日志

所有重要操作必须记录 Audit Log：

- 谁（Who）：Agent ID 或 User ID
- 什么时候（When）：时间戳
- 做了什么（What）：操作类型和目标
- 结果（Result）：成功/失败

**禁止记录**：Password, API Key, Private Key, Token

---

## 7. AI Agent 安全

### 7.1 Tool Calling 必须

Agent 必须使用 Tool Calling。

禁止让 LLM 直接生成命令并自动执行。

### 7.2 命令执行流程

```
Tool → Permission → Risk Assessment → Approval → Execution → Verification
```

### 7.3 操作隔离

禁止浏览器直接执行命令。

所有操作必须走：

```
Frontend → API → Service → Agent → Target
```

---

## 8. 容器安全

- 优先使用容器 SDK / API
- 禁止通过字符串拼接 Shell 管理容器
- 生产容器不使用 root 用户
- 敏感信息通过 Secret 或环境变量注入

---

## 9. 安全审查

以下场景必须触发安全审查：

- 新增第三方依赖
- 修改认证/授权逻辑
- 修改数据库 schema
- 新增 API 端点
- 修改权限配置
