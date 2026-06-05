# Phase 2：Auth 域迁移试点

**版本**: 1.0  
**基于**: Phase 1 文档 (00-constitution.md, 05-backend-target-architecture.md, 09-first-pr-plan.md)  
**状态**: 未开始 → 本文档为施工单

---

## 1. 目标

将旧 `backend/` 中的用户注册和登录功能迁移到 `backend-next/`，验证新骨架（chi router、统一 response envelope、domain 层、app use case 层、storage 接口、内存实现）是否跑通。

迁移完成后：
- `backend-next` 可以独立处理 `/api/register` 和 `/api/login`
- JWT token 签发格式与旧 backend 一致（`middleware.SignJWT` 的输出）
- 旧 backend 仍然提供相同的路由，互不干扰
- 前端不需任何改动（因为路径不变）

---

## 2. 边界

### 2.1 迁移范围 ✅

| 功能 | 说明 |
|------|------|
| `POST /api/register` | 用户注册 → 创建 Player + Company → 返回 JWT |
| `POST /api/login` | 用户登录 → 验证密码 → 返回 JWT |
| JWT 签发 | 使用与旧 backend 兼容的 HS256 格式 |
| bcrypt 密码哈希 | 使用 `golang.org/x/crypto/bcrypt` |
| Dev 模式自动创建 | 如果 `SIM_API_DEV_MODE=true`，启动时创建 dev 用户 |
| Auth middleware | `Bearer <token>` → 解析出 player_id + company_id 注入 context |

### 2.2 绝对不碰 ❌

| 模块 | 原因 |
|------|------|
| `backend/internal/service/` (旧) | 不动旧业务逻辑 |
| `backend/internal/handler/auth.go` (旧) | 旧路由继续提供，前端可能还在用 |
| `backend/internal/middleware/auth.go` (旧) | JWT 验证逻辑不动旧的 |
| `backend/internal/model/types.go` (旧) | domain model 不动旧的 |
| `backend-next/internal/domain/market/...` | 不是 auth 范围 |
| `backend-next/internal/domain/production/...` | 不是 auth 范围 |
| `backend-next/internal/domain/finance/...` | 不是 auth 范围 |
| `backend-next/internal/domain/research/...` | 不是 auth 范围 |
| `backend-next/internal/domain/social/...` | 不是 auth 范围 |
| `backend-next/internal/storage/postgres/...` | 暂不接 PostgreSQL |
| `client/` (前端) | 不动前端一行代码 |
| `docs/` | 不修改已有文档 |

### 2.3 暂时不做的（但会在 auth 跑通后做）

- OpenAPI codegen（Phase 3）
- PostgreSQL migration（Phase 4）
- 移除旧 auth handler（至少等前端确认切换后）

---

## 3. 文件清单

### 3.1 新建文件

| 文件 | 职责 |
|------|------|
| `backend-next/internal/app/auth/service.go` | Auth use case（Register、Login、DevBootstrap） |
| `backend-next/internal/app/auth/jwt.go` | JWT 签发（与旧 backend 兼容的 HS256） |
| `backend-next/internal/httpapi/auth_handler.go` | HTTP handler（/api/register, /api/login） |
| `backend-next/internal/httpapi/auth_middleware.go` | JWT 验证中间件（提取 player_id + company_id） |
| `backend-next/internal/domain/auth/types.go` | ✅ 已有（LoginRequest, LoginResponse, RegisterRequest, Player） |

### 3.2 修改文件

| 文件 | 改动 |
|------|------|
| `backend-next/internal/httpapi/router.go` | 注册 auth 路由 + auth middleware |
| `backend-next/cmd/simapi-next/main.go` | 创建 App, 初始化 storage, DevMode bootstrap |
| `backend-next/internal/app/app.go` | 添加 AuthService 字段 |
| `backend-next/internal/config/config.go` | 添加 DevMode 需要的字段 |

### 3.3 依赖新增

| 依赖 | 用途 |
|------|------|
| `golang.org/x/crypto` | bcrypt 密码哈希 |
| `github.com/google/uuid` | 生成 JWT kid（可选） |

---

## 4. 架构决策

### 4.1 JWT 兼容性

旧 backend 使用自造 HS256 JWT（`middleware/auth.go`）。Phase 2 直接在 `app/auth/jwt.go` 中使用 `golang.org/x/crypto` + 标准 `crypto/hmac` + `encoding/base64` 复制相同的签发逻辑，确保生成的 token 可以被旧 middleware 解析（和反过来）。

