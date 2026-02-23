# AgentSkills — Go Backend Architecture Design

**Version:** 2.0
**Date:** 2026-02-23
**Author:** LIU YU KAI
**Purpose:** 將後端從 Python/FastAPI 遷移至 Go，實現 CLI / Server 分離 binary 架構 (build tags)。

---

## 1. 設計目標

| 目標 | 說明 |
|------|------|
| **分離 binary** | CLI (~8MB) 和 Server (~18MB) 透過 build tags 分開編譯 |
| **雙模式資料庫** | 開發/個人用 SQLite，生產環境用 PostgreSQL |
| **雙模式儲存** | 開發/個人用本地檔案系統，生產環境用 S3/MinIO |
| **跨平台** | `GOOS=windows/linux/darwin` 一鍵編譯 |
| **API 完全相容** | 所有 REST 端點與原 SDD.md §5.2 規格完全一致 |
| **Docker 極簡** | 最終鏡像 < 30MB (Alpine + static binary) |

---

## 2. 分離 Binary 指令設計

使用 Go build tags 將 CLI 和 Server 分開編譯：

```bash
# === CLI binary (agentskills) ===
# 不含 Server/DB/Storage 依賴，體積小 (~8MB)
agentskills init [name]
agentskills push [path]
agentskills pull <name>[@version]
agentskills search <keyword>
agentskills login
agentskills version

# === Server binary (agentskills-server) ===
# 包含完整 Server + DB + Storage，build tag: server
agentskills-server serve                              # 預設: SQLite + LocalFS, port 8000
agentskills-server serve --port 9000                  # 自訂 port
agentskills-server serve --db postgres://u:p@host/db  # 使用 PostgreSQL
agentskills-server serve --storage s3://endpoint       # 使用 S3/MinIO
agentskills-server migrate                            # 執行資料庫 migration
agentskills-server version
```

### 2.0 Build Tags 策略

```go
// cmd/serve.go
//go:build server

package cmd

// 此檔案只在 -tags server 時編譯
// Server binary 同時包含 CLI 指令 + serve 指令
```

```makefile
# Makefile
build-cli:
    go build -o bin/agentskills .                       # 只有 CLI 指令

build-server:
    go build -tags server -o bin/agentskills-server .    # CLI + Server + DB + Storage
```

### 2.1 環境變數對應

所有 flag 均可透過環境變數設定，方便 Docker 部署：

```bash
AGENTSKILLS_PORT=8000
AGENTSKILLS_DB_DRIVER=sqlite          # sqlite | postgres
AGENTSKILLS_DB_DSN=./data/agentskills.db
AGENTSKILLS_STORAGE_DRIVER=local      # local | s3
AGENTSKILLS_STORAGE_PATH=./data/bundles
AGENTSKILLS_S3_ENDPOINT=http://localhost:9000
AGENTSKILLS_S3_ACCESS_KEY=minioadmin
AGENTSKILLS_S3_SECRET_KEY=minioadmin
AGENTSKILLS_S3_BUCKET=skills
AGENTSKILLS_MAX_BUNDLE_SIZE=52428800  # 50MB
```

---

## 3. 專案結構

