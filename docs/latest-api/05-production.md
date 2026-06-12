# 生产

本文件覆盖生产任务列表、开始生产、领取产出、队列和取消等接口。

## GET /E1/production/jobs/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `ProductionJobListResponse`

### 响应示例

```json
{
  "jobs": [
    {
      "id": "job-uuid",
      "building_id": "bld-1",
      "resource_id": 3,
      "quantity": 10,
      "target_quantity": 10,
      "started_at": "2026-06-12T08:00:00Z",
      "duration_seconds": 3600,
      "claimed_amount": 0,
      "claimable_amount": 0,
      "status": "running"
    }
  ]
}
```

## POST /E1/production/start/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `StartProductionRequest`

### 请求示例

```json
{
  "building_id": "bld-1",
  "resource_id": 3,
  "quantity": 10
}
```

### 响应示例

```json
{
  "job": {
    "id": "job-uuid",
    "building_id": "bld-1",
    "resource_id": 3,
    "quantity": 10,
    "target_quantity": 10,
    "started_at": "2026-06-12T08:00:00Z",
    "duration_seconds": 3600,
    "claimed_amount": 0,
    "claimable_amount": 0,
    "status": "running"
  },
  "building": {
    "id": "bld-1",
    "busy": true,
    "job_id": "job-uuid"
  }
}
```

## POST /E1/production/claim/{jobId}/

认证: 需要
路径参数: `jobId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `ClaimProductionResponse`

### 响应示例

```json
{
  "job_id": "job-uuid",
  "status": "claimed",
  "output": {
    "3": 10
  },
  "claimed_amount": 10,
  "remaining": 0,
  "xp": 25,
  "level": 6,
  "market_unlocked": true
}
```

## GET /E1/production/claimable/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `ClaimableJobListResponse`

### 响应示例

```json
{
  "jobs": [
    {
      "job_id": "job-uuid",
      "building_id": "bld-1",
      "resource_id": 3,
      "total_amount": 10,
      "claimed_amount": 0,
      "claimable_amount": 10
    }
  ]
}
```

## GET /E1/production/queue/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `ProductionQueueResponse`

### 响应示例

```json
{
  "byBuilding": {
    "bld-1": [
      {
        "id": "job-uuid",
        "building_id": "bld-1",
        "resource_id": 3,
        "quantity": 10,
        "target_quantity": 10,
        "started_at": "2026-06-12T08:00:00Z",
        "duration_seconds": 3600,
        "claimed_amount": 0,
        "claimable_amount": 0,
        "status": "running"
      }
    ]
  },
  "inUse": 1,
  "maxSlots": 3
}
```

## POST /E1/production/cancel/

认证: 需要
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `CancelJobRequest`

### 请求示例

```json
{
  "jobId": "job-uuid"
}
```

### 响应示例

```json
{
  "jobId": "job-uuid",
  "status": "cancelled"
}
```

## POST /E1/production/claim-all/

认证: 需要
请求: 无 body

### 响应示例

```json
{
  "claimed": [],
  "errors": [],
  "total": 0
}
```

## GET /E1/buildings/{buildingId}/production-options/

认证: 需要
路径参数: `buildingId`
请求: 无 body

### 响应说明

返回建筑可生产资源列表，结构以 [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `ResourceDefinition` 为准。

## POST /E1/buildings/{buildingId}/busy/

认证: 需要
路径参数: `buildingId`
来源: [`heritage/openapi-draft.yaml`](../../heritage/openapi-draft.yaml) 中 `StartProductionV1Request`

### 请求示例

```json
{
  "kind": 3,
  "amount": 10,
  "estimatedSecondsToFinish": 3600
}
```

### 说明

- 这是旧版兼容路由
- 如果新前端可控，优先使用 [`POST /E1/production/start/`](./05-production.md)

