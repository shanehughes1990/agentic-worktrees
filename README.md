# Agentic Git Worktrees

Durable autonomous task orchestration with Asynq + Redis, ADK-backed board planning, and file-based auditability.

## Binaries

- `cmd/cli`: operational CLI (`preflight`, `status`, `ingest`, `version`)
- `cmd/worker`: Asynq worker that executes ADK planning tasks

## Quick Start

1. Start Redis:
   - `docker compose up -d redis`
2. Set required environment variables:
   - `APP_COPILOT_ADK_BOARD_URL` (required for worker)
   - optional: `APP_REDIS_ADDR`, `APP_ASYNQ_QUEUE`, `APP_BOARD_PATH`, `APP_CHECKPOINT_PATH`, `APP_AUDIT_PATH`
3. Run preflight:
   - `go run ./cmd/cli preflight`
4. Check runtime status:
   - `go run ./cmd/cli status`
5. Enqueue board planning from docs:
   - `go run ./cmd/cli ingest --scope docs`
6. Start worker:
   - `go run ./cmd/worker`

## Runtime Artifacts

- board output: `state/board.json`
- checkpoints: `state/checkpoints.json`
- audit log: `logs/audit.log`

## Local Tasks

Use `Taskfile.yml` shortcuts:

- `task build`
- `task test`
- `task cli:preflight`
- `task cli:status`
- `task cli:ingest`
- `task worker:run`

## Notes

- ADK execution is worker-only via Asynq tasks.
- CLI only validates input and enqueues deterministic task payloads.
