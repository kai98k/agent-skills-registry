# AgentSkills — Software Design Document (SDD)

**Version:** 2.0
**Date:** 2026-02-23
**Author:** LIU YU KAI
**Purpose:** 本文件為 Claude Code 的開發藍圖。請嚴格依照本文件的規格、目錄結構、API Contract 與 DB Schema 進行開發與驗證。

---

## 1. 專案概述

AgentSkills 是一個 AI Agent Skill 的集中式 Registry 平台，類似 npm 或 Docker Hub。開發者可透過 CLI 工具上傳（push）與下載（pull）標準化的 Skill Bundle，平台負責版本控制、Metadata 解析與檔案儲存。

**技術架構：**

- 全 Go 實作（CLI + Server），透過 build tags 分離為兩個 binary
- 雙模式資料庫：SQLite（嵌入式，開發/個人用）+ PostgreSQL（生產環境）
- 雙模式儲存：本地檔案系統（開發/個人用）+ S3/MinIO（生產環境）
- 跨平台編譯：Linux / macOS / Windows

**MVP 範圍包含：**

- Go HTTP Server（chi router）
- Go CLI 工具（Cobra）
- SQLite 嵌入式資料庫（預設）/ PostgreSQL（可選）
- 本地檔案儲存（預設）/ MinIO S3 相容儲存（可選）
- Docker 部署（簡易模式 + 生產模式）
- 跨平台 binary 編譯（Linux / macOS / Windows）

**MVP 明確不包含：**

- Web UI 前端
- OAuth / 第三方登入
- Semver range resolution（`^1.0.0`）
- Skill 之間的依賴關係
- unpublish / deprecate 功能
- 自動版本號 bump

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
| version | string | ✅ | 嚴格 semver，使用 `golang.org/x/mod/semver` 驗證 |
| description | string | ✅ | 1-256 字元 |
| author | string | ✅ | 必須與 API Token 對應的 username 一致 |
| tags | list[string] | ❌ | 最多 10 個，每個 `/^[a-z0-9\-]{1,32}$/` |
| license | string | ❌ | 若提供需為合法 SPDX identifier |
| min_agent_version | string | ❌ | MVP 階段僅儲存，不做邏輯判斷 |

---

## 3. 系統架構

### 3.1 簡易模式（預設）— 零外部依賴

```
┌─────────────────────────────────────┐
│     agentskills-server binary       │
│                                     │
│  ┌──────────────────────────────┐   │
│  │ net/http + chi (HTTP Server) │   │
│  │ port: 8000                   │   │
│  └──────────────┬───────────────┘   │
│                 │                   │
│  ┌──────────────┴───────────────┐   │
│  │ SQLite (嵌入式, 純 Go)       │   │
│  │ ./data/agentskills.db        │   │
│  └──────────────────────────────┘   │
│  ┌──────────────────────────────┐   │
│  │ Local FileSystem             │   │
│  │ ./data/bundles/              │   │
│  └──────────────────────────────┘   │
└───────────────────┬─────────────────┘
                    │
             ┌──────┴──────┐
             │ agentskills │
             │  CLI binary │
             │ (本機執行)   │
             └─────────────┘
```

### 3.2 生產模式 — 外部 PostgreSQL + S3

```
┌──────────────────────────────────────────┐
│            docker-compose                │
│                                          │
│  ┌─────────────┐    ┌─────────────────┐  │
│  │ PostgreSQL   │    │ MinIO (S3)      │  │
│  │ port: 5432   │    │ API:  9000      │  │
│  │              │    │ Console: 9001   │  │
│  └──────┬───────┘    └──────┬──────────┘  │
│         │                   │             │
│         └─────────┬─────────┘             │
│              ┌────┴─────────┐             │
│              │ agentskills  │             │
│              │ -server      │             │
│              │ port: 8000   │             │
│              └────┬─────────┘             │
│                   │                       │
└───────────────────┼───────────────────────┘
                    │
             ┌──────┴──────┐
             │ agentskills │
             │  CLI binary │
             │ (本機執行)   │
             └─────────────┘
```

