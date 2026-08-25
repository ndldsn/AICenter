---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: Dev
related: [AGENTS.md, docs/architecture.md]
---

# Technical Spec Reference

> **技术规范参考** — 技术实现细节速查。
>
> 本文件承载项目的技术实现规范，包括接口定义、配置结构、协议格式等。
>
> 架构概述参见 `docs/architecture.md`。

---

## 1. AI Provider 接口

```go
package ai

type AIProvider interface {
    Name() string
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ListModels(ctx context.Context) ([]string, error)
}

type ChatRequest struct {
    Model       string
    Messages    []Message
    Temperature float32
    MaxTokens   int
}

type ChatResponse struct {
    Content string
    Usage   Usage
}

type Message struct {
    Role    string // "system", "user", "assistant"
    Content string
}

type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

### 已实现的 Provider

| Provider | 实现类 | 备注 |
|----------|--------|------|
| OpenAI | `OpenAIProvider` | 兼容 OpenAI API |
| Anthropic | `AnthropicProvider` | Claude API |
| Gemini | `GeminiProvider` | Google Gemini |
| DeepSeek | `DeepSeekProvider` | DeepSeek API |
| Ollama | `OllamaProvider` | 本地模型 |

---

## 2. 数据库 Migration

Migration 文件位于 `migrations/` 目录。

**规则**：

- 新增表或字段必须新增 migration 文件
- Migration 文件必须幂等（可重复执行）
- 禁止修改已有 migration 文件

---

## 3. WebSocket 协议

### 3.1 Terminal Session

```
Connect:
GET /ws/terminal?session=<id>

Message Format (JSON):
{
  "type": "input" | "output" | "resize" | "error",
  "data": "<string>"
}
```

### 3.2 Event Streaming

```
Server → Client:
{
  "type": "event",
  "event": "task_started" | "task_progress" | "task_completed" | "task_failed",
  "data": { ... }
}
```

---

## 4. 缓存策略

### 4.1 In-Memory Cache

- 使用 TTL + LRU 缓存
- 缓存热点数据
- 写入操作时无效化相关缓存

### 4.2 Provider Rate Limiting

- 每个 Provider 设置并发限制
- 防止 API 429 风暴

---

## 5. 配置结构

```go
type Config struct {
    Server      ServerConfig
    Database    DatabaseConfig
    JWT         JWTConfig
    AI          AIConfig
    Docker      DockerConfig
    Log         LogConfig
}

type ServerConfig struct {
    Host string
    Port int
}

type DatabaseConfig struct {
    Driver string  // "sqlite" or "postgres"
    DSN    string
}

type JWTConfig struct {
    Secret        string
    Expiry        time.Duration
    RefreshExpiry time.Duration
}

type AIConfig struct {
    DefaultProvider string
    Providers       map[string]ProviderConfig
    Concurrency     int
}
```

### 5.1 必选配置

- `JWT_SECRET` — JWT 签名密钥，缺失时拒绝启动
- `DATABASE_URL` — 数据库连接字符串

### 5.2 可选配置

- `AI_PROVIDER` — 默认 AI Provider
- `AI_API_KEY` — AI API Key
- `DOCKER_HOST` — Docker host

---

## 6. 错误码

| 错误码 | 含义 |
|--------|------|
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Unprocessable Entity |
| 429 | Rate Limited |
| 500 | Internal Server Error |

---

## 7. 响应格式

```go
type ApiResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T    `json:"data,omitempty"`
}
```

---

## 8. 测试工具

- 前端测试：Vitest
- 后端测试：Go test

详见 `docs/testing.md`。

---

## 9. 工作方式

每次开始任务：

```
先检查现状 → 然后分析 → 然后制定实现方案 → 然后修改 → 然后测试 → 最后汇报
```

汇报格式：

```
Changed: <修改内容>
Tested: <测试情况>
Issues: <遗留问题>
Next Step: <下一步计划>
```
