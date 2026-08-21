-- AI Providers and Models
CREATE TABLE IF NOT EXISTS ai_providers (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    display_name    TEXT,
    base_url        TEXT,
    api_key_enc     TEXT,
    api_type        TEXT DEFAULT 'openai-compatible',
    is_enabled      INTEGER DEFAULT 1,
    is_default      INTEGER DEFAULT 0,
    config          TEXT,
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS ai_models (
    id              TEXT PRIMARY KEY,
    provider_id     TEXT REFERENCES ai_providers(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    model_id        TEXT NOT NULL,
    model_type      TEXT DEFAULT 'chat',
    max_tokens      INTEGER,
    supports_stream INTEGER DEFAULT 1,
    supports_tools  INTEGER DEFAULT 0,
    is_enabled      INTEGER DEFAULT 1,
    is_default      INTEGER DEFAULT 0,
    config          TEXT,
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);
