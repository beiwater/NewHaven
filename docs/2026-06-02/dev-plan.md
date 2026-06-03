# Development Plan

**Date:** 2026-06-02  
**Scope:** Gap closure plan for frontend and backend

---

## Frontend Completion Plan

### 1. Research Page — 新页面 (侧边栏已有入口)

**当前状态:** 侧边栏有 Research 导航项（使用 timer 图标），但 `activeView === 'research'` 时无对应内容，约等于空。

**需要实现:**

| 功能 | API | 描述 |
|------|-----|------|
| 研发项目列表 | `GET /api/v2/research/` | 显示可用的研究项目（名称、描述、时长、消耗） |
| 开始研究 | `POST /api/v2/research/start/` | 选择一个项目开始研究，扣除资源/资金 |
| 研究进度 | `GET /api/v2/research/progress/` | 进度条 + 预计剩余时间 |
| 完成研究 | `POST /api/v2/research/complete/` | 领取研究成果 |

**资产需求:** 无特殊图片需求，可用现有 icon。

**文件:** `src/features/research/ResearchPage.tsx`（新建），sidebar 导航项已有。

---

### 2. Financial Pages — 新页面

**当前状态:** 后端 6 个财务端点完整实现，前端页面为 0。

**需要实现:**

| 功能 | API | 描述 |
|------|-----|------|
| 损益表 | `GET /api/v2/companies/me/income-statement/` | 收入/支出/净利润表格 |
| 资产负债表 | `GET /api/v2/companies/me/balance-sheet/` | 资产/负债/权益 |
| 现金流量表 | `GET /api/v2/companies/me/cashflow-statement/` | 经营/投资/筹资现金流 |
| 最近流水 | `GET /api/v2/companies/me/cashflow/recent/` | 最近 100 条流水记录列表 |
| 历史财务 | `GET /api/v3/companies/me/past-finances/` | 历史每周净利润趋势图 |

**入口建议:** TopBar 上点击资金数字打开财务面板，或左侧新增 Financial 导航项。

**文件:** `src/features/financial/`（新建目录）

---

### 3. Bond Market — 新页面（先不用）

**当前状态:** 后端 6 个债券端点完整，前端页面为 0。

**需要实现:**

| 功能 | API | 描述 |
|------|-----|------|
| 债券市场列表 | `GET /api/bonds/` | 显示所有可交易债券（发行人、利率、评级、面值） |
| 债券详情 | `GET /api/bonds/{id}/` | 债券详细信息 |
| 购买债券 | `POST /api/bonds/{id}/buy/` | 输入数量购买 |
| 发行债券 | `POST /api/bonds/` | 公司自己发行债券（金额、利率） |
| 持有的债券 | `GET /api/v2/companies/me/bonds/owned/` | 我购买的债券列表 |
| 利息结算 | `POST /api/bonds/settle-interest/` | 手动领取利息 |

**入口建议:** 左侧新增 Bonds 导航项。

**文件:** `src/features/bonds/`（新建目录）

---

### 4. Executive (高管) — 新页面

**当前状态:** 后端 6 个高管端点完整，前端页面为 0。

**需要实现:**

| 功能 | API | 描述 |
|------|-----|------|
| 高管市场 | `POST /api/v2/executives/search/` | 搜索可用高管（列表卡片，显示技能等级）|
| 招募 | `POST /api/v2/executives/recruit/` | 招募选定高管 |
| 培训 | `POST /api/v2/executives/train/{id}/` | 提升高管技能 |
| 挖角 | `POST /api/v3/executives/poach/` | 从其他公司挖人 （这个先不要做）|
| 报价管理 | `GET/POST /api/v3/executives/offers/` | 查看/回应挖角报价 （这个先不要做）|
| 高管详情 | `GET /api/v3/executives/{id}/` | 单个高管详情 |

**入口建议:** 左侧新增 Executives 导航项。

**文件:** `src/features/executives/`（新建目录）

---

### 5. Auction (拍卖) — 新页面（先不做）

**当前状态:** 后端 3 个拍卖端点完整，前端页面为 0。

**需要实现:**

| 功能 | API | 描述 |
|------|-----|------|
| 拍卖列表 | `GET /api/v2/auctions/` | 正在进行的建筑拍卖 |
| 拍卖详情 | `GET /api/v2/auctions/{id}/` | 详情 + 当前最高出价 |
| 出价 | `POST /api/v2/auctions/{id}/bid/` | 输入金额出价 |
| 我的拍卖 | `GET /api/v2/companies/me/auctions/` | 我参与的拍卖 |

**入口建议:** 放到 Build 页面下作为子标签（Tab），或在 Build 页面加一个 "Auctions" 按钮。

**文件:** 可合并进 `src/features/buildings/` 或新建 `src/features/auction/`

---

### 6. Aerospace (航空/火箭) — 新页面（先不做）

**当前状态:** 后端 5 个端点完整，前端页面为 0。

