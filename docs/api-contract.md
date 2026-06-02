# API Contract

> 生成时间: 2026-06-01
> 总路由数: 101
> 稳定等级: 🟢 稳定 / 🟡 实验 / 🔴 废弃风险 / ⚪ 开发/调试

---

## 🟢 稳定 (33 routes)

这些接口行为已锁定，返回结构以 typed struct 为主，handler 测试覆盖率需后续补充。

### Auth

| Method | Path | Request | Response |
|--------|------|---------|----------|
| POST | `/api/register` | `{"username":"str"}` | `{"player":Player,"company":Company,"companyID":int}` |
| POST | `/api/login` | `{"username":"str"}` | `{"player":Player,"company":Company}` |
| GET | `/healthz` | — | `{"status":"ok"}` |
| GET | `/readyz` | — | `{"status":"ready"}` |

### Market Orders

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET/POST | `/api/v2/market-order/` | `{"resourceId":int,"kind":0\|1,"quality":int,"quantity":int,"price":float}` | `{"order":MarketOrder}` |
| POST | `/api/v2/market-order/cancel/` | `{"orderId":"str"}` | `{"order":MarketOrder,"cancelled":true}` |
| GET | `/api/v2/market-order/take/` | `{"resourceId":int,"quantity":int,"quality":int,"maxPrice":float}` | `{"amountBought":int,"trades":[]Trade,"moneyDelta":float}` |
| GET | `/api/v3/market-ticker/{id}/` | — | `map[string]any` (ticker data) |
| GET | `/api/v3/market-depth/{id}/{quality}/` | — | `{"buys":[],"sells":[]}` |
| GET | `/api/v3/market/{id}/` | — | `{"orders":[]MarketOrder}` |
| GET | `/api/market/buy/orders/` | — | `[]MarketOrder` |
| GET | `/api/v3/resources/` | — | `{"resources":[]map}` |
| GET | `/api/v3/resources-info/{id}/` | — | `map[string]any` |

### Production

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/production/jobs/` | — | `{"jobs":[]ProductionJob}` |
| GET/POST | `/api/v2/production/claim/` | `{"jobId":"str"}` | `{"jobId":"str","status":"str","output":"str","quality":int}` |
| GET | `/api/v2/production/claimable/` | — | `{"claimable":[]ProductionJob}` |
| POST | `/api/v2/production/claim-all/` | — | `{"claimed":int,"results":[]map}` |
| GET | `/api/v2/production/queue/` | — | `{"queue":[]ProductionJob,"maxSlots":int,"usedSlots":int}` |
| POST | `/api/v2/production/slots/add/` | — | `{"slots":int,"cost":float}` |
| POST | `/api/v2/production/cancel/` | `{"jobId":"str"}` | `{"jobId":"str","status":"cancelled"}` |

### Buildings

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/buildings/market/` | — | `[]map` (available buildings) |
| POST | `/api/v2/buildings/buy/` | `{"buildingId":"str"}` | `{"building":map,"money":float}` |
| POST | `/api/v2/buildings/place/` | `{"buildingId":"str","x":int,"y":int}` | `{"building":map,"money":float}` |
| POST | `/api/v2/buildings/move/` | `{"buildingId":"str","x":int,"y":int}` | `{"building":map}` |
| POST | `/api/v2/buildings/demolish/` | `{"buildingId":"str"}` | `{"building":map,"money":float}` |
| POST | `/api/v2/companies/me/warehouse/upgrade/` | — | `{"level":int,"capacity":int,"cost":float,"money":float}` |
| GET | `/api/v1/buildings/{id}/` | — | `{"building":map}` (legacy) |
| GET | `/api/v2/companies/me/buildings/` | — | `[building]` |
| GET | `/api/v2/companies/me/administration-overhead/` | — | `{"baseOverhead":float,"cooSkill":float,"multiplier":float}` |

### Daily Orders

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/orders/daily/` | — | `{"orders":[]Order,"date":"str"}` |
| POST | `/api/v2/orders/daily/complete/{id}/` | — | `{"id":"str","status":"completed"}` |
| POST | `/api/v2/orders/daily/claim/` | — | `{"id":"str","cash":float,"xp":int,"money":float}` |

### Bonds

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/bonds/` | — | `[]map` |
| POST | `/api/bonds/settle-interest/` | — | `{"dailyBondIncome":float,"dailyBondExpense":float,"defaults":[]}` |
| GET | `/api/v2/companies/me/bonds/owned/` | — | `[]map` |
| GET | `/api/v2/companies/me/bonds/sold/` | — | `[]map` |
| POST | `/api/bonds/issue/` | `{"amount":int,"interest":float}` | `{"bond":map}` |
| POST | `/api/bonds/buy/` | `{"id":"str","amount":int}` | `{"bond":map,"money":float}` |
| POST | `/api/bonds/call/` | `{"id":"str","amount":int}` | `{"bond":map,"money":float}` |

