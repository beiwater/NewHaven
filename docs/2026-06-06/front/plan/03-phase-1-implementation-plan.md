# Phase 1 前端基础设施实施计划

目标：先把前端工程底座补齐，但不改变业务页面行为。第一阶段完成后，后续页面迁移可以沿统一 API hooks、统一 UI 组件、统一表单和统一 Mock 框架推进。

## 范围

第一阶段只做：

1. 安装和初始化必要开源项目。
2. 建立 `api/client.ts`、`queryKeys.ts`、`QueryProvider` 的统一边界。
3. 建立 shadcn/ui 基础组件目录。
4. 建立 MSW mock 框架。
5. 建立前端开源栈说明文档。
6. 不改业务页面行为。

## 目标目录

```txt
client/atlas-foods-client/src/
  app/
    providers/
      QueryProvider.tsx

  api/
    client.ts
    queryKeys.ts
    errors.ts
    generated/
    hooks/

  components/
    ui/
    game/

  store/
    ui.store.ts
    game.store.ts
    settings.store.ts

  mocks/
    handlers/
      research.ts
      financial.ts
      executives.ts
      chat.ts
      market.ts
    browser.ts
    server.ts

  tests/
    e2e/
    unit/
```

说明：现有项目已经有 `src/api/client.ts`、`src/app/providers.tsx`、`src/store/ui.store.ts` 等文件，实施时要先读现状，再按最小改动补齐，不要机械创建重复文件。

## 任务 1：依赖审计和安装

先检查 `client/atlas-foods-client/package.json` 是否已有：

```txt
@tanstack/react-query
zustand
pixi.js
tailwindcss
```

再决定是否新增：

```txt
react-hook-form
zod
recharts
msw
vitest
@playwright/test
openapi-typescript
openapi-fetch
```

验收：

- `npm install` 后 lockfile 更新清晰。
- `npm run build` 能通过。
- 没有引入 Material UI、Ant Design、Redux、Three.js 等非目标依赖。

## 任务 2：QueryProvider 和 queryKeys

建立或整理 QueryProvider：

```txt
QueryClient
QueryClientProvider
默认 staleTime / retry 策略
开发环境 React Query Devtools 可后置
```

建立统一 query keys：

```txt
company
buildings
inventory
production
market
research
financial
executives
chat
leaderboard
powerups
```

验收：

- 现有 TanStack Query hooks 可以继续运行。
- 新 hooks 不再各自手写散乱 key 字符串。

## 任务 3：API client 边界

整理 `src/api/client.ts`：

```txt
base URL
JWT Authorization
JSON request/response
统一 ApiError
401 清 token 或抛出认证错误
get/post/put/delete helpers
```

验收：

- 现有 API 文件不用大改也能继续使用。
- 错误结构可被页面统一展示。

## 任务 4：shadcn/ui 基础组件

优先建立最小组件：

```txt
Button
Card
Dialog
Tabs
Table
Input
Label
Select
Textarea
Tooltip
Toast/Sonner
Sheet
Badge
Progress
```

验收：

- 组件进入 `src/components/ui/`。
- 后续业务组件从项目内部路径导入，而不是依赖外部黑盒风格。
- 不在第一阶段大规模改现有页面 UI。

## 任务 5：MSW mock 框架

建议结构：

```txt
src/mocks/
  handlers/
    research.ts
    financial.ts
    executives.ts
    chat.ts
    market.ts
  browser.ts
  server.ts
```

第一阶段只需要框架可用和少量示例 handler，不要求把所有 API 都 mock 完。

验收：

- 开发环境可以按开关启用 MSW。
- 测试环境可以复用 `server.ts`。
- Mock DTO 不乱造字段，尽量贴近现有后端返回。

## 任务 6：测试框架落点

先准备关键路径，不追求覆盖率：

```txt
1. 登录成功后进入游戏
2. 打开 Market 页面
3. 创建订单
4. 领取生产
5. 打开 Research 页面
6. 打开 Financial 页面
7. 打开 Executive 页面
8. 发送聊天消息
```

验收：

- Vitest 可运行至少一个最小测试。
- Playwright 配置可以后置，但目录和计划清晰。

## 不做

```txt
不迁移 Market 页面
不迁移 Warehouse 页面
不迁移 Production 页面
不迁移 Buildings 页面
不重做 GameCanvas
不改业务页面视觉风格
不修改后端 API 行为
不一次性接入 OpenAPI 生成全量类型
```

## Phase 2 入口

第一阶段完成且 build 通过后，再进入第二阶段：

```txt
Market       → 统一 hooks + 表单 mutation + Recharts
Warehouse    → 统一 hooks + 表格/库存组件
Production   → 统一 hooks + 队列 mutation
Buildings    → 统一 hooks + 建造/升级/移动/拆除 mutation
```