### 3.3 Build Tags 分離策略

使用 Go build tags 將 CLI 和 Server 分開編譯，降低 CLI binary 體積：

| Binary | Build 指令 | 包含內容 | 體積 |
|--------|-----------|---------|------|
| `agentskills` | `go build .` | CLI 指令 (push/pull/search/init/login) | ~8MB |
| `agentskills-server` | `go build -tags server .` | CLI 指令 + HTTP Server + DB + Storage | ~18MB |

---

## 4. 資料庫設計

### 4.1 Schema (PostgreSQL 版)

```sql
-- init.sql (PostgreSQL)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==========================================
-- USERS
-- ==========================================
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(64)  UNIQUE NOT NULL,
    api_token   VARCHAR(128) UNIQUE NOT NULL,
    created_at  TIMESTAMPTZ  DEFAULT now()
);

-- ==========================================
-- SKILLS (一個 name 一筆)
-- ==========================================
CREATE TABLE skills (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(128) UNIQUE NOT NULL,
    owner_id    UUID NOT NULL REFERENCES users(id),
    downloads   BIGINT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

-- ==========================================
-- SKILL VERSIONS (每次 publish 一筆, immutable)
-- ==========================================
CREATE TABLE skill_versions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    skill_id      UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    version       VARCHAR(32) NOT NULL,
    bundle_key    TEXT NOT NULL,           -- storage object key, e.g. "code-review-agent/1.0.0.tar.gz"
    metadata      JSONB NOT NULL,          -- 完整 frontmatter
    checksum      VARCHAR(64) NOT NULL,    -- SHA-256 hex digest
    size_bytes    BIGINT NOT NULL,         -- bundle 檔案大小
    published_at  TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT uq_skill_version UNIQUE (skill_id, version)
);

CREATE INDEX idx_skill_versions_latest
    ON skill_versions (skill_id, published_at DESC);

CREATE INDEX idx_skills_name
    ON skills (name);

-- ==========================================
-- SEED DATA (開發用測試帳號)
-- ==========================================
INSERT INTO users (username, api_token)
VALUES ('dev', 'dev-token-12345');
```

### 4.2 Schema (SQLite 版，嵌入式)

```sql
-- migrations/001_init.sql (SQLite)
CREATE TABLE IF NOT EXISTS users (
    id          TEXT PRIMARY KEY,
    username    TEXT UNIQUE NOT NULL,
    api_token   TEXT UNIQUE NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
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
    metadata      TEXT NOT NULL,     -- JSON string
    checksum      TEXT NOT NULL,
    size_bytes    INTEGER NOT NULL,
    published_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(skill_id, version)
);

CREATE INDEX IF NOT EXISTS idx_skill_versions_latest
    ON skill_versions (skill_id, published_at DESC);

CREATE INDEX IF NOT EXISTS idx_skills_name
    ON skills (name);

-- Seed data
INSERT OR IGNORE INTO users (id, username, api_token)
VALUES ('00000000-0000-0000-0000-000000000001', 'dev', 'dev-token-12345');
```

### 4.3 設計決策

- **兩表分離**：`skills` 存身份與聚合資料（downloads），`skill_versions` 存每次發布的 immutable 記錄。
- **Immutable publish**：同一 `(skill_id, version)` 不可覆寫，嘗試重複發布回傳 `409 Conflict`。
- **Soft latest**：最新版透過 `published_at DESC LIMIT 1` 查詢，不額外維護 `latest` 欄位。
- **JSONB / JSON metadata**：frontmatter 全文存入，PostgreSQL 用 JSONB，SQLite 用 TEXT (JSON string)。
- **UUID 主鍵**：PostgreSQL 用 `gen_random_uuid()`，SQLite 由 Go 程式碼以 `google/uuid` 產生。
- **雙 Schema 策略**：PostgreSQL 使用 `init.sql`（外部初始化），SQLite 使用 `go:embed` 嵌入 migration SQL。

