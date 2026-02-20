# AgentSkills — Software Design Document (SDD)

**Version:** 2.0 (Full Platform)
**Date:** 2026-02-20
**Author:** LIU YU KAI
**Purpose:** 本文件為 Claude Code 的開發藍圖。請嚴格依照本文件的規格、目錄結構、API Contract 與 DB Schema 進行開發與驗證。

---

## 1. 專案概述

AgentSkills 是一個 AI Agent Skill 的集中式 Registry 平台，類似 npm 或 Docker Hub，但專為 AI Agent Skills 設計。開發者可透過 **Web UI** 瀏覽與發現 Skills，透過 **CLI** 工具上傳（push）與下載（pull）標準化的 Skill Bundle，平台負責版本控制、Metadata 解析、全文搜尋與檔案儲存。

**平台範圍包含：**

- **Web 前端**：Next.js App Router（瀏覽、搜尋、Skill 詳情、使用者頁面）
- **後端 API**：FastAPI（Python）
- **CLI 工具**：Go + Cobra
- **資料庫**：PostgreSQL（含全文搜尋 `tsvector`）
- **物件儲存**：MinIO（S3 相容）
- **認證**：GitHub OAuth（Web）+ Bearer Token（CLI/API）
- **基礎設施**：Docker Compose 本地開發環境

**本版本明確不包含：**

- Semver range resolution（`^1.0.0`）
- Skill 之間的依賴關係
- unpublish / deprecate 功能
- 自動版本號 bump
- 向量語意搜尋（使用 PostgreSQL 全文搜尋替代）
- 付費 / 私有 Skills

---

## 2. Skill Bundle 標準

每個 Skill 是一個目錄，打包為 `.tar.gz` 上傳。

### 2.1 目錄結構

```
my-skill/
├── SKILL.md         (必填) 核心定義：YAML Frontmatter + Markdown 指令
├── scripts/         (選填) Agent 可呼叫的腳本
├── references/      (選填) RAG / Few-shot 參考文件
└── assets/          (選填) 靜態模板與資源
```

### 2.2 SKILL.md Frontmatter 規格

```yaml
---
name: "code-review-agent"           # 必填, 全域唯一, 格式: [a-z0-9\-], 3-64 字元
version: "1.0.0"                    # 必填, 嚴格 semver (MAJOR.MINOR.PATCH)
description: "PR code review skill" # 必填, 最長 256 字元
author: "liuyukai"                  # 必填, 與上傳者帳號一致
tags:                               # 選填, 最多 10 個, 每個最長 32 字元
  - code-review
  - github
license: "MIT"                      # 選填, SPDX identifier
min_agent_version: ">=0.1.0"        # 選填, 保留欄位 (MVP 不驗證)
---

# Code Review Agent

以下為 Markdown 格式的 Skill 指令內容...
```

### 2.3 Frontmatter 驗證規則

| 欄位 | 類型 | 必填 | 驗證規則 |
|------|------|------|----------|
| name | string | ✅ | `/^[a-z0-9\-]{3,64}$/`，不允許連續 `--` |
| version | string | ✅ | 嚴格 semver，使用 Python `semver` 套件驗證 |
| description | string | ✅ | 1-256 字元 |
| author | string | ✅ | 必須與 API Token 對應的 username 一致 |
| tags | list[string] | ❌ | 最多 10 個，每個 `/^[a-z0-9\-]{1,32}$/` |
| license | string | ❌ | 若提供需為合法 SPDX identifier |
| min_agent_version | string | ❌ | MVP 階段僅儲存，不做邏輯判斷 |

---

## 3. 系統架構

```
┌─────────────────────────────────────────────────┐
│                docker-compose                    │
│                                                  │
│  ┌─────────────┐  ┌──────────┐  ┌────────────┐  │
│  │ PostgreSQL   │  │ MinIO    │  │ Next.js    │  │
│  │ port: 5432   │  │ API:9000 │  │ port: 3000 │  │
│  │              │  │ UI: 9001 │  │ (SSR)      │  │
│  └──────┬───────┘  └────┬─────┘  └─────┬──────┘  │
│         │               │              │         │
│         └───────┬───────┘              │         │
│            ┌────┴────┐                 │         │
│            │ FastAPI  │◄───────────────┘         │
│            │ port:8000│                          │
│            └────┬─────┘                          │
│                 │                                │
└─────────────────┼────────────────────────────────┘
                  │
       ┌──────────┼──────────┐
       │          │          │
┌──────┴──┐  ┌───┴──────┐  ┌┴──────────┐
│ Browser  │  │  Go CLI  │  │ External  │
│ (直連    │  │ (本機)    │  │ Clients   │
│  Next.js)│  └──────────┘  └───────────┘
└─────────┘
```

### 3.1 元件職責

| 元件 | 職責 | 技術 |
|------|------|------|
| **Next.js** | Web 前端：SSR 頁面、GitHub OAuth、Skill 瀏覽/搜尋 | Next.js 15+, React 19, Tailwind CSS, shadcn/ui |
| **FastAPI** | 核心 API：Skill CRUD、Bundle 上傳下載、搜尋、認證 | Python 3.12+, async SQLAlchemy, boto3 |
| **Go CLI** | 開發者工具：push/pull/search/init/login | Go 1.22+, Cobra, Viper |
| **PostgreSQL** | 資料持久化、全文搜尋（`tsvector`）、JSONB metadata | PostgreSQL 16 |
| **MinIO** | Skill Bundle (.tar.gz) 物件儲存 | S3 相容 API |

### 3.2 資料流

```
Web UI 使用者流程:
Browser → Next.js (SSR) → FastAPI API → PostgreSQL / MinIO

CLI 使用者流程:
Terminal → Go CLI → FastAPI API → PostgreSQL / MinIO

認證流程 (Web):
Browser → Next.js → GitHub OAuth → FastAPI (驗證/建立使用者) → JWT cookie

認證流程 (CLI):
Terminal → agentskills login → 儲存 Bearer Token → 後續 API 呼叫帶 Token
```

---

## 4. 資料庫設計 (PostgreSQL)

### 4.1 Schema

