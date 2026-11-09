-- One durable settlement per enemy spawn life. The composite key is the
-- idempotency boundary: retries and concurrent defeat paths cannot settle twice.
CREATE TABLE IF NOT EXISTS enemy_encounter_settlements (
    enemy_id TEXT NOT NULL,
    encounter_id TEXT NOT NULL,
    settled_at TIMESTAMPTZ NOT NULL,
    reward_policy_version INTEGER NOT NULL DEFAULT 1 CHECK (reward_policy_version > 0),
    PRIMARY KEY (enemy_id, encounter_id)
);

-- Personal reward claims are durable and independent of bag capacity. A full
-- inventory leaves the claim pending rather than discarding valuable rewards.
CREATE TABLE IF NOT EXISTS enemy_encounter_reward_claims (
    enemy_id TEXT NOT NULL,
    encounter_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    experience_skill_id TEXT NOT NULL,
    experience_amount INTEGER NOT NULL DEFAULT 0 CHECK (experience_amount >= 0),
    parent_skill_id TEXT NULL,
    parent_experience_amount INTEGER NOT NULL DEFAULT 0 CHECK (parent_experience_amount >= 0),
    loot_item_id TEXT NULL,
    loot_quantity INTEGER NOT NULL DEFAULT 0 CHECK (loot_quantity >= 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'claimed')),
    created_at TIMESTAMPTZ NOT NULL,
    claimed_at TIMESTAMPTZ NULL,
    PRIMARY KEY (enemy_id, encounter_id, character_id),
    FOREIGN KEY (enemy_id, encounter_id)
        REFERENCES enemy_encounter_settlements (enemy_id, encounter_id) ON DELETE RESTRICT,
    FOREIGN KEY (character_id, account_id)
        REFERENCES players (id, account_id) ON DELETE CASCADE,
    CHECK ((loot_item_id IS NULL) = (loot_quantity = 0)),
    CHECK ((status = 'claimed') = (claimed_at IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS enemy_encounter_reward_claims_account_idx
    ON enemy_encounter_reward_claims (account_id, character_id, created_at DESC);

CREATE INDEX IF NOT EXISTS enemy_encounter_reward_claims_pending_idx
    ON enemy_encounter_reward_claims (character_id, created_at)
    WHERE status = 'pending';
