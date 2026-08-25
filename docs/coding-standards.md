---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: Lead
related: [AGENTS.md, docs/architecture.md]
---

# Coding Standards

> **编码规范** — 项目的代码风格和质量标准。
>
> 适用对象：所有需要修改代码的 Agent（Dev, QA, Lead）。

---

## 1. 前端规范

### 1.1 组件设计

- 优先使用项目指定的 UI 组件库，不要重复实现已提供的基础组件
- 一个页面必须拆分成合理的 Components
- 业务逻辑不能全部写进 UI Component
- 组件命名：PascalCase
- Hook 命名：use + PascalCase

### 1.2 TypeScript

- 必须保持 strict mode
- 避免 `any` 类型
- 使用 Interface 定义数据结构
- 使用 Type 定义联合类型和映射类型

### 1.3 代码组织

```
src/
├── components/      # 通用组件
├── pages/           # 页面组件
├── hooks/           # 自定义 Hooks
├── services/        # API 服务
├── types/           # TypeScript 类型定义
├── utils/           # 工具函数
└── assets/          # 静态资源
```

---

## 2. 后端规范

### 2.1 分层架构

```
api/             # HTTP Handler（薄层，只做参数解析和响应）
service/         # 业务逻辑（核心层）
repository/      # 数据访问（数据库操作）
models/          # 数据模型
runtime/         # Agent Runtime
bridge/          # SSH Bridge
auth/            # 认证授权
permission/      # 权限系统
task/            # 任务管理
monitor/         # 监控
websocket/       # WebSocket 处理
```

**禁止**：在 Handler 中写业务逻辑。

### 2.2 命名规范

- 包名：全小写，简短有意义
- 类型名：PascalCase
- 函数名：PascalCase（导出）/ camelCase（包内）
- 接口名：PascalCase，以 `er` 或 `able` 结尾

### 2.3 错误处理

- 使用标准错误包装方式带 `%w` 包装错误
- 不要在业务逻辑中 panic
- 返回错误给调用方，由上层统一处理

```go
// 正确
if err != nil {
    return fmt.Errorf("failed to auth user: %w", err)
}

// 错误
if err != nil {
    panic(err)
}
```

### 2.4 配置管理

- 敏感配置必须来自环境变量
- 使用 `config.Load()` 统一加载配置
- 必要配置缺失时拒绝启动

---

## 3. 通用规范

### 3.1 注释

- 导出符号必须有注释
- 复杂逻辑必须有注释说明"为什么"而非"是什么"
- 禁止遗留调试注释

### 3.2 日志

- 使用结构化日志
- 禁止在生产代码中遗留调试输出
- 敏感信息不得写入日志

### 3.3 代码修改纪律

修改代码前必须：

1. 阅读相关代码
2. 理解现有架构
3. 找到真正的调用链
4. 评估影响范围
5. 再修改

**禁止**：为了修复一个问题而大范围重构。

---

## 4. 禁止事项

- 硬编码 API Key / 密码
- 让 AI 默认 root 执行
- 绕过权限系统
- 绕过 Approval 流程
- 把所有逻辑写进一个文件
- 为了方便而破坏架构分层
