# AgentSkills Registry

AI Agent Skill Registry — CLI + API platform for publishing and pulling standardized skill bundles. Similar to npm or Docker Hub, but for AI Agent Skills.

## Architecture

```
┌──────────────────────────────────────────┐
│            docker-compose                │
│                                          │
│  ┌─────────────┐    ┌─────────────────┐  │
│  │ PostgreSQL   │    │ MinIO (S3)      │  │
│  │ port: 5432   │    │ API:  9000      │  │
│  │              │    │ Console: 9001   │  │
│  └──────┬───────┘    └──────┬──────────┘  │
│         └─────────┬─────────┘             │
│              ┌────┴────┐                  │
│              │ FastAPI  │                  │
│              │ port:8000│                  │
│              └────┬─────┘                  │
└───────────────────┼────────────────────────┘
                    │
             ┌──────┴──────┐
             │   Go CLI    │
             └─────────────┘
```

- **Backend**: FastAPI (Python 3.12+) — `api/`
- **CLI**: Go 1.22+ with Cobra — `cli/`
- **Database**: PostgreSQL 16 (via Docker)
- **Object Storage**: MinIO S3-compatible (via Docker)

## Quick Start

### 1. Start Infrastructure

```bash
docker compose up -d
```

This starts PostgreSQL (port 5432) and MinIO (ports 9000/9001) with a seeded `dev` user.

### 2. Start the API Server

```bash
cd api
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

### 3. Build the CLI

```bash
cd cli
go build -o agentskills .
```

Or run directly:

```bash
cd cli && go run main.go <command>
```

### 4. Login

```bash
agentskills login
# Enter API token: dev-token-12345
```

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
# Auto-detect provider from current directory
agentskills init my-new-skill

# Explicitly target Claude
agentskills init my-new-skill --provider claude
```

**Output:**

```
Created my-new-skill/
  ├── SKILL.md        (template for claude)
  ├── scripts/
  ├── references/
  └── assets/
```

The generated `SKILL.md` includes provider-specific fields:

```yaml
---
name: "my-new-skill"
version: "0.1.0"
description: "Brief description of what this skill does and when Claude should use it."
author: ""
tags: []
compatibility: "Designed for Claude Code"
---

# my-new-skill

## When to use this skill
Use this skill when...

## Instructions
1. Step one...
2. Step two...

## Examples
[Concrete examples of using this skill]
```

---

### `agentskills push [path]`

Pack and upload a skill bundle to the registry.

```bash
# Push from current directory
agentskills push

# Push a specific directory
agentskills push ./my-skill

# Push with explicit provider
agentskills push ./my-skill --provider gemini
```

**Output:**

```
Validating SKILL.md...        OK
Provider: claude (auto-detected)
Packing bundle...             OK (12.3 KB)
Uploading my-skill@1.0.0...   OK
Checksum: sha256:a1b2c3d4...

Published my-skill@1.0.0 successfully.
  Providers: claude
```

**What it does:**

