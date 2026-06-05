# Phase 5: Warehouse Read Migration

**版本**: 1.0  
**日期**: 2026-06-06  
**范围**: backend-next 仅  
**基于**: Phase 4 Company Read

---

## 1. 目标

迁移低风险仓库读取端点 `GET /api/v2/companies/me/warehouse/`，验证 company-owned resource state 的读取链路。

Phase 5 完成后 backend-next 能完成：
```
auth token → player id → current company → warehouse items → generated DTO → uniform response
```

---

## 2. 当前状态

可用端点：
- `GET /healthz` `/GET /readyz`
- `POST /api/register` `POST /api/login`
- `GET /api/v2/players/me/companies/`

迁移流水线已验证：
```
auth middleware → ctx player id → app service → storage interface → generated OpenAPI DTO → uniform response → handler tests
```

---

## 3. Target Endpoint

```
GET /api/v2/companies/me/warehouse/
```

Response 200:
```json
{
  "data": {
    "company_id": 1000002,
    "capacity": 1000,
    "used_capacity": 0,
    "items": []
  },
  "error": null
}
```

DTO:
- WarehouseItemDTO: resource_id, resource_name, quality, amount
- GetMyWarehouseData: company_id, capacity, used_capacity, items[]
- GetMyWarehouseResponse: data + error + meta

---

## 4. 范围内

1. 更新 OpenAPI spec → 新增 warehouse endpoint + DTO
2. 重新运行 oapi-codegen
3. 新增 `internal/app/warehouse/service.go`
4. 新增 `internal/httpapi/warehouse_handler.go`
5. 扩展 memory storage 支持 warehouse read
6. 新增 service tests + handler tests
7. 确认 auth/company 端点不变

---

## 5. 范围外

❌ 不实现仓库升级/写入/库存扣减  
❌ 不搬 production / market / finance / bonds  
❌ 不接 PostgreSQL  
❌ 不改旧 backend / 前端 / 经济数据  
❌ 不手改 generated 文件  
❌ 不新增 `map[string]any` response  

---

## 6. 验收标准

1. OpenAPI spec 含 warehouse endpoint
2. oapi-codegen 重新生成成功
3. 未登录返回 401
4. 登录后返回 200 + warehouse data
5. 空仓库返回 `[]` 不是 `null`
6. auth/login/company 行为不变
7. `go test ./...` `go vet ./...` `go build` 全部通过

---

## 7. 执行指令

你现在执行 Phase 5：Warehouse Read Migration。

必须先阅读：
- `docs/2026-06-06/rebuild/00-constitution.md`
- `docs/2026-06-06/rebuild/12-phase-2-auth-migration.md`
- `docs/2026-06-06/rebuild/14-phase-4-company-read-migration.md`
- `backend-next/openapi/openapi-draft.yaml`
- `backend-next/internal/generated/openapi/types.gen.go`
- `backend-next/internal/httpapi/auth_middleware.go`
- `backend-next/internal/httpapi/company_handler.go`
- `backend-next/internal/storage/memory/memory.go`

任务：
1. 在 OpenAPI spec 中定义 warehouse endpoint 和 DTO
2. 运行 `scripts/generate-openapi.sh` 重新生成
3. 创建 `internal/app/warehouse/service.go`
4. 创建 `internal/httpapi/warehouse_handler.go`
5. 扩展 memory storage 的 warehouse read
6. 写 service_test.go 和 warehouse_handler_test.go
7. 注册路由到 router.go + main.go
8. `go test ./...` `go vet ./...` `go build`
9. 确认 auth、login、company 端点不变

硬限制：不改旧 backend、不改前端、不接数据库、不搬生产/市场/金融、不手改 generated、不新增 map[string]any。
