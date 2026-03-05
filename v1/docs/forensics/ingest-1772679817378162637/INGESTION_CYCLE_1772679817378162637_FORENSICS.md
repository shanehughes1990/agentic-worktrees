# Ingestion Cycle Forensics: ingest-1772679817378162637

## Scope
This document captures filesystem and runtime artifacts that correlate to the specific ingestion cycle with:

- `run_id`: `ingest-1772679817378162637`
- `board_id`: `v1_documentation_1772679817378162637`
- `stream_id` in prompt payload: `ingestion-agent-1772679817378651512`

All observations were taken from the running worker container `v1-worker-1`.

## Worker and Runtime Context

- Container: `v1-worker-1`
- Copilot home inside container: `/root/.copilot`
- Session state root: `/root/.copilot/session-state`
- Ingestion sandbox root for this run: `/tmp/ingestion-sandbox-2107252471/sandbox`

## Session Correlation

Two session directories were present at observation time:

- `/root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947` (newer)
- `/root/.copilot/session-state/63ddd77c-b217-4569-a393-52914746b11b` (older)

The ingestion cycle maps to:

- `/root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947`

because `events.jsonl` in that directory contains a `session.start` entry with:

- `sessionId`: `75c3f3fd-b59c-44f5-b245-5ca13f805947`
- `selectedModel`: `gpt-5.3-codex`
- `startTime`: `2026-03-05T03:03:39.693Z`
- `context.cwd`: `/tmp/ingestion-sandbox-2107252471/sandbox`

and a user payload that explicitly includes:

- `board_id = v1_documentation_1772679817378162637`
- `run_id = ingest-1772679817378162637`
- output path requirement: `/tmp/ingestion-sandbox-2107252471/sandbox/taskboard.json`

## Session Artifact Files (Correlated)

Primary chat/event artifact:

- `/root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/events.jsonl`

Companion session artifacts:

- `/root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/workspace.yaml`
- `/root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/session.db`
- `/root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/checkpoints/index.md`
- `/root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/plan.md`

Observed file mtimes (UTC) for key files:

- `2026-03-05 03:07:33.398665010 +0000` `events.jsonl`
- `2026-03-05 03:03:40.409510014 +0000` `workspace.yaml`
- `2026-03-05 03:04:06.356294012 +0000` `session.db`

## Run-Local Tool Output Files (Correlated)

Inside `/tmp`, this cycle produced Copilot tool-output spill files:

- `/tmp/copilot-tool-output-1772679825703-8cxzmw.txt`
- `/tmp/copilot-tool-output-1772679829902-e8ebiy.txt`
- `/tmp/copilot-tool-output-1772679840707-zghvrw.txt`
- `/tmp/copilot-tool-output-1772679956712-83u1gd.txt`

These paths are referenced in `events.jsonl` for the same session, alongside operations over the same sandbox path.

## Repository and Branch Inputs Captured in Session Payload

The same `events.jsonl` payload includes synchronized source repository metadata:

- `repository_id=agentic-repo-1`
- `repository_url=https://github.com/shanehughes1990/agentic-worktrees`
- `source_branch=v1`
- `local_dir=repos/agentic-worktrees`

This matches the sandbox repository layout observed under:

- `/tmp/ingestion-sandbox-2107252471/sandbox/repos/agentic-worktrees`

## Non-Correlated Neighbor Session

The older session directory:

- `/root/.copilot/session-state/63ddd77c-b217-4569-a393-52914746b11b`

contains a prompt-refinement conversation (different objective and different `cwd`), so it is not the ingestion execution cycle above.

## Quick Retrieval Commands

```bash
# Enter container shell
docker exec -it v1-worker-1 bash

# Read correlated session events
tail -n 200 /root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/events.jsonl

# Check session file timestamps
stat -c '%y %n' \
  /root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/events.jsonl \
  /root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/workspace.yaml \
  /root/.copilot/session-state/75c3f3fd-b59c-44f5-b245-5ca13f805947/session.db

# Inspect sandbox root used by this cycle
find /tmp/ingestion-sandbox-2107252471/sandbox -maxdepth 3 -type d | head -n 50
```
