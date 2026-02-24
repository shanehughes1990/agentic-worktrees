# Kanban Task Board + Document Ingestion Architecture

## 1) Purpose

This document defines the architecture for turning a folder of source documentation into a concurrent dependency-graph kanban board and autonomously executing work until completion.

Primary outcome:

- Ingest a folder of documents.
- Use Copilot SDK + a constrained prompt contract to decompose the grand objective.
- Emit an epic/task/micro-task board with explicit dependencies.
- Execute all ready work concurrently.
- Re-plan remaining work until the board is fully complete.

## 2) Scope and Non-Goals

### In Scope

- Folder-based document intake and normalization.
- Prompt-driven decomposition into epics, tasks, and micro-micro-tasks.
- Dependency graph creation and validation.
- Concurrent scheduling of ready tasks.
- Task lifecycle tracking (`not-started`, `in-progress`, `completed`, `blocked`).
- Autonomous convergence loop (plan -> execute -> reconcile -> re-plan).
- Persistence abstraction via repository ports, with JSON file storage as the initial backend.

### Out of Scope

- Rich UI design details for the kanban board.
- External multi-tenant concerns.
- Replacing the initial JSON backend with other storage engines in this phase.
- Human approval workflows beyond optional operator pause/resume controls.

## 3) Domain Terminology

- **IngestionRun**: a single execution unit over one documentation folder snapshot.
- **DocumentAsset**: one normalized source document.
- **Board**: the full task board for an `IngestionRun`.
- **Epic**: a thematic grouping of related tasks.
- **Task**: a unit of work under an epic.
- **MicroTask**: smallest independently executable unit (micro-micro task granularity).
- **DependencyEdge**: directed edge from prerequisite to dependent work item.
- **Ready Set**: all non-completed items whose dependencies are satisfied.
- **ExecutionJob**: one queued worker execution attempt for a task/micro-task.

## 4) End-to-End Flow

1. **Folder Intake**
   - Read all supported files from an input directory.
   - Normalize content (path metadata, canonical text extraction, stable ordering).
   - Compute deterministic input hash for idempotency.

2. **Decomposition Request**
   - Build a strict prompt payload containing:
     - normalized document corpus,
     - objective statement,
     - decomposition constraints,
     - schema for epics/tasks/micro-tasks/dependencies.
   - Enqueue an Asynq task to execute the Copilot SDK decomposition call.

3. **Board Materialization**
   - Validate decomposition output against schema and invariants.
   - Persist board graph (epics, tasks, micro-tasks, edges) through repository ports.
   - Reject/repair invalid graph fragments before scheduler start.

4. **Concurrent Execution Loop**
   - Compute ready set from dependency graph.
   - Dispatch ready items in parallel (bounded concurrency).
   - Record lifecycle transitions and execution artifacts.
   - Recompute ready set on each completion or unblock event.

5. **Convergence**
   - Finish when all items are `completed`.
   - If blocked items remain, either:
     - trigger targeted re-planning for unresolved branches, or
     - mark run `needs-operator-input` with unblock reasons.

## 5) Board Structure and Invariants

## Hierarchy

- `Board` has many `Epic`.
- `Epic` has many `Task`.
- `Task` has many `MicroTask`.
- Dependencies are allowed across hierarchy boundaries, including cross-epic edges.

## Graph Rules

- The dependency graph must be a DAG (no cycles).
- Every node must belong to exactly one board.
- A node is `ready` only when all predecessors are `completed`.
- A node may transition to `blocked` only with typed reason metadata.

## Status Machine

Allowed transitions:

- `not-started -> in-progress`
- `in-progress -> completed`
- `in-progress -> blocked`
- `blocked -> not-started` (after unblock condition met)

Forbidden transitions are rejected by domain rules.

## 6) Decomposition Contract (Copilot SDK)

All Copilot SDK interactions are mandatory Asynq task executions:

- No direct SDK calls from CLI/MCP handlers.
- No direct SDK calls from application services.
- SDK calls are allowed only in Asynq worker task handlers.

Decomposition prompt must request structured output only:

- `epics[]`
- `tasks[]`
- `microTasks[]`
- `dependencies[]`
- `acceptanceCriteria[]`
- `estimation` and `risk` fields (optional but typed)

Micro-micro-task requirements:

- Each micro-task must be independently executable.
- Each micro-task must include explicit done criteria.
- Each micro-task must have dependency edges that allow concurrent scheduling where possible.
- Large tasks must be recursively decomposed until they satisfy micro-task constraints.

Validation requirements:

- Schema validation (required fields, enum values, ids).
- Referential integrity (all parent/child ids exist).
- Dependency integrity (all edge endpoints exist).
- DAG validation prior to persistence.
- Duplicate semantic work detection (content similarity threshold) to reduce overlap.

If validation fails:

- quarantine invalid items,
- request a focused re-decomposition for failed segments,
- do not start scheduler for unresolved critical graph violations.

## 7) Concurrent Scheduler Design

## Scheduling Policy

1. Readiness gate first (dependency-satisfied only).
2. Priority within ready set.
3. FIFO tie-breaker.

## Dispatch Model

- Pull next ready batch up to `max_concurrency`.
- Enqueue one `ExecutionJob` per task/micro-task.
- Maintain idempotent dispatch key per `(run_id, node_id, attempt_group)`.
- Persist state transition before and after enqueue to avoid duplicate scheduling.

## Completion Handling

On job completion:

- Persist result artifact and status change.
- Emit event for dependency resolver.
- Recompute impacted subtree readiness.
- Enqueue newly-ready nodes.

## Failure Handling

- `transient` errors: retry with bounded exponential backoff.
- `terminal` errors: mark node `blocked` with actionable reason.
- Repeated terminal concentration in one branch triggers targeted re-plan.

