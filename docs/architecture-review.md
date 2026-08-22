# AICenter 架构评审报告

**评审对象**：`ARCHITECTURE.md`（3247 行）
**评审依据**：文档自洽性 + 与真实代码库（backend 72 .go / frontend 45 .ts/.tsx / 8 SQL 迁移）的对齐度
**日期**：2026-08-22
**评审结论**：文档结构完整、无空洞章节；作为"开发蓝图"与真实实现存在 **6 处高优 + 4 处中优 + 若干低优** 偏差。可进入评审，但不能原样作为交付验收依据。

---

## 图例

- 🔴 **High** — 文档承诺 vs 实现缺失，阻碍对齐
- 🟠 **Medium** — 文档"目标态"与"当前态"混淆，易误导
- 🟡 **Low** — 标注、命名、一致性小问题

---

## 🔴 高优先级

### H1. PostgreSQL 支持：文档承诺、代码未交付
- **文档**：§1.3/§1.4「SQLite(开发) + PostgreSQL(生产)」，§6.2 SQL Schema，§18.2 生产 compose
- **实现**：`backend/go.mod` **无 pgx 驱动**；`internal/database/database.go` 有 `postgresql:` URL 分支但未接入任何驱动（仅 import `glebarez/go-sqlite`）
- **判定**：当前为 "SQLite only + 永不生效的 postgres 分支"
- **建议**：(a) 加 `pgx` 并实现 postgres driver；或 (b) 文档降级为"仅 SQLite，PostgreSQL 为 V1.x 计划"

### H2. 认证/RBAC：文档纳入 Phase 1、路由里全 TODO
- **文档**：§11 RBAC + Tool Permission、§1「JWT + Refresh Token」、Phase 1/2 均列 Auth/RBAC
- **实现**：`router.go` 里 `/auth/login`、`/auth/register`、`/auth/refresh` 三个端点返回 `login endpoint - TODO`；`internal/auth/` 只有 `jwt.go` 骨架，**无 `rbac.go`、`permission.go`**（文档 §4 明确列出）
- **判定**：核心安全模块未完成却被 Phase 1 标成已纳入；当前靠 `MockAuth()` 硬过鉴权
- **建议**：把 `MockAuth` 明确标成"开发期占位"；登录/注册/刷新 + RBAC + 权限中间件作为 Phase 1.5 补交

### H3. 独立 Agent 进程：文档核心、代码只有库
- **文档**：§5「Agent 目录结构 + 部署方式」、§14「多服务器模型：Agent(Node1)/Agent(Node2) 边缘节点」
- **实现**：`backend/cmd/` 只有 `aicenter/` 一个入口，**无 `agent/`**；agent 逻辑全在 `internal/agent/{runtime,tools,planner,approval,permissions}` 进程内库
- **判定**：文档的"三层边缘架构"实际退化为单进程；§14 Agent 注册/心跳/通信协议未落地
- **建议**：拍板——边缘 Agent 则 Phase 1 补 `cmd/agent/main.go` + 独立部署；进程内 Agent 则重写 §1/§5/§14 改"两层"

### H4. JWT Secret 硬编码默认值（违反 AGENTS.md）
- **文档/AGENTS.md**：明确「不要硬编码 API Key / 密码」
- **实现**：`config.go:55` `getEnv("JWT_SECRET", "aIcEnTeR-...")`
- **状态**：✅ **已在本次评审中修复**（commit 待提交）
  - `config.Load()` 改为 `JWT_SECRET` 必选，缺失时 `return nil, error`
  - 新增 `config_test.go`：`TestLoad_RejectsMissingJWTSecret` + `TestLoad_AcceptsProvidedJWTSecret`
  - 未触发其他调用方改动（`main.go` 已正确处理 `Load()` 的 error）

---

## 🟠 中优先级

### M1. Redis/NATS 依赖与实现不一致
- 文档 §1.3 数据层「PostgreSQL + Redis + NATS」；实现无对应依赖，用进程内 LRU + WebSocket Hub 替代
- **建议**：补依赖，或把 §1.3 改成"当前：SQLite + in-process cache + WS；多实例扩展：PostgreSQL + Redis + NATS"

