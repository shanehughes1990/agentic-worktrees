# Slice 09 — Postgres Persistence Conversion

## Objective

Consolidate durable application state into PostgreSQL (single primary database), while keeping Redis limited to queue transport/cache concerns.

## Scope Constraint

- PostgreSQL is the only durable system of record.
- Redis remains allowed for `asynq` enqueue/consume mechanics and short-lived queue internals.
- Filesystem remains allowed for SCM checkout/repository content and local runtime artifacts, but not as the canonical source of business state.

## Inventory — What SHOULD Persist to Postgres

| Domain Surface                                     | Current State                                                              | Source Files                                                                                                                                                                      | Postgres Requirement                                                                                                                                                           |
| -------------------------------------------------- | -------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Workflow admission ledger (run/task/job lifecycle) | **Postgres-backed (implemented)** for enqueue admission writes (`queued`)  | `internal/application/taskengine/scheduler.go`, `internal/infrastructure/taskengine/postgres/admission_ledger.go`, `internal/interface/graphql/resolvers/scm_resolver_helpers.go` | Persist every admitted job and lifecycle transition (`queued`, `running`, `succeeded`, `failed`, `skipped`, `dead_lettered`) keyed by `run_id/task_id/job_id/idempotency_key`. |
| Retry checkpoints                                  | **Postgres-backed (implemented)**                                          | `internal/infrastructure/taskengine/postgres/checkpoint_store.go`, worker handlers                                                                                                | Keep durable checkpoint persistence in Postgres as source of truth.                                                                                                            |
| Execution journal                                  | **Postgres-backed (implemented)**                                          | `internal/infrastructure/taskengine/postgres/execution_journal.go`, worker handlers                                                                                               | Keep durable execution records in Postgres for replay/audit/debug.                                                                                                             |
| Dead-letter triage history                         | **Postgres-backed (implemented for requeue audit events)**                 | `internal/infrastructure/queue/asynq/dead_letter_manager.go`, `internal/infrastructure/taskengine/postgres/dead_letter_audit.go`                                                  | Persist dead-letter snapshots and operator actions (`requeue`, `discard`, reason, actor, timestamp).                                                                           |
| SCM repository lease coordination                  | **Postgres-backed (implemented)**                                          | `internal/infrastructure/scm/postgres_repo_lease_manager.go`, `internal/application/scm/coordinator.go`                                                                           | Keep lease table as durable coordination layer across workers.                                                                                                                 |
| Tracker canonical board model                      | **Postgres-backed (implemented: snapshots + normalized relational model)** | `internal/infrastructure/tracker/postgres_board_snapshot_provider.go`, `internal/infrastructure/tracker/postgres_normalized_provider.go`, tracker service/bootstrap wiring        | Keep snapshot and normalized canonical persistence paths in Postgres.                                                                                                          |
| Session state snapshots                            | Derived at runtime; not durably stored                                     | `internal/domain/agent/contracts.go`, `internal/application/agent/service.go`                                                                                                     | Persist session snapshots and last-known SCM/task state for query/recovery.                                                                                                    |
| Supervisor decision/event history                  | **Postgres-backed (implemented)**                                          | `internal/infrastructure/supervisor/postgres/event_store.go`, `internal/application/supervisor/service.go`, `internal/bootstrap/api.go`                                           | Keep state transitions + reason codes as append-only decision history.                                                                                                         |
| Stream replay store                                | Planned in roadmap, not implemented                                        | `docs/roadmap/05-realtime-streams.md`                                                                                                                                             | Persist stream events for replay/reconnect diagnostics.                                                                                                                        |
| Control-plane query read models                    | GraphQL mostly dispatch-only today                                         | `internal/interface/graphql/schema/scm.graphqls`, resolvers                                                                                                                       | Persist query-optimized read models in Postgres for GraphQL list/detail/status operations.                                                                                     |

## What Should NOT Move to Postgres

- Queue transport internals (`asynq` queues, retries, worker pull mechanics) remain Redis-backed.
- Git repository working trees and checkout filesystem content remain on disk.
- Telemetry/log pipelines remain in observability backends, not relational business tables.

## Target Data Model (High-Level)

- `workflow_runs` — run-level metadata and status.
- `workflow_tasks` — task-level state per run.
- `workflow_jobs` — job records keyed by queue task ID + correlation IDs. **(enqueue admission writes implemented)**
- `job_checkpoints` — durable step/token checkpoints. **(implemented)**
- `job_execution_events` — append-only execution journal. **(implemented as upserted current-state records)**
- `dead_letter_events` — dead-letter snapshots and operator actions. **(requeue audit implemented)**
- `scm_repo_leases` — lease ownership + expiry. **(implemented)**
- `tracker_board_snapshots` — canonical board snapshots by run/board. **(implemented)**
- `tracker_boards`, `tracker_epics`, `tracker_tasks`, `tracker_task_outcomes` — full normalized tracker model. **(implemented)**
- `agent_session_snapshots` — latest session view + resumable context.
- `supervisor_events` — policy decisions and transition history. **(implemented)**
- `stream_events` (future slice dependency) — replayable event stream records.

## Conversion Roadmap

### Phase 0 — Foundation

- [x] Add shared Postgres client infrastructure and DSN validation.
- [x] Add Postgres lifecycle wiring in worker and API bootstrap.
- [ ] Add formal migration framework and schema versioning.
- [ ] Define transactional boundary helpers and typed repository errors.

### Phase 1 — Execution Reliability First

- [x] Replace Redis checkpoint store with Postgres checkpoint repository.
- [x] Replace Redis execution journal with Postgres execution repository.
- [x] Add workflow run/task/job admission ledger writes at enqueue time.
- [x] Persist dead-letter triage events alongside asynq inspection actions (requeue action path).

### Phase 2 — SCM Coordination Durability

- [x] Replace `InMemoryRepoLeaseManager` with Postgres lease manager.
- [ ] Enforce lease expiry and owner-token constraints in SQL (TTL/expiry semantics).
- [ ] Add conflict and stale-lease recovery policies with deterministic tests.

### Phase 3 — Tracker and Session Persistence

- [x] Persist board snapshots on ingestion sync via Postgres snapshot provider.
- [x] Introduce full Postgres tracker provider for canonical relational model.
- [ ] Persist agent session snapshots + last checkpoint/session state.

### Phase 4 — Control Plane Read Models

- [ ] Add query repositories for GraphQL run/task/job views.
- [ ] Expose historical execution and dead-letter status in control-plane queries.
- [ ] Add pagination/filtering contracts backed by Postgres indexes.

### Phase 5 — Supervisor + Streams Alignment

- [x] Persist supervisor decisions as append-only events.
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
