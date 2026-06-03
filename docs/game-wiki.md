# Go Sim API 游戏Wiki

> 基于 Sim Companies 游戏经济的 Go API 复刻版。涵盖 14 个子系统、50+ 端点，纯 Go 标准库实现。

---

## 1. 公司系统 (Company)

### 概述
公司是玩家的核心实体。每个玩家可拥有多家公司，每家公司独立维护资金 `money`、库存 `inventory`、等级 `level`、建筑 `placedBuildings` 等属性。外部经济由两个机器人公司参与市场提供流动性。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v3/companies/{id}/`|公司详情（含资金/等级/库存/建筑）|
|GET|`/api/v3/companies/{id}/executives/`|高管列表示例|
|GET|`/api/v2/companies/me/buildings/`|已放置的建筑列表|
|GET|`/api/v2/companies/me/administration-overhead/`|行政开销（受COO技能影响）|
|GET|`/api/v2/companies/me/game-notifications/`|游戏内通知列表|
|GET|`/api/v2/companies/me/market-orders/`|公司的市场挂单|
|GET|`/api/v2/companies/me/achievements/`|成就列表|
|GET|`/api/v2/players/me/companies/`|玩家的公司列表|
|POST|`/api/v2/players/{id}/preferences/`|保存偏好设置|
|GET|`/api/csrf/`|获取 CSRF Token|
|GET|`/api/v2/companies/me/collectibles/`|收藏品（存根）|
|GET|`/api/v2/companies/me/certificates/`|证书（存根）|
|GET|`/api/v2/companies/me/display-case/`|展示柜（存根）|
|GET|`/api/v2/companies/me/former-executives/`|前高管记录（存根）|
|GET|`/api/v2/companies/me/royalties/`|版税记录（存根）|
|GET|`/api/v2/companies/me/egg-collection/`|彩蛋收集（存根）|
|GET/POST|`/api/v2/companies/{id}/tags/`|公司标签|

### 数据模型

```go
type Company struct {
    ID              int                `json:"id"`
    Name            string             `json:"name"`
    Money           float64            `json:"money"`
    Level           int                `json:"level"`
    Inventory       map[int]int        `json:"inventory"`       // {resourceId: quantity}
    PlacedBuildings []map[string]any   `json:"placedBuildings"` // 已放置的建筑
}
```

### 业务逻辑
- 公司的所有操作通过 `sync.Mutex` 保护，状态完全在内存中维护
- `Inventory` 的 key 为资源ID，value 为持有数量
- 资金以 `float64` 存储；扣款时需检查 `money >= 0`
- `handleV2Companies` 作为路由分发器，根据路径后缀派发到不同子处理器（collectibles、notifications、market-orders 等）

---

## 2. 财务系统 (Financial)

### 概述
三大财务报表从流水账（Ledger）实时汇总生成。流水账每笔记录 kind（类型）、amount（金额）、direction（in/out）和自定义元数据。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v2/companies/me/income-statement/`|利润表：总收入 vs 总支出|
|GET|`/api/v2/companies/me/balance-sheet/`|资产负债表：资产=现金+库存×10−负债，负债=已发行债券总额|
|GET|`/api/v2/companies/me/cashflow-statement/`|现金流量表：经营/投资/筹资三类|
|GET|`/api/v2/companies/me/cashflow/recent/`|最近流水记录|
|GET|`/api/v2/companies/me/bonds/owned/`|持有的债券|
|GET|`/api/v3/contracts-incoming/`|合同收入汇总|
|GET|`/api/v3/contracts-outgoing/me/`|合同支出汇总|
|GET|`/api/dev/ledger/`|完整流水账查询（开发用）|

### 财务报表生成逻辑

**利润表**：遍历 Ledger，direction=in 累加 revenue，direction=out 累加 expenses。

**资产负债表**：
- 资产 = `money` + `Σ (inventory[qty] × 10)` + 持有的 Bonds 面值
- 负债 = `Σ (已发行Bonds数量 × BondFaceValue)`
- 权益 = 资产 − 负债

**现金流量表**：按 Ledger 记录的 kind 前缀分类：
- `production_*`, `market_trade_*` → 经营活动
- `bond_*` → 筹资活动
- 其余 → 投资活动

