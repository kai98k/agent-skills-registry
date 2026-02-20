# AgentSkills

AI Agent Skill Registry — CLI + API platform for publishing and pulling standardized skill bundles. Similar to npm or Docker Hub, but for AI Agent Skills.

## Project Status

This project is in the **specification phase**. The comprehensive Software Design Document (`reference/SDD.md`) defines the complete architecture. No implementation code exists yet — all source files are to be built from the SDD blueprint.

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

- **Backend**: FastAPI (Python 3.12+) in `api/`
- **CLI**: Go 1.22+ with Cobra in `cli/`
- **Database**: PostgreSQL 16 (via Docker)
- **Object Storage**: MinIO S3-compatible (via Docker)
- **Spec**: `reference/SDD.md` — the authoritative design document

## Repository Structure

```
agent-skills-registry/
├── CLAUDE.md                          # This file
├── LICENSE                            # MIT License
├── README.md                          # Project overview
├── reference/
│   ├── SDD.md                         # Software Design Document (authoritative spec)
│   ├── what-are-agent-skills.md       # Agent Skills open format specification
│   ├── claude-agent-skills-intro.md   # Claude platform skills documentation
│   └── gemini-agent-skills-intro.md   # Gemini CLI skills documentation
├── api/                               # (to be created) FastAPI backend
│   ├── app/
│   │   ├── main.py                    # FastAPI app entry, lifespan, middleware
│   │   ├── config.py                  # pydantic-settings, env var management
│   │   ├── dependencies.py            # DI: get_db, get_current_user, get_s3
│   │   ├── models.py                  # SQLAlchemy ORM models
│   │   ├── schemas.py                 # Pydantic request/response schemas
│   │   ├── routes/
│   │   │   ├── skills.py              # All /v1/skills endpoints
│   │   │   └── health.py              # /v1/health
│   │   └── services/
│   │       ├── storage.py             # MinIO/S3 upload, download, presigned URL
│   │       ├── parser.py              # .tar.gz extraction, SKILL.md parsing
│   │       └── auth.py                # API token authentication
│   ├── tests/
│   │   ├── conftest.py                # pytest fixtures: test DB, mock S3, test client
│   │   ├── test_publish.py
│   │   ├── test_pull.py
│   │   ├── test_search.py
│   │   └── test_parser.py
│   ├── requirements.txt
│   └── Dockerfile
├── cli/                               # (to be created) Go CLI tool
│   ├── main.go
│   ├── cmd/
│   │   ├── root.go                    # Cobra root command, global flags
│   │   ├── init_cmd.go                # agentskills init
│   │   ├── push.go                    # agentskills push
│   │   ├── pull.go                    # agentskills pull
│   │   ├── search.go                  # agentskills search
│   │   └── login.go                   # agentskills login
│   ├── internal/
│   │   ├── config/config.go           # ~/.agentskills/config.yaml management
│   │   ├── api/client.go              # HTTP client wrapping all API calls
│   │   ├── bundle/pack.go             # tar.gz packing and extraction
│   │   └── parser/frontmatter.go      # Local SKILL.md validation (pre-push check)
│   └── Makefile
├── docker-compose.yml                 # (to be created) PostgreSQL + MinIO
└── init.sql                           # (to be created) DB schema initialization
```

## Infrastructure

```bash
# Start PostgreSQL (5432) + MinIO (9000/9001)
docker compose up -d

# Verify PostgreSQL
docker compose exec postgres psql -U dev -d agentskills -c "SELECT COUNT(*) FROM users;"

# Verify MinIO
curl -s http://localhost:9000/minio/health/live

# Tear down with volumes
docker compose down -v
```

## Development — API (Python)

```bash
cd api && pip install -r requirements.txt
cd api && uvicorn app.main:app --reload --port 8000
cd api && pytest -xvs
```

- Python 3.12+
- Dependencies defined in `api/requirements.txt`
- Async SQLAlchemy with asyncpg — no sync database calls
- boto3 with `endpoint_url` for MinIO — path-style addressing required

## Development — CLI (Go)

```bash
cd cli && go run main.go <command>
cd cli && go test ./...
cd cli && make build
```

- Go 1.22+
- Cobra framework for CLI commands
- Viper for configuration management

## Key Rules

