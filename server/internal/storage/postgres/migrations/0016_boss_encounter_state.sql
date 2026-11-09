-- Boss state is isolated by the same encounter-cycle identity used by the
-- enemy contribution and reward ledgers. Deleting a spawn row cascades its
-- transient boss state. Old cycles remain isolated and cannot overwrite a
-- newly respawned boss.
CREATE TABLE IF NOT EXISTS boss_encounter_states (
    enemy_id TEXT NOT NULL REFERENCES enemy_spawn_state(enemy_id) ON DELETE CASCADE,
    encounter_id TEXT NOT NULL CHECK (length(encounter_id) > 0),
    definition_id TEXT NOT NULL CHECK (length(definition_id) > 0),
    definition_version INTEGER NOT NULL CHECK (definition_version > 0),
    phase_id TEXT NOT NULL CHECK (length(phase_id) > 0),
    phase_started_at TIMESTAMPTZ NOT NULL,
    active_ability_id TEXT NULL,
    telegraph_started_at TIMESTAMPTZ NULL,
    telegraph_until TIMESTAMPTZ NULL,
    ability_cooldowns JSONB NOT NULL DEFAULT '{}'::jsonb
        CHECK (jsonb_typeof(ability_cooldowns) = 'object'),
    next_ability_index INTEGER NOT NULL DEFAULT 0 CHECK (next_ability_index >= 0),
    last_updated_at TIMESTAMPTZ NOT NULL,
    defeated BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (enemy_id, encounter_id),
    CHECK (
        (active_ability_id IS NULL AND telegraph_started_at IS NULL AND telegraph_until IS NULL)
        OR
        (active_ability_id IS NOT NULL AND telegraph_started_at IS NOT NULL
            AND telegraph_until IS NOT NULL AND telegraph_until >= telegraph_started_at)
    ),
    CHECK (NOT defeated OR active_ability_id IS NULL)
);

CREATE INDEX IF NOT EXISTS boss_encounter_states_cycle_updated_idx
    ON boss_encounter_states (enemy_id, encounter_id, last_updated_at DESC);
