-- Initial migration: Create users table
CREATE TABLE IF NOT EXISTS users (
    id              TEXT PRIMARY KEY,
    username        TEXT UNIQUE NOT NULL,
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT DEFAULT 'viewer',
    is_active       INTEGER DEFAULT 1,
    last_login_at   TEXT,
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS roles (
    id              TEXT PRIMARY KEY,
    name            TEXT UNIQUE NOT NULL,
    description     TEXT,
    is_system       INTEGER DEFAULT 0,
    created_at      TEXT DEFAULT (datetime('now'))
);

-- Insert default roles
INSERT INTO roles (id, name, description, is_system) VALUES
    ('role-superadmin', 'superadmin', 'Super Administrator with all permissions', 1),
    ('role-admin', 'admin', 'Administrator with server and user management', 1),
    ('role-operator', 'operator', 'Operator with daily operations', 1),
    ('role-viewer', 'viewer', 'Read-only viewer', 1);

-- Insert default admin user (password: admin123)
INSERT INTO users (id, username, email, password_hash, role) VALUES
    ('user-admin', 'admin', 'admin@aicenter.local', '$2a$10$rI8eQj3J3q3J3q3J3q3J3O', 'superadmin');
