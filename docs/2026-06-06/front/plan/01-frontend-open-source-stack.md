# 前端开源栈融合方案

目标：在不一次性重写前端的前提下，把成熟开源项目融入 New Haven，减少自研基础设施，把精力集中到经营模拟、经济系统、地图交互和游戏 UI 上。

## 推荐技术路线

```txt
React + Vite + Tailwind
  ├─ UI 组件：shadcn/ui
  ├─ 表单：React Hook Form + Zod
  ├─ 服务器数据：TanStack Query
  ├─ 本地 UI 状态：Zustand
  ├─ API 类型：openapi-typescript / openapi-fetch
  ├─ 图表：Recharts
  ├─ Mock API：MSW
  ├─ 地图渲染：PixiJS / PixiJS React
  └─ 测试：Vitest + Playwright
```

## 必须融入

| 类别 | 项目 | 用途 | 边界 |
| --- | --- | --- | --- |
| UI | shadcn/ui | 快速建立 Button、Card、Dialog、Tabs、Table、Toast、Tooltip、Sheet 等基础组件 | 复制进项目后再二次封装，不在业务页面里散乱粘贴 |
| Server State | TanStack Query | 管理 API 数据、缓存、刷新、mutation、错误和 loading 状态 | 只放服务器数据，不管地图 hover、弹窗开关等 UI 状态 |
| Local State | Zustand | 管理本地 UI 状态 | 不长期保存库存、订单、公司资金、研究进度等服务器数据 |
| API Type | openapi-typescript / openapi-fetch | 从 OpenAPI schema 生成类型安全 API 客户端 | 后端契约稳定后逐步接入，不能为了生成类型而阻塞现有业务 |
| Form | React Hook Form | 管理下单、研究、高管、债券、政府竞标等复杂表单 | 不用临时 `useState` 拼复杂表单状态 |
| Validation | Zod | 表单字段和 API 输入校验 | 表单 schema 与提交 DTO 尽量复用 |
| Chart | Recharts | 财务、市场、利润、现金流、饱和度图表 | 先做 Line、Bar、Area、Pie，不急着做复杂 K 线 |
| Mock | MSW | 后端未完成时让前端独立开发和测试 | Mock 数据必须贴近真实 API DTO |
| Unit Test | Vitest | hooks、schema、组件逻辑测试 | 重点覆盖新增 API hooks 和表单 schema |
| E2E | Playwright | 登录、市场、生产、研究、财务、高管、聊天等关键路径 | 先做少量稳定路径，不铺满所有 UI |

## 可以后面再融入

```txt
TanStack Router     → 如果以后真要多页面路由
Storybook           → 如果需要独立预览 UI 组件
Floating UI         → 如果 tooltip/popover 复杂到 shadcn 默认能力不够
Framer Motion       → 如果后续要做高级 UI 动画
```

## 先不要融入

```txt
Redux               → 当前 Zustand + TanStack Query 足够
Material UI         → 风格偏企业后台，游戏感不足
Ant Design          → 风格也偏后台系统
Three.js            → 当前是 2D 经济游戏，没有 3D 必要
完整开源游戏模板     → 架构污染太大，容易拖慢现有项目
Unity WebGL / Unreal → 体积和工作流都不适合当前浏览器经营模拟项目
```

## 页面映射

```txt
MarketPage     → Card + Table + Tabs + Dialog + Recharts
FinancialPage  → Card + Tabs + Table + Recharts
ExecutivePage  → Card + Dialog + Badge + Form
ResearchPage   → Card + Progress + Button + Form
ChatPanel      → Sheet + ScrollArea + Input
BuildView      → Card + Dialog + Zustand UI 状态
GameCanvas     → PixiJS 地图、建筑 sprite、点击、缩放、进度圈
```

## 一句话方向

用开源项目把底层问题外包掉：shadcn/ui 解决 UI，TanStack Query 解决 API 状态，Zustand 解决本地状态，MSW 解决假后端，Recharts 解决图表，React Hook Form + Zod 解决表单。项目只专心做游戏业务。

