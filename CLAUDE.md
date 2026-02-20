# AgentSkills

AI Agent Skill Registry — Web UI + CLI + API platform for publishing, discovering, and pulling standardized skill bundles. Similar to npm or Docker Hub, but for AI Agent Skills.

## Project Status

The comprehensive Software Design Document (`reference/SDD.md`) defines the complete architecture. Backend API and CLI have initial implementations. Web frontend (Next.js) is to be built.

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

- **Frontend**: Next.js 15+ (App Router, SSR, shadcn/ui) in `web/`
- **Backend**: FastAPI (Python 3.12+) in `api/`
- **CLI**: Go 1.22+ with Cobra in `cli/`
- **Database**: PostgreSQL 16 with full-text search (`tsvector`)
- **Object Storage**: MinIO S3-compatible (via Docker)
- **Auth**: GitHub OAuth (Web) + Bearer Token (CLI/API)
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
├── api/                               # FastAPI backend
│   ├── app/
│   │   ├── main.py                    # FastAPI app entry, lifespan, CORS middleware
│   │   ├── config.py                  # pydantic-settings, env var management
│   │   ├── dependencies.py            # DI: get_db, get_current_user, get_s3, get_optional_user
│   │   ├── models.py                  # SQLAlchemy ORM models
│   │   ├── schemas.py                 # Pydantic request/response schemas
│   │   ├── routes/
│   │   │   ├── skills.py              # /v1/skills CRUD + search + star
│   │   │   ├── auth.py               # /v1/auth/github OAuth
│   │   │   ├── categories.py         # /v1/categories
│   │   │   ├── users.py              # /v1/users/{username}
│   │   │   └── health.py             # /v1/health
│   │   └── services/
│   │       ├── storage.py             # MinIO/S3 upload, download, presigned URL
│   │       ├── parser.py              # .tar.gz extraction, SKILL.md parsing
│   │       ├── auth.py                # API token + GitHub OAuth logic
│   │       └── markdown.py            # SKILL.md → safe HTML rendering
│   ├── tests/
│   │   ├── conftest.py                # pytest fixtures
│   │   ├── test_publish.py
│   │   ├── test_pull.py
│   │   ├── test_search.py
│   │   ├── test_parser.py
│   │   ├── test_stars.py
│   │   └── test_auth.py
│   ├── requirements.txt
│   └── Dockerfile
├── cli/                               # Go CLI tool
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
├── web/                               # Next.js frontend
│   ├── package.json
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── .env.local.example
│   ├── src/
│   │   ├── app/                       # App Router pages
│   │   │   ├── layout.tsx             # Root layout, providers, Header/Footer
│   │   │   ├── page.tsx               # Homepage: hero + categories + featured
│   │   │   ├── skills/[name]/page.tsx # Skill detail + markdown + sidebar
│   │   │   ├── search/page.tsx        # Search results + filters
│   │   │   ├── categories/[category]/page.tsx
│   │   │   ├── user/[username]/page.tsx
│   │   │   └── api/auth/[...nextauth]/route.ts
│   │   ├── components/                # React components
│   │   │   ├── ui/                    # shadcn/ui base components
│   │   │   ├── layout/               # Header, Footer
│   │   │   ├── skills/               # SkillCard, SkillDetail, StarButton
│   │   │   ├── search/               # SearchBar, SearchFilters
│   │   │   ├── home/                 # Hero, CategoryGrid, FeaturedSkills
│   │   │   └── markdown/             # MarkdownRenderer
│   │   ├── lib/                       # API client, auth config, utils
│   │   └── types/                     # TypeScript types
│   └── Dockerfile
├── docker-compose.yml                 # PostgreSQL + MinIO
└── init.sql                           # DB schema initialization
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
- CORS enabled for `http://localhost:3000` (Next.js dev server)

## Development — CLI (Go)

```bash
cd cli && go run main.go <command>
cd cli && go test ./...
cd cli && make build
```

- Go 1.22+
- Cobra framework for CLI commands
- Viper for configuration management

## Development — Web Frontend (Next.js)