1. **ALWAYS read `reference/SDD.md` before making architectural decisions** — it is the single source of truth for the project design
2. **DB schema is defined in `init.sql`** — do not deviate from the schema in SDD.md §4.1
3. **API responses must match SDD.md §5.2 exactly** — response shapes, status codes, and error formats are strictly specified
4. **Frontmatter validation rules are in SDD.md §2.3** — implement all of them without exception
5. **Tests are mandatory** for every endpoint and parser function — see SDD.md §9 for required test cases
6. **Use async SQLAlchemy (asyncpg)** — no synchronous database calls
7. **Use boto3 with `endpoint_url` for MinIO** — path-style addressing is required (MinIO does not support virtual-hosted-style)
8. **Never execute uploaded scripts server-side** — only parse SKILL.md metadata
9. **Immutable publishing** — same `(skill_id, version)` cannot be overwritten (409 Conflict)
10. **Bearer token auth for MVP** — `Authorization: Bearer <token>`, only POST endpoints require auth

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/v1/skills/publish` | Required | Upload a skill bundle (.tar.gz) |
| `GET` | `/v1/skills/{name}` | No | Get skill info + latest version |
| `GET` | `/v1/skills/{name}/versions` | No | List all versions of a skill |
| `GET` | `/v1/skills/{name}/versions/{version}/download` | No | Download a specific version bundle |
| `GET` | `/v1/skills?q=&tag=&page=&per_page=` | No | Search skills |
| `GET` | `/v1/health` | No | Health check (DB + storage) |

## Database Schema (PostgreSQL)

Three tables: `users`, `skills`, `skill_versions`. Key design decisions:
- Two-table separation: `skills` for identity/aggregates, `skill_versions` for immutable publish records
- Latest version determined by `published_at DESC LIMIT 1` (no explicit `latest` column)
- JSONB `metadata` column stores full frontmatter for forward compatibility
- Seed data includes a `dev` user with token `dev-token-12345`

## SKILL.md Bundle Format

Skills are directories packaged as `.tar.gz` with a required `SKILL.md` file containing YAML frontmatter:

```yaml
---
name: "skill-name"        # Required: [a-z0-9\-], 3-64 chars, no consecutive --
version: "1.0.0"          # Required: strict semver
description: "What it does" # Required: 1-256 chars
author: "username"         # Required: must match API token user
tags:                      # Optional: max 10, each 1-32 chars [a-z0-9\-]
  - tag-name
license: "MIT"             # Optional: SPDX identifier
min_agent_version: ">=0.1" # Optional: reserved, not validated in MVP
---
```

## Security Considerations

| Threat | Protection |
|--------|-----------|
| Zip bomb | 200MB max decompressed size |
| Path traversal | All extraction paths validated within temp directory |
| Arbitrary execution | Server only parses SKILL.md, never executes scripts/ |
| Token leakage | CLI config file chmod 0600; logs never record full tokens |
| SQL injection | SQLAlchemy ORM with parameterized queries |
| Large file DoS | 50MB request body limit at FastAPI layer |

## Development Order

Follow SDD.md §11 for implementation sequence:

1. Infrastructure — `docker-compose.yml` + `init.sql`
2. FastAPI skeleton — `main.py`, `config.py`, `health.py`
3. Parser module — `parser.py` + `test_parser.py`
4. Storage module — `storage.py` (MinIO upload/download)
5. Publish endpoint — `POST /v1/skills/publish` + tests
6. Query endpoints — all GET routes + tests
7. Go CLI skeleton — `root.go`, `config.go`, `login.go`
8. CLI push/pull — full CLI-to-API workflow
9. CLI search + init — remaining CLI commands
10. Integration verification — end-to-end validation

## Testing

### Backend (pytest)
- Test DB: SQLite async (`aiosqlite`) with per-test rollback
- Mock S3: `moto` library to simulate MinIO
- Integration: FastAPI `TestClient` for endpoint testing
- Required cases listed in SDD.md §9.1

### CLI (go test)
- Local frontmatter parsing and validation
- tar.gz bundling with exclusion list
- SHA-256 checksum computation
- API client calls via `httptest` mock server

## Bundle Exclusion List (CLI packing)

When creating `.tar.gz` bundles, the CLI excludes:
```
.git/
.DS_Store
node_modules/
__pycache__/
*.pyc
.env
```
