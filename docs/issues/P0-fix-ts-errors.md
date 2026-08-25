# P0: 修复前端 TypeScript 编译错误

## 问题描述

`tsc --noEmit` 报错，共 18 处编译错误，集中在 8 个测试文件中。

### 错误分类

1. **未使用导入**（TS6133）— 9 处
   - `useT`（DashboardPage.test.tsx, ServerListPage.test.tsx）
   - `Message`（AddServerModal.test.tsx）
   - `waitFor`（ServerListPage.test.tsx）
   - `act`（useConfirm.test.ts）
   - `useCreateServer`、`mockedCreate`（hooks.test.ts）
   - `useContainerLogs`（hooks.test.ts）

2. **全局类型未声明**（TS2593）— `beforeEach` 未找到
   - tsconfig 缺少 `vitest/globals` types

3. **Jest 类型误用**（TS2694）— `jest.Mock` 不存在
   - `servers.test.ts` 中 `jest.Mock` 应改为 `ReturnType<typeof vi.fn>`

4. **泛型参数错误**（TS2558）— `vi.fn<[string], void>()` 不支持双泛型
   - `api.test.ts` 中 3 处

5. **缺失必填 prop**（TS2741）— `onSuccess` 缺失
   - `AddServerModal.test.tsx` 中 2 处

6. **类型不匹配**（TS2345）— `CreateServerRequest` 缺少 `username`
   - `servers.test.ts` 中 1 处

## 根因分析

- `tsconfig.json` 和 `tsconfig.app.json` 的 `types` 字段缺少 `vitest/globals`
- 测试文件中有大量从 Jest 迁移到 Vitest 后残留的旧导入
- Vitest v4 不支持 `vi.fn<[Args], Return>()` 双泛型写法

## 修复方案

1. 在 `tsconfig.json` 和 `tsconfig.app.json` 中添加 `"vitest/globals"` 到 `types`
2. 移除所有未使用导入
3. 将 `jest.Mock` 类型断言改为 `ReturnType<typeof vi.fn>`
4. 将 `vi.fn<[string], void>()` 改为 `vi.fn()`
5. 补全 `AddServerModal` 组件调用中的 `onSuccess` prop
6. 补全 `CreateServerRequest` 中的 `username` 字段

## 影响范围

- 8 个测试文件 + 2 个 tsconfig 文件
- 不影响任何业务代码

## 验证标准

- `tsc --noEmit` 零错误
- `npm test` 全部 97 个测试通过