---

## 5. API 設計 (Go HTTP Server)

**Base URL:** `http://localhost:8000/v1`

### 5.1 認證

MVP 使用靜態 API Token，透過 `Authorization: Bearer <token>` Header 傳遞。

```
Authorization: Bearer dev-token-12345
```

認證失敗回傳 `401 Unauthorized`。僅 `POST` (publish) 需要認證，`GET` 端點皆為公開。

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
10. 上傳 `.tar.gz` 至 Storage → key: `{name}/{version}.tar.gz`
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

**行為：** 直接串流檔案內容。

**Response:** `200 OK`

- `Content-Type: application/gzip`
- `Content-Disposition: attachment; filename="code-review-agent-1.0.0.tar.gz"`
- `X-Checksum-SHA256: a1b2c3d4...`
- Body: raw binary

**Side effect:** `skills.downloads += 1`

**Error:** `404` 若 name 或 version 不存在。

---

#### `GET /v1/skills?q={keyword}&tag={tag}&page={n}&per_page={n}`

搜尋 Skills。

**Query Parameters:**

| Param | Type | Default | 說明 |
|-------|------|---------|------|
| q | string | - | 搜尋 name 和 description（ILIKE / LIKE） |
| tag | string | - | 精確匹配 metadata tags（可多次傳遞） |
| page | int | 1 | 頁碼 |
| per_page | int | 20 | 每頁數量，最大 100 |

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
      "downloads": 42,
      "latest_version": "1.2.0",
      "updated_at": "2026-02-20T12:00:00Z",
      "tags": ["code-review", "github"]
    }
  ]
}
```

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

## 6. 後端專案結構 (Go)

```
agentskills/
├── main.go                     # 程式入口
├── go.mod
├── go.sum
├── Makefile                    # 跨平台編譯
├── Dockerfile                  # 多階段 build (Server)
├── docker-compose.yml          # 簡易模式 (SQLite + LocalFS)
├── docker-compose.prod.yml     # 生產模式 (PostgreSQL + MinIO)
├── init.sql                    # PostgreSQL schema (給外部 PG 使用)
│
├── cmd/                        # Cobra 指令定義
│   ├── root.go                 # Root command + global flags
│   ├── serve.go                # agentskills-server serve (build tag: server)
│   ├── migrate.go              # agentskills-server migrate (build tag: server)
│   ├── init_cmd.go             # agentskills init
│   ├── push.go                 # agentskills push
│   ├── pull.go                 # agentskills pull
│   ├── search.go               # agentskills search
│   ├── login.go                # agentskills login
│   └── version.go              # agentskills version
│
├── internal/
│   ├── server/                 # HTTP Server 核心 (build tag: server)
│   │   ├── server.go           # Server 結構體、啟動/關閉
│   │   ├── router.go           # 路由定義 (chi router)
│   │   ├── middleware.go       # 認證、logging、recovery
│   │   └── handlers/           # HTTP Handlers
│   │       ├── health.go       # GET  /v1/health
│   │       ├── publish.go      # POST /v1/skills/publish
│   │       ├── skills.go       # GET  /v1/skills/{name}
│   │       ├── versions.go     # GET  /v1/skills/{name}/versions
│   │       ├── download.go     # GET  /v1/skills/{name}/versions/{version}/download
│   │       └── search.go       # GET  /v1/skills?q=...
│   │
│   ├── database/               # 資料庫抽象層 (build tag: server)
│   │   ├── database.go         # Database interface 定義
│   │   ├── models.go           # Go struct (User, Skill, SkillVersion)
│   │   ├── sqlite.go           # SQLite 實作 (modernc.org/sqlite, 純 Go)
│   │   ├── postgres.go         # PostgreSQL 實作 (lib/pq)
│   │   └── migrate.go          # Schema migration (嵌入 SQL)
│   │
│   ├── storage/                # 儲存抽象層 (build tag: server)
│   │   ├── storage.go          # Storage interface 定義
│   │   ├── local.go            # 本地檔案系統實作
│   │   └── s3.go               # S3/MinIO 實作 (aws-sdk-go-v2)
│   │
│   ├── bundle/                 # Bundle 處理 (CLI + Server 共用)
│   │   ├── pack.go             # tar.gz 打包 (CLI 用)
│   │   ├── unpack.go           # tar.gz 解壓 (Server 用)
│   │   └── checksum.go         # SHA-256 計算
│   │
│   ├── parser/                 # SKILL.md 解析 (CLI + Server 共用)
│   │   ├── frontmatter.go      # YAML Frontmatter 解析
│   │   └── validate.go         # 驗證規則 (§2.3)
│   │
│   ├── api/                    # CLI HTTP client
│   │   └── client.go           # 封裝所有 API 呼叫
│   │
│   └── config/                 # 配置管理
│       └── config.go           # CLI config (~/.agentskills/config.yaml)
│
├── migrations/                 # 嵌入式 SQL migrations (build tag: server)
│   ├── embed.go                # go:embed 載入 SQL 檔案
│   └── 001_init.sql            # SQLite 初始 schema
│
└── tests/                      # 測試
    ├── integration_test.go     # 端對端測試
    ├── publish_test.go
    ├── pull_test.go
    ├── search_test.go
    ├── parser_test.go
    └── testdata/               # 測試用 skill bundles
        └── valid-skill/
            └── SKILL.md
