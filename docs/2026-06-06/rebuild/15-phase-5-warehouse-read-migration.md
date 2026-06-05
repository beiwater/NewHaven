# Phase 5: Warehouse Read Migration

**版本**: 1.1 (内联执行上下文版)  
**日期**: 2026-06-06  
**范围**: backend-next 仅  
**基于**: Phase 4 Company Read

---

## 1. 目标

迁移 `GET /api/v2/companies/me/warehouse/`，验证 company-owned resource state 读取链路。

---

## 2. Execution Context Pack（子 Agent 直接使用本节的 exact 信息）

### 2.1 本任务必须遵守的风格规则

1. Handler 只能做：读参数（CompanyIDFromCtx）→ 调 app service → 用 generated DTO → writeSuccess。
2. 禁止 `map[string]any`。所有 response 用 `openapi.GetMyWarehouseData` 和 `openapi.WarehouseItem`。
3. 不手改 `internal/generated/openapi/types.gen.go`。改 OpenAPI spec → 重新生成。
4. 所有跨边界函数第一个参数是 `context.Context`。
5. Storage interface 新增方法必须写 exact Go signature。
6. Router 注册必须照 §2.3 的 exact pattern。
7. 测试必须包含：401、200、空数组不 null、回归旧端点。

### 2.2 Related Existing Code 模板

**company_handler.go handler pattern:**
```go
func (h *CompanyHandler) ListMyCompanies(w http.ResponseWriter, r *http.Request) {
	companyID, ok := httpapi.CompanyIDFromCtx(r.Context())
	if !ok {
		httpapi.WriteErr(w, 401, httpapi.ErrorUnauthorized, "company not found in context", nil)
		return
	}
	resp, err := h.svc.ListMyCompanies(r.Context(), companyID)
	if err != nil {
		httpapi.WriteErr(w, 500, httpapi.ErrorInternal, "failed to list companies", nil)
		return
	}
	httpapi.WriteSuccess(w, 200, resp)
}
```

**company app service pattern:**
```go
type Service struct {
	companies storage.CompanyStorage
	logger    *platform.Logger
}

func NewService(companies storage.CompanyStorage, logger *platform.Logger) *Service {
	return &Service{companies: companies, logger: logger}
}

func (s *Service) ListMyCompanies(ctx context.Context, companyID int) (*openapi.ListMyCompaniesResponse, error) {
	c, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return &openapi.ListMyCompaniesResponse{
		Companies: &[]openapi.CompanySummary{{
			Id:    &c.ID,
			Name:  &c.Name,
			Money: &c.Money,
			Level: &c.Level,
		}},
	}, nil
}
```

**router registration pattern:**
```go
// In router.go NewRouter(), after auth routes:
r.Group(func(r chi.Router) {
	r.Use(httpapi.AuthRequired(cfg.JWTSigningKey))
	r.Get("/api/v2/players/me/companies/", companyHandler.ListMyCompanies)
	r.Get("/api/v2/companies/me/warehouse/", warehouseHandler.GetMyWarehouse)
})
```

### 2.3 Exact Interface Signatures

Auth middleware（`internal/httpapi/auth_middleware.go`）:
```go
func AuthRequired(jwtKey string) func(http.Handler) http.Handler
func PlayerIDFromCtx(ctx context.Context) (int, bool)
func CompanyIDFromCtx(ctx context.Context) (int, bool)
```

Storage interface 需要新增（`internal/storage/interfaces.go`）:
```go
type WarehouseStorage interface {
	GetWarehouse(ctx context.Context, companyID int) (*warehouse.Warehouse, error)
}
```
或者在 `CompanyStorage` 中加方法，不要新建超大 interface。

Response helpers（`internal/httpapi/response.go`）:
```go
func WriteSuccess(w http.ResponseWriter, status int, data any)
func WriteErr(w http.ResponseWriter, status int, code, message string, details any)
```