### 流水账数据结构

```go
// addLedger(kind, amount, direction, meta)
entry := map[string]any{
    "id":        uuid,
    "kind":      kind,        // 如 "market_trade_sell", "bond_issue", "gov_bid_deposit"
    "amount":    amount,
    "direction": direction,   // "in" 或 "out"
    "meta":      meta,        // 自定义附加数据
    "createdAt": now,
}
```

---

## 3. 资源系统 (Resources)

### 概述
游戏内有 151 种资源，每项资源定义其基础价格、是否可交易、经济模型参数等。数据从 `decompiled/data/resources.json` 加载。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v3/resources/{id}/`|资源详情（通过 `dbLetter` 字段匹配ID）|
|GET|`/api/v3/market-ticker/{id}/`|K线数据（含机器人流动性刷新）|

### 数据来源

资源定义加载自静态 JSON：
- `resources.json`：151 种资源的定义（名称、基础价格、`dbLetter`、`dbLetter2` 等）
- `economy_model.json`：三种经济状态下每项资源的产能/成本/销量参数
- `resource_lookups.json`：资源 ID 与名称的映射查找表

### 重要资源ID

|ID|资源名称|说明|
|---|---|---|
|1|Power|电力，基础能源|
|2|Water|水，基础资源|
|3|Apples|水果|
|4|Wheat|小麦，基础农业|
|8|Beef|牛肉，高价值食品|
|9|Steak|牛排，牛肉加工品|
|10|Bread|面包|
|11|Cake|蛋糕|
|12|Pizza|披萨|

### 资源详情字段

`GET /api/v3/resources/{id}/` 返回的字段（从 `resources.json` 直接输出）：
- `dbLetter`: 资源 ID（整数）
- `name`: 名称
- `dbLetter2`: 子类型/品质标识
- `basePrice`: 基础价格
- `isTradeable`: 是否可交易
- `isHQ`: 是否总部资源
- `isCollectible`: 是否收藏品

---

## 4. 市场系统 (Market)

### 概述
限价订单簿（Limit Order Book）。买卖双方挂单，价格匹配时自动撮合。两个机器人持续提供流动性，日循环模拟真实市场行为。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v3/market-ticker/{id}/`|K线数据（最近48小时，含机器人流动性刷新）|
|GET|`/api/v3/market/{resource}/{quality}/`|订单簿（指定资源和品质）|
|POST|`/api/v2/market-order/`|创建挂单（kind=1 买单, kind=0 卖单）|
|DELETE|`/api/v2/market-order/cancel/{id}/`|撤单|
|POST|`/api/market/buy/orders/{resource}/{quality}/`|买单快捷接口|
|POST|`/api/v2/market-order/take/`|市价吃单|
|GET|`/api/v2/weather/{id}/`|天气（影响零售 sellingSpeed）|
|GET|`/api/v2/production-modifiers/`|生产修正系数|

### 创建订单请求体

```json
{
  "resourceId": 8,
  "kind": 1,
  "quality": 1,
  "quantity": 100,
  "price": 12.50
}
```

### 吃单请求体

```json
{
  "resource": 8,
  "quantity": 50,
  "quality": 1,
  "maxPrice": 13.00
}
```

### 订单簿数据结构

```go
type MarketOrder struct {
    ID         string  `json:"id"`
    ResourceID int     `json:"resourceId"`
    Kind       int     `json:"kind"` // 0=卖单, 1=买单
    Price      float64 `json:"price"`
    Quality    int     `json:"quality"`
    Quantity   int     `json:"quantity"`
    Remaining  int     `json:"remaining"`
    CompanyID  int     `json:"companyId"`
    CreatedAt  string  `json:"createdAt"`
}

type Trade struct {
    ID          string  `json:"id"`
    ResourceID  int     `json:"resourceId"`
    Quality     int     `json:"quality"`
    Quantity    int     `json:"quantity"`
    Price       float64 `json:"price"`
    BuyOrderID  string  `json:"buyOrderId"`
    SellOrderID string  `json:"sellOrderId"`
    CreatedAt   string  `json:"createdAt"`
}
```