### Government

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v3/government-orders/` | — | `[]GovContract` |
| POST | `/api/v3/government-orders/bid/` | `{"contractId":"str","unitPrice":float}` | `{"contract":GovContract}` |
| POST | `/api/v3/government-orders/deliver/` | `{"contractId":"str"}` | `{"contract":GovContract,"reward":float}` |
| GET/POST | `/api/v3/government-orders/award/` | — | `{"awarded":[]GovContract}` |

### Level / XP / Boost

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/players/me/level/` | — | `{"level":int,"currentXp":int,"xpToNextLevel":int,"buildingSlots":int,"buildingsUsed":int}` |
| POST | `/api/v2/players/me/xp/` | — | `{"xp":int,"level":int}` |
| GET | `/api/v2/players/me/level-rewards/` | — | `{"rewards":[]map}` |
| GET | `/api/v2/players/simboosts/` | — | `[]map` |
| POST | `/api/v2/players/simboosts-use/` | `{"boostId":"str"}` | `{"boost":map}` |
| GET | `/api/v2/players/me/offline-income/` | — | `{"offlineIncome":float,"money":float}` |

### Research

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/research/` | — | `{"projects":[]ResearchProject}` |
| POST | `/api/v2/research/start/` | `{"kind":int}` | `{"project":ResearchProject}` |
| GET | `/api/v2/research/progress/` | — | `{"progress":[]map}` |
| POST | `/api/v2/research/complete/` | `{"id":"str"}` | `{"completed":true,"project":ResearchProject}` |

### Executive

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/executives/search/` | — | `[]map` |
| POST | `/api/v2/executives/recruit/` | `{"execId":"str"}` | `{"exec":map}` |
| POST | `/api/v2/executives/train/` | `{"id":"str"}` | `{"exec":map,"cost":float}` |
| POST | `/api/v3/executives/poach/` | `{"execId":"str","targetCompanyId":int}` | `{"exec":map,"cost":float}` |
| GET/POST | `/api/v3/executives/offers/` | `{"offerId":"str","accept":bool}` | `{"exec":map}` |
| GET | `/api/v3/executives/{id}/` | — | `{"exec":map}` |

### Recipes

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/recipes/` | — | `[]map` |
| GET | `/api/v2/recipes/{id}/` | — | `{"recipe":map}` |

### Warehouse

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/companies/me/warehouse/` | — | `{"inventory":[],"capacity":int,"used":int}` |

### Other Company

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/companies/{id}/` | — | `{"company":map}` |
| GET | `/api/v3/companies/{id}/` | — | `{"company":map}` |
| GET | `/api/v2/players/me/companies/` | — | `[]Company` |
| GET | `/api/v2/companies/me/auctions/` | — | `{"auctions":[]Auction}` |
| GET/POST | `/api/v2/auctions/{id}/` | bid: `{"amount":float}` | `{"auctions":[]Auction}` or `{"auction":Auction}` |
| PATCH | `/api/v2/companies/me/tags/` | `{"tags":[]str}` | `{"ok":true}` |

### Production Modifiers

| Method | Path | Request | Response |
|--------|------|---------|----------|
| GET | `/api/v2/production-modifiers/` | — | `{"boostMultiplier":float,"offlineMultiplier":float}` |
| GET | `/api/v2/weather/` | — | `{"weather":"str","season":"str"}` |

---

## 🟡 实验 (34 routes)

这些接口功能可用但返回结构不稳定（含 `map[string]any`），或缺少错误处理路径。

### Financial (mock historical data)

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/v2/companies/me/income-statement/` | 返回 `map[string]any`，结构依赖 Service.FinancialStatements |
| GET | `/api/v2/companies/me/balance-sheet/` | 同上 |
| GET | `/api/v2/companies/me/cashflow-statement/` | 同上 |
| GET | `/api/v2/companies/me/cashflow/recent/` | 含 mock weeklyNet 数据 |
| GET | `/api/v2/companies/me/past-finances-overview/` | **纯 mock**：硬编码 `{"weeklyNet":[2100,3800,2750,4200]}` |
| GET | `/api/v3/companies/me/past-finances/` | **纯 mock**：硬编码 series 数据 |

