# Phase 1 — 前端基础设施

目标：补齐前端工程底座，不改业务页面行为。为后续页面迁移（Market、Warehouse、Production、Buildings）提供统一 API hooks、UI 组件、表单和 Mock 框架。

目标代码路径：`clinet-next/`（新前端项目目录。目录名沿用现状拼写，非 typo）。

---

## 范围

### 做
1. 初始化 `clinet-next/` 为可构建的 Vite + React + TS 项目（从现有 client 复制基础工程）
2. 安装全部目标依赖
3. 建立 `api/errors.ts`、`api/queryKeys.ts`、整理 `api/client.ts` 边界
4. 建立 `components/ui/` 基础组件目录（shadcn/ui 组件，二次封装）
5. 建立 `mocks/` MSW mock 框架（browser + server + 示例 handler）
6. 建立测试基础设施（Vitest setup + Playwright 配置骨架）
7. 更新开源栈说明文档

### 不做
- 不迁移 Market / Warehouse / Production / Buildings 业务页面
- 不重做 GameCanvas 或 PixiJS 地图
- 不改业务页面视觉风格
- 不引入 Redux、MUI、Ant Design、Three.js、完整开源游戏模板
- 不一次性用 OpenAPI 生成全量类型
- 不修改后端 API 行为

---

## 现有项目状态对照

| 依赖/能力 | 现有项目 (`client/atlas-foods-client`) | clinet-next 目标 |
|-----------|---------------------------------------|------------------|
| `@tanstack/react-query` | ✅ 已安装 + Provider | 沿用 |
| `zustand` | ✅ 已安装 + 2 个 store | 沿用 |
| `pixi.js` | ✅ 已安装 | 沿用 |
| `tailwindcss` v4 | ✅ 已安装 + vite plugin | 沿用 |
| `vitest` | ✅ 已安装 | 沿用 |
| `react-router-dom` | ✅ 已安装 | 沿用 |
| `i18next` | ✅ 已安装 | 沿用 |
| `api/client.ts` | ✅ 已有 JWT + 错误处理 + get/post/patch/delete | 整理边界 |
| `api/errors.ts` | ❌ ApiError 嵌在 client.ts 内 | 抽取 |
| `api/queryKeys.ts` | ❌ 各 API 文件散写字符串 key | 新建 |
| `shadcn/ui` + 16 组件 | ❌ 无 `components/ui/` | 新建 |
| `react-hook-form` | ❌ 未安装 | 新增 |
| `zod` | ❌ 未安装 | 新增 |
| `recharts` | ❌ 未安装 | 新增 |
| `msw` | ❌ 未安装 | 新增 |
| `@playwright/test` | ❌ 未安装 | 新增 |
| `mocks/` | ❌ 不存在 | 新建 |
| `tests/` | ❌ 不存在 | 新建 |
| `settings.store.ts` | ❌ 不存在 | 新建 |

---

## 依赖清单（所有需安装的 npm 包）

### dependencies（新增）
| 包 | 用途 | 备注 |
|---|------|------|
| `react-hook-form` | 表单管理 |  |
| `zod` | schema 校验 |  |
| `recharts` | 图表 |  |
| `class-variance-authority` | shadcn 组件变体 | shadcn 间接依赖 |
| `clsx` | class 拼接 | shadcn 间接依赖 |
| `tailwind-merge` | Tailwind class 合并 | shadcn 间接依赖 |
| `lucide-react` | 图标 | shadcn 默认图标集 |
| `@radix-ui/react-dialog` | Dialog 组件基座 | shadcn 间接依赖 |
| `@radix-ui/react-tabs` | Tabs 组件基座 | shadcn 间接依赖 |
| `@radix-ui/react-select` | Select 组件基座 | shadcn 间接依赖 |
| `@radix-ui/react-tooltip` | Tooltip 组件基座 | shadcn 间接依赖 |
| `@radix-ui/react-slot` | 组件 asChild 模式 | shadcn 间接依赖 |
| `@radix-ui/react-label` | Label 组件基座 | shadcn 间接依赖 |
| `@radix-ui/react-separator` | Separator 组件基座 | shadcn 间接依赖 |
| `@radix-ui/react-progress` | Progress 组件基座 | shadcn 间接依赖 |
| `cmdk` | Command 组件基座 | shadcn 间接依赖 |
| `sonner` | Toast 组件 | shadcn Sonner toast |
| `vaul` | Sheet Drawer 组件基座 | shadcn 可选依赖 |
| `@hookform/resolvers` | React Hook Form + Zod 桥接 |  |

