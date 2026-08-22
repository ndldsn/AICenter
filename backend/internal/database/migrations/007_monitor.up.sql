-- Monitor: metrics, alert rules, alert events
CREATE TABLE IF NOT EXISTS monitor_metrics (
    id            TEXT PRIMARY KEY,
    server_id     TEXT,
    metric_name   TEXT NOT NULL,
    value         REAL NOT NULL,
    unit          TEXT,
    labels        TEXT DEFAULT '{}',
    collected_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_monitor_metrics_server ON monitor_metrics(server_id, metric_name, collected_at);
CREATE INDEX IF NOT EXISTS idx_monitor_metrics_time ON monitor_metrics(collected_at);

CREATE TABLE IF NOT EXISTS alert_rules (
    id             TEXT PRIMARY KEY,
    name           TEXT NOT NULL,
    metric_name    TEXT NOT NULL,
    condition      TEXT NOT NULL, -- gt / lt / gte / lte
    threshold      REAL NOT NULL,
    duration       INTEGER DEFAULT 0,   -- seconds the condition must hold
    severity       TEXT DEFAULT 'warning', -- info / warning / critical
    server_id      TEXT,                -- NULL = global (all servers)
    is_enabled     INTEGER DEFAULT 1,
    cooldown       INTEGER DEFAULT 300, -- seconds between repeat alerts
    created_at     TEXT DEFAULT (datetime('now')),
    updated_at     TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS alert_events (
    id           TEXT PRIMARY KEY,
    rule_id      TEXT REFERENCES alert_rules(id) ON DELETE CASCADE,
    rule_name    TEXT,
    server_id    TEXT,
    metric_name  TEXT,
    value        REAL,
    threshold    REAL,
    condition    TEXT,
    severity     TEXT DEFAULT 'warning',
    message      TEXT,
    status       TEXT DEFAULT 'firing',   -- firing / acknowledged / resolved
    triggered_at TEXT DEFAULT (datetime('now')),
    acknowledged_by TEXT,
    acknowledged_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_alert_events_status ON alert_events(status, triggered_at);
