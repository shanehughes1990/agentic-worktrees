# Agentic Git Worktrees

Task-board generation pipeline using a single CLI entrypoint, Asynq + Redis queueing, and Copilot SDK worker execution.

## What This CLI Does

- `generate-task-board`: loads documentation files and enqueues a task-board generation job
- `run-worker`: runs the Asynq worker that executes Copilot generation and publishes result tasks

## Current Requirements

- Go `1.25.3`
- Redis reachable by the CLI/worker process
- GitHub Copilot access for the account/token used by Copilot SDK
- Copilot CLI available (default discovery) or explicitly configured path/url

## Environment Variables

### Common (optional)

- `LOG_LEVEL` (default: `info`)
- `LOG_FORMAT` (default: `json`)
- `DATABASE_DSN` (default: `sqlite:///./agentic-worktrees.db`)

### Queue/worker runtime

- `REDIS_ADDR` (default: `127.0.0.1:6379`)
- `ASYNQ_QUEUE` (default: `default`)
- `ASYNQ_RESULT_QUEUE` (default: `<ASYNQ_QUEUE>-result`)
- `ASYNQ_CONCURRENCY` (default: `10`)

### Copilot SDK runtime

- `GITHUB_TOKEN` (optional; recommended for non-interactive/CI)
- `COPILOT_CLI_PATH` (optional)
- `COPILOT_CLI_URL` (optional)

## Copilot Authentication

You can authenticate in either of these ways:

1. Logged-in Copilot CLI user session
   - Sign in with your installed Copilot CLI auth command (commonly `copilot auth login` depending on your CLI version).
   - If `GITHUB_TOKEN` is not set, the SDK uses the logged-in user path.

2. Token-based auth
   - Export `GITHUB_TOKEN` before running worker mode.
   - When token is present, SDK uses token auth for session creation.

## Quick Test (Now)

1. Start Redis:
   - `docker compose up -d redis`

2. Terminal A: run worker mode:
   - `go run ./cmd/cli run-worker`

3. Terminal B: enqueue board generation from docs:
   - `go run ./cmd/cli generate-task-board --ROOT_DIRECTORY docs --MAX_DEPTH 3`

4. Optional overrides:
   - Prompt override:
     - `go run ./cmd/cli generate-task-board --PROMPT "<your prompt>"`
   - Model override:
     - `go run ./cmd/cli generate-task-board --MODEL "<model-name>"`

## Prompt and Model Defaults

- Default model: `gpt-5.3-codex`
- Default prompt: detailed board-schema-oriented decomposition prompt embedded in CLI command defaults
- Both are overrideable per enqueue call via `--MODEL` and `--PROMPT`

## Local Shortcuts

- `task build`
- `task test`
- `task worker:run`
