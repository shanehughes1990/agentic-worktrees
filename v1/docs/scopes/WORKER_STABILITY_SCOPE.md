# Feature Scope: Worker Stability Monitoring and Intervention

## Objective

Define a v1 feature scope that detects unstable/failed agent runs early, classifies failure modes deterministically, and emits actionable intervention feedback to desktop and optional external channels.

## Problem Statement

Current worker execution can fail in ways that are hard to see quickly:

- Agent process exits unexpectedly.
- Agent is alive but not making progress (stuck/hung).
- Agent is waiting for user approval/input with no operator visibility.
- Session context exists on disk but is not translated into operational state.

The result is delayed intervention, silent run degradation, and poor operator trust.

## Inputs and Evidence

This scope is based on:

- Forensic ingestion bundle:
  - `v1/docs/forensics/ingest-1772679817378162637/`
- Session-state artifact analysis:
  - `events.jsonl` is the strongest live activity signal.
  - `session.db` is durable context/history.
  - `workspace.yaml` provides context binding but weak liveness value.
  - `checkpoints/index.md` and `plan.md` are useful for human intervention context.
- Live worker investigation findings (non-`events.jsonl` probes):
  - Process liveness via `ps` (PID, elapsed time, process state, CPU) is a strong runtime-active signal.
  - Copilot process log growth (`~/.copilot/logs/process-*.log` mtime/size) is a strong progress signal.
  - Copilot JSON stdout lifecycle (`assistant.turn_start`, `tool.execution_start`, `tool.execution_complete`, `assistant.turn_end`) is a strong session activity signal when available.
  - `workspace.yaml`, `checkpoints/index.md`, and `plan.md` are useful context signals, but weak as sole liveness proof.
  - `session.db` existence/size may be inconsistent during active execution and must not be the primary runtime liveness signal.
- Reference architecture findings from `.docs/agent-orchestrator`:
  - Runtime liveness is treated as authoritative first.
  - Session metadata is control-plane truth.
  - Restore is guarded by explicit restorable/non-restorable state rules.
  - Agent-specific activity enrichments are layered on top of liveness.

## Scope Boundaries

In scope:

- Worker-side stability classifier based on multiple signals.
- Persistent stability state model for sessions/runs.
- Intervention event emission for desktop via internal control-plane listeners.
- Abstract listener contract that external providers can implement in phase 2.
- Deterministic retry/intervention thresholds.
- Operator-visible feedback with correlation IDs.

Out of scope:

- Re-architecting execution engine beyond stability surfaces.
- New control-plane REST endpoints (must remain GraphQL + health constraints).
- Agent model behavior changes unrelated to monitoring/intervention.

## Stability Signal Model

Use a layered signal model; no single source is sufficient.

Primary signals:

1. Runtime liveness

- Process/tmux/container alive checks for active session.

1. Session activity freshness

- Last event age from `events.jsonl` (or equivalent stream source).
- Event type pattern indicates `active`, `ready`, `waiting_input`, `blocked`.

1. Progress/checkpoint drift

- No meaningful checkpoint/output advancement for threshold window while runtime is still alive.

1. Durable session context

- `session.db` and persisted run/session metadata used for recovery context and post-mortem mapping.

### Heartbeat Coverage and Gaps (Current)

Current heartbeat coverage already implemented:

- Worker registry heartbeat request/response flow validates worker lease and epoch health.
- Deadline miss invalidates/deregisters unhealthy worker leases.

What this covers well:

- API <-> worker runtime reachability.
- Worker process-level health and lease continuity.

What this does not guarantee:

- Copilot CLI session progress for a specific `session_id`.
- Whether agent is actively running tools/model turns versus idle/waiting.
- Whether task-level execution is stuck while worker process remains healthy.

### Layered Session Heartbeats (Required Additions)

Add session-scoped heartbeat layers for session inspection ingestion, persistence, and upstream fan-out.

Layer 0: runtime lease heartbeat (existing)

- Source: worker registry heartbeat request/response.
- Scope: worker-level (`worker_id`, `epoch`).
- Purpose: runtime transport/lease health baseline.

Layer 1: session process heartbeat

- Source: worker-side process probe for correlated Copilot process (`pid`, `stat`, `%cpu`, `etime`).
- Scope: `session_id` correlation track.
- Purpose: prove session process exists and is not exited.

Layer 2: session activity heartbeat

- Source: session activity freshness from parsed session lifecycle signals (including `events.jsonl`-derived events).
- Scope: `session_id`.
- Purpose: prove recent assistant/tool/session event activity.

Layer 3: tool/turn heartbeat

- Source: Copilot tool/turn lifecycle (`tool.execution_start`, `tool.execution_complete`, `assistant.turn_start`, `assistant.turn_end`) from session stream and/or JSON stdout capture when present.
- Scope: `session_id`.
- Purpose: distinguish active execution from passive runtime survival.

Layer 4: log-progress heartbeat

- Source: process log growth (`process-*.log` mtime/size deltas) and model-request group cadence.
- Scope: `session_id` to process-log correlation.
- Purpose: independent progress signal outside session-state events.

Layer 5: context/progress heartbeat (weak signal)

- Source: `checkpoints/index.md`, `plan.md`, and other session companion artifact updates.
- Scope: `session_id`.
- Purpose: context drift/progress hints, not primary liveness.

### Heartbeat Quorum Rule for Running Confidence

Session status confidence must be derived from quorum, not one source.

Required confidence model:

- `running_confident` when:
  - runtime lease heartbeat is fresh
  - session process heartbeat is fresh
  - and at least one of {session activity heartbeat, tool/turn heartbeat, log-progress heartbeat} is fresh
- `running_degraded` when runtime+process are fresh but no progress heartbeat is fresh in grace window.
- `idle_suspected` when process is alive but progress heartbeats are stale for threshold window.
- `done_or_exited` when process heartbeat is absent or explicit terminal signal is observed.

Freshness windows (initial defaults):

- runtime lease heartbeat: <= 30s
- session process heartbeat: <= 15s
- activity/tool/log progress heartbeat: <= 20s
- idle suspicion promotion: >= 2 consecutive stale windows

## Stability State Contract

Define explicit health/intervention states:

- `healthy_active`
- `healthy_waiting_input`
- `stale_needs_nudge`
- `stuck_needs_intervention`
- `exited_unexpected`

Required properties on each state transition event:

- `run_id`
- `task_id`
- `job_id`
- `project_id`
- `session_id`
- `previous_state`
- `new_state`
- `detected_at`
- `reason_code`
- `reason_summary`

## Detection and Threshold Policy

Initial policy defaults (configurable via typed config, not ad-hoc flags):

