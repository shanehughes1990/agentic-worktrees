# Slice 09 — Postgres Persistence Conversion

## Objective

Consolidate durable application state into PostgreSQL (single primary database), while keeping Redis limited to queue transport/cache concerns.

## Scope Constraint

- PostgreSQL is the only durable system of record.
- Redis remains allowed for `asynq` enqueue/consume mechanics and short-lived queue internals.
- Filesystem remains allowed for SCM checkout/worktree content and local runtime artifacts, but not as the canonical source of business state.

## Inventory — What SHOULD Persist to Postgres

| Domain Surface                                     | Current State                                                                  | Source Files                                                                                                                                        | Postgres Requirement                                                                                                                                                           |
| -------------------------------------------------- | ------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Workflow admission ledger (run/task/job lifecycle) | Not persisted as first-class model; only queue task IDs and payload dispatch   | `internal/application/taskengine/scheduler.go`, `internal/interface/graphql/resolvers/scm_resolver_helpers.go`                                      | Persist every admitted job and lifecycle transition (`queued`, `running`, `succeeded`, `failed`, `skipped`, `dead_lettered`) keyed by `run_id/task_id/job_id/idempotency_key`. |
| Retry checkpoints                                  | Redis key-value with TTL                                                       | `internal/infrastructure/queue/asynq/checkpoint_store.go`, `internal/interface/worker/agent_handler.go`, `internal/interface/worker/scm_handler.go` | Persist checkpoints durably in Postgres with no implicit TTL loss.                                                                                                             |
| Execution journal                                  | Redis key-value with TTL                                                       | `internal/infrastructure/queue/asynq/execution_journal.go`, worker handlers                                                                         | Persist execution records in Postgres for replay/audit/debug.                                                                                                                  |
| Dead-letter triage history                         | Read directly from asynq archived tasks; no durable triage/requeue audit trail | `internal/infrastructure/queue/asynq/dead_letter_manager.go`                                                                                        | Persist dead-letter snapshots and operator actions (`requeue`, `discard`, reason, actor, timestamp).                                                                           |
| SCM repository lease coordination                  | In-memory mutex map only (single-process scope)                                | `internal/infrastructure/scm/repo_lease_manager.go`, `internal/application/scm/coordinator.go`                                                      | Persist lease table with expiry/owner/token for multi-worker safe coordination.                                                                                                |
| Tracker canonical board model                      | Local JSON source read on demand; not persisted in canonical store             | `internal/infrastructure/tracker/local_json_provider.go`, `internal/domain/tracker/contracts.go`                                                    | Persist normalized board/epic/task/outcome model in Postgres, regardless of source provider.                                                                                   |
| Session state snapshots                            | Derived at runtime; not durably stored                                         | `internal/domain/agent/contracts.go`, `internal/application/agent/service.go`                                                                       | Persist session snapshots and last-known SCM/task state for query/recovery.                                                                                                    |
| Worker capability + heartbeat registry             | Contract exists; no concrete durable advertiser implementation                 | `internal/application/taskengine/worker_capability_contract.go`, `internal/bootstrap/worker.go`                                                     | Persist worker identity, capabilities, heartbeat timestamps, liveness windows.                                                                                                 |
| Supervisor decision/event history                  | Planned in roadmap, not implemented                                            | `docs/roadmap/03-orchestrator-supervisor.md`                                                                                                        | Persist state transitions + reason codes as append-only decision history.                                                                                                      |
| Stream replay store                                | Planned in roadmap, not implemented                                            | `docs/roadmap/05-realtime-streams.md`                                                                                                               | Persist stream events for replay/reconnect diagnostics.                                                                                                                        |
| Control-plane query read models                    | GraphQL mostly dispatch-only today                                             | `internal/interface/graphql/schema/scm.graphqls`, resolvers                                                                                         | Persist query-optimized read models in Postgres for GraphQL list/detail/status operations.                                                                                     |

## What Should NOT Move to Postgres

- Queue transport internals (`asynq` queues, retries, worker pull mechanics) remain Redis-backed.
- Git repository working trees and checkout filesystem content remain on disk.
- Telemetry/log pipelines remain in observability backends, not relational business tables.

## Target Data Model (High-Level)

- `workflow_runs` — run-level metadata and status.
- `workflow_tasks` — task-level state per run.
- `workflow_jobs` — job records keyed by queue task ID + correlation IDs.
- `job_checkpoints` — durable step/token checkpoints.
- `job_execution_events` — append-only execution journal.
- `dead_letter_events` — dead-letter snapshots and operator actions.
- `scm_repo_leases` — lease ownership + expiry.
- `tracker_boards`, `tracker_epics`, `tracker_tasks`, `tracker_task_outcomes` — canonical tracker/taskboard persistence.
- `agent_session_snapshots` — latest session view + resumable context.
- `worker_registry` — worker capabilities + heartbeat/liveness.
- `supervisor_events` (future slice dependency) — policy decisions and transition history.
- `stream_events` (future slice dependency) — replayable event stream records.

## Conversion Roadmap

### Phase 0 — Foundation

- [ ] Add shared Postgres migration framework and schema versioning.
- [ ] Add Postgres health checks in API/worker bootstrap.
- [ ] Define transactional boundary helpers and typed repository errors.

### Phase 1 — Execution Reliability First

- [ ] Replace Redis checkpoint store with Postgres checkpoint repository.
- [ ] Replace Redis execution journal with Postgres execution event repository.
- [ ] Add workflow run/task/job admission ledger writes at enqueue time.
- [ ] Persist dead-letter triage events alongside asynq inspection actions.

### Phase 2 — SCM Coordination Durability

- [ ] Replace `InMemoryRepoLeaseManager` with Postgres lease manager.
- [ ] Enforce lease expiry and owner-token constraints in SQL.
- [ ] Add conflict and stale-lease recovery policies with deterministic tests.

### Phase 3 — Tracker and Session Persistence

- [ ] Introduce Postgres tracker provider for canonical board/task persistence.
- [ ] Persist normalized board snapshots on ingestion sync.
- [ ] Persist agent session snapshots + last checkpoint/session state.

### Phase 4 — Control Plane Read Models

- [ ] Add query repositories for GraphQL run/task/job/worker views.
- [ ] Expose historical execution and dead-letter status in control-plane queries.
- [ ] Add pagination/filtering contracts backed by Postgres indexes.

### Phase 5 — Supervisor + Streams Alignment

- [ ] Persist supervisor decisions as append-only events.
- [ ] Persist stream events with replay cursor semantics.
- [ ] Align subscription replay protocol with persisted stream offsets.

## DDD Placement Rules for This Conversion

- Domain: keep entities/value objects/invariants only.
- Application: define persistence ports/use-case orchestration.
- Infrastructure: implement Postgres repositories/adapters.
- Interface: map request/response only; no direct SQL access.

## Acceptance Criteria

- Durable workflow history survives Redis flush/restart.
- Any run can be reconstructed by `run_id` from Postgres alone.
- Tracker/taskboard canonical state is queryable without local JSON dependency.
- Lease coordination is safe across multiple worker processes.
- Control-plane queries read from Postgres-backed models (not transient runtime memory).

## Dependencies

- Existing Postgres client + observability adapter (`internal/infrastructure/database/postgres`).
- Slices 02, 04, 05, 06, and 03 alignment for event/read-model adoption.

## Exit Check

This conversion is complete when Postgres is the authoritative store for all business-critical and operationally relevant state, and Redis is reduced to queue transport/cache behavior only.