### devDependencies（新增）
| 包 | 用途 | 备注 |
|---|------|------|
| `msw` | Mock API | v2 |
| `@playwright/test` | E2E 测试 |  |
| `@testing-library/react` | Vitest 组件测试 |  |
| `@testing-library/jest-dom` | DOM matchers |  |
| `@testing-library/user-event` | 用户事件模拟 | 可选，建议加 |

### 从 client/atlas-foods-client 继承（已存在，不重复安装）
`@tanstack/react-query`, `zustand`, `pixi.js`, `tailwindcss`, `@tailwindcss/vite`, `react`, `react-dom`, `react-router-dom`, `i18next`, `react-i18next`, `i18next-browser-languagedetector`, `@vitejs/plugin-react`, `typescript`, `vite`, `vitest`, `jsdom`, `eslint`, `typescript-eslint`

---

## 子 Agent 分工

### 第 -1 步：初始化 clinet-next（串行）

**Agent: Init-Project**
- 目标：`clinet-next/` 从空目录变为可 `npm run build` 的 Vite + React + TS 项目
- 步骤：
  1. 在 `clinet-next/` 运行 `npm create vite@latest . -- --template react-ts` 或复制 `client/atlas-foods-client/` 的基础工程文件
     - 推荐复制：`package.json`、`tsconfig*.json`、`vite.config.ts`、`index.html`、`src/main.tsx`、`src/App.tsx`、`src/vite-env.d.ts`、`src/styles/globals.css`、`eslint.config.js`
     - 去掉业务内容（`src/features/`、`src/game/`、`src/api/*.api.ts`、`src/audio/`、`src/i18n/`、`src/store/`）
     - `src/App.tsx` 只保留 `<div>Hello Phase 1</div>`
  2. 合并依赖：将 client 已有依赖 + Phase 1 新增依赖一并写入 `package.json`
  3. `npm install`
  4. `npm run build` 通过
- 验收：
  - `clinet-next/package.json` 包含全部目标依赖
  - `npm run build` 通过
  - 不含 `features/`、`game/` 等业务目录

---

### 第 0 步：安装依赖（串行，因 Init-Project 已完成 package.json）

**Agent: Dependencies**
- 读取 `clinet-next/package.json`，核对依赖清单
- 如有缺漏：`npm install <pkg>` 补上
- 验收：`npm install` 成功，lockfile 更新清晰

---

### 第 1 步（5 个 Agent 并行）

**⚠️ 文件冲突警告**：以下 5 个 agent 写入的源文件互不相交。shadcn/ui CLI（`npx shadcn@latest init`）会修改 `package.json` 并生成 `src/lib/utils.ts` 和 `components.json`——**UI-Components agent 禁止运行 CLI**，只手动创建组件文件。

#### Agent: API-Boundary
**操作文件**：`clinet-next/src/api/`（新建 `errors.ts`、`queryKeys.ts`；修改 `client.ts`）

- **创建** `clinet-next/src/api/errors.ts`
  - 从 `client.ts` 抽取 `ApiError` 类
  - 导出 `ApiError`、`isApiError()` 类型守卫
- **创建** `clinet-next/src/api/queryKeys.ts`
  - 导出统一 query key 工厂函数/常量
  - 领域：`company`, `buildings`, `inventory`, `production`, `market`, `research`, `financial`, `executives`, `chat`, `leaderboard`, `powerups`
- **修改** `clinet-next/src/api/client.ts`
  - `import { ApiError } from './errors'`
  - 保持 get/post/patch/delete 签名不变
  - 保持 401 自动清 token 逻辑
- **不动** 任何 `*.api.ts` 文件（hooks 文件）
- 验收：
  - 现有 `*.api.ts` 文件不改也能继续用
  - `ApiError` 可从 `errors.ts` 导入
  - query key 有完整领域覆盖

