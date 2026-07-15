# Database Migrations

This directory contains goose migration files for the backend PostgreSQL schema.

## Usage

```bash
# Create a new migration
goose create add_some_table sql

# Apply all pending migrations
goose up

# Rollback the last migration
goose down

# Check status
goose status
```

## Naming Convention

Files: `YYYYMMDDHHMMSS_description.sql`

Always provide both `-- +goose Up` and `-- +goose Down` sections.

## Initial Schema

To be designed from `docs/2026-06-06/rebuild/03-database-design-v1.md`.

Core tables planned:
- players
- companies
- company_buildings
- warehouse_items
- production_jobs
- market_orders
- market_trades
- market_tickers
- ledger_entries
- bonds
- bond_holdings
- research_progress
- chat_messages
- notifications
- audit_events
