# Grid Farm Game Materials (Project Binder)

Last updated: 2026-06-01 (local workspace snapshot)

This document is the consolidated binder for the current game prototype. It collects the design rules, math models, data definitions, interaction conventions, assets, and API surface that exist in the workspace.

If something conflicts with code, treat code as the current source of truth and update this binder next.

## 0. Project Goals (Current)

- A web-based grid farm management game (8x8 farm).
- Player can build on tiles with a safe, intentional flow (no accidental builds).
- Core loop: place buildings -> produce resources -> configure restaurant -> 12h restaurant settlement -> earn cash -> expand.
- All economic settlement is server-authoritative; the client submits intents only.
- Add a global economic cycle (changes every 7 days) that affects restaurant demand and profitability.
- i18n is required for the real experiment; MVP keeps a lightweight structure.

## 1. Versioning Convention

Experiment versions use `UTC timestamp + short random suffix`:

```txt
yyyyMMddTHHmmssZ-6hex
Example: 20260601T000132Z-e32490
```

Where it lives:

- `shared/version.ts`
- Updated via `npm run version:stamp` which writes `shared/version.ts`

Related note:

- Documented in `docs/project-conventions.md`

## 2. Folder Map (Important Files)

Design + math:

- `docs/restaurant-math-model.md` (restaurant math model)
- `docs/project-conventions.md` (versioning + i18n convention)
- `docs/game-materials.md` (this binder)

Shared types + rules:

- `shared/types.ts` (domain types)
- `shared/data.ts` (static definitions: resources, buildings, restaurant menu/staff/style, economic cycles)
- `shared/economy.ts` (math + placement rules)

Server:

- `server/index.ts` (Express routes)
- `server/gameState.ts` (authoritative in-memory state + settlement logic)
- `server/security.ts` (request signature + nonce replay protection)

Client:

- `src/ui/App.tsx` (UI shell and panels)
- `src/game/FarmScene.ts` (Phaser scene: grid, sprites, input)
- `src/game/PhaserFarm.tsx` (Phaser boot + wiring)
- `src/api/client.ts` (signed API client)
- `src/store/gameStore.ts` (Zustand state + UI modes)
- `src/i18n/*` (lightweight i18n)

Assets:

- `src/assets/buildings/*.png` (building sprites currently used by Phaser + UI)

## 3. Domain Overview

### 3.1 Farm Grid

- Grid size: `8 x 8`
- Buildings occupy rectangular footprints (width/height), with `normal` and `rotated` orientations.
- Placement constraints:
  - Must be within grid bounds.
  - Must not overlap existing building footprint.
  - Server validates all placement/moves (client preview is not authoritative).

Relevant implementation:

- `shared/economy.ts`: `footprint`, `occupiedCells`, `canPlaceBuilding`

### 3.2 Resources (MVP)

Resource set:

- `wheat`
- `vegetable`
- `meat`
- `bread`
- `meal`

Defined in:

- `shared/data.ts` (`resources`)

### 3.3 Buildings (MVP)

Building set:

- `field` (producer, 1x1): produces wheat
- `garden` (producer, 1x1): produces vegetable
- `ranch` (producer, 2x1): produces meat
- `bakery` (processor, 2x1): wheat -> bread
- `kitchen` (processor, 2x2): bread + vegetable + meat -> meal
- `restaurant` (restaurant, 2x2): sells via restaurant cycle
- `warehouse` (storage, 1x2): storage placeholder (no special mechanics yet)

Defined in:

- `shared/data.ts` (`buildings`)

### 3.4 Restaurant System (Key Variables)

Restaurant is modeled as a 12-hour retail cycle. Player configures:

- Menu configuration: selected menu items across categories
- Menu price: `60..350`
- Staff type (ability + wage cost)
- Style (seats + operating cost + rating multiplier)

Server computes:

- `rating` (0..100)
- `cycleCapacity` (bounded by seats and food availability)
- `occupancyRate` (0..1, depends on price/rating/reputation/economic cycle/volatility)
- `soldCustomers`
- `grossRevenue`
- `operatingCost`
- `netIncome` (currently floored at 0)
- `reputation` update

Documented in:

- `docs/restaurant-math-model.md`

Implemented in:

- `shared/economy.ts`: restaurant formulas
- `server/gameState.ts`: restaurant settle pipeline + inventory consumption + logs

### 3.5 Economic Cycle (7-day)

Global economic cycle changes every 7 days, and affects restaurant demand:

- Cycle length: `7 days`
- The cycle is determined by a fixed epoch and cycling through predefined cycle list.
- It supplies:
  - `demandMultiplier`
  - `priceSensitivity`
  - `ratingSensitivity`
  - `volatility`

Where it lives:

- `shared/data.ts`: `economicCycles`
- `server/gameState.ts`: computes `economyCycle`, `nextEconomyCycleAt`
- Included in API state for top-bar display

### 3.6 Security / Anti-cheat (MVP)

For non-GET requests, the client includes:

- `x-game-timestamp` (ms since epoch)
- `x-game-nonce` (UUID)
- `x-game-signature` (HMAC SHA-256 of `timestamp.nonce.body`)

Server checks:

- Timestamp window (5 minutes)
- Nonce replay within window
- Signature correctness

Implementation:

- `server/security.ts` + `src/api/client.ts`

Note:

- This is a “raising the bar” measure only; server-authoritative rules remain the real anti-cheat.