### 撮合逻辑

1. **排序**：买单按价格从高到低排序，卖单从低到高（同价则按时间优先）
2. **成交条件**：买单 `price >=` 卖单 `price` 时撮合
3. **成交价**：以卖单价格成交（`price = sell.price`）
4. **数量**：取买卖双方的较小剩余量
5. **手续费**：`ceil(数量 × 价格 × ExchangeFeePct)`（默认 4%）
6. **资金结算**：买方获得库存，卖方获得资金（扣除手续费）
7. **订单状态**：完全成交后从订单簿移除；部分成交更新 `Remaining`

### 机器人流动性

每次查询 K 线（`/api/v3/market-ticker/{id}/`）时自动触发 `RunBotMarketCycle()`：

- 按小时周期模拟供需波动：`cycleVol = 1.0 + BotCycleAmplitude × sin(hour / 24 × 2π)`
- 覆盖资源 8~12（牛肉、牛排、面包、蛋糕、披萨）
- 每个资源双向挂单（买一档 + 卖一档），价差约 5%
- 每日清理过期订单，总数上限 `MaxBotOrders`（默认 600）

### 价格 Tick

价格必须落在合法 Tick 步长上，否则拒绝：

|价格区间|Tick步长|
|---|---|
|≥ 20000|500|
|10000 ~ 19999.99|100|
|5000 ~ 9999.99|25|
|1000 ~ 4999.99|10|
|500 ~ 999.99|5|
|200 ~ 499.99|2|
|100 ~ 199.99|1|
|50 ~ 99.99|0.5|
|20 ~ 49.99|0.25|
|5 ~ 19.99|0.1|
|2 ~ 4.99|0.05|
|1 ~ 1.99|0.01|
|0.5 ~ 0.99|0.005|
|< 0.5|0.001|

### 天气

`GET /api/v2/weather/{id}/` 返回：
- `id`: 所在区域（默认 0）
- `since` ~ `until`: 天气有效期
- `sellingSpeedMultiplier`: 零售速度倍率（默认 1.06）

---

## 5. 生产系统 (Production)

### 概述
玩家通过建筑生产资源。选择配方（资源ID）、投入原材料，等待生产完成后收货。

### 关键端点

|方法|路径|说明|
|---|---|---|
|POST|`/api/v1/buildings/{id}/busy/`|让指定建筑开始生产|
|GET|`/api/v2/production/jobs/`|生产任务列表（自动刷新状态）|
|POST|`/api/v2/production/claim/{jobId}/`|收货完成|

### 开始生产请求体

```json
{
  "kind": 9,
  "amount": 10,
  "estimatedSecondsToFinish": 120
}
```

### 生产任务模型

```go
type ProductionJob struct {
    ID          string         `json:"id"`
    BuildingID  string         `json:"buildingId"`
    ResourceID  int            `json:"resourceId"`
    Amount      int            `json:"amount"`
    Input       map[int]int    `json:"input"`       // {resourceId: consumedQty}
    Output      map[int]int    `json:"output"`      // {resourceId: producedQty}
    StartedAt   string         `json:"startedAt"`
    CompletesAt string         `json:"completesAt"`
    Status      string         `json:"status"`      // running | ready | claimed
    Meta        map[string]any `json:"meta,omitempty"`
}
```

### 业务逻辑

- 开始生产时查找配方消耗，从库存中扣除原材料
- 输入不足时拒绝
- 未指定 `estimatedSecondsToFinish` 时使用 `max(30, amount × 6)` 作为工期
- 查询 `/api/v2/production/jobs/` 时自动将超时任务标记为 `ready`
- 收货时将产出库存添加到公司，流水标记 `production_output`


### 建筑等级 (Building Level)

每个建筑有 `level` 字段，影响产能和升级成本：

| 等级 | 产出倍率 | 累计成本公式 |
|------|---------|------------|
| Lv1  | 1× base | baseCost |
| Lv2  | 2× base | (1+2) × baseCost = 3× baseCost |
| Lv3  | 3× base | (1+2+3) × baseCost = 6× baseCost |
| LvN  | N× base | N(N+1)/2 × baseCost |

#### 升级

