-- PostgreSQL schema for AgentSkills Registry
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(64)  UNIQUE NOT NULL,
    api_token   VARCHAR(128) UNIQUE NOT NULL,
    created_at  TIMESTAMPTZ  DEFAULT now()
);

CREATE TABLE tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         VARCHAR(64) NOT NULL,
    token_hash   VARCHAR(128) UNIQUE NOT NULL,
    token_prefix VARCHAR(32) NOT NULL,
    expires_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT now(),
    UNIQUE(user_id, name)
);

CREATE TABLE skills (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(128) UNIQUE NOT NULL,
    owner_id    UUID NOT NULL REFERENCES users(id),
    downloads   BIGINT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE skill_versions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id      UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version       VARCHAR(32) NOT NULL,
    bundle_key    TEXT NOT NULL,
    metadata      JSONB NOT NULL,
    checksum      VARCHAR(64) NOT NULL,
    size_bytes    BIGINT NOT NULL,
    published_at  TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT uq_skill_version UNIQUE (skill_id, version)
);

CREATE INDEX idx_skill_versions_latest ON skill_versions (skill_id, published_at DESC);
CREATE INDEX idx_skills_name ON skills (name);
CREATE INDEX idx_tokens_hash ON tokens (token_hash);
CREATE INDEX idx_tokens_user ON tokens (user_id);

-- Seed data (dev account)
INSERT INTO users (username, api_token)
VALUES ('dev', 'dev-token-12345');