- Waiting-input threshold: 2 minutes -> emit intervention notification.
- Stuck threshold: 5 minutes without event/checkpoint progress while runtime alive.
- Unexpected exit: immediate critical intervention event.
- Flapping guard: debounce transitions to avoid noisy oscillation.
- Heartbeat quorum degraded threshold: 45 seconds.
- Idle suspected threshold (process alive, no progress heartbeats): 2 minutes.

Policy requirements:

- Deterministic evaluation order: liveness -> activity freshness -> checkpoint drift.
- Typed failure class mapping (`transient` vs `terminal`) for queue policy compatibility.
- Idempotent transition writes.
- Deterministic heartbeat evaluation order: runtime lease -> session process -> activity/tool/log progress -> weak context signals.
- Stuck classification must require quorum-aware evidence (not a single stale signal).

## Intervention Feedback Channels

Required:

- Desktop operator feedback (primary intervention UI signal).

Optional but supported:

- Additional provider adapters are phase 2 (webhook/slack/bus), behind the same listener contract.

Notification requirements:

- Include session/run/task correlation IDs.
- Include actionable suggestion (approve input, restart, inspect logs, restore session).
- Emit only on state transitions (not repeated polling duplicates).

## Data and Persistence Expectations

Persist stability snapshots/events with enough fidelity for replay and audit:

- Last known runtime liveness timestamp.
- Last activity event timestamp and classifier result.
- Last checkpoint/progress timestamp.
- Current stability state.
- Transition history with reason codes.
- Heartbeat layer freshness and confidence (`running_confident`, `running_degraded`, `idle_suspected`).

Mandatory heartbeat snapshot fields (minimum):

- `last_runtime_heartbeat_at`
- `last_process_heartbeat_at`
- `last_activity_heartbeat_at`
- `last_tool_heartbeat_at`
- `last_log_heartbeat_at`
- `heartbeat_quorum_state`
- `heartbeat_confidence_score` (normalized 0-100)

Heartbeat ingestion requirement:

- Each heartbeat layer is an ingestion point into session inspection, persisted into lifecycle history, then fanned out upstream through existing ordered event transport.
- Heartbeat-derived transitions must be emitted only when materially changed (idempotent transition policy).

Mandatory persistence targets:

- A generic session table (recommended name: `project_sessions`) must persist live stability evaluation outputs for each worker session lifecycle tick.
- A generic append-only history table (recommended name: `project_session_history`) must persist all lifecycle transitions and intervention events.
- All worker calls must be covered, including asynq tasks and non-asynq worker pipelines (ingestion, agent task worker, and future pipeline stages).
- History coverage must include: enqueue, start, heartbeat, activity classification updates, waiting-input, stuck detection, retry scheduling, dead-letter routing, completion, failure, termination, and manual intervention actions.
- Persisted rows must carry full correlation identifiers (`project_id`, `run_id`, `task_id`, `job_id`, `session_id`) and transition timestamps.
- Persistence must be append-safe and idempotent so repeated checks do not lose or corrupt lifecycle history.

Ordered feedback requirement:

- Live feedback must be persisted in strict order per session using a monotonic sequence (`event_seq`) and event timestamp (`occurred_at`).
- Consumers (desktop/UI/notifiers) must read ordered persisted events, not transient in-memory state, as the source of truth for intervention signals.
- If out-of-order writes occur, ordering must be deterministically recoverable by (`session_id`, `event_seq`, `occurred_at`).

Proposed persistence schema (initial):

- `project_sessions`:
  - one row per active/logical worker session
  - latest state snapshot (`current_state`, `last_liveness_at`, `last_activity_at`, `last_checkpoint_at`, `last_reason_code`)
  - correlation fields (`project_id`, `run_id`, `task_id`, `job_id`, `session_id`)
  - lifecycle bounds (`started_at`, `ended_at`)
- `project_session_history`:
  - append-only ordered event log per session
  - ordering fields (`event_seq`, `occurred_at`)
  - event fields (`event_type`, `previous_state`, `new_state`, `reason_code`, `reason_summary`, `payload_json`)
  - feedback fields (`feedback_channel`, `feedback_emitted_at`, `feedback_status`)
  - correlation fields (`project_id`, `run_id`, `task_id`, `job_id`, `session_id`)
- optional supporting table(s) for delivery attempts:
  - `project_session_feedback_deliveries` for notifier retries and outcome tracking

## Persistence Specification (Concrete)

This section defines the required persistence contract for implementation.

### Table 1: `project_sessions` (current snapshot)

Purpose:

- One row per logical worker session.
- Fast read model for current health/state in board and API.

Required columns:

- `id` (uuid/text primary key)
- `project_id` (text, not null)
- `run_id` (text, nullable for non-run-bound jobs)
- `pipeline_id` (text, nullable)
- `pipeline_type` (text, not null)
- `task_id` (text, nullable)
- `job_id` (text, nullable)
- `session_id` (text, not null, unique)
- `worker_id` (text, nullable)
- `source_runtime` (text, not null)
- `current_state` (text, not null)
- `current_severity` (text, not null)
- `last_reason_code` (text, nullable)
- `last_reason_summary` (text, nullable)
- `last_liveness_at` (timestamptz, nullable)
- `last_activity_at` (timestamptz, nullable)
- `last_checkpoint_at` (timestamptz, nullable)
- `last_event_seq` (bigint, not null, default 0)
- `last_project_event_seq` (bigint, not null, default 0)
- `started_at` (timestamptz, not null)
- `ended_at` (timestamptz, nullable)
- `created_at` (timestamptz, not null)
- `updated_at` (timestamptz, not null)

Required extension columns for layered heartbeats:

- `last_runtime_heartbeat_at` (timestamptz, nullable)
- `last_process_heartbeat_at` (timestamptz, nullable)
- `last_activity_heartbeat_at` (timestamptz, nullable)
- `last_tool_heartbeat_at` (timestamptz, nullable)
- `last_log_heartbeat_at` (timestamptz, nullable)
- `heartbeat_quorum_state` (text, nullable)
- `heartbeat_confidence_score` (int, nullable)

Required constraints:

- unique (`session_id`)
- check `last_event_seq >= 0`
- check `last_project_event_seq >= 0`
- check `ended_at is null or ended_at >= started_at`

Required indexes:

- (`project_id`, `updated_at` desc)
- (`project_id`, `current_state`)
- (`project_id`, `pipeline_type`)
- (`run_id`)
- (`task_id`)
- (`job_id`)

### Table 2: `project_session_history` (append-only lifecycle events)

Purpose:

- Canonical ordered worker lifecycle timeline.
- Source of truth for replay, subscriptions, audit, and fan-out.

