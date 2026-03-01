# Version 1 Roadmap (High-Level)

## Purpose

Define the high-level path to V1 release and the release-level outcomes that must be achieved.

Detailed execution plans belong in `docs/roadmap/*`.

## V1 Mission

Ship a GraphQL-first orchestrator platform with local + remote workers, real-time operational streams, tracker-agnostic planning, and a cross-platform client experience, all under strict DDD boundaries, container-first runtime parity, and PostgreSQL as the durable system of record.

## Non-Negotiable Constraints

- V1 is a ground-up rewrite in `v1/`.
- `mvp/` is inspiration only; no copy/migrate/reuse of source.
- Canonical slots: `agent`, `scm`, `tracker`, `notifier`, `client`.
- Dependency direction: `interface -> application -> domain` (infrastructure implements inner-layer ports).
- End-user product UX is client-first (not terminal/tmux/CLI).
- Containerized runtime is the default deployable contract.
- PostgreSQL is the durable source of truth for business/operational state.
- Redis is limited to queue transport internals for `asynq`.

## Release Pillars

1. **Architecture Foundation**
   - Enforced placement rules and slot boundaries.
   - Stable composition roots and runtime contracts.
   - Explicit persistence placement rules by layer.

2. **Execution Plane**
   - Local and remote worker parity.
   - SCM-backed remote bootstrap and resumable checkpoints.
   - Durable checkpoint/journal/lease/worker-registry state in Postgres.

3. **Control Plane**
   - GraphQL schema as primary contract.
   - Queries, mutations, and subscriptions for orchestrator operations.
   - Postgres-backed query/read models for run/task/job/worker history.

4. **Tracker-Agnostic Planning**
   - Canonical taskboard domain model.
   - Local JSON adapter + external provider extension boundary.
   - Postgres canonical tracker persistence (snapshots + normalized relational model).

5. **Realtime Observability and Streams**
   - Session activity, agent output, supervisor/orchestrator decision streams.
   - Correlation IDs through execution and events.
   - Postgres-backed stream replay storage for reconnect diagnostics.

6. **Cross-Platform Client**
   - Runtime-configurable GraphQL endpoint.
   - Operational dashboard + session detail/controls.
   - Live + historical operational visibility from persisted control-plane models.

7. **Container-First Delivery**
   - API and worker container images.
   - Compose-based local parity and release-ready artifact flow.
   - Postgres migration/init lifecycle included in deployment contract.

## High-Level Delivery Phases

### Phase 0 — Foundation

- Finalize scope and placement governance.
- Establish baseline project layout and startup lifecycle contracts.
- Lock persistence baseline (Postgres durable state, Redis queue transport only).

### Phase 1 — Core Runtime Contracts (SCM/Agent/Worker First)

- Define `agent` and `scm` slot interfaces with initial adapters.
- Define worker lease/capability and execution contracts.
- Validate SCM-backed execution baseline before supervisor policy depth.
- Implement durable execution reliability persistence (checkpoint, execution journal, admission ledger, dead-letter triage, lease records, worker registry).

### Phase 2 — Orchestrator + GraphQL Core

- Implement supervisor state/event taxonomy over real execution context.
- Persist supervisor decision history in Postgres.
- Establish GraphQL contract surface and core control actions backed by application/query services.

### Phase 3 — Tracker + Realtime + Client Surface

- Deliver canonical tracker/taskboard ingestion flows with durable persistence.
- Extend tracker persistence from snapshots to normalized relational model.
- Deliver subscription/event stream pathways with persisted replay offsets.
- Deliver first usable cross-platform client flows with replay-aware resilience.

### Phase 4 — Release Hardening

- Container/runtime parity validation.
- Postgres schema migration/versioning hardening.
- Reliability, operability, and release criteria closure.

## Current Persistence Status (Aligned to Slice Plans)

- **Implemented now**: Postgres client/lifecycle wiring, checkpoint persistence, execution journal persistence, SCM lease persistence, worker registry persistence, tracker board snapshot persistence, normalized tracker relational persistence (`tracker_boards`, `tracker_epics`, `tracker_tasks`, `tracker_task_outcomes`), admission ledger enqueue writes, dead-letter requeue audit events, supervisor decision history persistence (`supervisor_events`).
- **Next persistence targets**: lease expiry/stale recovery, session snapshots, stream replay store, GraphQL read models.

## V1 Release Exit Criteria

V1 is release-ready only when all are true:

1. GraphQL control plane supports core orchestrator workflows.
2. Local and remote worker modes both execute via shared orchestration contracts.
3. Realtime session/agent/supervisor streams are available live and replayable from persisted storage.
4. Tracker flows operate on canonical internal model with durable Postgres backing.
5. Cross-platform client can operate the system without terminal-first UX.
6. API and worker are containerized and validated under compose-based runtime profiles with Postgres initialization/migration lifecycle.
7. Postgres is authoritative for business-critical and operationally relevant state; Redis remains queue transport only.
8. DDD boundary and placement constraints are maintained across implementation.

## Risks to Track

- Slot-boundary leakage under delivery pressure.
- Remote worker bootstrap complexity (SCM auth, checkout, resume fidelity).
- Incomplete persistence coverage causing split-brain operational state.
- Stream volume/performance under concurrent sessions.
- Drift between API contract and client expectations.
- Host-native shortcuts eroding container-first parity.

## Immediate Priority Focus

1. Complete remaining persistence conversions identified in `docs/roadmap/09-postgres-persistence-conversion.md`.
2. Land stream persistence (`stream_events`) and replay-path wiring before full client polish.
3. Back GraphQL query/subscription surfaces with Postgres read models.
4. Add formal schema migration/versioning and deployment lifecycle checks.
