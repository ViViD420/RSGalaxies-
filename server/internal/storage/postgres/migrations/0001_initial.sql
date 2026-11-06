CREATE TABLE IF NOT EXISTS players (
    id TEXT PRIMARY KEY,
    name VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    skills JSONB NOT NULL,
    inventory JSONB NOT NULL,
    zone_id TEXT NOT NULL,
    position_x DOUBLE PRECISION NOT NULL,
    position_z DOUBLE PRECISION NOT NULL,
    last_move_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS players_name_idx ON players (name);

CREATE TABLE IF NOT EXISTS gathering_node_state (
    node_id TEXT PRIMARY KEY,
    charges INTEGER NOT NULL DEFAULT 0 CHECK (charges >= 0),
    respawn_at TIMESTAMPTZ NULL,
    initialized BOOLEAN NOT NULL DEFAULT FALSE
);
