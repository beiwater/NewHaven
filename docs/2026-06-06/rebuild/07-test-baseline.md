# 07 — 测试基线文档

> Phase 1 只读分析，不修改业务代码。本文档记录当前测试状态、覆盖缺口以及各模块的测试计划。

---

## 1. 当前测试状态

### `go test ./...` — 全部通过

```
ok  go-sim-api/internal/aml         (cached)
ok  go-sim-api/internal/anticheat   (cached)
ok  go-sim-api/internal/config      (cached)
ok  go-sim-api/internal/formula     (cached)
ok  go-sim-api/internal/handler     1.846s
ok  go-sim-api/internal/service     5.660s
ok  go-sim-api/internal/storage     1.545s
ok  go-sim-api/tests                (cached)
```

所有 8 个包均通过，无编译失败、无超时、无 panic。

### `go vet ./...` — 无输出（无警告）

`go vet` 对全部包静默退出，代码为 0。无任何 vet 告警。

### 现有测试文件分布

| 包 | 文件数 | 规模 | 覆盖方向 |
|---|---|---|---|
| `internal/formula/` | 1 | 19 KB | 生产、债券、饱和度、成本、零售、市场 tick |
| `internal/service/` | ~12 | 各数百行 | 公司、认证、生产、市场、债券、政府、排行榜、订单、并发 |
| `internal/handler/` | 1 | 3.7 KB | healthz、CSRF、公司资料、市场深度、建造 |
| `internal/config/` | 1 | — | 默认配置一致性、时间解析、JSON 同步 |
| `internal/storage/` | 1 | — | NoopStorage、GameState/Company/MarketOrder/LedgerEntry JSON round-trip |
| `internal/anticheat/` | 1 | — | 反作弊检测 |
| `internal/aml/` | 1 | — | AML 大额/快速交易检测 |
| `internal/tests/` | 1 | — | 脚本检测 |

---

## 2. 测试计划

以下各节描述 Phase 1 之后 **需要新增** 的测试。不修改现有测试，仅在现有基础上补充。

---

### 2.1 Formula 黄金测试

**目标：** 每个公开公式函数至少有一个黄金测试，用已知输入验证精确输出，防止回归时公式意外变更。

**来源文件：** `backend/internal/formula/*.go`

| 函数 | 输入组合 | 预期输出 |
|---|---|---|
| `OutputPerHour` | `(100, 10, 1)` → 110.0；`(100, 0, 5)` → 500.0 | 与 spec v1.3.1 一致 |
| `ProductionDurationSeconds` | `(10, 3600, 1, 1.0)` → 36000；`(10, 3600, 2, 1.5)` → 12000 | 时间/层级/加速 |
| `DailyBondInterest` | `(1, 1.2)` → 60.0；`(10, 0.5)` → 250.0 | FaceValue=5000 |
| `MaxIssuableBonds` | `(5000000, 500)` → 500；`(100000, 0)` → 20 | |
| `SaturationPriceMultiplier` | `(1.0, 0.15)` → 1.0；`(2.0, 0.15)` → 0.85；`(0.5, 0.15)` → 1.075，clamp 0.70/1.10 | |
| `EffectivePrice` | `(100, 0.8, 1.0)` → 80.0 | |
| `GroupOf` | `(1)` → GroupGrain；`(999)` → GroupGeneralMarket | |
| `LaborCostPerHour` | `(100, 1.0, 1)` → 100；`(100, 1.2, 3)` → 360 | |
| `EnergyCostPerHour` | 同上模式 | |
| `InputCost` | `(100, 2, 10, 1.0)` → 2000 | |
| `MaintenanceCostPerHour` | `(50, 1)` → 50；`(50, 5)` → 250 | |
| `ManagementCostPerHour` | `(100, 7, 7)` → 700；`(100, 10, 7)` → 1428.57 | quadratic 分支 |
| `TaxCost` | `(1000, 0.1)` → 100；`(1000, 0)` → 0 | |
| `UpgradeCost` | `(5000, 3)` → 15000 | |
| `TotalBuildingCost` | `(5000, 3)` → 30000 | |
| `TickStep` | 每个区间边界值 + 中点：20001→500，9999→100，10→1，1.5→0.01，0.4→0.001 | |
| `IsValidTick` | `(10.0)` → true；`(10.05)` → false (step=1) | |
| `ExchangeFee` | `(100, 10, 0.01)` → 10 (ceil) | |
| `AdminOverheadWithCOO` | `(1.5, 50)` → 1.25 | |
| `CTOProductionMultiplier` | `(10)` → 1.2 | |
| `UnitsSoldPerHour` | 正常市场/高饱和度/零饱和度/底价/NaN 保护 | 各场景预期范围 |
| `clamp`（内部） | `(0.5, 0, 1)` → 0.5；`(-1, 0, 1)` → 0；`(2, 0, 1)` → 1 | |

