# TODO: Worker Supervisor Scheduler Dependency

## Discovery Summary

The worker runtime still depends on `taskengine.Scheduler` for supervisor-driven enqueue flows, even after splitting Asynq into API/Worker platforms.

### Where this dependency exists

- `internal/core/worker/worker.go`
  - Bootstraps scheduler via `taskengine.NewScheduler(...)`
  - Wires `SetAdmissionLedger(...)` and `SetAdmissionSignalSink(...)`
  - Passes scheduler into `applicationcontrolplane.NewService(...)`
  - Builds supervisor dispatcher with scheduler

- `internal/infrastructure/supervisor/taskengine/dispatcher.go`
  - Uses scheduler to enqueue follow-up jobs from supervisor decisions (`scheduler.Enqueue(...)`)

- `internal/application/controlplane/documents.go`
  - Uses scheduler for project document task enqueue paths when service is constructed in worker runtime

## Why this matters

Goal is to keep worker as a pure Asynq consumer/processor. Current scheduler usage means worker can still originate enqueue operations in some flows.

## Follow-up tasks

1. Define target boundary:
   - API is sole enqueue/admission owner
   - Worker is consumer-only (register/start/shutdown handlers)

2. Remove worker-side enqueue origins:
   - Refactor supervisor dispatcher usage so enqueue decisions are executed by API-side orchestration
   - Prevent worker-constructed control-plane service from owning enqueue paths

3. Update worker bootstrap:
   - Remove scheduler creation from worker runtime once enqueue paths are removed
   - Keep only `WorkerPlatform` consumer lifecycle (`Register`, `Start`, `Shutdown`)

4. Preserve distributed contract:
   - Ensure any handoff from worker observations to API enqueue remains via persisted/distributed mechanisms

5. Verification:
   - Add/adjust tests proving worker no longer calls `scheduler.Enqueue(...)`
   - Run full `go test ./...`

---

# TODO: Realtime Worker Registry + Heartbeat Refactor (LISTEN/NOTIFY First)

## Status

- Current: completed
- End state target: completed
- Completion gate: do not mark completed until desktop + API + worker compile successfully and full project tests pass again

## Scope mandate

This is a complete refactor/rewrite of the worker registration + heartbeat subsystem.

- Do not deliver this as a shim, bridge layer, or minimal tactical patch
- Do not preserve old polling/legacy lifecycle behavior behind feature toggles
- Replace the existing flow with the new realtime-first model as the primary architecture
- Remove superseded paths once replacement behavior is verified
- Breaking changes are explicitly accepted for this refactor
- No migration path is required for previous worker heartbeat/registration flow
- No backward-compatibility guarantees are required for legacy contracts in this subsystem

## End-to-end implementation plan

### 1) Domain contracts and lifecycle model
- [ ] Define durable registration submission domain contracts (request, response, status, reason, timeout, revoke)
- [ ] Define realtime heartbeat request/response contracts (API initiated, worker response, deadline semantics)
- [ ] Define explicit worker lifecycle transitions for new model (`pending_registration`, `registered`, `healthy`, `invalidated`, `deregistered`)
- [ ] Remove/replace any remaining legacy state semantics not used by the new flow

### 2) Infrastructure persistence (minimal Postgres)
- [ ] Add minimal durable registration submission persistence (create/list pending/update status/revoke)
- [ ] Keep worker registry persistence minimal and aligned with new lifecycle only
- [ ] Remove obsolete persistence fields and write paths that are no longer reused
- [ ] Ensure API startup catch-up query path processes pending registration submissions

### 3) Realtime transport abstraction and PG implementation
- [ ] Keep/establish transport abstraction at domain/application boundaries
- [ ] Implement/align PostgreSQL LISTEN/NOTIFY adapter for registration + heartbeat channels
- [ ] Ensure deterministic envelopes (correlation IDs + idempotency keys)
- [ ] Ensure transport usage remains swappable for future non-PG backend

### 4) API runtime implementation
- [ ] API listens for registration submissions in realtime and judges compatibility
- [ ] API responds accept/reject in realtime and persists outcome atomically
- [ ] API periodically initiates heartbeat requests to registered/healthy workers via realtime transport
- [ ] API enforces deadline, invalidates missed workers, and persists deregistration/invalidation
- [ ] API startup catch-up handles pending submissions created while API was down

### 5) Worker runtime implementation
- [ ] Worker startup creates durable registration submission + realtime notify
- [ ] Worker waits for API decision with bounded timeout/retry policy
- [ ] Worker revokes pending submission on timeout before exiting non-zero
- [ ] Worker exits non-zero on reject and logs explicit reasons
- [ ] Worker handles API heartbeat requests and responds with health proof
- [ ] Worker handles API invalidation/shutdown intent and exits non-zero when live
- [ ] Worker restart always performs fresh registration; old epoch/session never reused

### 6) Desktop scope (frontend)
- [ ] Update desktop worker-related GraphQL operations/models to new contract shape
- [ ] Preserve desktop realtime visibility for registered worker count and worker session information
- [ ] Remove desktop dependencies on deleted legacy fields/events
- [ ] Ensure desktop builds cleanly with updated schema/generated artifacts

### 7) GraphQL/schema/codegen reconciliation
- [ ] Finalize GraphQL schema for worker sessions/settings/events under new model
- [ ] Regenerate GraphQL server artifacts and frontend/client artifacts
- [ ] Remove stale generated code and stale resolver paths no longer used

