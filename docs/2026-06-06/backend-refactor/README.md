# Backend Refactor Pack

Last updated: 2026-06-06

This folder is the working packet for the backend refactor cycle. It collects the constitution, inventories, copied baseline resources, and future architecture decisions in one place.

Nothing in this folder is runtime code. Files here are planning, reference, and review material.

## Contents

| Path | Purpose |
|------|---------|
| `../backend-refactor-constitution.md` | Main refactor constitution and phase plan. |
| `backend-inventory.md` | Current backend package and responsibility inventory. |
| `api-route-inventory.md` | Current HTTP route groups and compatibility notes. |
| `resource-baseline.md` | Copied config/data files and how to treat them during refactor. |
| `backend-next-plan.md` | Plan for building a parallel `backend-next` without breaking the current backend. |
| `bridge-routing-plan.md` | Bridge route strategy for old/next/shadow migration. |
| `migration-modes.md` | Definitions and safety rules for old, next, shadow, and off modes. |
| `prototypes/` | Pseudocode and design sketches before implementation. |
| `reference/backend-config/game.json` | Snapshot of `backend/configs/game.json`. |
| `reference/dependency-baseline/go.mod` | Snapshot of backend module requirements. |
| `reference/dependency-baseline/go.sum` | Snapshot of backend dependency checksums. |
| `reference/game-data/*.json` | Snapshot of static game data from `decompiled/data`. |
| `adr/` | Future architecture decision records. |

## How To Use This Pack

- Start with `../backend-refactor-constitution.md` before creating backend refactor tasks.
- Use `backend-inventory.md` to choose one backend domain at a time.
- Use `api-route-inventory.md` before moving handlers or changing DTOs.
- Use `resource-baseline.md` before touching data loaders, economy config, formulas, or production logic.
- Use `prototypes/` before writing `backend-next` code.
- Use `bridge-routing-plan.md` before adding any old-to-next routing behavior.
- Add ADRs for decisions that change long-term architecture, dependency policy, storage strategy, or API versioning.

## Refactor Rule

The copied files under `reference/` are evidence, not source of truth. Runtime code should keep reading from the original project locations unless a later implementation task explicitly changes that.