## 4. Interaction Conventions (Current UI Behavior)

### 4.1 Normal Mode

- Left-click on a building: open details panel.
- Right-click on a building: open image popover preview.
- No building placement is possible in normal mode.

### 4.2 Build Mode (Safety / Insurance)

Building placement requires an explicit build mode:

1. Player chooses a building and presses `Build` in the shop panel.
2. Cursor movement previews placement on the grid.
3. Player left-clicks a position: this creates a pending placement (still not built).
4. A confirm bar appears on the map: `Confirm Build` / `Cancel`.
5. Only `Confirm Build` triggers server placement.

Escape handling:

- `Esc` cancels build mode.

### 4.3 Restaurant Panel (Structure)

Restaurant UI contains:

- Current cycle stats (rating, occupancy, price, meal stock)
- Staff selector
- Style selector
- Price input
- Menu item picker
- Save config + Settle now

## 5. Restaurant Math Model (Summary)

For full detail, see `docs/restaurant-math-model.md`.

Key functional blocks:

- Menu validity: must contain at least 1 item in each of `starter`, `main`, `drink`.
- Rating: derived from avg menu quality, menu bonuses, staff ability/bonus, style bonuses, reputation, level; style can multiply rating.
- Capacity:
  - `foodCapacity` limited by inventory across the selected menu
  - `cycleCapacity = min(styleSeats, foodCapacity)`
  - menu inputs are consumed for the cycle (prepared/allocated), not only sold units
- Occupancy:
  - depends on `recommendedPrice/menuPrice`, rating, reputation, and economic cycle multipliers
  - includes deterministic “volatilityFactor” bounded to `[0.85, 1.15]`
- Income:
  - `gross = soldCustomers * menuPrice`
  - `operatingCost = staffWageCost + styleOperatingCost`
  - `netIncome = max(0, gross - operatingCost)`
- Reputation:
  - moves slowly based on occupancy banding

## 6. Static Data Definitions (What Exists)

### 6.1 Restaurant Menu Items

Defined in `shared/data.ts` as `restaurantMenuItems`.

Each item has:

- `category`: `starter | main | drink`
- `input`: resource consumption per serving
- `quality`: 0..100
- `ratingBonus`: integer bonus

### 6.2 Restaurant Staff

Defined in `shared/data.ts` as `restaurantStaff`.

Each staff type has:

- `ability`: 0..100
- `wageCost`: per 12h cycle
- `ratingBonus`

### 6.3 Restaurant Styles

Defined in `shared/data.ts` as `restaurantStyles`.

Each style has:

- `seats`: maximum customers per cycle (upper bound)
- `operatingCost`: fixed operating cost per cycle
- `ratingBonus`
- `ratingMultiplier`

### 6.4 Economic Cycles

Defined in `shared/data.ts` as `economicCycles`.

Each cycle has:

- `demandMultiplier`
- `priceSensitivity`
- `ratingSensitivity`
- `volatility`

## 7. Assets (What’s Available)

Building sprites currently used by the game:

- `src/assets/buildings/field.png`
- `src/assets/buildings/garden.png`
- `src/assets/buildings/ranch.png`
- `src/assets/buildings/bakery.png`
- `src/assets/buildings/kitchen.png`
- `src/assets/buildings/restaurant.png`
- `src/assets/buildings/warehouse.png`

Source atlas (generated):

- `src/assets/buildings/building-atlas.png`

## 8. API Surface (Current)

Base:

- Server: `http://127.0.0.1:8787`
- Client dev: proxied via Vite `/api` to the server

Endpoints (MVP):

- `GET /api/state` -> `GameState`
- `POST /api/dev/reset` -> reset in-memory state
- `POST /api/buildings` -> place building
- `PATCH /api/buildings/:id` -> move building
- `DELETE /api/buildings/:id` -> demolish building
- `POST /api/production/:buildingId/start` -> start a job
- `POST /api/production/:jobId/claim` -> claim job output
- `POST /api/restaurant/menu` -> set menu price only (legacy helper)
- `POST /api/restaurant/config` -> configure restaurant (price + staff + style + menu items)
- `POST /api/restaurant/settle` -> settle restaurant cycle (dev forces settlement)
- `POST /api/market/orders` -> create market order

All non-GET endpoints require signed headers (see security section).

## 9. i18n Notes

MVP has a minimal i18n structure:

- `src/i18n/messages.ts`
- `src/i18n/index.ts`

Real experiment requirement:

- All UI copy should route through `t(key)` and message catalogs should be expanded.

## 10. Known Gaps / TODO (Product)

- Build mode polish: show explicit invalid reasons in preview, not only at confirm time.
- Right-click behavior might become a context menu later; current “image popover” is placeholder.
- Restaurant history: currently logs show latest few lines; needs proper charting and longer series.
- Market system: currently minimal; later can incorporate economy cycle and demand links.
- Persistence: state is in-memory; Prisma schema exists but is not wired in yet.

## 11. How To Run (For This Workspace)

Install (PowerShell):

```powershell
cd C:\Users\QWQ\Desktop\newgame
npm.cmd install
```

Dev:

```powershell
npm.cmd run dev
```

Then open:

- Client: `http://127.0.0.1:5173`
- API: `http://127.0.0.1:8787/api/state`

Note:

On this machine, `npm` may be blocked by PowerShell execution policy; use `npm.cmd`.
