# API Contract Plan — Phase 1

> 本文档定义 Phase 1 的 API 契约规范，包含统一响应格式、DTO 命名约定、版本策略、认证方案、错误码编目、首批端点定义及 OpenAPI 规范文件位置。

---

## 1. OpenAPI Tag 定义（首批）

| Tag | 说明 | 对应 handler 文件 |
|-----|------|------------------|
| `Auth` | 登录、注册、Token 获取 | `backend/internal/handler/auth.go` |
| `Companies` | 公司档案、库存、建筑列表、偏好设置 | `backend/internal/handler/company.go` |
| `Buildings` | 建筑详情、升级、生产选项 | `backend/internal/handler/production.go` |
| `Warehouse` | 仓库管理、物品移动 | `backend/internal/handler/company.go` (handleWarehouse) |
| `Production` | 生产任务、领取、队列管理 | `backend/internal/handler/production.go`, `production_queue.go` |
| `Market` | 市场行情、深度、下单、撤单 | `backend/internal/handler/market.go`, `order.go` |
| `Resources` | 资源定义、资源信息查询 | `backend/internal/handler/market.go` |
| `Finance` | 财务报表、现金流、债券 | `backend/internal/handler/financial.go`, `bond.go` |
| `Research` | 研发项目、进度、完成 | `backend/internal/handler/dev.go` |
| `Executives` | 高管搜索、招募、培训、挖角 | `backend/internal/handler/company.go` |
| `Chat` | 消息系统（预留） | `backend/internal/handler/message.go` |
| `Health` | 健康检查、就绪检查 | `backend/internal/handler/health.go` |

---

## 2. 统一响应格式

### 2.1 成功响应

```json
{
  "data": { ... },
  "error": null,
  "meta": {
    "request_id": "req-xxxxxxxxxxxx"
  }
}
```

- `data`：业务负载。单实体为对象，集合为数组。
- `error`：固定 `null`。
- `meta.request_id`：可选，由客户端传入 `X-Request-Id` 头回显，便于跟踪。

### 2.2 错误响应

```json
{
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable description",
    "details": {}
  },
  "meta": {}
}
```

- `data`：固定 `null`。
- `error.code`：字符串枚举值，见第 6 节编目。
- `error.message`：面向开发者的英文描述，前端可根据 locale 做 i18n。
- `error.details`：可选的结构化上下文（如字段校验失败的字段名列表）。
- `meta`：当前阶段固定空对象，预留扩展。

### 2.3 HTTP 状态码策略

| 场景 | HTTP Status | 说明 |
|------|-------------|------|
| 成功 | 200 | 业务正常完成 |
| 创建成功 | 201 | POST 创建资源（极少数场景，当前阶段绝大部分用 200） |
| 请求体格式错误 | 400 | JSON 无法解析 / 缺少必填字段 |
| 认证失败 | 401 | Token 缺失或无效 |
| 授权不足 | 403 | Token 有效但无权执行该操作 |
| 资源不存在 | 404 | 实体未找到（使用对应 `*_NOT_FOUND` 错误码） |
| 请求冲突 | 409 | 业务逻辑拒绝（如余额不足、库存不足） |
| 频率限制 | 429 | 请求过多 |
| 服务器内部错误 | 500 | 未捕获异常，返回 `INTERNAL_ERROR` |

### 2.4 当前状态 → 目标迁移

**当前问题**（全部 handler 文件）：

```go
// 1. 直接返回 map，无统一信封
writeJSON(w, 200, map[string]any{"csrfToken": token})

// 2. 错误使用 http.Error（纯文本）
http.Error(w, "bad request", 400)

// 3. 错误使用 JSON map，但无错误码
writeErr(w, 400, "username and password required")
```

**目标实现方式**（新增 `handler/response.go`）：

```go
// 统一成功响应
type SuccessResponse struct {
    Data  any            `json:"data"`
    Error *ErrorPayload  `json:"error,omitempty"`
    Meta  MetaPayload    `json:"meta"`
}

type ErrorPayload struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

type MetaPayload struct {
    RequestID string `json:"request_id,omitempty"`
}
```

---

## 3. DTO 命名约定

### 3.1 请求 DTO

| 模式 | 示例 | 场景 |
|------|------|------|
| `Create{Xxx}Request` | `CreateOrderRequest` | 创建/提交资源的 POST 请求体 |
| `{Verb}{Noun}Request` | `LoginRequest`, `ClaimRequest` | 动作类请求 |

### 3.2 响应 DTO

