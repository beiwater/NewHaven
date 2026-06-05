# 风险登记册 — NewHaven Phase 1 后端重构

> 文档日期: 2026-06-05
> 版本: v1.0
> 状态: 初稿

## 风险概览

| 严重程度 | 数量 |
|----------|------|
| Critical | 6    |
| High     | 12   |
| Med      | 6    |
| Low      | 1    |

## 风险明细

| ID | 类别 | 风险描述 | 严重程度 | 概率 | 证据 (文件:行号) | 影响 | 缓解措施 | 所需测试 | 责任人 |
|----|------|----------|----------|------|------------------|------|----------|----------|--------|
| MARKET_MATCH-001 | 市场撮合 | 订单撮合中价格/数量计算的整数与浮点数精度问题导致舍入误差 | Critical | High | `market_trade.go:35` `float64(quantity) * price` 乘积后用 `<=` 与 `company.Money` 比较；`market_trade.go:155` `ExchangeFee()` 返回 `math.Ceil` 取整后的 float64；`market_trade.go:169,179,181` 多处 `float64(fill) * execPrice` 浮点乘法；`market_match.go:82` `cost := float64(fill)*sell.Price + fee` 累加浮点值；`formula/market.go:44` `ExchangeFee` 使用 `math.Ceil(float64(amount) * price * feeRate)` 但 amount 是 int 时先转 float64，高单价(>20000)下 float64 精度 ~0.5 无法精确表示整数 | 玩家资金/库存因舍入偏差异步，系统账目不自洽，GameState 持久化后金额偏差累计 | ① 所有金额计算改用 `int64`（分为单位），仅在 UI 展示时转 float ② `ExecuteFee` 返回 `int64` 而非 float64 ③ 在 `executeMatch`/`executeTakeFill` 入口处添加金额守恒断言 | 精确性测试：对已知价格/数量组合的撮合结果进行分单位断言，验证总入=总出+费用 | 后端组 |
| MARKET_MATCH-002 | 市场撮合 | 多个订单同时撮合时部分成交状态损坏 | Critical | High | `market_trade.go:156-163` `executeMatch` 直接修改 `buy.Remaining` 和 `sell.Remaining`，而 `matchLimitOrders`(`market_trade.go:116-130`) 在循环中调用 `executeMatch` 时无重入保护；`market_trade.go:57` `s.State.Orders = append([]model.MarketOrder{order}, s.State.Orders...)` 新订单插入队首，撮合循环中遍历的 slice 引用可能已经改变；`CollectOrders`(`market_trade.go:132-147`) 返回指针切片，多个撮合可能修改同一个指针指向的底层数组 | 订单剩余量被覆盖，同一库存被多次分配或漏分配，导致 GameState 状态不一致 | ① 将 `matchLimitOrders` 重构为在副本上计算撮合，再原子性应用到 State ② 每笔撮合完成后检查 buy/sell 订单的 `Version` 字段（需新增乐观锁）③ 限制单个撮合周期内最大匹配轮次 | 并发撮合测试：N 个 goroutine 同时创建和取订单，验证最终剩余量与 ledger 总和一致 | 后端组 |
| MARKET_MATCH-003 | 市场撮合 | 手续费计算不一致（ceil vs floor vs round） | High | Med | `formula/market.go:44` `ExchangeFee = math.Ceil(amount * price * feeRate)`；`market_trade.go:155` `executeMatch` 调用 `ExchangeFee` 并从卖方收款中扣除；`market_match.go:81-82` `executeTakeFill` 同样调用 `ExchangeFee` 但加到买方成本中；`bond.go:207` `formula.DailyBondInterest` 使用 `math.Floor`；`formula/bonds.go:22` 利息用 Floor 而市场费用用 Ceil，策略不统一 | 分属不同路径（限价单 vs 市价单）的手续费计算策略一致但语义不同，结算差异虽小但在高频交易场景下会积累偏差 | ① 统一费用取整策略（推荐 Ceil 作为系统收入取整，避免浮点向下舍入导致系统亏钱）② `ExchangeFee` 改为 `int64` 参数和返回值，消除浮点精度影响 | 费用一致性测试：对同一笔模拟交易分别走限价单和市价单路径，验证费用计算一致 | 后端组 |
| PRODUCTION-001 | 生产系统 | 定时刷新 `refreshProductionJobs` 调用频率不足导致时间状态漂移 | High | Med | `production.go:291-332` `refreshProductionJobs` 仅在 `ClaimProduction`(`production_claim.go:23`) 和显式 `RefreshProductionJobs` 时对单个公司调用；`scheduler/scheduler.go:86` `RunAllProductionJobs()` 每 60s 调度一次全量刷新，但两轮 tick 间的生产进度只能靠玩家主动 claim 来推进 | 玩家在线时长不均导致生产完成时间感知不一致；高频在线的玩家能比离线玩家更及时地 claim，获得不公平优势 | ① 将生产进度推进改为纯计算模式：`claimableAmountForJob` 每次传 `now` 参数实时计算(e.g. `production.go:334-354` 已部分实现)，不依赖定时刷新更新状态 ② scheduler tick 中增加部分完成状态的通知推送 | 时间漂移测试：冻结模拟时间后创建 job，设置 CompletesAt 过去 1 小时，验证 claimableAmount 直接返回全额 | 后端组 |
| PRODUCTION-002 | 生产系统 | 生产任务声明后槽位未释放（slot leak） | High | High | `production.go:321-326` 刷新时按 `Status == "running" || Status == "ready"` 统计占用槽位；`production_claim.go:46` claim 后将 `Status` 设为 `"claimed"`，因此 claim 后不再计入占用——但 `production.go:327-331` 设置 `busy` 字段仅用于 UI 展示，实际 `checkBuildingSlot`(`production.go:103-113`) 按 job 列表中的 running/ready 数量判断 | 若因 bug 导致已 claim 的 job 状态未更新，槽位永久泄漏，玩家无法启动新生产；极端情况所有槽位被"幽灵"任务占用 | ① `checkBuildingSlot` 增加 `Status == "claimed"` 的检查，强制跳过已声明任务 ② 新增定时清理：CompletesAt 超过 7 天的 running 任务自动标记为 claimed ③ 在每个 `ClaimProduction` 后断言该 building 的 running 任务数 <= slots | 槽位泄漏测试：创建→claim→验证新建任务成功；手动将 Status 置为 running 并设置 CompletesAt 已过时，验证清理逻辑 | 后端组 |
| PRODUCTION-003 | 生产系统 | 原料已扣除但任务创建失败（部分执行） | Critical | Med | `production.go:67-69` `deductInputs` 执行成功后如果后续步骤（如 `production.go:80` append 到 State）panic 或被中断，原料已扣除但任务未创建；`deductInputs`(`production.go:244-253`) 先做存在性检查再逐个扣除，但扣除过程非原子（多轮 `inventorySub`） | 玩家永久损失原料，无对应生产任务，且无回滚机制 | ① 使用延迟的 inventory 快照+回滚模式：先保存库存快照，任务创建成功后释放快照，失败则回扣 ② 将该代码块改为先构建完整 job 对象，最后原子性 append 到 State 并保存 | 部分执行恢复测试：注入 `StartBuildingProduction` 中途 panic，验证库存被正确归还 | 后端组 |
| INVENTORY-001 | 库存系统 | 并发 add/sub 操作导致库存为负或物品丢失 | Critical | High | `service.go:478-505` `inventoryAdd` 和 `inventorySub` 无锁（调用者应已持有 `s.mu`），但 `market_trade.go:49` 在锁内调用 `inventorySub`，而 `market_match.go:88` `executeTakeFill` 也在锁内调用 `inventoryAdd`；`executeMatch`(`market_trade.go:168,174`) 同时调用 `inventoryAdd` 和 `Money` 增减；`cancelOrder`(`market_trade.go:105`) 调用 `inventoryAdd`——这些操作依赖调用者正确持有锁，一旦锁遗漏则并发写导致 data race | 库存数值严重不一致：物品凭空消失、负库存、重复发放 | ① 在 `inventoryAdd`/`inventorySub` 中添加检测当前 goroutine 是否持有锁的断言（debug build 开启）② 所有库存修改路径代码审查，确保锁覆盖率 ③ 考虑将库存改为无锁 atomic 操作或单一写者模式 | 并发库存测试：用 `-race` flag 启动，20 个 goroutine 并发创建/取消订单 100 轮，zero data race | 后端组 |
| INVENTORY-002 | 库存系统 | 库存变更未同步写入账本 | High | High | `inventoryAdd`/`inventorySub`(`service.go:478-505`) 是纯赋值操作，不调用 `addLedger`；`market_trade.go:47-50` `CreateOrder` 中 `company.Money -= total` 和 `inventorySub` 后只有特定 kind 调用了 `addLedger`；`executeMatch`(`market_trade.go:188-189`) 的 ledger 记录的是 trade/fee 而非库存变化本身；`cancelOrder`(`market_trade.go:101-106`) 中退款记了 ledger(`market_buy_refund`) 但退货的 `inventoryAdd` 无对应 ledger | 审计溯源缺口：无法通过账本重建库存变动历史，违反金融系统可审计性要求 | ① `inventoryAdd`/`inventorySub` 签名增加 reason 参数，内部强制调用 `addLedger` ② 或新增 `InventoryChange` 独立的事件存储，与库存修改在同一个事务中写入 | 审计完整性测试：对每条库存变更路径，检查 ledger 中存在对应的 entry 记录 | 后端组 |
| LEDGER-001 | 账本系统 | 账本条目被修改或删除，违背追加写不可变性 | High | High | `model/types.go:43-50` `LedgerEntry` 无不可变标识（无 append-only 校验）；`service_save.go:26` `s.State.Ledger = append([]model.LedgerEntry{entry}, s.State.Ledger...)` 追加到 slice 头部，但 `service_save.go:28` `s.State.Ledger = s.State.Ledger[:s.Cfg.Game.MaxLedgerEntries]` 有截断操作——`cleanup.go:34-36` 中 `CleanupOrders` 也截断 ledger；任何代码路径都可能通过 `s.State.Ledger[i] = ...` 修改已有条目 | 账本记录可被篡改或删除，审计失效，无法作为经济系统争议的客观依据 | ① LedgerEntry 增加 `hash` 字段：每个条目包含前一条目哈希链 ② 截断操作只允许从旧端（尾部）删除，不允许随机删除或修改 ③ 引入只读的 `Ledger()` API 禁止外部直接修改、新增 `LedgerWriter` 接口 | 不可变性测试：尝试通过 State 引用修改 ledger 条目后重新读取，验证哈希链校验失败 | 后端组 |
| LEDGER-002 | 账本系统 | 并发条目导致 `balance_after` 计算错误 | Med | Med | `model/types.go:43-50` `LedgerEntry` 无 `balance_after` 字段；`addLedger`(`service_save.go:18-30`) 不计算余额；现有账本仅记录单笔金额和方向，不记录累计余额 | 无法快速验证账本自洽性（sum(收入) - sum(支出) != 当前余额），并发创建条目时查询的余额快照可能过时 | ① 新增 `balance_after` 字段，在 `addLedger` 中原子性计算（需要读当前余额）② 或提供离线校验工具定期检查 ledger sum 是否等于 GameState 中的 money 总和 | 账本平衡测试：随机注入 1000 笔交易后执行校验脚本，验证 sum(ledger) == sum(money across all companies) | 后端组 |
| BOND-001 | 债券系统 | 大量债券利息叠加时浮点精度聚合误差 | Med | Med | `formula/bonds.go:22` `DailyBondInterest = math.Floor(amount * BondFaceValue * interestRatePct / 100.0)` 每笔利息立即 floor；`bond.go:207-212` 每 tick 对所有债券遍历计算，将 floor 后的值累加到 `totalIncome`/`totalExpense`（float64 累加）；高利率小额多次场景下单笔 floor 损失虽小，但全服数千笔债券每 tick 聚合后的系统负债偏差持续扩大 | 系统发行债券总额高时，利息聚合偏差导致经济系统通胀/通缩压力不可控 | ① 利息累加改为 `int64`（分单位），所有金额计算在整数空间进行 ② 保留 `math.Floor` 策略但确保最后一次利息结算时补齐 | 利息精度测试：创建 10000 笔 1 单位债券，计算总利息并与高精度十进制预期值比较，偏差 < 1 分 | 后端组 |
| BOND-002 | 债券系统 | 发行方公司破产后已发行债券未处理 | High | Med | `bond.go:193-248` `SettleBondInterest` 仅在资金不足时记录 `MissedPayments` 和调整 `RestructurePct`，不进行债券强制平仓；`ResolveGovernmentDefaults`(`government.go:149-171`) 处理政府合同违约但不处理债券违约；`model/types.go:9-21` `Bond` 结构体无 `status` 字段标记已违约 | 违约债券持续悬挂在市场上，其他玩家仍可买入已违约的无价值债券；公司破产后债券持有人无法追索 | ① 新增 `status` 字段 (`active`/`defaulted`/`restructured`) ② `SettleBondInterest` 中连续 MissedPayments >= 3 时触发强制平仓和销毁 ③ 禁止购买 `MissedPayments > 0` 的债券 | 违约处理测试：制造一家公司连续违约 3 次，验证债券自动平仓且不可购买 | 后端组 |
| GOV-001 | 政府合同 | 同一合同因竞态条件授予多个投标方 | Critical | Low | `government.go:67-107` `AwardGovernmentContracts` 由 scheduler tick 调用（`scheduler.go:70`），单线程串行执行；`PlaceGovernmentBid`(`government.go:11-65`) 在用户请求路径上，两者都持有 `s.mu`——只要锁正确，不会并发授予 | 理论上竞态条件概率低，但一旦发生，同份合同多次发放报酬，经济系统通胀 | ① 在 `AwardGovernmentContracts` 中增加已授予校验：`c.Status` 从 `"open"` 改为 `"awarded"` 使用 CAS 语义 ② 输出 `out` 中去重 ③ 添加 contract 的 version 乐观锁 | 合同授予幂等性测试：手动将合同状态设为 `"open"` 但不设 Bids，验证 `AwardGovernmentContracts` 跳过 | 后端组 |
| SCHEDULER-001 | 调度器 | Tick 操作顺序依赖构成隐含业务逻辑：债券结算必须在机器人周期之前 | High | Med | `scheduler.go:60-95` tick 顺序固定：①结算债券 ②授予合同 ③违约处理 ④机器人周期 ⑤锁检查 ⑥清理 ⑦生产刷新 ⑧日订单刷新 ⑨持久化；服务没有显式定义步骤间的数据依赖契约 | 若某步骤抛出 panic 导致后续步骤全部跳过，经济状态停留在不一致的中途状态；重构时若调整顺序（如债券结算移到机器人周期之后）会导致债券利息计算使用错误的价格指数 | ① 将每个步骤封装为独立函数，输出明确的依赖数据契约 ② 引入 `TickPhase` 枚举，每 phase 完成后检查不变量 ③ 每个 tick 包裹 recover，panic 后至少保证 SaveAll 执行 | Tick 顺序测试：冻结时间后执行完整 tick 序列，验证每一步的读/写集与下一步的输入集相符 | 后端组 |
| SCHEDULER-002 | 调度器 | 机器人做市商行为创造套利机会 | High | High | `market_competition.go:166-232` `RunBotMarketCycle` 每小时创建 buy/sell 订单，使用 `math.Round(buyBase*cycleVol*(1-spread)*100)/100` 计算价格；`market_competition.go:205-206` 买卖数量根据库存偏差调整，但价格计算 (`market_competition.go:209,215`) 直接用 `basePrice` 和 `buyBase` 两个不同价格源；`ComputeChainPrice`(`market_competition.go:197-198`) 的输出依赖链式公式 | 玩家可通过观察机器人订单价格模式（每小时更新、可预测的正弦波动）进行套利交易，从机器人手中无风险获利 | ① 增加机器人价格噪声 ±3% 随机偏移 ② 缩短机器人调价周期到 15-30 分钟 ③ 限制玩家与机器人订单的成交比例 | 套利空间测试：运行模拟 24 小时，统计玩家-机器人交易中玩家平均盈利率，超过 10% 则告警 | 后端组 |
| SCHEDULER-003 | 调度器 | `SaveAll` 与 API handler 修改竞速 | Critical | High | `cleanup.go:46-57` `SaveAll` 获取 `s.mu` 后直接调用 `s.Store.SaveState(ctx, &s.State)` 持久化完整 State，期间持有锁——但 `SaveAll` 从 scheduler goroutine 调用（`scheduler.go:92`），而所有 API handler 也通过 `s.mu` 同步，两方共享同一把锁 | 虽因同一把锁不会导致数据竞争，但 `SaveState` 是 I/O 操作（PostgreSQL upsert），锁持有时间长达毫秒级，60 秒 tick 中的 SaveAll 阻塞所有 API 请求，造成玩家可见的卡顿 | ① `SaveAll` 改为 State 的快照（`Snapshot()` + 深度拷贝）后在无锁状态下存储 ② 使用写时复制（Copy-on-Write）机制 ③ 增加 SaveAll 的超时上限，超时后当前 tick 跳过保存 | 锁竞争测试：模拟 100 QPS API 请求 + 每 tick SaveAll，检测 p99 响应延迟，不得超过 500ms | 后端组 |
| FORMULA-001 | 经济公式 | 公式修改静默改变游戏经济 | High | Med | `formula/market.go` `TickStep` `ExchangeFee` `IsValidTick` 均为纯函数，可被独立单元测试；`formula/production.go` `OutputPerHour` `ProductionDurationSeconds` 依赖外部传入的 `baseOutputPerHour` 和 `secondsPerUnit`；`formula/costs.go` 所有公式无版本标识；`StaticData`(`data/loader.go:10-15`) 中 `map[string]any` 类型的 `EconomyModel`、`Buildings` 等字段通过 `intFromAny`/`floatFromAny` 提取值——实际生产环境中若 JSON 被修改（如 `producedPerHourRaw` 从 `"1.5"` 变为 `"1.2"`），游戏进入静默不通告的新版本 | 经济参数变更无版本控制，线上热更新 JSON 后现有玩家经济平衡被破坏，且无法回滚 | ① 所有公式参数和 JSON 字段增加版本号（`formula_version` / `data_version`）并在 game.json 中记录 ② `StaticData` 加载时校验 hash ③ 公式变更时通过 game.json 的 breaking_change 标记拒绝加载旧存档 | 公式版本测试：加载旧版本 JSON 应触发明确的错误或迁移路径，而非静默接受 | 后端组 |
| STATIC_DATA-001 | 静态数据 | `map[string]any` 类型断言在字段缺失时 panic | High | Med | `data/loader.go:10-15` `Resources []map[string]any`, `Buildings map[string]any`；`service.go:423-435` `intFromAny(v any) int` 和 `service.go:454-464` `floatFromAny(v any) float64` 使用 `switch t := v.(type)` 处理 json unmarshal 后的类型映射，但 `intFromAny` 未处理 `nil` 或 `string` 类型（JSON 中数值可能被 unmarshal 为 `float64` 或 `json.Number`）；`production.go:280-288` `outputPerHour` 中 `intFromAny(r["dbLetter"])` 和 `floatFromAny(r["producedPerHourRaw"])` 假设字段存在，若 JSON 字段缺失则传入 `nil` | 字段缺失或类型异常时生产环境 panic，service 不可用 | ① `intFromAny`/`floatFromAny` 增加 nil 检查和类型兜底，不 panic 而返回 (0, error) ② 加载时使用 JSON schema 验证或转为强类型 struct ③ 在 `Load()` 之后执行验证回调查看所有必要字段 | 静态数据加载测试：提供缺失字段的 JSON 文件，验证加载不 panic 且返回明确错误 | 后端组 |
| API-001 | API 兼容性 | 破坏性 API 变更未被检测（字段重命名、类型变更） | High | Med | `handler/handler.go:62-81` `Register` 注册的 handler 返回 `map[string]any`，无 OpenAPI 规范或结构化响应类型；`handler/production.go`、`handler/market.go` 等 handler 通过 `writeJSON(w, 200, map[string]any{...})` 返回无 schema 约束的 JSON，字段名在 handler 层与 service 层重复映射 | 前端编译时无法检测到字段重命名，运行时因 JSON key 不匹配而解析失败 | ① 为所有 API 响应定义 Go struct（至少关键 path），handler 返回强类型 ② 引入 API 契约测试：快照当前响应 JSON 结构，CI 中 diff 检测字段变化 ③ 在 handler 层统一处理 snake_case 转换 | API 兼容性测试：每个 handler 的响应 JSON 与已知 schema 对比，新增字段必须显式标记 | 后端组 |
| API-002 | API 兼容性 | 旧 API 路由在前端迁移完成前被移除 | Med | Med | `handler/handler.go:62-81` 路由在 `Register` 中集中注册，每条路由直接对应 handler 方法；前端代码未与后端路由定义共享——无路由注册表或前端对应的 API 客户端自动生成 | 前端部署时序不当导致线上部分功能不可用 | ① 路由标记 deprecation 日期，保留至少 2 个版本周期 ② 使用 API 版本前缀 `/v1/`、`/v2/`并行提供服务 ③ 引入 `Deprecation` 响应头 | 路由兼容性测试：对标记为 deprecated 的路由断言返回 `Deprecation` header 而非 404 | 后端组 |
| DB_MIGRATION-001 | 数据迁移 | JSONB 快照转范式化存储时数据丢失 | Critical | Med | `storage/postgres.go:229-323` `LoadState` 先尝试 `game_state` 快照表，若 companies 为空则降级到从独立 domain 表组装——独立表与快照表之间的迁移缺少字段映射验证；`postgres.go:42-51` `migrate` 仅有基本的表结构创建，无渐进式 migration 版本链；存储层使用 `upsert`(`postgres.go:53-63`) 以 JSONB 写入 `data` 列，无针对嵌套字段的校验 | 迁移过程中若一方新增字段而另一方无对应列，数据静默丢失；回滚后完整状态不可恢复 | ① 每次 save 时同时写入 game_state 快照和独立 domain 实体，确保冗余 ② 迁移脚本必须经过 staging 环境的校验测试 ③ 引入 migration 版本号，拒绝不兼容的降级 | 迁移回滚测试：在 Postgres 中存储 v1 格式的 snapshot，应用 v2 迁移后回滚，验证 v1 数据完整可恢复 | 后端组 |
| WS-001 | 网络通信 | WebSocket 事件顺序与 REST API 状态不一致 | Med | High | `scheduler/scheduler.go:60-95` tick 中无 WebSocket 事件广播；当前代码库无 WebSocket 实现（搜索未发现 `websocket`、`ws`、`upgrade`、`gorilla/websocket` 引用）——若 Phase 1 仅 REST API 则此风险延期 | WebSocket 事件基于 REST 响应间隔推送，玩家看到的状态更新滞后于实际游戏 tick，导致 UI 闪烁或数据回滚 | ① 若 Phase 1 不接入 WebSocket，将该风险标记为已接受并延期 ② 若接入 WebSocket，每个 Scheduler tick 完成后向所有连接广播状态 diff ③ 在客户端实现乐观更新 + tick 覆写 | 事件顺序测试：模拟 tick 结算后查询 REST API，验证返回状态与广播的事件最终一致 | 后端组 |
| ANTI_CHEAT-001 | 反作弊 | 重构中绕过或删除了反作弊限速 | High | High | `anticheat/anticheat.go:83-101` `CheckRateLimit` 在每个业务入口调用（`CreateOrder:23-24`, `TakeOrder:22-23`, `ClaimProduction:18-19`, `PlaceGovernmentBid:19-20`, `IssueOrAdjustBond:36-37`），但 `BuyBond`(`bond.go:74`) 和 `CallBond`(`bond.go:156`) 和 `DeliverGovernmentContract`(`government.go:109`) 缺少对应的 `CheckRateLimit` 调用——重构时若从这些路径复制代码，限速检查可能被遗漏 | 机器人攻击者可通过未限速的接口无限制执行经济操作，破坏游戏平衡 | ① 创建通用的 `withRateLimit(func) (map[string]any, error)` 包装装饰器 ② 审查所有 service method 调用入口，补充缺失的 `CheckRateLimit` ③ 新增一个全局中间件层自动对所有 API 请求限速 | 限速覆盖测试：枚举所有 service method 入口，检查是否有对应的 rate limit 调用 | 后端组 |
| ANTI_CHEAT-002 | 反作弊 | 脚本检测在 data migration 过程中不触发 | Med | Low | `anticheat/detector.go:36-64` `RecordAction` 仅在玩家在线操作时触发（`service.go:27` `s.SD.RecordAction(pid)` 在 CreateOrder、TakeOrder、ClaimProduction、PlaceGovernmentBid、IssueOrAdjustBond 等入口调用）；`storage/postgres.go:229-323` `LoadState` 在服务启动时批量加载公司和订单，不经过 `RecordAction` | 若迁移脚本在后台批量执行玩家操作（如重放订单），这些操作不经过脚本检测，可被恶意玩家利用 | ① 迁移脚本中关闭 `SD.enabled` 或标记操作为 `system` 类型 ② 批量操作不走业务 service method，走独立的 admin API | 脚本检测隔离测试：通过批量导入路径注入大量订单，验证 SD 记录未被污染 | 后端组 |
| STORAGE-001 | 存储 | NoopStorage 内存模式重启数据丢失（设计如此，但需文档化） | Low | High | `storage/storage.go:21-28` `NoopStorage` 所有方法为空实现，`LoadState` 返回 `nil, nil`；`cleanup.go:47-48` `SaveAll` 中 `if s.Store == nil { return }`——开发模式下使用 `New` (`storage.go:17-19`) 返回 postgres 而非 NoopStorage；`service/New`(`service.go:69-220`) 接受 storage 参数但依赖调用方决定使用 `New` 或 `NoopStorage` | 开发/演示环境无持久化，服务重启后所有玩家进度丢失 | ① NoopStorage 每次 `SaveState` 时在内存中保留副本 ② 开发模式默认启用 Postgres ③ 文档中明确标注 NoopStorage 的使用场景 | 存储切换测试：从 Postgres 切换到 NoopStorage 再切回，验证 `LoadState` 能按预期工作 | 后端组 |

