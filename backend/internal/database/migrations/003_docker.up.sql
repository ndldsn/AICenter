-- Docker
CREATE TABLE IF NOT EXISTS docker_hosts (
    id              TEXT PRIMARY KEY,
    server_id       TEXT REFERENCES servers(id) ON DELETE CASCADE,
    name            TEXT,
    socket_path     TEXT DEFAULT '/var/run/docker.sock',
    api_url         TEXT,
    version         TEXT,
    running         INTEGER DEFAULT 1,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS docker_containers (
    id              TEXT PRIMARY KEY,
    docker_host_id  TEXT REFERENCES docker_hosts(id) ON DELETE CASCADE,
    container_id    TEXT NOT NULL,
    name            TEXT NOT NULL,
    image           TEXT,
    state           TEXT,
    status          TEXT,
    last_inspected  TEXT,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP),
    updated_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS docker_images (
    id              TEXT PRIMARY KEY,
    docker_host_id  TEXT REFERENCES docker_hosts(id) ON DELETE CASCADE,
    image_id        TEXT NOT NULL,
    repo_tags       TEXT,
    size            INTEGER,
    last_inspected  TEXT,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS docker_volumes (
    id              TEXT PRIMARY KEY,
    docker_host_id  TEXT REFERENCES docker_hosts(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    mountpoint      TEXT,
    driver          TEXT,
    created_at      TEXT DEFAULT (CURRENT_TIMESTAMP)
);