**实现方式：** 每个测试用 `t.Run` 子测试 + 硬编码 input/output 对，不依赖外部数据。对浮点结果使用 `math.Abs(got-want) < 1e-9` 或 `cmpopts.EquateApprox`。

---

### 2.2 Handler 契约测试

**目标：** HTTP 级别验证：状态码 + 响应字段形状，确保 handler 不因内部重构改变对外契约。

**现有文件：** `backend/internal/handler/handler_test.go`（目前 6 个测试，使用 `httptest`）

**覆盖缺口：**

| 端点 | 当前覆盖 | 需增计划 |
|---|---|---|
| `GET /healthz` | 有 | 保持 |
| `POST /csrf` | 有 | 保持 |
| `GET /company/profile` | 有 | 增加认证 header 缺失 → 401 |
| `GET /market/depth` | 有 | 增加 resourceID 范围外 → 400 |
| `GET /level` | 有 | 保持 |
| `POST /building/buy` | 有 | 增加缺少参数 → 400；余额不足 → 400/403 |
| `POST /building/place` | 无 | 新建测试：非法坐标 → 400 |
| `POST /production/start` | 无（仅 service 层有） | 新建：building 不存在 → 404；资源不足 → 400 |
| `POST /production/claim` | 无 | 新建：idle building→400；已领取→400 |
| `POST /order/create` | 无 | 新建：无效 kind→400；库存不足→400 |
| `POST /order/cancel` | 无 | 新建：不存在的订单→404 |
| `POST /bond/issue` | 无 | 新建：非法面额/利率→400 |
| `POST /bond/buy` | 无 | 新建：额度不足→400 |
| `GET /finance` | 无 | 新建：返回数据结构含 `ledger`、`statements` |
| `POST /player/register` | 无 | 新建：重名→409；空 username→400 |

**模式：**

```go
func TestXxxEndpoint(t *testing.T) {
    h := newTestHandler()
    body := jsonEncode(map[string]any{...})
    req := httptest.NewRequest(http.MethodPost, "/path", body)
    req.Header.Set("X-Auth", "...")
    w := httptest.NewRecorder()
    h.ServeHTTP(w, req)
    assert.Equal(t, 200, w.Code)
    var resp map[string]any
    json.NewDecoder(w.Body).Decode(&resp)
    assert.Contains(t, resp, "expectedField")
}
```

---

### 2.3 市场订单撮合测试

**目标：** 覆盖订单簿撮合引擎的全部核心场景。

**来源：** `backend/internal/service/market_test.go`、`backend/internal/service/order_test.go`

**场景表格：**

| 场景 | 测试方法 | 断言 |
|---|---|---|
| 限价买单与卖单完全匹配 | 创建 sell 10@10，再创建 buy 10@10 | 双方 remaining=0，生成 Trade |
| 多价格优先匹配 | 卖单 12/10/9，买单 15 扫盘 | 最低价 9 先成交 |
| 部分填单 | 卖 5@10，买 10@10 | 卖单 remaining=0，买单 remaining=5 |
| 不匹配（价格区间错开） | 卖 10@15，买 5@10 | 双方 remaining 不变，无 Trade |
| 市价单 | 创建 kind=1 MarketBuy（价格不设或极大值），match 所有卖单 | 按价格顺序吃光 |
| 手续费计算 | 创建 order 后检查公司 Money 扣减 = amount*price*(1+ExchangeFeePct) | |
| Tick 验证 | 提交非法价格（如 10.03，tick=1 不是整数倍） | 返回错误 |
| 撤销后重新匹配 | cancel 部分未成交订单，创建新反方向订单 | 新订单与剩余未成交匹配 |
| 多资源隔离 | 同公司对 resource 3 的订单不影响 resource 1 的订单 | |
| 质量订单匹配 | 创建 sell quality=1，buy quality=1 | 准确匹配 quality |