```

### 6.1 核心依賴 (go.mod)

```
module github.com/liuyukai/agentskills

go 1.22

require (
    // CLI Framework
    github.com/spf13/cobra       v1.8+
    github.com/spf13/viper       v1.19+

    // HTTP Router
    github.com/go-chi/chi/v5     v5.1+

    // Database
    modernc.org/sqlite           v1.29+   // 純 Go SQLite (無 CGO)
    github.com/lib/pq            v1.10+   // PostgreSQL driver

    // S3/MinIO
    github.com/aws/aws-sdk-go-v2 v1.30+
    github.com/aws/aws-sdk-go-v2/service/s3

    // YAML
    gopkg.in/yaml.v3             v3.0+

    // UUID
    github.com/google/uuid       v1.6+

    // Semver
    golang.org/x/mod                       // semver validation

    // Testing
    github.com/stretchr/testify  v1.9+
)
```

### 6.2 Server 配置

透過環境變數或 CLI flag 設定，所有 flag 均可透過 `AGENTSKILLS_` 前綴的環境變數覆蓋：

```bash
# Server
AGENTSKILLS_PORT=8000

# Database
AGENTSKILLS_DB_DRIVER=sqlite              # sqlite | postgres
AGENTSKILLS_DB_DSN=./data/agentskills.db  # SQLite 檔案路徑 或 PostgreSQL DSN

# Storage
AGENTSKILLS_STORAGE_DRIVER=local          # local | s3
AGENTSKILLS_STORAGE_PATH=./data/bundles   # 本地儲存路徑

# S3/MinIO (當 STORAGE_DRIVER=s3)
AGENTSKILLS_S3_ENDPOINT=http://localhost:9000
AGENTSKILLS_S3_ACCESS_KEY=minioadmin
AGENTSKILLS_S3_SECRET_KEY=minioadmin
AGENTSKILLS_S3_BUCKET=skills
AGENTSKILLS_S3_REGION=us-east-1

# Limits
AGENTSKILLS_MAX_BUNDLE_SIZE=52428800      # 50MB
```

### 6.3 parser 核心邏輯

```
SKILL.md 解析流程：
1. 接收上傳的 .tar.gz bytes
2. 解壓至暫存目錄 (os.MkdirTemp)
3. 遍歷解壓後的檔案，找到 SKILL.md（必須在根目錄或一層子目錄內）
4. 使用 gopkg.in/yaml.v3 解析 YAML Frontmatter + Markdown body
5. 對 YAML 執行 §2.3 驗證規則
6. 回傳 ParsedSkill struct

