# Codex 前端重构执行 Prompt

下面这段 Prompt 可直接交给 Codex，用于执行 New Haven 前端开源栈重构的第一阶段。

```txt
你是一名前端架构师，负责 New Haven 多人网页经济模拟游戏的前端重构。

项目是 Go API 后端 + React/Vite/PixiJS 前端的 monorepo。前端路径是：

client/atlas-foods-client/

这次重构允许融入成熟开源项目，但禁止无节制加依赖。

推荐开源项目：
- shadcn/ui：作为项目内部 UI 组件基础
- TanStack Query：管理服务器数据
- Zustand：只管理本地 UI 状态
- openapi-typescript / openapi-fetch：根据 OpenAPI 生成前端 API 类型
- React Hook Form + Zod：处理复杂表单和校验
- Recharts：处理市场、财务、利润图表
- MSW：处理前端 mock API
- Vitest：单元测试
- Playwright：端到端测试
- PixiJS：只负责地图和建筑 sprite 渲染

硬规则：
1. 页面组件不能直接 fetch。
2. 服务器数据不能长期放进 Zustand。
3. PixiJS 不能直接调用业务 API。
4. shadcn/ui 组件必须复制进项目后再二次封装，不要到处散乱使用。
5. 表单必须使用 React Hook Form + Zod。
6. 图表统一用 Recharts。
7. 新页面在后端未完成时必须先用 MSW mock。
8. 不允许一次性重写整个前端。
9. 每个 PR 必须能 build、能独立回滚。

目标目录：

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
    ui/              # shadcn/ui 组件
    game/            # 游戏通用组件

  store/
    ui.store.ts
    game.store.ts
    settings.store.ts

  features/
    market/
    production/
    buildings/
    warehouse/
    research/
    financial/
    executives/
    chat/
    powerups/
    map/

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

执行前必须先读：
- package.json
- src/api/client.ts
- src/app/providers.tsx
- src/store/ui.store.ts
- src/game/GameCanvas.tsx
- 现有 src/api/*.api.ts
- 现有 src/features/* 目录

第一阶段只做：
1. 审计现有依赖和目录，不重复创建已有能力。
2. 安装和初始化必要开源项目。
3. 建立或整理 api/client.ts、queryKeys.ts、QueryProvider。
4. 建立 shadcn/ui 基础组件目录和最小组件集。
5. 建立 MSW mock 框架。
6. 建立或更新 frontend-open-source-stack.md，说明每个库的用途、边界和禁止事项。
7. 不改业务页面行为。

完成后必须验证：
- npm run build
- npm run lint
- 如新增测试框架，则运行最小测试命令

不要做：
- 不迁移业务页面。
- 不重做 GameCanvas。
- 不引入 Redux、Material UI、Ant Design、Three.js、完整开源游戏模板。
- 不一次性用 OpenAPI 重写全部 API。

完成第一阶段后，再进入第二阶段：
把 Market、Warehouse、Production、Buildings 迁移到统一 hooks。
```

