# NewHaven API — JSON 请求/响应示例

生成方式: OpenAPI schema 定义 + 测试文件中的真实 JSON 提取

---

## 测试已验证的真实 JSON 请求

以下 JSON 直接从测试代码中提取，保证可用：

### POST /api/register

来源: `auth_handler_test.go:38`

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

### POST /api/register (空 body 验证错误)

来源: `auth_handler_test.go:86`

```json
{}
```

→ 返回 400 `VALIDATION_ERROR`

### POST /api/login

来源: `auth_handler_test.go:179`

```json
{
  "username": "bob",
  "password": "hunter2"
}
```

### POST /api/login (错误密码)

来源: `auth_handler_test.go:233`

```json
{
  "username": "carol",
  "password": "wrongpw"
}
```

→ 返回 401

### POST /api/bonds/

来源: `bond_handler_test.go:159`

```json
{
  "amount": 5,
  "interest": 1.2
}
```

### POST /api/bonds/ (验证边界)

来源: `bond_handler_test.go:136`

```json
{
  "amount": 0,
  "interest": 0
}
```

→ 返回 400

### POST /api/v2/production/start/

来源: `production_handler_test.go:234`

```json
{
  "building_id": "bld-1",
  "resource_id": 3,
  "quantity": 10
}
```

### POST /api/v2/market-order/

来源: `market_handler_test.go:396`

```json
{
  "resourceId": 1,
  "kind": 1,
  "quality": 0,
  "quantity": 5,
  "price": 10.0
}
```

`kind`: 0=buy, 1=sell

### DELETE /api/v2/market-order/cancel/{orderId}/

来源: `market_handler_test.go`

路径参数 `orderId`, 无请求 body

### POST /api/v2/market-order/take/

来源: `market_handler_test.go:626`

```json
{
  "resource": 1,
  "quantity": 5,
  "quality": 0,
  "maxPrice": 100.0
}
```

---

## 从 OpenAPI Schema 推断的请求示例

以下 JSON 由 schema 定义自动生成，字段名/类型与 OpenAPI 一致，实际值需按游戏内情况调整。

### POST /api/register

Schema: `RegisterRequest`

```json
{
  "username": "newplayer",
  "password": "secret123",
  "name": "Alice",
  "gender": "M",
  "email": "alice@example.com"
}
```

### POST /api/login

Schema: `LoginRequest`

```json
{
  "username": "newplayer",
  "password": "secret123"
}
```

### POST /api/v2/buildings/buy/

Schema: `BuyBuildingRequest`

```json
{
  "buildingId": "bld-factory-1",
  "requestId": "req-a1b2c3d4"
}
```

`requestId` 是幂等键（可选），`buildingId` 来自建筑市场列表。

### POST /api/v2/buildings/place/

Schema: `PlaceBuildingRequest`

```json
{
  "buildingId": "bld-factory-1",
  "mapId": "map-main",
  "slotId": "A1",
  "x": 15,
  "y": 30
}
```

### POST /api/v2/buildings/move/

Schema: `MoveBuildingRequest`

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

Schema: `DemolishBuildingRequest`

```json
{
  "buildingId": "bld-uuid-abc"
}
```

### POST /api/v2/buildings/{buildingId}/stock/

Schema: `StockShelfRequest`

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "quantity": 50,
  "price": 12.5
}
```

`price` 可选，不传则沿用上次价格。

### POST /api/v2/buildings/{buildingId}/unstock/

Schema: `UnstockShelfRequest`

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "quantity": 20
}
```

### POST /api/v2/buildings/{buildingId}/shelf-price/

Schema: `SetShelfPriceRequest`

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "price": 15.0,
  "lock": true
}
```

`lock: true` 防止自动调价覆盖。

### POST /api/v2/market-order/

Schema: `CreateOrderRequestFrontend`

```json
{
  "resourceId": 1,
  "kind": 1,
  "quality": 0,
  "quantity": 5,
  "price": 10.0
}
```

### POST /api/v2/market-order/take/

Schema: `TakeOrderRequest`

```json
{
  "resource": 1,
  "quantity": 5,
  "quality": 0,
  "maxPrice": 100.0
}
```

### PATCH /api/v2/companies/me/story-progress/

Schema: `StoryProgressRequest`

```json
{
  "storyId": "tutorial_intro",
  "stepId": "step_build_factory",
  "status": "completed"
}
```

`status` 枚举值: `not_started`, `in_progress`, `completed`, `skipped`

### POST /api/bonds/

Schema: `CreateBondRequest`

```json
{
  "amount": 5,
  "interest": 1.2
}
```

### POST /api/v2/production/start/

Schema: `StartProductionRequest`

```json
{
  "building_id": "bld-factory-1",
  "resource_id": 3,
  "quantity": 10
}
```

---

## 响应示例（从 Schema 推断）

### 200 OK — 登录成功

Schema: `LoginResponse`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "player_id": 42,
  "company_id": 1,
  "username": "newplayer"
}
```