注意事項：
- 解壓時檢查 zip bomb（解壓後總大小不超過 200MB）
- 路徑穿越攻擊防護（所有解壓路徑必須在暫存目錄下）
- 暫存目錄用完即刪
```

### 6.4 storage 核心邏輯

```
Storage Interface 設計：
- Init(): 初始化（建立目錄 或 檢查 S3 Bucket）
- HealthCheck(): 健康檢查
- Upload(key, reader, size): 上傳 bundle
- Download(key): 下載 bundle，回傳 io.ReadCloser
- Exists(key): 檢查 key 是否存在
- Delete(key): 刪除（預留）

LocalStorage:
- 基於 os 套件操作本地檔案系統
- basePath 下以 {name}/{version}.tar.gz 結構儲存

S3Storage:
- 使用 aws-sdk-go-v2 連接 MinIO（S3 相容 API）
- 使用 path-style addressing（MinIO 不支援 virtual-hosted-style）
```

---

## 7. CLI 設計 (Go / Cobra)

### 7.1 指令規格

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

### 7.2 本地設定檔

路徑: `~/.agentskills/config.yaml`

```yaml
api_url: "http://localhost:8000"
token: "dev-token-12345"
```

---

## 8. Docker 部署

### 8.1 Dockerfile (多階段 Build)

```dockerfile
# === Build Stage ===
FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -tags server -ldflags="-s -w" -o /agentskills-server .

# === Runtime Stage ===
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /agentskills-server /usr/local/bin/agentskills-server

RUN mkdir -p /data/bundles
VOLUME ["/data"]

EXPOSE 8000

ENTRYPOINT ["agentskills-server"]
CMD ["serve", "--port", "8000"]
```

最終鏡像大小：~25MB (Alpine 7MB + Go binary ~18MB)

### 8.2 docker-compose.yml (簡易模式)

```yaml
services:
  agentskills:
    build: .
    ports:
      - "8000:8000"
    volumes:
      - data:/data
    # 預設使用 SQLite + 本地檔案系統，零配置

volumes:
  data:
```

### 8.3 docker-compose.prod.yml (生產模式)

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: agentskills
      POSTGRES_USER: prod
      POSTGRES_PASSWORD: ${PG_PASSWORD}
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U prod -d agentskills"]
      interval: 5s
      timeout: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD}
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
      mc alias set local http://minio:9000 $${MINIO_USER} $${MINIO_PASSWORD};
      mc mb --ignore-existing local/skills;
      echo 'Bucket skills created';
      "

  agentskills:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      postgres:
        condition: service_healthy
      minio-init:
        condition: service_completed_successfully
    environment:
      AGENTSKILLS_DB_DRIVER: postgres
      AGENTSKILLS_DB_DSN: postgres://prod:${PG_PASSWORD}@postgres:5432/agentskills?sslmode=disable
      AGENTSKILLS_STORAGE_DRIVER: s3
      AGENTSKILLS_S3_ENDPOINT: http://minio:9000
      AGENTSKILLS_S3_ACCESS_KEY: ${MINIO_USER}
      AGENTSKILLS_S3_SECRET_KEY: ${MINIO_PASSWORD}
      AGENTSKILLS_S3_BUCKET: skills
    command: ["serve", "--port", "8000"]

volumes:
  pgdata:
  miniodata:
```

### 8.4 啟動與驗證指令