```http
POST /api/v1/buildings/{id}/upgrade/
```

- LvN → Lv(N+1) 花费 = (N+1) × baseCost
- `baseCost` = building kind × 5000（kind=2 → $10000）
- Lv3→Lv4 花费 4×$10000 = $40000
- 产出倍率从 3× 提升至 4×

#### 产出公式

实际产出量 = `amount(请求base)` × `buildingLevel`

例如：`POST /api/v1/buildings/b-1/busy/` 请求 `amount=10`，
建筑 Lv3 产出 30 单位，Lv4 产出 40 单位。

#### 竞拍建筑

通过建筑竞拍获得的建筑初始为 Lv1，`baseCost=10000`，与普通建筑一样可升级。
### 核心公式

位于 `internal/formula/production.go`：

**BaseProductionRate** — 薪资修正后的基础产率：
```
baseRate = producedPerHourRaw × (AverageSalary / SalaryMid[salaryLevel]) ^ salaryModifier
```

**ProducedPerHour** — 含规模、薪资、机器人、品质、速度修正的实际产率：
```
adjusted = baseRate
if isMining: adjusted ×= qualityPct / 100
adjusted ×= (speedModifierPct / 100) + 1
effectiveSalary = salaryPercent
if isAccumulator: effectiveSalary += RobotBonus × robotCount
den = max(1 − effectiveSalary / 100, 0.01)
result = size × adjusted / den
```

**ProductionTimeSeconds** — 生产耗时（核心公式）：
```
time = (345 × buildingSalaryModifier × size / producedPerHour) × eventMultiplier
```

其中 `SalaryMid = {0: 655, 1: 700, 2: 745}`，`AverageSalary = 345`，`RobotBonus = 4`。

---

## 6. 债券系统 (Bonds)

### 概述
公司可以发行债券融资，亦可购买他人发行的债券获取利息收入。债券评级分级限制可购买数量。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/bonds/`|债券市场列表（支持 `?rating=` 筛选）|
|PATCH|`/api/bonds/`|发行新债券|
|PATCH|`/api/bonds/{id}/`|购买债券|
|PUT|`/api/bonds/{id}/`|赎回债券|
|POST|`/api/bonds/settle-interest/`|结算每日利息|
|GET|`/api/v2/companies/me/bonds/owned/`|持有的债券|
|GET|`/api/v2/companies/me/bonds/sold/`|已发行的债券|

### 债券数据模型

债券在内存中表示为 `map[string]any`，包含：
- `id`: 债券ID
- `issuerId`: 发行公司ID
- `issuerName`: 发行公司名称
- `rating`: 评级（由发行公司等级决定）
- `interestRatePct`: 利率百分比
- `faceValue`: 面值（默认 5000）
- `totalAmount`: 总发行量（份数）
- `remainingAmount`: 剩余未购份数
- `ownedAmount`: 已购份数（对持有者视角）
- `status`: 状态（active/called/defaulted）
- `issuedAt`: 发行时间

### 核心公式

**面值**：`BondFaceValue = 5000`（可通过环境变量 `SIM_API_BOND_FACE` 调整）

**日利息**（每日结算）：
```
dailyInterest = floor(持有数量 × 50 × 利率%)
```

**期间利息**（用于报价展示）：
```
periodInterest = floor(数量 × BondFaceValue × 利率%) / 100
```

**可发行上限**：
```
maxIssuable = floor(总建筑价值 / BondFaceValue) − 已发行数量
if maxIssuable < 0 → 0
```

### 评级分组

评级从发行公司等级自动计算（示例映射）：
- 等级 60+ → AAA
- 等级 50-59 → AA
- 等级 40-49 → A
- 等级 30-39 → BBB
- 等级 20-29 → BB
- 等级 10-19 → B
- 等级 < 10 → CCC

|评级分组|等级|购买上限（占公司现金%）|
|---|---|---|
|AAA ~ AA|最高|1%|
|A ~ BBB-|中等|3%|
|BB+ 及以下|投机|5%|

### 业务逻辑

**发行**：发行人立即收到 `amount × faceValue` 资金，流水标记 `bond_issue`。

**购买**：利率在 `BondMinInterest ~ BondMaxInterest` 之间（默认 0.5% ~ 2.0%），按评级限制检查可购买量上限。

**赎回**：发行人按面值购回债券，资金从公司扣除。

**利息结算**：遍历所有活跃债券，对每个持有者按日利息公式计算并支付，流水标记 `bond_interest`。

---

## 7. 政府合同 (Government)

### 概述
政府发布采购合同，公司投标竞争，最低价中标。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v3/government-orders/`|可用合同列表|
|POST|`/api/v3/government-orders/bid/`|投标（需交押金）|
|POST|`/api/v3/government-orders/award/`|评标授标|
|POST|`/api/v3/government-orders/deliver/`|交付合同|
|POST|`/api/v3/government-orders/resolve-defaults/`|处理违约|