### M2. 前端目录 §3 是"完整目标态"、实现只落约 1/3
- 文档 §3 列 ~100+ 组件文件；实现 `frontend/src/components/` 仅 `editors/YamlEditor.tsx` 一个真实组件
- **建议**：§3 加"实现状态"列，或拆"已实现 / 规划"两段

### M3. 数据库迁移命名与文档不符
- 文档：`001_create_users.up/down.sql`（按 entity + 含 down）；实现：`001_init.up.sql ... 008_notifications.up.sql`（按批次、无 down）
- **建议**：统一迁移约定写入文档；决定是否支持 down 迁移（生产回滚用）

### M4. §22.2 待决取舍需评审前拍板
- Agent 通信 WS vs gRPC、是否引入消息队列、是否多机多实例 —— 直接决定 H1/H3/M1 修复方向
- **建议**：列为评审议程第一优先

---

## 🟡 低优先级

- **§20 路线图与已实现功能重叠**：V1.2 列 "Web Terminal / 批量操作"，但 Phase 7.1/7.2 已实现。更新勾选态
- **§9 Provider 适配器**：文档列 5 家；实现仅 `openai_compat.go` + `anthropic.go` + `mock.go`，Gemini/DeepSeek/Ollama 缺失。§9 标注"已实现 OpenAI/Anthropic"
- **WebSocket 端点无鉴权**：`router.go` 自注 `// TODO: auth via query param`，与 §8 不符
- **附录 B**：列顶层 `agent/` 目录，实际无此目录

---

## ✅ 亮点（文档与实现对齐良好）

- **Phase 7.1–7.6 完成状态**：逐项关联到具体 commit，可追溯 `git log`
- **Agent runtime 核心库**：`runtime`/`tools`/`planner`/`approval`/`permissions` 齐全，§10/§12 有落地
- **Docker 客户端**：`mock.go`/`real.go` 接口一致，符合 Docker SDK 优先规则
- **并发限流**：`ai/limited.go` + 单测，与 §21 风险项对应

---

## 推荐评审议程（按优先级）

1. 拍板架构形态（H3）：边缘 Agent vs 进程内 Agent → 决定 §1/§5/§14 重写方向
2. 拍板生产数据库（H1）：本迭代交付 PostgreSQL 还是标成后续计划
3. 拍板认证交付范围（H2）：RBAC 本迭代 vs 后续
4. 采纳 H4 修复（已完成，待 commit）
5. 标注文档"当前态 vs 目标态"（M1/M2/M4）

---

## 评审动作记录

| # | 动作 | 状态 |
|---|------|------|
| 1 | 评审报告落盘 `docs/architecture-review.md` | ✅ |
| 2 | H4 硬编码 JWT secret 修复 | ✅ commit `4e2f81a` |
| 3 | H2 批1 交付（真实 JWT 认证 + bcrypt + login/register/refresh + 移除 MockAuth） | ✅ commit `f4bebd3` |
| 4 | seed 默认 admin 假 hash bug 修复（改为动态 bcrypt，admin/Admin@123!） | ✅ 同上 |
| 5 | H1 PostgreSQL 兼容（pgx 接入 + SQL 方言收敛 + 双库兼容迁移） | ✅ commit `987…` |
| 6 | 后续：H2 批2（RBAC 权限中间件 + /users /roles CRUD）/ H3 文档对齐 | ⏳ 待用户决策 |

> 注：本环境无 Docker，无法实机启动 PostgreSQL 做端到端验证。H1 已完成代码层面的双库兼容
> （SQLite 独占语法全部清除、`datetime('now')`→`CURRENT_TIMESTAMP`、`INSERT OR IGNORE`→`ON CONFLICT`、
> migrate.go 命名参数、时间解析 helper），并以 SQLite 路径做完整 E2E 回归验证；
> 真实 PG 验证需在装有 docker 的开发机/CI 跑一次 `DATABASE_URL=postgresql://...` 启动即可确认。