```sql
-- init.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==========================================
-- USERS
-- ==========================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(64)  UNIQUE NOT NULL,
    api_token       VARCHAR(128) UNIQUE NOT NULL,
    display_name    VARCHAR(128),                        -- 顯示名稱（可選）
    avatar_url      TEXT,                                -- 頭像 URL（GitHub 頭像）
    github_id       BIGINT UNIQUE,                       -- GitHub user ID（OAuth 登入）
    bio             VARCHAR(256),                        -- 個人簡介
    created_at      TIMESTAMPTZ  DEFAULT now()
);

-- ==========================================
-- CATEGORIES (技能分類)
-- ==========================================
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(64)  UNIQUE NOT NULL,            -- e.g. "development", "productivity"
    label       VARCHAR(128) NOT NULL,                   -- e.g. "Development", "Productivity"
    description VARCHAR(256),
    icon        VARCHAR(64),                             -- icon name, e.g. "code", "zap"
    sort_order  INT DEFAULT 0
);

-- ==========================================
-- SKILLS (一個 name 一筆)
-- ==========================================
CREATE TABLE skills (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(128) UNIQUE NOT NULL,
    owner_id      UUID NOT NULL REFERENCES users(id),
    category_id   UUID REFERENCES categories(id),        -- 所屬分類（可選）
    downloads     BIGINT DEFAULT 0,
    stars_count   BIGINT DEFAULT 0,                      -- 冗餘計數，快速排序用
    readme_html   TEXT,                                  -- 最新版 SKILL.md body 渲染後的 HTML（快取）
    search_vector TSVECTOR,                              -- 全文搜尋向量
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

-- ==========================================
-- SKILL VERSIONS (每次 publish 一筆, immutable)
-- ==========================================
CREATE TABLE skill_versions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id      UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version       VARCHAR(32) NOT NULL,
    bundle_key    TEXT NOT NULL,           -- MinIO object key, e.g. "code-review-agent/1.0.0.tar.gz"
    metadata      JSONB NOT NULL,          -- 完整 frontmatter
    checksum      VARCHAR(64) NOT NULL,    -- SHA-256 hex digest
    size_bytes    BIGINT NOT NULL,         -- bundle 檔案大小
    providers     TEXT[] DEFAULT '{}',     -- 支援的 Agent 平台
    readme_raw    TEXT,                    -- SKILL.md markdown body（此版本）
    published_at  TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT uq_skill_version UNIQUE (skill_id, version)
);

-- ==========================================
-- STARS (使用者收藏)
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
-- TRIGGER: 自動更新 search_vector
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
-- 開發用測試帳號
INSERT INTO users (username, api_token)
VALUES ('dev', 'dev-token-12345');

-- 預設分類（參考 ClawHub 分類架構）
INSERT INTO categories (name, label, icon, sort_order) VALUES
    ('development',   'Development',     'code',        1),
    ('productivity',  'Productivity',    'zap',         2),
    ('ai-ml',         'AI & ML',         'brain',       3),
    ('devops',        'DevOps & Infra',  'server',      4),
    ('data',          'Data & Analytics', 'bar-chart',   5),
    ('security',      'Security',        'shield',      6),
    ('testing',       'Testing & QA',    'check-circle', 7),
    ('documentation', 'Documentation',   'file-text',   8),
    ('integration',   'Integration',     'link',        9),
    ('utility',       'Utility',         'wrench',      10);
```

### 4.2 設計決策

- **兩表分離**：`skills` 存身份與聚合資料（downloads, stars_count），`skill_versions` 存每次發布的 immutable 記錄。
- **Immutable publish**：同一 `(skill_id, version)` 不可覆寫，嘗試重複發布回傳 `409 Conflict`。
- **Soft latest**：最新版透過 `published_at DESC LIMIT 1` 查詢，不額外維護 `latest` 欄位。
- **JSONB metadata**：frontmatter 全文存入，支援未來擴充欄位時不需 migration。
- **PostgreSQL 全文搜尋**：使用 `tsvector` + `GIN` 索引，權重分配：name (A) > description (B) > readme (C)，免除外部搜尋引擎依賴。
- **Stars 計數冗餘**：`skills.stars_count` 由應用層維護（star/unstar 時 +1/-1），避免每次排序都 JOIN + COUNT。
- **Categories**：預設 10 個分類，Skill publish 時可選擇分類（CLI 或 Web UI）。
- **GitHub OAuth 欄位**：`users` 表增加 `github_id`, `display_name`, `avatar_url`, `bio`，支援 Web OAuth 登入自動建立帳號。
- **README 快取**：`skills.readme_html` 快取最新版 SKILL.md body 的渲染結果，publish 時更新，避免每次 SSR 都即時渲染。

---

## 5. API 設計 (FastAPI)

**Base URL:** `http://localhost:8000/v1`

### 5.1 認證

支援兩種認證方式：

#### 5.1.1 Bearer Token（CLI / API 直接呼叫）

```
Authorization: Bearer dev-token-12345
```

靜態 API Token，適用於 CLI 和程式化 API 呼叫。認證失敗回傳 `401 Unauthorized`。

#### 5.1.2 GitHub OAuth（Web UI）

Web 前端透過 NextAuth.js 處理 GitHub OAuth 流程：

1. 使用者點擊「Sign in with GitHub」
2. 重導至 GitHub 授權頁
3. GitHub callback → NextAuth.js → 呼叫 FastAPI `POST /v1/auth/github`
4. FastAPI 用 GitHub access token 取得使用者資料
5. 自動建立或更新 `users` 記錄（以 `github_id` 為識別）
6. 回傳 API Token → NextAuth.js 存入 session/cookie

兩種方式共用 `users` 表，GitHub OAuth 使用者同樣擁有 `api_token`，可在 Web UI 設定頁複製 Token 供 CLI 使用。

#### 5.1.3 認證規則

| 端點類型 | 認證要求 |
|----------|----------|
| `GET` 查詢類 | 不需要認證（公開） |
| `POST /v1/skills/publish` | 需要 Bearer Token |
| `POST /v1/skills/{name}/star` | 需要 Bearer Token |
| `DELETE /v1/skills/{name}/star` | 需要 Bearer Token |
| `POST /v1/auth/github` | 不需要（用 GitHub token 換 API token） |

### 5.2 端點規格

#### `POST /v1/skills/publish`

上傳一個 Skill Bundle。

**Request:**

- Header: `Authorization: Bearer <token>`
- Body: `multipart/form-data`
  - `file`: `.tar.gz` 檔案 (最大 50MB)

**Server 端處理流程：**

1. 驗證 API Token → 取得 `user`
2. 解壓縮 `.tar.gz` 至暫存目錄
3. 找到並解析 `SKILL.md` 的 YAML Frontmatter
4. 執行 Frontmatter 驗證（見 §2.3）
5. 確認 `author` == `user.username`
6. 計算整個 `.tar.gz` 的 SHA-256 checksum
7. 查詢 DB：若 `name` 不存在 → 新建 `skills` 記錄（owner = user）
8. 查詢 DB：若 `name` 存在但 `owner_id != user.id` → 403 Forbidden
9. 查詢 DB：若 `(skill_id, version)` 已存在 → 409 Conflict
10. 上傳 `.tar.gz` 至 MinIO → key: `{name}/{version}.tar.gz`
11. 寫入 `skill_versions` 記錄
12. 更新 `skills.updated_at`

