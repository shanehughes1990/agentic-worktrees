# 14 - Persistence and Realtime Abstractions

## Purpose

Define a minimal, strict contract so the kanban/task-tree system can run on any persistent backend while starting with JSON PRD files.

Execution durability is provided by Asynq + Redis and is intentionally separate from board/PRD storage abstraction.

Ingestion and automatic board generation from scope files are defined in [16 - Kanban Ingestion and Board Seeding Pipeline](./16-kanban-ingestion-and-seeding.md).

## Core Principle

Core orchestration logic must only depend on interfaces. Storage, watch mechanics, and transport infrastructure are adapter concerns.

Queue execution semantics are first-class and must remain explicit in contracts.

## Required Interfaces (Conceptual)

### Board Repository

Responsibilities:

- load/save board state
- atomic task status transitions
- dependency graph validation hooks

### PRD Repository

Responsibilities:

- load/save task PRD documents
- list PRDs by task scope
- template materialization support

### Event Repository

Responsibilities:

- append normalized change events
- provide ordered read by offset/time
- support replay for lagging subscribers

### Queue Execution Adapter (Asynq)

Responsibilities:

- enqueue typed lifecycle jobs
- expose queue health and depth
- apply retry/dead-letter policies per task type
- provide job-level introspection for run correlation

### Change Watcher

Responsibilities:

- subscribe to board/prd/backend changes
- normalize backend-specific events into canonical event envelope
- surface transient/permanent watch failures

### Broadcaster

Responsibilities:

- publish canonical events to multiple subscribers
- non-blocking broadcast with bounded queues
- lag detection and subscriber resync signaling

## Canonical Event Envelope

Each change event should include:

- `event_id`
- `source` (`board`, `prd`, `planner`, `worker`)
- `entity_id`
- `event_type`
- `at` (UTC timestamp)
- `revision` (monotonic per source)
- `payload`
- `job_id` (when event originates from queue execution)

## JSON-First Backend (Initial Adapter)

### Directory Layout

- `tasks/board.json`
- `tasks/prd/<task-id>.json`
- `tasks/events.jsonl`

### Adapter Rules

- write operations are atomic (`tmp` file + rename)
- board updates increment board revision
- every write appends canonical event record
- watcher detects file changes and emits normalized events

## Schema Versioning and Compatibility Contract

- every persisted artifact must include a mandatory `schema_version`
- new schema versions must be additive or include deterministic migration rules
- runtime must support backward-read compatibility for all versions in the supported window
- writes should emit current canonical version unless explicitly running migration tooling
- unsupported future versions must fail closed with actionable migration error

### Supported Compatibility Window

- define a rolling supported window (example: current major + previous major)
- compatibility window policy must be documented per release
- dropping support for an old version requires migration tooling and release note callout

### Migration Rules

- migrations must be idempotent and resumable
- maintain backup/snapshot before destructive transforms
- record migration run metadata (from_version, to_version, status, timestamp)
- partial migration must never publish mixed active state without compatibility guarantees

## CLI Seeding Specification

The CLI must expose a seed/init command that:

1. creates required directory and base files
2. writes default board schema with standard lists
3. writes at least one PRD template file
4. is idempotent by default
5. supports explicit overwrite via `--force`

## Realtime Multi-Thread Consistency Model

- all workers subscribe to a shared canonical event stream
- ordering guarantee is per source revision
- workers maintain local cursor and acknowledge processed offsets
- on cursor gap or overflow, worker performs snapshot reload + replay from latest durable offset

## Backend Portability Expectations

Any future backend (SQLite, Postgres, Redis Streams, etc.) must preserve:

- canonical event schema
- ordering and replay semantics
- idempotent writes and deterministic read behavior
- compatibility with existing planner/workflow/queue contracts

## Queue Contract Expectations (Asynq Core)

- at-least-once delivery is assumed
- handlers must be idempotent and checkpoint-aware
- retry policy is explicit per task type
- terminal failures move to dead-letter/archive with reason metadata

## Failure Points and Required Handling

- write contention: retry with version check and bounded backoff
- partial write: atomic-write guarantees prevent torn state
- watch drop: automatic re-subscribe with jittered backoff
- subscriber lag: resync mode with snapshot + replay
- corrupted file record: quarantine bad record and continue from last valid offset