```
agentskills/
├── main.go                     # 程式入口
├── go.mod
├── go.sum
├── Makefile                    # 跨平台編譯
├── Dockerfile                  # 多階段 build
├── docker-compose.yml          # 開發環境 (PostgreSQL + MinIO 可選)
├── docker-compose.prod.yml     # 生產環境
├── init.sql                    # PostgreSQL schema (給外部 PG 使用)
│
├── cmd/                        # Cobra 指令定義
│   ├── root.go                 # Root command + global flags
│   ├── serve.go                # agentskills serve (HTTP Server)
│   ├── init_cmd.go             # agentskills init
│   ├── push.go                 # agentskills push
│   ├── pull.go                 # agentskills pull
│   ├── search.go               # agentskills search
│   ├── login.go                # agentskills login
│   ├── migrate.go              # agentskills migrate
│   └── version.go              # agentskills version
│
├── internal/
│   ├── server/                 # HTTP Server 核心
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
│   ├── database/               # 資料庫抽象層
│   │   ├── database.go         # Database interface 定義
│   │   ├── models.go           # Go struct (User, Skill, SkillVersion)
│   │   ├── sqlite.go           # SQLite 實作
│   │   ├── postgres.go         # PostgreSQL 實作
│   │   └── migrate.go          # Schema migration (嵌入 SQL)
│   │
│   ├── storage/                # 儲存抽象層
│   │   ├── storage.go          # Storage interface 定義
│   │   ├── local.go            # 本地檔案系統實作
│   │   └── s3.go               # S3/MinIO 實作
│   │
│   ├── bundle/                 # Bundle 處理
│   │   ├── pack.go             # tar.gz 打包 (CLI 用)
│   │   ├── unpack.go           # tar.gz 解壓 (Server 用)
│   │   └── checksum.go         # SHA-256 計算
│   │
│   ├── parser/                 # SKILL.md 解析
│   │   ├── frontmatter.go      # YAML Frontmatter 解析
│   │   └── validate.go         # 驗證規則 (§2.3)
│   │
│   ├── api/                    # CLI HTTP client
│   │   └── client.go           # 封裝所有 API 呼叫
│   │
│   └── config/                 # 配置管理
│       └── config.go           # CLI config (~/.agentskills/config.yaml)
│
├── migrations/                 # 嵌入式 SQL migrations
│   ├── embed.go                # go:embed 載入 SQL 檔案
│   ├── 001_init.sql            # 初始 schema
│   └── 002_xxx.sql             # 未來 migration
│
└── tests/                      # 整合測試
    ├── integration_test.go     # 端對端測試
    ├── publish_test.go
    ├── pull_test.go
    ├── search_test.go
    └── testdata/               # 測試用 skill bundles
        └── valid-skill/
            └── SKILL.md
```

---

## 4. 核心 Interface 設計

### 4.1 Database Interface

```go
package database

import (
    "context"
    "time"
)

// ============================================
// Models
// ============================================

type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    APIToken  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
}

type Skill struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    OwnerID   string    `json:"owner_id"`
    Downloads int64     `json:"downloads"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type SkillVersion struct {
    ID          string                 `json:"id"`
    SkillID     string                 `json:"skill_id"`
    Version     string                 `json:"version"`
    BundleKey   string                 `json:"bundle_key"`
    Metadata    map[string]interface{} `json:"metadata"`
    Checksum    string                 `json:"checksum"`
    SizeBytes   int64                  `json:"size_bytes"`
    PublishedAt time.Time              `json:"published_at"`
}

// ============================================
// Database Interface
// ============================================

type Database interface {
    // Lifecycle
    Open() error
    Close() error
    Migrate() error

    // Users
    GetUserByToken(ctx context.Context, token string) (*User, error)

    // Skills
    GetSkillByName(ctx context.Context, name string) (*Skill, error)
    CreateSkill(ctx context.Context, name string, ownerID string) (*Skill, error)
    IncrementDownloads(ctx context.Context, skillID string) error
    UpdateSkillTimestamp(ctx context.Context, skillID string) error
    SearchSkills(ctx context.Context, query string, tag string, page int, perPage int) ([]SkillSearchResult, int, error)

    // Versions
    CreateVersion(ctx context.Context, v *SkillVersion) error
    GetVersion(ctx context.Context, skillID string, version string) (*SkillVersion, error)
    GetLatestVersion(ctx context.Context, skillID string) (*SkillVersion, error)
    ListVersions(ctx context.Context, skillID string) ([]SkillVersion, error)
}

type SkillSearchResult struct {
    Name          string   `json:"name"`
    Description   string   `json:"description"`
    Owner         string   `json:"owner"`
    Downloads     int64    `json:"downloads"`
    LatestVersion string   `json:"latest_version"`
    UpdatedAt     string   `json:"updated_at"`
    Tags          []string `json:"tags"`
}
```

### 4.2 Storage Interface

```go
package storage

import (
    "context"
    "io"
)

type Storage interface {
    // Lifecycle
    Init() error
    HealthCheck(ctx context.Context) error

    // Operations
    Upload(ctx context.Context, key string, reader io.Reader, size int64) error
    Download(ctx context.Context, key string) (io.ReadCloser, int64, error)
    Exists(ctx context.Context, key string) (bool, error)
    Delete(ctx context.Context, key string) error
}
```

---

## 5. 雙模式實作細節

### 5.1 SQLite 模式 (預設)

```go
// internal/database/sqlite.go
package database

