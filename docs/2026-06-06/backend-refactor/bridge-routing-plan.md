# Bridge Routing Plan

Last updated: 2026-06-06

This document defines the route bridge that will allow the current backend A and future `backend-next` B to coexist during migration.

## 1. Goal

The bridge lets the project migrate backend domains one at a time without forcing the frontend to switch URLs manually.

Target model:

```txt
Frontend
  -> current backend port
  -> bridge decision
      -> old handler A
      -> backend-next B
      -> shadow compare A and B
```

The old backend remains the stable public entry point until a later deployment decision changes that.

## 2. Route Modes

Each migratable domain can run in one of these modes:

| Mode | User response comes from | What happens |
|------|--------------------------|--------------|
| `old` | Current backend A | Existing behavior only. |
| `next` | `backend-next` B | Route is served by new backend. |
| `shadow` | Current backend A | Bridge also calls B and records comparison result. |
| `off` | Current backend A | Bridge logic disabled for that domain. |

Default mode must be `old`.

## 3. Suggested Environment Flags

Use one flag per domain so migration can be granular.

```txt
SIM_API_NEXT_BASE_URL=http://127.0.0.1:8090
SIM_API_ROUTE_SYSTEM=old
SIM_API_ROUTE_AUTH=old
SIM_API_ROUTE_CATALOG=old
SIM_API_ROUTE_COMPANY=old
SIM_API_ROUTE_PRODUCTION=old
SIM_API_ROUTE_MARKET=old
SIM_API_ROUTE_FINANCE=old
SIM_API_ROUTE_RESEARCH=old
SIM_API_ROUTE_SOCIAL=old
```

Allowed values:

```txt
old
next
shadow
off
```

## 4. Bridge Decision Flow

```txt
request arrives
  -> identify route domain
  -> read domain route mode
  -> old:
       call old handler
  -> next:
       proxy request to backend-next
  -> shadow:
       call old handler for real response
       copy request to backend-next in background or bounded sync path
       compare status/body shape
       log comparison result
  -> off:
       call old handler
```

## 5. Shadow Mode Rules

Shadow mode is for comparison, not mutation.

Safe shadow candidates:

- Health.
- Static catalog/resource routes.
- Company read models.
- Production read models.
- Market ticker/resource info/depth.
- Financial read models.

Dangerous shadow candidates:

- Login/register if token issuance mutates state.
- Production start/cancel/claim.
- Market create/cancel/take.
- Bond buy/call/settle.
- Scheduler ticks.

For dangerous routes, shadow mode must either:

- Use a read-only simulation endpoint in B, or
- Run against an isolated copied state, or
- Be skipped until a safe comparison method exists.

## 6. Comparison Levels

Not all routes need byte-for-byte equality.

| Level | Meaning | Use case |
|-------|---------|----------|
| `status` | HTTP status matches. | Early bridge wiring. |
| `shape` | JSON keys and types match. | Most compatibility routes. |
| `semantic` | Important fields match with tolerated ordering/time differences. | Market, production, finance. |
| `exact` | Full JSON response matches. | Static data and deterministic reads. |

Comparison output should include:

- Route.
- Mode.
- HTTP method.
- Old status.
- New status.
- Comparison level.
- Difference summary.
- Request ID if available.

## 7. Proxy Rules

When proxying to `backend-next`:

- Forward method, path, query, body, and auth header.
- Preserve request context cancellation.
- Set a bounded timeout.
- Do not leak internal B errors directly if old API error shape differs.
- Normalize response headers to the old backend's API expectations.

Suggested timeout defaults:

```txt
read routes: 2s
command routes: 5s
shadow calls: 1s
```

## 8. First Bridge Candidates

Start with routes that are read-only and deterministic:

1. `/healthz`
2. `/readyz`
3. `/api/v3/resources/`
4. `/api/v3/resources-info/`
5. `/api/v2/recipes/`
6. `/api/v2/companies/me/buildings/`
7. `/api/v2/production/jobs/`
8. `/api/v3/market-ticker/`
9. `/api/v3/market-depth/`

Do not begin with:

- `/api/v2/market-order/`
- `/api/v2/market-order/take/`
- `/api/v2/production/claim/`
- `/api/bonds/`

## 9. Rollback Rule

Every bridged domain must be able to return to `old` mode without code changes.

If `next` mode causes regressions:

```txt
SIM_API_ROUTE_<DOMAIN>=old
```

Then restart the backend if config is only read at startup.

Later, hot reload can be considered, but it is not required for the first bridge.

## 10. Bridge Done Criteria

The bridge is acceptable when:

- Default behavior is identical to current backend.
- Domain route mode can be configured.
- `next` mode can proxy one safe route to B.
- `shadow` mode can compare one safe read route.
- Timeout and error handling are bounded.
- Logs show clear comparison summaries.
- Rollback to `old` is simple.

