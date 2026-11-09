ALTER TABLE enemy_spawn_state
    ADD COLUMN IF NOT EXISTS encounter_id TEXT NULL;

CREATE TABLE IF NOT EXISTS enemy_encounter_participants (
    encounter_id TEXT NOT NULL,
    enemy_id TEXT NOT NULL,
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    character_id TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    damage_dealt BIGINT NOT NULL DEFAULT 0 CHECK (damage_dealt >= 0),
    damage_received BIGINT NOT NULL DEFAULT 0 CHECK (damage_received >= 0),
    action_count BIGINT NOT NULL DEFAULT 0 CHECK (action_count >= 0),
    first_contribution_at TIMESTAMPTZ NOT NULL,
    last_contribution_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (encounter_id, character_id)
);

CREATE INDEX IF NOT EXISTS enemy_encounter_participants_enemy_idx
    ON enemy_encounter_participants (enemy_id, encounter_id);

CREATE INDEX IF NOT EXISTS enemy_encounter_participants_character_idx
    ON enemy_encounter_participants (character_id, last_contribution_at DESC);