Required columns:

- `id` (uuid/text primary key)
- `event_id` (text, not null)
- `schema_version` (int, not null)
- `project_id` (text, not null)
- `run_id` (text, nullable)
- `pipeline_id` (text, nullable)
- `pipeline_type` (text, not null)
- `task_id` (text, nullable)
- `job_id` (text, nullable)
- `session_id` (text, not null)
- `worker_id` (text, nullable)
- `source_runtime` (text, not null)
- `event_seq` (bigint, not null)
- `project_event_seq` (bigint, not null)
- `occurred_at` (timestamptz, not null)
- `ingested_at` (timestamptz, not null)
- `event_type` (text, not null)
- `previous_state` (text, nullable)
- `new_state` (text, nullable)
- `severity` (text, not null)
- `reason_code` (text, nullable)
- `reason_summary` (text, nullable)
- `payload_json` (jsonb, not null, default '{}')
- `feedback_channel` (text, nullable)
- `feedback_status` (text, nullable)
- `feedback_emitted_at` (timestamptz, nullable)
- `actor_type` (text, nullable)  # system|worker|user|automation
- `actor_id` (text, nullable)
- `created_at` (timestamptz, not null)

Required heartbeat metadata in `payload_json` (minimum):

- `heartbeat_layer`
- `heartbeat_source`
- `heartbeat_freshness_ms`
- `heartbeat_quorum_state`
- `heartbeat_confidence_score`

Required constraints:

- unique (`event_id`)
- unique (`session_id`, `event_seq`)
- unique (`project_id`, `project_event_seq`)
- check `event_seq >= 1`
- check `project_event_seq >= 1`
- check `ingested_at >= occurred_at`
- check `schema_version >= 1`

Required indexes:

- (`project_id`, `occurred_at` desc)
- (`project_id`, `project_event_seq`)
- (`project_id`, `session_id`, `event_seq`)
- (`project_id`, `run_id`, `occurred_at` desc)
- (`project_id`, `pipeline_type`, `occurred_at` desc)
- (`project_id`, `event_type`, `occurred_at` desc)
- gin (`payload_json`)

### Table 3: `project_session_feedback_deliveries` (listener delivery tracking)

Purpose:

- Track fan-out delivery attempts/outcomes per listener target.
- Support retry/backoff/audit and listener lag observability.

Required columns:

- `id` (uuid/text primary key)
- `event_id` (text, not null)
- `listener_id` (text, not null)
- `listener_type` (text, not null)  # graphql|internal in phase 1; webhook|slack|bus in phase 2
- `delivery_status` (text, not null)  # pending|delivered|retrying|failed_terminal
- `attempt_count` (int, not null, default 0)
- `last_error` (text, nullable)
- `next_attempt_at` (timestamptz, nullable)
- `last_attempt_at` (timestamptz, nullable)
- `delivered_at` (timestamptz, nullable)
- `cursor_event_seq` (bigint, nullable)
- `created_at` (timestamptz, not null)
- `updated_at` (timestamptz, not null)

Required constraints:

- unique (`event_id`, `listener_id`)
- check `attempt_count >= 0`

Required indexes:

- (`listener_id`, `delivery_status`, `next_attempt_at`)
- (`event_id`)
- (`listener_type`, `updated_at` desc)

### Event Type Enumeration (minimum required)

Minimum event types that must be persisted in `project_session_history`:

- `enqueued`
- `started`
- `heartbeat`
- `runtime_heartbeat`
- `process_heartbeat`
- `activity_heartbeat`
- `tool_heartbeat`
- `log_heartbeat`
- `heartbeat_quorum_degraded`
- `heartbeat_quorum_recovered`
- `activity_classified`
- `waiting_input`
- `stuck_detected`
- `stale_detected`
- `retry_scheduled`
- `retry_started`
- `dead_lettered`
- `manual_intervention`
- `feedback_emitted`
- `feedback_delivery_failed`
- `completed`
- `failed`
- `terminated`
- `gap_detected`
- `gap_reconciled`

### State Enumeration (minimum required)

`current_state` / `new_state` should support at least:

- `healthy_active`
- `healthy_waiting_input`
- `stale_needs_nudge`
- `stuck_needs_intervention`
- `exited_unexpected`
- `terminated`
- `completed`

### Write and Ordering Rules

Required behavior:

- Write `project_session_history` first (append event), then update `project_sessions` snapshot in same transaction boundary when possible.
- If transactional atomicity is unavailable for both writes, use compensating write/reconciliation job and emit `gap_detected` when needed.
- `event_seq` allocation is per `session_id` and monotonic.
- `project_event_seq` allocation is per `project_id` and monotonic.
- Session ordering is (`session_id`, `event_seq`), with `occurred_at` as diagnostic tie-breaker only.
- Project-wide ordering is (`project_id`, `project_event_seq`), with `occurred_at` as diagnostic tie-breaker only.

### Retention and Archival

Initial retention policy:

- `project_sessions`: keep only active/latest row per session (long-lived).
- `project_session_history`: hot retention default 90 days, then archive to cold storage with deterministic replay preserved.
- `project_session_feedback_deliveries`: retain long enough for audit/SLA reporting (recommended >= history retention).
- Cold archive retention must prioritize auditability over minimization (recommended long-term retention).

### Security Requirements for Persistence

- `payload_json` must be pre-redacted according to allow/deny policy before insert.
- Secrets/tokens/credentials must never be written in clear text.
- Redaction policy version should be recorded in payload metadata (`payload_json.redaction_policy_version`).

## Database Shape and Indexing Contract (Design Only)

This section is normative for schema shape and index strategy, but intentionally does not require immediate migration work while the system remains unstable.

### Design status

- These are required table/index targets for the stable implementation phase.
- Current unstable tables can be dropped/rebuilt as needed during iteration.
- Do not optimize for backward compatibility yet; optimize for deterministic ordering and audit correctness.

### Canonical query patterns to design for

1. Project Events Board live stream:

- Filter by `project_id` with deterministic global order by `project_event_seq`.

1. Pipeline drilldown timeline:

- Filter by `project_id`, `run_id` and/or `pipeline_type`, ordered by `project_event_seq`.

1. Session deep inspection:

- Filter by `session_id`, ordered by `event_seq`.

1. Active intervention queue:

- Filter snapshots by `project_id`, `current_state`, recency (`updated_at`).

1. Fan-out retry worker:

- Filter delivery rows by `listener_id`, `delivery_status`, `next_attempt_at`.

### Required index set by table

`project_sessions` required indexes:

