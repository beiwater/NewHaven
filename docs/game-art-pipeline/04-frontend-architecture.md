# 前端技术架构文档

## 目标

本文档只定义后续实现架构，不在本轮实现代码。
核心目标是让地图、建筑、市场价格和素材加载都配置化，避免代码变成屎山。

## 推荐架构

- React：负责外层 UI，如顶部状态、左侧导航、右侧详情、底部价格条。
- Phaser：负责地图、建筑精灵、动画、点击、拖拽和收取反馈。
- Zustand：负责前端状态缓存，不负责业务真相。
- Go API Client：统一请求后端，不允许组件直接 fetch。
- Adapter：把 Go API 返回结构转换成前端稳定模型。
- Asset Registry：根据 building id、level、state 查找素材。
- Game Config：定义建筑、产业链、资源、等级、占地、动画状态。

## 边界

React 不直接绘制地图。
Phaser 不直接请求后端。
组件不直接拼图片路径。
组件不硬编码建筑等级差异。
Go API 数据必须先进入 adapter，再进入 store。

## 数据流

`Go API -> api client -> adapter -> store -> React UI / Phaser Scene`

用户操作：

`React/Phaser event -> action -> api client -> adapter -> store update -> render`

## 素材加载策略

后续实现时应该存在一个资产注册表，概念如下：

- 输入：`building_id`, `level`, `state`
- 输出：对应 PNG 或 spritesheet 元数据

不要让 UI 写：

`if buildingId === "bakery" use bakeryLv2Image`

应该让 UI 写：

`assetRegistry.getBuildingSprite(buildingId, level, state)`

## Go API 接入原则

正式前端应该优先面向 Go API 合约。
本地 Node demo API 只能作为旧原型参考，不应该和正式 API 混成两套真相。

优先接口：

- 登录/注册：`/api/login`, `/api/register`
- 建筑市场：`/api/v2/buildings/market/`
- 我的建筑：`/api/v2/companies/me/buildings/`
- 放置建筑：`/api/v2/buildings/place/`
- 移动建筑：`/api/v2/buildings/move/`
- 生产任务：`/api/v2/production/jobs/`
- 可领取任务：`/api/v2/production/claimable/`
- 市场最新价：`/api/v3/market-ticker/{id}/`
- 资源信息：`/api/v3/resources/`

## 底部价格条

底部显示的是最新价格，不是库存。
库存属于仓库面板。

价格条应展示：

- 资源名
- 最新价格
- 涨跌幅
- 小趋势线
- 质量或等级标识，如果后端支持

## 后续验收标准

- 新增建筑只改配置和素材，不改 UI 判断链。
- 新增等级只补素材和配置，不复制组件。
- 新增产业链只增加数据，不重写地图逻辑。
- 替换素材不会影响 API 或业务逻辑。