**Success Response:** `201 Created`

```json
{
  "name": "code-review-agent",
  "version": "1.0.0",
  "checksum": "sha256:a1b2c3d4...",
  "published_at": "2026-02-20T10:00:00Z"
}
```

**Error Responses:**

| Code | Condition | Body |
|------|-----------|------|
| 400 | 無 SKILL.md / Frontmatter 驗證失敗 / 非 .tar.gz | `{"error": "具體錯誤訊息"}` |
| 401 | Token 無效或缺少 | `{"error": "Unauthorized"}` |
| 403 | name 已被其他使用者佔用 | `{"error": "Skill 'x' is owned by another user"}` |
| 409 | 版本已存在 | `{"error": "Version 1.0.0 already exists"}` |
| 413 | 檔案超過 50MB | `{"error": "Bundle exceeds 50MB limit"}` |

---

#### `GET /v1/skills/{name}`

取得 Skill 資訊與最新版本。

**Success Response:** `200 OK`

```json
{
  "name": "code-review-agent",
  "owner": "liuyukai",
  "downloads": 42,
  "created_at": "2026-02-20T10:00:00Z",
  "latest_version": {
    "version": "1.2.0",
    "description": "PR code review skill",
    "checksum": "sha256:a1b2c3d4...",
    "size_bytes": 15360,
    "published_at": "2026-02-20T12:00:00Z",
    "metadata": { ... }
  }
}
```

**Error:** `404 Not Found` 若 name 不存在。

---

#### `GET /v1/skills/{name}/versions`

列出 Skill 所有版本。

**Success Response:** `200 OK`

```json
{
  "name": "code-review-agent",
  "versions": [
    {
      "version": "1.2.0",
      "checksum": "sha256:...",
      "size_bytes": 15360,
      "published_at": "2026-02-20T12:00:00Z"
    },
    {
      "version": "1.0.0",
      "checksum": "sha256:...",
      "size_bytes": 12288,
      "published_at": "2026-02-20T10:00:00Z"
    }
  ]
}
```

---

#### `GET /v1/skills/{name}/versions/{version}/download`

下載指定版本的 Bundle。

**行為：** 回傳 MinIO presigned URL 做 302 Redirect，或直接串流檔案內容（MVP 用串流較簡單）。

**Response:** `200 OK`

- `Content-Type: application/gzip`
- `Content-Disposition: attachment; filename="code-review-agent-1.0.0.tar.gz"`
- `X-Checksum-SHA256: a1b2c3d4...`
- Body: raw binary

**Side effect:** `skills.downloads += 1`

**Error:** `404` 若 name 或 version 不存在。

---

#### `GET /v1/skills?q={keyword}&tag={tag}&category={cat}&sort={sort}&page={n}&per_page={n}`

搜尋 Skills。

**Query Parameters:**

| Param | Type | Default | 說明 |
|-------|------|---------|------|
| q | string | - | 全文搜尋（使用 PostgreSQL `tsvector`，匹配 name, description, readme） |
| tag | string | - | 精確匹配 metadata tags（可多次傳遞） |
| category | string | - | 按分類篩選（category name） |
| sort | string | `relevance` | 排序方式：`relevance`（搜尋相關度）, `downloads`, `stars`, `newest`, `updated` |
| page | int | 1 | 頁碼 |
| per_page | int | 20 | 每頁數量，最大 100 |

**搜尋實作邏輯：**

```sql
-- 當有 q 參數時，使用 tsvector 全文搜尋 + 排名
SELECT s.*, ts_rank(s.search_vector, plainto_tsquery('english', :q)) AS rank
FROM skills s
WHERE s.search_vector @@ plainto_tsquery('english', :q)
ORDER BY rank DESC;

-- 無 q 時，按 sort 參數排序
-- sort=downloads → ORDER BY downloads DESC
-- sort=stars     → ORDER BY stars_count DESC
-- sort=newest    → ORDER BY created_at DESC
-- sort=updated   → ORDER BY updated_at DESC
```

**Success Response:** `200 OK`

```json
{
  "total": 45,
  "page": 1,
  "per_page": 20,
  "results": [
    {
      "name": "code-review-agent",
      "description": "PR code review skill",
      "owner": "liuyukai",
      "owner_avatar_url": "https://avatars.githubusercontent.com/u/12345",
      "downloads": 42,
      "stars_count": 15,
      "latest_version": "1.2.0",
      "category": "development",
      "updated_at": "2026-02-20T12:00:00Z",
      "tags": ["code-review", "github"]
    }
  ]
}
```

---

#### `GET /v1/skills/{name}`（更新版）

取得 Skill 資訊與最新版本，增加 stars、category、readme 欄位。

**Success Response:** `200 OK`

```json
{
  "name": "code-review-agent",
  "owner": "liuyukai",
  "owner_avatar_url": "https://avatars.githubusercontent.com/u/12345",
  "downloads": 42,
  "stars_count": 15,
  "starred_by_me": false,
  "category": "development",
  "readme_html": "<h1>Code Review Agent</h1><p>...</p>",
  "created_at": "2026-02-20T10:00:00Z",
  "latest_version": {
    "version": "1.2.0",
    "description": "PR code review skill",
    "checksum": "sha256:a1b2c3d4...",
    "size_bytes": 15360,
    "published_at": "2026-02-20T12:00:00Z",
    "metadata": { ... }
  }
}
```

> 注意：`starred_by_me` 僅在請求帶有有效 Auth header 時計算，否則為 `false`。

---

#### `POST /v1/skills/{name}/star`

收藏 Skill。需要認證。

**Response:** `200 OK`

```json
{
  "starred": true,
  "stars_count": 16
}
```

**Error:** `404` 若 skill 不存在，`409` 若已收藏。

---

#### `DELETE /v1/skills/{name}/star`

取消收藏。需要認證。

**Response:** `200 OK`

```json
{
  "starred": false,
  "stars_count": 15
}
```

---

#### `GET /v1/categories`

列出所有分類及各分類的 skill 數量。

**Response:** `200 OK`

```json
{
  "categories": [
    {
      "name": "development",
      "label": "Development",
      "icon": "code",
      "skill_count": 128
    },
    {
      "name": "productivity",
      "label": "Productivity",
      "icon": "zap",
      "skill_count": 85
    }
  ]
}
```

---

#### `POST /v1/auth/github`

