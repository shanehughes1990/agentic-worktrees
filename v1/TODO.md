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
