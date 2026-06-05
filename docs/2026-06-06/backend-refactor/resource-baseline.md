# Resource Baseline

Last updated: 2026-06-06

This document records the copied baseline resources used for backend refactor planning.

## Copied Files

| Copied file | Source file | Why it matters |
|-------------|-------------|----------------|
| `reference/backend-config/game.json` | `backend/configs/game.json` | Runtime economy tuning, bot behavior, fees, rates, costs. |
| `reference/dependency-baseline/go.mod` | `backend/go.mod` | Backend module and dependency baseline. |
| `reference/dependency-baseline/go.sum` | `backend/go.sum` | Dependency checksum baseline. |
| `reference/game-data/buildings.json` | `decompiled/data/buildings.json` | Building definitions used by shop, placement, production, and map behavior. |
| `reference/game-data/economy_model.json` | `decompiled/data/economy_model.json` | Economy model data used by pricing and resource chains. |
| `reference/game-data/resources.json` | `decompiled/data/resources.json` | Resource definitions, production chain inputs, and output rates. |
| `reference/game-data/resource_lookups.json` | `decompiled/data/resource_lookups.json` | Resource lookup metadata. |

## Handling Rules

- Treat these files as a point-in-time snapshot.
- Do not make runtime code read from `docs/backend-refactor/reference`.
- If the source files change during implementation, refresh this snapshot only in a docs/reference update.
- When formula, production, market, or loader behavior changes, compare against these files before deciding whether behavior changed intentionally.

## Refactor Uses

Use this baseline when planning:

- Formula extraction.
- Resource loader cleanup.
- Production recipe validation.
- Building shop and placement validation.
- Market price-chain calculation.
- Bot order generation and economy tuning.
- Persistence snapshot compatibility.

## Future Additions

Good candidates to add later:

- A generated resource dependency graph.
- A generated building-to-production map.
- A generated endpoint-to-frontend-usage map.
- A saved `go test ./...` baseline result.
- A saved `go vet ./...` baseline result.