| 模式 | 示例 | 场景 |
|------|------|------|
| `{Noun}Response` | `CompanyResponse` | 单个实体响应 |
| `{Noun}ListResponse` | `BuildingListResponse` | 实体集合响应，包含 `items` 字段 |
| `LoginResponse` | `LoginResponse` | 特殊登录响应，含 token |
| `{Verb}Response` | `ClaimResponse` | 动作结果响应 |

### 3.3 DTO 字段规范

- 使用 `snake_case` JSON 字段名（Go: `json:"field_name"`）。
- 所有非空的 `int`/`float64`/`string` 必填字段不带 `omitempty`。
- 可空字段使用指针 `*Type` 或 `omitempty`。
- 时间字段统一使用 RFC3339 字符串。

### 3.4 DTO Go 文件组织

```
handler/
  dto_auth.go        # LoginRequest, LoginResponse, RegisterRequest, RegisterResponse
  dto_company.go     # CompanyResponse, BuildingListResponse, ...
  dto_production.go  # ProductionJobResponse, ClaimRequest, ClaimResponse, ...
  dto_market.go      # CreateOrderRequest, OrderResponse, MarketTickerListResponse, ...
  dto_finance.go     # FinancialStatementResponse, BondMarketListResponse, ...
  dto_research.go    # ResearchResponse (预留)
  dto_executive.go   # ExecutiveResponse (预留)
  dto_response.go    # 统一信封 DTO + writeSuccess/writeError 辅助函数
  error_codes.go     # ErrorCode 常量枚举
```

### 3.5 ErrorCode 枚举

```go
// handler/error_codes.go

type ErrorCode string

const (
    ErrorInsufficientFunds    ErrorCode = "INSUFFICIENT_FUNDS"
    ErrorInsufficientInventory ErrorCode = "INSUFFICIENT_INVENTORY"
    ErrorInvalidPrice         ErrorCode = "INVALID_PRICE"
    ErrorInvalidTick          ErrorCode = "INVALID_TICK"
    ErrorOrderNotFound        ErrorCode = "ORDER_NOT_FOUND"
    ErrorCompanyNotFound      ErrorCode = "COMPANY_NOT_FOUND"
    ErrorBuildingNotFound     ErrorCode = "BUILDING_NOT_FOUND"
    ErrorProductionNotFound   ErrorCode = "PRODUCTION_NOT_FOUND"
    ErrorProductionNotReady   ErrorCode = "PRODUCTION_NOT_READY"
    ErrorBondNotFound         ErrorCode = "BOND_NOT_FOUND"
    ErrorResearchNotFound     ErrorCode = "RESEARCH_NOT_FOUND"
    ErrorExecutiveNotFound    ErrorCode = "EXECUTIVE_NOT_FOUND"
    ErrorContractNotFound     ErrorCode = "CONTRACT_NOT_FOUND"
    ErrorBidNotFound          ErrorCode = "BID_NOT_FOUND"
    ErrorInvalidRequest       ErrorCode = "INVALID_REQUEST"
    ErrorUnauthorized         ErrorCode = "UNAUTHORIZED"
    ErrorForbidden            ErrorCode = "FORBIDDEN"
    ErrorRateLimited          ErrorCode = "RATE_LIMITED"
    ErrorInternal             ErrorCode = "INTERNAL_ERROR"
    ErrorMaintenanceMode      ErrorCode = "MAINTENANCE_MODE"
)
```

---

## 4. API 版本策略

### 4.1 现状

| 前缀 | 用途 | 路由数（近似） | 处理方式 |
|------|------|---------------|----------|
| `/api/` | 早期端点（login, register, csrf, bonds） | ~6 | 无版本号，逐步迁移到 v2 |
| `/api/v1/` | 第一代建筑/生产端点 | ~2 | 已在淘汰，标记 legacy |
| `/api/v2/` | 当前前端主力版本 | ~25 | 保留兼容，新业务优先放 v2 |
| `/api/v3/` | 市场、政府、高管等较新端点 | ~10 | 保留现有，新增适当放 v3 |
| `/api/v4/` | 实验性端点（payment 等） | ~2 | **STOP 添加**，已存在的留用但标记 deprecated |
| `/api/dev/` | 开发调试端点 | ~4 | 保留，不纳入合同 |
| `/healthz`, `/readyz` | 健康检查 | 2 | 保留，不纳入 OpenAPI 主合同 |

### 4.2 新端点分配策略

