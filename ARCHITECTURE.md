# AICenter 系统架构设计文档

> Version: 1.0 | Date: 2026-08-21 | Author: AICenter Architect

---

## 目录

1. [总体架构](#1-总体架构)
2. [系统模块图](#2-系统模块图)
3. [前端目录结构](#3-前端目录结构)
4. [Backend 目录结构](#4-backend-目录结构)
5. [控制面 Runtime 组件结构](#5-控制面-runtime-组件结构)
6. [数据库 Schema](#6-数据库-schema)
7. [REST API 设计](#7-rest-api-设计)
8. [WebSocket 协议设计](#8-websocket-协议设计)
9. [AI Provider 抽象层](#9-ai-provider-抽象层)
10. [Agent Tool 系统](#10-agent-tool-系统)
11. [权限模型](#11-权限模型)
12. [审批模型](#12-审批模型)
13. [Audit Log 模型](#13-audit-log-模型)
14. [多服务器模型](#14-多服务器模型)
15. [Docker 管理模型](#15-docker-管理模型)
16. [监控模型](#16-监控模型)
17. [任务模型](#17-任务模型)
18. [部署方案](#18-部署方案)
19. [MVP 开发顺序](#19-mvp-开发顺序)
20. [后续迭代路线](#20-后续迭代路线)
21. [风险与技术难点](#21-风险与技术难点)
22. [设计取舍](#22-设计取舍)

---

## 1. 总体架构

### 1.1 系统定位

AICenter 是一个生产级的 AI 驱动统一运维控制平台，采用 **前后端分离 + 两层控制面** 架构：

- **控制面（Control Plane）** — 后端 Go 进程内承载所有业务服务（API / WS / RBAC / Agent Runtime / Task / Monitor / Approval / Audit），Agent Runtime 作为**进程内组件**而非独立进程运行。
- **被管节点（Managed Nodes）** — 通过 **SSH Bridge** 从控制面发起会话，执行命令 / Docker / 采集，不在节点侧常驻任何 AICenter 进程。

架构选择参见 [ADR-001](#c-架构决策记录-adr)。

### 1.2 架构总览

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser (React SPA)                   │
│   Dashboard │ Servers │ Docker │ Models │ Runtime │ Tasks   │
└─────────────┴────────┴────────┴────────┴────────┴────────────┘
              │ REST API                    │ WebSocket
              ▼                             ▼
┌─────────────────────────────────────────────────────────────┐
│         Go Backend  (Control Plane, single process)          │
│  ┌──────────────────────────────────────────────────┐       │
│  │                   API Layer                        │       │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐ │       │
│  │  │  Auth   │ │ Server  │ │ Docker  │ │   AI   │ │       │
│  │  │ Service │ │ Service │ │ Service │ │Service │ │       │
│  │  └─────────┘ └─────────┘ └─────────┘ └────────┘ │       │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐ │       │
│  │  │  Task   │ │  Audit  │ │Approval │ │Monitor │ │       │
│  │  │ Service │ │ Service │ │ Service │ │Service │ │       │
│  │  └─────────┘ └─────────┘ └─────────┘ └────────┘ │       │
│  └──────────────────────────────────────────────────┘       │
│                                                              │
│  ┌──────────────────────────────────────────────────┐       │
│  │              AI Provider Layer (interface)         │       │
│  │  OpenAI │ Anthropic │ Gemini │ DeepSeek │ Ollama   │       │
│  └──────────────────────────────────────────────────┘       │
│                                                              │
│  ┌──────────────────────────────────────────────────┐       │
│  │           Agent Runtime (进程内组件)                │       │
│  │  Session │ Tool Executor │ Planner │ Approval     │       │
│  │  + Tool 注册中心 + Tool Permission 检查              │       │
│  └──────────────────────────────────────────────────┘       │
│                                                              │
│  ┌──────────────────────────────────────────────────┐       │
│  │              SSH Bridge (Server Executor)          │       │
│  │  SSH Session 管理 │ 命令执行 │ Docker SDK │ 指标采集  │       │
│  └──────────────────────────────────────────────────┘       │
└───────────────────────────────┬─────────────────────────────┘
                                │ SSH (SSH-2, 密钥认证 + 权限校验)
                                │
              ┌─────────────────┼─────────────────────┐
              │                 │                     │
         ┌────▼─────┐     ┌────▼─────┐        ┌────▼─────┐
         │ Node 001 │     │ Node 002 │        │ Node 00N │
         │ (无常驻   │     │ (无常驻   │        │ (无常驻   │
         │  AICenter│     │  AICenter│        │  AICenter│
         │  进程)   │     │ 进程)    │        │  进程)   │
         └──────────┘     └──────────┘        └──────────┘
```

### 1.3 分层说明

| 层级 | 职责 | 技术 |
|------|------|------|
| 展示层 | 用户界面、数据可视化、交互 | React + Arco Design + ECharts |
| 网关层 | 认证、路由、限流、审计、WS 网关 | Go + Gin |
| 服务层 | 业务逻辑编排（API / Task / Monitor / Approval / Audit） | Go Services |
| AI 层 | Provider 抽象（OpenAI/Anthropic/.../Ollama） | Go + Provider SDKs |
| Runtime 层 | Agent Runtime（Session / Tool Executor / Planner / Approval / Tool 注册中心） — **进程内组件** | Go（backend/internal/runtime） |
| 执行层 | SSH Bridge：通过 SSH 对被管节点执行命令 / Docker / 采集（不在节点侧常驻） | Go（backend/internal/executor + Go SSH lib） |
| 数据层 | 持久化、缓存 | SQLite(开发) + PostgreSQL(生产) |

### 1.4 关键架构决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 前后端通信 | REST + WebSocket | REST 用于 CRUD，WS 用于实时推送（Agent 会话、审批事件、指标流） |
| Agent 架构 | **进程内 Runtime（见 ADR-001）** | 降低部署面、消除 Agent 进程间一致性问题，把复杂度收到控制面 |
| 节点通信 | **SSH Bridge（自控制面发起）** | 节点侧无常驻进程，天然符合"默认只读、高危操作走审批"原则 |
| 数据库 | SQLite(开发) + PostgreSQL(生产) | 开发简单，生产可靠（dual-driver 已落地） |
| 缓存 | Redis | 会话、实时数据、分布式锁（可选；MVP 可省） |
| 认证 | JWT + Refresh Token + RBAC | 无状态，权限基于角色 → 权限组 → 权限链 |
| 消息 | 内存通道 + WS（MVP）；可选 Redis Pub/Sub | 单控制面进程内无需外部 MQ；多实例时再升级 |

---

## 2. 系统模块图

```
                         ┌──────────────┐
                         │   AICenter   │
                         │   Frontend   │
                         └──────┬───────┘
                                │
                    ┌───────────┼───────────┐
                    │           │           │
              ┌─────▼─────┐ ┌───▼───┐ ┌─────▼─────┐
              │  Auth     │ │  API  │ │ WebSocket │
              │  Module   │ │Gateway│ │   Hub     │
              └─────┬─────┘ └───┬───┘ └─────┬─────┘
                    │           │           │
        ┌───────────┼───────────┼───────────┼───────────┐
        │           │           │           │           │
   ┌────▼───┐ ┌────▼───┐ ┌─────▼──┐ ┌─────▼──┐ ┌─────▼──┐
   │ User   │ │ Server │ │ Docker │ │   AI   │ │  Task  │
   │Module  │ │Module  │ │Module  │ │Module  │ │Module  │
   └────┬───┘ └────┬───┘ └────┬───┘ └────┬───┘ └────┬───┘
        │          │          │          │          │
   ┌────▼───┐ ┌────▼───┐ ┌─────▼──┐ ┌─────▼──┐ ┌─────▼──┐
   │ Audit  │ │Approval│ │Monitor │ │Runtime │ │Notification│
   │Module  │ │Module  │ │Module  │ │Module  │ │Module     │
   └────┬───┘ └────┬───┘ └────┬───┘ └────┬───┘ └────────┬─┘
        │          │          │          │               │
        │    ┌─────▼──────────▼──────────▼───────────────┘
        │    │
   ┌────▼────▼──────────────────────────────────────┐
   │               SSH Bridge Layer                  │
   │  Session 池 │ 命令执行 │ Docker SDK │ 指标采集    │
   └────────────┬───────────────────────────────────┘
                │ SSH（控制面发起，节点侧无常驻进程）
        ┌───────┼──────────────┐
        │       │              │
   ┌────▼───┐ ┌▼─────┐ ┌─────▼───┐
   │ Node A │ │Node B│ │ Node C  │
   │(无常驻) │ │(无常驻│ │(无常驻)  │
   └────────┘ └──────┘ └─────────┘
```

### 模块职责

| 模块 | 职责 | 关键实体 |
|------|------|----------|
| User | 认证、授权、角色、偏好 | User, Role, Permission |
| Server | 服务器管理、分组、连接 | Server, ServerGroup, ServerMetric |
| Docker | 容器/镜像/Volume/Compose 管理 | Container, Image, Volume, ComposeProject |
| AI | Provider 管理、模型管理 | Provider, AIModel |
| Task | 任务调度、执行、历史 | Task, TaskStep, ExecutionLog |
| Audit | 操作审计、合规 | AuditLog |
| Approval | 审批流程、Dry Run | ApprovalRequest, ExecutionPlan |
| Monitor | 指标采集（via SSH）、告警 | Metric, AlertRule, AlertEvent |
| **Runtime** | Agent Runtime（进程内组件）：会话、消息、Tool 注册、执行、Planner | AgentSession, AgentMessage, Tool, ToolRegistry |
| **SSH Bridge** | 通过 SSH 对被管节点执行命令 / Docker / 采集（进程内执行层） | SSHSession, SSHExecResult |
| Notification | 通知渠道 | NotificationChannel |

---

## 3. 前端目录结构

```
frontend/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── public/
│   └── favicon.ico
├── src/
│   ├── main.tsx                          # 入口
│   ├── App.tsx                           # 根组件
│   ├── env.d.ts                          # 环境类型声明
│   │
│   ├── app/                              # App 级配置
│   │   ├── router.tsx                    # 路由配置
│   │   ├── query-client.ts               # TanStack Query 配置
│   │   ├── axios.ts                      # Axios 实例 + 拦截器
│   │   ├── auth.tsx                      # Auth Provider
│   │   └── bootstrap.ts                  # 启动引导
│   │
│   ├── assets/                           # 静态资源
│   │   ├── images/
│   │   ├── icons/
│   │   └── styles/
│   │       ├── global.css
│   │       ├── tailwind.css
│   │       └── theme.ts                  # Arco Design 主题
│   │
│   ├── components/                       # 通用组件
│   │   ├── common/
│   │   │   ├── PageHeader.tsx
│   │   │   ├── EmptyState.tsx
│   │   │   ├── Loading.tsx
│   │   │   ├── ConfirmModal.tsx
│   │   │   ├── SearchInput.tsx
│   │   │   ├── StatusBadge.tsx
│   │   │   └── RiskBadge.tsx             # 风险等级标签
│   │   ├── charts/
│   │   │   ├── LineChart.tsx             # ECharts 折线图
│   │   │   ├── AreaChart.tsx
│   │   │   ├── GaugeChart.tsx
│   │   │   └── RealtimeChart.tsx         # 实时滚动图表
│   │   ├── editors/
│   │   │   ├── CodeEditor.tsx            # Monaco Editor 封装
│   │   │   ├── JsonEditor.tsx
│   │   │   └── YamlEditor.tsx
│   │   ├── terminal/
│   │   │   ├── WebTerminal.tsx           # xterm.js 封装
│   │   │   └── TerminalPanel.tsx
│   │   ├── docker/
│   │   │   ├── ContainerTable.tsx
│   │   │   ├── ContainerLogs.tsx
│   │   │   ├── ImageTable.tsx
│   │   │   ├── VolumeTable.tsx
│   │   │   └── ComposeEditor.tsx
│   │   ├── agent/
│   │   │   ├── AgentChat.tsx             # Agent 对话界面
│   │   │   ├── MessageBubble.tsx
│   │   │   ├── ToolCallCard.tsx          # Tool 调用展示卡片
│   │   │   ├── ExecutionPlanCard.tsx     # 执行计划展示
│   │   │   ├── DryRunPreview.tsx         # Dry Run 预览
│   │   │   └── ThinkingBlock.tsx         # Agent 思考过程
│   │   ├── approval/
│   │   │   ├── ApprovalModal.tsx         # 审批弹窗
│   │   │   ├── ApprovalList.tsx
│   │   │   └── ApprovalDetail.tsx
│   │   └── monitor/
│   │       ├── MetricCard.tsx
│   │       ├── ServerStatusBoard.tsx
│   │       └── AlertBanner.tsx
│   │
│   ├── features/                         # 按业务域组织的特性模块
│   │   ├── auth/
│   │   │   ├── LoginPage.tsx
│   │   │   ├── RegisterPage.tsx
│   │   │   ├── stores/
│   │   │   │   └── authStore.ts
│   │   │   └── hooks/
│   │   │       └── useAuth.ts
│   │   │
│   │   ├── dashboard/
│   │   │   ├── DashboardPage.tsx
│   │   │   ├── OverviewCards.tsx
│   │   │   ├── RecentTasks.tsx
│   │   │   └── AlertSummary.tsx
│   │   │
│   │   ├── servers/
│   │   │   ├── ServerListPage.tsx
│   │   │   ├── ServerDetailPage.tsx
│   │   │   ├── ServerFormModal.tsx
│   │   │   ├── ServerMonitorPage.tsx
│   │   │   ├── ServerTerminalPage.tsx
│   │   │   ├── components/
│   │   │   │   ├── ServerConnectionForm.tsx
│   │   │   │   ├── ServerMetricsGrid.tsx
│   │   │   │   └── ServerGroupTree.tsx
│   │   │   ├── stores/
│   │   │   │   └── serverStore.ts
│   │   │   └── hooks/
│   │   │       ├── useServers.ts
│   │   │       ├── useServerMetrics.ts
│   │   │       └── useServerConnection.ts
│   │   │
│   │   ├── docker/
│   │   │   ├── DockerDashboardPage.tsx
│   │   │   ├── ContainerPage.tsx
│   │   │   ├── ImagePage.tsx
│   │   │   ├── VolumePage.tsx
│   │   │   ├── ComposePage.tsx
│   │   │   ├── components/
│   │   │   │   ├── ContainerActions.tsx
│   │   │   │   ├── ContainerStats.tsx
│   │   │   │   ├── ComposeDeployModal.tsx
│   │   │   │   └── ImagePullModal.tsx
│   │   │   └── hooks/
│   │   │       ├── useContainers.ts
│   │   │       ├── useDockerEvents.ts
│   │   │       └── useCompose.ts
│   │   │
│   │   ├── models/
│   │   │   ├── ModelListPage.tsx
│   │   │   ├── ModelConfigModal.tsx
│   │   │   ├── ProviderListPage.tsx
│   │   │   ├── ProviderConfigModal.tsx
│   │   │   └── hooks/
│   │   │       ├── useModels.ts
│   │   │       └── useProviders.ts
│   │   │
│   │   ├── agents/
│   │   │   ├── AgentListPage.tsx
│   │   │   ├── AgentDetailPage.tsx
│   │   │   ├── AgentChatPage.tsx
│   │   │   ├── AgentConfigModal.tsx
│   │   │   ├── components/
│   │   │   │   ├── AgentSessionList.tsx
│   │   │   │   ├── MessageThread.tsx
│   │   │   │   ├── ToolPermissionEditor.tsx
│   │   │   │   └── ExecutionTimeline.tsx
│   │   │   └── hooks/
│   │   │       ├── useAgentSessions.ts
│   │   │       ├── useAgentMessages.ts
│   │   │       └── useAgentExecution.ts
│   │   │
│   │   ├── tasks/
│   │   │   ├── TaskListPage.tsx
│   │   │   ├── TaskDetailPage.tsx
│   │   │   ├── components/
│   │   │   │   ├── TaskStepList.tsx
│   │   │   │   ├── TaskProgressBar.tsx
│   │   │   │   └── TaskLogViewer.tsx
│   │   │   └── hooks/
│   │   │       └── useTasks.ts
│   │   │
│   │   ├── monitor/
│   │   │   ├── MonitorDashboardPage.tsx
│   │   │   ├── MetricsExplorerPage.tsx
│   │   │   ├── AlertRulePage.tsx
│   │   │   └── AlertHistoryPage.tsx
│   │   │
│   │   ├── audit/
│   │   │   ├── AuditLogPage.tsx
│   │   │   └── AuditDetailDrawer.tsx
│   │   │
│   │   ├── approvals/
│   │   │   ├── PendingApprovalPage.tsx
│   │   │   └── ApprovalHistoryPage.tsx
│   │   │
│   │   └── settings/
│   │       ├── SettingsPage.tsx
│   │       ├── UserManagement.tsx
│   │       ├── RoleManagement.tsx
│   │       └── SystemSettings.tsx
│   │
│   ├── hooks/                            # 全局共享 Hooks
│   │   ├── useWebSocket.ts               # WebSocket Hook
│   │   ├── usePageTitle.ts
│   │   ├── useDebounce.ts
│   │   ├── useClipboard.ts
│   │   ├── useMediaQuery.ts
│   │   └── useTheme.ts
│   │
│   ├── layouts/                          # 布局组件
│   │   ├── MainLayout.tsx                # 主布局（侧边栏+顶栏+内容）
│   │   ├── Sidebar.tsx
│   │   ├── Navbar.tsx
│   │   ├── Breadcrumb.tsx
│   │   └── PageContainer.tsx
│   │
│   ├── pages/                            # 页面（路由对应，调用 features）
│   │   ├── index.ts                      # 页面导出聚合
│   │   └── ...                           # 引用 features/* 下的页面
│   │
│   ├── routes/                           # 路由配置
│   │   ├── index.tsx
│   │   ├── routes.tsx                    # 路由定义
│   │   ├── ProtectedRoute.tsx
│   │   └── routeConfig.ts                # 路由元数据（标题、权限）
│   │
│   ├── services/                         # API 服务层
│   │   ├── api.ts                        # API 基础封装
│   │   ├── auth.ts
│   │   ├── servers.ts
│   │   ├── docker.ts
│   │   ├── models.ts
│   │   ├── providers.ts
│   │   ├── agents.ts
│   │   ├── tasks.ts
│   │   ├── monitor.ts
│   │   ├── audit.ts
│   │   ├── approvals.ts
│   │   └── websocket.ts                  # WS 消息类型与发送
│   │
│   ├── stores/                           # Zustand 状态管理
│   │   ├── index.ts
│   │   ├── uiStore.ts                    # UI 状态（侧边栏折叠等）
│   │   ├── notificationStore.ts          # 通知状态
│   │   └── realtimeStore.ts              # 实时数据缓存
│   │
│   ├── types/                            # TypeScript 类型定义
│   │   ├── index.ts
│   │   ├── auth.ts
│   │   ├── server.ts
│   │   ├── docker.ts
│   │   ├── ai.ts
│   │   ├── agent.ts
│   │   ├── task.ts
│   │   ├── monitor.ts
│   │   ├── audit.ts
│   │   ├── approval.ts
│   │   └── websocket.ts                  # WS 消息类型
│   │
│   └── utils/                            # 工具函数
│       ├── format.ts                     # 格式化（时间、字节、数字）
│       ├── validate.ts                   # 校验
│       ├── storage.ts                    # localStorage 封装
│       ├── crypto.ts                     # 加密辅助
│       └── constants.ts                  # 常量
│
└── .env                                  # 环境变量
```

---

## 4. Backend 目录结构

```
backend/
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── .env.example
│
├── cmd/                                  # 入口
│   ├── aicenter/
│   │   └── main.go                       # API Server 入口
│   ├── agent/
│   │   └── main.go                       # Agent 入口
│   └── migrate/
│       └── main.go                       # 数据库迁移工具
│
├── configs/                              # 配置文件模板
│   ├── config.yaml.example
│   └── agent.yaml.example
│
├── deployments/                          # 部署相关
│   ├── docker-compose.yml
│   ├── Dockerfile.api
│   └── Dockerfile.agent
│
├── docs/                                 # 文档
│   ├── api.md
│   └── architecture.md
│
├── migrations/                           # 数据库迁移文件
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   ├── 002_create_servers.up.sql
│   └── ...
│
├── internal/                             # 私有包
│   │
│   ├── api/                              # HTTP 处理层
│   │   ├── handler/                      # 请求处理器
│   │   │   ├── auth_handler.go
│   │   │   ├── server_handler.go
│   │   │   ├── docker_handler.go
│   │   │   ├── model_handler.go
│   │   │   ├── provider_handler.go
│   │   │   ├── agent_handler.go
│   │   │   ├── task_handler.go
│   │   │   ├── monitor_handler.go
│   │   │   ├── audit_handler.go
│   │   │   ├── approval_handler.go
│   │   │   └── websocket_handler.go
│   │   ├── middleware/                   # 中间件
│   │   │   ├── auth.go                   # JWT 认证
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   ├── audit.go                  # 审计中间件
│   │   │   ├── permission.go             # 权限检查
│   │   │   ├── recovery.go
│   │   │   └── request_id.go
│   │   ├── request/                      # 请求 DTO
│   │   │   ├── auth.go
│   │   │   ├── server.go
│   │   │   ├── docker.go
│   │   │   └── ...
│   │   ├── response/                     # 响应 DTO
│   │   │   ├── common.go
│   │   │   ├── server.go
│   │   │   └── ...
│   │   └── router/                       # 路由注册
│   │       ├── router.go
│   │       ├── auth.go
│   │       ├── server.go
│   │       ├── docker.go
│   │       ├── ai.go
│   │       ├── agent.go
│   │       └── ...
│   │
│   ├── auth/                             # 认证逻辑
│   │   ├── jwt.go
│   │   ├── password.go
│   │   ├── rbac.go                       # 基于角色的访问控制
│   │   └── permission.go                 # 权限定义与检查
│   │
│   ├── config/                           # 配置加载
│   │   ├── config.go
│   │   └── loader.go
│   │
│   ├── database/                         # 数据库层
│   │   ├── db.go                         # 连接管理
│   │   ├── migrate.go                    # 迁移逻辑
│   │   └── seed.go                       # 初始数据
│   │
│   ├── models/                           # 数据模型（实体）
│   │   ├── user.go
│   │   ├── role.go
│   │   ├── permission.go
│   │   ├── server.go
│   │   ├── server_metric.go
│   │   ├── docker.go
│   │   ├── ai_model.go
│   │   ├── ai_provider.go
│   │   ├── agent.go
│   │   ├── agent_session.go
│   │   ├── agent_message.go
│   │   ├── task.go
│   │   ├── tool.go
│   │   ├── approval.go
│   │   ├── audit_log.go
│   │   ├── monitor_metric.go
│   │   ├── alert.go
│   │   └── base.go                       # 公共字段
│   │
│   ├── repository/                       # 数据访问层
│   │   ├── user_repo.go
│   │   ├── server_repo.go
│   │   ├── docker_repo.go
│   │   ├── ai_repo.go
│   │   ├── agent_repo.go
│   │   ├── task_repo.go
│   │   ├── approval_repo.go
│   │   ├── audit_repo.go
│   │   ├── monitor_repo.go
│   │   └── interfaces.go                # Repo 接口定义
│   │
│   ├── service/                          # 业务逻辑层
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── server_service.go
│   │   ├── docker_service.go
│   │   ├── ai_service.go
│   │   ├── provider_service.go
│   │   ├── agent_service.go
│   │   ├── task_service.go
│   │   ├── monitor_service.go
│   │   ├── audit_service.go
│   │   ├── approval_service.go
│   │   ├── notification_service.go
│   │   └── interfaces.go
│   │
│   ├── provider/                         # AI Provider 抽象层
│   │   ├── provider.go                   # Provider 接口
│   │   ├── manager.go                    # Provider 管理
│   │   ├── factory.go                    # Provider 工厂
│   │   ├── openai/
│   │   │   └── openai_provider.go
│   │   ├── anthropic/
│   │   │   └── anthropic_provider.go
│   │   ├── gemini/
│   │   │   └── gemini_provider.go
│   │   ├── deepseek/
│   │   │   └── deepseek_provider.go
│   │   ├── ollama/
│   │   │   └── ollama_provider.go
│   │   └── custom/
│   │       └── custom_provider.go
│   │
│   ├── agent/                            # Agent 运行时
│   │   ├── runtime/
│   │   │   ├── runtime.go                # Agent 运行时主控
│   │   │   ├── session.go                # 会话管理
│   │   │   ├── context.go                # 上下文管理
│   │   │   ├── memory.go                 # 会话记忆
│   │   │   └── loop.go                   # Agent Loop（思考-行动循环）
│   │   ├── planner/
│   │   │   ├── planner.go                # 任务规划器
│   │   │   └── plan.go                   # 执行计划模型
│   │   ├── tools/                        # 内置 Tool
│   │   │   ├── registry.go               # Tool 注册表
│   │   │   ├── interfaces.go             # Tool 接口定义
│   │   │   ├── read/                     # 只读 Tool
│   │   │   │   ├── system_info.go
│   │   │   │   ├── system_load.go
│   │   │   │   ├── disk_usage.go
│   │   │   │   ├── memory_usage.go
│   │   │   │   ├── network_info.go
│   │   │   │   ├── process_list.go
│   │   │   │   ├── read_log.go
│   │   │   │   ├── read_file.go
│   │   │   │   ├── docker_list.go
│   │   │   │   ├── docker_inspect.go
│   │   │   │   ├── docker_logs.go
│   │   │   │   └── docker_stats.go
│   │   │   ├── write/                    # 写入 Tool
│   │   │   │   ├── write_file.go
│   │   │   │   ├── edit_file.go
│   │   │   │   └── create_file.go
│   │   │   ├── docker/                   # Docker Tool
│   │   │   │   ├── container_start.go
│   │   │   │   ├── container_stop.go
│   │   │   │   ├── container_restart.go
│   │   │   │   ├── container_delete.go
│   │   │   │   ├── image_delete.go
│   │   │   │   ├── volume_delete.go
│   │   │   │   ├── compose_up.go
│   │   │   │   ├── compose_down.go
│   │   │   │   └── compose_update.go
│   │   │   ├── exec/                     # 执行 Tool
│   │   │   │   ├── exec_command.go
│   │   │   │   └── exec_shell.go
│   │   │   └── system/                   # 系统 Tool
│   │   │       ├── reboot.go
│   │   │       ├── service_restart.go
│   │   │       ├── install_package.go
│   │   │       └── modify_firewall.go
│   │   ├── permissions/                  # Tool 权限
│   │   │   ├── permission.go             # 权限定义
│   │   │   ├── enforcer.go               # 权限检查器
│   │   │   └── risk.go                   # 风险评估
│   │   └── approval/                     # 审批逻辑
│   │       ├── approval.go
│   │       ├── dryrun.go
│   │       └── executor.go               # 受控执行器
│   │
│   ├── monitor/                          # 监控模块
│   │   ├── collector.go                  # 指标采集器接口
│   │   ├── store.go                      # 指标存储
│   │   ├── alert.go                      # 告警引擎
│   │   └── notifier.go                   # 告警通知
│   │
│   ├── task/                             # 任务系统
│   │   ├── scheduler.go                  # 任务调度器
│   │   ├── executor.go                   # 任务执行器
│   │   ├── worker.go                     # 工作池
│   │   └── types.go                      # 任务类型定义
│   │
│   ├── websocket/                        # WebSocket Hub
│   │   ├── hub.go                        # Hub 管理
│   │   ├── client.go                     # 客户端连接
│   │   ├── message.go                    # 消息协议
│   │   └── rooms.go                      # 房间（按 server/agent 分组）
│   │
│   └── pkg/                              # 内部共享包
│       ├── logger/
│       │   └── logger.go                 # Zap Logger 封装
│       ├── crypto/
│       │   └── crypto.go                 # 加密/解密
│       ├── docker/
│       │   └── client.go                 # Docker Client 封装
│       ├── ssh/
│       │   └── client.go                 # SSH Client 封装
│       ├── utils/
│       │   ├── string.go
│       │   ├── time.go
│       │   └── pointer.go
│       └── version/
│           └── version.go
│
└── scripts/                              # 辅助脚本
    ├── build.sh
    ├── test.sh
    └── release.sh
```

---

## 5. 控制面 Runtime 组件结构

Agent 不再部署为被管节点上的独立进程，而是作为 **控制面 Go 进程内的 Runtime 组件** 运行。运行时与 SSH Bridge 之间通过 Go 接口解耦：Runtime 发起执行意图 → Bridge 通过 SSH 投递到目标节点 → 结果回喂 Runtime。

```
backend/internal/runtime/
├── runtime.go                  # Runtime 主控：启动、生命周期、会话路由
├── session/
│   ├── session.go              # AgentSession 生命周期
│   ├── message.go              # AgentMessage（用户输入 / 助手输出 / tool_call / tool_result）
│   └── stream.go               # 结果流（对接 WS Hub 推送）
│
├── planner/
│   ├── planner.go              # 思考-行动循环（Planner）
│   └── policy.go               # 策略：只读优先、禁止直接 shell、默认走审批
│
├── tool/
│   ├── registry.go             # Tool 注册中心（含 group/permission）
│   ├── definitions.go          # 22+ Tool 定义
│   ├── tool.go                 # Tool 接口 + 执行器
│   ├── execute.go              # 执行编排（含权限检查 + Approval 钩子）
│   ├── sandbox.go              # 沙箱执行策略（可选）
│   └── builtin/
│       ├── shell.go            # Shell Tool（默认 DENY，需审批）
│       ├── docker.go           # Docker Tool（读取 SDK）
│       ├── k8s.go              # K8s Tool
│       ├── server.go           # 服务器读取/操作
│       ├── file.go             # 文件读写
│       └── net.go              # 网络探测
│
├── approval/
│   └── hook.go                 # 与 Approval Service 的对接（Dry Run / 需审批）
│
└── executor/                   # 进程内执行层（对节点的操作入口）
    ├── executor.go             # Executor 接口
    ├── ssh.go                  # SSH Bridge：session 池 + 命令执行 + 上传下载
    ├── docker.go               # Docker SDK 调用（远端 node 通过 SSH 隧道）
    └── collector.go            # 指标采集（通过 SSH 执行采集命令）
```

### 5.1 Runtime 与 SSH Bridge 协作

```
用户发起操作 / Agent 决定动作
        │
        ▼
┌──────────────────────────────┐
│  Agent Runtime (进程内)       │
│  - 权限检查 (RequirePermission)│
│  - Approval 钩子（若需审批）   │
│  - 拼装 Tool 参数             │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│  SSH Bridge (进程内执行层)     │
│  - 从 Server DB 取 SSH 凭据   │
│  - 复用 SSH Session 池        │
│  - 执行命令 / 调 Docker SDK  │
│  - 采集指标（如 uname / df）  │
└──────────────┬───────────────┘
               │ SSH（自控制面发起）
               ▼
        ┌──────────────┐
        │  被管节点     │
        │  (无常驻进程) │
        └──────┬───────┘
               │
               ▼
        结果回喂 Runtime
```

### 5.2 Runtime 设计原则

| 原则 | 说明 |
|------|------|
| 默认只读 | Runtime 所有 Tool 默认为 READ，写操作必须走 Approval |
| 禁止直接 shell | LLM 不能直接生成 shell 命令执行，必须通过 Tool → Permission → Risk Assessment → Approval → Execution → Verification |
| 进程内 | Runtime 是 Go 进程内组件，无独立二进制部署，跟随 control plane 启动/升级 |
| 可观测 | 每次 Tool 调用写入 AuditLog，含 tool、参数、结果、审批状态 |
| 可扩展 | Tool 通过 registry 注册，新增 Tool 只需加定义 + 权限项 |

### 5.3 SSH Bridge 设计要点

- **凭据存储**：SSH 密钥 / 密码存于 `servers.ssh_private_key`（AES-GCM 加密），控制面启动时按需提供
- **Session 池**：每个 Server 维护一个 SSH Session 复用池，避免每次命令都握手
- **心跳/健康检测**：Monitor 每 30s 通过 SSH 执行 `uptime` / `df` 采集，非节点侧心跳
- **超时与失败**：单命令 60s 超时；节点不可达时 Server 状态置为 `offline`，不重试风暴
- **审计**：每条命令的 stdin / stdout / stderr 全部入 AuditLog（敏感值脱敏）
- **安全**：不缓存明文密钥；禁止交互式 shell；所有写操作需 Approval

---

## 6. 数据库 Schema

### 6.1 ER 关系图 (文字描述)

```
User ───< Role ───< Permission
User ───< AuditLog
User ───< ApprovalRequest

Server ───< ServerGroup
Server ───< ServerMetric
Server ───< DockerHost
Server ───< Task

DockerHost ───< Container
DockerHost ───< Image
DockerHost ───< Volume
DockerHost ───< ComposeProject

AIModel ───< AIProvider

Agent ───< AgentSession ───< AgentMessage
Agent ───< ToolPermission (via Role)

Task ───< TaskStep ───< ExecutionLog

ApprovalRequest ───< ExecutionPlan

AuditLog (独立审计表)

AlertRule ───< AlertEvent
```

### 6.2 完整 SQL Schema

```sql
-- ============================================
-- 1. 用户与权限
-- ============================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    role_id         UUID NOT NULL REFERENCES roles(id),
    is_active       BOOLEAN DEFAULT true,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(50) UNIQUE NOT NULL,
    description     TEXT,
    is_system       BOOLEAN DEFAULT false,   -- 系统内置角色不可删除
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 预置角色: superadmin, admin, operator, viewer
INSERT INTO roles (name, description, is_system) VALUES
    ('superadmin', '超级管理员，拥有所有权限', true),
    ('admin', '管理员，管理服务器和用户', true),
    ('operator', '运维人员，日常运维操作', true),
    ('viewer', '只读用户，只能查看', true);

CREATE TABLE permissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) UNIQUE NOT NULL,
    code            VARCHAR(100) UNIQUE NOT NULL,
    description     TEXT,
    category        VARCHAR(50)              -- server/docker/agent/audit/system
);

-- 预置权限
INSERT INTO permissions (name, code, category) VALUES
    -- Server
    ('查看服务器', 'server:read', 'server'),
    ('创建服务器', 'server:create', 'server'),
    ('修改服务器', 'server:update', 'server'),
    ('删除服务器', 'server:delete', 'server'),
    ('连接服务器终端', 'server:terminal', 'server'),
    -- Docker
    ('查看 Docker', 'docker:read', 'docker'),
    ('操作 Docker', 'docker:manage', 'docker'),
    ('管理 Compose', 'docker:compose', 'docker'),
    -- AI
    ('查看 AI 模型', 'ai:read', 'ai'),
    ('管理 AI Provider', 'ai:provider', 'ai'),
    ('管理 AI Agent', 'ai:agent', 'ai'),
    -- Agent
    ('执行 Agent', 'agent:execute', 'agent'),
    ('管理 Agent 配置', 'agent:config', 'agent'),
    -- Task
    ('查看任务', 'task:read', 'task'),
    ('创建任务', 'task:create', 'task'),
    ('管理任务', 'task:manage', 'task'),
    -- Audit
    ('查看审计日志', 'audit:read', 'audit'),
    -- Approval
    ('审批操作', 'approval:approve', 'approval'),
    ('查看审批', 'approval:read', 'approval'),
    -- Monitor
    ('查看监控', 'monitor:read', 'monitor'),
    ('管理告警规则', 'monitor:alert', 'monitor'),
    -- System
    ('系统设置', 'system:settings', 'system'),
    ('用户管理', 'system:user', 'system'),
    ('角色管理', 'system:role', 'system');

CREATE TABLE role_permissions (
    role_id         UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id   UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- ============================================
-- 2. 服务器管理
-- ============================================

CREATE TABLE server_groups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    parent_id       UUID REFERENCES server_groups(id),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE servers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    host            VARCHAR(255) NOT NULL,
    port            INTEGER DEFAULT 22,
    username        VARCHAR(100) DEFAULT 'root',
    auth_type       VARCHAR(20) DEFAULT 'password',  -- password, key
    password_enc    TEXT,                            -- 加密存储
    private_key_enc TEXT,                            -- 加密存储
    agent_connected BOOLEAN DEFAULT false,
    agent_token     VARCHAR(255),                    -- Agent 注册令牌
    group_id        UUID REFERENCES server_groups(id),
    tags            JSONB DEFAULT '[]',
    os_info         JSONB,                           -- {distribution, kernel, arch}
    hardware_info   JSONB,                           -- {cpu_model, cpu_cores, memory_gb, disk_gb}
    status          VARCHAR(20) DEFAULT 'unknown',   -- online, offline, unknown
    last_heartbeat  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_servers_status ON servers(status);
CREATE INDEX idx_servers_group ON servers(group_id);

-- ============================================
-- 3. Docker 管理
-- ============================================

CREATE TABLE docker_hosts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID REFERENCES servers(id) ON DELETE CASCADE,
    name            VARCHAR(100),
    socket_path     VARCHAR(255) DEFAULT '/var/run/docker.sock',
    api_url         VARCHAR(255),                    -- 远程 Docker API
    version         VARCHAR(50),
    running         BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE docker_containers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    docker_host_id  UUID REFERENCES docker_hosts(id) ON DELETE CASCADE,
    container_id    VARCHAR(12) NOT NULL,            -- Docker 短 ID
    name            VARCHAR(255) NOT NULL,
    image           VARCHAR(255),
    command         TEXT,
    state           VARCHAR(20),                     -- running, stopped, paused, exited
    status          TEXT,                            -- Up 2 hours
    ports           JSONB,                           -- [{host_port, container_port, protocol}]
    mounts          JSONB,
    env             JSONB,
    network_mode    VARCHAR(50),
    restart_policy  VARCHAR(50),
    cpu_usage       DECIMAL(5,2),
    memory_usage    BIGINT,
    memory_limit    BIGINT,
    network_rx      BIGINT,
    network_tx      BIGINT,
    health          VARCHAR(20),                     -- healthy, unhealthy, starting, none
    last_inspected  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_containers_host ON docker_containers(docker_host_id);
CREATE INDEX idx_containers_state ON docker_containers(state);

CREATE TABLE docker_images (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    docker_host_id  UUID REFERENCES docker_hosts(id) ON DELETE CASCADE,
    image_id        VARCHAR(12) NOT NULL,
    repo_tags       JSONB,
    size            BIGINT,
    created         TIMESTAMPTZ,
    last_inspected  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE docker_volumes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    docker_host_id  UUID REFERENCES docker_hosts(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    mountpoint      VARCHAR(500),
    driver          VARCHAR(50),
    size            BIGINT,
    last_inspected  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE docker_networks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    docker_host_id  UUID REFERENCES docker_hosts(id) ON DELETE CASCADE,
    network_id      VARCHAR(12) NOT NULL,
    name            VARCHAR(100),
    driver          VARCHAR(50),
    scope           VARCHAR(50),
    subnet          VARCHAR(50),
    gateway         VARCHAR(50),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE compose_projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    docker_host_id  UUID REFERENCES docker_hosts(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    working_dir     VARCHAR(500),                    -- Compose 文件所在目录
    file_path       VARCHAR(500),                    -- docker-compose.yml 路径
    project_name    VARCHAR(100),                    -- Docker Compose project name
    containers      JSONB,                           -- 关联的容器列表
    status          VARCHAR(20) DEFAULT 'stopped',   -- running, stopped, partial
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- 4. AI Provider 与模型
-- ============================================

CREATE TABLE ai_providers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,           -- OpenAI, Anthropic, Gemini, DeepSeek, Ollama
    display_name    VARCHAR(100),
    base_url        VARCHAR(500),
    api_key_enc     TEXT,                            -- 加密存储
    api_type        VARCHAR(50),                     -- openai-compatible, anthropic, gemini
    is_enabled      BOOLEAN DEFAULT true,
    is_default      BOOLEAN DEFAULT false,
    config          JSONB,                           -- 额外配置
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE ai_models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id     UUID REFERENCES ai_providers(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,           -- 模型名称
    model_id        VARCHAR(100) NOT NULL,           -- API 调用时的 model 参数
    model_type      VARCHAR(50),                     -- chat, embedding, image, audio
    max_tokens      INTEGER,
    supports_stream BOOLEAN DEFAULT true,
    supports_tools  BOOLEAN DEFAULT false,
    is_enabled      BOOLEAN DEFAULT true,
    is_default      BOOLEAN DEFAULT false,
    cost_per_1k_input  DECIMAL(10,6),
    cost_per_1k_output DECIMAL(10,6),
    config          JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- 5. Agent
-- ============================================

CREATE TABLE agents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    model_id        UUID REFERENCES ai_models(id),
    system_prompt   TEXT,
    temperature     DECIMAL(3,2) DEFAULT 0.7,
    max_tokens      INTEGER DEFAULT 4096,
    max_iterations  INTEGER DEFAULT 10,              -- Agent Loop 最大迭代次数
    tools           JSONB DEFAULT '[]',              -- 允许使用的 Tool 列表
    tool_permission_mode VARCHAR(20) DEFAULT 'deny_all', -- allow_all, deny_list, allow_list, approval_required
    require_approval_for JSONB DEFAULT '[]',          -- 需要审批的 tool 列表
    is_enabled      BOOLEAN DEFAULT true,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE agent_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id        UUID REFERENCES agents(id),
    user_id         UUID REFERENCES users(id),
    server_id       UUID REFERENCES servers(id),
    title           VARCHAR(255),
    status          VARCHAR(20) DEFAULT 'active',    -- active, completed, error, cancelled
    context_summary TEXT,                            -- 会话上下文摘要
    token_input     INTEGER DEFAULT 0,
    token_output    INTEGER DEFAULT 0,
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    ended_at        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE agent_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID REFERENCES agent_sessions(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL,            -- user, assistant, tool, system
    content         TEXT,
    tool_call_id    VARCHAR(100),
    tool_name       VARCHAR(100),
    tool_args       JSONB,
    tool_result     JSONB,
    metadata        JSONB,                           -- token usage, timing, etc.
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_agent_messages_session ON agent_messages(session_id);
CREATE INDEX idx_agent_sessions_user ON agent_sessions(user_id);

-- ============================================
-- 6. Tool 系统
-- ============================================

CREATE TABLE tools (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) UNIQUE NOT NULL,    -- system_info, read_file, etc.
    display_name    VARCHAR(100),
    description     TEXT,
    category        VARCHAR(50),                     -- read, write, docker, exec, system
    risk_level      VARCHAR(20) NOT NULL,            -- none, low, medium, high, critical
    permission_code VARCHAR(100) UNIQUE NOT NULL,    -- READ_SYSTEM, WRITE_FILE, etc.
    parameters      JSONB,                           -- JSON Schema
    handler         VARCHAR(100),                    -- 后端 handler 标识
    is_builtin      BOOLEAN DEFAULT true,
    is_enabled      BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 预置 Tool
INSERT INTO tools (name, display_name, description, category, risk_level, permission_code, handler) VALUES
    -- 只读
    ('system_info', '系统信息', '获取操作系统信息', 'read', 'none', 'READ_SYSTEM', 'system.Info'),
    ('system_load', '系统负载', '获取系统负载', 'read', 'none', 'READ_SYSTEM', 'system.Load'),
    ('disk_usage', '磁盘使用', '获取磁盘使用情况', 'read', 'none', 'READ_SYSTEM', 'system.DiskUsage'),
    ('memory_usage', '内存使用', '获取内存使用情况', 'read', 'none', 'READ_SYSTEM', 'system.MemoryUsage'),
    ('network_info', '网络信息', '获取网络配置', 'read', 'none', 'READ_NETWORK', 'system.NetworkInfo'),
    ('process_list', '进程列表', '获取进程列表', 'read', 'none', 'READ_SYSTEM', 'system.ProcessList'),
    ('read_log', '读取日志', '读取系统/应用日志', 'read', 'low', 'READ_LOG', 'system.ReadLog'),
    ('read_file', '读取文件', '读取文件内容', 'read', 'low', 'READ_FILE', 'system.ReadFile'),
    ('docker_list', 'Docker 列表', '获取容器列表', 'read', 'none', 'READ_DOCKER', 'docker.List'),
    ('docker_inspect', 'Docker 检查', '检查容器详情', 'read', 'none', 'READ_DOCKER', 'docker.Inspect'),
    ('docker_logs', 'Docker 日志', '获取容器日志', 'read', 'none', 'READ_DOCKER', 'docker.Logs'),
    ('docker_stats', 'Docker 统计', '获取容器资源统计', 'read', 'none', 'READ_DOCKER', 'docker.Stats'),
    ('docker_images', '镜像列表', '获取镜像列表', 'read', 'none', 'READ_DOCKER', 'docker.Images'),
    ('docker_volumes', 'Volume 列表', '获取 Volume 列表', 'read', 'none', 'READ_DOCKER', 'docker.Volumes'),
    ('compose_status', 'Compose 状态', '获取 Compose 项目状态', 'read', 'none', 'READ_DOCKER', 'docker.ComposeStatus'),
    -- 写入
    ('write_file', '写入文件', '写入文件内容', 'write', 'high', 'WRITE_FILE', 'system.WriteFile'),
    ('edit_file', '编辑文件', '编辑文件内容', 'write', 'medium', 'WRITE_FILE', 'system.EditFile'),
    ('create_file', '创建文件', '创建新文件', 'write', 'medium', 'WRITE_FILE', 'system.CreateFile'),
    -- Docker 操作
    ('container_start', '启动容器', '启动 Docker 容器', 'docker', 'medium', 'DOCKER_START', 'docker.Start'),
    ('container_stop', '停止容器', '停止 Docker 容器', 'docker', 'high', 'DOCKER_STOP', 'docker.Stop'),
    ('container_restart', '重启容器', '重启 Docker 容器', 'docker', 'medium', 'DOCKER_RESTART', 'docker.Restart'),
    ('container_delete', '删除容器', '删除 Docker 容器', 'docker', 'critical', 'DOCKER_DELETE', 'docker.Delete'),
    ('image_delete', '删除镜像', '删除 Docker 镜像', 'docker', 'critical', 'DOCKER_DELETE', 'docker.DeleteImage'),
    ('volume_delete', '删除 Volume', '删除 Docker Volume', 'docker', 'critical', 'DOCKER_DELETE', 'docker.DeleteVolume'),
    ('compose_up', 'Compose Up', '启动 Compose 项目', 'docker', 'medium', 'DOCKER_COMPOSE', 'docker.ComposeUp'),
    ('compose_down', 'Compose Down', '停止 Compose 项目', 'docker', 'high', 'DOCKER_COMPOSE', 'docker.ComposeDown'),
    ('compose_update', 'Compose 更新', '更新并重新部署 Compose', 'docker', 'high', 'DOCKER_COMPOSE', 'docker.ComposeUpdate'),
    -- 执行
    ('exec_command', '执行命令', '执行单个命令', 'exec', 'high', 'EXEC_COMMAND', 'exec.Command'),
    ('exec_shell', '执行 Shell', '执行 Shell 脚本', 'exec', 'critical', 'EXEC_SHELL', 'exec.Shell'),
    -- 系统
    ('system_reboot', '重启服务器', '重启服务器', 'system', 'critical', 'SYSTEM_REBOOT', 'system.Reboot'),
    ('service_restart', '重启服务', '重启系统服务', 'system', 'high', 'SYSTEM_SERVICE', 'system.ServiceRestart'),
    ('install_package', '安装软件', '安装系统包', 'system', 'high', 'SYSTEM_INSTALL', 'system.InstallPackage'),
    ('modify_firewall', '修改防火墙', '修改防火墙规则', 'system', 'critical', 'NETWORK_MODIFY', 'system.ModifyFirewall'),
    ('network_modify', '修改网络', '修改网络配置', 'system', 'critical', 'NETWORK_MODIFY', 'system.NetworkModify'),
    ('user_permission_modify', '修改用户权限', '修改用户/权限', 'system', 'critical', 'USER_MODIFY', 'system.UserPermissionModify');

CREATE TABLE agent_tool_permissions (
    agent_id        UUID REFERENCES agents(id) ON DELETE CASCADE,
    tool_id         UUID REFERENCES tools(id) ON DELETE CASCADE,
    is_allowed      BOOLEAN DEFAULT true,
    requires_approval BOOLEAN DEFAULT false,
    PRIMARY KEY (agent_id, tool_id)
);

-- ============================================
-- 7. 任务系统
-- ============================================

CREATE TABLE tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    type            VARCHAR(50),                     -- manual, scheduled, agent, webhook
    status          VARCHAR(20) DEFAULT 'pending',   -- pending, running, completed, failed, cancelled
    priority        INTEGER DEFAULT 5,               -- 1-10
    server_id       UUID REFERENCES servers(id),
    agent_id        UUID REFERENCES agents(id),
    agent_session_id UUID REFERENCES agent_sessions(id),
    created_by      UUID REFERENCES users(id),
    scheduled_at    TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    error_message   TEXT,
    metadata        JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_server ON tasks(server_id);

CREATE TABLE task_steps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID REFERENCES tasks(id) ON DELETE CASCADE,
    step_number     INTEGER NOT NULL,
    name            VARCHAR(255),
    description     TEXT,
    tool_name       VARCHAR(100),
    tool_args       JSONB,
    status          VARCHAR(20) DEFAULT 'pending',   -- pending, running, completed, failed, skipped
    requires_approval BOOLEAN DEFAULT false,
    approved_by     UUID REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    result          JSONB,
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_task_steps_task ON task_steps(task_id);

-- ============================================
-- 8. 审批系统
-- ============================================

CREATE TABLE approval_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_type    VARCHAR(50),                     -- tool_execution, task_execution, manual
    status          VARCHAR(20) DEFAULT 'pending',   -- pending, approved, rejected, expired, cancelled
    requested_by    UUID REFERENCES users(id),
    requested_at    TIMESTAMPTZ DEFAULT NOW(),
    approved_by     UUID REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    rejected_by     UUID REFERENCES users(id),
    rejected_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    task_id         UUID REFERENCES tasks(id),
    agent_session_id UUID REFERENCES agent_sessions(id),
    server_id       UUID REFERENCES servers(id),
    tool_name       VARCHAR(100),
    tool_args       JSONB,
    dry_run_result  JSONB,                            -- Dry Run 结果
    risk_level      VARCHAR(20),
    impact_analysis JSONB,                           -- 影响分析
    expires_at      TIMESTAMPTZ DEFAULT NOW() + INTERVAL '24 hours',
    metadata        JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_approvals_status ON approval_requests(status);
CREATE INDEX idx_approvals_expires ON approval_requests(expires_at);

-- ============================================
-- 9. 审计日志
-- ============================================

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),
    username        VARCHAR(50),                     -- 冗余存储（用户删除后仍保留）
    action          VARCHAR(100) NOT NULL,           -- server.create, docker.stop, agent.execute
    resource_type   VARCHAR(50),                     -- server, docker, agent, user, system
    resource_id     VARCHAR(100),
    resource_name   VARCHAR(255),
    method          VARCHAR(10),                     -- HTTP method or internal method
    path            VARCHAR(500),
    ip_address      INET,
    user_agent      TEXT,
    status_code     INTEGER,
    request_body    JSONB,
    response_body   JSONB,
    before_state    JSONB,                           -- 操作前状态
    after_state     JSONB,                           -- 操作后状态
    duration_ms     INTEGER,
    error_message   TEXT,
    server_id       UUID,                            -- 关联的服务器
    agent_session_id UUID,                           -- 关联的 Agent 会话
    approval_id     UUID REFERENCES approval_requests(id),
    metadata        JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at);
CREATE INDEX idx_audit_server ON audit_logs(server_id);

-- 审计日志按时间分区（PostgreSQL 12+）
-- 使用 pg_partman 或手动分区

-- ============================================
-- 10. 监控
-- ============================================

CREATE TABLE monitor_metrics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id       UUID REFERENCES servers(id),
    metric_name     VARCHAR(100) NOT NULL,           -- cpu.usage, memory.usage, disk.usage
    metric_value    DOUBLE PRECISION NOT NULL,
    unit            VARCHAR(20),
    labels          JSONB,                           -- {device="/dev/sda1", interface="eth0"}
    collected_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_metrics_server ON monitor_metrics(server_id, metric_name, collected_at);
CREATE INDEX idx_metrics_time ON monitor_metrics(collected_at);

-- 时序数据 hypertable (TimescaleDB)
-- SELECT create_hypertable('monitor_metrics', 'collected_at');

CREATE TABLE alert_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    metric_name     VARCHAR(100) NOT NULL,
    condition       VARCHAR(20),                     -- gt, lt, eq, ne, gte, lte
    threshold       DOUBLE PRECISION NOT NULL,
    duration        INTEGER DEFAULT 0,               -- 持续秒数才触发
    severity        VARCHAR(20) DEFAULT 'warning',   -- info, warning, critical
    server_id       UUID REFERENCES servers(id),      -- NULL 表示全局规则
    server_group_id UUID REFERENCES server_groups(id),
    is_enabled      BOOLEAN DEFAULT true,
    notification_channels JSONB DEFAULT '[]',
    cooldown        INTEGER DEFAULT 300,             -- 冷却时间（秒）
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE alert_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id         UUID REFERENCES alert_rules(id),
    server_id       UUID REFERENCES servers(id),
    metric_name     VARCHAR(100),
    metric_value    DOUBLE PRECISION,
    threshold       DOUBLE PRECISION,
    severity        VARCHAR(20),
    status          VARCHAR(20) DEFAULT 'firing',    -- firing, resolved
    fired_at        TIMESTAMPTZ DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ,
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_alerts_status ON alert_events(status);
CREATE INDEX idx_alerts_server ON alert_events(server_id);

-- ============================================
-- 11. 系统配置
-- ============================================

CREATE TABLE system_settings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(100) UNIQUE NOT NULL,
    value           JSONB NOT NULL,
    description     TEXT,
    updated_by      UUID REFERENCES users(id),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- 12. 预置数据
-- ============================================

-- 默认管理员
INSERT INTO users (username, email, password_hash, role_id)
VALUES ('admin', 'admin@aicenter.local', '$2a$10$...', (SELECT id FROM roles WHERE name = 'superadmin'));
```

---

## 7. REST API 设计

### 7.1 设计原则

- 路径前缀：`/api/v1`
- 资源命名：复数名词（`/servers`, `/containers`）
- 状态码：标准 HTTP 状态码
- 响应格式：统一 JSON 信封
- 错误格式：统一错误结构
- 分页：`?page=1&limit=20` 或 cursor-based

### 7.2 响应格式

```json
// 成功
{
    "code": 0,
    "message": "success",
    "data": { ... }
}

// 列表
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [ ... ],
        "total": 100,
        "page": 1,
        "limit": 20
    }
}

// 错误
{
    "code": 40001,
    "message": "参数校验失败",
    "errors": [
        {"field": "host", "message": "不能为空"}
    ]
}
```

### 7.3 API 路由表

```
/api/v1/
├── POST   /auth/register          注册
├── POST   /auth/login             登录
├── POST   /auth/logout            登出
├── POST   /auth/refresh           刷新 Token
├── GET    /auth/me                当前用户信息
├── PUT    /auth/me                更新个人信息
│
├── GET    /servers                服务器列表
├── POST   /servers                添加服务器
├── GET    /servers/:id            服务器详情
├── PUT    /servers/:id            更新服务器
├── DELETE /servers/:id            删除服务器
├── POST   /servers/:id/connect    测试连接
├── GET    /servers/:id/metrics    服务器指标
├── GET    /servers/:id/processes  进程列表
├── GET    /servers/:id/terminal   Terminal 会话 ID (WS)
├── POST   /terminal/sessions      创建 terminal 会话 (PTY/WebSocket 桥)
├── GET    /terminal/sessions      列出激活的 terminal 会话
├── DELETE /terminal/sessions/:id  关闭 terminal 会话
├── POST   /servers/batch/command  批量执行 (多服务器同命令)
├── GET    /servers/groups         服务器分组
├── POST   /servers/groups         创建分组
├── PUT    /servers/groups/:id     更新分组
├── DELETE /servers/groups/:id     删除分组
│
├── GET    /docker/hosts           Docker 主机列表
├── GET    /docker/containers      容器列表 (可跨主机)
├── GET    /docker/containers/:id  容器详情
├── POST   /docker/containers/:id/start      启动 (审批)
├── POST   /docker/containers/:id/stop       停止 (审批)
├── POST   /docker/containers/:id/restart    重启
├── DELETE /docker/containers/:id            删除 (审批)
├── GET    /docker/containers/:id/logs       日志
├── GET    /docker/containers/:id/stats      实时统计 (WS)
├── GET    /docker/images           镜像列表
├── DELETE /docker/images/:id       删除镜像 (审批)
├── GET    /docker/volumes          Volume 列表
├── DELETE /docker/volumes/:id      删除 Volume (审批)
├── GET    /docker/networks         网络列表
├── GET    /docker/compose          Compose 项目列表
├── POST   /docker/compose          创建 Compose 项目
├── PUT    /docker/compose/:id      更新 Compose
├── POST   /docker/compose/:id/up   启动 Compose (审批)
├── POST   /docker/compose/:id/down 停止 Compose (审批)
│
├── GET    /ai/providers           Provider 列表
├── POST   /ai/providers           添加 Provider
├── PUT    /ai/providers/:id       更新 Provider
├── DELETE /ai/providers/:id       删除 Provider
├── POST   /ai/providers/:id/test  测试连接
├── GET    /ai/models               模型列表
├── POST   /ai/models               添加模型
├── PUT    /ai/models/:id           更新模型
├── DELETE /ai/models/:id           删除模型
├── POST   /ai/chat                 对话（直接调用）
├── POST   /ai/chat/stream          流式对话 (SSE)
│
├── GET    /agents                  Agent 列表
├── POST   /agents                  创建 Agent
├── GET    /agents/:id              Agent 详情
├── PUT    /agents/:id              更新 Agent
├── DELETE /agents/:id              删除 Agent
├── POST   /agents/:id/sessions     创建会话
├── GET    /agents/sessions         会话列表
├── GET    /agents/sessions/:id     会话详情
├── POST   /agents/sessions/:id/messages  发送消息
├── DELETE /agents/sessions/:id     关闭会话
│
├── GET    /tasks                   任务列表
├── POST    /tasks                  创建任务
├── GET    /tasks/:id               任务详情
├── PUT    /tasks/:id               更新任务
├── DELETE /tasks/:id               删除任务
├── POST   /tasks/:id/execute       执行任务
├── POST   /tasks/:id/cancel        取消任务
│
├── GET    /approvals               审批列表
├── POST    /approvals              创建审批
├── GET    /approvals/:id           审批详情
├── POST    /approvals/:id/approve  批准
├── POST    /approvals/:id/reject   拒绝
│
├── GET    /monitor/metrics         指标查询
├── GET    /monitor/metrics/latest  最新指标
├── GET    /monitor/alert-rules     告警规则
├── POST    /monitor/alert-rules    创建告警规则
├── PUT    /monitor/alert-rules/:id 更新告警规则
├── DELETE /monitor/alert-rules/:id 删除告警规则
├── GET    /monitor/alerts          告警事件
├── POST    /monitor/alerts/:id/ack  确认告警
│
├── GET    /audit-logs              审计日志
├── GET    /audit-logs/:id          审计详情
│
├── GET    /users                   用户列表
├── POST    /users                  创建用户
├── PUT    /users/:id               更新用户
├── DELETE /users/:id               删除用户
├── GET    /roles                   角色列表
├── POST    /roles                  创建角色
├── PUT    /roles/:id               更新角色
├── DELETE /roles/:id               删除角色
│
├── GET    /settings                系统配置
├── PUT    /settings                更新系统配置
│
└── WS     /ws                      WebSocket 端点
```

### 7.4 关键 API 设计细节

#### 7.4.1 Agent 对话 (SSE/Webstream)

```
POST /api/v1/agents/sessions/:id/messages
Content-Type: application/json

{
    "content": "帮我看一下这台服务器为什么 CPU 高",
    "stream": true
}

Response (SSE):
event: message
data: {"type": "thinking", "content": "我来分析一下..."}

event: message
data: {"type": "tool_call", "tool": "system_load", "args": {}}

event: message
data: {"type": "tool_result", "tool": "system_load", "result": {"load1": 3.5, ...}}

event: message
data: {"type": "text", "content": "CPU 负载偏高，主要是因为..."}

event: approval_required
data: {
    "approval_id": "uuid",
    "tool": "exec_command",
    "args": {"command": "systemctl restart nginx"},
    "dry_run": {...},
    "risk_level": "high"
}

event: message
data: {"type": "result", "content": "执行完成"}
```

#### 7.4.2 Dry Run 预览

```
POST /api/v1/agents/sessions/:id/messages
{
    "content": "清理 Docker",
    "dry_run": true
}

Response:
{
    "code": 0,
    "data": {
        "execution_plan": {
            "steps": [
                {
                    "step": 1,
                    "tool": "docker.List",
                    "description": "列出所有容器",
                    "risk_level": "none"
                },
                {
                    "step": 2,
                    "tool": "container_stop",
                    "description": "停止 my-app 容器",
                    "risk_level": "high",
                    "impact": "服务将不可用",
                    "command": "docker stop my-app"
                },
                {
                    "step": 3,
                    "tool": "container_delete",
                    "description": "删除 my-app 容器",
                    "risk_level": "critical",
                    "impact": "容器将被永久删除，数据可能丢失",
                    "command": "docker rm my-app"
                }
            ],
            "requires_approval": true,
            "approval_id": "uuid"
        }
    }
}
```

---

## 8. WebSocket 协议设计

### 8.1 连接

```
WS /api/v1/ws?token=<JWT_TOKEN>
```

### 8.2 消息帧格式

```json
{
    "id": "msg-uuid",
    "type": "event_type",
    "channel": "channel_name",
    "timestamp": "2026-08-21T10:00:00Z",
    "data": { ... }
}
```

### 8.3 事件类型 (Server → Client)

| 事件 | Channel | 说明 |
|------|---------|------|
| `connected` | `system` | 连接成功 |
| `server.status` | `server.:id` | 服务器在线状态变化 |
| `server.metric` | `server.:id` | 实时指标推送 |
| `server.heartbeat` | `server.:id` | 心跳 |
| `docker.event` | `server.:id` | Docker 事件 (start/stop/die) |
| `container.stats` | `container.:id` | 容器实时统计 |
| `container.log` | `container.:id` | 容器日志流 |
| `agent.message` | `agent_session.:id` | Agent 消息流 |
| `agent.tool_call` | `agent_session.:id` | Agent Tool 调用 |
| `agent.thinking` | `agent_session.:id` | Agent 思考过程 |
| `agent.approval_required` | `agent_session.:id` | 需要审批 |
| `agent.status` | `agent_session.:id` | Agent 会话状态变化 |
| `task.progress` | `task.:id` | 任务进度 |
`task.step_update` | `task.:id` | 任务步骤更新 |
| `approval.requested` | `approvals` | 新审批请求 |
| `approval.resolved` | `approvals` | 审批结果 |
| `alert.fired` | `alerts` | 告警触发 |
| `alert.resolved` | `alerts` | 告警恢复 |
| `notification` | `user.:id` | 系统通知 |
| `terminal.output` | `terminal.:id` | Terminal 输出 |

### 8.4 事件类型 (Client → Server)

| 事件 | 说明 |
|------|------|
| `subscribe` | 订阅频道 |
| `unsubscribe` | 取消订阅 |
| `ping` | 心跳 |
| `terminal.input` | Terminal 输入 |
| `terminal.resize` | Terminal 窗口大小调整 |
| `approval.approve` | 批准审批 |
| `approval.reject` | 拒绝审批 |

### 8.5 订阅示例

```json
// Client → Server
{
    "id": "sub-1",
    "type": "subscribe",
    "data": {
        "channels": [
            "server.550e8400-e29b-41d4-a716-446655440000",
            "agent_session.6ba7b810-9dad-11d1-80b4-00c04fd430c8",
            "task.6ba7b811-9dad-11d1-80b4-00c04fd430c8",
            "approvals",
            "alerts"
        ]
    }
}
```

### 8.6 Agent 消息流示例

```json
// 1. Agent 开始思考
{
    "type": "agent.thinking",
    "channel": "agent_session.xxx",
    "data": {
        "content": "我来分析一下服务器 CPU 高的原因...",
        "iteration": 1
    }
}

// 2. Agent 调用 Tool
{
    "type": "agent.tool_call",
    "channel": "agent_session.xxx",
    "data": {
        "tool_call_id": "call_abc123",
        "tool_name": "system_load",
        "arguments": {}
    }
}

// 3. Tool 执行结果
{
    "type": "agent.message",
    "channel": "agent_session.xxx",
    "data": {
        "role": "tool",
        "tool_call_id": "call_abc123",
        "tool_name": "system_load",
        "content": "{\"load1\": 3.5, \"load5\": 2.1, \"load15\": 1.8}"
    }
}

// 4. Agent 生成回复
{
    "type": "agent.message",
    "channel": "agent_session.xxx",
    "data": {
        "role": "assistant",
        "content": "当前系统负载偏高，1分钟负载为 3.5。让我进一步查看进程..."
    }
}

// 5. 需要审批
{
    "type": "agent.approval_required",
    "channel": "agent_session.xxx",
    "data": {
        "approval_id": "approval-uuid",
        "tool_name": "exec_command",
        "tool_arguments": {"command": "kill -9 12345"},
        "risk_level": "critical",
        "dry_run": {
            "will_execute": "kill -9 12345",
            "reason": "终止占用 CPU 最高的进程",
            "impact": "进程 12345 (nginx) 将被强制终止",
            "risk": "可能导致服务中断",
            "expected_result": "CPU 负载降低"
        }
    }
}
```

---

## 9. AI Provider 抽象层

### 9.1 接口定义

```go
package provider

// Provider 是所有 AI Provider 必须实现的接口
type Provider interface {
    // 基础信息
    Name() string
    DisplayName() string
    APIType() string  // "openai-compatible", "anthropic", "gemini"

    // 模型管理
    ListModels(ctx context.Context) ([]Model, error)
    DefaultModel() string

    // 对话
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)

    // 工具调用
    SupportsTools() bool
    ChatWithTools(ctx context.Context, req ToolChatRequest) (*ToolChatResponse, error)
    StreamChatWithTools(ctx context.Context, req ToolChatRequest) (<-chan ToolStreamChunk, error)

    // 健康检查
    HealthCheck(ctx context.Context) error
}

type Model struct {
    ID            string
    Name          string
    Type          string  // "chat", "embedding", "image"
    MaxTokens     int
    SupportsStream bool
    SupportsTools  bool
}

type ChatRequest struct {
    Model       string
    Messages    []Message
    Temperature float32
    MaxTokens   int
    Stream      bool
    Tools       []ToolDefinition  // OpenAI format tools
}

type Message struct {
    Role       string  // "user", "assistant", "system", "tool"
    Content    string
    ToolCallID string     // for tool response
    ToolCalls  []ToolCall // for assistant
}

type ToolCall struct {
    ID        string
    Type      string
    Function  FunctionCall
}

type FunctionCall struct {
    Name      string
    Arguments string  // JSON string
}

type ChatResponse struct {
    Content   string
    ToolCalls []ToolCall
    Usage     Usage
    Model     string
}

type StreamChunk struct {
    Content   string
    ToolCall  *ToolCall
    Done      bool
    Usage     *Usage
    Err       error
}

type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

### 9.2 Provider 工厂

```go
type Factory interface {
    Create(name string, config ProviderConfig) (Provider, error)
    CreateFromDB(ctx context.Context, providerID string) (Provider, error)
}

type ProviderConfig struct {
    Name       string
    BaseURL    string
    APIKey     string
    APIType    string  // openai, anthropic, gemini
    Extra      map[string]interface{}
}

// 注册
var providers = map[string]func(ProviderConfig) (Provider, error){
    "openai": func(cfg ProviderConfig) (Provider, error) {
        return openai.New(cfg)
    },
    "openai-compatible": func(cfg ProviderConfig) (Provider, error) {
        return openai.New(cfg)
    },
    "anthropic": func(cfg ProviderConfig) (Provider, error) {
        return anthropic.New(cfg)
    },
    "gemini": func(cfg ProviderConfig) (Provider, error) {
        return gemini.New(cfg)
    },
    "deepseek": func(cfg ProviderConfig) (Provider, error) {
        return openai.New(cfg)  // OpenAI 兼容
    },
    "ollama": func(cfg ProviderConfig) (Provider, error) {
        return openai.New(cfg)  // OpenAI 兼容
    },
    "mock": func(cfg ProviderConfig) (Provider, error) {
        return mock.New(cfg)     // 单元测试 / demo
    },
}
```

### 9.3 Provider 适配器说明

| Provider | 适配器 | 特殊处理 |
|----------|--------|----------|
| OpenAI | `openai-compatible` | 原生兼容 |
| Anthropic | `anthropic` | Tool 格式转换、System Prompt 分离 |
| Gemini | `gemini` | Tool 格式转换、Safety Settings |
| DeepSeek | `openai-compatible` | 配置 base_url |
| Ollama | `openai-compatible` | 配置 base_url、长超时 |
| 自定义 | `openai-compatible` | 只要兼容 OpenAI API 格式即可 |

### 9.4 统一工具调用格式

```
OpenAI Tool Format (标准格式，所有 Provider 统一到此格式):

{
    "tools": [
        {
            "type": "function",
            "function": {
                "name": "read_file",
                "description": "读取文件内容",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "文件路径"
                        }
                    },
                    "required": ["path"]
                }
            }
        }
    ]
}
```

---

## 10. Agent Tool 系统

### 10.1 Tool 接口定义

```go
type Tool interface {
    // 元信息
    Name() string
    DisplayName() string
    Description() string
    Category() string  // "read", "write", "docker", "exec", "system"

    // 权限
    RiskLevel() RiskLevel      // none, low, medium, high, critical
    PermissionCode() string    // READ_SYSTEM, WRITE_FILE, etc.

    // 参数
    Parameters() jsonschema.Schema

    // 执行
    Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
    DryRun(ctx context.Context, args json.RawMessage) (*DryRunResult, error)
}

type ToolResult struct {
    Success bool
    Data    interface{}
    Message string
    Metadata map[string]interface{}
}

type DryRunResult struct {
    WillExecute string      // 将要执行的操作
    Reason      string      // 为什么执行
    Impact      string      // 影响什么
    Risk        string      // 风险描述
    Command     string      // 具体命令/参数
    Expected    string      // 预计结果
    CanUndo     bool        // 是否可撤销
    UndoCommand string      // 撤销命令（如果可撤销）
}
```

### 10.2 Tool 分类与权限

| 分类 | Tool | 风险等级 | 权限码 | 需审批 |
|------|------|----------|--------|--------|
| **只读** | system_info | none | READ_SYSTEM | 否 |
| | system_load | none | READ_SYSTEM | 否 |
| | disk_usage | none | READ_SYSTEM | 否 |
| | memory_usage | none | READ_SYSTEM | 否 |
| | network_info | none | READ_NETWORK | 否 |
| | process_list | none | READ_SYSTEM | 否 |
| | read_log | low | READ_LOG | 否 |
| | read_file | low | READ_FILE | 否 |
| | docker_list | none | READ_DOCKER | 否 |
| | docker_inspect | none | READ_DOCKER | 否 |
| | docker_logs | none | READ_DOCKER | 否 |
| | docker_stats | none | READ_DOCKER | 否 |
| | docker_images | none | READ_DOCKER | 否 |
| | docker_volumes | none | READ_DOCKER | 否 |
| | compose_status | none | READ_DOCKER | 否 |
| **写入** | write_file | high | WRITE_FILE | 是 |
| | edit_file | medium | WRITE_FILE | 是 |
| | create_file | medium | WRITE_FILE | 是 |
| **Docker** | container_start | medium | DOCKER_START | 否 |
| | container_stop | high | DOCKER_STOP | 是 |
| | container_restart | medium | DOCKER_RESTART | 否 |
| | container_delete | critical | DOCKER_DELETE | 是 |
| | image_delete | critical | DOCKER_DELETE | 是 |
| | volume_delete | critical | DOCKER_DELETE | 是 |
| | compose_up | medium | DOCKER_COMPOSE | 否 |
| | compose_down | high | DOCKER_COMPOSE | 是 |
| | compose_update | high | DOCKER_COMPOSE | 是 |
| **执行** | exec_command | high | EXEC_COMMAND | 是 |
| | exec_shell | critical | EXEC_SHELL | 是 |
| **系统** | system_reboot | critical | SYSTEM_REBOOT | 是 |
| | service_restart | high | SYSTEM_SERVICE | 是 |
| | install_package | high | SYSTEM_INSTALL | 是 |
| | modify_firewall | critical | NETWORK_MODIFY | 是 |
| | network_modify | critical | NETWORK_MODIFY | 是 |
| | user_permission_modify | critical | USER_MODIFY | 是 |

### 10.3 Tool 执行流程

```
用户请求
    ↓
Agent Runtime 接收
    ↓
LLM 分析 → 决定调用 Tool
    ↓
Tool Registry 查找 Tool
    ↓
检查 Agent 是否有权使用此 Tool
    ↓
检查 Tool 风险等级
    ↓
┌─────────────────────────────────────┐
│ 风险等级 = none/low                  │
│ → 直接执行                           │
├─────────────────────────────────────┤
│ 风险等级 = medium/high/critical       │
│ → 生成 Dry Run                       │
│ → 创建 Approval Request              │
│ → 推送审批通知 (WebSocket)            │
│ → 等待审批结果                        │
│   ├─ 批准 → 执行                     │
│   └─ 拒绝 → 返回拒绝原因              │
└─────────────────────────────────────┘
    ↓
执行 Tool
    ↓
记录 Audit Log
    ↓
返回结果给 Agent
    ↓
Agent 继续推理
```

### 10.4 Tool 沙箱设计

```go
// 高风险 Tool 在沙箱中执行
type SandboxConfig struct {
    Timeout         time.Duration   // 执行超时
    MaxOutputSize   int64          // 最大输出大小
    AllowedPaths    []string       // 允许读写的路径
    DeniedPaths     []string       // 禁止访问的路径
    MaxMemoryMB     int64          // 内存限制
    NetworkAccess   bool           // 是否允许网络
    AllowPrivileged bool           // 是否允许特权操作
}

// 默认沙箱配置
var DefaultSandbox = SandboxConfig{
    Timeout:       30 * time.Second,
    MaxOutputSize: 10 * 1024 * 1024,  // 10MB
    DeniedPaths:   []string{"/etc/shadow", "/etc/passwd", "/root/.ssh"},
    MaxMemoryMB:   256,
    NetworkAccess: false,
}
```

---

## 11. 权限模型

### 11.1 RBAC + Tool Permission 混合模型

```
┌──────────────────────────────────────────────┐
│                   User                        │
│                     │                         │
│                     ▼                         │
│                  Role                          │
│               ┌─────┴─────┐                   │
│               │           │                   │
│               ▼           ▼                   │
│       System Perms   Tool Perms               │
│       (server:read)  (READ_SYSTEM)            │
│       (docker:manage) (DOCKER_STOP)           │
│       (agent:execute) (EXEC_COMMAND)          │
└──────────────────────────────────────────────┘
```

### 11.2 权限层级

| 层级 | 说明 | 示例 |
|------|------|------|
| System Permission | 系统级功能访问 | `server:read`, `docker:manage` |
| Tool Permission | Agent Tool 使用权限 | `READ_SYSTEM`, `DOCKER_STOP` |
| Resource Permission | 特定资源访问 | `server:server-uuid:manage` |
| Approval Permission | 审批权限 | `approval:approve` |

### 11.3 角色权限矩阵

| 权限 | SuperAdmin | Admin | Operator | Viewer |
|------|-----------|-------|----------|--------|
| server:read | ✅ | ✅ | ✅ | ✅ |
| server:create | ✅ | ✅ | ❌ | ❌ |
| server:update | ✅ | ✅ | ❌ | ❌ |
| server:delete | ✅ | ❌ | ❌ | ❌ |
| server:terminal | ✅ | ✅ | ✅ | ❌ |
| docker:read | ✅ | ✅ | ✅ | ✅ |
| docker:manage | ✅ | ✅ | ✅ | ❌ |
| docker:compose | ✅ | ✅ | ✅ | ❌ |
| ai:read | ✅ | ✅ | ✅ | ✅ |
| ai:provider | ✅ | ✅ | ❌ | ❌ |
| ai:agent | ✅ | ✅ | ❌ | ❌ |
| agent:execute | ✅ | ✅ | ✅ | ❌ |
| agent:config | ✅ | ✅ | ❌ | ❌ |
| task:read | ✅ | ✅ | ✅ | ✅ |
| task:create | ✅ | ✅ | ✅ | ❌ |
| task:manage | ✅ | ✅ | ❌ | ❌ |
| audit:read | ✅ | ✅ | ❌ | ❌ |
| approval:approve | ✅ | ✅ | ❌ | ❌ |
| approval:read | ✅ | ✅ | ✅ | ❌ |
| monitor:read | ✅ | ✅ | ✅ | ✅ |
| monitor:alert | ✅ | ✅ | ❌ | ❌ |
| system:settings | ✅ | ❌ | ❌ | ❌ |
| system:user | ✅ | ✅ | ❌ | ❌ |
| system:role | ✅ | ❌ | ❌ | ❌ |

### 11.4 权限检查中间件

```go
// 系统权限检查
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetCurrentUser(c)
        if !user.HasPermission(permission) {
            c.AbortWithStatusJSON(403, ErrorResponse("无权限"))
            return
        }
        c.Next()
    }
}

// Tool 权限检查
func RequireToolPermission(agentID string, toolName string) error {
    // 1. 检查 Agent 是否有权使用此 Tool
    // 2. 检查用户是否有此 Tool 的权限
    // 3. 检查是否需要审批
}
```

---

## 12. 审批模型

### 12.1 审批流程

```
Agent 生成 Execution Plan
        ↓
检查每个 Step 的风险等级
        ↓
┌──────────────────────────────────────┐
│ 存在 medium+ 风险操作？               │
│                                      │
│ 否 → 直接执行                        │
│ 是 → 进入审批流程                     │
└──────────────────────────────────────┘
        ↓
创建 ApprovalRequest
        ↓
推送审批通知 (WebSocket + 可选邮件)
        ↓
审批人收到通知
        ↓
审批人查看详情 (Dry Run 结果)
        ↓
┌──────────────────────────────────────┐
│ 审批人决定：                          │
│   ├─ 批准 → 执行操作                  │
│   ├─ 拒绝 → 返回拒绝原因              │
│   ├─ 修改 → 修改参数后执行            │
│   └─ 超时 → 自动取消                  │
└──────────────────────────────────────┘
        ↓
记录 Audit Log
        ↓
返回结果给 Agent
```

### 12.2 审批策略

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| `auto_approve` | 自动批准 | 低风险操作、测试环境 |
| `require_approval` | 需要审批 | 高风险操作 |
| `multi_approval` | 多人审批 | 关键操作（如生产重启） |
| `time_limited` | 限时审批 | 超时自动取消 |
| `escalation` | 升级机制 | 审批人无响应时升级 |

### 12.3 审批请求数据结构

```go
type ApprovalRequest struct {
    ID              string
    Type            string          // "tool_execution", "task_execution"
    Status          string          // "pending", "approved", "rejected", "expired"
    RequestedBy     string          // User ID
    RequestedAt     time.Time
    ApprovedBy      *string
    ApprovedAt      *time.Time
    TaskID          *string
    AgentSessionID  *string
    ServerID        *string
    ToolName        string
    ToolArgs        json.RawMessage
    DryRunResult    *DryRunResult
    RiskLevel       string
    ImpactAnalysis  *ImpactAnalysis
    ExpiresAt       time.Time
}

type ImpactAnalysis struct {
    AffectedServices   []string
    AffectedUsers      int
    DataLossRisk       bool
    DowntimeRisk       bool
    Reversible         bool
    RollbackPlan       string
}
```

### 12.4 审批规则配置

```yaml
approval:
  default_timeout: 24h
  auto_reject_after: 72h
  rules:
    - risk_level: critical
      require_approval: true
      min_approvers: 2
      escalate_after: 2h
    - risk_level: high
      require_approval: true
      min_approvers: 1
      escalate_after: 4h
    - risk_level: medium
      require_approval: true
      min_approvers: 1
    - risk_level: low
      require_approval: false
    - risk_level: none
      require_approval: false
```

---

## 13. Audit Log 模型

### 13.1 审计范围

| 操作类型 | 是否审计 | 记录内容 |
|----------|----------|----------|
| 用户登录/登出 | ✅ | IP、UA、时间、结果 |
| CRUD 操作 | ✅ | 前后状态变化 |
| Agent 执行 | ✅ | 完整会话、Tool 调用、结果 |
| Docker 操作 | ✅ | 操作类型、容器、结果 |
| 系统命令 | ✅ | 命令、参数、输出 |
| 审批操作 | ✅ | 审批人、决定、原因 |
| API 调用 | ✅ | 请求/响应摘要 |
| 配置变更 | ✅ | 变更前后值 |

### 13.2 审计日志结构

```go
type AuditLog struct {
    ID              string
    UserID          *string
    Username        string          // 冗余
    Action          string          // "server.create", "docker.stop", "agent.execute"
    ResourceType    string          // "server", "docker", "agent", "user"
    ResourceID      *string
    ResourceName    *string
    Method          string          // HTTP method
    Path            string
    IPAddress       net.IP
    UserAgent       string
    StatusCode      int
    RequestBody     json.RawMessage
    ResponseBody    json.RawMessage
    BeforeState     json.RawMessage // 操作前
    AfterState      json.RawMessage // 操作后
    DurationMs      int
    ErrorMessage    *string
    ServerID        *string         // 关联服务器
    AgentSessionID  *string         // 关联 Agent 会话
    ApprovalID      *string         // 关联审批
    Metadata        json.RawMessage
    CreatedAt       time.Time
}
```

### 13.3 审计日志保留策略

```yaml
audit:
  retention:
    default: 90d
    security: 365d
    system: 30d
  storage:
    primary: postgresql
    archive: s3  # 过期后归档到 S3
  export:
    formats: [json, csv, pdf]
```

---

## 14. 多服务器模型

### 14.1 架构

被管节点**无常驻 AICenter 进程**。控制面维护 `servers` 表（含 SSH 凭据），需要操作节点时通过 **SSH Bridge** 主动发起连接。

```
                    ┌──────────────┐
                    │   Backend    │
                    │ (Control P.) │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │ SSH Bridge   │
                    │ 控制面进程内  │
                    └──────┬───────┘
                           │ SSH (控制面发起)
              ┌────────────┼──────────────┐
              │            │              │
         ┌────▼─────┐ ┌───▼────┐ ┌───────▼───────┐
         │ Node A   │ │ Node B │ │ Node C        │
         │(无常驻)   │ │(无常驻) │ │(无常驻)        │
         └──────────┘ └────────┘ └───────────────┘
```

### 14.2 服务器注册流程

```
1. 用户在 AICenter 添加服务器，填入 SSH 信息（Host / Port / 用户 / 密钥或密码）
2. Backend 校验 SSH 连通性（执行 ssh user@host 'echo OK'）
3. 校验通过后，SSH 凭据经 AES-GCM 加密写入 servers 表
4. Server 状态置为 online，开始被 Monitor 通过 SSH 采集指标
```

### 14.3 SSH Bridge 通信

```go
// Server（被管节点）模型 —— 凭据加密存储
type Server struct {
    ID                 string
    Name               string
    Host               string
    Port               int
    User               string
    SSHPrivateKeyEnc   []byte   // AES-GCM 加密后的私钥
    SSHPassphraseEnc   []byte   // 可选
    Status             string   // online | offline | unknown
    LastSeenAt         *time.Time
    CreatedAt, UpdatedAt time.Time
}

// Bridge 执行结果
type SSHExecResult struct {
    ServerID    string
    Command     string
    ExitCode    int
    Stdout      string
    Stderr      string
    Duration    time.Duration
    AuditLogID  string
}
```

### 14.4 服务器分组

```go
type ServerGroup struct {
    ID          string
    Name        string
    Description string
    ParentID    *string         // 支持树形分组
    Tags        []string
    CreatedAt   time.Time
}

// 分组示例:
// - Production
//   - Web Servers
//   - DB Servers
// - Staging
// - Development
```

---

## 15. Docker 管理模型

### 15.1 架构

```
┌─────────────────────────────────────────────┐
│                  Backend                     │
│  ┌─────────────────────────────────────┐    │
│  │         Docker Service               │    │
│  │  ┌───────────┐  ┌───────────────┐  │    │
│  │  │ Local     │  │ Remote        │  │    │
│  │  │ Docker    │  │ Docker        │  │    │
│  │  │ Client    │  │ Client        │  │    │
│  │  │ (Socket)  │  │ (API/TCP)     │  │    │
│  │  └───────────┘  └───────────────┘  │    │
│  └─────────────────────────────────────┘    │
│                    │                         │
│              ┌─────┴─────┐                   │
│              │   Agent   │ (备选路径)         │
│              └───────────┘                   │
└─────────────────────────────────────────────┘
```

### 15.2 Docker 连接方式

| 方式 | 适用场景 | 安全 |
|------|----------|------|
| Unix Socket | Backend 与 Docker 同机 | 中 |
| TCP + TLS | 远程 Docker | 高 |
| Agent 代理 | 所有场景 | 高 |
| SSH 隧道 | 跨网络 | 高 |

### 15.3 Docker 操作安全

```go
// 危险操作映射
var DangerousDockerOps = map[string]RiskLevel{
    "container:stop":    High,
    "container:delete":  Critical,
    "container:kill":    Critical,
    "image:delete":      Critical,
    "volume:delete":     Critical,
    "compose:down":      High,
    "compose:update":    High,
    "network:delete":    Critical,
    "system:prune":      Critical,
}

// 所有危险操作必须经过审批
func (s *DockerService) executeWithApproval(
    ctx context.Context,
    op string,
    target string,
    args interface{},
) error {
    risk := DangerousDockerOps[op]
    if risk >= High {
        return s.approvalService.Request(ctx, ApprovalRequest{
            Type:     "docker_operation",
            ToolName: op,
            RiskLevel: risk,
            ...
        })
    }
    return s.execute(ctx, op, target, args)
}
```

### 15.4 Compose 管理

```go
type ComposeProject struct {
    ID           string
    DockerHostID string
    Name         string
    WorkingDir   string
    FilePath     string
    ProjectName  string
    Containers   []string
    Status       string
    Content      string  // docker-compose.yml 内容
}

// 操作:
// - Deploy (up -d)
// - Update (pull + up -d)
// - Stop (down)
// - Restart
// - Scale
// - View Logs
// - Edit Config
```

---

## 16. 监控模型

### 16.1 指标采集

```go
type Metric struct {
    ServerID    string
    Name        string      // "cpu.usage", "memory.usage", "disk.io"
    Value       float64
    Unit        string      // "%", "bytes", "ops/s"
    Labels      map[string]string
    CollectedAt time.Time
}

// 采集指标列表
var SystemMetrics = []string{
    "cpu.usage",           // CPU 使用率
    "cpu.load.1",          // 1分钟负载
    "cpu.load.5",          // 5分钟负载
    "cpu.load.15",         // 15分钟负载
    "memory.usage",        // 内存使用率
    "memory.used",         // 已用内存
    "memory.available",    // 可用内存
    "disk.usage",          // 磁盘使用率
    "disk.io.read",        // 磁盘读
    "disk.io.write",       // 磁盘写
    "network.rx",          // 网络接收
    "network.tx",          // 网络发送
    "network.connections", // 连接数
    "process.count",       // 进程数
    "swap.usage",          // Swap 使用率
}

var DockerMetrics = []string{
    "docker.container.cpu",
    "docker.container.memory",
    "docker.container.network_rx",
    "docker.container.network_tx",
    "docker.container.block_io",
}
```

### 16.2 告警规则

```go
type AlertRule struct {
    ID              string
    Name            string
    MetricName      string
    Condition       string      // "gt", "lt", "gte", "lte"
    Threshold       float64
    Duration        int         // 持续秒数
    Severity        string      // "info", "warning", "critical"
    ServerID        *string     // NULL = 全局
    ServerGroupID   *string
    IsEnabled       bool
    NotificationChannels []string
    Cooldown        int
}

// 示例规则
// - CPU > 80% 持续 5 分钟 → warning
// - CPU > 95% 持续 2 分钟 → critical
// - 内存 > 90% 持续 3 分钟 → critical
// - 磁盘 > 85% 持续 10 分钟 → warning
// - 容器退出 → critical
```

### 16.3 监控数据流

```
Agent Collector (每15s)
    ↓
Backend WebSocket Hub
    ↓
┌───────────────────┐
│  Monitor Service   │
│  ├─ 存储 (DB)      │
│  ├─ 告警检查       │
│  └─ 实时推送 (WS)  │
└───────────────────┘
    ↓
Frontend (实时图表)
```

### 16.4 存储策略

| 数据 | 保留期 | 粒度 |
|------|--------|------|
| 原始指标 | 7 天 | 15s |
| 1 分钟聚合 | 30 天 | 1 分钟 |
| 5 分钟聚合 | 90 天 | 5 分钟 |
| 1 小时聚合 | 1 年 | 1 小时 |

---

## 17. 任务模型

### 17.1 任务类型

| 类型 | 说明 | 触发方式 |
|------|------|----------|
| `manual` | 手动创建 | 用户点击 |
| `scheduled` | 定时任务 | Cron 表达式 |
| `agent` | Agent 创建 | Agent 自动 |
| `webhook` | Webhook 触发 | 外部系统 |
| `event` | 事件触发 | 告警/状态变化 |

### 17.2 任务结构

```go
type Task struct {
    ID          string
    Title       string
    Description string
    Type        string
    Status      string          // pending, running, completed, failed, cancelled
    Priority    int             // 1-10
    ServerID    *string
    AgentID     *string
    CreatedBy   string
    ScheduledAt *time.Time
    StartedAt   *time.Time
    CompletedAt *time.Time
    Steps       []TaskStep
    ErrorMessage *string
    Metadata    map[string]interface{}
}

type TaskStep struct {
    ID              string
    TaskID          string
    StepNumber      int
    Name            string
    Description     string
    ToolName        string
    ToolArgs        json.RawMessage
    Status          string      // pending, running, completed, failed, skipped
    RequiresApproval bool
    ApprovedBy      *string
    ApprovedAt      *time.Time
    StartedAt       *time.Time
    CompletedAt     *time.Time
    Result          interface{}
    ErrorMessage    *string
}
```

### 17.3 任务执行引擎

```go
type TaskExecutor struct {
    workerCount int
    queue       chan *Task
    workers     []*Worker
}

// 工作池模式
// - 默认 5 个 worker
// - 支持并发执行多个任务
// - 每个任务内的步骤串行执行
// - 需要审批的步骤暂停等待
```

### 17.4 定时任务

```go
type ScheduledTask struct {
    ID          string
    TaskTemplateID string
    CronExpr    string
    Timezone    string
    IsEnabled   bool
    LastRunAt   *time.Time
    NextRunAt   *time.Time
    RunCount    int
}

// 使用 robfig/cron 或类似库
```

---

## 18. 部署方案

### 18.1 开发环境

```yaml
# docker-compose.dev.yml
version: "3.8"

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ./backend:/app
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DATABASE_URL=sqlite:///data/aicenter.db
      - JWT_SECRET=dev-secret-key
      - LOG_LEVEL=debug
    depends_on:
      - redis

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.dev
    ports:
      - "5173:5173"
    volumes:
      - ./frontend:/app
    environment:
      - VITE_API_URL=http://localhost:8080

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # 开发用 PostgreSQL (可选)
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: aicenter
      POSTGRES_USER: aicenter
      POSTGRES_PASSWORD: aicenter
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

### 18.2 生产环境

```yaml
# docker-compose.yml
version: "3.8"

services:
  api:
    build:
      context: ./backend
      dockerfile: Dockerfile
    restart: always
    environment:
      - DATABASE_URL=postgresql://aicenter:${DB_PASSWORD}@postgres:5432/aicenter
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - LOG_LEVEL=info
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - app-data:/data
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 512M

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf
      - ./certs:/etc/nginx/certs
    depends_on:
      - api

  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_DB: aicenter
      POSTGRES_USER: aicenter
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U aicenter"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  pgdata:
  redis-data:
  app-data:
```

### 18.3 节点接入方式

被管节点**不部署任何 AICenter 进程**。接入方式仅为 SSH 凭据注册：

```bash
# 1. 在控制面 UI/API 上添加服务器
curl -X POST https://aicenter.example.com/api/v1/servers \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "node-001",
    "host": "10.0.0.1",
    "port": 22,
    "user": "deploy",
    "ssh_private_key": "<ENCRYPTED-SHELL>"
  }'

# 2. 控制面自动执行 SSH 连通性校验
# 3. 成功后节点状态=online，Monitor 开始通过 SSH 采集指标
```

> 备注：被管节点只需开放 22/TCP 端口，无需安装 Agent、无需 systemd 服务、无需 Docker。

### 18.4 部署拓扑

```
┌─────────────────────────────────────────────────────────┐
│                     Load Balancer                        │
│                   (Nginx / Traefik)                      │
└─────────────────────┬───────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
  ┌─────▼─────┐ ┌────▼──────┐ ┌────▼──────┐
  │ Frontend  │ │ Frontend  │ │ Frontend  │
  │  (Nginx)  │ │  (Nginx)  │ │  (Nginx)  │
  └─────┬─────┘ └────┬──────┘ └────┬──────┘
        │             │             │
        └─────────────┼─────────────┘
                      │
              ┌───────▼───────┐
              │  API Gateway  │
              │   (Nginx)     │
              └───────┬───────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
  ┌─────▼─────┐ ┌────▼──────┐ ┌────▼──────┐
  │  Backend  │ │  Backend  │ │  Backend  │
  │  (API+Runtime│ │(API+Runtime│ │(API+Runtime│
  │   单实例)   │ │  单实例)  │ │  单实例)  │
  └─────┬─────┘ └────┬──────┘ └────┬──────┘
        │             │             │
        └─────────────┼─────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
  ┌─────▼─────┐ ┌────▼──────┐ ┌────▼──────┐
  │ PostgreSQL│ │   Redis   │ │  (可选)   │
  │ (Primary) │ │ (Cache)   │ │  Pub/Sub  │
  └───────────┘ └───────────┘ └───────────┘
                      │
                      │ SSH (控制面发起)
         ┌────────────┼──────────────────────┐
         │            │                      │
    ┌────▼─────┐ ┌───▼────┐        ┌───────▼───────┐
    │ Node 001 │ │Node 002│        │ Node 00N      │
    │(无常驻)   │ │(无常驻) │        │(无常驻)        │
    └──────────┘ └────────┘        └───────────────┘
```

> 生产部署建议 Backend 至少 2 实例；SSH 凭据解密后的明文密钥只在内存中存在，不写入日志/DB。

---

## 19. MVP 开发顺序

### Phase 1: 基础框架 (2-3 周)

```
目标: 跑通前后端基础框架，实现用户认证

1.1 Backend 基础
   ├── Go 项目初始化
   ├── 配置管理
   ├── 数据库连接 (SQLite)
   ├── 用户认证 (JWT)
   ├── RBAC 基础
   ├── 统一响应格式
   └── 日志系统

1.2 Frontend 基础
   ├── Vite + React + TS 项目初始化
   ├── Arco Design 集成
   ├── 路由配置
   ├── 登录/注册页面
   ├── 主布局 (Sidebar + Navbar)
   ├── Axios 封装
   └── Zustand Store 基础

1.3 部署
   ├── Docker Compose 开发环境
   └── 热重载配置
```

### Phase 2: 服务器管理 (2-3 周)

```
目标: 添加服务器，查看系统状态

2.1 Backend
   ├── Server CRUD API
   ├── Server Group CRUD
   ├── SSH 连接管理
   ├── Agent 注册协议
   ├── 心跳检测
   └── 基础指标采集

2.2 Agent
   ├── Agent 基础框架
   ├── 系统信息采集
   ├── WebSocket 通信
   └── 心跳

2.3 Frontend
   ├── 服务器列表/详情
   ├── 添加服务器表单
   ├── 服务器分组
   ├── 系统状态展示
   └── 基础监控图表
```

### Phase 3: Docker 管理 (2-3 周)

```
目标: 管理 Docker 容器、镜像、Volume、Compose

3.1 Backend
   ├── Docker Client 封装
   ├── Container CRUD
   |   ├── 列表/详情/日志/统计
   |   ├── 启动/停止/重启/删除
   ├── Image 管理
   ├── Volume 管理
   ├── Network 管理
   ├── Compose 管理
   └── Docker 事件监听

3.2 Frontend
   ├── Docker Dashboard
   ├── 容器列表/操作
   ├── 镜像管理
   ├── Volume 管理
   ├── Compose 编辑器
   └── 实时日志查看
```

### Phase 4: AI Provider 集成 (1-2 周)

```
目标: 接入多个 AI Provider，实现对话功能

4.1 Backend
   ├── Provider 抽象层
   ├── OpenAI 适配器
   ├── Anthropic 适配器
   ├── Gemini 适配器
   ├── DeepSeek 适配器
   ├── Ollama 适配器
   ├── Provider CRUD API
   ├── Model CRUD API
   └── 对话 API (流式)

4.2 Frontend
   ├── Provider 配置页面
   ├── Model 管理页面
   ├── 对话测试页面
   └── 流式输出展示
```

### Phase 5: Runtime 系统 (3-4 周)

```
目标: 在控制面进程内实现 Agent Runtime（思考-行动循环 + Tool 调用 +
     审批），并通过 SSH Bridge 把执行能力扩展到被管节点。

5.1 Backend
   ├── Runtime CRUD（Agent 配置、模型、system prompt）
   ├── Runtime Session 管理
   ├── Tool 注册框架（Read-only / Approve / Deny 分类）
   ├── Runtime Loop（思考-行动循环，进程内）
   ├── SSH Bridge 接口 + 实现（密钥管理、通道复用）
   ├── 受控执行器（risk + approval gate）
   └── Audit Log

5.2 Frontend
   ├── Runtime 配置页面
   ├── 对话界面（会话 / 消息 / Tool 步骤可视化）
   ├── 执行计划展示
   ├── 审批界面
   └── Audit Log 查看

5.3 关键设计决策
   ├── Agent 进程内化（ADR-001）：无边缘 Agent 进程
   ├── Tool 分层：只读直出 / 写前校验 / 高风险拒绝
   ├── 远程执行：SSH Bridge 通道复用，按 operation-id 聚合输出
   └── 审批事件：WebSocket 从控制面推送，前端弹窗
```

### Phase 6: 任务与监控 (2-3 周)

```
目标: 任务调度、告警、完整监控

6.1 Backend
   ├── 任务系统
   ├── 定时任务
   ├── 告警引擎
   ├── 通知系统
   └── 指标聚合

6.2 Frontend
   ├── 任务中心
   ├── 告警管理
   ├── 监控大屏
   └── 通知中心
```

### Phase 7: 完善与优化 (持续)

```
- [x] Web Terminal (PTY/WebSocket backend + xterm.js frontend)
- [x] 多服务器批量操作 (batch command service + handler + batch UI)
- [x] 性能优化 (in-process TTL+LRU cache for server reads)
- [x] 安全加固 (per-provider concurrency limiter)
- [x] 文档完善
- [x] 测试覆盖
```

#### 7.1 Web Terminal 状态

**已完成（commit `462b914`）**：
- 后端 `internal/terminal/`：跨平台会话管理器。POSIX 用 `creack/pty` 真 PTY；Windows 用 stdin/stdout pipe 多路复用 stderr。保持统一的 WebSocket 信元（`input`/`resize`/`data`/`exit`）。
- REST 路由 `/api/v1/terminal/sessions`（增/查/删） + WebSocket `/ws/terminal?session=<id>`。
- 前端 `features/servers/Terminal.tsx`（xterm.js fit/Resize） + `ServerDetailPage.tsx` 服务器详情页「概览/终端」双 Tab，点击服务器跳转。
- E2E 验证：真实 PTY 回显 `echo HELLO_TERMINAL_E2E`，状态 200，列表与关闭均 PASS。

#### 7.2 多服务器批量操作 状态

**已完成（commit `c794ccc`）**：
- 服务 `internal/service/batch_service.go`：接收 `command` + `server_ids[]`，并发执行，单 server 超时 + 进程树清理，返回 `BatchResult` 列表（`server_id / host / stdout / stderr / exit_code / duration / status / error`）。
- localhost/127.0.0.1 走本地 `exec`（可验证）；其它 host 走已存的 `pkg/ssh` 客户端。
- 路由 `POST /api/v1/servers/batch/command`，鉴权透传 `MockAuth`。
- 前端 `features/servers/BatchCommandPage.tsx`：服务器多选 + 命令输入 + 超时 + 实时结果表格；路由 `/servers/batch`，导航栏新 Tab。
- E2E：echo 回显 / 非零退出码 / 超时 surfaces 为 failure，全部 PASS。

#### 7.3 性能优化 状态

**已完成（commit `5fc8210`）**：
- 新增 `internal/pkg/cache`：依赖-free 内存 TTL+LRU `Store` 接口 (`Get/Set/Delete/Clear/Stats`)。
- 注入 `ServerService`：`ListServers`/`GetServer` 走缓存-旁路；`Create/Update/Delete` 使缓存失效。
- Benchmarks：`BenchmarkMemoryStore_GetHit` **28.37 ns/op，0 allocs**；miss 10.27 ns/op。
- 验证：在缓存层激活后，terminal + batch 两个 E2E 均仍 PASS。

#### 7.4 安全加固 状态

**已完成（commit `9babbf9`）**：
- `internal/ai/limited.go`：`NewLimited(Client, max)` 包装器，用信号量 `chan struct{}` 硬控每 provider 最大并发调用数 (`DefaultProviderConcurrency=4`)，防止 credential abuse / 429 风暴 / 单 provider 耗尽后端。
- 所有 provider client 通过 `Factory.Build` 自动被 `NewLimited` 包装，无需调用方改动。
- 单元测试 `limited_test.go`：peak concurrency == cap；超出 cap 时 context-deadline fast-fail；`max<=0` 透传。
+ 修复 `terminal.Session` JSON 序列化 (`ID`→`id`) 及 terminal E2E 列表顺序（在 WS 断连前列出会话）。

#### 7.5 文档完善 状态

**已完成（commit `9b8d29e`）**：
- README 新增 **Features** + **API** 段落，列出 Web Terminal / 批量操作 / 缓存 / 限流 4 大特性及端点。
- ARCHITECTURE §7 REST API 树补充 `/terminal/sessions` + `/servers/batch/command` 路由。
- ARCHITECTURE §19 各 Phase 状态对勾到 `[x]`，并补充 7.2/7.3/7.4 完成摘要。

#### 7.6 测试覆盖 状态

**已完成（commit `e297004`）**：
- `internal/service/batch_service_test.go`：6 个单元测试（空命令 / 本地 echo / 非零退出码 / 超时 surfaces 为 failed / 列表错错 / 并发执行）。
- `internal/ai/limited_test.go`：3 个单元测试（peak==cap / 超容 fast-fail / max<=0 透传）。
- `internal/pkg/cache/bench_test.go`：GetHit 28.37 ns/op；`go test ./internal/...` 全部 PASS。
- 回归：terminal + batch 两个 E2E 套件在缓存层激活后均 PASS。

---

## 20. 后续迭代路线

### V1.1 - 增强 Runtime
- [ ] Runtime 记忆系统 (长期记忆)
- [ ] Runtime 协作 (多 Agent 协同)
- [ ] Runtime 技能市场 (共享 Tool)
- [ ] Prompt 模板库
- [ ] 边缘 SSH Bridge 通道复用优化

### V1.2 - 增强运维
- [ ] Web Terminal (xterm.js)
- [ ] 文件管理器
- [ ] 批量操作
- [ ] 配置管理 (Ansible 集成)
- [ ] 备份管理

### V1.3 - 增强 AI
- [ ] RAG 集成 (知识库)
- [ ] Agent 自动调参
- [ ] 多模态支持 (图片理解)
- [ ] 语音交互

### V1.4 - 增强协作
- [ ] 团队协作
- [ ] 操作分享
- [ ] 评论/标注
- [ ] 操作回放

### V1.5 - 增强安全
- [ ] 2FA
- [ ] IP 白名单
- [ ] 操作录像
- [ ] 合规报告

### V2.0 - 平台化
- [ ] 插件系统
- [ ] 开放 API
- [ ] 多租户
- [ ] 移动端
- [ ] 桌面端 (Electron)
- [ ] 边缘节点进程内 Agent（见 ADR-001；MVP 暂不做，V2.0 再评估）

---

## 21. 风险与技术难点

### 21.1 高风险项

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| AI Agent 误操作导致服务中断 | 极高 | 中 | 多层审批、Dry Run、只读默认、沙箱 |
| WebSocket 连接不稳定 | 高 | 中 | 自动重连、心跳检测、状态恢复 |
| Docker Socket 安全 | 高 | 中 | TLS 认证、Agent 代理、权限最小化 |
| 凭证泄露 | 极高 | 低 | 加密存储、审计日志、定期轮换 |
| AI Provider API 限流 | 中 | 高 | 多 Provider 切换、队列、缓存 |

### 21.2 技术难点

| 难点 | 说明 | 建议方案 |
|------|------|----------|
| Agent Loop 稳定性 | LLM 可能陷入循环、幻觉调用 | 最大迭代限制、超时、人工干预 |
| 多 Provider 兼容 | 各 Provider Tool 格式不同 | 统一适配层、充分测试 |
| 实时性 | 大量 WebSocket 连接和指标推送 | 连接池、消息压缩、采样降级 |
| 审批延迟 | 人工审批阻塞自动化 | 可配置审批策略、超时自动处理 |
| 跨版本兼容 | SSH Bridge 与远端工具链版本不一致 | 通道级能力协商、特性降级、版本上报 |
| 大规模监控 | 多服务器高频指标采集 | 边缘聚合、采样、TSDB |
| 会话状态恢复 | Agent 会话中断后恢复 | 持久化上下文、断点续传 |

### 21.3 安全注意事项

```
1. Docker Socket 暴露
   - 不要将 Docker Socket 直接暴露到公网
   - 使用 Agent 代理或 TLS 认证

2. Agent Token 管理
   - Token 必须加密存储
   - 支持 Token 轮换
   - 离职/更换时立即吊销

3. AI Prompt 注入
   - 用户输入必须清洗
   - Agent System Prompt 不可被用户覆盖
   - 日志中的敏感信息脱敏

4. 命令注入
   - 所有 Shell 命令必须参数化
   - 禁止字符串拼接命令
   - 使用白名单限制可执行命令

5. 权限提升
   - Agent 不应以 root 运行
   - 使用最小权限原则
   - 敏感操作需要二次确认
```

---

## 22. 设计取舍

### 22.1 已做的取舍

| 决策 | 选择 | 放弃 | 理由 |
|------|------|------|------|
| 数据库 | PostgreSQL | MySQL, MongoDB | JSON 支持好、时序扩展、开源 |
| 缓存 | Redis | Memcached | 功能更丰富、支持 Pub/Sub |
| 通信 | WebSocket | gRPC, MQTT | 兼容性好、穿透力强 |
| 前端框架 | React | Vue, Sango | 生态好、Arco Design 支持 |
| UI 库 | Arco Design | Ant Design, Material | 更现代、TS 支持好 |
| 状态管理 | Zustand | Redux, MobX | 轻量、TS 友好 |
| Agent 部署 | **进程内 Runtime + SSH Bridge** | 独立边缘 Agent 进程、Sidecar、DaemonSet | 降低部署面（节点零常驻）、消除 Agent 进程升级/状态一致性问题；节点只需 SSH |
| 认证 | JWT | Session, OAuth | 无状态、易扩展 |
| 配置格式 | YAML | JSON, TOML | 可读性好、支持注释 |

### 22.2 需要进一步确认的取舍

| 决策 | 选项 A | 选项 B | 建议 |
|------|--------|--------|------|
| 时序数据 | PostgreSQL + 分区 | TimescaleDB / InfluxDB | 初期用 PG，后期迁移 |
| 消息队列 | Redis Pub/Sub | NATS / RabbitMQ | 初期用 Redis，后期可换 |
| Agent 通信 | SSH（控制面发起） | WebSocket Agent、gRPC Agent、节点常驻 Agent | SSH 是节点侧已存在的能力，无需额外部署；配合密钥管理天然安全 |
| 监控存储 | 自建 | Prometheus + Grafana | 建议自建 (统一体验) |
| 日志收集 | 自建 | ELK / Loki | 建议自建 (简化架构) |
| 前端图表 | ECharts | D3.js, Chart.js | 建议 ECharts (功能全) |
| 终端方案 | xterm.js | ttyd (后端渲染) | 建议 xterm.js (前端渲染) |
| 容器编排 | Docker Compose | K3s | 初期 Compose，后期可选 K3s |

### 22.3 架构演进建议

```
MVP (V1.0)
├── 单机部署
├── SQLite 数据库
├── 基础 Agent
├── 基础审批
└── 基础监控

Growth (V1.5)
├── Docker Compose 多副本
├── PostgreSQL 主从
├── Redis 集群
├── 高级 Agent (记忆、协作)
└── 完整告警

Scale (V2.0)
├── Kubernetes 部署
├── TimescaleDB
├── NATS 集群
├── 多租户
└── 插件系统
```

---

## 附录

### A. 技术栈版本建议

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.22+ | 最新稳定版 |
| React | 18.x | 最新稳定版 |
| TypeScript | 5.x | 最新稳定版 |
| Vite | 5.x | 最新稳定版 |
| Arco Design | 2.x | 最新稳定版 |
| PostgreSQL | 16.x | 最新稳定版 |
| Redis | 7.x | 最新稳定版 |
| Docker | 24+ | 最新稳定版 |

### B. 项目目录总览

```
AI_Server_Center/
├── ARCHITECTURE.md          # 本文档
├── IDEA.md                  # 项目定位
├── README.md                # 项目说明
├── .gitignore
│
├── backend/                 # Go 后端
│   ├── cmd/
│   │   └── aicenter/        # 主入口
│   ├── internal/
│   │   ├── api/             # Gin 路由、handler、middleware
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   └── router/
│   │   ├── database/        # 连接、迁移、seed
│   │   ├── models/          # 领域模型
│   │   ├── repository/      # 数据访问（server/provider/role/user 等）
│   │   ├── service/         # 业务逻辑（Server/AI/Auth/User 等）
│   │   ├── permission/      # RBAC 权限注册中心
│   │   ├── monitor/         # 指标采集、存储、告警
│   │   ├── task/            # 任务调度、执行、工作池
│   │   ├── approval/        # 审批、Dry Run、受控执行器
│   │   ├── websocket/       # WS Hub
│   │   └── runtime/         # Agent Runtime（进程内：Loop + Tool + Session）
│   ├── pkg/                 # 内部共享（crypto/ssh/docker/utils/logger）
│   ├── migrations/          # SQL 迁移
│   └── configs/             # 环境配置
│
├── frontend/                # React 前端
│   ├── src/
│   ├── public/
│   └── ...

### C. 架构决策记录 (ADR)

#### ADR-001: Agent 部署形态 — 进程内 Runtime + SSH Bridge

**状态**: Accepted · 2026-08-22

**背景**
AICenter 需要把 LLM 的"思考-行动循环"（Runtime）与"在远端节点执行命令"（Docker API / Shell）解耦。
Runtime 本身是 CPU/内存密集的应用逻辑；节点执行依赖 Docker Socket 或 SSH。

**备选方案**

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| A. 边缘独立 Agent 进程 | 每个被管节点跑一个常驻 AICenter-agent，与 Backend WS 长连 | 节点本地上下文强；可批量分发；离线也能缓存 | 部署面翻倍（N 节点 = N+1 进程要运维）；升级/配置不一致；心跳/探活/自愈工程量大 |
| B. 进程内 Runtime + SSH Bridge（本次选择）| Runtime 作为控制面 Go 进程的组件；控制面通过 SSH 主动发起连接 | 单一进程、无 Agent 部署面；版本强一致；密钥由控制面统一管理 | 控制面并发受限于进程；远端命令延迟含 SSH 握手开销；控制面成单点 |
| C. Sidecar / DaemonSet | Agent 作为容器 Sidecar 或 K8s DaemonSet | 编排原生、自动扩缩 | 绑定 K8s；小团队运维成本不划算 |
| D. 纯 Pull API | 节点提供 API，Backend 轮询 | 无常驻依赖 | 暴露面扩大；安全边界弱 |

**决策**
采用 **B** 作为 MVP（V1.0–V1.5）；**A** 留到 V2.0 再评估（当节点数 > 数百、网络分区场景明确时）。

**理由**
1. MVP 的节点规模通常在个位到数十台，进程内方案工程成本远低于独立 Agent 集群的部署/升级/探活开销。
2. 消除 Agent 与 Backend 版本不一致这一类跨版本兼容风险。
3. 密钥、审计、审批、Token 全在控制面同一内存边界，攻击面更可控。
4. 边缘 SSH 已存在，节点零部署成本。

**后果**
- ✅ 正面：部署面 -1、密钥集中、审计完整、无 Agent 探活/自愈代码
- ⚠️ 负向：控制面成单点（V1.5 起可水平扩展 Backend + Redis 共享 Session）；远端命令延迟含 SSH 握手（用通道复用缓解）
- 未来如节点数 > 500 或需要弱网容忍，需回切方案 A

**相关引用**
- `backend/internal/runtime/` — Runtime 组件目录（待实现）
- `backend/internal/pkg/ssh/` — SSH Bridge 封装
- `backend/internal/database/migrations/009_rbac.up.sql` — 权限/角色表

---

> 文档结束。此架构设计涵盖了 AICenter 项目的完整技术方案，可作为后续开发的蓝图。建议在正式开发前，对关键模块（Runtime、审批系统、安全模型）进行详细设计和评审。