### 合同数据模型

合同在内存中表示为 `map[string]any`，包含：
- `id`: 合同ID
- `resourceId`: 所需资源ID
- `quantity`: 需求数量
- `unitPrice`: 单位结算价（授标时确定）
- `depositRate`: 投标押金比例
- `status`: 状态（open/awarded/delivered/defaulted）
- `bids`: 竞价列表
- `awardedBidder`: 中标者ID
- `deadline`: 交付截止时间

### 业务流程

**投标** (Bid)：
1. 检查合同状态为 `open`
2. 计算押金 = `quantity × unitPrice × depositRate`
3. 检查公司资金充足并扣除押金
4. 记录投标（同一公司可更新报价）

**授标** (Award)：
1. 对每个 `open` 状态的合同
2. 按 `unitPrice` 升序选取最低价投标者
3. 未中标者退还押金 × `GovBidRefundRate`（默认 80%）
4. 合同状态变更为 `awarded`

**交付** (Deliver)：
1. 检查公司是否有足够库存
2. 扣除库存，支付 `quantity × unitPrice`
3. 合同状态变更为 `delivered`

**违约处理** (ResolveDefaults)：
1. 遍历所有 `awarded` 合同
2. 检查是否超过截止时间未交付
3. 违约则罚没押金，合同状态变更为 `defaulted`

---

## 8. 高管系统 (Executive)

### 概述
高管为公司提供技能加成。通过搜索、招募、培训、挖角等方式获取。

### 关键端点

|方法|路径|说明|
|---|---|---|
|POST|`/api/v2/executives/search/`|搜索可用高管|
|POST|`/api/v2/executives/recruit/`|招聘高管|
|POST|`/api/v2/executives/train/{id}/`|培训高管（提升技能）|
|POST|`/api/v3/executives/poach/`|挖角他人高管|
|GET|`/api/v3/executives/{id}/`|高管详情|
|GET/POST|`/api/v3/executives/offers/`|查看/回复报价|

### 技能系统

|技能|适用高管|效果|
|---|---|---|
|COO (运营)|CEO / COO|降低行政开销：`AdminOverheadWithCOO(base, coo)`|
|CTO (技术)|CTO|提升生产效率：`CTOProductionMultiplier(cto)`|
|CFO (金融)|CFO|影响债券相关操作|

### 核心公式

**行政开销折减**（`internal/formula/admin.go`）：
```
AdminOverheadWithCOO(adminOverhead, cooSkill) = 
    adminOverhead − (adminOverhead − 1) × cooSkill / 100
```

**CTO 生产效率加成**：
```
CTOProductionMultiplier(ctoSkill) = (100 + ctoSkill × 2) / 100
```

---

## 9. 报社系统 (Newspaper)