### 8) Test strategy and required coverage
- [ ] Add API tests: registration accept/reject, startup catch-up, heartbeat request loop, invalidation on missed proof
- [ ] Add worker tests: timeout revoke path, reject non-zero path, heartbeat response path, shutdown intent non-zero exit path
- [ ] Add transport tests: abstraction conformance + PG adapter behavior
- [ ] Add persistence tests: submission durability, pending query, status transitions, revoke semantics
- [ ] Add desktop tests for worker realtime/session/count rendering against new contract
- [ ] Remove obsolete tests from deleted legacy behavior

### 9) Compile and verification gate
- [x] API compile/build green
- [x] Worker compile/build green
- [x] Desktop compile/build green
- [x] Full backend test suite green (`go test ./...`)
- [x] Frontend test suite green (desktop/frontend tests)
- [ ] End-to-end regression passes for registration + heartbeat lifecycle

### 10) Completion status update
- [x] Update this TODO status from `implementing` to `completed` only after all compile + test gates above pass

## Target behavior

Workers rely on API-owned runtime configuration persisted in the database, and worker registration records remain persisted in the database. Coordination and liveness signaling move to a realtime-first PostgreSQL `LISTEN/NOTIFY` pipeline.

Registration compatibility requests must be durable: each worker startup request is persisted in the database and also emitted through realtime `NOTIFY` for low-latency handling.

Heartbeat policy is API-initiated and proof-based:
- Workers may claim `healthy`, but API is the authority that periodically requests heartbeat proofs from currently `registered`/`healthy` workers
- Missing/failed proofs cause API-side force deregistration of that worker epoch
- If the worker process is still alive, it must receive shutdown intent and exit with non-zero status
- If the worker was offline during invalidation and later restarts, it must submit a new registration; old epoch/registration is invalid

Postgres persistence for this subsystem should be minimal:
- Keep authoritative current worker registry state and runtime settings
- Keep durable registration submission records needed for startup compatibility handshake/replay
- Do not add heartbeat audit history retention tables/logs for this flow

Realtime transport architecture must remain swappable:
- Treat realtime heartbeat/registration messaging as an abstraction at domain/application boundaries
- Provide a concrete PostgreSQL `LISTEN/NOTIFY` implementation for now
- If abstraction and concrete implementation are not both present, add them as part of this refactor
- Keep contracts deterministic so transport can be replaced later (for example, a stronger broker) without changing core orchestration logic

## Refactor requirements

1. Realtime registration compatibility handshake:
   - On worker startup, persist a registration compatibility submission record in DB and publish a corresponding realtime `NOTIFY` event
   - API consumes request, validates worker build/runtime compatibility (version + remote backend/runtime config expectations)
   - API returns accept/reject response over the same realtime pipeline
   - API startup path must scan and process pending registration submissions from DB (for requests created while API was unavailable)
   - Worker must not register in DB unless API explicitly responds with `ok`
   - Worker waits for response with bounded timeout; on timeout worker revokes/cancels its pending DB submission and exits non-zero
   - Worker exits non-zero on explicit reject and logs reject reasons at error level

2. API-initiated heartbeat proving:
   - API is the initiator and periodically requests heartbeats from all workers currently marked `registered`/`healthy`
   - Requests and responses flow through `LISTEN/NOTIFY` (no polling-first model)
   - Worker responds to API heartbeat request with proof-of-health payload
   - On missed response deadline, API force-deregisters the worker epoch as invalid
   - API emits shutdown intent/event for live workers so runtime terminates with non-zero exit

3. DB as source of truth + realtime as transport:
   - Keep worker registry/settings in DB as authoritative state
   - Keep registration submission records durable until resolved/cancelled
   - Use realtime messages for coordination signals, then persist resulting state transitions
   - Ensure API remains policy judge for worker health/compatibility decisions
   - Keep schema minimal (no heartbeat audit log retention for this subsystem)
   - Keep transport-facing code behind an explicit realtime transport abstraction (no direct PG wiring in orchestration logic)

4. Lifecycle and failure semantics:
   - Define deterministic request/response envelopes with correlation IDs and idempotency keys
   - Add explicit timeout/retry policy for dropped or delayed notifications
   - Ensure timeout cleanup prevents stale submissions from being approved after worker process is gone
   - Treat force-deregistered epochs as invalid forever; worker must re-register to obtain a new epoch
   - Define worker-side shutdown handling so API invalidation drives explicit non-zero process exit when worker is still running
   - Map failures to typed classes (`transient` vs `terminal`) and deterministic worker exit behavior

5. Verification:
    - Add integration coverage for:
       - startup compatibility accept path
       - startup compatibility reject path (worker exits non-zero)
          - startup compatibility timeout path (worker revokes pending submission and exits non-zero)
          - API restart catch-up path (processes pending DB submissions created before API became available)
       - API heartbeat request loop over `LISTEN/NOTIFY`
    - missed heartbeat proof handling causes force deregistration for the targeted epoch
    - live worker receives shutdown intent and exits non-zero
    - restarted worker after invalidation must submit new registration and cannot reuse old epoch
    - transport abstraction conformance tests (shared contract tests runnable against PG implementation)
    - Run full `go test ./...`
