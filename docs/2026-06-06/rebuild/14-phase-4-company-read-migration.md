# Phase 4: Company Read Migration

**版本**: 1.0  
**日期**: 2026-06-06  
**范围**: backend-next 仅  
**基于**: Phase 2 Auth / Phase 3 OpenAPI Codegen

---

## 1. 目标

把低风险 company 读端点 `GET /api/v2/players/me/companies/` 迁移到 `backend-next`，使用 Phase 3 已跑通的 OpenAPI generated types 做 contract-first 开发。

本阶段只做读端点，不做写操作。不碰 market、production、finance、bonds、scheduler。

---

## 2. 当前依赖

- ✅ Phase 1: docs/rebuild/*
- ✅ Phase 2: auth/register/login 可运行
- ✅ Phase 3: OpenAPI codegen 可运行
- ✅ `internal/generated/openapi/types.gen.go`
- ✅ `internal/httpapi/auth_middleware.go`
- ✅ `internal/storage/memory/memory.go`

---

## 3. 范围内

1. 更新 `openapi/openapi-draft.yaml` 定义 company 端点
2. 重新运行 oapi-codegen
3. 使用 generated DTO（不手写新 DTO）
4. 新增 company app service
5. 新增 company HTTP handler
6. 扩展 memory storage 的 company read 方法
7. 新增 service tests + handler tests
8. 保持 auth/register/login 不变

---

## 4. 范围外

❌ 不搬 market / production / finance / bonds / scheduler  
❌ 不接 PostgreSQL  
❌ 不改旧 backend  
❌ 不改前端  
❌ 不改 JWT 签发  
❌ 不改经济公式和数据  
❌ 不新增 `map[string]any` response  
❌ 不手改 generated 文件  

---

## 5. API Contract

```
GET /api/v2/players/me/companies/
```

Auth: 必需 (JWT Bearer)  
Response 200:
```json
{
  "data": {
    "companies": [
      {"id": 1000002, "name": "Dev Player", "money": 100000, "level": 1}
    ]
  },
  "error": null,
  "meta": {"request_id": "..."}
}
```

Response 401:
```json
{
  "data": null,
  "error": {"code": "UNAUTHORIZED", "message": "missing authorization token"}
}
```

---

## 6. 验收标准

1. OpenAPI spec 包含 company 端点
2. oapi-codegen 重新生成成功
3. company handler 返回 401 当未登录
4. company handler 返回 200 + company list 当已登录
5. login/register/healthz/readyz 行为不变
6. `go test ./...` 全部通过
7. `go vet ./...` 全部通过
8. 不改旧 backend、前端、经济数据

---

## 7. 执行指令

你现在执行 Phase 4：Company Read Migration。

必须先阅读：
- `docs/2026-06-06/rebuild/00-constitution.md`
- `docs/2026-06-06/rebuild/12-phase-2-auth-migration.md`
- `docs/2026-06-06/rebuild/13-phase-3-openapi-codegen.md`
- `backend-next/openapi/openapi-draft.yaml`
- `backend-next/internal/generated/openapi/types.gen.go`
- `backend-next/internal/httpapi/auth_middleware.go`
- `backend-next/internal/storage/memory/memory.go`

任务：
1. 在 OpenAPI spec 中定义 `GET /api/v2/players/me/companies/`
2. 运行 `scripts/generate-openapi.sh` 重新生成
3. 创建 `internal/app/company/service.go` — 从 storage 获取公司列表
4. 创建 `internal/httpapi/company_handler.go` — 用 generated DTO 返回
5. 扩展 `internal/storage/memory/memory.go` 支持 `GetCompanyByPlayerID`
6. 写 service_test.go 和 company_handler_test.go
7. 注册路由到 router.go
8. 运行 `go test ./...` `go vet ./...`
9. 冒烟测试确认 `/api/register` `/api/login` `/healthz` 不变

硬限制：不改旧 backend、不改前端、不接数据库、不手改 generated、不新增 map[string]any。
