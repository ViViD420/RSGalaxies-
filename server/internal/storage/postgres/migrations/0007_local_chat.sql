CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    zone_id TEXT NOT NULL,
    sender_account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    sender_character_id TEXT NOT NULL,
    sender_character_name VARCHAR(20) NOT NULL,
    message VARCHAR(240) NOT NULL CHECK (length(message) BETWEEN 1 AND 240),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS chat_zone_cursor_idx ON chat_messages (zone_id, id);
CREATE INDEX IF NOT EXISTS chat_sender_cooldown_idx ON chat_messages (sender_character_id, id DESC);
