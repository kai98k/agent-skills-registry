# AgentSkills Registry

**English** | [中文](./README.zh-TW.md)

AgentSkills is a centralized registry platform for AI Agent Skills, similar to npm or Docker Hub. Developers can publish (push) and download (pull) standardized Skill Bundles via the CLI tool. The platform handles version control, metadata parsing, and file storage.

## Architecture

- **Backend API**: FastAPI (Python 3.12+) in `api/`
- **CLI Tool**: Go + Cobra in `cli/`
- **Database**: PostgreSQL 16 (Docker)
- **Object Storage**: MinIO S3-compatible (Docker)
- **Spec**: See [`reference/SDD.md`](./reference/SDD.md) for the complete design document

## Quick Start

### Start Infrastructure

```bash
docker compose up -d
```

This starts PostgreSQL (port 5432) and MinIO (API: 9000, Console: 9001).

### Start the API (Development)

```bash
cd api
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

### Build the CLI

```bash
cd cli
go build -o bin/agentskills .
```

### Verify

```bash
# Health check
curl http://localhost:8000/v1/health

# CLI usage
./cli/bin/agentskills --help
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `agentskills init <name>` | Create a new skill skeleton directory |
| `agentskills login` | Save API token to local config |
| `agentskills push <path>` | Pack and upload a skill bundle |
| `agentskills pull <name>[@version]` | Download and extract a skill bundle |
| `agentskills search <keyword>` | Search for skills on the registry |

## Project Structure

```
.
├── api/                  # FastAPI backend
│   ├── app/
│   │   ├── routes/       # API endpoints
│   │   └── services/     # Business logic
│   ├── tests/
│   ├── Dockerfile
│   └── requirements.txt
├── cli/                  # Go CLI tool
│   ├── cmd/              # Cobra commands
│   ├── internal/         # Internal packages
│   ├── Dockerfile
│   ├── Makefile
│   └── go.mod
├── reference/            # Design documents
│   └── SDD.md
├── docker-compose.yml
└── init.sql
```

## License

See [LICENSE](./LICENSE).
