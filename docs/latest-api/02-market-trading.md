# 市场 / 交易

本文件覆盖资源查询、行情、深度、订单列表以及下单、撤单、吃单等接口。

## GET /V1/resources/

认证: 无需
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `ResourcesResponse`

### 响应示例

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

## GET /V1/market-ticker/{resourceId}/

认证: 无需
路径参数: `resourceId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `MarketTickerResponse`

### 响应示例

```json
{
  "resource": 1,
  "series": [
    {
      "price": 12.5,
      "time": "2026-06-12T10:30:00Z"
    }
  ]
}
```

## GET /V1/market-depth/{resourceId}/{quality}/

认证: 无需
路径参数: `resourceId`, `quality`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `MarketDepthResponse`

### 响应示例

```json
{
  "buys": [
    { "price": 9.5, "quantity": 100, "qty": 100 },
    { "price": 9.0, "quantity": 200, "qty": 200 }
  ],
  "sells": [
    { "price": 10.5, "quantity": 150, "qty": 150 },
    { "price": 11.0, "quantity": 80, "qty": 80 }
  ]
}
```

## GET /V1/market/{resourceId}/{quality}/

认证: 无需
路径参数: `resourceId`, `quality`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `MarketOrderListResponse`

### 响应示例

```json
{
  "orders": [
    {
      "id": "ord-uuid-xyz",
      "resourceId": 1,
      "kind": 0,
      "price": 10,
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

## POST /E1/market-order/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `CreateOrderRequestFrontend`

### 请求示例

```json
{
  "resourceId": 1,
  "kind": 1,
  "quality": 0,
  "quantity": 5,
  "price": 10
}
```

### 响应示例

```json
{
  "order": {
    "id": "ord-uuid-xyz",
    "resourceId": 1,
    "kind": 1,
    "price": 10,
    "quality": 0,
    "quantity": 5,
    "remaining": 5,
    "companyId": 1,
    "createdAt": "2026-06-12T08:00:00Z",
    "status": "open"
  }
}
```

### 字段说明

- `kind`: `0` = buy，`1` = sell
- `requestId`: 可选幂等键
- 该接口使用 `camelCase` 字段名

## DELETE /E1/market-order/cancel/{orderId}/

认证: 需要
路径参数: `orderId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `CancelOrderResponse`

### 响应示例

```json
{
  "id": "ord-uuid-xyz",
  "status": "cancelled"
}
```

## POST /E1/market-order/take/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `TakeOrderRequest`

### 请求示例

```json
{
  "resource": 1,
  "quantity": 5,
  "quality": 0,
  "maxPrice": 100
}
```

### 响应示例

```json
{
  "amountBought": 5,
  "trades": [
    {
      "id": "trade-uuid",
      "resourceId": 1,
      "quality": 0,
      "quantity": 5,
      "price": 10,
      "buyOrderId": "ord-buy",
      "sellOrderId": "ord-sell",
      "createdAt": "2026-06-12T08:00:00Z"
    }
  ],
  "moneyDelta": -50
}
```

### 常见错误

- `400 VALIDATION_ERROR`: 数量、价格或质量不合法
- `400 INSUFFICIENT_FUNDS`: 余额不足
- `400 INSUFFICIENT_INVENTORY`: 卖单库存不足

