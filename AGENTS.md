# Repository Guidelines

## Project Overview

New Haven is a browser-based multiplayer economy simulation game — build factories, produce goods, trade on a player-driven market, research upgrades, and compete on leaderboards. Monorepo with a Go API backend and a React+PixiJS frontend.

## Active Rewrite Rules

- `backend/` is the canonical Go API. The former `backend-next/` rewrite has replaced the legacy backend; do not reintroduce a parallel backend tree.
- Codex is responsible for planning, implementation, code review, verification, and final judgment. Do not delegate repository work to OMP unless the user explicitly asks to use OMP again.
- Complete changes end to end: define or update the OpenAPI contract, implement the service and HTTP route, connect the frontend callsite when applicable, add focused tests, and verify the real user flow.
- Commit and push each meaningful, verified rewrite batch. Do not include unrelated user changes in a commit.
- A handwritten source file over 300 lines is a split-risk signal. Prefer domain-focused files before extending an already large file. Generated files, the root OpenAPI document, and shared global styles are exceptions but must still be reviewed carefully.
- Preserve the standard backend response envelope: `{ "data": ..., "error": null }` on success and a typed application error on failure.

### Story Progress

- The backend is the authoritative source for account story progress. Session storage may improve UX, but must never decide permanent story state.
- Newly registered accounts initialize the arrival story as `not_started` at its first step. Existing accounts without story progress are migration-compatible and must not be forced into the story.
- Story progress records a `storyId`, `stepId`, and status: `not_started`, `in_progress`, `completed`, or `skipped`.
- The frontend must save progress while advancing, resume from the saved step after refresh/login, and provide visible continue and skip controls.
- Terminal story states (`completed` and `skipped`) must not regress because of late or out-of-order progress requests.
- The `dev` bootstrap account must enter the game directly and must not play the new-account story.

### Deferred Extension Boundaries

- Anti-cheat/AML, aerospace, executives, research, government contracts, and the remaining bond lifecycle are deferred extension/plugin domains.
- Preserve or document API/plugin boundaries for deferred domains, but do not implement them during core rewrite work unless the user explicitly changes their priority.
- Bond listing, detail, issue, owned, and sold compatibility routes may remain; bond purchase, redemption/call, settlement, and default processing are deferred.

## Architecture & Data Flow

```
Browser (PixiJS Canvas + React UI)
  │  REST/JSON (Bearer JWT)       WebSocket (market ticks)
  ▼
Go backend (net/http, in-memory state, optional PostgreSQL)
  ├── handler/   HTTP endpoints, typed DTOs, JWT auth via middleware
  ├── service/   Business logic, sync.Mutex-guarded game state
  ├── formula/   Pure economic functions (no side effects)
  ├── scheduler/ Background tasks (market matching, bot orders)
  ├── model/     Shared types (Company, MarketOrder, ProductionJob, etc.)
  ├── config/    Env vars + game.json tuning
  └── storage/   Optional PostgreSQL persistence (memory-only by default)
```

**Auth flow**: POST `/api/login` → JWT token → `Authorization: Bearer <token>` header → middleware extracts `player_id`/`company_id` into request context.

**State management**: All mutable state lives in `model.GameState` struct, guarded by `sync.Mutex`. No shared mutable state outside `service.Service`.

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `backend/cmd/simapi/` | Go entry point (30-line main, graceful shutdown) |
| `backend/internal/handler/` | HTTP handlers (one file per domain: company, market, bond, research…) |
| `backend/internal/service/` | Business logic (same domain split, all methods on `*Service`) |
| `backend/internal/model/` | Data types (`types.go` — Company, MarketOrder, ProductionJob, GameState) |
| `backend/internal/formula/` | Pure functions (production rates, bond interest, saturation, costs) |
| `backend/internal/config/` | Environment-based config + `game.json` defaults |
| `backend/internal/scheduler/` | Background tasks (market matching tick, bot replenishment) |
| `backend/configs/game.json` | Economy tuning (bond rates, fees, costs, bot params) |
| `client/atlas-foods-client/src/app/` | App shell (AuthGate, ErrorBoundary, layout routing) |
| `client/atlas-foods-client/src/game/` | PixiJS map canvas, building rendering, texture loading |
| `client/atlas-foods-client/src/features/` | Feature modules (buildings, market, research, chat, executives…) |
| `client/atlas-foods-client/src/api/` | API client (`client.ts` fetch wrapper), TanStack Query hooks |
| `client/atlas-foods-client/src/store/` | Zustand stores (`ui.store.ts`) |
| `client/atlas-foods-client/public/assets/` | Static assets served by Vite (icons, buildings, items, backgrounds) |
| `decompiled/data/` | Static JSON data (resources, buildings, economy model, lookups) |

## Development Commands

### Backend (Go 1.25+)
```bash
cd backend
go run ./cmd/simapi/              # Start (port 8088, memory-only)
go build -o simapi ./cmd/simapi/  # Build binary
go vet ./...                       # Lint
```
Env vars: `SIM_API_DEV_MODE` (default `true` — creates dev user/password), `SIM_API_DATABASE_URL` (optional PostgreSQL).

