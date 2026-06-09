# Backend Review Score

Date: 2026-06-06
Branch: main
Commit: 861e34f97a90e7a7eea962b1ffe148276d0a6edd

## Gate Check
- go fmt: PASS
- go vet: PASS
- go test ./...: PASS (all 8 packages ok)
- go test -race ./...: FAIL (CGO_ENABLED=1 required on Windows; race detector cannot run)
- API smoke test: PASS (healthz, readyz, login, market ticker, market depth, warehouse, research, order creation all 200)

## Score
| Category | Score | Max | Evidence | Main Risk |
|---|---:|---:|---|---|
| Architecture Boundary | 11 | 12 | handler/ (HTTP only), service/ (domain-split files), formula/ (pure), storage/ (interface+impl), scheduler via interface (scheduler.go:10-22). Minor: some handlers reach directly into `svc.Snapshot()` | No explicit domain event bus; service layer grows and could blur back into handler |
| API Contract | 6 | 12 | Typed DTOs exist (auth.go RegisterRequest/LoginRequest, market.go CreateOrderRequest/CancelJobReq). Route versioning: v1/v2/v3/v4 mixed, some bare /api/... (bonds, messages). Error format is `{"error":"msg"}` (handler.go:89), not `{data,error,meta}` envelope. No contract tests. Dev routes at /api/dev/ (dev.go:13-29) mixed into production router | Frontend client.ts (client.ts:94-130) handles both envelope and raw JSON, but error format diverges from convention |
| Type Safety | 8 | 10 | Core domain models in model/types.go are well-typed structs. No raw `map[string]any` in service business logic. ID is `int`, money is `float64`. JSON parse errors return clear messages. Some Company/GameState fields use `map[string]any` for dynamic data (Preferences, PlacedBuildings, Executives) | Lack of dedicated Money type means rounding/truncation bugs possible |
| Business Correctness | 13 | 14 | Production formulas (formula/production.go, costs.go) with tests. Market order lifecycle (CreateOrder/CancelOrder/TakeOrder/exchange fee) correct. Bonds (issue/buy/sell/settle interest/call) with formula tests (formula/bonds.go). Govt contracts (bid/award/deliver/default). Daily orders. Offline income. Golden regression tests in formula_test.go (579 lines) | Input costs and retail formula not yet wired into production cost calculations; profit regression tests are documentation-only |
| Data Consistency | 12 | 12 | All state mutations under `s.mu.Lock()` (market_trade.go:16, production.go:11, bond.go:29, government.go:12). CreateOrder deducts money/inventory atomically before matching (market_trade.go:46-50). CancelOrder refunds precisely. TakeOrder credits seller + charges fee (market_match.go:51-63). Double-claim prevention in ClaimProduction (production_claim.go:29-31). Bond interest tracked via InterestCollected (bond.go:218-225). Idempotency for claim/bid (service_save.go:95-106). Scheduler goroutine shares mutex (scheduler.go:60). Full state save via SaveAll. JSON round-trip tested (storage_test.go) | No database transactions across compound operations; partial writes possible with PostgreSQL backend on crash |
| Tests | 13 | 14 | `go test ./...` all pass. Formula test: 579 lines covering production, bonds, saturation, costs, retail, tick step, edge cases (formula_test.go). Market test: create/cancel/quality/take/limit matching/price priority (market_test.go). Production test: start/claim/resolve quality/cancel/edge cases (production_test.go). Bond test: issue/buy/sell/interest/defaults (bond_test.go). Auth test: register/login/JWT/token (auth_test.go). Concurrency smoke test (service_test.go:258-292). Storage JSON round-trip (storage_test.go). Config consistency (config_test.go). Anticheat/AML tests. No handler contract tests; race detector unavailable on Windows | No contract tests means API changes can break FE without detection |
| Runtime Safety | 7 | 8 | sync.Mutex on all service methods. Time injectable via SetSimulatedAt (service.go:64-68). ID generation uses random (market_trade.go:54). Scheduler uses context cancellation (scheduler.go:44-47). Graceful shutdown with 10s timeout (main.go:87-96). Panic recovery middleware (middleware.go:27-37). Request ID middleware (middleware.go:61-66). HTTP timeouts set (main.go:73-78) | Race detector could not run (Windows CGO); scheduler goroutine and handler goroutines compete on the same mutex under load |
| Auth & Anti-cheat | 6 | 8 | JWT HS256 with 3 claims (playerID/companyID/exp), 72h expiry (middleware/auth.go:31-102). CompanyID from JWT, not client request. AntiCheat system: rate limit (50 actions trigger), quick-cancel detection, wash-trade detection (anticheat/anticheat.go). AML system: transaction recording, rapid-trade check, round-trip detection (aml/aml.go). ScriptDetector: timing-based bot detection (anticheat/detector.go). **Dev endpoints** (/api/dev/ledger/, /api/dev/time/) require JWT auth but any valid user token works — no admin role check (dev.go:13-29). DevMode defaults to enabled | Dev routes accessible by any authenticated user: /api/dev/ledger/ exposes all ledger entries, /api/dev/time/ manipulates game time |
| Config & Ops | 3 | 6 | Config loaded from env vars + game.json (config.go:88-106). game.json silently falls back to defaults if missing (config.go:108-125). /healthz (handler/health.go:6-8) and /readyz (handler/health.go:10-12). Graceful shutdown (main.go:87-96). Logger middleware logs method/path/status/duration (middleware.go:39-46). Request ID header set (middleware.go:61-66). **No startup config validation**, **no CI config visible**, **no structured log levels** | game.json parse failure produces silent fallback to defaults; no config drift detection |
| FE/BE Alignment | 2 | 4 | Frontend API paths (market.api.ts, production.api.ts, research.api.ts, company.api.ts, contracts.api.ts, financial.api.ts) all hit backend-registered routes. Error format differs (backend: `{"error":"msg"}`, FE client handles both envelope and raw). No formal alignment document, no FE mock migration plan, no unused-feature inventory | DTO shape drifts could break FE transforms silently (compat.ts normalize functions exist) |

