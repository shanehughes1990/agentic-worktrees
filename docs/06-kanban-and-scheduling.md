# 06 - Kanban Model and Scheduling

## Goal

Provide a mini-kanban system that supports parallel task lists with dependencies while maintaining predictable agent throughput.

Execution is queue-driven: runnable nodes are materialized into durable Asynq jobs.

For ingestion from scope files and automatic board generation, see [16 - Kanban Ingestion and Board Seeding Pipeline](./16-kanban-ingestion-and-seeding.md).

## Board Model

- Multiple lists (e.g., `Backlog`, `Ready`, `In Progress`, `Review`, `Done`, `Blocked`)
- Task cards with metadata:
  - id, title, description
  - priority, labels
  - dependency IDs
  - child task IDs (parallel task tree edges)
  - attempt count and retry budget
  - target branch
  - PRD file reference

## Minimal v1 Persistence Shape (JSON-first)

- `tasks/board.json`: lists, task summaries, dependency edges, status
- `tasks/prd/<task-id>.json`: detailed PRD for each task
- `tasks/events.log` (or JSONL): append-only local change ledger

This format is the first backend implementation and acts as the seed template contract.

## Dependency Semantics

- Dependencies form a DAG
- Parent/child tree relationships are allowed when acyclic
- A task becomes `ready` only when all dependencies are `done`
- Cycle detection runs on every dependency update
- Invalid dependency mutations are rejected

## Scheduling Strategy

Recommended v1 strategy:

1. Build ready queue from DAG + status filters
2. Sort by explicit priority, then age
3. Convert selected tasks into typed Asynq jobs
4. Route jobs to priority queues with configured weights
5. Respect max parallelism and worker concurrency constraints

## Parallelism Controls

- Global concurrency limit
- Per-repo/branch concurrency limit
- Optional per-label limits (e.g., migration tasks)
- Queue-level concurrency, strict-priority, and weighted fair scheduling controls
- Handler-specific timeout and max retry policy

## CLI Seed Contract

CLI must provide an initialization command that:

1. creates missing `tasks/` directory structure
2. writes default `board.json`
3. writes PRD template file(s)
4. does not overwrite existing files unless `--force` is explicitly provided

## Realtime Watch Contract

- every backend write emits a normalized change event
- active worker threads subscribe to scoped streams (board/task/global)
- events are broadcast to all subscribers with per-subscriber cursor tracking
- if consumer lag exceeds threshold, subscriber enters resync mode from persisted snapshot
- queue lifecycle events are included for each task transition

## Retry and Dead-Letter Policy

- Retry transient failures with per-task-type Asynq retry policy and capped exponential backoff
- Move terminal failures to dead-letter/archive after max attempts
- Require explicit operator action to requeue dead-letter task

## Task Materialization Rules

- one DAG node may emit multiple lifecycle jobs, but only one active lifecycle stage at a time
- enqueue must include idempotency key derived from task ID + lifecycle stage + revision
- downstream stage is enqueued only after upstream handler success checkpoint

## Starvation Prevention

- Priority aging to prevent low-priority indefinite waiting
- Periodic planner/queue rebalance pass
