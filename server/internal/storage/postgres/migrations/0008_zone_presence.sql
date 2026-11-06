CREATE TABLE IF NOT EXISTS zone_presence (
    character_id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    zone_id TEXT NOT NULL,
    position_x DOUBLE PRECISION NOT NULL,
    position_z DOUBLE PRECISION NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS zone_presence_active_idx ON zone_presence (zone_id, last_seen_at DESC);
CREATE INDEX IF NOT EXISTS zone_presence_account_idx ON zone_presence (account_id, last_seen_at DESC);
