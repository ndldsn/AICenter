-- Agent and sessions
CREATE TABLE IF NOT EXISTS agents (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    description     TEXT,
    model_id        TEXT REFERENCES ai_models(id),
    system_prompt   TEXT,
    temperature     REAL DEFAULT 0.7,
    max_tokens      INTEGER DEFAULT 4096,
    max_iterations  INTEGER DEFAULT 10,
    tools           TEXT DEFAULT '[]',
    tool_permission_mode TEXT DEFAULT 'deny_all',
    require_approval_for TEXT DEFAULT '[]',
    is_enabled      INTEGER DEFAULT 1,
    created_by      TEXT REFERENCES users(id),
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS agent_sessions (
    id              TEXT PRIMARY KEY,
    agent_id        TEXT REFERENCES agents(id),
    user_id         TEXT REFERENCES users(id),
    server_id       TEXT REFERENCES servers(id),
    title           TEXT,
    status          TEXT DEFAULT 'active',
    context_summary TEXT,
    token_input     INTEGER DEFAULT 0,
    token_output    INTEGER DEFAULT 0,
    started_at      TEXT DEFAULT (datetime('now')),
    ended_at        TEXT,
    created_at      TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS agent_messages (
    id              TEXT PRIMARY KEY,
    session_id      TEXT REFERENCES agent_sessions(id) ON DELETE CASCADE,
    role            TEXT NOT NULL,
    content         TEXT,
    tool_call_id    TEXT,
    tool_name       TEXT,
    tool_args       TEXT,
    tool_result     TEXT,
    metadata        TEXT,
    created_at      TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_agent_messages_session ON agent_messages(session_id);