```bash
# ==========================================
# 方式 A: 簡易模式 (SQLite + LocalFS)
# ==========================================

# 1. 啟動
docker compose up -d

# 2. 驗證
curl http://localhost:8000/v1/health
# 預期: {"status":"ok","database":"connected","storage":"connected"}

# ==========================================
# 方式 B: 生產模式 (PostgreSQL + MinIO)
# ==========================================

# 1. 設定環境變數
export PG_PASSWORD=your-secure-password
export MINIO_USER=minioadmin
export MINIO_PASSWORD=minioadmin

# 2. 啟動
docker compose -f docker-compose.prod.yml up -d

# 3. 驗證 PostgreSQL
docker compose -f docker-compose.prod.yml exec postgres \
  psql -U prod -d agentskills -c "SELECT COUNT(*) FROM users;"

# 4. 驗證 API
curl http://localhost:8000/v1/health

# ==========================================
# 方式 C: 直接執行 binary (不需要 Docker)
# ==========================================

# 1. 啟動 Server
./agentskills-server serve
# → SQLite: ./data/agentskills.db
# → Bundles: ./data/bundles/

# 2. 驗證
curl http://localhost:8000/v1/health

# ==========================================
# 完整功能測試
# ==========================================

# Publish 測試
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

# Pull 測試
curl http://localhost:8000/v1/skills/test-skill
# 預期: 200 OK

curl -OJ http://localhost:8000/v1/skills/test-skill/versions/0.1.0/download
# 預期: 下載 .tar.gz

# Immutable 測試
curl -X POST http://localhost:8000/v1/skills/publish \
  -H "Authorization: Bearer dev-token-12345" \
  -F "file=@test-skill.tar.gz"
# 預期: 409 Conflict

# CLI 測試 (Server 啟動後)
./agentskills login
./agentskills push /tmp/test-skill
./agentskills pull test-skill
./agentskills search test
```

---

## 9. 跨平台編譯

### 9.1 Makefile

```makefile
VERSION     := $(shell git describe --tags --always --dirty)
BUILD_TIME  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

.PHONY: build-cli build-server build-all clean test

# === CLI only (輕量，不含 Server/DB/Storage) ===
build-cli:
	go build -ldflags "$(LDFLAGS)" -o bin/agentskills .

# === Server (包含 CLI + Server + DB + Storage) ===
build-server:
	go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server .

# === 跨平台完整編譯 ===
build-all: build-all-cli build-all-server

build-all-cli:
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/agentskills-windows-amd64.exe .

build-all-server:
	GOOS=linux   GOARCH=amd64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -tags server -ldflags "$(LDFLAGS)" -o bin/agentskills-server-windows-amd64.exe .

test:
	go test -v -race -tags server ./...

clean:
	rm -rf bin/
```

### 9.2 支援平台

| 平台 | CLI Binary | Server Binary |
|------|-----------|--------------|
| Linux amd64 | `agentskills-linux-amd64` | `agentskills-server-linux-amd64` |
| Linux arm64 | `agentskills-linux-arm64` | `agentskills-server-linux-arm64` |
| macOS amd64 (Intel) | `agentskills-darwin-amd64` | `agentskills-server-darwin-amd64` |
| macOS arm64 (Apple Silicon) | `agentskills-darwin-arm64` | `agentskills-server-darwin-arm64` |
| Windows amd64 | `agentskills-windows-amd64.exe` | `agentskills-server-windows-amd64.exe` |

---

## 10. 測試策略

### 10.1 後端測試 (go test)

測試使用 SQLite in-memory 資料庫 + LocalFS temp 目錄，不需要外部依賴。

**必要測試案例：**

| 測試檔案 | 案例 | 預期 |
|----------|------|------|
| parser_test.go | 合法 SKILL.md | 正確解析所有欄位 |
| parser_test.go | 缺少 name 欄位 | ValidationError |
| parser_test.go | version 非 semver | ValidationError |
| parser_test.go | name 含大寫或特殊字元 | ValidationError |
| parser_test.go | 無 SKILL.md 的 tar.gz | FileNotFoundError |
| publish_test.go | 正常 publish | 201, DB 有記錄, Storage 有檔案 |
| publish_test.go | 無 auth header | 401 |
| publish_test.go | 重複版本 | 409 |
| publish_test.go | name 被他人佔用 | 403 |
| publish_test.go | 超過 50MB | 413 |
| pull_test.go | 下載 latest | 200, 正確 binary |
| pull_test.go | 下載指定版本 | 200, checksum 正確 |
| pull_test.go | 不存在的 skill | 404 |
| search_test.go | keyword 搜尋 | 回傳匹配結果 |
| search_test.go | tag 篩選 | 僅回傳有該 tag 的結果 |
| search_test.go | 空結果 | 200, results: [] |