1. Validates `SKILL.md` frontmatter locally (name, version, description, author)
2. Applies provider-specific name rules (e.g., Claude skills can't contain "anthropic" or "claude")
3. Packs the directory as `.tar.gz` (excludes `.git/`, `node_modules/`, `__pycache__/`, `*.pyc`, `.env`, `.DS_Store`)
4. Uploads to the registry API with provider metadata
5. Verifies server checksum matches local checksum

---

### `agentskills pull <name>[@version]`

Download and extract a skill bundle into the correct provider discovery path.

```bash
# Pull latest version (auto-detect provider)
agentskills pull code-review-agent

# Pull a specific version
agentskills pull code-review-agent@1.0.0

# Pull for a specific provider
agentskills pull code-review-agent --provider claude

# Install to user-level path instead of project-level
agentskills pull code-review-agent --provider claude --scope user
```

**Output:**

```
Downloading code-review-agent@1.2.0...  OK
Verifying checksum...          OK
Provider: claude
Extracted to ./.claude/skills/code-review-agent/
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--scope` | `workspace` | `workspace` = project-level, `user` = user-level (`~/`) |
| `--provider` | auto-detect | Target agent provider |

**Install paths by provider and scope:**

| Provider | `--scope workspace` | `--scope user` |
|----------|---------------------|----------------|
| claude | `./.claude/skills/{name}/` | `~/.claude/skills/{name}/` |
| gemini | `./.agents/skills/{name}/` | `~/.agents/skills/{name}/` |
| codex | `./.agents/skills/{name}/` | `~/.codex/skills/{name}/` |
| copilot | `./.github/skills/{name}/` | `./{name}/` |
| cursor | `./.cursor/skills/{name}/` | `~/.cursor/skills/{name}/` |
| windsurf | `./.windsurf/skills/{name}/` | `~/.codeium/skills/{name}/` |
| antigravity | `./.agent/skills/{name}/` | `~/.antigravity/skills/{name}/` |
| generic | `./{name}/` | `./{name}/` |

---

### `agentskills search <keyword>`

Search the registry for skills.

```bash
# Search by keyword
agentskills search code-review

# Filter by provider
agentskills search code-review --provider claude

# Filter by tag
agentskills search code-review --tag github
```

**Output:**

```
NAME                  VERSION  DOWNLOADS  PROVIDERS      DESCRIPTION
code-review-agent     1.2.0    42         claude,gemini  PR code review skill
code-review-lite      0.3.0    7          generic        Lightweight review helper
```

---

### `agentskills login`

Save your API token to local config (`~/.agentskills/config.yaml`).

```bash
agentskills login
# Enter API token: ********
# Token saved to ~/.agentskills/config.yaml
```

**Config file format** (`~/.agentskills/config.yaml`):

```yaml
api_url: "http://localhost:8000"
token: "dev-token-12345"
default_provider: "claude"  # optional
```

---

## SKILL.md Format

Every skill bundle requires a `SKILL.md` file with YAML frontmatter:

```yaml
---
name: "code-review-agent"           # Required: [a-z0-9\-], 3-64 chars, no --
version: "1.0.0"                    # Required: strict semver (MAJOR.MINOR.PATCH)
description: "PR code review skill" # Required: 1-256 chars
author: "username"                  # Required: must match API token user
tags:                               # Optional: max 10, each [a-z0-9\-], 1-32 chars
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
| `POST` | `/v1/skills/publish` | Bearer token | Upload a skill bundle (.tar.gz) |
| `GET` | `/v1/skills/{name}` | No | Get skill info + latest version |
| `GET` | `/v1/skills/{name}/versions` | No | List all versions |
| `GET` | `/v1/skills/{name}/versions/{version}/download` | No | Download a specific version |
| `GET` | `/v1/skills?q=&tag=&provider=&page=&per_page=` | No | Search skills |
| `GET` | `/v1/health` | No | Health check |

---

## Development

### Run API Tests

```bash
cd api && pytest -xvs
# 77 tests (parser, publish, pull, search)
```

### Run CLI Tests

```bash
cd cli && go test ./... -v
# 21 tests (provider detection, name validation, install paths)
```

### Infrastructure Commands

```bash
# Start services
docker compose up -d

# Verify PostgreSQL
docker compose exec postgres psql -U dev -d agentskills -c "SELECT COUNT(*) FROM users;"

# Verify MinIO
curl -s http://localhost:9000/minio/health/live

# Tear down with volumes
docker compose down -v
```

---

## End-to-End Example

```bash
# 1. Start infrastructure + API
docker compose up -d
cd api && uvicorn app.main:app --reload --port 8000 &

# 2. Login
cd cli && go run main.go login
# Enter: dev-token-12345

# 3. Create a new skill for Claude
go run main.go init my-skill --provider claude

# 4. Edit my-skill/SKILL.md (fill in description, author, etc.)

# 5. Publish
go run main.go push ./my-skill --provider claude

# 6. Search for it
go run main.go search my-skill

# 7. Pull it into another project
cd /tmp/other-project
mkdir -p .claude
go run /path/to/cli/main.go pull my-skill
# → Extracted to ./.claude/skills/my-skill/
```

## License

MIT
