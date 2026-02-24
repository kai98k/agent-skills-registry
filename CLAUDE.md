# AgentSkills

AI Agent Skill Registry — CLI + Server for publishing and pulling skill bundles.

## Architecture

- All Go: CLI + Server in single repo, separated by build tags
- CLI binary: `go build .` → `agentskills` (~8MB)
- Server binary: `go build -tags server .` → `agentskills-server` (~18MB)
- DB: SQLite (embedded, default) or PostgreSQL (production)
- Storage: Local filesystem (default) or S3/MinIO (production)
- Auth: Legacy static token + multi-PAT (Personal Access Tokens) with SHA-256 hashing
- Spec: See `reference/SDD.md` for complete design document
- Architecture: See `reference/GO-BACKEND-ARCHITECTURE.md` for detailed Go design

## Quick Start

```bash
# Start server (SQLite + LocalFS, zero config)
make build-server && ./bin/agentskills-server serve

# Or with Docker
docker compose up -d
```

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
- Token format: `ask_<40 hex chars>`, stored as SHA-256 hash
