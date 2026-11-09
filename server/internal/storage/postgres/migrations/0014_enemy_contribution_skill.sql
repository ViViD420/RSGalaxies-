ALTER TABLE enemy_encounter_participants
    ADD COLUMN IF NOT EXISTS experience_skill_id TEXT NOT NULL DEFAULT 'unarmed';