GitHub OAuth 登入/註冊。由 Next.js 前端在 OAuth callback 後呼叫。

**Request:**

```json
{
  "github_access_token": "gho_xxxxxxxxxxxx"
}
```

**Server 端處理流程：**

1. 使用 `github_access_token` 呼叫 GitHub API `GET /user` 取得使用者資料
2. 以 `github_id` 查詢 `users` 表
3. 若不存在 → 建立新使用者（username = GitHub login, 自動產生 api_token）
4. 若已存在 → 更新 `display_name`, `avatar_url`
5. 回傳使用者資訊與 `api_token`

**Response:** `200 OK`

```json
{
  "username": "liuyukai",
  "display_name": "Liu Yu Kai",
  "avatar_url": "https://avatars.githubusercontent.com/u/12345",
  "api_token": "ask-xxxxxxxxxxxxxxxx"
}
```

---

#### `GET /v1/users/{username}`

取得使用者公開資料與其發布的 Skills。

**Response:** `200 OK`

```json
{
  "username": "liuyukai",
  "display_name": "Liu Yu Kai",
  "avatar_url": "https://avatars.githubusercontent.com/u/12345",
  "bio": "Backend developer, AI enthusiast",
  "created_at": "2026-01-15T10:00:00Z",
  "skills": [
    {
      "name": "code-review-agent",
      "description": "PR code review skill",
      "downloads": 42,
      "stars_count": 15,
      "latest_version": "1.2.0",
      "updated_at": "2026-02-20T12:00:00Z"
    }
  ],
  "total_downloads": 156,
  "total_stars": 47
}
```

**Error:** `404` 若使用者不存在。

---

#### `GET /v1/health`

健康檢查端點。

**Response:** `200 OK`

```json
{
  "status": "ok",
  "database": "connected",
  "storage": "connected"
}
```

---

## 6. 後端專案結構 (Python / FastAPI)

```
api/
├── app/
│   ├── __init__.py
│   ├── main.py               # FastAPI app 入口, lifespan, CORS middleware
│   ├── config.py              # pydantic-settings, 環境變數讀取
│   ├── dependencies.py        # Depends: get_db, get_current_user, get_s3, get_optional_user
│   ├── models.py              # SQLAlchemy ORM models (對應 §4.1 schema)
│   ├── schemas.py             # Pydantic request/response schemas
│   ├── routes/
│   │   ├── __init__.py
│   │   ├── skills.py          # /v1/skills CRUD + search + star 端點
│   │   ├── auth.py            # /v1/auth/github OAuth 端點
│   │   ├── categories.py      # /v1/categories 端點
│   │   ├── users.py           # /v1/users/{username} 端點
│   │   └── health.py          # /v1/health
│   └── services/
│       ├── __init__.py
│       ├── storage.py          # MinIO/S3 上傳、下載、presigned URL
│       ├── parser.py           # .tar.gz 解壓、SKILL.md 解析、YAML 驗證
│       ├── auth.py             # API Token 驗證 + GitHub OAuth 邏輯
│       └── markdown.py         # SKILL.md Markdown → HTML 渲染（安全過濾）
├── tests/
│   ├── conftest.py            # pytest fixtures: test DB, test S3, test client
│   ├── test_publish.py        # publish 端點完整測試
│   ├── test_pull.py           # download 端點測試
│   ├── test_search.py         # search 端點測試（含全文搜尋）
│   ├── test_parser.py         # SKILL.md 解析與驗證測試
│   ├── test_stars.py          # star/unstar 測試
│   └── test_auth.py           # GitHub OAuth 測試
├── requirements.txt
└── Dockerfile
```

### 6.1 核心依賴 (requirements.txt)

```
fastapi>=0.115.0
uvicorn[standard]>=0.30.0
sqlalchemy[asyncio]>=2.0
asyncpg>=0.30.0
pydantic>=2.0
pydantic-settings>=2.0
boto3>=1.35.0
python-multipart>=0.0.9
pyyaml>=6.0
semver>=3.0
python-frontmatter>=1.1
httpx>=0.27.0
markdown>=3.7                  # SKILL.md → HTML 渲染
bleach>=6.0                    # HTML 安全過濾（防 XSS）
pygments>=2.18                 # Markdown code block 語法高亮
pytest>=8.0
pytest-asyncio>=0.24.0
```

### 6.2 config.py 規格

```python
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # Database
    database_url: str = "postgresql+asyncpg://dev:devpass@localhost:5432/agentskills"

    # MinIO / S3
    s3_endpoint: str = "http://localhost:9000"
    s3_access_key: str = "minioadmin"
    s3_secret_key: str = "minioadmin"
    s3_bucket: str = "skills"
    s3_region: str = "us-east-1"

    # GitHub OAuth
    github_client_id: str = ""
    github_client_secret: str = ""

    # CORS (允許 Next.js 前端)
    cors_origins: list[str] = ["http://localhost:3000"]

    # App
    max_bundle_size: int = 50 * 1024 * 1024  # 50MB
    api_prefix: str = "/v1"

    class Config:
        env_file = ".env"
```

### 6.3 parser.py 核心邏輯

```python
"""
SKILL.md 解析流程：
1. 接收上傳的 .tar.gz bytes
2. 解壓至暫存目錄 (tempfile.mkdtemp)
3. 遍歷解壓後的檔案，找到 SKILL.md（必須在根目錄或一層子目錄內）
4. 使用 python-frontmatter 解析 YAML + Markdown body
5. 對 YAML 執行 §2.3 驗證規則
6. 回傳 ParsedSkill dataclass

注意事項：
- 解壓時檢查 zip bomb（解壓後總大小不超過 200MB）
- 路徑穿越攻擊防護（所有解壓路徑必須在暫存目錄下）
- 暫存目錄用完即刪
"""
```

### 6.4 storage.py 核心邏輯

```python
"""
使用 boto3 連接 MinIO（S3 相容 API）。

關鍵設定（確保 MinIO 相容）：
- endpoint_url: 必須設定為 config.s3_endpoint
- 使用 path-style addressing（MinIO 不支援 virtual-hosted-style）

核心方法：
- upload_bundle(name, version, file_bytes) -> bundle_key
- download_bundle(bundle_key) -> StreamingResponse
- check_health() -> bool
"""
```

---

## 7. CLI 設計 (Go / Cobra)

### 7.1 專案結構

