package storage

import (
	"context"
	"go-sim-api/internal/model"
)

type Storage interface {
	Close()
	LoadState(ctx context.Context) (*model.GameState, error)
	SaveState(ctx context.Context, state *model.GameState) error
	SaveCompany(ctx context.Context, c *model.Company) error
	SaveOrders(ctx context.Context, orders []model.MarketOrder) error
	SaveTrades(ctx context.Context, trades []model.Trade) error
}

func New(ctx context.Context, connStr string) (Storage, error) {
	return newPostgres(ctx, connStr)
}

type NoopStorage struct{}

func (n *NoopStorage) Close()                                                    {}
func (n *NoopStorage) LoadState(_ context.Context) (*model.GameState, error)     { return nil, nil }
func (n *NoopStorage) SaveState(_ context.Context, _ *model.GameState) error     { return nil }
func (n *NoopStorage) SaveCompany(_ context.Context, _ *model.Company) error     { return nil }
func (n *NoopStorage) SaveOrders(_ context.Context, _ []model.MarketOrder) error { return nil }
func (n *NoopStorage) SaveTrades(_ context.Context, _ []model.Trade) error       { return nil }
