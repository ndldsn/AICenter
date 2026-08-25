---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: DevOps
related: [AGENTS.md, docs/security.md]
---

# Deployment Standards

> **部署规范** — 项目的 Docker 和部署策略。
>
> 适用对象：DevOps, Lead。

---

## 1. 部署架构

### 1.1 开发环境

```bash
# Terminal 1: Start backend
cd backend
make dev

# Terminal 2: Start frontend
cd frontend
npm install
npm run dev
```

### 1.2 容器部署

```bash
docker compose -f deployments/docker-compose.yml up -d
```

---

## 2. Docker 规范

### 2.1 多阶段构建

Backend 使用多阶段构建：

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app ./cmd/server

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./app"]
```

### 2.2 禁止事项

- 禁止在容器中硬编码敏感信息
- 禁止使用 root 用户运行服务
- 禁止通过字符串拼接 Shell 管理 Docker

---

## 3. 环境变量

### 3.1 必选变量

```bash
JWT_SECRET=<required>           # JWT 签名密钥
DATABASE_URL=<required>         # 数据库连接字符串
LOG_LEVEL=info                  # 日志级别
```

### 3.2 可选变量

```bash
AI_PROVIDER=openai              # AI Provider
AI_API_KEY=<optional>           # AI API Key
DOCKER_HOST=<optional>          # Docker host
```

### 3.3 配置模板

使用 `.env.example` 提供配置模板（不含真实值）：

```bash
# Copy to .env and fill in values
cp .env.example .env
```

---

## 4. CI/CD

### 4.1 CI 检查项

PR 合并前必须通过：

1. 后端测试
2. 前端测试
3. 静态分析
4. 前端 lint
5. 镜像构建

### 4.2 部署流程

```
main 分支合并
    ↓
自动触发 CI
    ↓
CI 通过
    ↓
自动部署到 Staging
    ↓
人工审批
    ↓
部署到 Production
```

---

## 5. Health Check

```yaml
services:
  backend:
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

---

## 6. Rollback

- 保留最近 3 个可用镜像
- 紧急回滚命令：`docker compose up -d --force-recreate`
- 数据库迁移不可回滚时，需手动执行反向迁移

---

## 7. Monitoring

- 监控指标：CPU、内存、请求延迟、错误率
- 日志聚合：结构化日志输出到 stdout
- 告警：P99 延迟 > 1s、错误率 > 1%