```
cli/
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go         # Cobra root command, global flags
│   ├── init_cmd.go     # agentskills init
│   ├── push.go         # agentskills push
│   ├── pull.go         # agentskills pull
│   ├── search.go       # agentskills search
│   └── login.go        # agentskills login (存 token 到 config)
├── internal/
│   ├── config/
│   │   └── config.go   # 讀取 ~/.agentskills/config.yaml
│   ├── api/
│   │   └── client.go   # HTTP client, 封裝所有 API 呼叫
│   ├── bundle/
│   │   └── pack.go     # tar.gz 打包與解壓邏輯
│   └── parser/
│       └── frontmatter.go  # 本地 SKILL.md 驗證 (push 前預檢)
└── Makefile            # build targets for linux/darwin/windows
```

### 7.2 指令規格

#### `agentskills init [name]`

在當前目錄建立 Skill 骨架。

```bash
$ agentskills init my-new-skill

Created my-new-skill/
  ├── SKILL.md        (已填入模板 frontmatter)
  ├── scripts/
  ├── references/
  └── assets/
```

SKILL.md 模板：

```yaml
---
name: "my-new-skill"
version: "0.1.0"
description: ""
author: ""
tags: []
---

# my-new-skill

Describe your skill here.
```

#### `agentskills push [path]`

打包並上傳 Skill Bundle。

```bash
$ agentskills push ./my-skill

Validating SKILL.md...        ✓
Packing bundle...             ✓ (12.3 KB)
Uploading my-skill@1.0.0...   ✓
Checksum: sha256:a1b2c3d4...

Published my-skill@1.0.0 successfully.
```

**流程：**

1. 讀取 `path/SKILL.md`，本地解析並驗證 frontmatter
2. 將整個目錄打包為 `.tar.gz`（排除 `.git`, `node_modules`, `__pycache__`）
3. 計算 SHA-256
4. POST 至 `/v1/skills/publish`
5. 驗證 server 回傳的 checksum 與本地一致
6. 輸出結果

**排除清單 (hardcoded)：**

```
.git/
.DS_Store
node_modules/
__pycache__/
*.pyc
.env
```

#### `agentskills pull <name>[@version]`

下載 Skill Bundle 並解壓至當前目錄。

```bash
$ agentskills pull code-review-agent
Downloading code-review-agent@1.2.0 (latest)...  ✓
Verifying checksum...                              ✓
Extracted to ./code-review-agent/

$ agentskills pull code-review-agent@1.0.0
Downloading code-review-agent@1.0.0...            ✓
Verifying checksum...                              ✓
Extracted to ./code-review-agent/
```

**流程：**

1. 解析 `name` 和可選的 `@version`
2. 若無 version → `GET /v1/skills/{name}` 取 latest version
3. `GET /v1/skills/{name}/versions/{version}/download` 下載 `.tar.gz`
4. 驗證 `X-Checksum-SHA256` header 與下載內容的 SHA-256 一致
5. 解壓至 `./{name}/`（若目錄已存在，提示覆蓋確認）

#### `agentskills search <keyword>`

搜尋平台上的 Skills。

```bash
$ agentskills search code-review

NAME                  VERSION  DOWNLOADS  DESCRIPTION
code-review-agent     1.2.0    42         PR code review skill
code-review-lite      0.3.0    7          Lightweight review helper
```

#### `agentskills login`

儲存 API Token 至本地設定。

```bash
$ agentskills login
Enter API token: ********
Token saved to ~/.agentskills/config.yaml
```

### 7.3 本地設定檔

路徑: `~/.agentskills/config.yaml`

```yaml
api_url: "http://localhost:8000"
token: "dev-token-12345"
```

### 7.4 核心依賴 (go.mod)

```
module github.com/liuyukai/agentskills-cli

go 1.22

require (
    github.com/spf13/cobra v1.8+
    github.com/spf13/viper v1.19+  // config 讀取
    gopkg.in/yaml.v3 v3.0+         // frontmatter 解析
)
```

---

## 8. Docker Compose 開發環境

### 8.1 docker-compose.yml

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: agentskills
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: devpass
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dev -d agentskills"]
      interval: 5s
      timeout: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - miniodata:/data
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 5s
      timeout: 5s
      retries: 5

  minio-init:
    image: minio/mc:latest
    depends_on:
      minio:
        condition: service_healthy
    entrypoint: >
      /bin/sh -c "
      mc alias set local http://minio:9000 minioadmin minioadmin;
      mc mb --ignore-existing local/skills;
      echo 'Bucket skills created';
      "

volumes:
  pgdata:
  miniodata:
```

> **注意**：Next.js 前端在開發時使用 `npm run dev`（port 3000），不包含在 docker-compose 中。生產部署時可另建 Dockerfile。

### 8.2 啟動與驗證指令

```bash
# 1. 啟動基礎設施
docker compose up -d

# 2. 驗證 PostgreSQL
docker compose exec postgres psql -U dev -d agentskills -c "SELECT COUNT(*) FROM users;"
# 預期輸出: 1 (dev 帳號)

# 3. 驗證 MinIO
curl -s http://localhost:9000/minio/health/live
# 預期輸出: HTTP 200

# 4. 啟動 FastAPI (開發模式)
cd api && uvicorn app.main:app --reload --port 8000

# 5. 驗證 API
curl http://localhost:8000/v1/health
# 預期輸出: {"status":"ok","database":"connected","storage":"connected"}

# 6. 完整 publish 測試
mkdir -p /tmp/test-skill && cat > /tmp/test-skill/SKILL.md << 'EOF'
---
name: "test-skill"
version: "0.1.0"
description: "A test skill for validation"
author: "dev"
tags:
  - test
---

# Test Skill

This is a test.
EOF

cd /tmp && tar -czf test-skill.tar.gz -C test-skill .
curl -X POST http://localhost:8000/v1/skills/publish \
  -H "Authorization: Bearer dev-token-12345" \
  -F "file=@test-skill.tar.gz"
# 預期: 201 Created

# 7. 驗證 pull
curl http://localhost:8000/v1/skills/test-skill
# 預期: 200 OK with latest_version.version == "0.1.0"

curl -O http://localhost:8000/v1/skills/test-skill/versions/0.1.0/download
# 預期: 下載 .tar.gz 檔案

# 8. 驗證 immutable publish (重複版本)
curl -X POST http://localhost:8000/v1/skills/publish \
  -H "Authorization: Bearer dev-token-12345" \
  -F "file=@test-skill.tar.gz"
# 預期: 409 Conflict

