# 03 - Architecture

## Architecture Style

Layered, hexagonal-inspired architecture:

- `core`: domain models and state machine
- `app`: orchestration services, enqueueing policies, and lifecycle use cases
- `adapters`: asynq, git, agent runtime, storage, telemetry
- `interfaces`: CLI first; additional interfaces later

## Major Components

### 0) Preflight Validator (Admission Gate)

- Verifies required external dependencies before any task admission
- Checks minimum supported versions for `git` and `gh`
- Verifies `gh` auth status and required capabilities for PR lifecycle operations
- Publishes admission status; planner/enqueuer must refuse task intake when unhealthy

### 1) Planner and Enqueuer

- Selects `ready` tasks from dependency-aware board state
- Converts runnable work into typed Asynq jobs
- Applies queue routing, priority, and dedup/idempotency keys

### 2) Asynq Server and Workers

- Durable asynchronous execution with Redis-backed queues
- Handles retries, backoff, timeout, and archive/dead-letter behavior
- Executes lifecycle handlers with at-least-once delivery semantics

### 3) Workflow Engine

- Drives deterministic state transitions
- Maintains checkpoints after each critical step
- Supports replay/resume after restart

### 4) Worktree Manager

- Creates and verifies worktree directory
- Maps task ↔ branch ↔ worktree identity
- Performs cleanup and stale worktree reconciliation

### 5) Git Orchestrator

- Fetch/pull/rebase on each iteration
- Commit, push, PR create/update
- Merge with policy guardrails
- Conflict detection and strategy selection

### 6) Agent Runtime Adapter

- Integrates with GitHub Copilot ADK
- Manages execution sessions, prompts, and artifacts
- Streams agent events to telemetry bus

### 7) Operator Session Adapter (Post-MVP)

- Optional terminal multiplexer integration for operator supervision
- Not required for MVP runtime correctness
- Designed as a swappable adapter, separate from core orchestration

### 8) Observability Pipeline

- Structured event bus
- Metrics exporter
- Log sinks and run timelines

### 9) State Store

- Durable workflow state and checkpoints
- Task metadata and dependency graph
- Retry/dead-letter bookkeeping

### 10) Queue Control Adapter (Asynq)

- queue inspection for depth/latency/retry state
- operator controls for pause/resume/drain per queue
- job lifecycle introspection for run correlation

### 11) Persistence Ports

- `TaskStore` for board/task CRUD and dependency queries
- `PRDStore` for task PRD document lifecycle
- `SeedProvider` for initialization templates
- all stores are interface contracts owned by core/application layers

### 12) Realtime Sync Bus

- `ChangeWatcher` abstraction receives backend change events
- `EventBroadcaster` fans events to all interested workers/threads
- supports multiple subscribers with non-blocking backpressure policy

## Data Model (Conceptual)

- `Task`: id, title, deps, priority, status, attempts
- `Run`: run_id, task_id, worker_id, timestamps, status
- `QueueJob`: job_id, queue, task_type, retry_count, max_retry, timeout
- `Worktree`: path, branch, base_branch, revision
- `PullRequest`: number, url, status, checks, mergeable
- `ConflictRecord`: files, type, action_taken, outcome
- `Event`: run_id, type, payload, time

## Interface-Agnostic Contract

The core services expose a stable contract to interface adapters:

- `StartTask(taskID)`
- `CancelRun(runID)`
- `RetryTask(taskID)`
- `GetRunStatus(runID)`
- `ListTasks(filter)`
- `ListEvents(runID, tail)`
- `EnqueueLifecycleTask(taskID, taskType)`
- `GetQueueHealth()`
- `SeedTemplates(targetPath)`
- `WatchBoard(scope)`

CLI and future interfaces consume the same contract, ensuring behavior parity.

## Infrastructure Swap Boundary

The following adapter groups are explicitly swappable without changing orchestration logic:

- persistence backend (JSON files first, DB/message bus later)
- runtime adapter (Copilot ADK first)
- event transport (local channels first, external broker later)
- operator session adapters (tmux or alternatives, post-MVP)

## Asynq Task Topology (MVP)

- `task.prepare_worktree`
- `task.execute_agent`
- `task.validate`
- `task.open_or_update_pr`
- `task.rebase_and_merge`
- `task.cleanup`

Each task type must define timeout, retry policy, idempotency key strategy, and emitted checkpoint boundaries.
