# NewHaven API — JSON 请求/响应示例

生成方式: OpenAPI schema 定义 + 测试文件中的真实 JSON 提取

适用场景: 复制 JSON 调接口、核对字段名、快速定位常用请求/响应结构。

---

## 目录

- [使用约定](#使用约定)
- [认证与健康检查](#认证与健康检查)
- [公司与玩家](#公司与玩家)
- [建筑与货架](#建筑与货架)
- [仓库](#仓库)
- [生产](#生产)
- [市场](#市场)
- [债券](#债券)
- [通用响应格式](#通用响应格式)
- [错误码](#错误码)
- [文档维护建议](#文档维护建议)

---

## 使用约定

### 字段来源

| 标记 | 含义 |
|------|------|
| 已测试 | JSON 来自测试文件，优先可信 |
| Schema | JSON 从 OpenAPI schema 推断，实际值需按游戏内状态调整 |
| 无 body | 请求不需要 JSON body，通常只依赖路径参数或 token |

### 认证

除以下接口外，其余接口默认需要登录 token：

- `/healthz`
- `/readyz`
- `/api/register`
- `/api/login`
- `/api/admin/snapshot/*`

请求头：

```http
Authorization: Bearer <token>
```

Token 从 `POST /api/login` 的响应 `data.token` 获取。

### 响应信封

业务数据通常包在 `data` 里：

```json
{
  "data": {
    "id": "example"
  },
  "error": null,
  "meta": {
    "request_id": "req-xxx"
  }
}
```

下方“响应示例”只展示 `data` 内部结构，除非专门说明完整信封。

---

## 认证与健康检查

### GET /healthz

认证: 不需要  
来源: Schema `HealthResponse`

响应 `data` 示例：

```json
{
  "status": "ok"
}
```

### GET /readyz

认证: 不需要  
来源: Schema `HealthResponse`

响应 `data` 示例：

```json
{
  "status": "ok"
}
```

### POST /api/register

认证: 不需要  
来源: 已测试 `auth_handler_test.go:38` + Schema `RegisterRequest`

最小请求：

```json
{
  "username": "alice",
  "password": "secret123"
}
```

带可选字段：

```json
{
  "username": "newplayer",
  "password": "secret123",
  "name": "Alice",
  "gender": "M",
  "email": "alice@example.com"
}
```

验证错误示例：

```json
{}
```

预期: `400 VALIDATION_ERROR`

### POST /api/login

认证: 不需要  
来源: 已测试 `auth_handler_test.go:179` + Schema `LoginRequest`

请求：

```json
{
  "username": "bob",
  "password": "hunter2"
}
```

错误密码示例：

```json
{
  "username": "carol",
  "password": "wrongpw"
}
```

预期: `401 UNAUTHORIZED`

成功响应 `data` 示例：

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "player_id": 42,
  "company_id": 1,
  "username": "newplayer"
}
```

---

## 公司与玩家

### GET /api/v2/players/me/companies/

认证: 需要  
来源: Schema `MyCompaniesResponse`

响应 `data` 示例：

```json
{
  "companies": [
    {
      "id": 1,
      "name": "NewHaven Corp",
      "money": 50000.0,
      "level": 5
    }
  ]
}
```

### GET /api/v3/companies/{companyId}/

认证: 需要  
路径参数: `companyId`  
来源: Schema `CompanyProfileResponse`

响应 `data` 示例：

```json
{
  "authCompany": {
    "id": 1,
    "name": "NewHaven Corp",
    "money": 50000.0,
    "level": 5,
    "xp": 1200
  },
  "authUser": {
    "id": 42,
    "username": "newplayer"
  },
  "levelInfo": {
    "level": 5,
    "xp": 1200,
    "xpToNext": 2000
  },
  "preferences": {
    "theme": "dark",
    "notifications": true
  }
}
```

### PATCH /api/v2/companies/me/story-progress/

认证: 需要  
来源: Schema `StoryProgressRequest`

请求：

```json
{
  "storyId": "tutorial_intro",
  "stepId": "step_build_factory",
  "status": "completed"
}
```

字段说明：

- `status`: `not_started`, `in_progress`, `completed`, `skipped`

响应 `data` 示例：

```json
{
  "status": "completed",
  "stepId": "step_build_factory"
}
```

---

## 建筑与货架

### GET /api/v3/companies/me/buildings/

认证: 需要  
来源: Schema `BuildingDTO`

响应 `data` 示例：

```json
{
  "id": "bld-uuid-abc123",
  "building_id": 1,
  "name": "Steel Factory",
  "level": 2,
  "map_id": "map-main",
  "slot_id": "A1",
  "x": 10,
  "y": 20,
  "robot_count": 3,
  "placed": true,
  "is_retail": false,
  "produces": [3],
  "shelves": []
}
```

### GET /api/v2/buildings/market/

认证: 需要  
来源: Schema `BuildingMarketItem`

响应 `data` 示例：

```json
{
  "id": "bld-factory-1",
  "name": "Steel Factory",
  "kind": 1,
  "cost": 5000.0,
  "unlock_level": 1,
  "description": "Produces steel bars from iron ore",
  "produces": [3],
  "starter_produces": [1, 3],
  "starter_role": "industrialist",
  "is_retail": false
}
```

### POST /api/v2/buildings/buy/

认证: 需要  
来源: Schema `BuyBuildingRequest`

请求：

```json
{
  "buildingId": "bld-factory-1",
  "requestId": "req-a1b2c3d4"
}
```

字段说明：

- `buildingId`: 来自建筑市场列表
- `requestId`: 幂等键，可选

响应 `data` 示例：

```json
{
  "building": {
    "id": "bld-uuid",
    "building_id": 1,
    "name": "Factory",
    "level": 1,
    "x": 10,
    "y": 20
  },
  "cost": 5000.0
}
```

### POST /api/v2/buildings/place/

认证: 需要  
来源: Schema `PlaceBuildingRequest`

请求：

```json
{
  "buildingId": "bld-factory-1",
  "mapId": "map-main",
  "slotId": "A1",
  "x": 15,
  "y": 30
}
```

响应 `data` 示例：

```json
{
  "building": {
    "id": "bld-uuid",
    "building_id": 1,
    "name": "Factory",
    "level": 1,
    "x": 15,
    "y": 30
  },
  "status": "ok"
}
```

### POST /api/v2/buildings/move/

认证: 需要  
来源: Schema `MoveBuildingRequest`

请求：

```json
{
  "buildingId": "bld-uuid-abc",
  "mapId": "map-main",
  "slotId": "B2",
  "x": 20,
  "y": 45
}
```

### POST /api/v2/buildings/demolish/

认证: 需要  
来源: Schema `DemolishBuildingRequest`

请求：

```json
{
  "buildingId": "bld-uuid-abc"
}
```

响应 `data` 示例：

```json
{
  "refund": 2500.0,
  "status": "ok"
}
```

### POST /api/v1/buildings/{buildingId}/upgrade/

认证: 需要  
路径参数: `buildingId`  
来源: Schema `UpgradeBuildingResponse`

请求: 无 body

响应 `data` 示例：

```json
{
  "building_id": "bld-uuid-abc",
  "old_level": 2,
  "new_level": 3,
  "cost": 8000.0,
  "output_multiplier": 150
}
```

### POST /api/v2/buildings/{buildingId}/stock/

认证: 需要  
路径参数: `buildingId`  
来源: Schema `StockShelfRequest`

请求：

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "quantity": 50,
  "price": 12.5
}
```

字段说明：

- `price`: 可选，不传则沿用上次价格
- 注意: 路径里已有 `buildingId`，body 里仍存在 `building_id` 字段

### POST /api/v2/buildings/{buildingId}/unstock/

认证: 需要  
路径参数: `buildingId`  
来源: Schema `UnstockShelfRequest`

请求：

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "quantity": 20
}
```

### POST /api/v2/buildings/{buildingId}/shelf-price/

认证: 需要  
路径参数: `buildingId`  
来源: Schema `SetShelfPriceRequest`

请求：

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "price": 15.0,
  "lock": true
}
```

字段说明：

- `lock: true`: 防止自动调价覆盖

### 货架响应示例

来源: Schema `ShelfItem`

```json
{
  "resource_id": 5,
  "quantity": 50,
  "max_qty": 200,
  "price": 12.5,
  "price_lock": false,
  "revenue": 625.0
}
```

---

## 仓库

### GET /api/v2/companies/me/warehouse/

认证: 需要  
来源: Schema `GetMyWarehouseData`

响应 `data` 示例：

```json
{
  "company_id": 1,
  "capacity": 5000,
  "used_capacity": 1250,
  "items": [
    {
      "resource_id": 1,
      "resource_name": "Iron Ore",
      "quality": 0,
      "amount": 500
    },
    {
      "resource_id": 3,
      "resource_name": "Steel Bar",
      "quality": 0,
      "amount": 200
    }
  ]
}
```

### POST /api/v2/companies/me/warehouse/upgrade/

认证: 需要  
来源: Schema `WarehouseUpgradeResponse`

请求: 无 body

响应 `data` 示例：

```json
{
  "level": 2,
  "capacity": 10000,
  "cost": 5000.0
}
```

---

## 生产

### GET /api/v2/production/jobs/

认证: 需要  
来源: Schema `ProductionJobListResponse`

请求: 无 body

用途: 查看当前生产任务列表。

### POST /api/v2/production/start/

认证: 需要  
来源: 已测试 `production_handler_test.go:234` + Schema `StartProductionRequest`

请求：

```json
{
  "building_id": "bld-1",
  "resource_id": 3,
  "quantity": 10
}
```

响应 `data` 示例：

```json
{
  "job": {
    "id": "job-uuid",
    "building_id": "bld-1",
    "resource_id": 3,
    "quantity": 10,
    "status": "running"
  }
}
```

### POST /api/v2/production/claim/{jobId}/

认证: 需要  
路径参数: `jobId`  
请求: 无 body

### GET /api/v2/production/claimable/

认证: 需要  
请求: 无 body

### GET /api/v2/production/queue/

认证: 需要  
请求: 无 body

### POST /api/v2/production/cancel/

认证: 需要  
来源: Schema `CancelJobRequest`

请求示例：

```json
{
  "job_id": "job-uuid"
}
```

### POST /api/v2/production/claim-all/

认证: 需要  
请求: 无 body

### GET /api/v2/buildings/{buildingId}/production-options/

认证: 需要  
路径参数: `buildingId`  
请求: 无 body

---

## 市场

### GET /api/v3/resources/

认证: 需要  
来源: Schema `ResourcesResponse`

响应 `data` 示例：

```json
{
  "resources": [
    {
      "resourceId": 1,
      "name": "Iron Ore",
      "producedFrom": {
        "mine": 2
      },
      "producedPerHourRaw": 100,
      "unitsSoldAnHour": 50,
      "hasEconomyModel": true
    }
  ]
}
```

### GET /api/v3/market-ticker/{resourceId}/

认证: 需要  
路径参数: `resourceId`  
来源: Schema `MarketTicker`

响应 `data` 示例：

```json
{
  "resource_id": 1,
  "last_price": 12.5,
  "volume_24h": 10000,
  "high_24h": 13.2,
  "low_24h": 11.8,
  "price_change_24h": 0.5,
  "updated_at": "2026-06-12T10:30:00Z"
}
```

### GET /api/v3/market-depth/{resourceId}/{quality}/

认证: 需要  
路径参数: `resourceId`, `quality`  
来源: Schema `MarketDepthResponse`

响应 `data` 示例：

```json
{
  "buys": [
    {
      "price": 9.5,
      "quantity": 100,
      "qty": 100
    },
    {
      "price": 9.0,
      "quantity": 200,
      "qty": 200
    }
  ],
  "sells": [
    {
      "price": 10.5,
      "quantity": 150,
      "qty": 150
    },
    {
      "price": 11.0,
      "quantity": 80,
      "qty": 80
    }
  ]
}
```

### GET /api/v3/market/{resourceId}/{quality}/

认证: 需要  
路径参数: `resourceId`, `quality`  
来源: Schema `MarketOrderListResponse`

响应 `data` 示例：

```json
{
  "orders": [
    {
      "id": "ord-uuid-xyz",
      "resourceId": 1,
      "kind": 0,
      "price": 10.0,
      "quality": 0,
      "quantity": 5,
      "remaining": 3,
      "companyId": 1,
      "createdAt": "2026-06-12T08:00:00Z",
      "status": "open"
    }
  ]
}
```

### POST /api/v2/market-order/

认证: 需要  
来源: 已测试 `market_handler_test.go:396` + Schema `CreateOrderRequestFrontend`

请求：

```json
{
  "resourceId": 1,
  "kind": 1,
  "quality": 0,
  "quantity": 5,
  "price": 10.0
}
```

字段说明：

- `kind`: `0` = buy, `1` = sell
- `requestId`: 可选幂等键
- 注意: 市场前端接口使用 camelCase 字段，如 `resourceId`

### DELETE /api/v2/market-order/cancel/{orderId}/

认证: 需要  
路径参数: `orderId`  
来源: 已测试 `market_handler_test.go`

请求: 无 body

响应 `data` 示例：

```json
{
  "id": "ord-uuid-xyz",
  "status": "cancelled"
}
```

### POST /api/v2/market-order/take/

认证: 需要  
来源: 已测试 `market_handler_test.go:626` + Schema `TakeOrderRequest`

请求：

```json
{
  "resource": 1,
  "quantity": 5,
  "quality": 0,
  "maxPrice": 100.0
}
```

响应 `data` 示例：

```json
{
  "amountBought": 5,
  "trades": [
    {
      "id": "trade-uuid",
      "resourceId": 1,
      "quality": 0,
      "quantity": 5,
      "price": 10.0,
      "buyOrderId": "ord-buy",
      "sellOrderId": "ord-sell",
      "createdAt": "2026-06-12T08:00:00Z"
    }
  ],
  "moneyDelta": -50.0
}
```

---

## 债券

### GET /api/bonds/

认证: 需要  
请求: 无 body
用途: 查看债券列表。

### POST /api/bonds/

认证: 需要  
来源: 已测试 `bond_handler_test.go:159` + Schema `CreateBondRequest`

请求：

```json
{
  "amount": 5,
  "interest": 1.2
}
```

验证边界示例：

```json
{
  "amount": 0,
  "interest": 0
}
```

预期: `400 VALIDATION_ERROR`

### GET /api/bonds/{bondId}/

认证: 需要  
路径参数: `bondId`  
请求: 无 body

### POST /api/bonds/{bondId}/call/

认证: 需要  
路径参数: `bondId`  
请求: 无 body

### POST /api/bonds/settle-interest/

认证: 需要  
请求: 无 body

### GET /api/v2/companies/me/bonds/owned/

认证: 需要  
请求: 无 body

### GET /api/v2/companies/me/bonds/sold/

认证: 需要  
请求: 无 body

### POST /api/v2/bonds/{bondId}/buy/

认证: 需要  
路径参数: `bondId`  
请求: 无 body

---

## 通用响应格式

成功：

```json
{
  "data": {
    "实际字段": "实际值"
  },
  "error": null,
  "meta": {
    "request_id": "req-xxx"
  }
}
```

失败：

```json
{
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "描述",
    "details": null
  }
}
```

错误响应示例：

```json
{
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "username is required",
    "details": null
  }
}
```

---

## 错误码

| 状态码 | 错误码 | 说明 |
|--------|--------|------|
| 400 | `BAD_REQUEST` | 请求格式错误 |
| 400 | `VALIDATION_ERROR` | 字段校验失败 |
| 400 | `INSUFFICIENT_FUNDS` | 余额不足 |
| 400 | `INSUFFICIENT_INVENTORY` | 库存不足 |
| 401 | `UNAUTHORIZED` | token 缺失或无效 |
| 403 | `FORBIDDEN` | 无权访问 |
| 404 | `NOT_FOUND` | 资源不存在 |
| 409 | `CONFLICT` | 冲突，如用户名已注册 |
| 429 | `RATE_LIMITED` | 频率限制 |
| 500 | `INTERNAL_ERROR` | 服务器内部错误 |

---

## 文档维护建议

### 推荐分类

当前文件按业务模块组织，建议与 `API-QUICKREF.md` 的模块保持一致：

| 模块 | 用途 |
|------|------|
| Auth / Health | 登录、注册、服务状态 |
| Company / Player | 公司资料、玩家设置、故事进度 |
| Building / Warehouse / Production | 核心经营链路 |
| Market / Bonds / Finance | 交易与财务 |
| Chat / Social | 社交消息 |
| Contract / Research / Executive / Leaderboard | 扩展玩法 |
| Admin | 开发工具 |

### 优先优化点

1. 补齐只在速查表中存在、但本文件尚未覆盖的模块示例：Chat、Social、Finance、Contract、Research、Executive、Leaderboard、Report、Admin。
2. 为每个写操作固定使用统一结构：认证、来源、请求 JSON、成功响应、错误响应。
3. 明确字段命名风格差异：建筑/仓库/生产多为 snake_case，市场前端接口多为 camelCase。
4. 标注无 body 请求，避免调试时误传空 JSON 导致 handler 行为差异。
5. 后续可从 OpenAPI 自动生成目录和 Schema 示例，再把测试提取的真实 JSON 覆盖到对应接口块。

---

## 三个文件的使用策略

| 文件 | 用途 | 搜索方式 |
|------|------|---------|
| `API-QUICKREF.md` | 查哪个接口对应哪个 handler | Ctrl+F 搜路径/模块名 |
| `API-EXAMPLES.md`（这个文件） | 复制粘贴 JSON 调接口 | Ctrl+F 搜路径名 |
| `openapi-draft.yaml` | 查字段定义、约束、类型 | `yq` 或 Ctrl+F |