**需要实现:**

| 功能 | API | 描述 |
|------|-----|------|
| 火箭项目列表 | `GET /api/v2/aerospace/projects/` | 可创建的火箭项目 |
| 创建项目 | `POST /api/v2/aerospace/projects/create/` | 选择项目，投入资源启动 |
| 发射历史 | `GET /api/v2/aerospace/launches/` | 历史发射记录 |
| 发射火箭 | `POST /api/v2/aerospace/launch/` | 执行发射 |
| 组件列表 | `GET /api/v2/aerospace/components/` | 可用的火箭组件 |

**入口建议:** 左侧新增 Aerospace 导航项，或放在 Research 附近。

**文件:** `src/features/aerospace/`（新建目录）

---

### 7. SimBoost 界面 — 新 UI （可以做但是不能吉叫这个名字改名叫power-up）

**当前状态:** 后端 3 个 SimBoost 端点，前端页面为 0。

**需要实现:**

| 功能 | API | 描述 |
|------|-----|------|
| 可用 Boost 列表 | `GET /api/v2/players/simboosts/` | 显示可用加速道具（类型、剩余次数） |
| 使用 Boost | `POST /api/v2/players/simboosts-use/` | 激活加速效果 |
| 活跃 Boost 状态 | 从 `GET /api/v2/players/simboosts-use/` 返回 | 显示剩余时间 |

**入口建议:** TopBar 加一个小火箭/闪电图标按钮，点击弹出使用面板。

**文件:** `src/features/simboost/`（新建目录）

---

### 8. Chat — 对接后端 API（需要做）

**当前状态:** 聊天 UI 完整但使用本地 mock。后端消息 API 完整。

**需要改造:**

| 功能 | API | 描述 |
|------|-----|------|
| 查询消息 | `GET /api/messages/` | 加载消息列表 |
| 发送消息 | `POST /api/messages/` | 发送文本消息 |
| 标记已读 | `POST /api/v2/message/{id}/read/` | 消息已读状态 |
| 聊天室 | `GET /api/v2/chatroom/` | 全局聊天室消息 |
| 联系人 | `GET /api/v2/contacts/` | 联系人列表 |

**改造:** 修改 `ChatPanel.tsx`，将对 mock 数据的引用替换为 API hook。

---

### 9. WebSocket 实时推送 — 新功能（先不做）

**当前状态:** 前端 `websocket.ts` 是空壳，后端无 WS 端点。

**需要实现:**

| 功能 | 描述 |
|------|------|
| 后端 WS 端点 | 新增 `/ws` 端点，推送市场行情、生产完成、通知 |
| 前端 WS 连接 | 实现 `useMarketWebSocket`、`useProductionWebSocket` 实际逻辑 |
| 断线重连 | 指数退避重连 |

---

### 10. 资产补齐（这个不用我之后自己弄）

| 缺失项 | 当前状态 | 建议 |
|--------|---------|------|
| 资源物品图 | 23 资源中只有 4 个有图 | 补充至少食品链核心资源（Steak, Cheese, Pizza, Dough, Butter, Sugar 等）|
| 资源图标映射 | `resourceIcon()` 函数只有 4 个 case | 补充全部资源 ID 映射，无图的用兜底图标 |
| InventoryBar 表情映射 | `RESOURCE_ICONS` 只有 6 个 emoji | 替换为实际 icon PNG |
| 市场分组图标 | `icon_restaurant_v1.png` 未实际使用 | 可以给 Kitchen Chain 用 |

---

### 11. 次要加固 （这个可以做）

| 问题 | 文件 | 修复 |
|------|------|------|
| BuildView 中建筑市场 API 路径 | `BuildView.tsx:25` | 实际调的是 `/api/v2/buildings/` 而非 `/api/v2/buildings/market/`，后端 `RegisterBuildingShop` 只注册了 `/api/v2/buildings/market/`，这段调用的 handler 是 production.go 的 `handleV2Buildings` — 需要确认是否走对了 |
| 前端 `icon_timer_v1.png` 用于 Research | 语义上更配 timer 还是 research？可能需要专门加个 research icon |
| TopBar 能量和工人为硬编码 | 120/120, 8/10 | 后端无此概念，要么前端本地算，要么后端新增字段 |

---

## Backend Completion Plan

### 1. WebSocket 端点 — 新增 （先不做）

**当前状态:** 无。

**需要实现:**

| 功能 | 描述 |
|------|------|
| `/ws` 端点 | 使用 `gorilla/websocket` 或 `nhooyr.io/websocket`（Go 1.25 建议选后者）|
| 广播市场成交 | 每次 `executeMatch` 后广播 ticker price 更新 |
| 广播生产完成 | 生产状态变更时推送给指定 company |
| 广播通知 | 新通知/消息推送 |
| 连接管理 | 连接池 + company ID 绑定 + 心跳保活 |

