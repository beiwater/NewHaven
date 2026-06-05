# 前端架构边界

目标：明确每一层负责什么，避免后续开发时把服务器状态、地图状态、业务 API、UI 组件混在一起。

## 总体结构

```txt
Browser
  ├─ React UI
  │   ├─ 页面、面板、表单、弹窗、图表
  │   └─ 通过 TanStack Query hooks 读取服务器数据
  │
  ├─ PixiJS Canvas
  │   ├─ 地图背景
  │   ├─ 建筑 sprite
  │   ├─ 地图缩放、拖拽、选中、高亮
  │   └─ 通过 Zustand 读写轻量 UI 状态
  │
  ├─ Zustand
  │   └─ 本地 UI 状态
  │
  ├─ TanStack Query
  │   └─ 服务器状态、缓存、mutation
  │
  └─ API client
      └─ REST/JSON + JWT
```

## React 负责

```txt
TopBar
LeftSidebar
BottomMarketTicker
RightBuildingPanel
MarketPage
FinancialPage
ExecutivePage
ResearchPage
Warehouse/Inventory
ChatPanel
Dialog/Sheet/Toast/Form/Table/Chart
```

React 页面只负责显示和交互编排，不直接散写 `fetch`。页面要通过 `api/*.api.ts` 或未来 `api/hooks/` 里的 hooks 读取数据。

## PixiJS 负责

```txt
地图背景
建筑图片
建筑点击区域
建筑选中态
可建造地块高亮
生产进度圈
简单粒子和收取动画
地图缩放、拖拽、镜头位置
```

PixiJS 不直接请求业务 API。它可以读取当前建筑列表渲染画面，也可以通过回调或 Zustand 更新 `selectedBuildingId`、`placementBuildingId`、`movingBuildingId` 这类 UI 状态。

## TanStack Query 负责

```txt
公司信息
库存
建筑列表
生产队列
市场价格
市场订单
研究项目
高管列表
财务报表
聊天消息
排行榜
Power-up 状态
```

这些数据来自后端，应该有缓存、失效、重试、错误和加载状态。不要长期复制到 Zustand。

## Zustand 负责

```txt
activeView
selectedBuildingId
placementBuildingId
movingBuildingId
chatOpen
currentModal
mapCamera
hoveredTile
settingsPanelOpen
```

Zustand 只保存客户端当前交互状态。页面刷新后可以丢失的状态，通常才适合放这里。

## API client 负责

```txt
统一 base URL
统一 JWT Authorization header
统一 JSON decode
统一错误类型
统一 401 处理
统一请求/响应类型
```

业务页面不应该关心 token 怎么取、错误怎么解析、URL 怎么拼。

## 表单边界

市场下单、开始研究、招募高管、训练高管、发行债券、政府竞标等表单统一走：

```txt
shadcn/ui 表单组件
  ↓
React Hook Form
  ↓
Zod schema
  ↓
TanStack Query mutation
  ↓
API client
```

## 图表边界

市场、财务、利润、现金流、成本结构、饱和度等图表统一用 Recharts。第一阶段只需要：

```txt
LineChart    价格趋势
BarChart     建筑利润
AreaChart    现金流
PieChart     成本结构
```

K 线和更复杂金融图后置。

## 硬规则

1. 页面组件不能直接 `fetch`。
2. 服务器数据不能长期放进 Zustand。
3. PixiJS 不能直接调用业务 API。
4. shadcn/ui 组件必须进入项目后再二次封装。
5. 表单必须使用 React Hook Form + Zod。
6. 图表统一用 Recharts。
7. 新页面在后端未完成时必须先用 MSW mock。
8. 不允许一次性重写整个前端。
9. 每个 PR 必须能 build、能独立回滚。

