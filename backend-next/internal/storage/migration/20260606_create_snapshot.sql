-- +goose Up
CREATE TABLE IF NOT EXISTS game_snapshot (
    id INTEGER PRIMARY KEY,
    data JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS game_snapshot;