# 9. Go CLI 測試 (API 啟動後)
cd cli && go run main.go push /tmp/test-skill
go run main.go pull test-skill
go run main.go search test
```

---

## 9. 測試策略

### 9.1 後端測試 (pytest)

**conftest.py 需提供：**

- `test_db`：使用 SQLite async (`aiosqlite`) 作為測試資料庫，每個 test function 自動 rollback
- `test_s3`：mock boto3 client 或使用 `moto` library 模擬 S3
- `test_client`：FastAPI `TestClient`，注入 test_db 和 test_s3

**必要測試案例：**

| 測試檔案 | 案例 | 預期 |
|----------|------|------|
| test_parser.py | 合法 SKILL.md | 正確解析所有欄位 |
| test_parser.py | 缺少 name 欄位 | ValidationError |
| test_parser.py | version 非 semver | ValidationError |
| test_parser.py | name 含大寫或特殊字元 | ValidationError |
| test_parser.py | 無 SKILL.md 的 tar.gz | FileNotFoundError |
| test_publish.py | 正常 publish | 201, DB 有記錄, MinIO 有檔案 |
| test_publish.py | 無 auth header | 401 |
| test_publish.py | 重複版本 | 409 |
| test_publish.py | name 被他人佔用 | 403 |
| test_publish.py | 超過 50MB | 413 |
| test_pull.py | 下載 latest | 200, 正確 binary |
| test_pull.py | 下載指定版本 | 200, checksum 正確 |
| test_pull.py | 不存在的 skill | 404 |
| test_search.py | keyword 搜尋 | 回傳匹配結果 |
| test_search.py | tag 篩選 | 僅回傳有該 tag 的結果 |
| test_search.py | category 篩選 | 僅回傳該分類結果 |
| test_search.py | sort 排序 | 按指定欄位排序 |
| test_search.py | 全文搜尋 tsvector | 使用 plainto_tsquery 正確匹配 |
| test_search.py | 空結果 | 200, results: [] |
| test_stars.py | star 一個 skill | 200, stars_count +1 |
| test_stars.py | unstar 一個 skill | 200, stars_count -1 |
| test_stars.py | 重複 star | 409 |
| test_stars.py | 未認證 star | 401 |
| test_auth.py | GitHub OAuth 新使用者 | 自動建立帳號，回傳 api_token |
| test_auth.py | GitHub OAuth 既有使用者 | 更新 profile，回傳相同 api_token |

### 9.2 CLI 測試 (go test)

- 本地 frontmatter 解析與驗證
- tar.gz 打包排除清單
- SHA-256 checksum 計算
- API client 呼叫 (使用 httptest mock server)
- config 檔讀寫

### 9.3 前端測試 (vitest + playwright)

- **Unit tests (vitest)**：API client functions、utility functions
- **Component tests (vitest + testing-library)**：SkillCard、SearchBar、MarkdownRenderer
- **E2E tests (playwright)**：首頁瀏覽、搜尋流程、Skill 詳情頁、登入流程

---

## 10. 安全性注意事項

| 威脅 | 防護措施 |
|------|----------|
| Zip bomb | 解壓時限制總大小 200MB，超過即中止 |
| 路徑穿越 (../../etc/passwd) | 所有解壓路徑檢查必須在暫存目錄下 |
| 任意檔案執行 | Server 端僅解析 SKILL.md，不執行 scripts/ 內任何檔案 |
| Token 洩漏 | CLI config 檔設 0600 權限；API logs 不記錄完整 token |
| SQL Injection | 使用 SQLAlchemy ORM parameterized queries |
| 超大檔案 DoS | FastAPI 層限制 request body 50MB |
| XSS (Markdown 渲染) | 使用 `bleach` 過濾 HTML，僅允許安全 tags (p, h1-h6, a, code, pre, ul, ol, li, strong, em, img) |
| CSRF | Next.js + NextAuth.js 內建 CSRF token 保護 |
| OAuth Token 劫持 | GitHub OAuth state 參數驗證 + HTTPS（生產環境） |
| CORS 錯誤設定 | FastAPI CORS middleware 限定允許的 origins（見 §6.2） |

---

## 11. 開發順序建議

以下為建議的實作優先順序，每個步驟完成後應可獨立驗證：

```
Phase 1: 基礎設施 + 後端 API
─────────────────────────────
Step 1: 基礎設施
  └─ docker-compose.yml + init.sql → `docker compose up` 驗證 DB 和 MinIO

Step 2: FastAPI 骨架
  └─ main.py + config.py + health.py + CORS → `/v1/health` 回傳 connected

Step 3: Parser 模組
  └─ parser.py + markdown.py + test_parser.py → 所有解析測試通過

Step 4: Storage 模組
  └─ storage.py → 可上傳/下載 MinIO 檔案

Step 5: Publish 端點
  └─ POST /v1/skills/publish + test_publish.py → 完整 publish 流程（含 readme_html 快取）

Step 6: Query 端點
  └─ GET /skills/{name}, /versions, /download, /search → 所有 GET 測試通過
  └─ 全文搜尋 tsvector + categories + sort

Step 7: Stars + Categories + Auth 端點
  └─ POST/DELETE /skills/{name}/star + GET /categories + POST /auth/github
  └─ GET /users/{username}

Phase 2: Go CLI
────────────────
Step 8: Go CLI 骨架
  └─ root.go + config.go + login.go → `agentskills login` 可存 token

Step 9: CLI push/pull
  └─ push.go + pull.go → 完整 CLI ↔ API 流程跑通

Step 10: CLI search + init
  └─ search.go + init_cmd.go → 所有 CLI 指令完成

Phase 3: Web 前端 (Next.js)
────────────────────────────
Step 11: Next.js 骨架
  └─ next.config.ts + layout.tsx + API client + shadcn/ui 初始化
  └─ 可存取首頁，Header/Footer 正常顯示

Step 12: 首頁 + 搜尋
  └─ 首頁 Hero + 分類瀏覽 + 搜尋列
  └─ /search 頁面 + SkillCard 列表 + 篩選/排序

Step 13: Skill 詳情頁
  └─ /skills/[name] 頁面：Markdown 渲染、版本歷史、metadata 側欄
  └─ Star 功能（登入後可操作）

Step 14: GitHub OAuth 登入
  └─ NextAuth.js + GitHub OAuth → 登入/登出/使用者 menu
  └─ /user/[username] 公開 profile 頁

Step 15: 整合驗證
  └─ 全流程：Web 瀏覽 → CLI publish → Web 可見 → CLI pull
  └─ 執行 §8.2 所有驗證指令 + Web UI 手動測試
