# 18 - Runtime Diagnostics and Watchdogs

## Purpose

Define required runtime diagnostics commands and watchdog/healthcheck behavior so operators can inspect the system while work is actively executing.

## Diagnostics Principles

- runtime insight must be available while tasks are in-flight
- diagnostics commands are read-only unless explicitly marked control-plane
- all diagnostics queries return correlation IDs and timestamps
- diagnostics output must align with file audit trails and event stream state
- all runtime diagnostics and control capabilities should be exposed through CLI and `go-mcp` with parity

## Required Runtime Command Surface (CLI Contract)

MCP parity requirement:

- each required runtime command must have equivalent MCP tool exposure
- MCP tool names should be stable and versioned
- responses must preserve command-level correlation and status fields

## Board and Queue Status

- `runtime board status`
  - summary of list counts (`backlog`, `ready`, `in_progress`, `blocked`, `review`, `done`, `failed`)
  - runnable frontier and blocked dependency reasons
- `runtime queue status`
  - queue depth, in-progress jobs, retries, archived/dead-letter counts
  - enqueue latency and oldest pending age

## Agent and Worker Health

- `runtime agents status`
  - active agents, heartbeat age, run phase, retry state
- `runtime workers status`
  - worker IDs, assigned jobs, uptime, last heartbeat, current queue
- `runtime health`
  - one-line overall health plus component statuses (`redis`, `queue`, `audit_sink`, `git`, `gh_auth`, `schema_gate`)

## Stream and Timeline Insights

- `runtime stream tail --run <run-id>`
  - live lifecycle events for a run
- `runtime stream tail --task <task-id>`
  - task-specific events across retries and job transitions
- `runtime events query --from <ts> --to <ts> [--component <name>]`
  - filtered historical event query for incident analysis

## Worktree and Git Runtime Insight

- `runtime worktrees status`
  - active worktree paths, mapped task IDs, branch/revision, age
- `runtime git status --run <run-id>`
  - latest fetch/rebase/push/merge outcomes, conflict state, pending action
- `runtime worktree logs --worktree <name>`
  - show metadata for `<worktree-name>.log` (path, size, last write, linked run/task)
- `runtime worktree logs tail --worktree <name>`
  - tail live output from `<worktree-name>.log`

## Audit Trail Runtime Insight

- `runtime audit status`
  - audit sink health, writable checks, file rotation status
- `runtime audit tail --run <run-id>`
  - live append-only audit records for lifecycle step/path/output inspection

## Control-Plane Runtime Commands

These commands modify runtime behavior and require explicit authorization/audit entries:

- `runtime queue pause --queue <name>`
- `runtime queue resume --queue <name>`
- `runtime run cancel --run <run-id>`
- `runtime run resume --run <run-id>`
- `runtime task resume --task <task-id>`
- `runtime task retry --task <task-id>`
- `runtime workers drain`

## Checkpoint and Resume Insight Commands

- `runtime checkpoints status --run <run-id>`
  - list latest durable checkpoints and resumable stage
- `runtime checkpoints diff --run <run-id>`
  - show delta between last checkpointed state and current observed state
- `runtime resume plan --run <run-id>`
  - display planned resume path and whether in-place worktree resume is possible

## Output Requirements

Every diagnostics command response must include:

- `timestamp`
- `request_id`
- `component`
- `status`
- `data`

When scoped:

- include `run_id`, `task_id`, and `job_id` when available

## Watchdogs and Healthchecks

## Mandatory Watchdogs

- `agent_heartbeat_watchdog`
  - triggers when heartbeat exceeds threshold
- `queue_stall_watchdog`
  - detects no progress with pending jobs
- `audit_sink_watchdog`
  - detects file sink unwritable/rotation failures
- `preflight_gate_watchdog`
  - detects drift in dependency/auth/schema/model gate status
- `worktree_orphan_watchdog`
  - detects stale/orphaned worktrees not mapped to active runs
- `stream_lag_watchdog`
  - detects event subscriber lag and replay gaps

## Watchdog Actions

By severity and policy:

1. emit alert event
2. annotate affected run/task state
3. trigger bounded self-healing action
4. escalate to operator if unresolved
5. block task intake if safety-critical watchdog is unhealthy

Safety-critical watchdogs:

- `audit_sink_watchdog`
- `preflight_gate_watchdog`
- `queue_stall_watchdog` (beyond critical threshold)

## Healthcheck Endpoints (Service-Level)

- liveness: process/event loop up
- readiness: dependencies and admission gates healthy
- diagnostics: command bus responsive, event stream readable
- watchdog: all mandatory watchdog loops running

## Runtime Insight During In-Flight Execution

The system must provide consistent in-flight visibility for:

- current board frontier and blocked dependency tree
- per-agent live phase and heartbeat
- per-job queue state and retries
- latest git/worktree action and output refs
- audit records for each lifecycle step

## Failure and Degraded Modes

If diagnostics subsystem is partially degraded:

- continue execution only if safety-critical visibility remains intact
- mark diagnostics status as degraded in `runtime health`
- emit alert and include missing insight surface details

If safety-critical diagnostics/audit become unavailable:

- fail closed on new task intake
- allow existing in-flight runs to proceed only per policy
- require operator acknowledgment to continue in degraded mode
