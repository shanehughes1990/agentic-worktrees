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

## Idle Detection Research (Outside `events.jsonl`)

This section documents probes run directly in the worker container to detect `working` vs `done` vs `idle` without relying on `events.jsonl` parsing.

### Controlled Long-Run Trigger Used

Executed Copilot CLI non-interactive prompts that force a long shell step (`sleep 60` / `sleep 90`) so the process remains active long enough for probing.

Example:

```bash
copilot --allow-all --output-format json -p "Run exactly this shell command first: echo start && date -u && sleep 60 && date -u && echo end. After it completes, summarize in one sentence."
```

### Probe 1: Process Liveness and CPU (Best Immediate Signal)

Observed during the forced long run:

- parent shell process (`sh -lc ...`) remains present
- child `copilot` process remains present with elapsed time and non-zero CPU at times

Sample observed output (sanitized):

```text
pid=... stat=Ss etime=00:59 args=sh -lc copilot --allow-all --output-format json -p ...
pid=... stat=Sl %cpu=7.4 etime=00:59 args=copilot --allow-all --output-format json -p ...
```

Interpretation:

- `working`: matching `copilot` process exists for active run
- `done`: matching `copilot` process disappears
- `idle`: process exists but no companion activity signals change for a grace window

### Probe 2: Copilot Process Logs (`~/.copilot/logs/process-*.log`)

Observed log markers include:

- startup + workspace init
- repeated `Sending request to the AI model` groups with `Start`/`End`
- transport close near process shutdown (`MCP transport ... closed`)

Useful signals outside `events.jsonl`:

- latest log file `mtime` and `size` increasing -> active work
- no growth for sustained window while process still alive -> likely idle/waiting
- final close markers + process exit -> done

### Probe 3: CLI Stdout JSON Event Stream (When Running With `--output-format json`)

Observed lifecycle in terminal output:

- `assistant.turn_start`
- `tool.execution_start`
- `tool.execution_partial_result`
- `tool.execution_complete`
- `assistant.turn_end`

This stream is distinct from session-state file tailing and can be consumed directly from the running command output.

### Probe 4: Session Companion Artifacts (Non-`events.jsonl`)

Observed in newest session directory:

- `workspace.yaml`: created early and stable
- `checkpoints/index.md`: present, may remain unchanged for short runs
- `plan.md`: not guaranteed to exist
- `session.db`: inconsistent as activity signal in this run set (absent initially, later created as 0 bytes in one session)

Conclusion: these are secondary hints, not primary liveness indicators.

### Practical Status Classification (Recommended)

For a single active Copilot run, combine probes in this order:

1. `done`
  - no matching `copilot` process for the tracked run
  - and log file no longer growing
2. `working`
  - matching `copilot` process exists
  - and at least one of: CPU activity, log growth, stdout JSON tool/turn events
3. `idle`
  - matching `copilot` process exists
  - but no log growth and no stdout lifecycle deltas for N seconds (suggest 20-30s)

### Why This Matters

This gives reliable run-state detection even if `events.jsonl` ingestion is delayed, temporarily unavailable, or intentionally excluded from a detection path.