```

---

## 12. Web 前端設計 (Next.js)

### 12.1 技術選型

| 技術 | 版本 | 用途 |
|------|------|------|
| Next.js | 15+ | React 全端框架，App Router + Server Components |
| React | 19+ | UI 元件 |
| TypeScript | 5+ | 型別安全 |
| Tailwind CSS | 4+ | Utility-first CSS 框架 |
| shadcn/ui | latest | 可客製化 UI 元件庫（基於 Radix UI） |
| NextAuth.js | 5+ | GitHub OAuth 認證 |
| react-markdown | latest | SKILL.md 內容渲染 |
| react-syntax-highlighter | latest | Code block 語法高亮 |
| lucide-react | latest | Icon 系統 |

### 12.2 專案結構

```
web/
├── package.json
├── next.config.ts
├── tailwind.config.ts
├── tsconfig.json
├── .env.local.example
├── public/
│   ├── logo.svg
│   └── favicon.ico
├── src/
│   ├── app/
│   │   ├── layout.tsx                # Root layout: providers, Header, Footer
│   │   ├── page.tsx                  # 首頁: Hero + 搜尋 + 分類 + 精選 Skills
│   │   ├── globals.css               # Tailwind + 自訂全域樣式
│   │   ├── skills/
│   │   │   └── [name]/
│   │   │       ├── page.tsx          # Skill 詳情頁: Markdown + metadata sidebar
│   │   │       └── versions/
│   │   │           └── page.tsx      # 版本歷史頁
│   │   ├── search/
│   │   │   └── page.tsx              # 搜尋結果頁 (含篩選/排序)
│   │   ├── categories/
│   │   │   └── [category]/
│   │   │       └── page.tsx          # 分類瀏覽頁
│   │   ├── user/
│   │   │   └── [username]/
│   │   │       └── page.tsx          # 使用者公開 Profile
│   │   ├── settings/
│   │   │   └── page.tsx              # 使用者設定 (API token 顯示/複製)
│   │   ├── login/
│   │   │   └── page.tsx              # 登入頁
│   │   └── api/
│   │       └── auth/
│   │           └── [...nextauth]/
│   │               └── route.ts      # NextAuth.js GitHub OAuth handler
│   ├── components/
│   │   ├── ui/                       # shadcn/ui 元件 (Button, Card, Input, Badge, etc.)
│   │   ├── layout/
│   │   │   ├── header.tsx            # Top nav: logo, 搜尋列, 登入/avatar
│   │   │   └── footer.tsx            # Footer: links, copyright
│   │   ├── skills/
│   │   │   ├── skill-card.tsx        # Skill 卡片 (搜尋結果、首頁列表)
│   │   │   ├── skill-detail.tsx      # Skill 詳情主內容
│   │   │   ├── skill-sidebar.tsx     # Metadata 側欄 (install, stats, tags)
│   │   │   ├── version-list.tsx      # 版本歷史表格
│   │   │   ├── star-button.tsx       # Star/Unstar 按鈕
│   │   │   └── install-command.tsx   # CLI 安裝指令 (一鍵複製)
│   │   ├── search/
│   │   │   ├── search-bar.tsx        # 全域搜尋輸入框 (含快捷鍵 ⌘K)
│   │   │   ├── search-filters.tsx    # 篩選: category, tag, sort
│   │   │   └── search-results.tsx    # 結果列表 + 分頁
│   │   ├── home/
│   │   │   ├── hero.tsx              # 首頁 Hero: 標語 + 搜尋列
│   │   │   ├── category-grid.tsx     # 分類卡片網格
│   │   │   ├── featured-skills.tsx   # 精選/熱門 Skills
│   │   │   └── stats-bar.tsx         # 平台統計 (Skills 數, 下載數)
│   │   ├── markdown/
│   │   │   └── markdown-renderer.tsx # SKILL.md 內容渲染 (安全 HTML)
│   │   └── auth/
│   │       ├── login-button.tsx      # GitHub 登入按鈕
│   │       └── user-menu.tsx         # 登入後使用者下拉選單
│   ├── lib/
│   │   ├── api.ts                    # FastAPI client (server-side + client-side fetch)
│   │   ├── auth.ts                   # NextAuth.js 設定
│   │   └── utils.ts                  # 格式化 (下載數, 日期, 檔案大小)
│   └── types/
│       └── index.ts                  # TypeScript 型別 (對應 API schemas)
├── Dockerfile
└── .dockerignore
```

### 12.3 頁面規格

#### 首頁 (`/`)

```
┌────────────────────────────────────────────────┐
│  🔎  AgentSkills        [Search...]   [Login]  │
├────────────────────────────────────────────────┤
│                                                │
│     Discover AI Agent Skills                   │
│     The open registry for agent capabilities   │
│                                                │
│     [═══════════ Search skills... ══════════]   │
│                                                │
├────────────────────────────────────────────────┤
│  Categories                                    │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐│
│  │ 💻   │ │ ⚡   │ │ 🧠   │ │ 🔧   │ │ 📊   ││
│  │ Dev  │ │Prod. │ │AI/ML │ │DevOps│ │ Data ││
│  │ 128  │ │  85  │ │  67  │ │  43  │ │  38  ││
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘│
├────────────────────────────────────────────────┤
│  Trending Skills                    [See all →]│
│  ┌──────────────────┐ ┌──────────────────┐    │
│  │ code-review-agent│ │ deploy-helper    │    │
│  │ ⭐ 42  ↓ 1.2k   │ │ ⭐ 38  ↓ 890    │    │
│  │ PR code review...│ │ Deploy workflow..│    │
│  └──────────────────┘ └──────────────────┘    │
├────────────────────────────────────────────────┤
│  Latest Skills                      [See all →]│
│  ┌──────────────────┐ ┌──────────────────┐    │
│  │ ...              │ │ ...              │    │
│  └──────────────────┘ └──────────────────┘    │
└────────────────────────────────────────────────┘
```

- **Server Component**：首頁所有資料透過 Server Component 在伺服器端取得
- **API 呼叫**：`GET /v1/categories` + `GET /v1/skills?sort=stars&per_page=6` + `GET /v1/skills?sort=newest&per_page=6`
- **SEO**：動態 metadata title/description

#### 搜尋頁 (`/search?q=&category=&tag=&sort=`)

```
┌────────────────────────────────────────────────┐
│  Header                                        │
├────────────────────────────────────────────────┤
│  Results for "code review"          45 results │
│                                                │
│  [Category ▾] [Sort: Relevance ▾]  [Tags: +]  │
│                                                │
│  ┌────────────────────────────────────────────┐│
│  │ 📦 code-review-agent        v1.2.0        ││
│  │ by liuyukai  ⭐ 42  ↓ 1.2k               ││
│  │ PR code review skill with GitHub...        ││
│  │ [code-review] [github] [development]       ││
│  ├────────────────────────────────────────────┤│
│  │ 📦 code-review-lite          v0.3.0       ││
│  │ by devuser   ⭐ 7   ↓ 120                 ││
│  │ Lightweight review helper...               ││
│  │ [code-review] [lightweight]                ││
│  └────────────────────────────────────────────┘│
│                                                │
│  [← Prev]  Page 1 of 3  [Next →]              │
└────────────────────────────────────────────────┘
```

- URL query params 驅動搜尋（支援 browser back/forward）
- Server Component + `searchParams` 做 SSR
- `GET /v1/skills?q=...&category=...&tag=...&sort=...&page=...`

#### Skill 詳情頁 (`/skills/[name]`)

```
┌────────────────────────────────────────────────┐
│  Header                                        │
├──────────────────────────┬─────────────────────┤
│  📦 code-review-agent    │  Install             │
│  by liuyukai             │  ┌─────────────────┐│
│  PR code review skill... │  │ agentskills pull ││
│  ⭐ Star (42)  ↓ 1.2k   │  │ code-review-agen││
│                          │  └────────── [📋] ──┘│
│  ─────────────────────── │                     │
│  # Code Review Agent     │  Version            │
│                          │  1.2.0 (latest)     │
│  This skill performs     │  Published 2d ago   │
│  automated code review   │                     │
│  of pull requests...     │  License            │
│                          │  MIT                │
│  ## Usage                │                     │
│  1. Configure the repo   │  Tags               │
│  2. Run the review...    │  [code-review]      │
│                          │  [github]           │
│  ## Configuration        │                     │
│  ```yaml                 │  Category           │
│  settings:               │  Development        │
│    threshold: 0.8        │                     │
│  ```                     │  Size               │
│                          │  15.3 KB            │
│                          │                     │
│                          │  [All versions →]   │
├──────────────────────────┴─────────────────────┤
│  Footer                                        │
└────────────────────────────────────────────────┘
```

- 左側：SKILL.md markdown body 渲染（使用 `react-markdown` + 語法高亮）
- 右側：Metadata sidebar（install command, version, license, tags, category, size, star button）
- **SSR + dynamic metadata**：`generateMetadata()` 產生 SEO title/description/og:image
- API：`GET /v1/skills/{name}`

#### 使用者頁面 (`/user/[username]`)

```
┌────────────────────────────────────────────────┐
│  Header                                        │
├────────────────────────────────────────────────┤
│  [Avatar]  liuyukai                            │
│  Liu Yu Kai                                    │
│  Backend developer, AI enthusiast              │
│  Joined Jan 2026  |  ↓ 156 total  ⭐ 47 total  │
├────────────────────────────────────────────────┤
│  Published Skills (3)                          │
│  ┌────────────────────────────────────────────┐│
│  │ code-review-agent  v1.2.0  ⭐42  ↓1.2k    ││
│  │ deploy-helper      v2.0.0  ⭐38  ↓890     ││
│  │ test-runner        v0.5.0  ⭐5   ↓45      ││
│  └────────────────────────────────────────────┘│
└────────────────────────────────────────────────┘
```

- API：`GET /v1/users/{username}`

### 12.4 API Client (`lib/api.ts`)

```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