import (
    "database/sql"
    _ "modernc.org/sqlite"   // 純 Go SQLite，無 CGO 依賴
)

type SQLiteDB struct {
    db   *sql.DB
    path string
}

func NewSQLite(dsn string) *SQLiteDB {
    return &SQLiteDB{path: dsn}
}

func (s *SQLiteDB) Open() error {
    db, err := sql.Open("sqlite", s.path+"?_journal_mode=WAL&_busy_timeout=5000")
    if err != nil {
        return err
    }
    s.db = db
    return nil
}
```

**關鍵設定:**
- WAL mode: 允許並發讀取
- Busy timeout: 避免 database locked 錯誤
- `modernc.org/sqlite`: 純 Go 實作，跨平台編譯無需 CGO

### 5.2 PostgreSQL 模式

```go
// internal/database/postgres.go
package database

import (
    "database/sql"
    _ "github.com/lib/pq"
)

type PostgresDB struct {
    db  *sql.DB
    dsn string
}

func NewPostgres(dsn string) *PostgresDB {
    return &PostgresDB{dsn: dsn}
}
```

### 5.3 Local Storage 模式 (預設)

```go
// internal/storage/local.go
package storage

import (
    "io"
    "os"
    "path/filepath"
)

type LocalStorage struct {
    basePath string  // 例如 ./data/bundles/
}

func NewLocalStorage(basePath string) *LocalStorage {
    return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
    path := filepath.Join(l.basePath, key)
    os.MkdirAll(filepath.Dir(path), 0755)
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, reader)
    return err
}
```

### 5.4 S3 Storage 模式

```go
// internal/storage/s3.go
package storage

import (
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
    client *s3.Client
    bucket string
}
```

---

## 6. 工廠模式初始化

```go
// cmd/serve.go — 啟動邏輯

func newServer(cfg *ServeConfig) (*server.Server, error) {
    // 1. 初始化 Database
    var db database.Database
    switch cfg.DBDriver {
    case "sqlite":
        db = database.NewSQLite(cfg.DBDSN)
    case "postgres":
        db = database.NewPostgres(cfg.DBDSN)
    default:
        return nil, fmt.Errorf("unsupported db driver: %s", cfg.DBDriver)
    }
    if err := db.Open(); err != nil {
        return nil, err
    }
    if err := db.Migrate(); err != nil {
        return nil, err
    }

    // 2. 初始化 Storage
    var store storage.Storage
    switch cfg.StorageDriver {
    case "local":
        store = storage.NewLocalStorage(cfg.StoragePath)
    case "s3":
        store = storage.NewS3Storage(cfg.S3Config)
    default:
        return nil, fmt.Errorf("unsupported storage driver: %s", cfg.StorageDriver)
    }
    if err := store.Init(); err != nil {
        return nil, err
    }

    // 3. 建立 Server
    return server.New(db, store, cfg.Port), nil
}
```

---

## 7. API 路由對應

使用 `go-chi/chi` 作為 router (輕量、相容 net/http):

```go
// internal/server/router.go

func (s *Server) setupRoutes() {
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)

    r.Route("/v1", func(r chi.Router) {
        // Public endpoints
        r.Get("/health", s.handlers.Health)
        r.Get("/skills", s.handlers.SearchSkills)
        r.Get("/skills/{name}", s.handlers.GetSkill)
        r.Get("/skills/{name}/versions", s.handlers.ListVersions)
        r.Get("/skills/{name}/versions/{version}/download", s.handlers.Download)

        // Protected endpoints
        r.Group(func(r chi.Router) {
            r.Use(s.middleware.Auth)
            r.Post("/skills/publish", s.handlers.Publish)
        })
    })

    s.router = r
}
```

**所有 API 回應格式與原 SDD.md §5.2 完全一致，無任何改動。**

---

## 8. Schema Migration 策略

使用 `go:embed` 嵌入 SQL 檔案，binary 自帶 migration：

```go
// migrations/embed.go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

```sql
-- migrations/001_init.sql
-- SQLite 版本 (與 PostgreSQL 的差異由程式碼處理)

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

```sql
-- init.sql (PostgreSQL 版本，保持與原 SDD 完全一致)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(64)  UNIQUE NOT NULL,
    api_token   VARCHAR(128) UNIQUE NOT NULL,
    created_at  TIMESTAMPTZ  DEFAULT now()
);

