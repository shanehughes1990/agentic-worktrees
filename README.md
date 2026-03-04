# 🚀 Agentic Repositories

Welcome! 🎉 `agentic-repositories` is a Go-powered orchestration system that turns scoped docs into execution-ready taskboards, then runs tasks through isolated Git repositories with queue-backed workflows.

Think of it as: **ingest docs ➜ generate board ➜ execute safely in parallel ➜ track workflow status** ✨

> The Go application codebase is isolated under `mvp/`.

Road to V1: see [`docs/VERSION_1_ROADMAP.md`](docs/VERSION_1_ROADMAP.md).

---

## ✅ Project Status

**MVP:** This codebase is now considered a **Minimum Viable Product (MVP)** with a working proof of concept, and is ready to proceed to the next phase.

---

## 🌟 What This Project Does

- 📥 Ingests a **file or folder** as source input
- 🧠 Builds/normalizes documentation for decomposition
- 🗂️ Produces taskboards and stores board/workflow state on disk
- 🌲 Executes task flows in isolated Git repositories
- 🧵 Uses Redis + Asynq for resilient queue/workflow processing
- 🖥️ Provides an interactive terminal dashboard for operations

---

## 🧱 Runtime Dependencies

You’ll need these installed/running:

| Dependency              | Why it’s needed                                 | Required               |
| ----------------------- | ----------------------------------------------- | ---------------------- |
| Go 1.25.3+              | Build and run the app                           | Yes                    |
| Redis 7+                | Queue backend for Asynq workflows               | Yes                    |
| Git                     | Repository checkout and merge flow execution    | Yes                    |
| GitHub Copilot CLI/auth | Decomposition + conflict-resolution agent calls | Yes for agent features |
| Docker (optional)       | Easy local Redis/Asynqmon via docker-compose    | No                     |
| Task CLI (optional)     | Convenience runner for `task` command           | No                     |

### Optional local infra with Docker 🐳

```bash
docker compose up -d
```

This starts:

- Redis on `localhost:6379`
- Asynqmon on `http://localhost:8085`

---

## ⚡ Quick Start

1. Create/update your `.env` (see full table below).
2. Start Redis (locally or with Docker).
3. Run the app:

```bash
cd v1
task
```

or

```bash
go run ./v1/cmd/main.go
```

---

## 🧭 Dashboard Flow

- **Settings**: set global Redis override
- **Ingestion**:
  - choose source type (`folder` or `file`)
  - if `folder`, configure walk depth + ignore paths/extensions
  - execute and auto-redirect to workflow status
- **Repositories**:
  - select board
  - choose source branch
  - run/monitor task execution pipeline

---

## 🗃️ Runtime Filesystem Layout

All runtime artifacts are rooted at `APP_ROOT_DIR` (default `.agentic-orchestrator`):

- `logs/` → application logs
- `taskboards/` → board JSON files
- `workflows/` → workflow/run/job JSON files
- `repositories/` → git repository directories used during execution

---

## 🔧 Environment Variables

The table below documents all currently supported public env vars.

| Variable                      | Description                                                                                                       | Default                 | Required |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------- | ----------------------- | -------- |
| `APP_ROOT_DIR`                | Runtime root directory for logs, taskboards, workflows, and repositories. Must be repo-relative and not `.`/`..`. | `.agentic-orchestrator` | No       |
| `LOG_FORMAT`                  | Log output format (`text` or `json`).                                                                             | `text`                  | No       |
| `LOG_LEVEL`                   | Log level (`debug`, `info`, `warn`, `error`, `fatal`, `panic`).                                                   | `info`                  | No       |
| `REDIS_URI`                   | Redis connection URI used by Asynq client/server.                                                                 | _(none)_                | Yes      |
| `COPILOT_MODEL`               | Preferred model for Copilot decomposition requests.                                                               | _(none)_                | No       |
| `GITHUB_TOKEN`                | GitHub token passed to Copilot client when provided.                                                              | _(none)_                | No       |
| `COPILOT_CLI_PATH`            | Override path to Copilot CLI executable.                                                                          | _(none)_                | No       |
| `COPILOT_CLI_URL`             | Optional Copilot CLI endpoint override.                                                                           | _(none)_                | No       |
| `COPILOT_AUTH_STATUS_COMMAND` | Command used to check Copilot auth status.                                                                        | `copilot auth status`   | No       |
| `COPILOT_AUTH_LOGIN_COMMAND`  | Command used to trigger Copilot login flow.                                                                       | `copilot auth login`    | No       |
| `COPILOT_SKILL_DIRECTORIES`   | Comma-separated skill directories for Copilot context.                                                            | _(none)_                | No       |

### Example `.env` 🧪

```env
APP_ROOT_DIR=.agentic-orchestrator
LOG_FORMAT=text
LOG_LEVEL=info
REDIS_URI=redis://localhost:6379/0

# Optional Copilot settings
# COPILOT_MODEL=
# GITHUB_TOKEN=
# COPILOT_CLI_PATH=
# COPILOT_CLI_URL=
# COPILOT_AUTH_STATUS_COMMAND=copilot auth status
# COPILOT_AUTH_LOGIN_COMMAND=copilot auth login
# COPILOT_SKILL_DIRECTORIES=
```

---

## 🏗️ Architecture (DDD)

Project layers follow strict DDD boundaries:

- `v1/internal/interface` → terminal/dashboard + worker handlers
- `v1/internal/application` → orchestration/use-cases
- `v1/internal/domain` → business rules and invariants
- `v1/internal/infrastructure` → adapters (redis, git, persistence, logging)

Dependency direction is inward: `interface -> application -> domain`, with infrastructure implementing required ports.

---

## 🧰 Tech Stack Highlights

- Go 1.25
- Asynq + Redis
- tview/tcell terminal UI
- Logrus + lumberjack log rotation
- Gonum for dependency graph traversal
- GitHub Copilot SDK integration

---

## 🙌 Contributing

Issues and PRs are welcome! Please keep changes scoped, tested, and aligned with the repository’s DDD boundaries and runtime safety conventions.

Happy building! 💙