### 10.2 CLI 測試 (go test)

- 本地 frontmatter 解析與驗證
- tar.gz 打包排除清單
- SHA-256 checksum 計算
- API client 呼叫 (使用 `net/http/httptest` mock server)
- config 檔讀寫

---

## 11. 安全性注意事項

| 威脅 | 防護措施 |
|------|----------|
| Zip bomb | 解壓時限制總大小 200MB，超過即中止 |
| 路徑穿越 (../../etc/passwd) | 所有解壓路徑檢查必須在暫存目錄下 |
| 任意檔案執行 | Server 端僅解析 SKILL.md，不執行 scripts/ 內任何檔案 |
| Token 洩漏 | CLI config 檔設 0600 權限；API logs 不記錄完整 token |
| SQL Injection | 使用 database/sql parameterized queries (`$1`, `?` placeholders) |
| 超大檔案 DoS | HTTP handler 層限制 request body 50MB (`http.MaxBytesReader`) |

---

## 12. 開發順序建議

以下為建議的實作優先順序，每個步驟完成後應可獨立驗證：

```
Phase 1: 專案骨架與基礎 (Step 1-3)
  Step 1: Go module 初始化 + 目錄結構 + Makefile
  Step 2: Database interface + SQLite 實作 + Migration
  Step 3: Storage interface + LocalFS 實作

Phase 2: HTTP Server (Step 4-6)
  Step 4: chi router + middleware (auth, logging, recovery)
  Step 5: Health + Publish handler + test
  Step 6: GetSkill + ListVersions + Download + Search handlers + test

Phase 3: CLI 整合 (Step 7-8)
  Step 7: serve command + migrate command (串接 Server)
  Step 8: push / pull / search / init / login (CLI 指令)

Phase 4: PostgreSQL + S3 擴充 (Step 9-10)
  Step 9:  PostgreSQL 實作 (database.Database interface)
  Step 10: S3 Storage 實作 (storage.Storage interface)

Phase 5: 建置與部署 (Step 11-12)
  Step 11: Dockerfile + docker-compose (簡易 + 生產)
  Step 12: 跨平台編譯驗證 + 最終整合測試
```

---

## 13. CLAUDE.md (放在 Repo 根目錄)

以下內容應放在 Repo 根目錄的 `CLAUDE.md`，供 Claude Code 開發時參考：

```markdown
# AgentSkills

AI Agent Skill Registry — CLI + Server for publishing and pulling skill bundles.

## Architecture

- All Go: CLI + Server in single repo, separated by build tags
- CLI binary: `go build .` → `agentskills` (~8MB)
- Server binary: `go build -tags server .` → `agentskills-server` (~18MB)
- DB: SQLite (embedded, default) or PostgreSQL (production)
- Storage: Local filesystem (default) or S3/MinIO (production)
- Spec: See `reference/SDD.md` for complete design document
- Architecture: See `reference/GO-BACKEND-ARCHITECTURE.md` for detailed Go design

## Quick Start

# Start server (SQLite + LocalFS, zero config)
make build-server && ./bin/agentskills-server serve

# Or with Docker
docker compose up -d

## Development

- `make build-cli`: Build CLI binary
- `make build-server`: Build Server binary
- `make test`: Run all tests
- `make build-all`: Cross-platform compilation
- Go 1.22+

## Key Rules

- ALWAYS read reference/SDD.md before making architectural decisions
- DB schema: SQLite in migrations/001_init.sql, PostgreSQL in init.sql
- API responses must match SDD.md §5.2 exactly
- Frontmatter validation rules are in SDD.md §2.3 — implement all of them
- Tests are mandatory for every handler and parser function
- Use database/sql with parameterized queries — no string concatenation
- Server-only code must have `//go:build server` tag
- Never execute uploaded scripts server-side
```

---

*End of Document*
