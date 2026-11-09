-- Preserve the single claim/idempotency boundary while allowing one claim to
-- carry multiple aggregated item stacks. Existing one-item claims are copied
-- before the new reader becomes authoritative.
CREATE TABLE IF NOT EXISTS enemy_encounter_reward_claim_items (
    enemy_id TEXT NOT NULL,
    encounter_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    item_id TEXT NOT NULL CHECK (length(item_id) > 0),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (enemy_id, encounter_id, character_id, item_id),
    FOREIGN KEY (enemy_id, encounter_id, character_id)
        REFERENCES enemy_encounter_reward_claims (enemy_id, encounter_id, character_id)
        ON DELETE CASCADE
);

INSERT INTO enemy_encounter_reward_claim_items (
    enemy_id, encounter_id, character_id, item_id, quantity
)
SELECT enemy_id, encounter_id, character_id, loot_item_id, loot_quantity
FROM enemy_encounter_reward_claims
WHERE loot_item_id IS NOT NULL AND loot_quantity > 0
ON CONFLICT (enemy_id, encounter_id, character_id, item_id) DO NOTHING;

CREATE INDEX IF NOT EXISTS enemy_encounter_reward_claim_items_claim_idx
    ON enemy_encounter_reward_claim_items (enemy_id, encounter_id, character_id);
