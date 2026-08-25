---
status: active
version: 1.0
last_reviewed: 2026-08-25
owner: QA
related: [AGENTS.md, docs/coding-standards.md]
---

# Testing Standards

> **测试规范** — 项目的测试策略和质量标准。
>
> 适用对象：QA, Dev, Lead。

---

## 1. 测试层级

### 1.1 单元测试（Unit Test）

- **位置**：与源码同目录，`*_test.go` / `*.test.ts`
- **覆盖**：每个重要模块必须有单元测试
- **工具**：按项目技术栈选择（Go testing / Vitest / Jest）
- **要求**：
  - 测试命名：`Test<FunctionName>_<Scenario>_<ExpectedResult>`
  - 每个测试独立，不依赖其他测试
  - 使用 table-driven tests 或 describe/it 结构

### 1.2 集成测试（Integration Test）

- **位置**：`tests/integration/`
- **覆盖**：模块间交互、API 端到端
- **要求**：
  - 使用测试数据库（in-memory 或 fixture）
  - 测试前准备数据，测试后清理
  - 模拟外部依赖

### 1.3 E2E 测试（可选）

- **位置**：`tests/e2e/`
- **覆盖**：完整用户流程
- **适用**：核心业务流程

---

## 2. 测试命名规范

### 通用格式

```
Test<FunctionName>_<Scenario>_<ExpectedResult>
```

示例：

```go
func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
    // ...
}

func TestUserService_CreateUser_Success(t *testing.T) {
    // ...
}
```

### TypeScript

```typescript
describe('AuthService', () => {
  it('should return token on valid credentials', () => {
    // ...
  });
  
  it('should throw error on invalid credentials', () => {
    // ...
  });
});
```

---

## 3. 测试数据

- 使用固定种子数据，保证测试可重复
- 禁止在测试中使用真实生产数据
- 测试数据应与业务逻辑解耦

---

## 4. 测试要求

### 4.1 必须测试的场景

- 正常流程（Happy Path）
- 边界条件（Empty input, Max length, etc.）
- 错误处理（Invalid input, Network error, etc.）
- 安全场景（Unauthorized access, SQL injection, etc.）

### 4.2 覆盖率要求

- 核心模块（auth, permission, runtime）：≥ 80%
- 一般模块（service, api）：≥ 60%
- 新增代码必须有对应测试

### 4.3 禁止事项

- **禁止简单删除失败的测试**
- 禁止使用 skip 掩盖问题
- 禁止在测试中硬编码敏感信息

---

## 5. 运行测试

### 通用命令

```bash
# 运行所有测试
<test-framework> test

# 运行指定包/文件
<test-framework> test ./path/to/package

# 运行指定测试
<test-framework> test -run <TestName>

# 查看覆盖率
<test-framework> test -coverprofile=coverage.out
```

### CI 测试要求

PR 必须通过以下测试检查：

1. 单元测试全绿
2. 集成测试全绿（如配置）
3. lint 无问题

---

## 6. 测试失败处理

当测试失败时：

1. **分析原因** — 不要简单删除失败的测试
2. **分类问题** — 是测试 bug 还是代码 bug
3. **报告问题** — 在 Issue 中说明失败原因
4. **建议修复** — 提供修复建议

**禁止**：为了通过测试而删除测试用例。