| 业务域 | 首选版本 | 理由 |
|--------|---------|------|
| Auth（login/register） | 不加版本，或 `/api/v2/auth/login` | 不涉及兼容问题，保留现有路径 |
| 公司、建筑、生产 | `/api/v2/` | 当前前端映射最多的路径 |
| 市场行情、深度、资源 | `/api/v3/` | 已经是 v3 路径，保持统一 |
| 财务报表、债券 | `/api/v2/` | 现有路径为 v2 |
| 研发、高管 | `/api/v2/` | 现有路径为 v2 |
| 新功能域（Chat 等） | `/api/v3/` | v2 已有大量路径，新域放 v3 减少冲突 |

### 4.3 标记 Legacy

在 OpenAPI 规范中，所有计划淘汰的端点添加：

```yaml
x-deprecated: true
x-deprecation-message: "Will be removed in 2026-Q3. Use /api/v2/xxx instead."
```

---

## 5. 认证方案

### 5.1 流程

```
1. 客户端 POST /api/login { "username": "...", "password": "..." }
2. 服务端验证凭据，返回 JWT token 及玩家/公司信息
3. 后续请求在 Authorization 头携带 Bearer token
4. 中间件 middleware/auth.go 验证 token 并注入 companyID / playerID 到 context
```

### 5.2 Token 格式

- **类型**：JWT (HMAC-SHA256)
- **Header**: `{ "alg": "HS256", "typ": "JWT" }`
- **Payload 字段**：
  - `sub`: player ID (string)
  - `company_id`: company ID (int)
  - `iat`: 签发时间
  - `exp`: 过期时间（当前 24 小时）
- **密钥来源**：`backend/configs/game.json` 或环境变量 `JWT_SECRET`

### 5.3 登录响应

```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "player_id": 1,
    "username": "player1",
    "company_id": 1
  },
  "error": null,
  "meta": {}
}
```

### 5.4 认证错误码

| 错误码 | 场景 |
|--------|------|
| `UNAUTHORIZED` | Token 缺失、过期、签名无效 |
| `FORBIDDEN` | Token 有效但无权访问该资源 |

### 5.5 当前实现文件

- `backend/internal/middleware/auth.go` — JWT 验证中间件
- `backend/internal/handler/auth.go` — `/api/login`, `/api/register` handler
- `backend/internal/service/auth_test.go` — 认证服务测试

---

## 6. 错误码编目（首批 20 个）

| # | 错误码 | HTTP | 场景 | 首次使用端点 |
|---|--------|------|------|-------------|
| 1 | `INSUFFICIENT_FUNDS` | 409 | 公司资金不足，无法完成购买/投资 | POST /api/v2/market-order |
| 2 | `INSUFFICIENT_INVENTORY` | 409 | 库存不足，无法出售或用于生产 | POST /api/v2/market-order |
| 3 | `INVALID_PRICE` | 400 | 下单价格超出市场允许范围 | POST /api/v2/market-order |
| 4 | `INVALID_TICK` | 400 | 生产 tick 数据异常 | POST /api/v2/production/claim |
| 5 | `ORDER_NOT_FOUND` | 404 | 指定的市场订单不存在 | POST /api/v2/market-order（撤单） |
| 6 | `COMPANY_NOT_FOUND` | 404 | 公司不存在或已删除 | GET /api/v2/companies/me |
| 7 | `BUILDING_NOT_FOUND` | 404 | 建筑不存在 | GET /api/v2/companies/me/buildings |
| 8 | `PRODUCTION_NOT_FOUND` | 404 | 生产任务不存在 | POST /api/v2/production/claim |
| 9 | `PRODUCTION_NOT_READY` | 409 | 生产尚未完成，不可领取 | POST /api/v2/production/claim |
| 10 | `BOND_NOT_FOUND` | 404 | 债券不存在 | GET /api/v2/bonds/market |
| 11 | `RESEARCH_NOT_FOUND` | 404 | 研发项目不存在 | GET /api/v2/research/start |
| 12 | `EXECUTIVE_NOT_FOUND` | 404 | 高管不存在 | POST /api/v2/executives/recruit |
| 13 | `CONTRACT_NOT_FOUND` | 404 | 合同不存在 | GET /api/v3/contracts-incoming |
| 14 | `BID_NOT_FOUND` | 404 | 竞标/拍卖出价不存在 | POST /api/v2/auctions（预留） |
| 15 | `INVALID_REQUEST` | 400 | 请求体格式错误、缺少必填字段 | 所有端点 |
| 16 | `UNAUTHORIZED` | 401 | Token 缺失或无效 | 所有需要认证的端点 |
| 17 | `FORBIDDEN` | 403 | Token 有效但无权操作 | 所有需要授权的端点 |
| 18 | `RATE_LIMITED` | 429 | 请求超过频率限制 | 全局中间件 |
| 19 | `INTERNAL_ERROR` | 500 | 服务器内部未捕获错误 | 所有端点 |
| 20 | `MAINTENANCE_MODE` | 503 | 服务器维护中 | 全局中间件 |

