# 01 - Scope and Product Definition

## Problem Statement

Teams want to run multiple AI agents against the same codebase in parallel. Traditional branch-based workflows create coordination overhead, stale state, and frequent merge pain. The project solves this by using Git worktrees as isolation boundaries and an orchestration engine that automates the full task cycle from intake to merged PR.

The core challenge is durability: task execution must survive crashes, restarts, and transient infrastructure failures without losing progress or producing unsafe duplicate side effects.

## Product Vision

Create a dependable orchestration platform where:

1. Tasks are scheduled into durable Asynq queues
2. Each task is executed in an isolated worktree
3. Agents operate with realtime supervision and policy controls
4. Rebase/conflict handling is automated as much as possible
5. Successful work is merged back safely into origin branch
6. The entire cycle repeats continuously using latest origin state
7. Failed work is retried or dead-lettered using explicit policies

## In Scope (v1)

- Go orchestration core
- First interface: CLI using `urfave/cli/v3`
- Asynq as the authoritative durable task execution engine
- Redis-backed durable queues for task processing
- Dependency-aware mini-kanban task system
- End-to-end automation cycle:
  - materialize runnable tasks from board/dependencies
  - enqueue typed Asynq jobs with deterministic payloads
  - claim and execute jobs in workers
  - create/sync worktree
  - run agent loop
  - validate gates
  - create/update PR
  - rebase/resolve/merge
  - ack success or retry/dead-letter on failure
- Realtime telemetry for all running agents
- Container-ready deployment and local dev

## Out of Scope (v1)

- Rich web UI (may be added as additional interface)
- tmux interface/session supervision (planned post-v1)
- Multi-repo orchestration at enterprise scale
- Non-Git backends
- Automatic semantic conflict resolution beyond defined policies
- Replacing Asynq with a custom in-memory execution core

## Primary Users

- Solo builders running many agents in parallel
- Small teams orchestrating AI-assisted delivery
- Platform engineers standardizing autonomous workflows

## Success Criteria

- Parallel execution with no shared-worktree collisions
- Deterministic lifecycle with resumable runs after failures
- Durable execution semantics across orchestrator/worker restarts
- Merge throughput increased vs manual branch workflow
- Conflict handling that is safe and operator-auditable
- Operational visibility sufficient for debugging every run