- `idx_project_sessions_project_updated` on (`project_id`, `updated_at` desc)
- `idx_project_sessions_project_state` on (`project_id`, `current_state`)
- `idx_project_sessions_project_pipeline` on (`project_id`, `pipeline_type`)
- `idx_project_sessions_run` on (`run_id`) where `run_id is not null`
- `idx_project_sessions_task` on (`task_id`) where `task_id is not null`
- `idx_project_sessions_job` on (`job_id`) where `job_id is not null`
- `idx_project_sessions_project_event_seq` on (`project_id`, `last_project_event_seq` desc)

`project_session_history` required indexes:

- `idx_psh_project_project_event_seq` on (`project_id`, `project_event_seq`)
- `idx_psh_project_session_event_seq` on (`project_id`, `session_id`, `event_seq`)
- `idx_psh_project_occurred` on (`project_id`, `occurred_at` desc)
- `idx_psh_project_run_project_event_seq` on (`project_id`, `run_id`, `project_event_seq`) where `run_id is not null`
- `idx_psh_project_pipeline_project_event_seq` on (`project_id`, `pipeline_type`, `project_event_seq`)
- `idx_psh_project_event_type_project_event_seq` on (`project_id`, `event_type`, `project_event_seq`)
- `idx_psh_payload_gin` gin (`payload_json`)

`project_session_feedback_deliveries` required indexes:

- `idx_psfd_listener_status_next_attempt` on (`listener_id`, `delivery_status`, `next_attempt_at`)
- `idx_psfd_event` on (`event_id`)
- `idx_psfd_listener_type_updated` on (`listener_type`, `updated_at` desc)
- `idx_psfd_retry_scan` on (`delivery_status`, `next_attempt_at`) where `delivery_status in ('pending', 'retrying')`

### Required uniqueness and key constraints

- `project_sessions`:
  - primary key (`id`)
  - unique (`session_id`)

- `project_session_history`:
  - primary key (`id`)
  - unique (`event_id`)
  - unique (`session_id`, `event_seq`)
  - unique (`project_id`, `project_event_seq`)

- `project_session_feedback_deliveries`:
  - primary key (`id`)
  - unique (`event_id`, `listener_id`)

### Deterministic ordering contract by layer

- Session layer order:
  - sort key is strictly (`session_id`, `event_seq`).

- Project layer order:
  - sort key is strictly (`project_id`, `project_event_seq`).

- Time fields (`occurred_at`, `ingested_at`) are diagnostic metadata and must not replace sequence-based ordering.

### Write-path integrity requirements

- Each event write must include both sequence values:
  - session sequence: `event_seq`
  - project sequence: `project_event_seq`
- Snapshot update must advance both `last_event_seq` and `last_project_event_seq`.
- Duplicate writer attempts must be safely rejected by uniqueness constraints and treated as idempotent success by application logic.

### Performance and scaling guardrails

- Design for high-write append behavior on `project_session_history`.
- Prefer narrow composite btree indexes aligned to query patterns; avoid speculative indexes.
- Keep JSON lookups secondary; primary filtering/order must use typed columns.
- Re-evaluate index bloat and scan plans after first production-like load tests.

### Partitioning guidance (deferred, but scoped)

- Partitioning is not required in the unstable phase.
- Stable-phase default: consider range partitioning of `project_session_history` by `occurred_at` (monthly) if table growth or retention jobs require it.
- Any partition plan must preserve uniqueness semantics for (`event_id`), (`session_id`, `event_seq`), and (`project_id`, `project_event_seq`).

Full audit trail requirement:

- Maintain a complete, queryable timeline of the entire worker lifecycle per session.
- Audit trail must support deterministic reconstruction of "what happened", "when", "why", and "what intervention was emitted".
- Stability checks are not ephemeral logs; they are durable operational records.

Retention:

- Keep recent transition history in primary table/read model.
- Longer retention can be folded into existing stream/event persistence strategy.

## DDD and Runtime Boundary Placement

Follow existing architecture mandates:

- Interface layer:
  - Worker handlers trigger stability evaluation hooks at lifecycle boundaries.
- Application layer:
  - Owns stability orchestration, state transitions, and notification dispatch policies.
- Domain layer:
  - Owns stability states, invariants, transition rules, and reason taxonomy.
- Infrastructure layer:
  - Implements liveness probes, session-state readers/parsers, persistence adapters, notifier adapters.

Boundary rules:

- No in-memory coupling between API and worker runtimes.
- Correlation data must cross boundaries through persisted/queued contracts.

## GraphQL and Control-Plane Constraints

Client-facing control plane remains GraphQL.

Potential GraphQL additions (if needed in this scope):

- Query: run/session stability status and recent transition history.
- Subscription/event stream integration for live stability updates.

No non-approved REST control-plane endpoints introduced by this feature.

## Project Events Board (Git-Tree Style)

Add a first-class Project Events Board that behaves like a git-tree explorer for worker execution.

Primary UX goals:

- Show a live project-wide event feed across all worker pipelines.
- Show only currently active runtime activity in Global Live (no historical/backfill rows).
- Allow drilldown from project -> run -> pipeline -> task -> session -> event.
- Preserve strict event ordering and lifecycle context at each node.
- Support both broad monitoring (all events) and deep inspection (single task pipeline).

Conceptual hierarchy:

- Project
- Run
- Pipeline (ingestion, agent-task, generic asynq workflow, future pipeline types)
- Task/Job
- Session
- Event stream entries

Required board modes:

1. Global live mode

- Real-time subscription for active-now worker activity only.
- Must not rely on persistence-read refresh cadence (`project_session_history`, snapshots, or generic table replay) to appear live.
- Must evict/age-out entries that are no longer active so the view represents current runtime state.
- Filterable by run, pipeline type, state, severity, worker, and time window.

1. Pipeline drilldown mode

- Focus view for a single task pipeline.
- Shows ordered event timeline from enqueue to terminal state.
- Includes transition reasons, intervention events, and delivery outcomes.

1. Session deep-inspection mode

- Session-level detail panel for liveness/activity/checkpoint drift signals.
- Shows latest snapshot from `project_sessions` and full history from `project_session_history`.

Git-tree interaction expectations:

- Expand/collapse nodes without losing live position.
- New events append under the correct branch in real time.
- Parent node badges summarize child health (active/waiting/stuck/failed/exited).
- Selecting a node scopes feed and metrics to that subtree.

Event ordering and consistency requirements:

- Board rendering order is driven by persisted order (`event_seq`, `occurred_at`), not arrival order.
- Late/out-of-order deliveries must be re-ordered deterministically.
- UI must indicate gaps or delayed segments when sequence continuity is broken.

## Realtime Contract Requirements for Events Board

