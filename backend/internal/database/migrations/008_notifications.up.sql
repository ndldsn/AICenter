-- ============================================
-- 14. 通知系统 (Phase 7)
-- ============================================

-- 通知渠道配置
CREATE TABLE IF NOT EXISTS notification_channels (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    type            TEXT NOT NULL,                 -- webhook, email, sms, im, console
    config          TEXT DEFAULT '{}',             -- 渠道专属配置 (URL/token/收件人等)
    is_enabled      INTEGER DEFAULT 1,
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_channels_type ON notification_channels(type);

-- 通知模板
CREATE TABLE IF NOT EXISTS notification_templates (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    event_type      TEXT NOT NULL,                 -- alert.fired, approval.requested, approval.resolved
    subject         TEXT,                          -- 邮件/IM 标题 (webhook 也可带)
    body            TEXT NOT NULL,                 -- 支持 {{.Var}} 占位符
    channels        TEXT DEFAULT '[]',             -- 默认渠道类型列表 ["email","webhook"]
    is_enabled      INTEGER DEFAULT 1,
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_templates_event ON notification_templates(event_type);

-- 投递日志
CREATE TABLE IF NOT EXISTS notification_delivery_logs (
    id              TEXT PRIMARY KEY,
    channel_id      TEXT REFERENCES notification_channels(id),
    channel_type    TEXT,
    template_id     TEXT,
    event_type      TEXT,
    subject         TEXT,
    body            TEXT,
    status          TEXT DEFAULT 'pending',         -- pending, sent, failed
    error_message   TEXT,
    created_at      TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_delivery_status ON notification_delivery_logs(status);
CREATE INDEX IF NOT EXISTS idx_delivery_event  ON notification_delivery_logs(event_type);

-- 给 alert_rules 增加通知渠道绑定 (JSON 数组, 存放 channel id)
ALTER TABLE alert_rules ADD COLUMN notification_channels TEXT DEFAULT '[]';