## 8) Ingestion Pipeline Details

## Input

- A root folder path provided at run start.
- Supported file classes are handled by pluggable normalizers.

## Normalization

- Canonical UTF-8 text extraction.
- Stable document ordering.
- Source provenance captured (relative path, hash, size, timestamp).

## Snapshot Identity

- `input_hash = hash(sorted(document_hashes + run_prompt + constraints))`
- Repeated runs with same `input_hash` are idempotently recognized.

## 9) DDD Layer Responsibility Map

### interface

- Accept run request (CLI/MCP/worker admission surface).
- Validate inputs and constraints.
- Enqueue application use-case requests.
- Never perform business decomposition or scheduling logic.
- Never call Copilot SDK directly.

### application

- Orchestrate use-cases:
  - start ingestion,
  - request decomposition via Asynq enqueue,
  - schedule ready tasks,
  - reconcile execution outcomes,
  - re-plan blocked branches.
- Own process boundaries and transaction sequencing.

### domain

- Own entities and invariants:
  - board hierarchy,
  - status transition rules,
  - dependency DAG validation,
  - readiness rules.
- No infrastructure SDK, queue, or transport dependencies.

### infrastructure

- Implement ports for:
  - board repository and run repository,
  - queue adapter (Asynq),
  - Copilot SDK adapter,
  - telemetry/audit sinks.
- Run Copilot SDK operations only inside Asynq worker handlers.

## 10) Persistence Abstraction and Initial Backend

Persistence is abstracted behind domain/application-owned repository ports.

Required repository contracts:

- `BoardRepository`
  - save board graph
  - load board graph
  - update node status atomically
- `RunRepository`
  - create ingestion run
  - persist checkpoints
  - mark terminal run state
- `JobRepository`
  - persist execution attempts
  - persist outputs/artifacts

Initial backend implementation:

- JSON file backend in infrastructure layer.
- JSON files are the single source of persisted board/run/job state for the initial release.
- Writes must be atomic (temp-file + rename) to reduce corruption risk.
- Repository interfaces remain storage-agnostic so backend replacement does not affect domain/application layers.

## 11) Resilience and Observability Requirements

Mandatory execution metadata on every operation:

- `run_id`
- `task_id`
- `job_id`
- `correlation_id`

Mandatory safeguards:

- Idempotency keys for decomposition and execution dispatch.
- Checkpoints at lifecycle transitions:
  - ingestion_started/completed,
  - decomposition_task_enqueued,
  - decomposition_started/completed,
  - job_started/completed/failed,
  - run_completed/blocked.
- Typed failure class: `transient` vs `terminal`.
- Dead-letter path for exhausted retries.

Audit events should capture:

- state transition,
- actor/worker identity,
- input/output hashes,
- retry count,
- unblock instructions for blocked nodes.

## 12) Minimal Data Contracts

## Board Aggregate

- `board_id`
- `run_id`
- `status`
- `created_at`, `updated_at`

## Work Item (Epic/Task/MicroTask)

- `item_id`
- `item_type`
- `board_id`
- `parent_id` (nullable by type)
- `title`, `description`
- `status`
- `priority`
- `acceptance_criteria[]`
- `metadata` (risk, estimate, labels)

## Dependency Edge

- `edge_id`
- `board_id`
- `from_item_id`
- `to_item_id`

## Execution Job

- `job_id`
- `run_id`
- `item_id`
- `attempt`
- `failure_class`
- `result_ref`
- `started_at`, `finished_at`

## 13) Re-Planning Strategy

Re-plan is triggered when:

- critical nodes are terminally blocked,
- decomposition gaps are detected,
- objective drift is discovered from execution outputs.

Re-plan behavior:

- preserve completed nodes,
- preserve valid unresolved nodes,
- generate only missing/replacement branches,
- revalidate DAG before reinsertion.

## 14) Implementation Milestones

### M1: Ingestion Foundation + JSON Persistence

- Folder intake + normalization + snapshot hashing.
- `IngestionRun` creation and checkpoint events.
- JSON repository adapters for board/run/job state.

### M2: Asynq Decomposition Engine

- Copilot SDK prompt contract.
- Asynq task handler for decomposition calls.
- Structured decomposition output validation.
- Initial board graph persistence.

### M3: Concurrent Scheduler

- Ready-set resolver.
- Parallel dispatch with bounded concurrency.
- Retry + terminal block handling.

### M4: Autonomous Completion Loop

- Continuous reconcile/re-plan cycle.
- Completion detection and terminal run states.
- End-to-end observability and audit trail.

## 15) Risks and Controls

- **Risk**: Oversized tasks reduce parallelism.
  - **Control**: enforce micro-task decomposition thresholds.
- **Risk**: Cyclic dependencies.
  - **Control**: strict DAG validation pre-scheduling.
- **Risk**: Duplicate or overlapping tasks.
  - **Control**: semantic dedupe pass during materialization.
- **Risk**: Non-deterministic decomposition outputs.
  - **Control**: normalized inputs, strict schema, idempotent run keys.
- **Risk**: JSON backend write corruption.
  - **Control**: atomic write strategy with checkpoint recovery.

## 16) Acceptance Criteria for This Architecture

The design is satisfied when:

1. A folder of documentation can produce a validated epic/task/micro-task board.
2. The board dependency graph is acyclic and schedulable.
3. Ready tasks execute concurrently with bounded worker limits.
4. Task states are tracked with deterministic transitions.
5. Blocked branches can be re-planned without discarding completed work.
6. Copilot SDK operations run only through Asynq task handlers.
7. Board and run state persist through abstract repositories with JSON backend implementation.
8. The run terminates only when all schedulable work is completed or explicitly blocked with reasons.
