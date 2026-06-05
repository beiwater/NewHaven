# Phase 7: Production Read Migration

**版本**: 1.0  
**日期**: 2026-06-06  
**范围**: backend-next 仅  
**基于**: Phase 6 Building Read

---

## 1. 目标

迁移生产队列读取端点 `GET /api/v2/production/jobs/`，读取当前 company 的所有生产任务。

本阶段只做读，不做 start / claim / cancel。

---

## 2. Execution Context Pack

### 2.1 风格规则

1. Handler 只能做：CompanyIDFromCtx → 调 service → 用 generated DTO → writeSuccess。
2. 禁止 `map[string]any`，用 `openapi.ProductionJobDTO`。
3. 不手改 generated 文件。
4. 所有跨边界函数第一参数 `context.Context`。
5. Storage interface 加方法写 exact signature。
6. Router 注册照 `r.With(AuthRequired).Get("/path", handler.Fn)`。
7. 测试包含: 401、200、空数组不 null、回归旧端点。

### 2.2 OpenAPI Spec 新增

```yaml
  /api/v2/production/jobs/:
    get:
      tags: [Production]
      operationId: listProductionJobs
      security:
        - BearerAuth: []
      responses:
        "200":
          description: Production job list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProductionJobListResponse'
        "401":
          $ref: '#/components/responses/UnauthorizedError'

components:
  schemas:
    ProductionJobDTO:
      type: object
      properties:
        id:
          type: string
        resource_id:
          type: integer
        quantity:
          type: integer
        target_quantity:
          type: integer
        started_at:
          type: string
          format: date-time
        duration_seconds:
          type: number
        status:
          type: string
          enum: [running, ready, claimed]
    ProductionJobListData:
      type: object
      properties:
        jobs:
          type: array
          items:
            $ref: '#/components/schemas/ProductionJobDTO'
    ProductionJobListResponse:
      type: object
      properties:
        data:
          $ref: '#/components/schemas/ProductionJobListData'
        error:
          type: 'null'
```

### 2.3 Storage Interface

```go
// In ProductionStorage:
type ProductionStorage interface {
    GetJobsByCompany(ctx context.Context, companyID int) ([]production.ProductionJob, error)
    GetJob(ctx context.Context, jobID string) (*production.ProductionJob, error)
}
```

已有 `internal/domain/production/types.go` 包含 `ProductionJob` 类型。

### 2.4 Handler Pattern

参照 building_handler.go，创建 `internal/httpapi/production_handler.go`。

### 2.5 Router Registration

加到现有 auth group:
```go
r.Get("/api/v2/production/jobs/", productionHandler.ListProductionJobs)
```

---

## 3. 范围内

1. 更新 OpenAPI spec
2. 重新生成 oapi-codegen
3. 新建 `internal/app/production/service.go`
4. 新建 `internal/httpapi/production_handler.go`
5. 扩展 `internal/storage/interfaces.go` + `memory/memory.go`
6. Service test + Handler test
7. 注册路由
8. 确认旧端点回归

## 4. 范围外

❌ 不实现 start production / claim / cancel  
❌ 不搬 market / finance  
❌ 不接 PostgreSQL  
❌ 不改旧 backend / 前端 / 经济数据  

## 5. 验收标准

1. `GET /api/v2/production/jobs/` 返回 200 + jobs 列表
2. 未登录返回 401
3. 空队列返回 `[]` 不是 `null`
4. `go test ./...` `go vet ./...` `go build` 全部通过
5. auth/company/warehouse/building 端点仍然工作
