# Phase 6: Building Read Migration

**版本**: 1.0  
**日期**: 2026-06-06  
**范围**: backend-next 仅  
**基于**: Phase 5 Warehouse Read

---

## 1. 目标

迁移建筑读取端点 `GET /api/v3/companies/me/buildings/`，读取当前 company 的所有已放置建筑。

本阶段只做读，不做 placement / upgrade / demolish。

---

## 2. Execution Context Pack

### 2.1 本任务必须遵守的风格规则

1. Handler 只能做：CompanyIDFromCtx → 调 service → 用 generated DTO → writeSuccess。不写业务逻辑。
2. 禁止 `map[string]any`，必须用 `openapi.BuildingDTO` / `openapi.BuildingListResponse`。
3. 不手改 `internal/generated/openapi/types.gen.go`。
4. 所有跨边界函数第一个参数是 `context.Context`。
5. Storage interface 新增方法写 exact Go signature。
6. Router 注册照已有 pattern：`r.With(AuthRequired(cfg.JWTSigningKey)).Get("/path", handler.Fn)`。
7. 测试包含: 401、200、空数组不 null、回归 auth/company/warehouse。

### 2.2 Exact Generated OpenAPI Type Names

先在 `openapi/openapi-draft.yaml` 新增：

```yaml
  /api/v3/companies/me/buildings/:
    get:
      tags: [Buildings]
      operationId: listMyBuildings
      security:
        - BearerAuth: []
      responses:
        "200":
          description: Building list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BuildingListResponse'
        "401":
          $ref: '#/components/responses/UnauthorizedError'

components:
  schemas:
    BuildingDTO:
      type: object
      properties:
        id:
          type: string
        building_id:
          type: integer
        name:
          type: string
        level:
          type: integer
        map_id:
          type: string
        slot_id:
          type: string
        x:
          type: integer
        y:
          type: integer
    BuildingListData:
      type: object
      properties:
        buildings:
          type: array
          items:
            $ref: '#/components/schemas/BuildingDTO'
    BuildingListResponse:
      type: object
      properties:
        data:
          $ref: '#/components/schemas/BuildingListData'
        error:
          type: 'null'
```

Regenerate 后预期的 Go 类型名（以实际生成输出为准，grep types.gen.go 确认）：
```go
openapi.BuildingDTO        // single building
openapi.BuildingListData   // wrapper with buildings []
openapi.BuildingListResponse  // envelope
```

⚠️ generated 类型用 `*[]openapi.BuildingDTO` + `*int`，赋值时注意 `&`。

### 2.3 Exact Interface Signatures

Storage (在 `internal/storage/interfaces.go` 的 CompanyStorage 上加):
```go
type CompanyStorage interface {
    // ... existing methods ...
    GetBuildings(ctx context.Context, companyID int) ([]company.Building, error)
}
```

Memory (在 `internal/storage/memory/memory.go` 实现):
```go
func (s *Store) GetBuildings(ctx context.Context, companyID int) ([]company.Building, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // return company's buildings
}
```

### 2.4 Exact Router Registration Pattern

```go
// In router.go, add to the existing auth group:
r.Group(func(r chi.Router) {
    r.Use(httpapi.AuthRequired(cfg.JWTSigningKey))
    r.Get("/api/v2/players/me/companies/", companyHandler.ListMyCompanies)
    r.Get("/api/v2/companies/me/warehouse/", warehouseHandler.GetMyWarehouse)
    r.Get("/api/v3/companies/me/buildings/", buildingHandler.ListMyBuildings)  // ← add this
})
```

### 2.5 Exact Handler Pattern

参照 warehouse_handler.go:
```go
type BuildingHandler struct {
    svc *building.Service
}

func NewBuildingHandler(svc *building.Service) *BuildingHandler {
    return &BuildingHandler{svc: svc}
}

func (h *BuildingHandler) ListMyBuildings(w http.ResponseWriter, r *http.Request) {
    companyID, ok := httpapi.CompanyIDFromCtx(r.Context())
    if !ok {
        httpapi.WriteErr(w, 401, httpapi.ErrorUnauthorized, "company not found in context", nil)
        return
    }
    resp, err := h.svc.ListMyBuildings(r.Context(), companyID)
    if err != nil {
        httpapi.WriteErr(w, 500, httpapi.ErrorInternal, "failed to list buildings", nil)
        return
    }
    httpapi.WriteSuccess(w, 200, resp)
}
```

### 2.6 Test Patterns

**Service test** (`internal/app/building/service_test.go`):
```go
func TestListMyBuildings_WithToken_ReturnsBuildings(t *testing.T) {
    // setup: memory store, register dev, get company
    // call service
    // verify buildings list not nil
}
```

**Handler test** (`internal/httpapi/building_handler_test.go`):
```go
func TestListMyBuildings_NoToken_401(t *testing.T) {
    // httptest, no auth header → 401 + UNAUTHORIZED
}
func TestListMyBuildings_WithToken_200(t *testing.T) {
    // register first, login to get token, request with auth → 200
}
func TestListMyBuildings_ReturnsEmptyArrayWhenNoBuildings(t *testing.T) {
    // dev company has no buildings → data.buildings is [] not null
}
```

---

## 3. 范围内

1. 更新 `openapi/openapi-draft.yaml`（BuildingDTO + endpoint）
2. `scripts/generate-openapi.sh` 重新生成
3. `internal/domain/building/types.go` 已有（确认存在，扩展如果需要）
4. `internal/app/building/service.go` — 新增
5. `internal/httpapi/building_handler.go` — 新增
6. `internal/storage/interfaces.go` — 加 GetBuildings
7. `internal/storage/memory/memory.go` — 实现 GetBuildings
8. `internal/app/building/service_test.go` + `internal/httpapi/building_handler_test.go`
9. 注册路由
10. 确认 auth/company/warehouse 回归

## 4. 范围外

❌ 不搬 building placement / upgrade / demolish  
❌ 不搬 production / market / finance  
❌ 不接 PostgreSQL  
❌ 不改旧 backend / 前端 / 经济数据  
❌ 不手改 generated 文件  

## 5. 验收标准

1. `GET /api/v3/companies/me/buildings/` 返回 200
2. 未登录返回 401
3. 空建筑返回 `[]` 不是 `null`
4. `go test ./...` `go vet ./...` `go build` 全部通过
5. auth/company/warehouse 端点仍然工作
6. 不改旧 backend、前端、经济数据
