-- Servers and groups
CREATE TABLE IF NOT EXISTS server_groups (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    description     TEXT,
    parent_id       TEXT REFERENCES server_groups(id),
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS servers (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    host            TEXT NOT NULL,
    port            INTEGER DEFAULT 22,
    username        TEXT DEFAULT 'root',
    auth_type       TEXT DEFAULT 'password',
    password_enc    TEXT,
    private_key_enc TEXT,
    agent_connected INTEGER DEFAULT 0,
    agent_token     TEXT,
    group_id        TEXT REFERENCES server_groups(id),
    tags            TEXT DEFAULT '[]',
    os_info         TEXT,
    hardware_info   TEXT,
    status          TEXT DEFAULT 'unknown',
    last_heartbeat  TEXT,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP),
    updated_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX IF NOT EXISTS idx_servers_status ON servers(status);
CREATE INDEX IF NOT EXISTS idx_servers_group ON servers(group_id);