#### Agent: UI-Components
**操作文件**：`clinet-next/src/components/ui/*.tsx`（新建 16+ 文件）；`src/lib/utils.ts`（新建）

- 手动创建 shadcn/ui 组件（**禁止运行 `npx shadcn init` CLI**，避免冲突 package.json）
- 先创建 `src/lib/utils.ts`：
  ```ts
  import { type ClassValue, clsx } from "clsx"
  import { twMerge } from "tailwind-merge"

  export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
  }
  ```
- 组件清单（从 shadcn/ui 源码手动复制核心逻辑，不要拿整个 dist）：
  Button, Card, Dialog, Tabs, Table, Input, Label, Select, Textarea, Tooltip, Toast（Sonner）, Sheet, Badge, Progress, Separator, Command
- 调色板引用现有 Tailwind v4 配置，不需要 `tailwind.config.js`
- 验收：
  - 组件可通过 `import { Button } from '@/components/ui/button'` 导入
  - TypeScript 编译通过

#### Agent: MSW-Framework
**操作文件**：`clinet-next/src/mocks/`（新建全部文件）

- **创建** `clinet-next/public/mockServiceWorker.js`
  - 运行 `npx msw init public --save` 生成 service worker
- **创建** `clinet-next/src/mocks/browser.ts`
  - 使用 `setupWorker` 注册 handler
  - 按 `VITE_ENABLE_MSW` 环境变量控制：仅 development + 显式启用时才加载
- **创建** `clinet-next/src/mocks/server.ts`
  - 使用 `setupServer` 供测试环境用
- **创建** `clinet-next/src/mocks/handlers/research.ts`（示例 handler）
  - mock GET `/api/v2/research/` 返回 `{ projects: [] }`
- **创建** `clinet-next/src/mocks/handlers/market.ts`（示例 handler）
  - mock GET `/api/v3/market-ticker/{id}/` 返回固定数据
- **创建** `clinet-next/src/mocks/handlers/` 下其余 3 个文件骨架（financial, executives, chat）— 只导出一个 `handlers` 数组
- 验收：
  - `browser.ts` 在 dev 模式可初始化不报错
  - 测试可 import `server.ts` 不报错
  - 至少 2 个 handler 有真实数据返回

#### Agent: Test-Setup
**操作文件**：`clinet-next/src/tests/`（新建全部文件）；`clinet-next/vitest.config.ts`（新建）

- **创建** `clinet-next/vitest.config.ts`
  ```ts
  /// <reference types="vitest/config" />
  import { defineConfig } from 'vitest/config'
  import react from '@vitejs/plugin-react'
  import path from 'path'

  export default defineConfig({
    plugins: [react()],
    resolve: {
      alias: { '@': path.resolve(__dirname, './src') },
    },
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: './src/tests/vitest.setup.ts',
    },
  })
  ```
- **创建** `clinet-next/src/tests/vitest.setup.ts`
  - 导入 `@testing-library/jest-dom`
  - 可选：全局 MSW server 初始化
- **创建** `clinet-next/src/tests/unit/` 目录 + 一条最小用例
- **创建** `clinet-next/tests/e2e/playwright.config.ts`（骨架）
  - 仅配置 baseURL 和 chromium，不写用例
- **修改** `clinet-next/package.json` 增加 scripts：
  ```json
  "test": "vitest run",
  "test:watch": "vitest",
  "test:e2e": "playwright test"
  ```
- 验收：
  - `npm test`（vitest run）通过
  - Playwright 配置可加载不报错

#### Agent: Store-Boundary
**操作文件**：`clinet-next/src/store/`（新建 `settings.store.ts`；整理已有 store）

- **检查** `clinet-next/src/store/` 已有文件
- **创建** `clinet-next/src/store/settings.store.ts`
  - 持久化用户偏好（音量、语言等）
  - 使用 `zustand/middleware` 的 `persist` 或 localStorage 简单包装
- **整理** store 职责注释：
  - `ui.store.ts` — 只放客户端交互状态（activeView, selectedBuildingId, chatOpen 等）
  - `game.store.ts` — 只放地图相机/选中态（zoom, pan, tick）
  - `settings.store.ts` — 用户偏好（音量、语言、主题）
  - **不复制** 服务器数据到 zustand
