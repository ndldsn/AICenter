-- Tasks, approvals, audit logs
CREATE TABLE IF NOT EXISTS tasks (
    id              TEXT PRIMARY KEY,
    title           TEXT NOT NULL,
    description     TEXT,
    type            TEXT DEFAULT 'manual',
    status          TEXT DEFAULT 'pending',
    priority        INTEGER DEFAULT 5,
    server_id       TEXT REFERENCES servers(id),
    agent_id        TEXT REFERENCES agents(id),
    created_by      TEXT REFERENCES users(id),
    scheduled_at    TEXT,
    started_at      TEXT,
    completed_at    TEXT,
    error_message   TEXT,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP),
    updated_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS task_steps (
    id              TEXT PRIMARY KEY,
    task_id         TEXT REFERENCES tasks(id) ON DELETE CASCADE,
    step_number     INTEGER NOT NULL,
    name            TEXT,
    tool_name       TEXT,
    tool_args       TEXT,
    status          TEXT DEFAULT 'pending',
    requires_approval INTEGER DEFAULT 0,
    approved_by     TEXT REFERENCES users(id),
    result          TEXT,
    error_message   TEXT,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS approval_requests (
    id              TEXT PRIMARY KEY,
    request_type    TEXT,
    status          TEXT DEFAULT 'pending',
    requested_by    TEXT REFERENCES users(id),
    requested_at    TEXT DEFAULT (CURRENT_TIMESTAMP),
    approved_by     TEXT REFERENCES users(id),
    approved_at     TEXT,
    tool_name       TEXT,
    tool_args       TEXT,
    risk_level      TEXT,
    dry_run_result  TEXT,
    expires_at      TEXT,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id              TEXT PRIMARY KEY,
    user_id         TEXT REFERENCES users(id),
    username        TEXT,
    action          TEXT NOT NULL,
    resource_type   TEXT,
    resource_id     TEXT,
    resource_name   TEXT,
    method          TEXT,
    path            TEXT,
    ip_address      TEXT,
    user_agent      TEXT,
    status_code     INTEGER,
    request_body    TEXT,
    response_body   TEXT,
    before_state    TEXT,
    after_state     TEXT,
    duration_ms     INTEGER,
    error_message   TEXT,
    server_id       TEXT,
    agent_session_id TEXT,
    approval_id     TEXT REFERENCES approval_requests(id),
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);
