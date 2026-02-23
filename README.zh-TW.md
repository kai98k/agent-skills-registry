# AgentSkills

AI Agent Skill 的集中式 Registry 平台 — 類似 npm 或 Docker Hub，但專為 AI Agent Skills 設計。

開發者可透過 CLI 工具上傳（push）與下載（pull）標準化的 Skill Bundle，平台負責版本控制、Metadata 解析與檔案儲存。

**[English README](README.md)**

---

## 為什麼需要 AgentSkills Registry？

### 背景：Agent Skills 正在成為 AI 基礎設施的核心

2025 年，AI Agent 的發展進入了一個關鍵轉折點。Anthropic（Claude）、Google（Gemini）等主要平台不約而同地推出了 **Agent Skills** 機制 — 一種讓 AI Agent 從「通才」變成「專家」的模組化擴展格式。

[Agent Skills 規範](https://agentskills.io)定義了一個簡單而強大的結構：一個 `SKILL.md` 檔案加上可選的腳本、參考文件和靜態資源。透過這個格式，任何人都可以將領域知識打包成可重複使用的「能力模組」，讓 AI Agent 在需要時自動載入並執行。

這就像是為 AI Agent 建立了一套「外掛系統」：

- Claude Code 可以在本地檔案系統中自動發現並使用 Skills
- Claude API 可以透過上傳機制載入 Skills
- Claude.ai 使用者可以在設定中安裝 Skills
- 任何相容的 Agent 框架都可以整合 Skills

### 問題：Skills 到處都是，卻無處可尋

但是，目前 Agent Skills 生態圈面臨著一個 **npm 誕生前 JavaScript 社群曾面臨的相同問題**：

**沒有集中的發現機制。** 想找一個「程式碼審查」的 Skill？你得翻遍 GitHub、部落格、Discord 社群。沒有一個地方可以搜尋「code-review skill」然後立刻找到並安裝。

**沒有版本保證。** 你從某個 repo 複製了一個 Skill，下個月作者改了內容，你的 Agent 行為就默默地變了。沒有 immutable versioning，沒有 changelog，沒有升級路徑。

**沒有完整性驗證。** 從網路上下載一個 `.tar.gz` 檔案，你怎麼確定它沒有被竄改？沒有 checksum 驗證，沒有簽名機制，安全性形同虛設。

**平台孤島。** Claude Code 的 Skills 存在本地檔案系統、Claude API 用上傳 endpoint、Claude.ai 用 zip 上傳。每個平台都是一座孤島，技能無法跨平台流通。

**缺乏標準化工作流程。** 沒有 `init`、`push`、`pull`、`search` 這樣的標準指令，每個團隊都在重新發明輪子。

### 解法：為 AI 時代打造套件管理基礎設施

AgentSkills Registry 解決這些問題的方式，和當年 npm 為 JavaScript、Docker Hub 為容器鏡像解決問題的方式如出一轍：

| npm 為 JS 做了什麼 | AgentSkills 為 Agent Skills 做什麼 |
|--------------------|------------------------------------|
| `npm publish` / `npm install` | `agentskills push` / `agentskills pull` |
| package.json + semver | SKILL.md frontmatter + 嚴格 semver |
| SHA integrity check | SHA-256 checksum 驗證每個 bundle |
| npmjs.com 搜尋 | `agentskills search` 依關鍵字與 tag 搜尋 |
| Scoped packages (`@org/pkg`) | 作者綁定 Skills（owner = API token 持有者） |

### 為什麼是現在？

**時機至關重要。** Agent Skills 規範還很年輕，生態圈正在快速成型。現在建立基礎設施，就像在 2010 年建立 npm — 在生態圈爆發前把基礎打好，之後所有人都能受益。

**開源驅動信任。** 一個由社群驅動的開源 Registry，比起各家廠商各自為政的封閉系統，更能促進 Skills 的跨平台流通和品質提升。

**降低入門門檻。** 個人開發者只需要一個 25MB 的 binary 就能架設私有 Registry；團隊可以用 Docker Compose 一鍵部署；企業可以接入 PostgreSQL + S3 擴展到生產規模。從個人到企業，同一套工具，零摩擦升級。

### 一句話總結

> **AgentSkills Registry 是 Agent Skills 生態圈中缺失的基礎設施層 — 讓散落各處的 Skill 檔案變成可發現、可版控、可驗證、可分享的標準化套件。**

---

## 功能一覽

| 功能 | 說明 |
|------|------|
| **Skill 發布 (push)** | 將本地 Skill 目錄打包為 `.tar.gz` 並上傳至 Registry |
| **Skill 下載 (pull)** | 從 Registry 下載指定 Skill（支援指定版本或自動取最新版） |
| **Skill 搜尋 (search)** | 以關鍵字或 tag 搜尋平台上的 Skills |
| **Skill 初始化 (init)** | 快速建立 Skill 骨架目錄與模板 |
| **版本控制** | 嚴格 Semantic Versioning，每個版本 immutable 不可覆寫 |
| **Checksum 驗證** | SHA-256 校驗確保上傳與下載的完整性 |
| **雙模式資料庫** | SQLite（嵌入式，零配置）或 PostgreSQL（生產環境） |
| **雙模式儲存** | 本地檔案系統（零配置）或 S3/MinIO（生產環境） |
| **跨平台** | 支援 Linux / macOS / Windows，單一 binary 零依賴 |
| **Docker 部署** | 25MB 極小鏡像，一鍵啟動 |

---

## 安裝方式

### 方式一：下載預編譯 Binary（推薦）

從 [Releases](../../releases) 頁面下載對應平台的 binary：

**CLI（給 Skill 開發者）：**

| 平台 | 檔案 |
|------|------|
| Linux (x64) | `agentskills-linux-amd64` |
| Linux (ARM64) | `agentskills-linux-arm64` |
| macOS (Intel) | `agentskills-darwin-amd64` |
| macOS (Apple Silicon) | `agentskills-darwin-arm64` |
| Windows (x64) | `agentskills-windows-amd64.exe` |

**Server（給 Registry 管理員）：**

| 平台 | 檔案 |
|------|------|
| Linux (x64) | `agentskills-server-linux-amd64` |
| Linux (ARM64) | `agentskills-server-linux-arm64` |
| macOS (Intel) | `agentskills-server-darwin-amd64` |
| macOS (Apple Silicon) | `agentskills-server-darwin-arm64` |
| Windows (x64) | `agentskills-server-windows-amd64.exe` |

```bash
# Linux / macOS 範例
curl -LO https://github.com/liuyukai/agentskills/releases/latest/download/agentskills-linux-amd64
chmod +x agentskills-linux-amd64
sudo mv agentskills-linux-amd64 /usr/local/bin/agentskills
```

```powershell
# Windows — 下載 .exe 後直接執行，不需要安裝
# 或加入 PATH 環境變數
```

### 方式二：Docker 鏡像（推薦用於 Server）

鏡像同時發布至 Docker Hub 和 GitHub Container Registry：

```bash
# Docker Hub
docker pull kai98k/agentskills-server:latest

# GitHub Container Registry
docker pull ghcr.io/kai98k/agentskills-server:latest
```

可用的 image tag：

| Tag | 說明 |
|-----|------|
| `latest` | 最新穩定版 |
| `v1.0.0` | 指定版本 |
| `sha-abc1234` | 指定 commit |

---

**簡易模式** — SQLite + 本地儲存，零配置，建立 `docker-compose.yml`：

```yaml
# docker-compose.yml
services:
  agentskills:
    image: kai98k/agentskills-server:latest
    # 或使用 ghcr.io:
    # image: ghcr.io/kai98k/agentskills-server:latest
    # 或從原始碼 build:
    # build: .
    ports:
      - "8000:8000"
    volumes:
      - data:/data
    # 預設 SQLite + 本地檔案系統，不需要任何環境變數

volumes:
  data:
```

```bash
docker compose up -d
curl http://localhost:8000/v1/health
# {"status":"ok","database":"connected","storage":"connected"}
```

---

**生產模式** — PostgreSQL + MinIO，建立 `docker-compose.prod.yml`：

```yaml
# docker-compose.prod.yml
services:
  # ── PostgreSQL ──────────────────────────────
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

  # ── MinIO (S3 相容儲存) ─────────────────────
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD}
    ports:
      - "9000:9000"    # S3 API
      - "9001:9001"    # Web Console
    volumes:
      - miniodata:/data
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 5s
      timeout: 5s
      retries: 5

  # ── MinIO 初始化 (自動建立 Bucket) ──────────
  minio-init:
    image: minio/mc:latest
    depends_on:
      minio:
        condition: service_healthy
    entrypoint: >
      /bin/sh -c "
      mc alias set local http://minio:9000 $${MINIO_USER} $${MINIO_PASSWORD};
      mc mb --ignore-existing local/skills;
      echo 'Bucket [skills] created';
      "

  # ── AgentSkills Server ──────────────────────
  agentskills:
    image: kai98k/agentskills-server:latest
    # 或使用 ghcr.io:
    # image: ghcr.io/kai98k/agentskills-server:latest
    # 或從原始碼 build:
    # build: .
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

```bash
# 建立 .env 檔案設定密碼
cat > .env << 'EOF'
PG_PASSWORD=your-secure-password
MINIO_USER=minioadmin
MINIO_PASSWORD=minioadmin
EOF

# 啟動
docker compose -f docker-compose.prod.yml up -d

# 驗證
curl http://localhost:8000/v1/health
```

---

**Dockerfile**（若要從原始碼自行 build）：

```dockerfile
# === Build Stage ===
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -tags server -ldflags="-s -w" -o /agentskills-server .

# === Runtime Stage (最終鏡像 ~25MB) ===
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /agentskills-server /usr/local/bin/agentskills-server
RUN mkdir -p /data/bundles
VOLUME ["/data"]
EXPOSE 8000
ENTRYPOINT ["agentskills-server"]
CMD ["serve", "--port", "8000"]
```

### 方式三：從原始碼編譯

需要 Go 1.22+：

```bash
git clone https://github.com/liuyukai/agentskills.git
cd agentskills

# 編譯 CLI
make build-cli
# → bin/agentskills

# 編譯 Server
make build-server
# → bin/agentskills-server

# 編譯所有平台
make build-all
# → bin/ 下包含 Linux / macOS / Windows 版本
```

---

## 快速開始

### 1. 啟動 Server

```bash
# 方式 A: 直接執行（SQLite + 本地儲存，零配置）
./agentskills-server serve

# 方式 B: 指定 port
./agentskills-server serve --port 9000

# 方式 C: 使用 PostgreSQL + S3
./agentskills-server serve \
  --db postgres://user:pass@localhost:5432/agentskills \
  --storage s3://localhost:9000

# 方式 D: Docker
docker compose up -d
```

Server 啟動後預設監聽 `http://localhost:8000`。

### 2. 設定 CLI

```bash
# 設定 Server 位址和 API Token
agentskills login
# Enter API URL: http://localhost:8000
# Enter API token: ********
# Token saved to ~/.agentskills/config.yaml
```

預設開發帳號：token 為 `dev-token-12345`

### 3. 建立第一個 Skill

```bash
# 初始化 Skill 骨架
agentskills init my-first-skill

# 編輯 SKILL.md，填入描述和指令內容
cd my-first-skill
# ... 編輯 SKILL.md ...
```

### 4. 發布 Skill

```bash
agentskills push ./my-first-skill

# Validating SKILL.md...        ✓
# Packing bundle...             ✓ (12.3 KB)
# Uploading my-first-skill@0.1.0...   ✓
# Checksum: sha256:a1b2c3d4...
#
# Published my-first-skill@0.1.0 successfully.
```

### 5. 下載 Skill

```bash
# 下載最新版
agentskills pull my-first-skill

# 下載指定版本
agentskills pull my-first-skill@0.1.0
```

### 6. 搜尋 Skill

```bash
agentskills search code-review

# NAME                  VERSION  DOWNLOADS  DESCRIPTION
# code-review-agent     1.2.0    42         PR code review skill
# code-review-lite      0.3.0    7          Lightweight review helper
```

---

## CLI 指令速查

| 指令 | 說明 | 範例 |
|------|------|------|
| `agentskills init [name]` | 建立 Skill 骨架 | `agentskills init my-skill` |
| `agentskills push [path]` | 打包上傳 Skill | `agentskills push ./my-skill` |
| `agentskills pull <name>[@ver]` | 下載 Skill | `agentskills pull my-skill@1.0.0` |
| `agentskills search <keyword>` | 搜尋 Skills | `agentskills search code-review` |
| `agentskills login` | 設定 API Token | `agentskills login` |
| `agentskills version` | 顯示版本 | `agentskills version` |

---

## Server 指令速查

| 指令 | 說明 | 範例 |
|------|------|------|
| `agentskills-server serve` | 啟動 HTTP Server | `agentskills-server serve --port 8000` |
| `agentskills-server migrate` | 執行資料庫 migration | `agentskills-server migrate` |
| `agentskills-server version` | 顯示版本 | `agentskills-server version` |

### Server 環境變數

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `AGENTSKILLS_PORT` | `8000` | HTTP 監聽 port |
| `AGENTSKILLS_DB_DRIVER` | `sqlite` | 資料庫類型：`sqlite` 或 `postgres` |
| `AGENTSKILLS_DB_DSN` | `./data/agentskills.db` | 資料庫連線字串 |
| `AGENTSKILLS_STORAGE_DRIVER` | `local` | 儲存類型：`local` 或 `s3` |
| `AGENTSKILLS_STORAGE_PATH` | `./data/bundles` | 本地儲存路徑 |
| `AGENTSKILLS_S3_ENDPOINT` | - | S3/MinIO endpoint |
| `AGENTSKILLS_S3_ACCESS_KEY` | - | S3 access key |
| `AGENTSKILLS_S3_SECRET_KEY` | - | S3 secret key |
| `AGENTSKILLS_S3_BUCKET` | `skills` | S3 bucket 名稱 |
| `AGENTSKILLS_MAX_BUNDLE_SIZE` | `52428800` | Bundle 最大大小 (bytes, 預設 50MB) |

---

## API 端點

Base URL: `http://localhost:8000/v1`

| Method | 端點 | 說明 | 認證 |
|--------|------|------|------|
| `GET` | `/v1/health` | 健康檢查 | 不需要 |
| `POST` | `/v1/skills/publish` | 上傳 Skill Bundle | Bearer Token |
| `GET` | `/v1/skills/{name}` | 取得 Skill 資訊 + 最新版本 | 不需要 |
| `GET` | `/v1/skills/{name}/versions` | 列出所有版本 | 不需要 |
| `GET` | `/v1/skills/{name}/versions/{ver}/download` | 下載指定版本 | 不需要 |
| `GET` | `/v1/skills?q=keyword&tag=tag` | 搜尋 Skills | 不需要 |

認證方式：HTTP Header `Authorization: Bearer <your-token>`

---

## Skill Bundle 格式

每個 Skill 是一個目錄，核心是 `SKILL.md` 檔案：

```
my-skill/
├── SKILL.md         (必填) YAML Frontmatter + Markdown 指令
├── scripts/         (選填) Agent 可呼叫的腳本
├── references/      (選填) 參考文件
└── assets/          (選填) 靜態資源
```

### SKILL.md 格式

```yaml
---
name: "my-skill"                    # 必填, 全域唯一, [a-z0-9-], 3-64 字元
version: "1.0.0"                    # 必填, 嚴格 semver
description: "My awesome skill"     # 必填, 最長 256 字元
author: "username"                  # 必填, 與 API Token 帳號一致
tags:                               # 選填, 最多 10 個
  - tag1
  - tag2
license: "MIT"                      # 選填, SPDX identifier
---

# My Skill

這裡寫 Agent 的指令內容...
```

---

## 部署場景

### 場景 A：個人開發者（最簡單）

```bash
# 下載 binary → 啟動 → 完成
./agentskills-server serve
# 資料存在 ./data/ 目錄，SQLite + 本地檔案，零配置
```

### 場景 B：團隊 / 小型組織

```bash
# 使用上方「簡易模式」的 docker-compose.yml
docker compose up -d
# 25MB 鏡像，SQLite + 本地儲存，自動初始化
```

### 場景 C：生產環境

```bash
# 使用上方「生產模式」的 docker-compose.prod.yml
# PostgreSQL + MinIO，完整生產配置
docker compose -f docker-compose.prod.yml up -d
```

### 場景 D：Windows 使用者

```powershell
# 下載 .exe → 雙擊或命令列啟動
agentskills-server-windows-amd64.exe serve
# 不需要安裝任何東西
```

---

## 技術架構

```
                                  ┌─── SQLite (嵌入式, 預設)
                                  │
agentskills-server ──── Database ─┤
     (Go binary)        Interface │
                                  └─── PostgreSQL (生產)

                                  ┌─── Local FS (預設)
                                  │
                         Storage ─┤
                        Interface │
                                  └─── S3/MinIO (生產)
```

- **語言**：Go 1.22+
- **HTTP Router**：go-chi/chi
- **CLI Framework**：spf13/cobra
- **嵌入式 DB**：modernc.org/sqlite（純 Go，無 CGO 依賴）
- **Build 策略**：Go build tags 分離 CLI / Server

---

## 開發指南

```bash
# 環境需求
# - Go 1.22+
# - (可選) Docker & Docker Compose

# 編譯
make build-cli        # CLI binary
make build-server     # Server binary
make build-all        # 所有平台

# 測試
make test             # 執行所有測試

# 啟動開發 Server
./bin/agentskills-server serve

# Docker 開發
docker compose up -d
```

詳細設計規格請參考 [`reference/SDD.md`](reference/SDD.md)。

---

## CI/CD 自動發布

專案使用 GitHub Actions 自動化建置與發布：

### 自動觸發流程

| 事件 | 觸發的 Workflow | 動作 |
|------|----------------|------|
| Push / PR 到 `main` | `ci.yml` | 執行測試 + 驗證 build + 驗證 Docker build |
| 推送 tag `v*` | `release.yml` | 測試 → 跨平台編譯 → Docker 鏡像推送 → GitHub Release |

### Release 流程

```bash
# 1. 打 tag
git tag v1.0.0
git push origin v1.0.0

# 2. GitHub Actions 自動：
#    - 執行測試
#    - 編譯 10 個 binary (5 平台 x CLI/Server)
#    - 建置 Docker 鏡像 (linux/amd64 + linux/arm64)
#    - 推送鏡像至 Docker Hub + GitHub Container Registry
#    - 建立 GitHub Release + 上傳 binary + SHA256 checksum
```

### 需要設定的 GitHub Secrets

在 GitHub repo → Settings → Secrets and variables → Actions 中設定：

| Secret | 說明 | 取得方式 |
|--------|------|---------|
| `DOCKERHUB_USERNAME` | Docker Hub 帳號 | [hub.docker.com](https://hub.docker.com) 註冊 |
| `DOCKERHUB_TOKEN` | Docker Hub Access Token | Docker Hub → Account Settings → Security → New Access Token |
| `GITHUB_TOKEN` | GitHub Token (自動提供) | 不需要手動設定，GitHub Actions 自帶 |

### Docker 鏡像 Tag 規則

推送 `v1.2.3` tag 後，會自動產生以下 Docker image tag：

```
kai98k/agentskills-server:1.2.3
kai98k/agentskills-server:1.2
kai98k/agentskills-server:1
kai98k/agentskills-server:latest
kai98k/agentskills-server:sha-abc1234

ghcr.io/kai98k/agentskills-server:1.2.3
ghcr.io/kai98k/agentskills-server:1.2
ghcr.io/kai98k/agentskills-server:1
ghcr.io/kai98k/agentskills-server:latest
ghcr.io/kai98k/agentskills-server:sha-abc1234
```

---

## License

MIT
