-- H2 batch 2: Role-based access control tables.
--
-- permissions: canonical list of grant names and their group membership.
-- role_permissions: which permission groups each role is granted.
--
-- The permission catalog is also registered in code (internal/permission) at
-- startup; here we persist it so admins can modify assignments at runtime.

CREATE TABLE IF NOT EXISTS permissions (
    id         TEXT PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,
    resource   TEXT NOT NULL,
    action     TEXT NOT NULL,
    group_id   TEXT,
    group_name TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    created_at    TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id),
    FOREIGN KEY (permission_id) REFERENCES permissions(id)
);

CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);