### Message / Chat

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/messages/` | 简化实现 |
| GET | `/api/messages_by_company/` | 简化 |
| GET/POST | `/api/v2/message/` | send/read 功能可用但返回 map |
| PATCH | `/api/v2/chatroom/` | 聊天室操作用 `body := map[string]any{}` |
| GET | `/api/v2/contacts/` | 返回空数组占位 |
| GET | `/api/v2/newspaper/articles-by-author/` | 占位，返回空数组 |
| GET | `/api/v2/newspaper/articles/` | 占位，`map[string]any` 混杂 |
| GET | `/api/v2/newspaper/publishing-costs/` | 占位，硬编码数据 |

### Achievements

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/v2/companies/me/achievements/` | 占位，返回空 |
| GET | `/api/v2/no-cache/companies/me/achievements/` | 占位 |
| DELETE | `/api/v2/no-cache/companies/achievements/` | 占位 |
| GET | `/api/v2/players/unlocked-hqs/` | 占位 |
| GET | `/api/v2/players/devices/` | 占位 |

### Contracts (internal)

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/v3/contracts-incoming/` | 占位 |
| GET | `/api/v3/contracts-outgoing/me/` | 占位 |
| GET | `/api/v2/contracts-history-incoming/` | 占位 |
| GET | `/api/v2/contracts-history-outgoing/` | 占位 |
| GET | `/api/v2/warehouse-contracts-summary/` | 占位 |

### Miscellaneous

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/csrf/` | 返回固定 token，简化实现 |
| GET | `/api/v2/players/{id}/` | 按 ID 查玩家 |
| POST | `/api/v2/companies/me/preferences/` | `body := map[string]any{}` |
| GET | `/api/v2/companies/me/collectibles/` | 占位 |
| GET | `/api/v2/companies/me/notifications/` | 占位 |
| GET | `/api/v2/companies/me/market-orders/` | 占位 |
| GET | `/api/v2/companies/me/certificates/` | 占位 |
| GET | `/api/v2/companies/me/display-case/` | 占位 |
| GET | `/api/v2/companies/me/former-executives/` | 占位 |
| GET | `/api/v2/companies/me/royalties/` | 占位 |
| GET | `/api/v2/companies/me/egg-collection/` | 占位 |
| PATCH | `/api/v2/research/progress/` | 返回空 |

### Production (legacy)

| Method | Path | Notes |
|--------|------|-------|
| GET/PATCH | `/api/v1/buildings/{id}/` | legacy v1 endpoint，`body := map[string]any{}` |

---

## 🔴 废弃风险 (15 routes)

这些接口要么是开发调试专用不应出现在生产，要么是 mock/空壳已不对齐实际游戏逻辑。

### Dev-only (should be disabled in production)

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/dev/ledger/` | 返回全量 ledger，调试用 |
| GET | `/api/dev/formulas/production/` | 公式调试 |
| GET | `/api/dev/formulas/retail/` | 公式调试 |
| GET | `/api/dev/formulas/retail-season-weather/` | 公式调试 |
| GET/POST/DELETE | `/api/dev/time/` | 模拟时间快进，破坏游戏公平性 |
| DELETE | `/api/v2/no-cache/companies/achievements/` | 名不副实，实际是删除成就 |

### V4 Aggregation (experimental dump)

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/v4/` | 把所有 state 和 config 一次性 dump 出来，结构不确定 |

### Aerospace (mock)

| Method | Path | Notes |
|--------|------|-------|
| GET | `/api/v2/aerospace/projects/` | 简化实现 |
| POST | `/api/v2/aerospace/projects/create/` | 简化 |
| GET | `/api/v2/aerospace/launches/` | 占位 |
| POST | `/api/v2/aerospace/launch/` | 简化 |
| GET | `/api/v2/aerospace/components/` | 占位 |

### Government (admin)

| Method | Path | Notes |
|--------|------|-------|
| POST | `/api/v3/government-orders/resolve-defaults/` | 管理端操作，无访问控制 |

---

## ⚪ 开发/调试 (19 routes)

### Government

| Method | Path | Notes |
|--------|------|-------|
| POST | `/api/v3/government-orders/award/` | 手动触发合同 award，需要管理员认证（尚无） |

---

## 风险指标

| 风险类型 | 数量 | 说明 |
|----------|------|------|
| 静默 JSON 解码错误 | 0 ✅ | 上次修复已清零 |
| `body := map[string]any{}` | 1 ⚠️ | `production.go` 遗留 |
| 返回 `map[string]any` 较多 | 10+ | financial/bond/message 等 |
| 忽略请求参数 (`_ *http.Request`) | 22+ | 这些 handler 拿不到 companyID |
| TODO/FIXME | 0 | 代码中无 TODO |
| 占位返回（空数组/硬编码） | 12+ | mock 数据主要在 financial/message/achievement |

## 稳定化路径

1. **P0** - 给所有 stable handler 补充契约测试（验证 request/response 结构）
2. **P1** - 消灭剩余 `body := map[string]any{}`
3. **P1** - financial mock 数据替换为真实计算
4. **P2** - aerospace/message/achievement 空壳实现
5. **P3** - `/api/dev/*` 编译期 feature flag 隔离（`//go:build !production`）
