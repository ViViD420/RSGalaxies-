CREATE UNIQUE INDEX IF NOT EXISTS players_id_account_id_trade_fk_idx
    ON players (id, account_id);

CREATE TABLE IF NOT EXISTS player_trades (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL CHECK (status IN ('pending', 'open', 'completed', 'cancelled', 'expired')),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    terminal_at TIMESTAMPTZ NULL,
    CHECK (expires_at > created_at),
    CHECK ((status IN ('completed', 'cancelled', 'expired')) = (terminal_at IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS player_trades_expiry_idx
    ON player_trades (expires_at)
    WHERE status IN ('pending', 'open');

CREATE TABLE IF NOT EXISTS player_trade_participants (
    trade_id TEXT NOT NULL REFERENCES player_trades(id) ON DELETE CASCADE,
    character_id TEXT NOT NULL,
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('initiator', 'target')),
    confirmed_revision BIGINT NULL,
    PRIMARY KEY (trade_id, character_id),
    UNIQUE (trade_id, role),
    FOREIGN KEY (character_id, account_id)
        REFERENCES players (id, account_id) ON DELETE CASCADE,
    CHECK (confirmed_revision IS NULL OR confirmed_revision > 0)
);

CREATE INDEX IF NOT EXISTS player_trade_participants_account_idx
    ON player_trade_participants (account_id, character_id);

CREATE TABLE IF NOT EXISTS player_trade_offers (
    trade_id TEXT NOT NULL,
    character_id TEXT NOT NULL,
    offered_coins BIGINT NOT NULL DEFAULT 0 CHECK (offered_coins >= 0),
    PRIMARY KEY (trade_id, character_id),
    FOREIGN KEY (trade_id, character_id)
        REFERENCES player_trade_participants (trade_id, character_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS player_trade_escrow_items (
    trade_id TEXT NOT NULL,
    owner_character_id TEXT NOT NULL,
    item_id TEXT NOT NULL CHECK (length(item_id) BETWEEN 1 AND 80),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (trade_id, owner_character_id, item_id),
    FOREIGN KEY (trade_id, owner_character_id)
        REFERENCES player_trade_offers (trade_id, character_id) ON DELETE CASCADE
);

-- This reservation table enforces one active trade per character even when a
-- character appears as the target in one trade and initiator in another.
CREATE TABLE IF NOT EXISTS player_trade_active_characters (
    character_id TEXT PRIMARY KEY,
    trade_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    FOREIGN KEY (character_id, account_id)
        REFERENCES players (id, account_id) ON DELETE CASCADE,
    FOREIGN KEY (trade_id, character_id)
        REFERENCES player_trade_participants (trade_id, character_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS player_trade_active_account_idx
    ON player_trade_active_characters (account_id, character_id);