## 风险矩阵图（按严重程度 x 概率）

```
                    概率
              Low     Med     High
    Critical  GOV-001         MARKET_MATCH-001
                              MARKET_MATCH-002
                              INVENTORY-001
                              PRODUCTION-003
                              SCHEDULER-003
                              DB_MIGRATION-001
    High      ANTI_CHEAT-002  PRODUCTION-001   MARKET_MATCH-003
                             PRODUCTION-002   INVENTORY-002
                             BOND-002         LEDGER-001
                             SCHEDULER-001    SCHEDULER-002
                             FORMULA-001      STATIC_DATA-001
                             ANTI_CHEAT-001   API-001
    Med                      BOND-001         LEDGER-002
                             API-002          WS-001
    Low                                                       STORAGE-001
```

## 前 5 高风险（按风险暴露排序）

1. **MARKET_MATCH-002** — 并发撮合状态损坏 (Critical × High)
2. **INVENTORY-001** — 并发库存竞争 (Critical × High)
3. **PRODUCTION-003** — 部分执行无回滚 (Critical × Med)
4. **SCHEDULER-003** — SaveAll 锁竞争 (Critical × High)
5. **DB_MIGRATION-001** — 迁移数据丢失 (Critical × Med)

## 已接受的风险

以下风险经评估为低优先级或在 Phase 1 范围外，已接受但不关闭跟踪：

- **BOND-001**（债券浮点精度）：影响仅在债券规模巨大时显现，Phase 1 的债券发行预期有限
- **WS-001**（WebSocket 顺序）：Phase 1 不包含 WebSocket 实现，移入 Phase 2 跟踪
- **STORAGE-001**（NoopStorage 数据丢失）：开发环境人工已知缺陷，不影响生产
- **ANTI_CHEAT-002**（脚本检测隔离）：概率低，仅影响 data migration 工具的使用

## 测试计划优先级

基于前 5 高风险，Phase 1 测试应优先覆盖：

1. **并发撮合测试** (`TestConcurrentMatch`)：20 goroutine 同时创建/取订单，验证最终状态一致
2. **原子性测试** (`TestProductionAtomicDeduction`)：注入 panic 验证库存回滚
3. **SaveAll 锁竞争测试** (`TestSaveAllContention`)：模拟高压 API + tick SaveAll，检测 p99 延迟
4. **库存竞争测试** (`TestInventoryRace`)：`-race` flag 下 100 轮并发操作
5. **迁移回滚测试** (`TestMigrationRollback`)：v1 → v2 → v1 往返数据完整性断言
