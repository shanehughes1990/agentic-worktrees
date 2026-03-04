# Worker Running CLI/Session Snapshot

Generated: 2026-03-04 UTC (captured live from running worker container)

## Scope

This document captures runtime details from the currently running worker container and the active Copilot CLI ingestion process.

## Runtime Target

- Compose project path: `/Volumes/External/workspace/projects/personal/agentic-worktrees/v1`
- Worker container: `v1-worker-1`
- Worker service image/entrypoint: `v1-worker` / `/app/worker`
- Worker container user: `root`

## Active Copilot CLI Process

- PID: `51`
- Parent PID: `1` (`/app/worker`)
- Elapsed during capture: ~`01:46`
- Binary: `/usr/local/bin/copilot`
- CWD: `/tmp/ingestion-sandbox-1706786542/sandbox`

### Observed command-line payload (sanitized summary)

The active process is a `copilot -p ...` invocation with a full ingestion prompt/contract inline, including:

- taskboard synthesis contract and schema constraints,
- output target path,
- selected model,
- source repository context,
- execution continuity fields.

Key extracted values from the running command line:

- `run_id`: `ingest-1772598849983546759`
- `stream_id`: `ingestion-agent-1772598849983559551`
- `session_id`: *(empty string in command payload)*
- model: `gpt-5.3-codex`
- output path: `/tmp/ingestion-sandbox-1706786542/sandbox/taskboard.json`
- source repository URL: `https://github.com/shanehughes1990/agentic-worktrees`
- source branch (repo context): `v1`

## Process-level Forensics

From `lsof -p 51`:

- Executable: `/usr/local/bin/copilot`
- Loaded addon path includes:
  - `/root/.copilot/pkg/linux-arm64/0.0.421/prebuilds/linux-arm64/pty.node`
- Outbound established TCP connections to GitHub HTTPS endpoint:
  - `lb-140-82-113-22-iad.github.com:https`
- stdin is `/dev/null`; stdout/stderr piped by parent worker process.

## Session/Environment Signals

### Env keys seen in worker container (filtered)

- `OTEL_SERVICE_NAME=agentic-worker`
- `DATABASE_DSN=postgres://postgres:postgres@postgres:5432/agentic_orchestrator?sslmode=disable`
- `GOOGLE_CLOUD_STORAGE_BUCKET=agentic-orchestrator-cdn`
- `GOOGLE_CDN_KEY_NAME=agentic-orchestrator-filestore`

### Env keys seen for copilot process (`/proc/51/environ`, filtered)

- `PWD=/tmp/ingestion-sandbox-1706786542/sandbox`

No dedicated `COPILOT_*`, `SESSION_*`, or `ACP_*` process env vars were present in filtered process-environment output.

## Filesystem Artifacts

### Worker project/cache directories

- `/app/.agentic-orchestrator/projects/agentic_orchestrator/repositories/agentic-worktrees`
- `/app/.agentic-orchestrator/projects/unscoped/repositories`

### Worker log directory

- `/app/.agentic-orchestrator/logs` exists, but no log files were present at capture time.

### Ingestion sandbox root

- `/tmp/ingestion-sandbox-1706786542/sandbox`
- Synchronized repo path:
  - `/tmp/ingestion-sandbox-1706786542/sandbox/repos/agentic-worktrees`

### Sandbox git state

- Branch: `v1`
- HEAD commit: `1f693eaa5edeb0d26e36de540391a267b880766f`
- Remote origin:
  - `https://github.com/shanehughes1990/agentic-worktrees`

### Taskboard output file

- Target path from process args:
  - `/tmp/ingestion-sandbox-1706786542/sandbox/taskboard.json`
- At capture time, preview output was empty (no JSON content returned by `cat ... | head`).

## Worker Logs (compose stream)

Observed in `docker compose logs worker`:

- Worker runtime startup succeeded.
- Asynq worker started successfully.
- Reconcile heartbeat loop healthy.
- Prior repository reconcile failure logged:
  - error mentions `git repository add -B ...` invalid git subcommand.

No explicit ingestion `run_id`/`task_id`/`job_id` log lines were emitted in the filtered log tail during this capture window.

## Command Evidence (executed)

1. `docker compose ps`
2. `docker exec v1-worker-1 sh -lc '... ps aux ... env ... ls ...'`
3. `docker exec v1-worker-1 sh -lc '... /proc/51/cmdline ... /proc/51/environ ...'`
4. `docker compose logs --no-color worker | tail -n 250`
5. `docker compose logs --no-color worker | grep -Ei "ingest-|ingestion-agent|copilot|taskboard.json|run_id|job_id|task_id|session_id|stream_id" | tail -n 200`
6. `docker exec v1-worker-1 sh -lc '... lsof -p 51 ... git rev-parse ...'`
7. `docker exec v1-worker-1 sh -lc 'find /app/.agentic-orchestrator/logs ...'`

## Notes

- The active CLI run is clearly present and tied to ingestion continuity via `stream_id`; `session_id` is currently blank in process arguments.
- If you want continuous tracking, run the same process and lsof captures repeatedly (or every N seconds) while the job is still active.