---

### 2.4 库存操作测试

**目标：** `inventoryAdd` / `inventorySub` / `inventoryGet` 正确性 + 并发安全。

**来源：** `backend/internal/service/service_test.go`  已有 `TestInventoryQuality`，但不够完整。

**新增计划：**

| 测试 | 细节 |
|---|---|
| `TestInventoryAdd_Accumulate` | 添加同资源多次，总量累加 |
| `TestInventorySub_Exact` | 扣减后归零 |
| `TestInventorySub_Insufficient` | 扣减超过持有量 → false，余额不变 |
| `TestInventorySub_Quality` | quality 维度扣减 |
| `TestInventoryGet_ZeroDefault` | 未初始化的资源返回 0 |
| `TestInventoryConcurrency` | `sync.WaitGroup` + 20 个 goroutine 并发 add/sub，最后总量 = 净差额 |
| `TestInventoryConsistencyAfterPanic` | 确保 `inventorySub` 内部不会在中间状态 panic 后留下脏数据（使用 defer recover 验证） |

---

### 2.5 生产队列测试

**目标：** 生产作业生命周期：启动 / 领取 / 取消 / 定时刷新。

**来源：** `backend/internal/service/production_test.go`  已有 15 个测试用例。

**新增计划：**

| 测试 | 细节 |
|---|---|
| `TestStartProduction_DeductsInput` | 启动后输入资源被扣减 |
| `TestStartProduction_NoInputInsufficient` | 输入不足 → 错误 |
| `TestClaimProduction_CreditsOutput` | 成功领取后 output 资源加到库存 |
| `TestClaimProduction_AddsXP` | 领取后 XP 增加 |
| `TestCancelProduction_RefundsInput` | 取消后输入资源退回 |
| `TestRefreshProductionJobs` | 模拟 tick 调用 `RunAllProductionJobs`，已过期 job 应变为 `ready`、输入已足的新 job 自动 claim |
| `TestProductionQueue_OrderPreserved` | 同一建筑多个 job 按 FIFO 完成 |
| `TestProduction48HourRejected` | 已有 `TestStartBuildingProductionRejectsOver48Hours`，确认覆盖 |
| `TestConcurrentProductionClaim` | 同一建筑并发 claim 不 double-credit |

---

### 2.6 财务总账测试

**目标：** 账本条目只追加、方向正确、金额不为负。

**来源：** `backend/internal/model/types.go` 中 `LedgerEntry` 结构；无专用 ledger 测试。

**新增 plan 测试文件：** `backend/internal/service/ledger_test.go`

| 测试 | 细节 |
|---|---|
| `TestLedgerAppendOnly` | 多次调用记账函数后，entries 长度递增，已有条目不被修改 |
| `TestLedgerDirection_Credit` | 收入交易 direction = "in" |
| `TestLedgerDirection_Debit` | 支出交易 direction = "out" |
| `TestLedgerKind_NotEmpty` | 每条 entry 有非空 kind 字段 |
| `TestLedgerAmount_Positive` | 所有 amount > 0 |
| `TestLedgerTimestamp_Exists` | 每条 at 非空且为有效时间格式 |
| `TestLedgerMeta_Optional` | 部分 entry 有 meta，部分无，均合法 |
| `TestLedgerEntriesReflectOperations` | 做买/卖/生产/税操作后，对应公司 ledger 包含对应 kind 的 entry |

---

### 2.7 债券结算测试

**目标：** 利息计算、发行人违约、次级市场交易。

**来源：** `backend/internal/service/bond_test.go`  已有 13 个测试用例。

**新增计划：**

| 测试 | 细节 |
|---|---|
| `TestBondInterest_Accrual` | 多次 `SettleAllBonds` 后利息逐日累加 |
| `TestBondInterest_AfterPartialCall` | CallBond 部分后，剩余本金继续计息 |
| `TestIssuerDefault_MissedPayments` | 连续拖欠超过阈值 → 自动违约 |
| `TestIssuerDefault_FundsInsufficient` | 发行人现金不足付利息 → missed_payments+1 |
| `TestIssuerDefault_NoMoneyNoInterest` | 发行人现金=0，利息记为 0，missed++（已有 `TestSettleBondInterestDefaultsWhenInsolvent`，确认覆盖） |
| `TestBondRestructure` | 违约后重组，RestructurePct 生效 |
| `TestBondSecondaryTrade` | A 将债券卖给 B，owner 转移 |
| `TestBondCallableWindow` | 在 callableAfter 之前不能 call |
| `TestBondRating_UpgradeDowngrade` | 发行人 level/财务变化 → 评级调整 |

