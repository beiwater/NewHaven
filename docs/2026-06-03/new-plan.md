# New Plan - 2026-06-03

目标：把 2026-06-02 的杂乱材料整理成一条可执行路线。后续模型或开发者应按本 plan 拆任务，不要直接把所有功能同时铺开。

## 总原则

1. 先补当前有后端、缺前端、能立刻形成产品闭环的功能。
2. 经济系统实现必须服从“极简建筑产量版”：建筑产量、成本指数、饱和度控价、机器人流动性。
3. 不引入工人管理、配方工作量黑箱、全局销量卡死等已删除变量。
4. 高管、研究、金融页面优先与精算表里的曲线和变量对齐。
5. 每个任务都要让模型先读现有代码模式，再动手，不允许凭空新建一套风格。

## Phase 0 - 代码现状审计

目的：确认真实代码状态，避免按过期文档开发。

交付物：

- 当前前端路由、导航、API hook、feature 目录结构摘要。
- 当前后端 handler、service、storage、formula、config、test 摘要。
- dev-plan 中每个“要做/先不做”的实际状态标记。
- 一张任务风险表：已存在、部分存在、缺失、接口不匹配。

验收：

- 能回答 Research、Financial、Executive、Chat、Power-up、Leaderboard 分别缺哪几层。
- 能确认建筑市场 API 到底该调用 `/api/v2/buildings/` 还是 `/api/v2/buildings/market/`。
- 能列出经济公式当前实现与 v1.3.1/v1.3.2 的差异。

## Phase 1 - 前端闭环优先包

### 1. Research Page

范围：

- 新建或补齐 Research 页面。
- 接入研究列表、开始研究、进度、完成研究 API。
- 显示研究节点的商品组、消耗、现金费用、进度和收益。
- 页面文案与经济模型一致：食品研究是资源/市场驱动，不是抽象科技树。

不做：

- WebSocket 实时进度。
- 新图标资产。

### 2. Financial Pages

范围：

- 新建 Financial 页面或面板。
- 接入损益表、资产负债表、现金流量表、最近流水、历史财务。
- 历史财务可先用现有图表库或轻量表格，核心是数据可读。

建议入口：

- TopBar 资金数字点击打开财务面板。
- 如果现有导航更适合，也可新增 Financial 导航项。

### 3. Executive Page

范围：

- 高管市场搜索。
- 招募。
- 培训。
- 高管详情。
- 展示生产加成、销售加成、管理折扣、工资和升级成本。

不做：

- 挖角。
- 报价管理。

### 4. Chat API 接入

范围：

- 将 `ChatPanel.tsx` 从 mock 数据迁移到消息 API。
- 支持加载消息、发送消息、标记已读、聊天室和联系人。
- 保留现有 UI 风格，避免重做聊天界面。

### 5. Power-up UI

范围：

- 接入可用 boost 列表、使用 boost、活跃状态。
- UI 展示名统一为 `Power-up`。
- 后端 API 名称可以保留 simboost，但用户可见文案不要出现 SimBoost。

## Phase 2 - 后端基础能力包

### 1. 多 Player 与真实认证

范围：

- 密码哈希。
- 签名 JWT。
- token 解析 playerId。
- 多玩家数据隔离。
- dev mode 与真实认证边界清晰。

验收：

- 不再使用 `base64(username)` 作为真实 token。
- 非 dev 模式不应固定 `dev-player`。
- 注册/登录/鉴权均有测试。

### 2. Postgres 持久化

范围：

- 审计 `storage/postgres.go` 是否覆盖内存 storage 的全部接口。
- 首次运行建表或迁移。
- Scheduler 保存逻辑支持 Postgres。
- 基础备份或快照策略。

验收：

- 内存模式和 Postgres 模式的核心行为一致。
- 至少覆盖玩家、公司、建筑、库存、市场订单、财务记录等核心数据。

### 3. Leaderboard

范围：

- 新增排行榜 API。
- 支持按净资产、等级或关键经营指标排序。
- 配套前端页面。

验收：

- 只展示允许公开的玩家/公司字段。
- 支持分页。
- 排名计算有测试。

### 4. 搜索、筛选、分页

范围：

- 市场订单。
- 高管搜索。
- 债券列表。

验收：

- 查询参数清晰。
- 后端有边界测试。
- 前端保留筛选状态，空结果和加载状态完整。

## Phase 3 - 经济模型对齐包

范围：

- 审计 `formula/production.go`、`formula/bonds.go`、`service/market_competition.go`、`service/government.go`、`service/market_trade.go`。
- 把已实现公式和 v1.3.1/v1.3.2 的公式表逐项对照。
- 给出差异清单，必要时再拆实现任务。

重点：

- 生产是否由建筑基础产量直接驱动。
- 饱和度是否按商品组控价。
- 机器人是否有限额，是否有价差，是否防套利。
- 高管和口碑是否相加。
- 等级甜蜜点后是否有管理成本收敛。
- 债券 50x 系数是否真的合理。

## Phase 4 - 测试与回归

优先测试：

- formula 纯函数。
- market trade 边界。
- bond 利息与违约。
- government contract 闭环。
- production 原料消耗、品质、产量。
- auth 注册、登录、JWT、玩家隔离。
- storage 内存/Postgres 一致性。

最低验收：

- 每个新增 API 至少有 happy path 和一个失败路径测试。
- 每个新增前端页面至少有加载、空状态、错误状态、主要操作状态。
- 经济公式修改必须有数值回归样例。

## 暂缓清单

以下内容先不要让模型实现，除非用户重新明确要求：

- Bond Market 页面。
- Auction 页面。
- Aerospace 页面。
- WebSocket 实时推送。
- 资源图标补齐。
- 动态天气和生产修正。
- 高管挖角和报价管理。

## 推荐执行顺序

1. 运行 Phase 0 审计提示词。
2. 根据审计结果，先做 Research Page。
3. 做 Executive Page，因为它直接连接高管升级曲线。
4. 做 Financial Pages，方便观察经济结果。
5. 做 Chat API 接入和 Power-up UI。
6. 做 Auth/Postgres/Leaderboard 后端包。
7. 最后做经济公式对齐和测试扩容。