### Global Live Active-Now Intent (Mandatory)

Global Live is an operations-now surface, not a history surface.

Required behavior:

- Global Live shows only currently active worker/session/runtime activity.
- Global Live must react to worker runtime signals immediately (sub-second target), without waiting for persistence commits to appear in Postgres-backed history tables.
- Historical/replay/persisted timelines belong only to Pipeline Drilldown and Session Inspection modes.
- If an activity is no longer active, it must leave Global Live (or move to an explicitly stale bucket) within bounded TTL.
- Empty Global Live while Session/Pipeline history has data is valid when nothing is active now.

Non-compliant behavior:

- Global Live updates only when persisted history/snapshot rows are queried/refreshed.
- Global Live surfaces historical-only events as if they are active-now.
- Global Live requires manual reload to reflect runtime activity changes.

Required GraphQL capabilities:

- Subscription: `projectEvents(projectId, filters...)` for global live mode.
- Subscription: `pipelineEvents(projectId, runId, pipelineId)` for focused drilldown.
- Query: paginated tree nodes and event history from persisted tables.
- Query: snapshot + history for a specific session.

Required event payload fields:

- Correlation: `project_id`, `run_id`, `pipeline_id`, `task_id`, `job_id`, `session_id`
- Ordering: `event_seq`, `occurred_at`
- State change: `previous_state`, `new_state`
- Classification: `event_type`, `reason_code`, `reason_summary`, `severity`
- Feedback: `feedback_channel`, `feedback_status`, `feedback_emitted_at`
- Diagnostics: compact `payload_json` for machine-readable details

Pagination and replay requirements:

- Cursor-based pagination for long histories.
- Replay from persisted history must reproduce the same node ordering as live mode.
- Subscriptions must be resumable from cursor/sequence checkpoints.

### Realtime Delivery Semantics (Runtime-Active Push)
Realtime for Global Live means push from runtime-active worker signals, not persistence-query refresh cadence.

Required behavior:

- API must expose a dedicated runtime-active stream for Global Live.
- Runtime-active stream messages must include correlation and liveness/activity fields required to render "active now" deterministically.
- Delivery path must be distributed and resumable across API/worker boundaries, but must not block Global Live rendering on persistence commit availability.
- Persisted tables remain canonical for audit/history modes; they are not the primary source for active-now Global Live rendering.

Mandatory platform requirement:

- Table-change watching must be implemented as a reusable infrastructure capability (agnostic watcher contract), not as a scope-specific one-off implementation.
- The watcher contract must allow additional domains/tables to register watchers without changing core worker-stability business logic.
- Watcher outputs must use deterministic envelopes (table/topic, project/session correlation, high-watermark sequence/cursor, commit timestamp).
- Watcher implementation must support backend swappability (Postgres implementation first, but not hard-coupled in application/domain contracts).
- The same watcher capability is required to be reusable for future non-worker pipelines and control-plane event features.

Explicitly out of scope as the primary realtime mechanism for Global Live:

- Low-latency polling loops as the main API->desktop delivery strategy.
- Postgres history/snapshot query-refresh as the source of active-now liveness UI.

## Event Stream Fan-Out and Listener Hooks

The worker lifecycle stream must be reusable beyond a single UI path.

Core requirement:

- A single canonical ordered event stream is produced once, then fanned out to multiple listeners.
- Desktop (via API/GraphQL) is one listener, not a special-case producer path.
- External listeners must be able to subscribe/receive the same event semantics without custom per-consumer logic.

Supported listener categories:

- Internal control-plane listeners (API stack -> desktop clients).
- Internal worker-side automation listeners (alerting, remediation, escalation workflows).

Phase 2 extension:

- External outbound listeners (webhook targets, message bus consumers, analytics sinks, ops tooling) are added by implementing the same listener contract without changing event semantics.

Fan-out contract requirements:

- Delivery input is the persisted canonical event record (`project_session_history` or equivalent).
- Each listener receives correlation, ordering, and state fields unchanged.
- Listener delivery state must be tracked per listener target (attempts, success/failure, last_error, last_attempt_at).
- Listener failures must not block core persistence of the canonical event stream.
- Retry/backoff policy for listeners must be typed (`transient` vs `terminal`) and auditable.

Hook model requirements:

- Register listeners declaratively by event type/state transition filters.
- Support scoped hooks (project-wide, pipeline-type, task/session specific).
- Support replay hooks from a stored cursor (`event_seq`/timestamp) for recovery and new listener bootstrap.

API/desktop propagation requirements:

- API runtime consumes canonical persisted events and exposes them via GraphQL subscriptions.
- Desktop can receive global and scoped streams with the same ordering guarantees as external listeners.
- No direct in-memory worker->desktop coupling; propagation remains via persisted/distributed contracts.
- API subscription fan-out is triggered by table-change watch signals, then resolved via ordered persisted cursors.
- Watch registration/configuration for table-change signals must be declarative and reusable across features.

## Event Contract Governance and Data Integrity

Canonical event contracts must be versioned and validated.

Requirements:

- Every persisted event includes `schema_version` and `event_id`.
- `event_id` is globally unique and idempotency-safe for reprocessing.
- `event_seq` must be unique per `session_id` and monotonically increasing.
- Enforce DB uniqueness constraints to prevent duplicate writes:
  - unique (`session_id`, `event_seq`)
  - unique (`event_id`)
- Validate required correlation fields at write time; reject partial events.
- Track ingestion/write provenance (`source_runtime`, `source_worker_id`, `ingested_at`).

Gap handling requirements:

- Detect sequence gaps per session and persist explicit `gap_detected` events.
- Provide deterministic reconciliation when late events arrive.
- Surface gap indicators in Project Events Board and subscription payloads.

## Reliability, Backpressure, and Delivery SLOs

Define reliability objectives for event capture and fan-out.

SLO targets (initial):

- Event persistence latency p95 <= 1s from worker emission.
- Desktop/API delivery latency p95 <= 2s from persistence commit.
- Fan-out success rate >= 99.9% with retries (excluding terminal listener failures).

Backpressure requirements:

- If downstream listeners are slow/unavailable, canonical event persistence continues.
- Fan-out workers must use bounded queues and retry policies without unbounded memory growth.
- Persist listener lag metrics and expose them for operator visibility.

## Security and Data Hygiene

Worker events may include sensitive operational payloads and must be protected.

Requirements:

- Redact or hash sensitive fields before persistence/fan-out (tokens, credentials, secrets, personal data).
- Define payload allowlist/denylist policy for `payload_json`.
- Enforce authz boundaries on GraphQL queries/subscriptions by project/session scope.
- Maintain auditability of redaction decisions (reason code or policy version).