// Server-side fetch（Server Components 直接呼叫，帶 cache control）
export async function getSkill(name: string) {
  const res = await fetch(`${API_BASE}/v1/skills/${name}`, {
    next: { revalidate: 60 }, // ISR: 60 秒快取
  });
  if (!res.ok) throw new Error(`Skill not found: ${name}`);
  return res.json();
}

// Client-side fetch（Star 等互動操作）
export async function starSkill(name: string, token: string) {
  return fetch(`${API_BASE}/v1/skills/${name}/star`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });
}
```

**快取策略（Next.js ISR）：**

| 頁面 | `revalidate` | 說明 |
|------|-------------|------|
| 首頁 | 300s (5min) | 分類統計、trending 不需即時 |
| 搜尋頁 | 0 (no cache) | 每次搜尋都打 API |
| Skill 詳情 | 60s | 大部分內容靜態，star/download 可延遲 |
| 使用者頁 | 120s | 不頻繁更新 |

### 12.5 環境變數 (.env.local)

```env
# FastAPI Backend
NEXT_PUBLIC_API_URL=http://localhost:8000

# NextAuth.js
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your-random-secret-here

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
```

### 12.6 設計風格

- **配色**：深色主題為主（類似 GitHub Dark），支援 Light/Dark 切換
- **字體**：系統字體堆疊 (`font-sans`)，程式碼用等寬字體 (`font-mono`)
- **元件庫**：shadcn/ui（可客製化，基於 Radix UI + Tailwind CSS）
- **動畫**：極簡，僅 hover/focus 狀態轉換，不使用花俏動畫
- **響應式**：Mobile-first，斷點 `sm:640px`, `md:768px`, `lg:1024px`, `xl:1280px`

### 12.7 核心依賴 (package.json)

```json
{
  "dependencies": {
    "next": "^15.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "next-auth": "^5.0.0",
    "react-markdown": "^9.0.0",
    "react-syntax-highlighter": "^15.0.0",
    "remark-gfm": "^4.0.0",
    "lucide-react": "^0.400.0",
    "class-variance-authority": "^0.7.0",
    "clsx": "^2.0.0",
    "tailwind-merge": "^2.0.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "tailwindcss": "^4.0.0",
    "@types/react": "^19.0.0",
    "vitest": "^2.0.0",
    "@testing-library/react": "^16.0.0",
    "playwright": "^1.45.0"
  }
}
```

---

## 13. API 端點總覽（完整版）

| Method | Endpoint | Auth | 說明 |
|--------|----------|------|------|
| `POST` | `/v1/skills/publish` | Bearer Token | 上傳 Skill Bundle (.tar.gz) |
| `GET` | `/v1/skills/{name}` | Optional | Skill 資訊 + 最新版本（帶 auth 時含 starred_by_me） |
| `GET` | `/v1/skills/{name}/versions` | No | 列出所有版本 |
| `GET` | `/v1/skills/{name}/versions/{version}/download` | No | 下載指定版本 Bundle |
| `GET` | `/v1/skills` | No | 搜尋 Skills（全文搜尋 + 篩選 + 排序） |
| `POST` | `/v1/skills/{name}/star` | Bearer Token | 收藏 Skill |
| `DELETE` | `/v1/skills/{name}/star` | Bearer Token | 取消收藏 |
| `GET` | `/v1/categories` | No | 列出分類 + 各分類 Skill 數 |
| `POST` | `/v1/auth/github` | No | GitHub OAuth 登入/註冊 |
| `GET` | `/v1/users/{username}` | No | 使用者公開資料 + 發布的 Skills |
| `GET` | `/v1/health` | No | 健康檢查 |

---

*End of Document*