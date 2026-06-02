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

// --- Entity-level CRUD ---
// Individual entity saves provide real-time crash recovery.
// They write to typed domains in sim_snapshot separate from the full game_state backup.

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

// SaveState persists the full GameState as a complete snapshot.
// It also persists individual entity domains so the migration path
// in LoadState stays consistent between full saves.
func (p *postgres) SaveState(ctx context.Context, state *model.GameState) error {
	if err := p.upsert(ctx, "game_state", state); err != nil {
		return err
	}

	// Write individual entity domains for crash recovery consistency.
	// These are the same domains that LoadState reads in the migration path.
	if state.Companies == nil {
		state.Companies = []model.Company{}
	}
	if state.BotCompanies == nil {
		state.BotCompanies = []model.Company{}
	}
	if state.Orders == nil {
		state.Orders = []model.MarketOrder{}
	}
	if state.Trades == nil {
		state.Trades = []model.Trade{}
	}
	if state.Players == nil {
		state.Players = []model.Player{}
	}
	if state.Bonds == nil {
		state.Bonds = []model.Bond{}
	}
	if state.Notifications == nil {
		state.Notifications = []model.Notification{}
	}
	if state.GovernmentContracts == nil {
		state.GovernmentContracts = []model.GovContract{}
	}
	if state.Messages == nil {
		state.Messages = []model.Message{}
	}
	if state.Ledger == nil {
		state.Ledger = []model.LedgerEntry{}
	}
	if state.ProductionJobs == nil {
		state.ProductionJobs = []model.ProductionJob{}
	}
	if state.Auctions == nil {
		state.Auctions = []model.Auction{}
	}

	_ = p.upsert(ctx, "companies", state.Companies)
	_ = p.upsert(ctx, "bot_companies", state.BotCompanies)
	_ = p.upsert(ctx, "orders", state.Orders)
	_ = p.upsert(ctx, "trades", state.Trades)
	_ = p.upsert(ctx, "players", state.Players)
	_ = p.upsert(ctx, "bonds", state.Bonds)
	_ = p.upsert(ctx, "notifications", state.Notifications)
	_ = p.upsert(ctx, "government_contracts", state.GovernmentContracts)
	_ = p.upsert(ctx, "executives", state.Executives)
	_ = p.upsert(ctx, "messages", state.Messages)
	_ = p.upsert(ctx, "ledger", state.Ledger)
	_ = p.upsert(ctx, "production_jobs", state.ProductionJobs)
	_ = p.upsert(ctx, "contracts_in", state.ContractsIn)
	_ = p.upsert(ctx, "contracts_out", state.ContractsOut)
	_ = p.upsert(ctx, "achievements", state.Achievements)
	_ = p.upsert(ctx, "next_player_id", state.NextPlayerID)
	_ = p.upsert(ctx, "player_preferences", state.PlayerPreferences)
	_ = p.upsert(ctx, "market_ticks", state.MarketTicks)
	_ = p.upsert(ctx, "csrf_token", state.CSRFToken)
	_ = p.upsert(ctx, "last_active_at", state.LastActiveAt)

	// Market competition state
	if state.MarketPressure == nil {
		state.MarketPressure = map[int]float64{}
	}
	if state.DailyTradeVolume == nil {
		state.DailyTradeVolume = map[int]float64{}
	}
	if state.DailyTradeQty == nil {
		state.DailyTradeQty = map[int]int{}
	}
	if state.DailyHighPrice == nil {
		state.DailyHighPrice = map[int]float64{}
	}
	if state.DailyLowPrice == nil {
		state.DailyLowPrice = map[int]float64{}
	}
	if state.YesterdayVolume == nil {
		state.YesterdayVolume = map[int]float64{}
	}
	if state.YesterdayHighPrice == nil {
		state.YesterdayHighPrice = map[int]float64{}
	}
	if state.YesterdayClose == nil {
		state.YesterdayClose = map[int]float64{}
	}
	if state.LastTradePrice == nil {
		state.LastTradePrice = map[int]float64{}
	}
	if state.MarketLocked == nil {
		state.MarketLocked = map[int]bool{}
	}
	if state.NationalTeamActive == nil {
		state.NationalTeamActive = map[int]bool{}
	}
	if state.UnlockedRecipes == nil {
		state.UnlockedRecipes = map[int]bool{}
	}
	if state.ResearchedQuality == nil {
		state.ResearchedQuality = map[int]int{}
	}
	if state.ResearchProjects == nil {
		state.ResearchProjects = []model.ResearchProject{}
	}

	_ = p.upsert(ctx, "market_pressure", state.MarketPressure)
	_ = p.upsert(ctx, "daily_trade_volume", state.DailyTradeVolume)
	_ = p.upsert(ctx, "daily_trade_qty", state.DailyTradeQty)
	_ = p.upsert(ctx, "daily_high_price", state.DailyHighPrice)
	_ = p.upsert(ctx, "daily_low_price", state.DailyLowPrice)
	_ = p.upsert(ctx, "yesterday_volume", state.YesterdayVolume)
	_ = p.upsert(ctx, "yesterday_high_price", state.YesterdayHighPrice)
	_ = p.upsert(ctx, "yesterday_close", state.YesterdayClose)
	_ = p.upsert(ctx, "last_trade_price", state.LastTradePrice)
	_ = p.upsert(ctx, "market_locked", state.MarketLocked)
	_ = p.upsert(ctx, "national_team_active", state.NationalTeamActive)
	_ = p.upsert(ctx, "unlocked_recipes", state.UnlockedRecipes)
	_ = p.upsert(ctx, "researched_quality", state.ResearchedQuality)
	_ = p.upsert(ctx, "research_projects", state.ResearchProjects)

	return nil
}