**影响文件:** `internal/handler/websocket.go`（新建），`internal/service/`（加 notify 方法），`go.mod`（加 ws 依赖）

---

### 2. 支持多个 Player 与真实认证 （要做）

**当前状态:** 登录/注册用 username 做 key，dev mode 固定 "dev-player"，token 无签名验证。

**需要实现:**

| 功能 | 描述 |
|------|------|
| 密码认证 | bcrypt/scrypt 哈希替代纯 username |
| JWT 签名 | 生成/验证签名 JWT（当前 token 用 `base64(username)` 伪造）|
| 多玩家隔离 | `RegisterPlayer` 应允许多个不同玩家各有多公司 |
| Player ID 查询 | 按 token 解析 playerId，非 dev 硬编码 |

**影响文件:** `internal/handler/auth.go`, `internal/service/service.go`（ValidateToken/Login/Register）, `internal/middleware/auth.go`

---

### 3. 数据持久化完善 （要做）

**当前状态:** 内存模式正常运行，PostgreSQL 实现 (`storage/postgres.go`) 需要确认是否完整。

**需要审计:**

| 问题 | 描述 |
|------|------|
| Postgres 实现完整性 | 确认 `postgres.go` 全部 CRUD 是否已实现 |
| 迁移/自动建表 | 首次运行时自动建表 |
| 定期保存 | Scheduler 的 `SaveAll()` 是否支持 pg 导出 |
| 备份策略 | 存档快照 |

**影响文件:** `internal/storage/storage.go`, `internal/storage/postgres.go`

---

### 4. 配置完善 要做

**当前状态:** 当前 `config/config.go` 读取 27 个环境变量。

**需要评估:**

| 项 | 状态 |
|----|------|
| 是否所有游戏参数都在 `game.json` 中可调？ | ✅ 完整 |
| 环境变量是否过多？ | 27 项适中 |
| 需要配置文件热重载？ | 当前不需要，重启即可 |

**当前无需改动。**

---

### 5. 缺少的 API 功能 / 改进 要做

| 功能 | 当前状态 | 建议 |
|------|---------|------|
| 天气系统 | `/api/v2/weather/` 返回固定值 | 可以加入随机/周期性天气变化，影响生产效率 |
| 生产修正 | `/api/v2/production-modifiers/` 返回固定值 | 同上，结合天气动态变化 （没想好先不做）|
| 反作弊 | `anticheat/detector.go` 需要审查 | 确认检测逻辑的合理性 已经有了不用做|
| 反洗钱 | `aml/aml.go` 需要审查 | 确认交易监控是否生效已经有了不用做 |
| 排行榜 | 无 | 可以新增 `/api/v2/leaderboard/` 按净资产/等级排名 要做还有单独新建页面 |
| 搜索/筛选 | 市场订单、高管、债券均不支持按条件搜索 | 加分页和查询参数 |
| API 版本对齐 | v1/v2/v3 混用 | 建议逐步迁移到 v3 统一前缀 |

---

### 6. 经济模型平衡性

**当前状态:** 公式系统通过 `game.json` 参数可调。

**需要审计:**

| 公式 | 文件 | 问题 |
|------|------|------|
| 生产效率 | `formula/production.go` | 确认 `BaseProductionRate` 的衰减符合预期 |
| 债券利率 | `formula/bonds.go` | 确认 `DailyBondInterest` 的 50x 系数是否合理 |
| 机器人市场 | `service/market_competition.go` | 确认机器人不会完全压制玩家定价 |
| 政府合同 | `service/government.go` | 确认授标算法公平 |
| 市场锁定 | `service/market_trade.go` | `market_lock_threshold` 逻辑是否生效 |

---

### 7. 单元测试覆盖

**当前状态:** 仅 8 个测试文件（handler_test, market_test, production_test, 等）。

| 模块 | 现有测试 | 建议 |
|------|---------|------|
| formula/ | 0 | 纯函数最容易测，应全覆盖 |
| service/market_trade | market_test.go | 关键路径，需要覆盖边界 |
| service/bond | bond_test.go | 确认利息、违约逻辑 |
| service/government | government_test.go | 竞标、授标、交付闭环 |
| service/production | production_test.go | 生产链、原料消耗、品质 |
| anticheat/ | 0 | 速率检测、作弊模式 |
| aml/ | aml_test.go | 洗钱检测 |

---

### 8. 杂项

| 问题 | 描述 |
|------|------|
| 市场订单 ID 冲突 | `model.MarketOrder.ID` 用 `fmt.Sprintf(...)` 生成，多线程下可能冲突？`sync.Mutex` 已保护 |
| 跨域配置 | middleware CORS 需要确认前端部署域名白名单 |
| 日志级别 | 当前只有 `log.Printf`，无级别区分 |
| graceful shutdown | `main.go` 已有信号处理 ✅ |
| 配置校验 | 启动时 `config.Load()` 是否校验了所有必填项？ |