## Operational Controls and Intervention Runbooks

Intervention must be actionable and safe, not just informational.

Required controls:

- Manual actions from board: nudge, retry, pause, terminate, restore.
- Every manual action emits a persisted event with actor identity and reason.
- Nudge expansion contract: `manual_nudge` should drive operator-facing follow-up (desktop notification and/or external listener action) in addition to audit history.
- Current implementation note: `manual_nudge` is plumbed end-to-end (board -> GraphQL -> persisted lifecycle event), but no dedicated listener currently consumes `manual_nudge` for downstream action execution.
- Pause semantics requirement: `pause` must never pause an entire Asynq queue because that would halt unrelated workloads.
- Pause execution requirement: `pause` must stop/halt only the correlated task execution (task/job scoped worker work), preserving queue availability for other tasks.
- Terminate semantics requirement: `terminate` is a destructive terminal action and must hard-stop the targeted session pipeline without graceful continuation.
- Terminate execution requirement: `terminate` must cancel/stop the correlated Asynq task and terminate any currently running live agent execution for that same correlation track (`project_id`, `run_id`, `task_id`, `job_id`, `session_id`) without impacting unrelated queue work.
- Restore automation requirement: when a session is classified as restorable from persisted context/checkpoint data, restore must be automatically orchestrated by the system without requiring manual operator intervention.
- Restore UX requirement: operators should not need a manual restore button for normal restorable flows; manual restore should be reserved for explicit break-glass/admin workflows only (if retained).
- Retry UX eligibility requirement: retry must be available only for terminal unhappy stop/failure states (for example `exited_unexpected`, failed terminal outcomes, dead-lettered failures).
- Retry UX guard requirement: retry must be hidden or disabled for successful terminal outcomes and for in-flight sessions; operators must not be able to press retry in those states.
- Runbook links per state (`waiting_input`, `stuck`, `exited_unexpected`) included in UI payloads.
- Circuit-breaker control to disable noisy listeners without stopping canonical persistence.

## Observability and Metrics

Emit platform metrics tied to stability and delivery quality.

Required metrics:

- State transition counts by pipeline/type/severity.
- Time spent in each state per session/task.
- Gap detection rate and reconciliation success rate.
- Listener delivery attempts/success/failure/lag.
- Unexpected exit rate and MTTR (mean time to recovery/intervention).

## Acceptance Criteria

1. Worker detects and classifies the five stability states deterministically.
2. Unexpected exits produce immediate intervention events with correlation IDs.
3. Waiting-input and stuck conditions trigger intervention events at configured thresholds.
4. Desktop receives actionable intervention feedback for transition events.
5. Stability transitions are persisted and queryable with run/task/job correlation.
6. Live worker stability checks are persisted into `project_sessions` and `project_session_history` (or equivalent generic names) with full correlation IDs.
7. Full worker lifecycle audit trail is queryable end-to-end from session creation through terminal state.
8. Recovery flow can use persisted context to support safe restart/restore decisions.
9. No architecture boundary violations (DDD layers and API/worker separation preserved).
10. Project Events Board supports global live feed and task pipeline drilldown with tree navigation.
11. Board event ordering is deterministic from persisted sequence fields and supports replay/resume.
12. Realtime subscriptions provide both all-events and scoped pipeline streams.
13. Realtime API->desktop propagation is table-change-watch driven (DB change signals), not polling-driven.
14. Canonical event stream can be fanned out to multiple listeners (desktop, webhooks, other consumers) without semantic drift.
15. Listener delivery attempts/outcomes are persisted and queryable for audit and replay.
16. Event contract versioning and uniqueness constraints are enforced (`event_id`, `session_id + event_seq`).
17. Sequence gaps are detected, persisted, and visible in board/subscription views.
18. Defined SLOs for persistence/delivery latency and fan-out success are measurable.
19. Sensitive payload redaction is enforced before persistence/fan-out.
20. Manual intervention actions are persisted with actor and reason, and replayable in audit trail.
21. Backpressure in listener pipelines does not block canonical event persistence.
22. Table-change watch infrastructure is reusable/agnostic and can onboard new table streams without architecture changes.
23. Existing worker lease heartbeat coverage is preserved and explicitly mapped as layer-0 heartbeat in session inspection.
24. Session inspection persists and streams layered heartbeat evidence (`runtime`, `process`, `activity`, `tool`, `log`) with deterministic ordering.
25. Running-confidence classification is quorum-based and prevents false "running" status from worker-lease heartbeat alone.

## Test Strategy

Required test coverage:

- Unit tests for transition rules and threshold policy.
- Integration tests for worker lifecycle -> stability transition emission.
- Persistence tests for transition history and idempotency.
- Persistence tests verifying writes to `project_sessions` and `project_session_history` (or equivalent generic lifecycle tables).
- Ordering tests validating monotonic `event_seq` and deterministic replay for live feedback.
- Audit reconstruction tests validating full lifecycle replay from persisted records only.
- Realtime subscription tests for global project feed and pipeline-scoped feed.
- Realtime watch-path tests proving DB change signal -> API cursor read -> GraphQL push flow.
- Negative tests proving polling-only delivery mode is not used as primary mechanism.
- Contract tests for generic watcher registration and deterministic envelope semantics across multiple table streams.
- Adapter parity tests ensuring watcher contract remains stable when swapping concrete backend implementation.
- Tree-view contract tests ensuring node drilldown maps correctly to persisted event branches.
- Fan-out tests validating identical payload semantics across API/desktop and external listeners.
- Listener retry/backoff tests with persisted delivery attempt history.
- Replay tests proving new listeners can bootstrap from stored cursors without data loss.
- Contract tests for schema versioning, idempotency, and uniqueness constraints.
- Gap detection/reconciliation tests with out-of-order and delayed events.
- Backpressure tests verifying canonical persistence remains healthy when listeners fail/lag.
- Redaction tests ensuring secrets are never persisted/fanned out in cleartext.
- Authorization tests for project-scoped event query/subscription access.
- Chaos tests for worker crash/restart, listener outages, and replay recovery.
- Notification adapter tests to prevent duplicate emissions on unchanged state.
- Failure-mode tests:
  - runtime dead + no activity
  - runtime alive + no progress
  - waiting input
  - blocked/error
- Layered heartbeat tests:
  - worker lease heartbeat fresh + no process heartbeat -> must not classify `running_confident`
  - process heartbeat fresh + stale activity/tool/log heartbeats -> classify `running_degraded` then `idle_suspected`
  - process missing + terminal/liveness failure -> classify `done_or_exited` / `exited_unexpected`
  - heartbeat quorum degradation/recovery emits ordered transition events exactly once per transition
  - heartbeat payload persistence includes layer/source/freshness/quorum/confidence metadata