// LoadState loads the full GameState from persistent storage.
// It first tries the complete game_state snapshot.
// If that has no companies, it falls back to the migration path
// which assembles state from individual entity domains.
func (p *postgres) LoadState(ctx context.Context) (*model.GameState, error) {
	state := &model.GameState{}
	_ = p.loadJSON(ctx, "game_state", state)
	if len(state.Companies) > 0 {
		return state, nil
	}

	// Migration/fallback path: assemble from individual entity domains.
	// This also covers the first Postgres run where game_state was never saved.
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
	loadMapJSONB(ctx, p, "bonds", &state.Bonds)
	_ = p.loadJSON(ctx, "contracts_in", &state.ContractsIn)
	_ = p.loadJSON(ctx, "contracts_out", &state.ContractsOut)
	_ = p.loadJSON(ctx, "achievements", &state.Achievements)
	loadMapJSONB(ctx, p, "notifications", &state.Notifications)
	loadMapJSONB(ctx, p, "government_contracts", &state.GovernmentContracts)
	_ = p.loadJSON(ctx, "executives", &state.Executives)
	loadMapJSONB(ctx, p, "messages", &state.Messages)
	loadMapJSONB(ctx, p, "ledger", &state.Ledger)
	loadMapJSONB(ctx, p, "production_jobs", &state.ProductionJobs)
	_ = p.loadJSON(ctx, "player_preferences", &state.PlayerPreferences)
	_ = p.loadJSON(ctx, "market_ticks", &state.MarketTicks)
	_ = p.loadJSON(ctx, "csrf_token", &state.CSRFToken)
	_ = p.loadJSON(ctx, "last_active_at", &state.LastActiveAt)

	// Market competition state
	_ = p.loadJSON(ctx, "market_pressure", &state.MarketPressure)
	_ = p.loadJSON(ctx, "daily_trade_volume", &state.DailyTradeVolume)
	_ = p.loadJSON(ctx, "daily_trade_qty", &state.DailyTradeQty)
	_ = p.loadJSON(ctx, "daily_high_price", &state.DailyHighPrice)
	_ = p.loadJSON(ctx, "daily_low_price", &state.DailyLowPrice)
	_ = p.loadJSON(ctx, "yesterday_volume", &state.YesterdayVolume)
	_ = p.loadJSON(ctx, "yesterday_high_price", &state.YesterdayHighPrice)
	_ = p.loadJSON(ctx, "yesterday_close", &state.YesterdayClose)
	_ = p.loadJSON(ctx, "last_trade_price", &state.LastTradePrice)
	_ = p.loadJSON(ctx, "market_locked", &state.MarketLocked)
	_ = p.loadJSON(ctx, "national_team_active", &state.NationalTeamActive)
	_ = p.loadJSON(ctx, "unlocked_recipes", &state.UnlockedRecipes)
	_ = p.loadJSON(ctx, "researched_quality", &state.ResearchedQuality)
	_ = p.loadJSON(ctx, "research_projects", &state.ResearchProjects)

	// Auctions
	if state.Auctions == nil {
		_ = p.loadJSON(ctx, "auctions", &state.Auctions)
	}

	// DailyOrders
	_ = p.loadJSON(ctx, "daily_orders", &state.DailyOrders)
	_ = p.loadJSON(ctx, "daily_orders_date", &state.DailyOrdersDate)

	// Boost state
	_ = p.loadJSON(ctx, "boost_type", &state.BoostType)
	_ = p.loadJSON(ctx, "boost_ends_at", &state.BoostEndsAt)
	_ = p.loadJSON(ctx, "boost_multiplier", &state.BoostMultiplier)

	// Simulated time
	_ = p.loadJSON(ctx, "simulated_at", &state.SimulatedAt)

	// Last bot cycle
	_ = p.loadJSON(ctx, "last_bot_cycle_at", &state.LastBotCycleAt)

	// XP
	_ = p.loadJSON(ctx, "xp", &state.XP)
	_ = p.loadJSON(ctx, "xp_to_next_level", &state.XpToNextLevel)

	return state, nil
}

func loadMapJSONB[T any](ctx context.Context, p *postgres, domain string, dest *[]T) {
	var v []T
	_ = p.loadJSON(ctx, domain, &v)
	*dest = v
}
