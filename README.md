# AgentSkills

A centralized Registry platform for AI Agent Skills — like npm or Docker Hub, but purpose-built for Agent Skills.

Developers can publish (push) and download (pull) standardized Skill Bundles via a CLI tool. The platform handles version control, metadata parsing, and file storage.

**[繁體中文版 README](README.zh-TW.md)**

---

## Why AgentSkills Registry?

### The problem: Skills are everywhere, but nowhere to find

AI Agent Skills — modular instruction sets that turn general-purpose agents into domain specialists — are rapidly becoming a core building block of the AI ecosystem. Anthropic's Claude, Google's Gemini, and others have all adopted the [Agent Skills specification](https://agentskills.io) as a standard way to extend agent capabilities.

But today, sharing and discovering Skills is fragmented:

- **No central discovery** — Skills are scattered across GitHub repos, blog posts, and internal wikis. There's no single place to search for "a code-review skill" or "a PDF-processing skill."
- **No versioning guarantee** — Without a registry enforcing immutable semantic versions, a skill you depend on could silently change or disappear.
- **No integrity verification** — Downloading a `.tar.gz` from a random URL offers no checksum validation. You can't be sure the bundle hasn't been tampered with.
- **Platform silos** — Claude Code stores skills on the local filesystem, the Claude API uses upload endpoints, and Claude.ai uses zip uploads. Each surface is an island with no cross-platform sharing.

### The solution: a package registry for the AI age

AgentSkills Registry solves these problems the same way npm solved them for JavaScript and Docker Hub solved them for container images:

| What npm did for JS | What AgentSkills does for Agent Skills |
|---------------------|---------------------------------------|
| `npm publish` / `npm install` | `agentskills push` / `agentskills pull` |
| package.json + semver | SKILL.md frontmatter + strict semver |
| SHA integrity check | SHA-256 checksum on every bundle |
| npmjs.com search | `agentskills search` by keyword & tag |
| Scoped packages (`@org/pkg`) | Author-scoped skills (owner = API token holder) |

**In short:** AgentSkills Registry is the missing infrastructure layer that turns ad-hoc skill files into a proper ecosystem — discoverable, versioned, verified, and shareable.

### Who is this for?

- **Skill authors** who want to publish reusable skills for the community
- **AI developers** who want to find and integrate battle-tested skills instead of writing from scratch
- **Teams & organizations** who want a private registry to share internal skills across projects
- **Platform builders** integrating skills into their own agent frameworks

---

## Features

| Feature | Description |
|---------|-------------|
| **Skill Publish (push)** | Pack a local Skill directory into `.tar.gz` and upload to the Registry |
| **Skill Download (pull)** | Download a specific Skill (supports version pinning or latest) |
| **Skill Search (search)** | Search by keyword or tag |
| **Skill Init (init)** | Scaffold a new Skill directory with templates |
| **Version Control** | Strict Semantic Versioning; every version is immutable |
| **Checksum Verification** | SHA-256 ensures upload/download integrity |
| **Dual-Mode Database** | SQLite (embedded, zero-config) or PostgreSQL (production) |
| **Dual-Mode Storage** | Local filesystem (zero-config) or S3/MinIO (production) |
| **Cross-Platform** | Linux / macOS / Windows, single binary with zero dependencies |
| **Docker Deployment** | 25 MB minimal image, one-command startup |

---

## Installation

### Option 1: Pre-compiled Binaries (Recommended)

Download from the [Releases](../../releases) page.

**CLI (for Skill developers):**

| Platform | File |
|----------|------|
| Linux (x64) | `agentskills-linux-amd64` |
| Linux (ARM64) | `agentskills-linux-arm64` |
| macOS (Intel) | `agentskills-darwin-amd64` |
| macOS (Apple Silicon) | `agentskills-darwin-arm64` |
| Windows (x64) | `agentskills-windows-amd64.exe` |

**Server (for Registry administrators):**

| Platform | File |
|----------|------|
| Linux (x64) | `agentskills-server-linux-amd64` |
| Linux (ARM64) | `agentskills-server-linux-arm64` |
| macOS (Intel) | `agentskills-server-darwin-amd64` |
| macOS (Apple Silicon) | `agentskills-server-darwin-arm64` |
| Windows (x64) | `agentskills-server-windows-amd64.exe` |

```bash
# Linux / macOS
curl -LO https://github.com/liuyukai/agentskills/releases/latest/download/agentskills-linux-amd64
chmod +x agentskills-linux-amd64
sudo mv agentskills-linux-amd64 /usr/local/bin/agentskills
```

```powershell
# Windows — download the .exe and run directly, no installation needed
```

### Option 2: Docker Image (Recommended for Server)

Images are published to both Docker Hub and GitHub Container Registry:

```bash
# Docker Hub
docker pull kai98k/agentskills-server:latest

# GitHub Container Registry
docker pull ghcr.io/kai98k/agentskills-server:latest
```

Available image tags:

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `v1.0.0` | Specific version |
| `sha-abc1234` | Specific commit |

---

**Simple mode** — SQLite + local storage, zero-config. Create a `docker-compose.yml`:

```yaml
# docker-compose.yml
services:
  agentskills:
    image: kai98k/agentskills-server:latest
    # Or use ghcr.io:
    # image: ghcr.io/kai98k/agentskills-server:latest
    # Or build from source:
    # build: .
    ports:
      - "8000:8000"
    volumes:
      - data:/data
    # Defaults to SQLite + local filesystem, no env vars needed

volumes:
  data:
```

```bash
docker compose up -d
curl http://localhost:8000/v1/health
# {"status":"ok","database":"connected","storage":"connected"}
```

---

**Production mode** — PostgreSQL + MinIO. Create a `docker-compose.prod.yml`:

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

  # ── MinIO (S3-compatible storage) ───────────
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

  # ── MinIO Init (auto-create bucket) ─────────
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
    # Or use ghcr.io:
    # image: ghcr.io/kai98k/agentskills-server:latest
    # Or build from source:
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
# Create .env file
cat > .env << 'EOF'
PG_PASSWORD=your-secure-password
MINIO_USER=minioadmin
MINIO_PASSWORD=minioadmin
EOF

# Start
docker compose -f docker-compose.prod.yml up -d

# Verify
curl http://localhost:8000/v1/health
```

---

**Dockerfile** (to build from source):

```dockerfile
# === Build Stage ===
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -tags server -ldflags="-s -w" -o /agentskills-server .

# === Runtime Stage (~25MB) ===
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /agentskills-server /usr/local/bin/agentskills-server
RUN mkdir -p /data/bundles
VOLUME ["/data"]
EXPOSE 8000
ENTRYPOINT ["agentskills-server"]
CMD ["serve", "--port", "8000"]
```

### Option 3: Build from Source

Requires Go 1.22+:

```bash
git clone https://github.com/liuyukai/agentskills.git
cd agentskills

# Build CLI
make build-cli
# → bin/agentskills

# Build Server
make build-server
# → bin/agentskills-server

# Build all platforms
make build-all
# → bin/ contains Linux / macOS / Windows binaries
```

---

## Quick Start

### 1. Start the Server

```bash
# Option A: Direct execution (SQLite + local storage, zero-config)
./agentskills-server serve

# Option B: Custom port
./agentskills-server serve --port 9000

# Option C: PostgreSQL + S3
./agentskills-server serve \
  --db postgres://user:pass@localhost:5432/agentskills \
  --storage s3://localhost:9000

# Option D: Docker
docker compose up -d
```

The server listens on `http://localhost:8000` by default.

### 2. Configure the CLI

```bash
# Set server URL and API token
agentskills login
# Enter API URL: http://localhost:8000
# Enter API token: ********
# Token saved to ~/.agentskills/config.yaml
```

Default dev account token: `dev-token-12345`

### 3. Create Your First Skill

```bash
# Scaffold a new Skill
agentskills init my-first-skill

# Edit SKILL.md with your description and instructions
cd my-first-skill
# ... edit SKILL.md ...
```

### 4. Publish a Skill

```bash
agentskills push ./my-first-skill

# Validating SKILL.md...        ✓
# Packing bundle...             ✓ (12.3 KB)
# Uploading my-first-skill@0.1.0...   ✓
# Checksum: sha256:a1b2c3d4...
#
# Published my-first-skill@0.1.0 successfully.
```

### 5. Download a Skill

```bash
# Download latest version
agentskills pull my-first-skill

# Download specific version
agentskills pull my-first-skill@0.1.0
```

### 6. Search for Skills

```bash
agentskills search code-review

# NAME                  VERSION  DOWNLOADS  DESCRIPTION
# code-review-agent     1.2.0    42         PR code review skill
# code-review-lite      0.3.0    7          Lightweight review helper
```

---

## CLI Reference

| Command | Description | Example |
|---------|-------------|---------|
| `agentskills init [name]` | Scaffold a Skill | `agentskills init my-skill` |
| `agentskills push [path]` | Pack and upload a Skill | `agentskills push ./my-skill` |
| `agentskills pull <name>[@ver]` | Download a Skill | `agentskills pull my-skill@1.0.0` |
| `agentskills search <keyword>` | Search for Skills | `agentskills search code-review` |
| `agentskills login` | Set API token | `agentskills login` |
| `agentskills version` | Show version | `agentskills version` |

---

## Server Reference

| Command | Description | Example |
|---------|-------------|---------|
| `agentskills-server serve` | Start HTTP server | `agentskills-server serve --port 8000` |
| `agentskills-server migrate` | Run database migrations | `agentskills-server migrate` |
| `agentskills-server version` | Show version | `agentskills-server version` |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTSKILLS_PORT` | `8000` | HTTP listen port |
| `AGENTSKILLS_DB_DRIVER` | `sqlite` | Database type: `sqlite` or `postgres` |
| `AGENTSKILLS_DB_DSN` | `./data/agentskills.db` | Database connection string |
| `AGENTSKILLS_STORAGE_DRIVER` | `local` | Storage type: `local` or `s3` |
| `AGENTSKILLS_STORAGE_PATH` | `./data/bundles` | Local storage path |
| `AGENTSKILLS_S3_ENDPOINT` | - | S3/MinIO endpoint |
| `AGENTSKILLS_S3_ACCESS_KEY` | - | S3 access key |
| `AGENTSKILLS_S3_SECRET_KEY` | - | S3 secret key |
| `AGENTSKILLS_S3_BUCKET` | `skills` | S3 bucket name |
| `AGENTSKILLS_MAX_BUNDLE_SIZE` | `52428800` | Max bundle size in bytes (default 50 MB) |

---

## API Endpoints

Base URL: `http://localhost:8000/v1`

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/v1/health` | Health check | No |
| `POST` | `/v1/skills/publish` | Upload a Skill bundle | Bearer Token |
| `GET` | `/v1/skills/{name}` | Get Skill info + latest version | No |
| `GET` | `/v1/skills/{name}/versions` | List all versions | No |
| `GET` | `/v1/skills/{name}/versions/{ver}/download` | Download a specific version | No |
| `GET` | `/v1/skills?q=keyword&tag=tag` | Search Skills | No |

Auth: `Authorization: Bearer <your-token>`

---

## Skill Bundle Format

Each Skill is a directory with a required `SKILL.md` file:

```
my-skill/
├── SKILL.md         (required) YAML frontmatter + Markdown instructions
├── scripts/         (optional) Scripts the agent can execute
├── references/      (optional) Reference documents
└── assets/          (optional) Static resources
```

### SKILL.md Format

```yaml
---
name: "my-skill"                    # required, globally unique, [a-z0-9-], 3-64 chars
version: "1.0.0"                    # required, strict semver
description: "My awesome skill"     # required, max 256 chars
author: "username"                  # required, must match API token owner
tags:                               # optional, max 10
  - tag1
  - tag2
license: "MIT"                      # optional, SPDX identifier
---

# My Skill

Agent instructions go here...
```

---

## Deployment Scenarios

### Scenario A: Individual Developer (Simplest)

```bash
# Download binary → start → done
./agentskills-server serve
# Data stored in ./data/, SQLite + local files, zero-config
```

### Scenario B: Team / Small Org

```bash
# Use the "Simple mode" docker-compose.yml above
docker compose up -d
# 25 MB image, SQLite + local storage, auto-initialized
```

### Scenario C: Production

```bash
# Use the "Production mode" docker-compose.prod.yml above
docker compose -f docker-compose.prod.yml up -d
```

### Scenario D: Windows

```powershell
# Download .exe → double-click or run from command line
agentskills-server-windows-amd64.exe serve
# No installation required
```

---

## Architecture

```
                                  ┌─── SQLite (embedded, default)
                                  │
agentskills-server ──── Database ─┤
     (Go binary)        Interface │
                                  └─── PostgreSQL (production)

                                  ┌─── Local FS (default)
                                  │
                         Storage ─┤
                        Interface │
                                  └─── S3/MinIO (production)
```

- **Language**: Go 1.22+
- **HTTP Router**: go-chi/chi
- **CLI Framework**: spf13/cobra
- **Embedded DB**: modernc.org/sqlite (pure Go, no CGO)
- **Build Strategy**: Go build tags separate CLI / Server binaries

---

## Development

```bash
# Requirements: Go 1.22+, (optional) Docker & Docker Compose

# Build
make build-cli        # CLI binary
make build-server     # Server binary
make build-all        # All platforms

# Test
make test             # Run all tests

# Dev server
./bin/agentskills-server serve

# Docker dev
docker compose up -d
```

See [`reference/SDD.md`](reference/SDD.md) for the full design specification.

---

## CI/CD

The project uses GitHub Actions for automated builds and releases.

### Trigger Rules

| Event | Workflow | Action |
|-------|----------|--------|
| Push / PR to `main` | `ci.yml` | Run tests + verify build + verify Docker build |
| Push tag `v*` | `release.yml` | Test → cross-compile → Docker push → GitHub Release |

### Release Process

```bash
# 1. Tag a release
git tag v1.0.0
git push origin v1.0.0

# 2. GitHub Actions automatically:
#    - Runs tests
#    - Compiles 10 binaries (5 platforms × CLI/Server)
#    - Builds Docker images (linux/amd64 + linux/arm64)
#    - Pushes images to Docker Hub + GitHub Container Registry
#    - Creates GitHub Release + uploads binaries + SHA256 checksums
```

### Required GitHub Secrets

Set these in GitHub repo → Settings → Secrets and variables → Actions:

| Secret | Description | How to get |
|--------|-------------|------------|
| `DOCKERHUB_USERNAME` | Docker Hub username | Register at [hub.docker.com](https://hub.docker.com) |
| `DOCKERHUB_TOKEN` | Docker Hub Access Token | Docker Hub → Account Settings → Security → New Access Token |
| `GITHUB_TOKEN` | GitHub Token (auto-provided) | No manual setup needed |

### Docker Image Tag Convention

Pushing tag `v1.2.3` automatically generates:

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
