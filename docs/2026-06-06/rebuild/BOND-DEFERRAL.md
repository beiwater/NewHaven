# Bond Deferred Capability

Date: 2026-06-06

Bonds are not part of the core rewrite completion criteria. Existing migrated
compatibility routes remain supported, but no additional bond lifecycle work is
required before legacy retirement.

## Existing Compatibility Surface

| Operation | Route | Status |
|---|---|---|
| List active bonds | `GET /api/bonds/` | Migrated |
| Issue bond | `POST /api/bonds/` | Migrated |
| Get bond detail | `GET /api/bonds/{bondId}/` | Migrated |
| List owned bonds | `GET /api/v2/companies/me/bonds/owned/` | Migrated |
| List sold bonds | `GET /api/v2/companies/me/bonds/sold/` | Migrated |

## Reserved Extension Surface

The following capabilities are reserved for a future optional bond extension:

- settle interest: `POST /api/bonds/settle-interest/`
- buy bonds on a secondary market
- call or redeem issued bonds early
- scheduler-driven interest and maturity processing

Reserved bond capabilities must be isolated from core startup, persistence,
scheduler operation, frontend cutover, and legacy retirement. They are
classified as `deferred-plugin` and do not count against core completion.
