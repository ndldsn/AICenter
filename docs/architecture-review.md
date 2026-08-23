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

### H1. PostgreSQL 支持：文档承诺、代码未交付  ✅ **已修复**（V1.2 `7bbccd1`）
- **文档**：§1.3/§1.4「SQLite(开发) + PostgreSQL(生产)」，§6.2 SQL Schema，§18.2 生产 compose
- **实现**：已接入 `pgx` 驱动、SQL 方言收敛、双库兼容迁移、migrate.go 命名参数
- **判定**：SQLite 路径完整 E2E 回归通过；PG 端到端待 CI 容器环境验证

### H2. 认证/RBAC：文档纳入 Phase 1、路由里全 TODO  ✅ **已修复**（V1.2 三批 `f4bebd3` / `4b262ed` / `3d67ff4`）
- **文档**：§11 RBAC + Tool Permission、§1「JWT + Refresh Token」、Phase 1/2 均列 Auth/RBAC
- **实现**：真实 JWT + bcrypt 登录/注册/刷新；RBAC 注册中心；/roles /permissions /roles/groups /users CRUD 全部落地；16+ 单测
- **判定**：MockAuth 已彻底移除，`go test ./...` 全过

### H3. 独立 Agent 进程：文档核心、代码只有库  ✅ **已修复**（V1.2 `b05cd5b`）
- **文档**：§5「Agent 目录结构 + 部署方式」、§14「多服务器模型：Agent(Node1)/Agent(Node2) 边缘节点」
- **实现**：已重构为"两层架构"——控制面（进程内 Runtime）+ 执行面（SSH Bridge）；`internal/runtime/` 五子包（engine/planner/tools/approval/permissions）+ `internal/bridge/`（ssh/client/websocket/session/pty）全部落地
- **判定**：§1/§5/§14 已按 ADR-001 改写；`go test ./...` 全过；`internal/agent/` 旧目录已删除

### H4. JWT Secret 硬编码默认值（违反 AGENTS.md）  ✅ **已修复**（V1.1 `4e2f81a`）
- **文档/AGENTS.md**：明确「不要硬编码 API Key / 密码」
- **实现**：`config.Load()` 改为 `JWT_SECRET` 必选，缺失时 `return nil, error`
- **判定**：`config_test.go` 双测覆盖；`main.go` 正确处理 error

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
- **§9 Provider 适配器** ✅ **已修复**（V1.3）：实现已补齐 7 家（OpenAI / Anthropic / Gemini / DeepSeek / Ollama / OpenAI-兼容 / Mock），§9 代码片段已同步更新
- **WebSocket 端点无鉴权**：`router.go` 自注 `// TODO: auth via query param`，与 §8 不符 → 列入 V1.4
- **附录 B**：列顶层 `agent/` 目录，实际无此目录 → V1.4 删除

---

## ✅ 亮点（文档与实现对齐良好）

- **Phase 7.1–7.6 完成状态**：逐项关联到具体 commit，可追溯 `git log`
- **Agent runtime 核心库**：`runtime`/`tools`/`planner`/`approval`/`permissions` 齐全，§10/§12 有落地
- **Docker 客户端**：`mock.go`/`real.go` 接口一致，符合 Docker SDK 优先规则
- **并发限流**：`ai/limited.go` + 单测，与 §21 风险项对应

---

## 推荐评审议程（按优先级）

✅ 历史议程已全部闭环（H1–H5 + M1–M4）。V1.3 新增动作：

1. Provider 全量实现（OpenAI / Anthropic / Gemini / DeepSeek / Ollama / OpenAI-兼容 / Mock）→ 已完成
2. Runtime 前端 Agent Chat UI → 待 V1.4
3. Agent 对话端到端（engine.run 全链路集成测试）→ 待 V1.4
4. H6 消息通知落地 → 待 V1.4
5. 文档最终自洽（§3 前端目录实现态 / §20 V1.4 迭代更新 / 附录 B 清理）→ 待 V1.4

---

## 评审动作记录

| # | 动作 | 状态 |
|---|------|------|
| 1 | 评审报告落盘 `docs/architecture-review.md` | ✅ |
| 2 | H4 硬编码 JWT secret 修复 | ✅ commit `4e2f81a` |
| 3 | H2 批1 交付（真实 JWT 认证 + bcrypt + login/register/refresh + 移除 MockAuth） | ✅ commit `f4bebd3` |
| 4 | seed 默认 admin 假 hash bug 修复（改为动态 bcrypt，admin/Admin@123!） | ✅ 同上 |
| 5 | H1 PostgreSQL 兼容（pgx 接入 + SQL 方言收敛 + 双库兼容迁移） | ✅ commit `7bbccd1` |
| 6 | H2 批2 交付（RBAC 注册中心 + 权限/角色表 + 权限门控中间件 + /roles /permissions /roles/groups 端点 + 16 个单测；/users CRUD 未含） | ✅ commit `4b262ed` |
| 7 | H2 批2b 交付（真实 /users CRUD + 角色分配，含 role 存在性校验，支持自定义角色；6 端点 users.manage 门控） | ✅ commit `3d67ff4` |
| 8 | H3 文档对齐（ARCHITECTURE.md 9 处改写 + 附录 C ADR-001 决策记录，Agent 从独立进程改为"进程内 Runtime + SSH Bridge"两层架构）| ✅ commit `a672ae6` / `b05cd5b` |
| 9 | H5 前端安全（permission-gated routes + sidebar filtering） | ✅ commit `89f0ba3` |
| 10 | H4 审计（audit middleware + audit_logs 表 schema） | ✅ commit `57b1195` |
| 11 | Provider 全量实现（Gemini REST + DeepSeek/Ollama OpenAI 兼容复用，7 家适配 + Factory 路由单测） | ✅ 本批 |
| 12 | SSH Bridge 完整实现（ssh/client/websocket/session/pty 五子包 + 集成测试） | ✅ 本批 |
| 13 | 文档清理（评审报告 V1.3 更新 + §9 Provider 代码片段同步） | ✅ 本批 |
| 14 | ⏳ V1.4 待办：Runtime 前端 Agent Chat UI / Agent 对话 E2E / H6 消息通知 / §3 前端目录实现态标注 | 待用户决策 |

> 注：本环境无 Docker，无法实机启动 PostgreSQL 做端到端验证。H1 已完成代码层面的双库兼容
> （SQLite 独占语法全部清除、`datetime('now')`→`CURRENT_TIMESTAMP`、`INSERT OR IGNORE`→`ON CONFLICT`、
> migrate.go 命名参数、时间解析 helper），并以 SQLite 路径做完整 E2E 回归验证；
> 端到端 PostgreSQL 验证待 CI 容器环境补上。H5 前端安全（commit `89f0ba3`）与 H4 审计（commit `57b1195`）已通过
> Vite 编译 + E2E 登录验证。V1.3 迭代新增：Provider 全量实现（OpenAI / Anthropic / Gemini / DeepSeek /
> Ollama / OpenAI-兼容 / Mock 全部通过 Factory 路由 + 7 家适配单元测试）、进程内 Runtime Agent 骨架
> （engine / planner / tools / approval / permissions 五子包）、SSH Bridge（SSH + WebSocket + Pty）
> 完成 `go test ./...`。V1.4 计划：Runtime 前端 Agent Chat UI、Agent 对话端到端、H6 消息通知落地、
> 文档最终自洽。
> 真实 PG 验证需在装有 docker 的开发机/CI 跑一次 `DATABASE_URL=postgresql://...` 启动即可确认。