Testing ethics:

- Assertions must encode safe/expected behavior.
- No fake-green tests that accept unstable behavior as acceptable.

## Rollout Plan

1. Implement read-only classifier and persistence first (no user-facing alerts).
2. Enable desktop intervention notifications behind feature toggle.
3. Enable optional webhook/slack adapters.
4. Promote thresholds from conservative defaults after observing production telemetry.

## Risks

- Over-alerting due to noisy/flapping signals.
- False stuck classification for legitimately long-running operations.
- Parser brittleness if session artifact formats evolve.

Mitigations:

- Debounce and transition-based notifications.
- Multi-signal gating (never classify on one weak signal only).
- Typed parser errors and fallback classification strategy.

## Deliverables

- Domain stability model and transition rules.
- Application stability orchestration service.
- Infrastructure readers/probes + persistence + notifier adapters.
- Database persistence updates for `project_sessions` + `project_session_history` + related feedback delivery tables.
- Desktop intervention integration.
- GraphQL query/subscription surfaces for Project Events Board and replay/resume.
- Listener hook/fan-out infrastructure for external sinks and internal subscribers.
- Delivery tracking model for per-listener status, retries, and replay cursor state.
- Event contract/versioning package and validation guards at write boundaries.
- Gap detection and reconciliation components with UI/API surfacing.
- Security/redaction policy implementation for event payloads.
- Operational runbooks and intervention action wiring in board/API.
- Metrics/telemetry dashboards for SLO tracking and incident response.
- Project Events Board UI with global live mode, pipeline drilldown, and session deep-inspection views.
- Full test coverage per strategy above.

## Actionable Vertical Slice Task List

Use this as the authoritative completion tracker for implementation.

Legend:

- Status: `TODO` -> `IN_PROGRESS` -> `DONE`
- Each slice is only `DONE` when all completion checks pass.

### Slice 1: Canonical Event Write Path

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Implement canonical event writer in worker path for all pipelines.
- [x] Enforce required fields (`project_id`, `session_id`, `event_id`, `event_seq`, `project_event_seq`, `event_type`, `occurred_at`).
- [x] Enforce idempotent duplicate handling using uniqueness constraints.
- [x] Persist `project_session_history` records before snapshot mutation.

Completion checks:

- [x] Deterministic write order validated in tests for session and project layers.
- [x] Duplicate event writes are safe (no semantic double-apply).
- [x] Required event contract validation failures are explicit and typed.

### Slice 2: Session Snapshot Materialization

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Upsert `project_sessions` on every accepted lifecycle event.
- [x] Advance `last_event_seq` and `last_project_event_seq` monotonically.
- [x] Persist latest state/classification timestamps (`last_liveness_at`, `last_activity_at`, `last_checkpoint_at`).
- [x] Persist terminal lifecycle bounds (`started_at`, `ended_at`).

Completion checks:

- [x] Snapshot always matches latest history event for the same session.
- [x] Snapshot query latency remains acceptable under expected load.
- [x] Race conditions do not regress ordering guarantees.

### Slice 3: Deterministic Classifier and Threshold Engine

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Implement evaluation order: liveness -> activity freshness -> checkpoint drift.
- [x] Emit deterministic state transitions for defined state taxonomy.
- [x] Implement debounce/flapping protections.
- [x] Map failures to typed classes (`transient` vs `terminal`).

Completion checks:

- [x] The five primary stability modes classify correctly in integration tests.
- [x] Threshold behavior is configurable via typed config.
- [x] No duplicate transition emissions on unchanged state.

### Slice 4: Gap Detection and Reconciliation

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Detect sequence discontinuities per session.
- [x] Persist `gap_detected` and `gap_reconciled` events.
- [x] Implement reconciliation for delayed/out-of-order events.
- [x] Surface gap state in API payloads.

Completion checks:

- [x] Out-of-order replay converges to deterministic final order.
- [x] Gap indicators appear in board/subscription output.
- [x] Reconciliation path is idempotent.

### Slice 5: Internal Fan-Out and Delivery Tracking (Phase 1)

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Implement listener abstraction with phase-1 internal listeners (`graphql`, `internal`).
- [x] Implement DB table-change watch path for canonical history/snapshot updates to wake API fan-out.
- [x] Persist per-listener delivery attempts in `project_session_feedback_deliveries`.
- [x] Implement retry/backoff with typed failure handling.
- [x] Ensure listener failures do not block canonical persistence.
- [x] Implement watcher abstraction as reusable infrastructure contract (feature-agnostic).
- [x] Implement Postgres watcher adapter under infrastructure that satisfies generic watcher contract.
- [x] Fix desktop live-update delivery so Project Events Board updates without requiring manual reload.
- [x] Implement dedicated runtime-active Global Live stream path that is not gated by Postgres history/snapshot commit visibility.
- [x] Ensure Global Live surface includes only active-now items and evicts non-active items via deterministic TTL/state transition rules.

Completion checks:

- [x] Desktop/API stream receives same event semantics as canonical history.
- [x] Desktop/API realtime updates are triggered from table-change watch signals, not polling loops.
- [x] Delivery retries and terminal failures are auditable.
- [x] Backpressure tests prove canonical stream remains healthy.
- [x] At least one non-worker demo stream can be wired through same watcher contract without contract changes.
- [x] Desktop reflects lifecycle/event changes in realtime without pressing reload.
- [x] Global Live updates immediately from runtime-active worker signals even when no new history row has been committed yet.
- [x] Global Live can be empty while Session Inspection/Pipeline Drilldown still show historical records.

Implementation status note (2026-03-05): completed with runtime-active worker->API push signals (`worker_runtime_activity`) backing Global Live, push-driven desktop updates (no polling loop as primary realtime mechanism), and deterministic active-now TTL/state-transition eviction semantics.

### Slice 6: GraphQL Query and Subscription Contracts

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Add global `projectEvents` subscription.
- [x] Add scoped `pipelineEvents` subscription.
- [x] Add query endpoints for session snapshot and replay history.
- [x] Add query endpoints for tree nodes.
- [x] Implement resumable cursor behavior.

Completion checks:

- [x] Subscription replay/resume is deterministic from persisted order keys.
- [x] Query/subscription authz is project-scoped and enforced.
- [x] Contract tests validate required payload fields and schema version behavior.

### Slice 7: Project Events Board Vertical UX

Status: `IN_PROGRESS`
Owner: `worker-runtime`
Target date: `2026-03-05`

Tasks:

- [x] Implement global live mode with filtering.
- [x] Implement pipeline drilldown mode.
- [x] Implement session deep-inspection mode.
- [x] Implement branch health badges and gap indicators.
- [ ] Gate retry action visibility/enabled state so it is only available for unhappy terminal stop/failure states.
- [ ] Hide/disable retry action for successful terminal outcomes and in-flight sessions.
- [ ] Move Project Events Board behind a dedicated launcher button in the Project Dashboard action area (where Edit/New controls live), instead of always-expanded inline rendering.
- [ ] Add a compact Global Live count indicator on the main Project Dashboard surface so event activity is visible without opening the board.
- [ ] Revamp Project Dashboard, Session, and related board views to a cohesive visual language aligned with the provided reference direction (clean ops-console cards, denser telemetry, stronger status contrast, and simplified layout hierarchy).
- [ ] Implement a dedicated git-tree-like events experience for Global Live, Pipeline Drilldown, and Session Inspection that visually matches the Session Matrix reference (row-centric session cards, terminal snippet affordances, status-first telemetry, and intervention affordance patterns).

Completion checks:

- [x] Tree expansion/collapse preserves live context.
- [x] UI ordering follows persisted sequence keys, not transport arrival order.
- [x] Runbook links and intervention actions are visible and actionable.
- [ ] Retry button cannot be invoked for successful terminal sessions or active in-flight sessions.
- [ ] Events Board is discoverable via launcher button and no longer clutters the default Project Dashboard view.
- [ ] Main Project Dashboard shows Global Live event count indicator that updates with event activity.
- [ ] Updated pages share the intended visual treatment and improve scanability of status/action controls.
- [ ] Global Live, Pipeline Drilldown, and Session Inspection share a Session Matrix-like git-tree presentation and interaction model consistent with the approved sample.
- [ ] Global Live displays only active-now worker/runtime activity; historical-only events are excluded from this mode.

Implementation status note (2026-03-05): retry action is currently surfaced too broadly in session inspection; eligibility gating by terminal unhappy state is not yet implemented.
Implementation status note (2026-03-05): Events Board is currently always visible inline and lacks the requested launcher-button navigation/placement model.
Implementation status note (2026-03-05): Main Project Dashboard currently lacks a compact Global Live count indicator in the top action area.
Implementation status note (2026-03-05): Current frontend styling does not yet reflect the requested reference visual language across project/session/event pages.
Implementation status note (2026-03-05): Session Matrix is the approved visual reference for the git-tree events page, but Global Live/Pipeline Drilldown/Session Inspection do not yet match that presentation.
Implementation status note (2026-03-05): Global Live currently can appear stale/misleading during active jobs because updates are still correlated with persisted event availability; this violates the active-now-only intent and must be corrected before marking Slice 7 done.

### Slice 8: Manual Intervention Actions and Auditability

Status: `IN_PROGRESS`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Implement manual actions (`nudge`, `retry`, `pause`, `terminate`, `restore`).
- [x] Persist actor identity, action reason, and resulting transition event.
- [x] Add circuit-breaker controls for noisy listeners.
- [x] Add operator-safe guardrails for destructive actions.
- [ ] Implement task-scoped `pause` execution path that halts only the correlated task/job worker work (not queue-wide pause).
- [ ] Implement task-scoped `terminate` execution path that cancels the correlated Asynq task and hard-stops any active live agent execution for the same session pipeline track.
- [ ] Implement automatic restore orchestration for restorable sessions using persisted checkpoint/session context (no operator button required for normal flow).
- [ ] Remove/deprecate manual restore button from default operator flow (or explicitly gate it as break-glass admin only).

Completion checks:

- [x] Every manual action is fully replayable from history.
- [x] Action authorization is enforced.
- [x] Intervention MTTR and action outcomes are measurable.
- [ ] `pause` performs task/job-scoped execution halt without pausing the queue.
- [ ] `terminate` performs destructive terminal halt for the targeted session pipeline by stopping correlated Asynq work and active live agent execution, without queue-wide interruption.
- [ ] Restorable sessions auto-restore without manual intervention, with deterministic audit trail of the automatic restore decision and execution.

Implementation status note (2026-03-05): `nudge` audit/event plumbing is complete, while listener-side action handling for `manual_nudge` remains a follow-up.
Implementation status note (2026-03-05): `pause` is currently audit/state plumbing only; task/job-scoped worker halt behavior is not yet implemented.
Implementation status note (2026-03-05): `terminate` is currently audit/state plumbing only; correlated Asynq task cancellation and active live-agent hard-stop behavior are not yet implemented.
Implementation status note (2026-03-05): `restore` is currently manual-button/audit-state plumbing (`manual_restore`) and does not automatically execute restore for restorable sessions.

### Slice 9: Security, Redaction, and Data Hygiene

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-04`

Tasks:

- [x] Implement payload allowlist/denylist policy.
- [x] Redact/hash sensitive fields before persistence and fan-out.
- [x] Stamp redaction policy version in payload metadata.
- [x] Add automated secret-leak regression checks for event payloads.

Completion checks:

- [x] No cleartext secrets in persisted/fanned-out payloads.
- [x] Redaction behavior is deterministic and tested.
- [x] Security tests pass for scoped data access and payload hygiene.

### Slice 10: Reliability SLOs, Metrics, and Chaos Validation

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-05`

Tasks:

- [x] Emit required metrics for transitions, gaps, deliveries, lag, and MTTR.
- [x] Build dashboards and SLO alerts.
- [x] Implement chaos scenarios (worker crash/restart, listener outage, delayed events).
- [x] Validate recovery and replay under fault conditions.

Completion checks:

- [x] SLOs are measurable in dashboard and alerting stack.
- [x] Chaos runs produce deterministic recovery behavior.
- [x] Incident runbook references validated against observed failures.

### Slice 11: Phase-2 External Listener Adapters (Deferred)

Status: `DONE`
Owner: `worker-runtime`
Target date: `2026-03-05`

Tasks:

- [x] Implement external adapters (`webhook`, `slack`, `bus`) using the same listener contract.
- [x] Reuse canonical payload semantics with no adapter-specific schema drift.
- [x] Add per-adapter retry policy tuning and delivery observability.

Completion checks:

- [x] External adapters consume canonical events without bypassing persistence.
- [x] Delivery outcomes are queryable with same audit model as phase 1.
- [x] External failure does not impact canonical history durability.

## Progress Rollup

Track cross-slice progress here (update weekly or at sprint close):

- Completed slices: `11 / 11`
- In-progress slices: `0 / 11`
- Blocked slices: `0 / 11`
- Overall confidence: `HIGH`
- Top risks this period: `Dashboard rollout UX polish and production alert tuning`
