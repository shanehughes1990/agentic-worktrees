# 09 - Containerization and Runtime

## Objectives

- Run reliably in local and containerized environments
- Keep developer experience simple
- Ensure deterministic dependencies for CI and runtime

## Runtime Requirements

- Go runtime/toolchain
- Git client
- shell utilities required by adapters
- Asynq runtime dependencies
- Redis connectivity for durable queue operations
- network access to git remote and Copilot ADK endpoints

## Mandatory Preflight Admission Checks

Before task acceptance/enqueueing, validate:

- `git` is installed and meets minimum supported version
- `gh` is installed and meets minimum supported version
- `gh auth status` is healthy for the configured host/account
- required GitHub capabilities/scopes are present for PR lifecycle operations
- default agent model is configured as `gpt-5.3-codex`
- PRD-specified per-task model overrides are validated against supported model allowlist
- unsupported model overrides are rewritten to `gpt-5.3-codex` with warning event

If any check fails, system must fail closed on task admission and keep queues in intake-blocked mode.

## Container Principles

- Multi-stage image build (build + minimal runtime)
- Non-root runtime user
- Read-only root filesystem where feasible
- Writable mounts only for:
  - state store
  - logs/artifacts
  - worktree root

## Filesystem Layout (Example)

- `/app/bin/orchestrator`
- `/data/state`
- `/data/logs`
- `/data/worktrees`

## Environment Configuration

- all secrets via environment or mounted secret files
- explicit config for concurrency and safety policies
- explicit feature flags for risky automation actions

## Health and Readiness

- liveness: process/event loop healthy
- readiness: dependencies reachable (redis/asynq, git remote, state store, runtime adapter)
- degraded mode when partial dependencies fail

## Deployment Modes

- local single-node
- container single-node
- future: distributed workers with shared state store

See [17 - Local Development Infrastructure](./17-local-development-infrastructure.md) for the concrete `docker compose` profiles and Air hot-reload workflows.

## Backup and Recovery

- persist state store on durable volume
- periodic snapshot/export of run ledger and task graph
- restart resumes from latest checkpoints