Token payload 格式：
```json
{"pid": 1, "cid": 1000001, "iat": 1234567890, "exp": 1234567890 + 259200}
```

签名算法：HS256（HMAC-SHA256），base64 URL 编码。

### 4.2 密码哈希

使用 `bcrypt.DefaultCost`，与旧 backend 一致。迁移后旧用户的密码哈希继续可用（同一算法）。

### 4.3 Response Envelope

使用 `internal/httpapi/response.go` 中已定义的统一格式：
```json
{"data": {...}, "error": null, "meta": {}}
```

### 4.4 Storage

继续使用 `internal/storage/memory/memory.go`（已实现 PlayerStorage 接口）。不接 PostgreSQL。

### 4.5 ID 生成

Player ID 和 Company ID 使用 `platform.IDGen`，与旧 backend 的递增策略兼容。

---

## 5. 验收标准

1. `cd backend-next && go build ./cmd/simapi-next/` → exit 0
2. `cd backend-next && go vet ./...` → exit 0
3. `go test ./internal/...` → all pass（含新增的 auth handler test）
4. 启动 backend-next，`curl POST /api/register` 返回 200 + valid JWT
5. `curl POST /api/login` 使用刚注册的凭据返回 200 + valid JWT
6. 错误的密码返回 401
7. 重复用户名返回 400
8. 旧 backend 的 `/api/register` 和 `/api/login` 仍然可用
9. Dev 模式（`SIM_API_DEV_MODE=true`）在启动时自动创建 dev 用户
10. 前端不报错（确认旧 API 没有被影响）

---

## 6. 回滚方式

```bash
# 后端回滚
cd backend-next
git checkout HEAD~1  # 只是 scaffold 没有旧业务影响
go run ./cmd/simapi-next/

# 前端回滚：不需要，前端没有改
```

实际上 Phase 2 没有改变旧 backend 的任何代码，也没有改变前端，所以"回滚"只影响 backend-next 自己。

---

## 7. 风险

| 风险 | 概率 | 缓解 |
|------|------|------|
| JWT token 格式不兼容，旧 middleware 无法解析 | Low | 复制旧 middleware 的签发逻辑，逐字段对比 |
| bcrypt cost 导致注册响应慢 | Low | DefaultCost 是 10，约 100ms，可接受 |
| Dev mode 用户创建时序问题 | Low | 服务启动完成后才监听端口，dev bootstrap 在 ListenAndServe 之前 |

---

## 8. AI 执行指令

### 8.1 任务

将旧 backend 的注册/登录功能迁移到 `backend-next/`。

### 8.2 具体步骤

1. 在 `app/auth/service.go` 实现：
   - `Register(ctx, req) → (*LoginResponse, error)` — 创建 Player、创建 Company、签发 JWT
   - `Login(ctx, req) → (*LoginResponse, error)` — 验证密码、签发 JWT
   - `DevBootstrap(ctx) error` — 如果 dev 模式且无用户，创建 dev/dev 用户
   - 内部调用 `storage.PlayerStorage` 和 `platform.IDGen`

2. 在 `app/auth/jwt.go` 实现：
   - `SignJWT(playerID, companyID int, key string) (string, error)` — 与旧 backend 兼容的 HS256

3. 在 `httpapi/auth_handler.go` 实现：
   - `handleRegister` — POST /api/register
   - `handleLogin` — POST /api/login
   - 使用 `response.go` 中的 `writeSuccess` 和 `writeErr`

4. 在 `httpapi/auth_middleware.go` 实现：
   - `AuthRequired` — 从 Authorization header 解析 JWT，提取 claims 注入 context

5. 修改 `httpapi/router.go`：在 router 中注册 auth 路由

6. 修改 `cmd/simapi-next/main.go`：初始化 App，调用 DevBootstrap

7. 写测试：`app/auth/service_test.go` 至少覆盖 Register、Login、错误密码、重复用户名

### 8.3 不可做的事

- 不要改 `backend/` 下的任何文件
- 不要改 `client/` 下的任何文件
- 不要改 `backend-next/internal/domain/market/`、`production/`、`finance/`、`research/`、`social/`
- 不要改 `backend-next/internal/storage/postgres/`
- 不要改已有文档

### 8.4 验证命令

```bash
cd backend-next
go build ./cmd/simapi-next/
go vet ./...
go test ./internal/app/auth/...
go test ./internal/httpapi/...
```
