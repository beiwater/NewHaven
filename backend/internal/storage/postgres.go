package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go-sim-api/internal/model"
)

type postgres struct {
	pool *pgxpool.Pool
}

func newPostgres(ctx context.Context, connStr string) (*postgres, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse conn str: %w", err)
	}
	cfg.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	p := &postgres{pool: pool}
	if err := p.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return p, nil
}

func (p *postgres) Close() { p.pool.Close() }

func (p *postgres) migrate(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS sim_snapshot (
			domain TEXT PRIMARY KEY,
			data JSONB NOT NULL DEFAULT '{}',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (p *postgres) upsert(ctx context.Context, domain string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", domain, err)
	}
	_, err = p.pool.Exec(ctx,
		`INSERT INTO sim_snapshot (domain, data, updated_at) VALUES ($1, $2, NOW())
		 ON CONFLICT (domain) DO UPDATE SET data = $2, updated_at = NOW()`,
		domain, string(b))
	return err
}

func (p *postgres) loadJSON(ctx context.Context, domain string, dest any) error {
	var raw string
	err := p.pool.QueryRow(ctx, `SELECT data FROM sim_snapshot WHERE domain = $1`, domain).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return json.Unmarshal([]byte(raw), dest)
}

func (p *postgres) SaveCompany(ctx context.Context, c *model.Company) error {
	return p.upsert(ctx, "company", c)
}

func (p *postgres) SaveOrders(ctx context.Context, orders []model.MarketOrder) error {
	if orders == nil {
		orders = []model.MarketOrder{}
	}
	return p.upsert(ctx, "orders", orders)
}

func (p *postgres) SaveTrades(ctx context.Context, trades []model.Trade) error {
	if trades == nil {
		trades = []model.Trade{}
	}
	return p.upsert(ctx, "trades", trades)
}

func (p *postgres) LoadState(ctx context.Context) (*model.GameState, error) {
	state := &model.GameState{}
	_ = p.loadJSON(ctx, "game_state", state)
	if len(state.Companies) > 0 {
		return state, nil
	}
	// Migration from old format: load individual company
	var oldCompany model.Company
	_ = p.loadJSON(ctx, "company", &oldCompany)
	if oldCompany.ID != 0 {
		if oldCompany.Inventory == nil {
			oldCompany.Inventory = map[int]int{}
		}
		state.Companies = []model.Company{oldCompany}
	}
	var orders []model.MarketOrder
	_ = p.loadJSON(ctx, "orders", &orders)
	state.Orders = orders
	if state.Orders == nil {
		state.Orders = []model.MarketOrder{}
	}
	var trades []model.Trade
	_ = p.loadJSON(ctx, "trades", &trades)
	state.Trades = trades
	if state.Trades == nil {
		state.Trades = []model.Trade{}
	}
	_ = p.loadJSON(ctx, "players", &state.Players)
	_ = p.loadJSON(ctx, "next_player_id", &state.NextPlayerID)
	_ = p.loadJSON(ctx, "companies", &state.Companies)
	_ = p.loadJSON(ctx, "bot_companies", &state.BotCompanies)
	loadMap := func(domain string) []map[string]any {
		var v []map[string]any
		_ = p.loadJSON(ctx, domain, &v)
		return v
	}
	state.Bonds = loadMapJSONB[model.Bond](ctx, p, "bonds")
	state.ContractsIn = loadMap("contracts_in")
	state.ContractsOut = loadMap("contracts_out")
	state.Achievements = loadMap("achievements")
	state.Notifications = loadMapJSONB[model.Notification](ctx, p, "notifications")
	state.GovernmentContracts = loadMapJSONB[model.GovContract](ctx, p, "government_contracts")
	state.Executives = loadMap("executives")
	state.Messages = loadMapJSONB[model.Message](ctx, p, "messages")
	state.Ledger = loadMapJSONB[model.LedgerEntry](ctx, p, "ledger")
	state.ProductionJobs = loadMapJSONB[model.ProductionJob](ctx, p, "production_jobs")
	_ = p.loadJSON(ctx, "player_preferences", &state.PlayerPreferences)
	_ = p.loadJSON(ctx, "market_ticks", &state.MarketTicks)
	_ = p.loadJSON(ctx, "csrf_token", &state.CSRFToken)
	_ = p.loadJSON(ctx, "last_active_at", &state.LastActiveAt)
	return state, nil
}

func loadMapJSONB[T any](ctx context.Context, p *postgres, domain string) []T {
	var v []T
	_ = p.loadJSON(ctx, domain, &v)
	return v
}
func (p *postgres) SaveState(ctx context.Context, state *model.GameState) error {
	return p.upsert(ctx, "game_state", state)
}