---

## 7. 首批端点详细定义

### 7.1 POST /api/login

| 字段 | 值 |
|------|-----|
| **Tag** | Auth |
| **Auth required** | 否 |
| **当前 handler 文件** | `backend/internal/handler/auth.go` (handleLogin) |
| **标记 legacy** | 否。保留现有路径，不加版本号 |
| **错误码** | `INVALID_REQUEST` (400), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request** (`LoginRequest`):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 玩家用户名 |
| `password` | string | 是 | 玩家密码 |

**Response** (`LoginResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | string | JWT Bearer token |
| `player_id` | int | 玩家 ID |
| `username` | string | 玩家用户名 |
| `company_id` | int | 当前公司 ID |

---

### 7.2 POST /api/register

| 字段 | 值 |
|------|-----|
| **Tag** | Auth |
| **Auth required** | 否 |
| **当前 handler 文件** | `backend/internal/handler/auth.go` (handleRegister) |
| **标记 legacy** | 否 |
| **错误码** | `INVALID_REQUEST` (400), `INTERNAL_ERROR` (500) |

**Request** (`RegisterRequest`):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `username` | string | 是 | 玩家用户名，唯一 |
| `password` | string | 是 | 密码（明文传输，服务端 bcrypt 加密） |
| `name` | string | 否 | 显示名称，默认同 username |
| `gender` | string | 否 | 性别，可选值 `male`/`female`/空 |
| `email` | string | 否 | 邮箱 |

**Response** (`LoginResponse`):

与 LoginResponse 结构一致（注册即登录，直接返回 token）。

---

### 7.3 GET /api/v2/companies/me

| 字段 | 值 |
|------|-----|
| **Tag** | Companies |
| **Auth required** | 是 (`UNAUTHORIZED`, `FORBIDDEN`) |
| **当前 handler 文件** | `backend/internal/handler/company.go` (handleV2Companies) |
| **标记 legacy** | 否 |
| **错误码** | `COMPANY_NOT_FOUND` (404), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`CompanyResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 公司 ID |
| `name` | string | 公司名称 |
| `money` | float64 | 当前现金 |
| `simcash` | float64 | 当前 SimCash（高级货币） |
| `level` | int | 公司等级 |
| `reputation` | int | 声望值 |
| `player_id` | int | 所属玩家 ID |
| `inventory` | []InventoryEntry | 库存明细（可选） |

**`InventoryEntry` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `resource_id` | string | 资源标识符 |
| `name` | string | 资源显示名称 |
| `quantity` | float64 | 当前持有数量 |
| `capacity` | float64 | 最大容量 |

---

### 7.4 GET /api/v2/companies/me/buildings

| 字段 | 值 |
|------|-----|
| **Tag** | Buildings |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/company.go` (handleCompaniesMeBuildings) |
| **标记 legacy** | 否 |
| **错误码** | `COMPANY_NOT_FOUND` (404), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`BuildingListResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | []BuildingResponse | 建筑列表 |
| `total` | int | 建筑总数 |

**`BuildingResponse` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 建筑 ID（如 `"farm"`, `"bakery"`） |
| `name` | string | 建筑显示名称 |
| `level` | int | 当前等级 |
| `production_status` | string | 状态：`idle` / `busy` / `locked` |
| `production_end_time` | string (RFC3339) | 生产结束时间（若 busy） |
| `position` | int | 地图上的位置索引 |
| `slots` | int | 生产队列槽位数 |

---

### 7.5 GET /api/v2/production/jobs

| 字段 | 值 |
|------|-----|
| **Tag** | Production |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/production.go` (handleProductionJobs) |
| **标记 legacy** | 否 |
| **错误码** | `COMPANY_NOT_FOUND` (404), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`ProductionJobListResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | []ProductionJobResponse | 活跃的生产作业列表 |

**`ProductionJobResponse` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 任务 ID |
| `building_id` | string | 所属建筑 ID |
| `recipe_id` | string | 配方 ID |
| `start_time` | string (RFC3339) | 开始时间 |
| `end_time` | string (RFC3339) | 预计结束时间 |
| `status` | string | `active` / `claimable` / `completed` |
| `output_resource_id` | string | 产出资源 ID |
| `output_quantity` | float64 | 产出数量 |

---

### 7.6 GET /api/v3/market-ticker

| 字段 | 值 |
|------|-----|
| **Tag** | Market |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/market.go` (handleMarketTicker) |
| **标记 legacy** | 否 |
| **错误码** | `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`MarketTickerListResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | []TickerEntryResponse | 各资源的行情快照 |
| `timestamp` | string (RFC3339) | 行情快照时间戳 |

**`TickerEntryResponse` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `resource_id` | string | 资源标识符 |
| `resource_name` | string | 资源名 |
| `price` | float64 | 当前价格 |
| `price_change_24h` | float64 | 24 小时价格变化（百分比） |
| `volume_24h` | float64 | 24 小时成交量 |
| `high_24h` | float64 | 24 小时最高价 |
| `low_24h` | float64 | 24 小时最低价 |

---

### 7.7 GET /api/v3/market-depth

| 字段 | 值 |
|------|-----|
| **Tag** | Market |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/market.go` (handleMarketDepth) |
| **标记 legacy** | 否 |
| **错误码** | `INVALID_REQUEST` (400), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request Query 参数**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `resource_id` | string | 是 | 资源 ID（如 `"wheat"`） |

**Response** (`MarketDepthResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `resource_id` | string | 资源 ID |
| `bids` | []DepthLevel | 买单深度 |
| `asks` | []DepthLevel | 卖单深度 |

**`DepthLevel` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `price` | float64 | 价格 |
| `quantity` | float64 | 数量 |
| `total` | float64 | 累计数量 |

---

### 7.8 GET /api/v3/resources

| 字段 | 值 |
|------|-----|
| **Tag** | Resources |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/market.go` (handleResources) |
| **标记 legacy** | 否 |
| **错误码** | `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`ResourceListResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | []ResourceResponse | 资源列表（含公司当前持有量） |

**`ResourceResponse` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 资源 ID |
| `name` | string | 资源名称 |
| `quantity` | float64 | 当前持有量 |
| `capacity` | float64 | 最大容量 |

---

### 7.9 GET /api/v3/resources-info

| 字段 | 值 |
|------|-----|
| **Tag** | Resources |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/market.go` (handleResourceInfo) |
| **标记 legacy** | 否 |
| **错误码** | `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`ResourceInfoListResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | []ResourceInfoResponse | 资源静态信息列表（全局定义，非公司持有） |

**`ResourceInfoResponse` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 资源 ID |
| `name` | string | 资源名称 |
| `category` | string | 分类（如 `"raw"`, `"processed"`, `"consumable"`） |
| `base_price` | float64 | 基础价格（来自 formula） |
| `unit` | string | 单位（如 `"kg"`, `"unit"`） |

---

### 7.10 GET /api/v2/companies/me/income-statement

| 字段 | 值 |
|------|-----|
| **Tag** | Finance |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/financial.go` (handleIncomeStatement) |
| **标记 legacy** | 否 |
| **错误码** | `COMPANY_NOT_FOUND` (404), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`FinancialStatementResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `revenue` | float64 | 总收入 |
| `cost_of_goods_sold` | float64 | 销货成本 |
| `gross_profit` | float64 | 毛利润 |
| `operating_expenses` | float64 | 运营开支 |
| `net_income` | float64 | 净利润 |
| `period_start` | string (RFC3339) | 报告期开始 |
| `period_end` | string (RFC3339) | 报告期结束 |

---

### 7.11 GET /api/v2/companies/me/balance-sheet

| 字段 | 值 |
|------|-----|
| **Tag** | Finance |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/financial.go` (handleBalanceSheet) |
| **标记 legacy** | 否 |
| **错误码** | `COMPANY_NOT_FOUND` (404), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`FinancialStatementResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `total_assets` | float64 | 总资产 |
| `total_liabilities` | float64 | 总负债 |
| `equity` | float64 | 所有者权益 |
| `cash` | float64 | 现金及现金等价物 |
| `inventory_value` | float64 | 库存价值 |
| `accounts_receivable` | float64 | 应收账款 |
| `short_term_debt` | float64 | 短期债务 |
| `long_term_debt` | float64 | 长期债务 |
| `as_of_date` | string (RFC3339) | 资产负债表日期 |

---

### 7.12 GET /api/v2/bonds/market

| 字段 | 值 |
|------|-----|
| **Tag** | Finance |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/bond.go` (handleBonds) |
| **标记 legacy** | 否 |
| **错误码** | `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request**: 无参数。

**Response** (`BondMarketListResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `items` | []BondResponse | 债券市场列表 |

**`BondResponse` 子结构**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 债券 ID |
| `issuer_company_id` | int | 发行公司 ID |
| `issuer_name` | string | 发行公司名称 |
| `face_value` | float64 | 面值 |
| `coupon_rate` | float64 | 票面利率（百分比） |
| `remaining_term_days` | int | 剩余期限（天） |
| `status` | string | `active` / `called` / `matured` |

---

### 7.13 POST /api/v2/market-order

| 字段 | 值 |
|------|-----|
| **Tag** | Market |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/market.go` (handleCreateOrder) |
| **标记 legacy** | 否 |
| **错误码** | `INSUFFICIENT_FUNDS` (409), `INSUFFICIENT_INVENTORY` (409), `INVALID_PRICE` (400), `INVALID_REQUEST` (400), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request** (`CreateOrderRequest`):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `resource_id` | string | 是 | 资源 ID |
| `order_type` | string | 是 | `buy` 或 `sell` |
| `price` | float64 | 是 | 单价，必须在市场允许范围内 |
| `quantity` | float64 | 是 | 数量，必须 > 0 |

**Response** (`OrderResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `order_id` | string | 订单 ID |
| `resource_id` | string | 资源 ID |
| `order_type` | string | `buy` / `sell` |
| `price` | float64 | 成交价或挂单价 |
| `quantity` | float64 | 成交数量 |
| `status` | string | `filled` / `partial` / `open` |
| `filled_quantity` | float64 | 已成交数量 |
| `created_at` | string (RFC3339) | 创建时间 |

---

### 7.14 POST /api/v2/production/claim

| 字段 | 值 |
|------|-----|
| **Tag** | Production |
| **Auth required** | 是 |
| **当前 handler 文件** | `backend/internal/handler/production.go` (handleClaimProduction) |
| **标记 legacy** | 否 |
| **错误码** | `PRODUCTION_NOT_FOUND` (404), `PRODUCTION_NOT_READY` (409), `INVALID_REQUEST` (400), `UNAUTHORIZED` (401), `INTERNAL_ERROR` (500) |

**Request** (`ClaimRequest`):

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `job_id` | string | 是 | 生产任务 ID |

**Response** (`ClaimResponse`):

| 字段 | 类型 | 说明 |
|------|------|------|
| `job_id` | string | 任务 ID |
| `resource_id` | string | 产出的资源 ID |
| `quantity` | float64 | 产出的数量 |
| `experience_gained` | int | 获得的经验值（如适用） |

---

## 8. 遗留端点标记（已识别）

以下端点已在代码中存在，需在 OpenAPI 规范中标记 `x-deprecated: true`：

| 端点 | 所在文件 | 说明 | 替代建议 |
|------|----------|------|---------|
| `/api/v1/buildings/` | `production.go` | V1 建筑端点 | `/api/v2/buildings/` |
| `/api/market/buy/orders/` | `market.go` | 无版本号市场端点 | `/api/v3/market-depth` |
| `/api/bonds/` | `bond.go` | 无版本号债券端点 | `/api/v2/bonds/market` |
| `/api/bonds/settle-interest/` | `bond.go` | 无版本号 | TBD |
| `/api/csrf/` | `company.go` | CSRF token（已不再必要） | TBD |
| `/api/v4/` | `dev.go` | V4 catch-all（payment-packages 等）| 不再添加新路由 |
| `/api/v2/market-order/take/` | `market.go` | 待评估 | TBD |
| `/api/v2/market-order/cancel/` | `market.go` | 待评估 | TBD |

---

## 9. OpenAPI 规范文件

### 9.1 存放位置

```
project-root/
  openapi/
    openapi-draft.yaml     # Phase 1 初始草案
```

### 9.2 文件结构概览

```yaml
openapi: 3.1.0
info:
  title: New Haven Game API
  version: 0.1.0-draft
  description: |
    Phase 1 API contract for New Haven backend refactor.
    Current API versions: /api, /api/v1 (deprecated), /api/v2, /api/v3, /api/v4 (deprecated).

servers:
  - url: http://localhost:8080
    description: Local dev server
  - url: https://api.newhaven.game
    description: Production (when available)

tags:
  - name: Auth
    description: Authentication and registration
  - name: Companies
    description: Company profile, inventory, buildings
  # …其余 tag

paths:
  /api/login:
    post:
      tags: [Auth]
      operationId: login
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        '400':
          $ref: '#/components/responses/InvalidRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'

  /api/v2/market-order:
    post:
      tags: [Market]
      operationId: createOrder
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '200':
          description: Order created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrderResponse'
        '400':
          $ref: '#/components/responses/InvalidRequest'
        '409':
          $ref: '#/components/responses/Conflict'
        # …401, 500
```

### 9.3 components 组织

```
components:
  schemas:
    # 信封
    ApiSuccessResponse
    ApiErrorResponse
    MetaPayload
    ErrorPayload

    # 业务 DTO（每个端点一个）
    LoginRequest, LoginResponse
    RegisterRequest
    CompanyResponse
    BuildingResponse, BuildingListResponse
    ProductionJobResponse, ProductionJobListResponse
    MarketTickerListResponse, TickerEntryResponse
    MarketDepthResponse, DepthLevel
    ResourceListResponse, ResourceResponse
    ResourceInfoListResponse, ResourceInfoResponse
    FinancialStatementResponse
    BondMarketListResponse, BondResponse
    CreateOrderRequest, OrderResponse
    ClaimRequest, ClaimResponse

    # 枚举
    ErrorCode
    OrderType (buy / sell)
    ProductionStatus (active / claimable / completed)
    BuildingStatus (idle / busy / locked)

  responses:
    InvalidRequest
    Unauthorized
    Forbidden
    NotFound
    Conflict
    RateLimited
    InternalError
    MaintenanceMode

  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

---

## 10. 实施建议

### 10.1 第一阶段（本 Phase 内完成）

1. **创建 `handler/dto_response.go`** — 实现 `writeSuccess(w, data)`, `writeError(w, status, code, msg, details)` 辅助函数。
2. **创建 `handler/error_codes.go`** — 定义 `ErrorCode` 常量和默认 HTTP 状态码映射。
3. **逐个端点迁移** — 从首批端点列表开始，为每个端点：
   - 定义 DTO struct（`handler/dto_xxx.go`）
   - 替换 `writeJSON(w, 200, map[string]any{...})` 为 `writeSuccess(w, dto)`
   - 替换 `writeErr(w, 400, msg)` 为 `writeError(w, 400, ErrorInvalidRequest, msg, nil)`
4. **编写 `openapi/openapi-draft.yaml`** — 覆盖首批端点。

### 10.2 第二阶段（Phase 1 收尾前）

1. 为所有现有 handler 增加 DTO（约 22 个 handler 文件）。
2. 用结构化错误码替换所有字符串错误。
3. 确保前端（`client/atlas-foods-client/`）使用新的统一响应解析路径。

### 10.3 不做的（Phase 1 范围外）

- 不生成 OpenAPI 代码（只写 spec 文件）。
- 不引入第三方路由库（保留 `http.ServeMux`，等端点稳定）。
- 不重构 service 层返回类型（只改变 handler 层 DTO）。
- 不修改前端现有 API 调用代码（响应信封由前端 adapter 层处理）。

---

## 附录 A：当前路由全景

以下为 Phase 1 开工前 `Register*()` 中注册的全部路由，按版本分组：

| 前缀 | 路由 | 方法 | Handler 文件 | 标签 |
|------|------|------|-------------|------|
| (无版本) | `/healthz` | GET | health.go | Health |
| (无版本) | `/readyz` | GET | health.go | Health |
| (无版本) | `/api/login` | POST | auth.go | Auth |
| (无版本) | `/api/register` | POST | auth.go | Auth |
| (无版本) | `/api/csrf/` | GET | company.go | Companies |
| (无版本) | `/api/bonds/` | GET/POST | bond.go | Finance |
| (无版本) | `/api/bonds/settle-interest/` | POST | bond.go | Finance |
| (无版本) | `/api/market/buy/orders/` | GET | market.go | Market |
| V1 | `/api/v1/buildings/` | * | production.go | Buildings |
| V2 | `/api/v2/buildings/` | * | production.go | Buildings |
| V2 | `/api/v2/players/me/companies/` | GET | company.go | Companies |
| V2 | `/api/v2/players/` | GET | company.go | Companies |
| V2 | `/api/v2/companies/me/buildings/` | GET | company.go | Buildings |
| V2 | `/api/v2/companies/me/administration-overhead/` | GET | company.go | Companies |
| V2 | `/api/v2/companies/` | GET | company.go | Companies |
| V2 | `/api/v2/companies/me/income-statement/` | GET | financial.go | Finance |
| V2 | `/api/v2/companies/me/balance-sheet/` | GET | financial.go | Finance |
| V2 | `/api/v2/companies/me/cashflow-statement/` | GET | financial.go | Finance |
| V2 | `/api/v2/companies/me/cashflow/recent/` | GET | financial.go | Finance |
| V2 | `/api/v2/companies/me/past-finances-overview/` | GET | financial.go | Finance |
| V2 | `/api/v2/production/jobs/` | GET | production.go | Production |
| V2 | `/api/v2/production/claim/` | POST | production.go | Production |
| V2 | `/api/v2/production/claimable/` | GET | production.go | Production |
| V2 | `/api/v2/production/claim-all/` | POST | production.go | Production |
| V2 | `/api/v2/production/queue/` | GET | production_queue.go | Production |
| V2 | `/api/v2/production/slots/add/` | POST | production_queue.go | Production |
| V2 | `/api/v2/production/cancel/` | POST | production_queue.go | Production |
| V2 | `/api/v2/production-modifiers/` | GET | market.go | Production |
| V2 | `/api/v2/market-order/` | POST | market.go | Market |
| V2 | `/api/v2/market-order/cancel/` | POST | market.go | Market |
| V2 | `/api/v2/market-order/take/` | POST | market.go | Market |
| V2 | `/api/v2/weather/` | GET | market.go | Market |
| V2 | `/api/v2/executives/search/` | GET | company.go | Executives |
| V2 | `/api/v2/executives/recruit/` | POST | company.go | Executives |
| V2 | `/api/v2/executives/train/` | POST | company.go | Executives |
| V2 | `/api/v2/companies/me/warehouse/` | GET | company.go | Warehouse |
| V2 | `/api/v2/companies/me/tutorial/` | POST | company.go | Companies |
| V2 | `/api/v2/companies/me/auctions/` | GET | company.go | Companies |
| V2 | `/api/v2/auctions/` | * | company.go | Companies |
| V2 | `/api/v2/research/` | GET | dev.go | Research |
| V2 | `/api/v2/research/start/` | POST | dev.go | Research |
| V2 | `/api/v2/research/progress/` | GET | dev.go | Research |
| V2 | `/api/v2/research/complete/` | POST | dev.go | Research |
| V2 | `/api/v2/companies/me/bonds/owned/` | GET | bond.go | Finance |
| V2 | `/api/v2/companies/me/bonds/sold/` | GET | bond.go | Finance |
| V2 | `/api/v2/contracts-history-incoming/` | GET | dev.go | (合同域) |
| V2 | `/api/v2/contracts-history-outgoing/` | GET | dev.go | (合同域) |
| V2 | `/api/v2/warehouse-contracts-summary/` | GET | dev.go | (合同域) |
| V2 | `/api/v2/orders/daily/` | GET | order.go | (日订单) |
| V2 | `/api/v2/orders/daily/complete/` | POST | order.go | (日订单) |
| V2 | `/api/v2/orders/daily/claim/` | POST | order.go | (日订单) |
| V3 | `/api/v3/market-ticker/` | GET | market.go | Market |
| V3 | `/api/v3/market/` | GET | market.go | Market |
| V3 | `/api/v3/market-depth/` | GET | market.go | Market |
| V3 | `/api/v3/resources/` | GET | market.go | Resources |
| V3 | `/api/v3/resources-info/` | GET | market.go | Resources |
| V3 | `/api/v3/companies/` | GET | company.go | Companies |
| V3 | `/api/v3/companies/me/past-finances/` | GET | financial.go | Finance |
| V3 | `/api/v3/executives/poach/` | POST | company.go | Executives |
| V3 | `/api/v3/executives/offers/` | GET | company.go | Executives |
| V3 | `/api/v3/executives/` | GET | company.go | Executives |
| V3 | `/api/v3/contracts-incoming/` | GET | dev.go | (合同域) |
| V3 | `/api/v3/contracts-outgoing/me/` | GET | dev.go | (合同域) |
| V4 | `/api/v4/` | * | dev.go | (实验性) |
| Dev | `/api/dev/ledger/` | GET | dev.go | (开发) |
| Dev | `/api/dev/formulas/production/` | GET | dev.go | (开发) |
| Dev | `/api/dev/formulas/retail/` | GET | dev.go | (开发) |
| Dev | `/api/dev/formulas/retail-season-weather/` | GET | dev.go | (开发) |
| Dev | `/api/dev/time/` | POST | dev.go | (开发) |

> 路由总数约 70+。首批端点定义覆盖其中 14 个最核心的路径。
