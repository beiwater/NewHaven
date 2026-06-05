# Migration Modes

Last updated: 2026-06-06

This document defines the operating modes used while migrating from the current backend A to `backend-next` B.

## 1. Mode Summary

| Mode | Purpose | Risk | Use when |
|------|---------|------|----------|
| `old` | Existing backend behavior | Lowest | Default state, rollback state. |
| `next` | New backend serves response | Medium/high | Domain is tested and ready for user-visible traffic. |
| `shadow` | Old serves response, new is compared | Low/medium | Need confidence before switching. |
| `off` | Bridge disabled | Lowest | Bridge code should not participate for a domain. |

## 2. Old Mode

`old` mode is the baseline.

Behavior:

- Request is handled by current backend A.
- `backend-next` is not called.
- User-visible behavior should match current project behavior exactly.

Use for:

- All domains by default.
- Rollback after any issue.
- Routes that have no B implementation yet.

## 3. Next Mode

`next` mode sends real user-visible traffic to `backend-next`.

Behavior:

- Request is proxied or dispatched to B.
- Response from B is returned to the frontend.
- A is not the response owner for that route.

Requirements before use:

- Contract documented.
- Handler tests exist.
- Domain service tests exist.
- Manual smoke test passes.
- Rollback flag exists.

Recommended first `next` candidates:

- Health/readiness.
- Static resources.
- Static building catalog.
- Read-only company profile.

## 4. Shadow Mode

`shadow` mode compares B without changing user-visible behavior.

Behavior:

- A handles the real request.
- B receives a copy when safe.
- Comparison result is logged.
- User receives A's response.

Shadow mode is best for:

- Read-only routes.
- Deterministic static data.
- Shape comparison for route contracts.
- Market and production read models before command migration.

Shadow mode is not automatically safe for write routes. A write route copied to B can double-spend inventory, create duplicate jobs, issue duplicate tokens, or mutate market state incorrectly.

## 5. Command Route Shadowing

For command routes, use one of these strategies:

### Strategy A: No Shadow

Skip shadow mode for the route until B can become the source of truth.

Best for:

- Market order taking.
- Bond buying/calling.
- Production claiming.

### Strategy B: Dry-Run Endpoint

B exposes a dry-run endpoint that validates and predicts the result without committing mutation.

Example:

```txt
POST /internal/shadow/production/claim/preview
```

Best for:

- Production start.
- Production claim.
- Order validation.

### Strategy C: Isolated State Copy

B runs the command against a copied snapshot, then discards it.

Best for:

- Complex market behavior.
- Finance calculations.
- Scheduler simulation.

## 6. Comparison Policy

For each migrated route, choose one comparison level:

| Level | Required match |
|-------|----------------|
| `status` | HTTP status only. |
| `shape` | JSON keys and value types. |
| `semantic` | Important values match with tolerances. |
| `exact` | Full response equality. |

Default recommendation:

- Static data: `exact`
- Health/config: `shape`
- Company read models: `semantic`
- Production read models: `semantic`
- Market ticker/depth: `semantic`
- Command previews: `semantic`

## 7. Domain Graduation

A domain graduates from A to B when:

- All selected routes have documented contracts.
- Read routes pass shadow comparison.
- Command routes have tests or dry-run comparisons.
- Manual smoke checks pass.
- Rollback mode is confirmed.
- The source of truth for that domain is clearly documented.

Graduation order should be:

1. `system`
2. `catalog`
3. `auth`
4. `company`
5. `production-read`
6. `production-command`
7. `finance-read`
8. `market-read`
9. `market-command`
10. `scheduler`
11. `storage`

## 8. Rollback Policy

Rollback must be boring.

Rollback action:

```txt
SIM_API_ROUTE_<DOMAIN>=old
```

Rollback should not require:

- Database migrations.
- Frontend changes.
- Manual cleanup of duplicated state.
- Deleting files.

If a B command route mutates shared persistent state, rollback is no longer simple. That route must not enter `next` mode until its state ownership is decided.

