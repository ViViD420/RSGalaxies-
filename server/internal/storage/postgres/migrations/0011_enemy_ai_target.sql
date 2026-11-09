ALTER TABLE enemy_spawn_state
    ADD COLUMN IF NOT EXISTS target_character_id TEXT NULL;

CREATE INDEX IF NOT EXISTS enemy_spawn_state_target_idx
    ON enemy_spawn_state (target_character_id)
    WHERE target_character_id IS NOT NULL;
