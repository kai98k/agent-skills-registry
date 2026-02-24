CREATE TABLE IF NOT EXISTS users (
    id          TEXT PRIMARY KEY,
    username    TEXT UNIQUE NOT NULL,
    api_token   TEXT UNIQUE NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tokens (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    token_hash   TEXT UNIQUE NOT NULL,
    token_prefix TEXT NOT NULL,
    expires_at   DATETIME,
    last_used_at DATETIME,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS skills (
    id          TEXT PRIMARY KEY,
    name        TEXT UNIQUE NOT NULL,
    owner_id    TEXT NOT NULL REFERENCES users(id),
    downloads   INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS skill_versions (
    id            TEXT PRIMARY KEY,
    skill_id      TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version       TEXT NOT NULL,
    bundle_key    TEXT NOT NULL,
    metadata      TEXT NOT NULL,
    checksum      TEXT NOT NULL,
    size_bytes    INTEGER NOT NULL,
    published_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(skill_id, version)
);

CREATE INDEX IF NOT EXISTS idx_skill_versions_latest
    ON skill_versions (skill_id, published_at DESC);

CREATE INDEX IF NOT EXISTS idx_skills_name
    ON skills (name);

CREATE INDEX IF NOT EXISTS idx_tokens_hash
    ON tokens (token_hash);

CREATE INDEX IF NOT EXISTS idx_tokens_user
    ON tokens (user_id);

-- Seed data (dev account)
INSERT OR IGNORE INTO users (id, username, api_token)
VALUES ('00000000-0000-0000-0000-000000000001', 'dev', 'dev-token-12345');