### 概述
模拟游戏内报纸系统，玩家可以发布和查看文章，影响市场情绪。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v2/newspaper/articles/`|文章列表（分页，支持 `?page=`）|
|POST|`/api/v2/newspaper/articles/`|发布文章|
|GET|`/api/v2/newspaper/articles/{id}/`|文章详情|
|GET|`/api/v2/newspaper/articles-by-author/`|特定作者的文章|
|GET|`/api/v2/newspaper/publishing-costs/`|发布费用查询|

### 文章数据字段

- `id`: 文章ID
- `title`: 标题
- `content`: 正文
- `authorId`: 作者ID
- `authorName`: 作者名称
- `createdAt`: 发布时间
- `tags`: 标签列表
- `readCount`: 阅读数

### 业务逻辑

- 文章列表支持分页，按 `?page=` 参数返回对应页
- `publishing-costs` 返回发布一篇文章所需的费用明细
- `articles-by-author` 按作者ID筛选

---

## 10. 研发系统 (Research)

### 概述
通过研发新科技来提升生产效率，解锁新产品或新功能。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v2/research/`|可研发项目列表|
|POST|`/api/v2/research/start/`|开始研发某个项目|
|GET|`/api/v2/research/progress/`|当前研发进度|
|POST|`/api/v2/research/complete/{id}/`|完成研发|

### 业务逻辑

- `GET /research/` 返回所有研究项目及其状态（locked/available/in_progress/completed）
- `POST /start/` 接受 `projectId` 参数，开始研发
- `GET /progress/` 返回研发进度百分比及预计剩余时间
- `POST /complete/{id}/` 消耗资源完成研发项目

---

## 11. 等级/经验系统 (Level & XP)

### 概述
通过经济活动获取经验值（XP），升级解锁更多功能并获得奖励。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v2/players/me/level/`|当前等级和XP信息|
|POST|`/api/v2/players/me/xp/`|增加XP（测试用）|
|GET|`/api/v2/players/me/level-rewards/{level}/`|指定等级的奖励|

### 等级信息响应

```json
{
  "level": 42,
  "xp": 15000,
  "xpToNext": 20000,
  "totalXp": 850000,
  "title": "生产大亨"
}
```

### 业务逻辑

- 等级最大 60 级
- `POST /xp/` 请求体 `{"xp": 100}`，未指定时默认加 100
- `GET /level-rewards/{level}/` 返回该等级对应的奖励（`level` 必须在 0~60 之间）
- 等级决定解锁内容：新建筑、高级配方、债券评级上限等

---

## 12. SimBoost (加速道具)

### 概述
游戏内的加速道具，可以加快生产速度、提高产量等。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v2/players/simboosts-use/`|可用Boost和活跃Boost列表|
|POST|`/api/v2/players/simboosts-use/`|使用Boost|
|GET|`/api/v2/players/simboosts/`|所有Boost类型定义|

### SimBoost 类型

预定义的 Boost 类型（示例）：
- `speed_boost`: 生产速度提升 2×
- `yield_boost`: 产量提升 50%
- `research_boost`: 研发速度提升 2×
- `revenue_boost`: 销售收入提升 25%

### 业务逻辑

- `GET simboosts-use/` 返回可用 Boost 数量、活跃 Boost 及剩余时间
- `POST simboosts-use/` 激活一个 Boost（消耗库存中的加速道具）
- `GET simboosts/` 返回所有 Boost 类型的定义（名称、效果倍率、持续时间等）

---

## 13. 建筑竞标系统 (Auctions)