### Frontend (Node 22+, npm)
```bash
cd client/atlas-foods-client
npm install        # Install dependencies
npm run dev        # Vite dev server (port 5173, proxies /api→:8088)
npm run build      # TypeScript check + Vite production build
npm run lint       # ESLint
npx tsc --noEmit   # TypeScript check only
```

## Code Conventions

### Go Backend
- **Module path**: `go-sim-api` (defined in `backend/go.mod`)
- **Handler pattern**: `func (h *Handler) handleXxx(w http.ResponseWriter, r *http.Request)` — registered via `mux.HandleFunc("/api/v2/.../", h.withAuth(h.handleXxx))`
- **Auth guard**: `h.withAuth(handler)` wraps handler with JWT validation; `h.companyID(r)` extracts company ID from context
- **Error responses**: `writeErr(w, statusCode, message)` and `writeJSON(w, statusCode, data)`
- **Service concurrency**: All public service methods acquire `s.mu.Lock()` / `s.mu.Unlock()`. Internal helpers use `Locked` suffix (e.g., `getCompanyLocked`)
- **ID generation**: Nanosecond timestamps: `fmt.Sprintf("msg-%d", time.Now().UnixNano())`
- **JSON data**: Resources/buildings loaded from `decompiled/data/*.json` — keyed by `dbLetter` (int resource ID)
- **Configuration**: `SIM_API_*` env vars + `backend/configs/game.json` for economy tuning

### TypeScript Frontend
- **Path alias**: `@/` → `src/` (configured in `tsconfig.app.json` and `vite.config.ts`)
- **API client**: `api.get<T>(path)`, `api.post<T>(path, body)` — thin fetch wrapper in `api/client.ts`, sends `Authorization: Bearer` from localStorage
- **Data fetching**: TanStack Query hooks (e.g., `useBuildings()`, `useProductionJobs()`) in `api/*.api.ts`
- **State**: Zustand in `store/ui.store.ts` — activeView, selectedBuildingId, placementBuildingId, movingBuildingId, chatOpen, etc.
- **Feature modules**: Each feature in `features/<name>/` — self-contained (components, no cross-feature imports except shared UI)
- **PixiJS**: v8, `Assets.load()` for texture preloading, canvas mounted in `GameCanvas.tsx` (lazy-loaded via `React.lazy`)
- **Styling**: Tailwind CSS v4 with `@apply`-free utility classes; custom amber-themed design system
- **Components**: Functional components with hooks; `useMemo` for derived data, `useRef` for mutable refs
- **Error boundaries**: `ErrorBoundary` wraps game canvas and entire app

## Important Files

| File | Purpose |
|------|---------|
| `backend/cmd/simapi/main.go` | Entry point, finds project root, wires layers |
| `backend/internal/handler/handler.go` | Handler struct, `withAuth`, `companyID`, route registration |
| `backend/internal/service/service.go` | `Service` struct, `New()`, `GameState` initialization, helper utils |
| `backend/internal/model/types.go` | All data types (Company, MarketOrder, ProductionJob, GameState, Message…) |
| `backend/configs/game.json` | Economy parameters |
| `decompiled/data/resources.json` | Resource definitions (dbLetter, producedFrom, producedPerHourRaw) |
| `client/atlas-foods-client/src/app/App.tsx` | Layout routing (tab-based, no react-router), GameLayout, MapSlot |
| `client/atlas-foods-client/src/api/client.ts` | Fetch wrapper, JWT storage, AuthGate integration |
| `client/atlas-foods-client/src/store/ui.store.ts` | Zustand store (activeView, selectedBuildingId, placement/moving ids) |
| `client/atlas-foods-client/src/game/GameCanvas.tsx` | PixiJS map init, texture preloading, building layer, pan/zoom/click |
| `client/atlas-foods-client/vite.config.ts` | Vite config (proxy /api→:8088, @ alias, port 5173) |
| `client/atlas-foods-client/src/features/buildings/BuildingCard.tsx` | Building detail card (production, upgrade, move, demolish) |

## Runtime / Tooling Preferences

- **Go**: 1.25+, standard library HTTP server, `golang.org/x/crypto` for bcrypt
- **Node**: 22+ with npm (pnpm optional — `package-lock.json` is committed)
- **Vite**: 8.x, dev server on :5173, proxies `/api`→`http://127.0.0.1:8088`, `/ws`→`ws://127.0.0.1:8088`
- **PixiJS**: 8.x, WebGL canvas with `Assets` loader
- **React**: 19.x with TypeScript 6.x strict mode
- **Tailwind CSS**: 4.x with `@tailwindcss/vite` plugin
- **State**: Zustand 5.x, TanStack Query 5.x

## Testing & QA

- **Backend tests**: `go test ./...` in `backend/` — service-level tests in `internal/service/*_test.go` and `internal/formula/formula_test.go`
- **Frontend**: TypeScript strict mode (`noUnusedLocals`, `noUnusedParameters`, `erasableSyntaxOnly`) enforces code quality
- **Manual verification**: Login as `dev`/`dev` (DevMode creates demo company with buildings, inventory, bonds, research projects)
- **No E2E/CI configured** — verify changes by starting both servers and testing in browser
