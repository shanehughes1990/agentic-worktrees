# Copilot CLI Session Filesystem Artifacts (Worker Exploration)

Date: 2026-03-04
Runtime inspected: running worker container `v1-worker-1`
Copilot CLI version: `0.0.421`

## Scope

This exploration executed exactly one tagged Copilot CLI request inside the running worker container, then enumerated filesystem artifacts produced by that request.

Tagged request:
- `PROBE_ID`: `copilot_probe_1772593497`
- Prompt: `Reply with exactly: copilot_probe_1772593497`
- Command used:

```bash
copilot -p "Reply with exactly: copilot_probe_1772593497" --model gpt-5.3-codex --allow-all --add-dir /tmp/copilot_probe_1772593497
```

Probe directory:
- `/tmp/copilot_probe_1772593497`

## Direct output artifacts from the request

Created/updated in probe dir:
- `/tmp/copilot_probe_1772593497/copilot_stdout.txt`
- `/tmp/copilot_probe_1772593497/copilot_stderr.txt`
- `/tmp/copilot_probe_1772593497/files_newer_than_marker.txt`

Observed stdout:
- `copilot_probe_1772593497`

Observed stderr usage metadata:
- `Total usage est: 1 Premium request`
- `API time spent: 2s`
- `Total session time: 3s`
- `Total code changes: +0 -0`
- Model accounting line includes `gpt-5.3-codex`

## Files modified after marker (request-correlated)

`find /root /tmp /app -type f -newer /tmp/copilot_probe_1772593497/.before` returned these request-correlated files:

- `/root/.copilot/logs/process-1772593497544-141.log`
- `/root/.copilot/session-state/9ab5f95d-be14-49d4-8069-2f41393e74e6/checkpoints/index.md`
- `/root/.copilot/session-state/9ab5f95d-be14-49d4-8069-2f41393e74e6/events.jsonl`
- `/root/.copilot/session-state/9ab5f95d-be14-49d4-8069-2f41393e74e6/workspace.yaml`
- `/tmp/copilot_probe_1772593497/copilot_stderr.txt`
- `/tmp/copilot_probe_1772593497/copilot_stdout.txt`
- `/tmp/copilot_probe_1772593497/files_newer_than_marker.txt`

## Session directory discovered for this exact request

- Session ID: `9ab5f95d-be14-49d4-8069-2f41393e74e6`
- Session path: `/root/.copilot/session-state/9ab5f95d-be14-49d4-8069-2f41393e74e6`

Session tree:
- `checkpoints/index.md`
- `events.jsonl`
- `workspace.yaml`
- `files/` (present, empty in this probe)
- `research/` (present, empty in this probe)

## Key artifact contents

### `workspace.yaml`
Contains session metadata:
- `id: 9ab5f95d-be14-49d4-8069-2f41393e74e6`
- `cwd: /app`
- `created_at` / `updated_at`
- `summary: "Reply with exactly: copilot_probe_1772593497"`

### `events.jsonl`
Contains structured event stream for the turn:
- `session.start` with selected model and cwd
- `user.message` containing prompt and transformed content
- `assistant.turn_start`
- `assistant.message` with response text (`copilot_probe_1772593497`) and opaque/encrypted payload fields
- `assistant.turn_end`

Notable detail from `user.message.transformedContent`:
- Copilot injected contextual wrapper text (current datetime and reminder block) around user prompt.

### `checkpoints/index.md`
- Present but minimal for this request (no substantive checkpoints created).

### `process-1772593497544-141.log`
Contains process-level runtime logs:
- CLI startup/version/node version
- MCP client startup/connection lines
- workspace initialization line with the same session ID
- model request lifecycle lines
- transport/client close lines at end

## Global Copilot filesystem layout observed in worker

Under `/root/.copilot`:
- `logs/` (per-process runtime logs)
- `session-state/<session-id>/` (per-session artifacts)
- `pkg/linux-arm64/0.0.421/` (installed Copilot CLI package/runtime assets)

Other session directories were also present from previous runs (for example `2a7c4cc6-c1cf-46ad-88ae-f690bcc632c6`), including richer files such as `session.db` and `plan.md` when workflows are more complex.

## Conclusions

For a single simple `copilot -p` run in this worker image, request-specific filesystem artifacts are primarily:
1. Probe command outputs in the working/probe directory (`stdout`/`stderr`).
2. A new per-process log file in `/root/.copilot/logs/`.
3. A new/updated per-session directory in `/root/.copilot/session-state/<session-id>/` containing:
   - `workspace.yaml`
   - `events.jsonl`
   - `checkpoints/index.md`
   - optional subdirs `files/` and `research/`.

This gives a practical forensic path for correlating one prompt to on-disk artifacts:
- start with marker timestamp,
- match process log file creation time,
- read session ID from logs/workspace,
- inspect `events.jsonl` for exact prompt/response sequence.
