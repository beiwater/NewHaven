# 建筑 / 货架

本文件覆盖建筑购买、放置、移动、拆除、升级以及货架相关操作。

## GET /E1/companies/me/buildings/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `BuildingListResponse`

### 响应示例

```json
{
  "buildings": [
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
  ]
}
```

## GET /V1/buildings/market/

认证: 无需
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `BuildingMarketItem`

### 响应示例

```json
[
  {
    "id": "bld-factory-1",
    "name": "Steel Factory",
    "kind": 1,
    "cost": 5000,
    "unlock_level": 1,
    "description": "Produces steel bars from iron ore",
    "produces": [3],
    "starter_produces": [1, 3],
    "starter_role": "industrialist",
    "is_retail": false
  }
]
```

## POST /E1/buildings/buy/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `BuyBuildingRequest`

### 请求示例

```json
{
  "buildingId": "bld-factory-1",
  "requestId": "req-a1b2c3d4"
}
```

### 响应示例

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
  "cost": 5000
}
```

## POST /E1/buildings/place/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `PlaceBuildingRequest`

### 请求示例

```json
{
  "buildingId": "bld-factory-1",
  "mapId": "map-main",
  "slotId": "A1",
  "x": 15,
  "y": 30
}
```

## POST /E1/buildings/move/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `MoveBuildingRequest`

### 请求示例

```json
{
  "buildingId": "bld-uuid-abc",
  "mapId": "map-main",
  "slotId": "B2",
  "x": 20,
  "y": 45
}
```

## POST /E1/buildings/demolish/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `DemolishBuildingRequest`

### 请求示例

```json
{
  "buildingId": "bld-uuid-abc"
}
```

### 响应示例

```json
{
  "refund": 2500,
  "status": "ok"
}
```

## POST /E1/buildings/{buildingId}/upgrade/

认证: 需要
路径参数: `buildingId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `UpgradeBuildingResponse`

### 响应示例

```json
{
  "building_id": "bld-uuid-abc",
  "old_level": 2,
  "new_level": 3,
  "cost": 8000,
  "output_multiplier": 150
}
```

## POST /E1/buildings/{buildingId}/stock/

认证: 需要
路径参数: `buildingId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `StockShelfRequest`

### 请求示例

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "quantity": 50,
  "price": 12.5
}
```

## POST /E1/buildings/{buildingId}/unstock/

认证: 需要
路径参数: `buildingId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `UnstockShelfRequest`

### 请求示例

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "quantity": 20
}
```

## POST /E1/buildings/{buildingId}/shelf-price/

认证: 需要
路径参数: `buildingId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `SetShelfPriceRequest`

### 请求示例

```json
{
  "building_id": "bld-shop-1",
  "resource_id": 5,
  "price": 15,
  "lock": true
}
```

### 货架响应示例

```json
{
  "resource_id": 5,
  "quantity": 50,
  "max_qty": 200,
  "price": 12.5,
  "price_lock": false,
  "revenue": 625
}
```