---

### 2.8 调度器 Tick 测试

**目标：** 调度器 `tick()` 内每个操作可独立单元测试。

**来源：** `backend/internal/scheduler/scheduler.go`

**设计：** 不测试 `Scheduler` 的定时循环本身（那需要 mock clock），而是将 `tick()` 分解为可测试的步骤函数。通过 `GameService` 接口的 mock 验证每个步骤被调用。

| 测试 | 模拟 | 断言 |
|---|---|---|
| `TestTick_BondSettlement` | 设置 `SettleAllBonds` 返回含 defaults | defaults 被 log，逻辑继续 |
| `TestTick_GovContractAward` | 验证 `AwardGovernmentContracts` 被调用 | 至少调用一次 |
| `TestTick_GovernmentDefault` | 验证 `ResolveGovernmentDefaults` 被调用 | 至少调用一次 |
| `TestTick_BotMarketCycle` | 验证 `RunBotMarketCycle` 被调用 | |
| `TestTick_MarketLockCheck` | `ResourcesWithMarket` 返回 [1,2]，检查 `CheckMarketLock` 被每个 resource 调用 | |
| `TestTick_OrderCleanup` | 验证 `CleanupOrders` 被调用 | |
| `TestTick_ProductionRefresh` | 验证 `RunAllProductionJobs` 被调用 | |
| `TestTick_DailyOrderRefresh` | 验证 `RefreshDailyOrders` 被调用 | |
| `TestTick_SaveAll` | 验证 `SaveAll` 被调用 | |

**Mock 实现方案（推荐）：** 在 `scheduler_test.go` 中定义 `mockGameService` 结构体实现 `GameService` 接口，每个方法记录调用次数/参数。

---

### 2.9 并发安全冒烟测试

**已有：** `TestConcurrentStateAccessSmoke` 在 `service_test.go`。

**补充：**

| 测试 | 细节 |
|---|---|
| `TestConcurrentCreateOrder` | 20 goroutine 同时创建订单，无 data race |
| `TestConcurrentProductionClaim` | 多个 goroutine 竞争同一 building claim |
| `TestConcurrentSnapshotAndMutate` | `Snapshot()` 读取与 mutate 操作并发 |
| `TestConcurrentBondIssueAndSettle` | Issue 和 Settle 并发执行 |

所有并发测试需加 `-race` flag 运行验证。

---

## 3. 基线捕获命令

以下命令应在本文档编写时（Phase 1 开始前）运行，输出存档作为对比基线。

```bash
# 进入 backend 目录
cd backend

# 全部测试（含缓存跳过），收集总耗时
go test ./... -count=1 -v 2>&1 | tee ../docs/2026-06-06/rebuild/test-baseline-output.txt

# 测试二进制大小（用于后续对比检测膨胀）
go test -c -o /dev/null ./internal/service 2>&1

# Vet 检查
go vet ./... 2>&1 | tee ../docs/2026-06-06/rebuild/vet-baseline-output.txt

# 竞态检测（不期望通过，但记录当前失败数）
go test ./... -race -count=1 2>&1 | tee ../docs/2026-06-06/rebuild/race-baseline-output.txt

# 测试覆盖率
go test ./... -coverprofile=../docs/2026-06-06/rebuild/coverage-baseline.out
go tool cover -func=../docs/2026-06-06/rebuild/coverage-baseline.out
```

> 注：当前 `go test ./...` 和 `go vet ./...` 均通过。`-race` 因本环境（Windows 无 CGO）无法执行，基线标记为 N/A，需在 Linux/CI 环境补充。Phase 1 不强制修复 race 问题。

---

## 4. 测试质量基线

当前统计（手动测量基准）：

| 指标 | 值 |
|---|---|
| 测试文件总数 | ~20 |
| 测试用例总数（估算） | ~110 |
| 各包通过率 | 100% |
| vet 告警 | 0 |
| `-race` 通过 | 待记录 |
| 测试执行耗时 | ~7 秒（全量） |
| 覆盖缺口优先级 | handler > ledger > scheduler > formula golden |