- 验收：
  - store 导入路径不变
  - 所有 store 文件顶部有职责注释

---

### 第 2 步（串行验证）

**Agent: Verify**
- 运行 `npm run build`（tsc + vite build）
- 运行 `npm run lint`（ESLint）
- 运行 `npx vitest run`（最小测试）
- 阅读 `package.json` 确认未引入禁止依赖（MUI、Ant Design、Redux、Three.js）
- 输出验证报告

---

## 硬规则（所有 Agent 必须遵守）

1. 页面组件不能直接 `fetch`。
2. 服务器数据不能长期放进 Zustand。
3. PixiJS 不能直接调用业务 API。
4. shadcn/ui 组件必须进入项目后再二次封装，不在业务页面里散乱粘贴原始样式。
5. 表单必须使用 React Hook Form + Zod。
6. 图表统一用 Recharts。
7. 新页面在后端未完成时必须先用 MSW mock。
8. 不允许一次性重写整个前端。
9. 每个 PR / 每个步骤必须能 build、能独立回滚。

---

## 目标目录结构

```
clinet-next/
  package.json
  tsconfig.json
  tsconfig.app.json
  tsconfig.node.json
  vite.config.ts              # 已有 + vitest 配置合并
  vitest.config.ts            # 新建（或整合进 vite.config.ts）
  index.html
  eslint.config.js
  public/
    mockServiceWorker.js      # MSW 生成

  src/
    main.tsx
    App.tsx
    vite-env.d.ts

    lib/
      utils.ts                # 新建（cn() 工具函数）

    app/
      providers.tsx           # 从 client 复制（实际是 src/app/providers.tsx）

    api/
      client.ts               # 从 client 复制 + 整理
      errors.ts               # 新建（从 client.ts 抽取 ApiError）
      queryKeys.ts            # 新建（统一 query key）
      generated/              # 预留（OpenAPI 类型生成）
      hooks/                  # 预留（未来迁移后的统一 hooks）

    components/
      ui/                     # shadcn/ui 基础组件（手动创建）
        button.tsx
        card.tsx
        dialog.tsx
        tabs.tsx
        table.tsx
        input.tsx
        label.tsx
        select.tsx
        textarea.tsx
        tooltip.tsx
        toast.tsx
        sheet.tsx
        badge.tsx
        progress.tsx
        separator.tsx
        command.tsx
      game/                   # 预留（游戏通用组件）

    store/
      ui.store.ts             # 从 client 复制
      game.store.ts           # 从 client 复制
      settings.store.ts       # 新建

    mocks/
      handlers/
        research.ts           # 新建（含示例 handler）
        financial.ts          # 新建（骨架）
        executives.ts         # 新建（骨架）
        chat.ts               # 新建（骨架）
        market.ts             # 新建（含示例 handler）
      browser.ts              # 新建
      server.ts               # 新建

    tests/
      vitest.setup.ts         # 新建
      unit/                   # 新建（最小用例）

  tests/
    e2e/
      playwright.config.ts    # 新建（骨架，项目根级）
```

---

## 验收标准（最终）

- `npm run build` 通过（tsc 零错误 + vite build 成功）
- `npm run lint` 通过
- `npm test` 通过（vitest run，至少 1 条用例）
- `package.json` 不含 MUI、Ant Design、Redux、Three.js
- `api/errors.ts` 和 `api/queryKeys.ts` 可被其他模块导入
- `components/ui/` 下 16 个组件存在且可导入
- MSW handler 可通过 `VITE_ENABLE_MSW` 环境变量在 dev 模式初始化
- 所有现有业务页面行为不变（TODO：后续在 clinet-next 迁移后验证）

---

## Phase 2 入口条件

Phase 1 通过上述验收后，进入 Phase 2：
- Market → 统一 hooks + 表单 mutation + Recharts
- Warehouse → 统一 hooks + 表格/库存组件
- Production → 统一 hooks + 队列 mutation
- Buildings → 统一 hooks + 建造/升级/移动/拆除 mutation
