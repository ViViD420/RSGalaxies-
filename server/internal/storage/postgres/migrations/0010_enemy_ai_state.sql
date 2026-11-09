ALTER TABLE enemy_spawn_state
    ADD COLUMN IF NOT EXISTS behavior TEXT NOT NULL DEFAULT 'idle'
        CHECK (behavior IN ('idle', 'chasing', 'telegraphing', 'returning')),
    ADD COLUMN IF NOT EXISTS position_zone_id TEXT NULL,
    ADD COLUMN IF NOT EXISTS position_x DOUBLE PRECISION NULL,
    ADD COLUMN IF NOT EXISTS position_z DOUBLE PRECISION NULL,
    ADD COLUMN IF NOT EXISTS telegraph_until TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS next_attack_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_ai_updated_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS enemy_spawn_state_respawn_idx
    ON enemy_spawn_state (respawn_at)
    WHERE respawn_at IS NOT NULL;