-- ... 其餘與原 SDD.md §4.1 完全相同
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

### 9.2 Docker (多階段 Build)

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

# 預設資料目錄
RUN mkdir -p /data/bundles
VOLUME ["/data"]

EXPOSE 8000

ENTRYPOINT ["agentskills-server"]
CMD ["serve", "--port", "8000", "--db", "sqlite:///data/agentskills.db", "--storage", "local:///data/bundles"]
```

**最終鏡像大小估算:** ~25MB (Alpine 7MB + Go binary ~18MB)

---

## 10. Docker Compose 配置

### 10.1 簡易模式 (SQLite + LocalFS)

```yaml
# docker-compose.yml — 最簡單的啟動方式
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

### 10.2 生產模式 (PostgreSQL + MinIO)

```yaml
# docker-compose.prod.yml
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

---

## 11. Go 核心依賴

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

---

## 12. 與原 SDD 的差異對照

| 項目 | 原 SDD (Python) | 新架構 (Go) | 影響 |
|------|-----------------|-------------|------|
| 後端語言 | Python 3.12+ / FastAPI | Go 1.22+ / chi | 核心改動 |
| CLI 語言 | Go / Cobra | Go / Cobra (不變) | 無影響 |
| DB Driver | asyncpg (async) | database/sql + modernc/sqlite 或 lib/pq | 同步但高效 |
| ORM | SQLAlchemy 2.0 | 手寫 SQL (database/sql) | 更直接 |
| S3 Client | boto3 | aws-sdk-go-v2 | 功能等價 |
| 嵌入式 DB | 不支援 | SQLite (純 Go) | **新增能力** |
| 本地儲存 | 不支援 | LocalFS | **新增能力** |
| Binary 數量 | 2 (Python app + Go CLI) | 2 Go binary (CLI ~8MB + Server ~18MB, build tags) | **核心改善** |
| Docker 鏡像 | ~200MB (Python) | ~25MB (Alpine + Go) | **體積降 87%** |
| API 規格 | SDD §5.2 | **完全相同** | 無影響 |
| DB Schema | SDD §4.1 | **完全相同** (+ SQLite 版) | 向下相容 |
| 安全防護 | SDD §10 | **完全相同** | 無影響 |

---

## 13. 部署場景對照

```
場景 A: 個人開發者，本機使用
  $ ./agentskills serve
  → SQLite: ./data/agentskills.db
  → Bundles: ./data/bundles/
  → 完全不需要 Docker、PostgreSQL、MinIO

場景 B: 團隊開發，Docker Compose
  $ docker compose up -d
  → 單一容器，SQLite + LocalFS
  → 或 docker compose -f docker-compose.prod.yml up -d (PG + MinIO)

場景 C: Windows 使用者
  > agentskills-windows-amd64.exe serve
  → 雙擊或命令列啟動，完全不需要安裝任何東西

場景 D: 生產環境 (Kubernetes)
  → Helm chart: agentskills container + 外部 PostgreSQL + 外部 S3
  → 環境變數配置，12-factor app
```

---

## 14. 實作順序

```
Phase 1: 專案骨架與基礎 (Step 1-3)
  Step 1: Go module 初始化 + 目錄結構 + Makefile
  Step 2: Database interface + SQLite 實作 + Migration
  Step 3: Storage interface + LocalFS 實作

Phase 2: HTTP Server (Step 4-6)
  Step 4: chi router + middleware (auth, logging, recovery)
  Step 5: Health + Publish handler
  Step 6: GetSkill + ListVersions + Download + Search handlers

Phase 3: CLI 整合 (Step 7-8)
  Step 7: serve command (串接 Server)
  Step 8: push / pull / search / init / login (CLI 指令)

Phase 4: PostgreSQL + S3 擴充 (Step 9-10)
  Step 9:  PostgreSQL 實作 (database.Database interface)
  Step 10: S3 Storage 實作 (storage.Storage interface)

Phase 5: 建置與部署 (Step 11-12)
  Step 11: Dockerfile + docker-compose (簡易 + 生產)
  Step 12: Makefile 跨平台編譯 + CI/CD

Phase 6: 測試 (貫穿所有階段)
  每個 Step 完成後撰寫對應的 unit test + integration test
```

---

*End of Architecture Design*
