# 仓库

本文件覆盖公司仓库查询与升级。

## GET /E1/companies/me/warehouse/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `GetMyWarehouseData`

### 响应示例

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

## POST /E1/companies/me/warehouse/upgrade/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `WarehouseUpgradeResponse`

### 响应示例

```json
{
  "level": 2,
  "capacity": 10000,
  "cost": 5000
}
```

### 常见错误

- `400 INSUFFICIENT_FUNDS`: 仓库升级费用不足
- `400 VALIDATION_ERROR`: 当前等级或条件不满足