```bash
cd web && npm install
cd web && npm run dev          # http://localhost:3000
cd web && npm run build        # Production build
cd web && npm run test         # Vitest unit tests
cd web && npx playwright test  # E2E tests
```

- Next.js 15+ with App Router and Server Components
- React 19+, TypeScript 5+
- Tailwind CSS 4+ with shadcn/ui components
- NextAuth.js 5+ for GitHub OAuth
- Dark/Light theme support

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
10. **Auth**: Bearer token for CLI/API, GitHub OAuth for Web UI — both share the `users` table
11. **Markdown safety** — always sanitize rendered HTML with `bleach` (backend) to prevent XSS
12. **Next.js SSR** — use Server Components for data fetching, ISR for caching (see SDD.md §12.4)

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/v1/skills/publish` | Bearer Token | Upload a skill bundle (.tar.gz) |
| `GET` | `/v1/skills/{name}` | Optional | Get skill info + latest version (starred_by_me with auth) |
| `GET` | `/v1/skills/{name}/versions` | No | List all versions of a skill |
| `GET` | `/v1/skills/{name}/versions/{version}/download` | No | Download a specific version bundle |
| `GET` | `/v1/skills?q=&tag=&category=&sort=&page=&per_page=` | No | Search skills (full-text + filters) |
| `POST` | `/v1/skills/{name}/star` | Bearer Token | Star a skill |
| `DELETE` | `/v1/skills/{name}/star` | Bearer Token | Unstar a skill |
| `GET` | `/v1/categories` | No | List categories with skill counts |
| `POST` | `/v1/auth/github` | No | GitHub OAuth login/register |
| `GET` | `/v1/users/{username}` | No | Public user profile + published skills |
| `GET` | `/v1/health` | No | Health check (DB + storage) |

## Database Schema (PostgreSQL)

Five tables: `users`, `categories`, `skills`, `skill_versions`, `stars`. Key design decisions:
- Two-table separation: `skills` for identity/aggregates, `skill_versions` for immutable publish records
- Latest version determined by `published_at DESC LIMIT 1` (no explicit `latest` column)
- JSONB `metadata` column stores full frontmatter for forward compatibility
- PostgreSQL full-text search with `tsvector` + GIN index on skills
- `stars` table with composite PK `(user_id, skill_id)`, redundant `stars_count` on skills for fast sorting
- `categories` table with 10 preset categories
- GitHub OAuth fields on `users`: `github_id`, `display_name`, `avatar_url`, `bio`
- Seed data includes a `dev` user with token `dev-token-12345` and 10 default categories

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
min_agent_version: ">=0.1" # Optional: reserved, not validated
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
| XSS (Markdown) | `bleach` HTML sanitizer, allow only safe tags |
| CSRF | NextAuth.js built-in CSRF token protection |
| CORS | FastAPI middleware restricted to allowed origins |

## Development Order

Follow SDD.md §11 for implementation sequence:

### Phase 1: Backend API
1. Infrastructure — `docker-compose.yml` + `init.sql`
2. FastAPI skeleton — `main.py`, `config.py`, `health.py`, CORS
3. Parser module — `parser.py` + `markdown.py` + `test_parser.py`
4. Storage module — `storage.py` (MinIO upload/download)
5. Publish endpoint — `POST /v1/skills/publish` + tests
6. Query endpoints — all GET routes + full-text search + tests
7. Stars + Categories + Auth — star/unstar, categories, GitHub OAuth, user profiles

### Phase 2: Go CLI
8. CLI skeleton — `root.go`, `config.go`, `login.go`
9. CLI push/pull — full CLI-to-API workflow
10. CLI search + init — remaining CLI commands

### Phase 3: Web Frontend (Next.js)
11. Next.js skeleton — layout, API client, shadcn/ui setup
12. Homepage + Search — hero, categories, search with filters
13. Skill detail page — markdown rendering, version history, star
14. GitHub OAuth — login/logout, user profile pages
15. Integration verification — end-to-end validation

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

### Frontend (vitest + playwright)
- Unit: API client, utility functions
- Component: SkillCard, SearchBar, MarkdownRenderer
- E2E: Homepage, search flow, skill detail, login flow

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
