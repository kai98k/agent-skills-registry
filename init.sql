-- init.sql — AgentSkills database schema
-- See reference/SDD.md §4.1

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==========================================
-- USERS
-- ==========================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(64)  UNIQUE NOT NULL,
    api_token       VARCHAR(128) UNIQUE NOT NULL,
    display_name    VARCHAR(128),
    avatar_url      TEXT,
    github_id       BIGINT UNIQUE,
    bio             VARCHAR(256),
    created_at      TIMESTAMPTZ  DEFAULT now()
);

-- ==========================================
-- CATEGORIES
-- ==========================================
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(64)  UNIQUE NOT NULL,
    label       VARCHAR(128) NOT NULL,
    description VARCHAR(256),
    icon        VARCHAR(64),
    sort_order  INT DEFAULT 0
);

-- ==========================================
-- SKILLS (one row per skill name)
-- ==========================================
CREATE TABLE skills (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(128) UNIQUE NOT NULL,
    owner_id      UUID NOT NULL REFERENCES users(id),
    category_id   UUID REFERENCES categories(id),
    downloads     BIGINT DEFAULT 0,
    stars_count   BIGINT DEFAULT 0,
    readme_html   TEXT,
    search_vector TSVECTOR,
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

-- ==========================================
-- SKILL VERSIONS (one row per publish, immutable)
-- ==========================================
CREATE TABLE skill_versions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id      UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version       VARCHAR(32) NOT NULL,
    bundle_key    TEXT NOT NULL,
    metadata      JSONB NOT NULL,
    checksum      VARCHAR(64) NOT NULL,
    size_bytes    BIGINT NOT NULL,
    providers     TEXT[] DEFAULT '{}',
    readme_raw    TEXT,
    published_at  TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT uq_skill_version UNIQUE (skill_id, version)
);

-- ==========================================
-- STARS
-- ==========================================
CREATE TABLE stars (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id   UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),

    PRIMARY KEY (user_id, skill_id)
);

-- ==========================================
-- INDEXES
-- ==========================================
CREATE INDEX idx_skill_versions_latest
    ON skill_versions (skill_id, published_at DESC);

CREATE INDEX idx_skills_name
    ON skills (name);

CREATE INDEX idx_skills_search
    ON skills USING GIN (search_vector);

CREATE INDEX idx_skills_category
    ON skills (category_id);

CREATE INDEX idx_skills_stars
    ON skills (stars_count DESC);

CREATE INDEX idx_skills_downloads
    ON skills (downloads DESC);

CREATE INDEX idx_stars_user
    ON stars (user_id);

CREATE INDEX idx_skill_versions_providers
    ON skill_versions USING GIN (providers);

-- ==========================================
-- TRIGGER: auto-update search_vector
-- ==========================================
CREATE OR REPLACE FUNCTION skills_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.readme_html, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_skills_search_vector
    BEFORE INSERT OR UPDATE OF name, readme_html
    ON skills
    FOR EACH ROW
    EXECUTE FUNCTION skills_search_vector_update();

-- ==========================================
-- SEED DATA
-- ==========================================
-- Dev account for local testing
INSERT INTO users (username, api_token)
VALUES ('dev', 'dev-token-12345');

-- Default categories
INSERT INTO categories (name, label, icon, sort_order) VALUES
    ('development',   'Development',      'code',         1),
    ('productivity',  'Productivity',     'zap',          2),
    ('ai-ml',         'AI & ML',          'brain',        3),
    ('devops',        'DevOps & Infra',   'server',       4),
    ('data',          'Data & Analytics',  'bar-chart',    5),
    ('security',      'Security',         'shield',       6),
    ('testing',       'Testing & QA',     'check-circle', 7),
    ('documentation', 'Documentation',    'file-text',    8),
    ('integration',   'Integration',      'link',         9),
    ('utility',       'Utility',          'wrench',       10);
