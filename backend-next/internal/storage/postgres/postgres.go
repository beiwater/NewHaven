package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newhaven/backend-next/internal/storage"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

// Store wraps a pgxpool to provide snapshot persistence alongside the in-memory store.
type Store struct {
	pool *pgxpool.Pool
	mem  *memory.Store
}

// New creates a new Postgres-backed snapshot store.
// It auto-creates the game_snapshot table if it does not exist.
func New(ctx context.Context, connStr string, mem *memory.Store) (*Store, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	// Auto-create snapshot table
	ddl := `CREATE TABLE IF NOT EXISTS game_snapshot (
		id INTEGER PRIMARY KEY,
		data JSONB NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`
	if _, err := pool.Exec(ctx, ddl); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: create table: %w", err)
	}

	return &Store{pool: pool, mem: mem}, nil
}

// Close shuts down the connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// SaveSnapshot serialises the current in-memory game state and stores it in Postgres.
func (s *Store) SaveSnapshot(ctx context.Context) error {
	snap := s.mem.GetSnapshotData()
	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("postgres: marshal snapshot: %w", err)
	}

	const upsert = `INSERT INTO game_snapshot (id, data, updated_at) VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE SET data = $1, updated_at = NOW()`
	if _, err := s.pool.Exec(ctx, upsert, data); err != nil {
		return fmt.Errorf("postgres: save snapshot: %w", err)
	}
	return nil
}

// LoadSnapshot reads the saved snapshot from Postgres and populates the in-memory store.
func (s *Store) LoadSnapshot(ctx context.Context) error {
	const query = `SELECT data FROM game_snapshot WHERE id = 1`
	var raw []byte
	if err := s.pool.QueryRow(ctx, query).Scan(&raw); err != nil {
		return fmt.Errorf("postgres: load snapshot: %w", err)
	}

	var snap storage.GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return fmt.Errorf("postgres: unmarshal snapshot: %w", err)
	}

	s.mem.LoadFromSnapshot(&snap)
	return nil
}