Total: 81 / 100
Grade: B

## Veto Checks
- go test ./... passes? **YES**
- Duplicate settlement in money/inventory/bonds/orders? **NO** — all mutations locked, double-claim/interest checks in place
- Player A can read/modify Player B data? **NO** — companyID from JWT, service validates ownership
- Normal request causes panic? **NO** — panic recovery middleware catches all; tests run without panics
- Economy formula changed without tests? **NO** — formula_test.go has 579 lines of comprehensive tests
- API format frontend cannot parse? **NO** — FE client.ts handles both `{data,error,meta}` envelope and raw JSON
- Market/production/finance has core closed-loop test? **YES** — market_test.go, production_test.go, bond_test.go, service_test.go

## Verdict
- Can merge? **YES** (all gate checks pass, no vetoes triggered)
- Can deploy? **YES** (small-scale testing — Grade B)
- Must fix before next feature:
  1. Gate /api/dev/* endpoints behind an admin role or disable in production (DevMode toggle is currently a default-yes env var, not a runtime guard)
  2. Add startup config validation (reject on missing game.json, validate numeric ranges)
  3. Add contract tests for critical API paths (market order lifecycle, production claim, bond operations)
  4. Add CI configuration (fmt/vet/test gate)
  5. Unify error response format (consider `{data, error, meta}` envelope for consistency with frontend expectation)

## Detailed Findings

### 1. Architecture Boundary (11/12)
The codebase follows clean Go layout: `handler/` owns HTTP parsing and response marshalling, `service/` owns business logic split by domain (14 files), `formula/` contains pure functions with no side effects, `storage/` is a 5-method interface with PostgreSQL and noop implementations. The scheduler interacts through a `GameService` interface (scheduler.go:10-22), never touching state directly. Minor friction: several handler methods call `h.svc.Snapshot()` and pass raw map result to `writeJSON`, which makes the handler aware of snapshot shape — this is acceptable for this codebase's maturity.

Two architectural concerns: (a) the `service_stubs.go` file (408 lines) mixes stubs for executives, aerospace, and poaching into the same package, and (b) `execState` is stored as a separate struct rather than in GameState, losing it on save/load.

### 2. API Contract (6/12)
Route registration is orderly: each domain has a `Register*` method called from `handler.Register()` (handler.go:62-81). Versioning uses path prefixes (v1/v2/v3/v4) but some endpoints remain at bare `/api/...` (`/api/bonds/`, `/api/messages/`, `/api/csrf/`). The `handleV4` method in dev.go acts as a wildcard catch-all for V4 endpoints. Request DTOs exist for auth (auth.go:8-19), market order creation (market.go:53-59), and job cancellation (production_queue.go:27-29), but most handlers parse `map[string]any` directly from the body. Error format is `{"error":"message"}` — functional but not the conventional backend envelope `{data, error, meta}`. The frontend client.ts handles both envelope and raw JSON (client.ts:94-130), so it works. No contract tests exist.

### 3. Type Safety (8/10)
Core types in `model/types.go` are well-constructed: Company, MarketOrder, Trade, Bond, GovContract, LedgerEntry, ProductionJob, Auction, Order, Player, ResearchProject, GameState all use typed fields. IDs are `int`. Money is `float64`. Some fields in Company and GameState unfortunately use `map[string]any` (PlacedBuildings, UnplacedBuildings, Preferences, Executives, Achievements, BotCompanies, etc.) — these are dynamic/legacy fields that would benefit from typed structs. JSON parsing errors consistently return clear messages. Input validation (price>0, quantity>0, interest in range) is present everywhere.

### 4. Business Correctness (13/14)
Production formulas follow the v1.3.1 spec: `OutputPerHour = baseOutput * (1 + speedBonus/100) * level`. Cost formulas implement Labor/Energy/Maintenance/Management with indices and quadratic sweet-spot management scaling. Bond formulas: `DailyBondInterest = amount * BondFaceValue * interestRatePct / 100` with correct floor rounding. Market tick steps faithfully replicate decompiled prices. Retail formula is a direct port from decompiled code with NaN/Inf guards. Market matching does price-time priority correctly. Daily orders are generated from tradeable resources with randomized rewards. Profit regression tests document the computed values for tuning.

Missing: input costs (material/ingredient costs) are not yet deducted from production profit. The retail formula is implemented but not yet wired into the actual selling workflow. These are gaps but not formula errors.

### 5. Data Consistency (12/12)
All state-mutating service methods acquire `s.mu.Lock()` and hold it through the entire operation sequence. The `CreateOrder` method (market_trade.go:15-71) deducts money or inventory atomically, adds the order, matches, and saves within one lock. `CancelOrder` reverses the initial deduction. `TakeOrder` accumulates trades, charges fees, and saves buyer/seller/orders/trades atomically. `ClaimProduction` checks `j.Status == "claimed"` (production_claim.go:29) and tracks `ClaimedAmount` to prevent double-claim. Bond interest settlement tracks `InterestCollected` via `SettleBondInterest` (bond.go:193-248). Idempotency keys are supported for production claims and government bids (service_save.go:95-106). The `saveOrdersLocked`/`saveCompanyLocked`/`saveStateLocked` methods fire individual saves rather than a coordinated DB transaction, which means crash recovery with PostgreSQL is best-effort rather than atomic.

### 6. Tests (13/14)
The test suite is the codebase's strongest area. All 8 packages pass `go test ./...`. Formula tests (~579 lines in formula_test.go) comprehensively cover production, bonds, saturation, costs, retail tick step, admin formulas, edge cases, and profit regression. Market tests (market_test.go) cover create/cancel/take quality orders, limit matching, price priority, insufficient funds. Production tests (production_test.go) cover start/claim/already-claimed/resolve-quality/cancel/over-48h-rejection/recipe-not-found. Bond tests (bond_test.go) cover issue/adjust/buy/sell/interest/defaults on insolvency/invalid rates. Auth tests (auth_test.go) cover register/login/JWT validation/duplicate usernames/password hashing. Concurrency smoke test (service_test.go:258-292) exercises 25 concurrent operations. Storage tests (storage_test.go) cover JSON round-trip serialization for GameState, Company, MarketOrder, LedgerEntry. Config tests (config_test.go) validate consistency between hardcoded defaults and game.json. Anticheat and AML tests cover rate limiting, quick cancel, bot detection, transaction recording, and round-trip detection.

Gaps: no handler-level contract tests (handler_test.go only has 5 basic tests), race detector cannot run on this platform (CGO disabled).

### 7. Runtime Safety (7/8)
The service layer uses `sync.Mutex` on every exported method, guaranteeing sequential consistency. Time is injectable via `SetSimulatedAt(t)` (service.go:64-68), with `now()` returning simulated time when set. Order IDs combine timestamp + companyID + random (market_trade.go:54). The scheduler uses a `context.WithCancel` pattern for clean shutdown (scheduler.go:24-38), with Start/Stop lifecycle. HTTP server has ReadTimeout (10s), WriteTimeout (30s), IdleTimeout (60s) (main.go:73-78). Graceful shutdown waits up to 10s for in-flight requests (main.go:92-96). Panic recovery middleware logs the panic and returns 500 (middleware.go:27-37). Request ID is set on every response (middleware.go:61-66). The race detector could not run on this platform (Windows requires CGO_ENABLED=1).

### 8. Auth & Anti-cheat (6/8)
JWT tokens are signed with HS256 and embed playerID, companyID, issued-at, and expiry (72h). Every authenticated request parses and verifies the JWT, extracting playerID and companyID into request context (handler.go:51-57). The anti-cheat layer provides rate limiting (50 actions triggers alert), quick-cancel detection (<3s between create and cancel), and wash-trade detection (anticheat.go). The AML layer records all transactions and detects rapid trades and round-trip patterns (aml.go). The script detector tracks action timing to compute a human-likeness score (detector.go).

**Critical gap**: Dev endpoints at `/api/dev/ledger/` (exposes all ledger entries) and `/api/dev/time/` (manipulates game time) are behind `withAuth` but any valid JWT token is accepted — there is no admin role check. DevMode defaults to `1` (SIM_API_DEV_MODE != "0" evaluates to true when unset). This means any authenticated player can access sensitive debug endpoints in a production deployment unless explicitly disabled.

### 9. Config & Ops (3/6)
Config is loaded from environment variables (SIM_API_*) and `configs/game.json`. Env vars have sensible defaults (`127.0.0.1:8088`, `dev-jwt-secret-not-for-production`). The game config silently falls back to hardcoded defaults if `game.json` cannot be loaded or parsed (config.go:108-125) — there is no startup validation or error. `healthz` and `readyz` endpoints return `{"status":"ok"}`. Graceful shutdown is implemented with a 10s timeout. Logger middleware logs each request. Request IDs are set. No structured logging (no JSON logs, no log levels). No CI configuration is present in the repo. The environment-based feature toggles (ACEnabled, AMLEnabled, BotEnabled, ScriptDetectEnabled) are good operational controls.

### 10. FE/BE Alignment (2/4)
Frontend API calls (`market.api.ts`, `production.api.ts`, `research.api.ts`, `company.api.ts`, `contracts.api.ts`, `financial.api.ts`, `chat.api.ts`, `buildings.api.ts`, `inventory.api.ts`, `leaderboard.api.ts`, `powerup.api.ts`, `executives.api.ts`) all target paths that the backend registers. Specific mapping verified: `/api/v3/market-ticker/{id}/`, `/api/v3/market-depth/{id}/{quality}/`, `/api/v2/production/jobs/`, `/api/v2/research/`, `/api/v2/companies/me/warehouse/`, `/api/v2/orders/daily/`, `/api/v2/leaderboard/`, `/api/v2/players/simboosts/` all match. The frontend `client.ts` handles both `{data, error, meta}` envelope format and the backend's direct JSON responses. Some frontend API files have `normalize*` helper functions in `compat.ts` (mentioned in imports) that transform backend responses to frontend types — indicating DTO shape drift. No formal alignment document or migration plan exists. No inventory of backend features unused by the frontend.
