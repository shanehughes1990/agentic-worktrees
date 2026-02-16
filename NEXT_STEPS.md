# NEXT_STEPS

## Mission (Confirmed)

Build the **minimum working application** for autonomous agentic programming with:

- durable execution (Asynq + Redis)
- worktree-based parallel task execution
- autonomous board generation + run-to-completion flow
- full test coverage for core packages and critical runtime paths
- clear runtime diagnostics, watchdogs, and auditability

## What Has Been Completed So Far

### 1) Core Product Documentation

Created and expanded full project documentation set in [docs/README.md](docs/README.md), including:

- scope and product definition
- functional/non-functional requirements
- architecture
- workflow/state machine
- git conflict strategy
- kanban + scheduling
- ingestion/board seeding pipeline
- observability, runtime diagnostics, watchdogs
- failure/recovery and preserve-resume policy
- security/safety controls
- test/SLO/release criteria
- delivery plan
- go-mcp interface integration

### 2) Critical Policies Now Documented

- autonomous walk-away mode as default
- preflight admission gate before task intake
- schema versioning + backwards compatibility requirements
- default model policy (`gpt-5.3-codex`) + per-task override fallback
- mandatory file-based audit logs by default
- per-worktree thread logs (`<worktree-name>.log`)
- preserve half-completed work and resume from checkpoint-first

### 3) Local Infrastructure Artifacts Added

- [docker-compose.yml](docker-compose.yml) (Redis + optional app/dev profiles)
- [Dockerfile](Dockerfile)
- [.dockerignore](.dockerignore)
- [.air.toml](.air.toml)
- [Taskfile.yml](Taskfile.yml)
- [.goreleaser.yaml](.goreleaser.yaml)
- [.gitignore](.gitignore) entries for task/goreleaser/air artifacts

### 4) Dev Workflow Documentation Added

- [docs/17-local-development-infrastructure.md](docs/17-local-development-infrastructure.md)

## Current Reality Check

The repository currently has **documentation and infrastructure scaffolding**, but the **core application code is not yet implemented**.

No runtime packages, command entrypoint, or tests exist yet for the orchestration system.

## Minimum Working Application Build Plan

## Phase A — Core Runtime Skeleton (Implement First)

1. Create app entrypoint: `cmd/agentic-worktrees`
2. Add config loader (env + file)
3. Add structured logger + mandatory file audit sink
4. Add Redis + Asynq bootstrap
5. Add health/preflight module (git/gh/auth/schema/model/audit sink checks)

**Exit criteria:** app boots, validates preflight, and blocks intake when unhealthy.

## Phase B — Domain + Persistence

1. Implement core domain models (Task, Run, QueueJob, Worktree, Event)
2. Implement JSON board/PRD repositories with versioned schema handling
3. Implement compatibility normalization for older schema versions
4. Implement ingestion pipeline (scope -> DAG -> lanes -> board artifacts)

**Exit criteria:** board generation works from input scope and writes valid artifacts deterministically.

## Phase C — Execution Pipeline

1. Implement Asynq task topology handlers:
   - prepare_worktree
   - execute_agent
   - validate
   - open_or_update_pr
   - rebase_and_merge
   - cleanup
2. Implement checkpoint store and resume logic
3. Implement preserve-work policy for recoverable failures

**Exit criteria:** tasks run end-to-end and resume in-place after simulated interruption.

## Phase D — Git/PR + Worktree Engine

1. Worktree lifecycle manager (`.worktrees/<task-id>-<slug>`)
2. Branch/merge target capture at run start
3. Rebase/conflict handling strategy execution
4. PR open/update/merge automation

**Exit criteria:** successful task PRs merge into captured origin start branch.

## Phase E — Runtime Diagnostics + Watchdogs

1. Implement runtime status command surfaces (board/queue/agents/worktrees/audit)
2. Implement stream tail and checkpoint insight commands
3. Implement mandatory watchdog loops

**Exit criteria:** operator/agent can inspect and control active runs in flight.

## Phase F — go-mcp Interface

1. Add `go-mcp` server adapter
2. Map core operations to MCP tools with parity to CLI
3. Apply same authz/audit/policy constraints

**Exit criteria:** MCP client can ingest/start/inspect/resume safely.

## Testing Strategy (Full Coverage Goal)

## Coverage Target

- Target 100% coverage for **core decision logic** and **critical safety paths**.
- Target very high coverage for integration surfaces (queue/persistence/git/runtime adapters), acknowledging some I/O glue may be best validated with integration tests rather than line-only metrics.

## Required Test Layers

1. **Unit tests**
   - state transitions
   - planner/lane assignment
   - schema migration/compatibility checks
   - preflight validators
   - model override fallback logic
   - retry/backoff classification logic

2. **Integration tests**
   - Redis/Asynq lifecycle
   - JSON repositories + schema upgrades
   - worktree/git operations on fixture repos
   - PR workflow with mocks/fakes

3. **E2E tests**
   - scope ingestion -> board generation -> autonomous run -> merged completion
   - interruption and resume-from-checkpoint behavior

4. **Failure/chaos tests**
   - agent crash mid-task
   - audit sink unavailable
   - queue stall
   - schema incompatibility
   - network/auth transient failure

## Definition of Done for MVP

MVP is done when all are true:

- board generation works from scope file
- autonomous run can be started and complete board tasks
- successful PRs merge into run-start captured origin branch
- resume-from-checkpoint works after injected failures
- mandatory audit trail and per-worktree logs are produced
- runtime diagnostics + watchdogs are available
- go-mcp interface supports core operations
- full core test suite passes with coverage target met

## Immediate Next Actions (Execution Order)

1. Scaffold `cmd/agentic-worktrees` and core package layout
2. Implement config + logger + audit sink + preflight module
3. Stand up Asynq worker server with one no-op lifecycle task
4. Add first test suite (state machine + preflight + schema compatibility)
5. Implement ingestion pipeline and JSON artifact writer
6. Iterate until end-to-end autonomous board run is green

## Notes

- Human intervention should remain exception-only.
- Any fallback from autonomous behavior must be explicit, auditable, and policy-driven.
- Work preservation and resume are default behavior, not optional behavior.
