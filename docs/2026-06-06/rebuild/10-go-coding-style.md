---

**精简版（子 Agent 优先读这个）**: 见 `docs/2026-06-06/basicprompt.md §0`。
**本文件用途**: 详细参考，不是执行清单。子 Agent 遇到风格争议时以本文件为准。

---

# NewHaven Go 编程风格与 AI 代码规则

本文件约束所有 AI 和人类开发者在 NewHaven Go 后端中的代码风格。

目标不是写“炫技 Go”，而是写稳定、清晰、可测试、可迁移的游戏服务器代码。

## 1. 总原则

1. 代码必须简单、明确、可测试。
2. 不允许为了少写几行代码牺牲可读性。
3. 不允许在 handler 中写业务规则。
4. 不允许在 handler 中写公式。
5. 不允许在 handler 中直接操作库存、市场订单、财务流水。
6. 不允许新增 `map[string]any` 作为正式 API 返回值。
7. 不允许 domain model 直接暴露给前端新 API。
8. 不允许业务代码读取 `docs/reference` 或 `docs/backend-refactor/reference`。
9. 不允许一次性重写 market、production、finance、scheduler。
10. 每个 PR 必须能单独运行、单独测试、单独回滚。

## 2. Go 官方基础规则

所有 Go 代码必须满足：

```bash
gofmt ./...
go vet ./...
go test ./...
```

新增代码必须符合：

* Effective Go
* Go Code Review Comments
* Go Doc Comments
* Google Go Style Guide 的基本命名、错误处理、包设计建议

## 3. 包命名规则

包名必须短、小写、表达职责。

推荐：

```txt
market
production
finance
building
research
player
formula
postgres
memory
http
middleware
```

禁止：

```txt
utils
common
helper
manager
core
business
logic
misc
```

如果真的需要工具函数，必须放到明确领域中，例如：

```txt
market/price.go
production/quality.go
finance/money.go
```

不要新建万能工具包。

## 4. 文件命名规则

文件名按领域和职责命名。

推荐：

```txt
order_create.go
order_match.go
order_cancel.go
production_start.go
production_claim.go
ledger_entry.go
bond_settle.go
```

禁止继续扩大：

```txt
service.go
helpers.go
utils.go
common.go
misc.go
```

一个文件超过 300 行必须考虑拆分。
一个函数超过 80 行必须考虑拆分。
一个函数同时做三件以上事情必须拆分。

## 5. 分层规则

后端分层必须遵守：

```txt
handler / adapter/http
    只负责 HTTP、参数、DTO、状态码

app / usecase
    负责编排业务流程

domain
    负责业务规则和领域模型

formula
    负责纯经济公式

storage / repository
    负责持久化

scheduler
    负责定时触发，不直接写复杂业务
```

handler 只能调用 app/service，不能直接调用 storage，不能直接改 GameState。

## 6. Handler 规则

handler 只能做这些事：

1. 读取 path/query/body。
2. 校验基础格式。
3. 调用 usecase。
4. 把结果转换为 response DTO。
5. 写 JSON。
6. 写 HTTP 状态码。

handler 禁止：

1. 计算市场价格。
2. 计算生产时间。
3. 修改库存。
4. 写 ledger。
5. 写订单撮合逻辑。
6. 直接返回 domain model。
7. 返回 `map[string]any`。
8. 手写复杂路径解析。

