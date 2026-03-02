# AgentSkills Registry

**English** | [中文](./README.zh-TW.md)

AgentSkills is a centralized registry platform for AI Agent Skills, similar to npm or Docker Hub. Developers can publish (push) and download (pull) standardized **Skill Bundles** via the CLI tool. The platform handles version control, metadata parsing, and file storage.

## Why AgentSkills?

AI agents (such as Claude Code, AutoGPT, LangChain agents) are becoming increasingly powerful, but they lack a standardized way to **share and reuse capabilities**. AgentSkills solves this by providing:

- **Standardized Skill Format** — Each skill is a directory containing a `SKILL.md` (YAML frontmatter + Markdown instructions), optional `scripts/` for callable tools, `references/` for RAG/few-shot examples, and `assets/` for templates.
- **One-command sharing** — `agentskills push` packages and uploads a skill; `agentskills pull` downloads and extracts it. Just like `npm publish` / `npm install`.
- **Agent-agnostic** — Skill Bundles are plain files. Any agent framework can consume them by reading `SKILL.md` for instructions and loading the accompanying scripts/references.
- **Version control** — Every skill is versioned with strict semver. Teams can pin specific versions and upgrade on their own schedule.
- **Discoverability** — `agentskills search` lets developers find community-contributed skills by keyword or tag (e.g., `code-review`, `data-analysis`, `devops`).

### Skill Bundle Structure

```
my-skill/
├── SKILL.md         # Required: YAML frontmatter (name, version, description, author, tags) + Markdown instructions
├── scripts/         # Optional: scripts the agent can execute
├── references/      # Optional: RAG / few-shot reference documents
└── assets/          # Optional: static templates and resources
```

## How It Works

```
AI Agent (Claude Code, AutoGPT, etc.)
        │
        │  Executes CLI commands in shell
        ▼
   agentskills CLI ──── HTTP REST API ────► agentskills server
   (Go binary)                              (Go binary, built with -tags server)
```

The CLI is the sole interface between agents and the registry server. Agents interact with the platform by executing shell commands — no need to know the underlying HTTP API:

- `agentskills search code-review` → finds matching skills on the server
- `agentskills pull code-review` → downloads and extracts the skill bundle locally
- `agentskills push ./my-skill` → packages and uploads a skill to the server

## Architecture

- **Backend API + CLI Tool**: Go + Cobra in `cli/` (server built with `-tags server`)
- **Storage**: File-based (bundles stored as `.tar.gz` with JSON metadata)
- **Spec**: See [`reference/SDD.md`](./reference/SDD.md) for the complete design document

## Quick Start

### Run the Server (Docker)

```bash
docker build -t agentskills-server .
docker run -p 8000:8000 -v agentskills-data:/data agentskills-server
```

### Build the CLI

```bash
cd cli
go build -o bin/agentskills .
```

### Verify

```bash
# CLI usage
./cli/bin/agentskills --help

# Search for skills
./cli/bin/agentskills search test
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
├── cli/                  # Go CLI + Server
│   ├── cmd/              # Cobra commands (incl. serve)
│   ├── server/           # HTTP handlers & file store
│   ├── internal/         # Internal packages
│   ├── Dockerfile
│   ├── Makefile
│   └── go.mod
├── reference/            # Design documents
│   └── SDD.md
├── Dockerfile            # Server Docker image
└── docker-compose.yml
```

## License

See [LICENSE](./LICENSE).
