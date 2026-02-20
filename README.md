# AgentSkills Registry

AI Agent Skill Registry — Web UI + CLI + API platform for publishing, discovering, and pulling standardized skill bundles. Similar to npm or Docker Hub, but for AI Agent Skills.

## Architecture

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
└─────────────────┼────────────────────────────────┘
                  │
       ┌──────────┼──────────┐
       │          │          │
┌──────┴──┐  ┌───┴──────┐  ┌┴──────────┐
│ Browser  │  │  Go CLI  │  │ External  │
└─────────┘  └──────────┘  └───────────┘
```

- **Frontend**: Next.js 15+ (App Router, SSR, shadcn/ui) — `web/`
- **Backend**: FastAPI (Python 3.12+) — `api/`
- **CLI**: Go 1.22+ with Cobra — `cli/`
- **Database**: PostgreSQL 16 with full-text search (`tsvector`)
- **Object Storage**: MinIO S3-compatible (via Docker)
- **Auth**: GitHub OAuth (Web) + Bearer Token (CLI/API)

## Quick Start

### 1. Start Infrastructure

```bash
docker compose up -d
```

This starts PostgreSQL (port 5432) and MinIO (ports 9000/9001) with a seeded `dev` user and 10 default categories.

### 2. Start the API Server

```bash
cd api
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

### 3. Start the Web Frontend

```bash
cd web
npm install
npm run dev
# → http://localhost:3000
```

### 4. Build the CLI

```bash
cd cli
go build -o agentskills .
```

Or run directly:

```bash
cd cli && go run main.go <command>
```

### 5. Login (CLI)

```bash
agentskills login
# Enter API token: dev-token-12345
```

---

## Web UI

The web frontend provides a browsing and discovery experience:

- **Homepage**: Hero search bar, category grid, trending & latest skills
- **Search**: Full-text search with category/tag filters and sort options
- **Skill Detail**: Rendered SKILL.md, version history, install command, star button
- **User Profiles**: Published skills, download/star counts
- **GitHub OAuth**: Sign in with GitHub, API token management

---

## CLI Usage

### Provider Auto-Detection

The CLI automatically detects which AI agent provider you're using by scanning your project directory for configuration files:

| Provider | Detected From | Install Path |
|----------|---------------|-------------|
| `claude` | `.claude/`, `CLAUDE.md` | `.claude/skills/{name}/` |
| `gemini` | `.gemini/`, `GEMINI.md` | `.agents/skills/{name}/` |
| `codex` | `.codex/`, `AGENTS.md` | `.agents/skills/{name}/` |
| `copilot` | `.github/copilot-instructions.md`, `.github/skills/` | `.github/skills/{name}/` |
| `cursor` | `.cursor/`, `.cursorrules` | `.cursor/skills/{name}/` |
| `windsurf` | `.windsurf/`, `.windsurfrules` | `.windsurf/skills/{name}/` |
| `antigravity` | `.antigravity/` | `.agent/skills/{name}/` |
| `generic` | (fallback) | `./{name}/` |

Override auto-detection with the global `--provider` flag:

```bash
agentskills --provider claude <command>
```

---

### `agentskills init <name>`

Create a new skill skeleton directory with a provider-appropriate SKILL.md template.

```bash
agentskills init my-new-skill
agentskills init my-new-skill --provider claude
```

### `agentskills push [path]`

Pack and upload a skill bundle to the registry.

```bash
agentskills push ./my-skill
```

### `agentskills pull <name>[@version]`

Download and extract a skill bundle into the correct provider discovery path.

```bash
agentskills pull code-review-agent
agentskills pull code-review-agent@1.0.0
agentskills pull code-review-agent --provider claude --scope user
```

### `agentskills search <keyword>`

Search the registry for skills.

```bash
agentskills search code-review
agentskills search code-review --provider claude --tag github
```

### `agentskills login`

Save your API token to local config (`~/.agentskills/config.yaml`).

```bash
agentskills login
```

---

## SKILL.md Format

Every skill bundle requires a `SKILL.md` file with YAML frontmatter:

```yaml
---
name: "code-review-agent"           # Required: [a-z0-9\-], 3-64 chars, no --
version: "1.0.0"                    # Required: strict semver
description: "PR code review skill" # Required: 1-256 chars
author: "username"                  # Required: must match API token user
tags:                               # Optional: max 10, each [a-z0-9\-]
  - code-review
  - github
license: "MIT"                      # Optional: SPDX identifier
compatibility: "Designed for Claude Code" # Optional: provider hint
---

# Code Review Agent

Your skill instructions in Markdown...
```

---

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/v1/skills/publish` | Bearer Token | Upload a skill bundle (.tar.gz) |
| `GET` | `/v1/skills/{name}` | Optional | Skill info + latest version |
| `GET` | `/v1/skills/{name}/versions` | No | List all versions |
| `GET` | `/v1/skills/{name}/versions/{version}/download` | No | Download a specific version |
| `GET` | `/v1/skills?q=&tag=&category=&sort=&page=&per_page=` | No | Search skills |
| `POST` | `/v1/skills/{name}/star` | Bearer Token | Star a skill |
| `DELETE` | `/v1/skills/{name}/star` | Bearer Token | Unstar a skill |
| `GET` | `/v1/categories` | No | List categories + skill counts |
| `POST` | `/v1/auth/github` | No | GitHub OAuth login/register |
| `GET` | `/v1/users/{username}` | No | Public user profile |
| `GET` | `/v1/health` | No | Health check |

---

## Development

### Run API Tests

```bash
cd api && pytest -xvs
```

### Run CLI Tests

```bash
cd cli && go test ./... -v
```

### Run Frontend Dev Server

```bash
cd web && npm run dev
```

### Infrastructure Commands

```bash
docker compose up -d           # Start services
docker compose down -v         # Tear down with volumes
```

---

## License

MIT