### 2.4 Exact Generated OpenAPI Type Names

从 `openapi-draft.yaml` 新增后 regenerate，types.gen.go 中会生成（名字以实际输出为准，以下为预期）:
```go
openapi.GetMyWarehouseData struct {
	CompanyId   *int            `json:"company_id,omitempty"`
	Capacity    *int            `json:"capacity,omitempty"`
	UsedCapacity *int           `json:"used_capacity,omitempty"`
	Items       *[]openapi.WarehouseItem `json:"items,omitempty"`
}
openapi.WarehouseItem struct {
	ResourceId   *int    `json:"resource_id,omitempty"`
	ResourceName *string `json:"resource_name,omitempty"`
	Quality      *int    `json:"quality,omitempty"`
	Amount       *int    `json:"amount,omitempty"`
}
```

⚠️ handler 返回时赋值给 `&data.Items`，不是 `data.Items = []`（因为 generated 用 `*[]`）。

### 2.5 Exact Router Registration Pattern

```go
// In router.go, inside NewRouter, add after company routes:
r.Group(func(r chi.Router) {
	r.Use(httpapi.AuthRequired(cfg.JWTSigningKey))
	r.Get("/api/v2/companies/me/warehouse/", warehouseHandler.GetMyWarehouse)
})

// In main.go:
wh := httpapi.NewWarehouseHandler(app.WarehouseService)
// pass wh to NewRouter, or register directly
```

### 2.6 Exact Test Fixture Pattern

**Service test** (create `internal/app/warehouse/service_test.go`):
```go
func TestGetMyWarehouse_ReturnsWarehouse(t *testing.T) {
	st := memory.New()
	// register dev user first
	// then call service
	svc := NewService(st, platform.NewLogger(slog.Default()))
	resp, err := svc.GetMyWarehouse(context.Background(), 1000002)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1000002, *resp.CompanyId)
	assert.NotNil(t, resp.Items) // must be empty array, not nil
}
```

**Handler test** (create `internal/httpapi/warehouse_handler_test.go`):
```go
func TestGetMyWarehouse_NoToken_401(t *testing.T) {
	// setup: app.New(cfg, st) → NewWarehouseHandler → register route on chi test router
	// httptest.NewServer or httptest.NewRecorder
	// verify 401 + error.code == "UNAUTHORIZED"
}
```

---

## 3. 范围内

1. 更新 OpenAPI spec → 新增 warehouse endpoint + schema
2. `scripts/generate-openapi.sh` 重新生成
3. 新增 `internal/domain/warehouse/types.go`
4. 新增 `internal/app/warehouse/service.go`（用 openapi types）
5. 新增 `internal/httpapi/warehouse_handler.go`
6. 扩展 `internal/storage/memory/memory.go` + `interfaces.go`
7. 新增 service test + handler test
8. 注册路由到 `router.go` + `main.go`
9. 确认 auth/company 端点不变

---

## 4. 范围外

❌ 不实现仓库升级/写入/库存扣减  
❌ 不搬 production / market / finance / bonds  
❌ 不接 PostgreSQL  
❌ 不改旧 backend / 前端 / 经济数据  
❌ 不手改 generated 文件  
❌ 不新增 `map[string]any` response  

---

## 5. 验收标准

1. OpenAPI spec 含 warehouse endpoint
2. oapi-codegen 重新生成成功
3. 未登录返回 401 + `UNAUTHORIZED`
4. 登录后返回 200 + warehouse data + `items` 是 `[]` 不是 `null`
5. auth/login/company 行为不变
6. `go test ./...` `go vet ./...` `go build` 全部通过
7. 不改旧 backend、前端、经济数据

---

## 6. 回滚方案

1. 回退 `openapi/openapi-draft.yaml`
2. 删除新增 handler / service / domain
3. 回退 memory storage 扩展
4. 重新运行 oapi-codegen
5. `go test ./...` 确认 auth 和 company 仍可用