### 概述
玩家可以通过拍卖竞拍稀有建筑。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v2/auctions/`|竞拍列表（含分页）|
|GET|`/api/v2/auctions/{id}/`|竞拍详情（含当前最高出价）|
|POST|`/api/v2/auctions/{id}/bid/`|对指定竞拍出价|
|GET|`/api/v2/companies/me/auctions/`|我的竞拍记录|

### 竞拍数据模型

- `id`: 竞拍ID
- `buildingId`: 拍卖的建筑ID
- `buildingName`: 建筑名称
- `currentBid`: 当前最高出价
- `minBid`: 最低出价
- `bidderId`: 最高出价者
- `endTime`: 竞拍结束时间
- `status`: 状态（active/ended/claimed）

### 业务逻辑

- `POST /{id}/bid/` 请求体 `{"amount": 500000}`
- 出价金额必须高于当前最高出价
- `MyAuctionList` 当前返回空列表（存根实现）
- 竞拍状态目前仅 `AvailableAuctionList()` 生成模拟数据

---

## 14. 航空系统 (Aerospace)

### 概述
制造和发射火箭，探索太空。

### 关键端点

|方法|路径|说明|
|---|---|---|
|GET|`/api/v2/aerospace/projects/`|火箭项目列表|
|POST|`/api/v2/aerospace/projects/create/`|创建火箭项目|
|GET|`/api/v2/aerospace/components/`|可用组件列表|
|GET|`/api/v2/aerospace/launches/`|发射记录（含分页）|
|POST|`/api/v2/aerospace/launch/`|执行发射|

### 业务逻辑

- **创建项目**：接受 `name`、`components` 等参数，创建火箭制造项目
- **可用组件**：返回当前可购买的火箭部件清单
- **发射**：消耗火箭组件，记录发射结果
- **发射记录**：返回历史发射列表（含成功/失败状态）

---

## 附录

### 环境变量

|变量名|默认值|说明|
|---|---|---|
|**服务器**|||
|`SIM_API_ADDR`|`127.0.0.1:8088`|监听地址|
|`SIM_API_DATA_DIR`|`decompiled/data`|JSON数据目录|
|`SIM_API_DEBUG`|`false`|调试模式|
|`SIM_API_DATABASE_URL`|`""`|PostgreSQL连接串（空=内存模式）|
|`SIM_API_CSRF_TOKEN`|`dev-csrf-token`|CSRF Token|
|**公司默认值**|||
|`SIM_API_COMPANY_ID`|`1234567`|主公司ID|
|`SIM_API_COMPANY_NAME`|`Example Company Inc`|主公司名称|
|`SIM_API_START_MONEY`|`200000`|初始资金|
|`SIM_API_START_LEVEL`|`42`|初始等级|
|**机器人**|||
|`SIM_API_BOT1_ID`|`900001`|机器人1 ID|
|`SIM_API_BOT2_ID`|`900002`|机器人2 ID|
|`SIM_API_BOT1_NAME`|`Atlas Trading Bot`|机器人1名称|
|`SIM_API_BOT2_NAME`|`Nova Market Bot`|机器人2名称|
|`SIM_API_BOT_MONEY`|`5000000`|机器人初始资金|
|`SIM_API_BOT_LEVEL`|`99`|机器人等级|
|**经济参数**|||
|`SIM_API_FEE_PCT`|`0.04`|市场手续费率（4%）|
|`SIM_API_ADMIN_BASE`|`1.35`|行政开销基准值|
|`SIM_API_BOND_FACE`|`5000`|债券面值|
|`SIM_API_BOND_MIN_PCT`|`0.5`|债券最低利率（%）|
|`SIM_API_BOND_MAX_PCT`|`2.0`|债券最高利率（%）|
|`SIM_API_GOV_REFUND_RATE`|`0.8`|政府合同未中标押金退还比例|
|`SIM_API_WEATHER_SPEED`|`1.06`|天气影响零售速度倍率|
|`SIM_API_PROD_MOD`|`1.02`|生产修正系数|
|`SIM_API_MAX_BOT_ORDERS`|`600`|机器人订单上限|
|`SIM_API_MAX_LEDGER`|`5000`|流水账最大条目数|
|`SIM_API_BOT_CYCLE_AMP`|`0.06`|机器人价格周期振幅|
|`SIM_API_BOT_BASE`|`8.0`|机器人基础价格偏移|

### 技术栈

|层|选型|
|---|---|
|语言|Go 1.22+|
|HTTP|标准库 `net/http.ServeMux`|
|存储|内存（默认）/ PostgreSQL（可选，通过 `pgx/v5`）|
|中间件|Recovery, Logger, CORS, RequestID|
|并发控制|`sync.Mutex`|
|数据来源|反编译 JSON（`decompiled/data/`）|

### 项目结构

```
go-sim-api/
├── cmd/simapi/main.go         # 入口
├── internal/
│   ├── config/                # 配置（27个环境变量）
│   ├── middleware/            # HTTP中间件
│   ├── handler/               # HTTP处理器（10个域文件）
│   ├── service/               # 业务逻辑（6个域文件）
│   ├── storage/               # 持久化接口（pgx/v5）
│   ├── model/                 # 数据类型
│   ├── data/                  # JSON数据加载
│   └── formula/               # 纯函数公式（production/market/bonds/admin/retail）
├── docs/                      # 文档
├── decompiled/                # 反编译参考数据（只读）
└── go.mod
```