错误示例：

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // 禁止：handler 中计算手续费、扣钱、改库存、撮合订单
}
```

正确方向：

```go
func (h *MarketHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    req, err := decodeJSON[CreateOrderRequest](r)
    if err != nil {
        writeError(w, err)
        return
    }

    res, err := h.market.CreateOrder(r.Context(), req.ToCommand())
    if err != nil {
        writeError(w, err)
        return
    }

    writeJSON(w, http.StatusCreated, CreateOrderResponseFromDomain(res))
}
```

## 7. DTO 规则

新 API 必须有明确 DTO。

命名：

```txt
CreateMarketOrderRequest
CreateMarketOrderResponse
MarketOrderDTO
MarketOrderListResponse
ErrorResponse
```

禁止：

```go
map[string]any
map[string]interface{}
interface{} // 作为正式 API 响应
```

Domain model 和 API DTO 必须分开。

例如：

```go
type MarketOrder struct {
    ID        int64
    CompanyID int64
    ResourceID int64
    PriceCents int64
}
```

不能直接返回给前端。

应该转换为：

```go
type MarketOrderDTO struct {
    ID       string `json:"id"`
    Resource string `json:"resource"`
    Price    string `json:"price"`
}
```

## 8. 错误处理规则

禁止忽略错误。

禁止：

```go
value, _ := strconv.Atoi(s)
```

必须：

```go
value, err := strconv.Atoi(s)
if err != nil {
    return fmt.Errorf("parse amount: %w", err)
}
```

错误必须带上下文：

```go
return fmt.Errorf("create market order: %w", err)
```

不要重复 log + return。
底层 return error，上层统一 log。

禁止在业务逻辑中随便 `panic`。
只有程序启动配置错误、不可恢复初始化错误可以 panic 或 fatal。

## 9. Context 规则

所有跨边界函数必须接受 `context.Context`。

推荐：

```go
func (s *Service) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*OrderResult, error)
```

禁止：

```go
func (s *Service) CreateOrder(cmd CreateOrderCommand) (*OrderResult, error)
```

`context.Context` 必须是第一个参数。
不允许把 context 存进 struct。
不允许传 `nil` context。

## 10. Interface 规则

interface 应该定义在使用方，而不是实现方。

推荐：

```go
type MarketStore interface {
    CreateOrder(ctx context.Context, order MarketOrder) error
    FindOpenOrders(ctx context.Context, resourceID int64) ([]MarketOrder, error)
}
```

禁止巨大 interface：

```go
type GameService interface {
    CreateOrder(...)
    CancelOrder(...)
    StartProduction(...)
    ClaimProduction(...)
    SettleBond(...)
    RunScheduler(...)
}
```

一个 interface 最好 1-5 个方法。
超过 7 个方法必须重新审查。

## 11. Service / Usecase 规则

一个 usecase 只做一个业务动作。

推荐：

```txt
CreateMarketOrder
CancelMarketOrder
TakeMarketOrder
StartProduction
ClaimProduction
IssueBond
SettleBondInterest
```

禁止一个函数里同时做：

```txt
读取请求 → 校验 → 扣库存 → 计算价格 → 撮合 → 写 ledger → 写通知 → 写响应
```

应该拆成：

```txt
Validate
LoadState
ApplyDomainRule
Persist
EmitEvent
ReturnDTO
```

## 12. Formula 规则

formula 包必须保持纯函数。

允许：

```go
func ExchangeFee(amount int64, priceCents int64, feeRate float64) int64
```

禁止：

```go
func ExchangeFee(s *Service, companyID int64) int64
```

formula 禁止：

1. 访问数据库。
2. 访问 GameState。
3. 访问 HTTP request。
4. 读写全局变量。
5. 产生随机数。
6. 直接读配置文件。

公式变更必须附带测试。
涉及经济数值变化必须说明是否是 intentional behavior change。

## 13. Money / 数值规则

金融、订单、ledger、债券金额优先使用整数最小单位。

推荐：

```go
PriceCents int64
AmountCents int64
```

如果旧系统暂时使用 float，不要在同一个 PR 里强行全改。
新模块优先使用整数金额。
所有 float 公式必须有测试覆盖，避免精度误差悄悄影响经济系统。

## 14. 时间规则

禁止业务代码中到处直接调用：

```go
time.Now()
```

推荐注入 clock：

```go
type Clock interface {
    Now() time.Time
}
```

scheduler、生产队列、债券结算、离线收益必须可测试。
所有时间字段使用 UTC 存储。

## 15. 并发规则

禁止无管理 goroutine。

禁止：

```go
go func() {
    for {
        doWork()
    }
}()
```

必须：

```go
go func(ctx context.Context) {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            doWork(ctx)
        }
    }
}(ctx)
```

所有 goroutine 必须能退出。
scheduler 必须幂等。
市场撮合、库存扣减、ledger 写入必须有事务边界或锁边界。

## 16. 数据库规则

handler 不能写 SQL。
domain 不能写 SQL。
SQL 只能在 repository/storage 层。

推荐：

```txt
internal/adapter/postgres/
internal/storage/
```

数据库 schema 必须通过 migration 管理。
复杂查询优先 sqlc。
禁止把 SQL 字符串散落在 handler 或 service 中。

## 17. 日志规则

使用结构化日志。

推荐：

```go
logger.Info("market order created",
    "order_id", orderID,
    "company_id", companyID,
    "resource_id", resourceID,
)
```

禁止：

```go
log.Printf("order created %v %v %v", orderID, companyID, resourceID)
```

不要记录密码、token、完整 JWT、敏感认证信息。

## 18. 测试规则

新增业务必须有测试。

优先级：

1. formula golden test
2. domain unit test
3. usecase test
4. handler contract test
5. repository integration test

高风险模块必须测试：

* market order matching
* inventory mutation
* production queue
* finance ledger
* bonds
* government contracts
* scheduler
* bot market maker
* static data loader

测试命名：

```go
func TestCreateMarketOrder_InsufficientFunds(t *testing.T)
func TestStartProduction_NotEnoughInputs(t *testing.T)
func TestExchangeFee_RoundsUp(t *testing.T)
```

## 19. AI 修改代码规则

AI 每次修改前必须说明：

1. 要改哪些文件。
2. 不改哪些文件。
3. 是否改变运行行为。
4. 是否改变 API 响应。
5. 是否改变经济公式。
6. 是否需要 migration。
7. 如何测试。
8. 如何回滚。

AI 禁止：

1. 未经确认删除旧 API。
2. 未经确认改公式。
3. 未经确认改经济参数。
4. 未经确认改 JSON 数据结构。
5. 未经确认创建 `backend-next/` 平行重写。
6. 未经确认引入大型框架。
7. 未经确认把 mock 数据当正式数据。
8. 未经确认改数据库 schema。
9. 未经确认全量替换 service 层。
10. 为了让测试通过而降低测试质量。

## 20. PR 规则

每个 PR 只解决一个主题。

推荐 PR 类型：

```txt
PR1 docs baseline
PR2 openapi draft
PR3 chi bridge
PR4 handler contract tests
PR5 database schema draft
PR6 market service split
PR7 production service split
```

禁止一个 PR 同时做：

```txt
换路由 + 改数据库 + 改市场撮合 + 改前端页面 + 改公式
```

每个 PR 必须包含：

* 变更摘要
* 风险说明
* 测试结果
* 回滚方式
* 是否改变 API
* 是否改变经济行为