### 200 OK — 公司资料

Schema: `CompanyProfileResponse`

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

### 200 OK — 我的公司列表

Schema: `MyCompaniesResponse`

```json
{
  "companies": [
    {"id": 1, "name": "NewHaven Corp", "money": 50000.0, "level": 5}
  ]
}
```

### 200 OK — 仓库

Schema: `GetMyWarehouseData`

```json
{
  "company_id": 1,
  "capacity": 5000,
  "used_capacity": 1250,
  "items": [
    {"resource_id": 1, "resource_name": "Iron Ore", "quality": 0, "amount": 500},
    {"resource_id": 3, "resource_name": "Steel Bar", "quality": 0, "amount": 200}
  ]
}
```

### 200 OK — 建筑列表

Schema: `BuildingDTO`

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

### 200 OK — 建筑市场

Schema: `BuildingMarketItem`

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

### 200 OK — 购买建筑响应

Schema: `BuyBuildingResponse`

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

### 200 OK — 升级结果

Schema: `UpgradeBuildingResponse`

```json
{
  "building_id": "bld-uuid-abc",
  "old_level": 2,
  "new_level": 3,
  "cost": 8000.0,
  "output_multiplier": 150
}
```

### 200 OK — 市场行情

Schema: `MarketTicker`

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

### 200 OK — 订单列表

Schema: `MarketOrderDTO`

```json
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
```

### 200 OK — 市场深度

Schema: `MarketDepthResponse`

```json
{
  "buys": [
    {"price": 9.5, "quantity": 100, "qty": 100},
    {"price": 9.0, "quantity": 200, "qty": 200}
  ],
  "sells": [
    {"price": 10.5, "quantity": 150, "qty": 150},
    {"price": 11.0, "quantity": 80, "qty": 80}
  ]
}
```

### 200 OK — 资源列表

Schema: `ResourcesResponse`

```json
{
  "resources": [
    {"resourceId": 1, "name": "Iron Ore", "producedFrom": {"mine": 2}, "producedPerHourRaw": 100, "unitsSoldAnHour": 50, "hasEconomyModel": true}
  ]
}
```

### 200 OK — 货架

Schema: `ShelfItem`

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

### 400/401/403 — 错误响应

Schema: `ErrorResponse`

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

### 200 OK — 升级仓库响应

Schema: `WarehouseUpgradeResponse`

```json
{
  "level": 2,
  "capacity": 10000,
  "cost": 5000.0
}
```

### 200 OK — 拆毁建筑响应

Schema: `DemolishBuildingResponse`

```json
{
  "refund": 2500.0,
  "status": "ok"
}
```

### 200 OK — 故事进度

Schema: `StoryProgress`

```json
{
  "status": "completed",
  "stepId": "step_build_factory"
}
```

### 200 OK — 放置建筑响应

Schema: `PlaceBuildingResponse`

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

---

## 通用响应格式

所有响应都包裹在标准信封里：

```json
{"data": <实际数据>, "error": null, "meta": {"request_id": "req-xxx"}}
```

错误时：

```json
{"data": null, "error": {"code": "ERROR_CODE", "message": "描述", "details": null}}
```

错误码:

| 状态码 | 错误码 | 说明 |
|--------|--------|------|
| 401 | `UNAUTHORIZED` | token 缺失或无效 |
| 403 | `FORBIDDEN` | 无权访问 |
| 404 | `NOT_FOUND` | 资源不存在 |
| 400 | `BAD_REQUEST` | 请求格式错误 |
| 400 | `VALIDATION_ERROR` | 字段校验失败 |
| 400 | `INSUFFICIENT_FUNDS` | 余额不足 |
| 400 | `INSUFFICIENT_INVENTORY` | 库存不足 |
| 409 | `CONFLICT` | 冲突（如用户名已注册） |
| 429 | `RATE_LIMITED` | 频率限制 |
| 500 | `INTERNAL_ERROR` | 服务器内部错误 |

---

## 认证方式

除 `/healthz`, `/readyz`, `/api/register`, `/api/login`, `/api/admin/snapshot/*` 外，所有接口都需要:

```
Authorization: Bearer <token>
```

Token 从 `POST /api/login` 的响应 `data.token` 获取。

---

## 三个文件的使用策略

| 文件 | 用途 | 搜索方式 |
|------|------|---------|
| `API-QUICKREF.md` | 查哪个接口对应哪个 handler | Ctrl+F 搜路径/模块名 |
| `API-EXAMPLES.md`（这个文件） | 复制粘贴 JSON 调接口 | Ctrl+F 搜路径名 |
| `openapi-draft.yaml` | 查字段定义、约束、类型 | `yq` 或 Ctrl+F |
