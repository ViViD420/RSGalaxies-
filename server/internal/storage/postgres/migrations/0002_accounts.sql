CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    email VARCHAR(320) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    disabled_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS accounts_email_lower_idx ON accounts (LOWER(email));

CREATE TABLE IF NOT EXISTS account_sessions (
    token_hash BYTEA PRIMARY KEY CHECK (octet_length(token_hash) = 32),
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    CHECK (expires_at > created_at)
);

CREATE INDEX IF NOT EXISTS account_sessions_account_idx ON account_sessions (account_id, expires_at);
CREATE INDEX IF NOT EXISTS account_sessions_expiry_idx ON account_sessions (expires_at);

ALTER TABLE players ADD COLUMN IF NOT EXISTS account_id TEXT NULL REFERENCES accounts(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS players_account_idx ON players (account_id, id);
CREATE UNIQUE INDEX IF NOT EXISTS players_account_name_lower_idx ON players (account_id, LOWER(name)) WHERE account_id IS NOT NULL;
