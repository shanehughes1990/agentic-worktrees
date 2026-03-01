# Slice 02 — Worker Execution Plane

## Status

- Completed: **Completed**
- Reviewed At: **2026-03-01T01:25:40Z**

## Objective

Deliver one execution plane that supports local and remote workers with shared dispatch, lease, and resume semantics, backed by durable Postgres persistence for execution-critical state.

## Task Checklist

- [x] Introduce queue/task-engine abstraction in application layer.
- [x] Add first concrete queue adapter (`asynq`) in infrastructure layer.
- [x] Wire worker runtime bootstrap to start queue consumer.
- [x] Register worker handlers through the task engine path.
- [x] Define worker capability advertisement contract.
- [x] Implement remote worker adapter contract and SCM-backed remote bootstrap sequence.
- [x] Persist checkpoints in Postgres (`job_checkpoints`).
- [x] Persist execution state in Postgres (`job_execution_events`).
- [x] Replace in-memory SCM lease coordination with Postgres lease manager (`scm_repo_leases`).

## Deliverables

- Dispatch + lease model with deterministic ownership semantics.
- Checkpoint/resume contract with Postgres as durability layer.
- Execution journal contract with Postgres as durability layer.
- Local/remote execution parity path under shared contracts.

## In Scope

- Queue/task engine abstraction and concrete adapter wiring.
- Worker lifecycle states and typed failure handling.
- Correlation IDs across dispatch, execution, and completion.
- Postgres persistence for execution-critical operational state.

## Out of Scope

- Provider-specific optimization beyond baseline reliability.
- UI-specific worker fleet controls beyond control-plane APIs.

## Acceptance Criteria

- Same job contract runs through local and remote execution paths.
- Transient failures resume from persisted Postgres checkpoints.
- Execution status is recoverable from Postgres after worker restart.

## Dependencies

- Slice 00 complete.
- Slice 01 complete.
- Container runtime alignment with slice 08.

## Exit Check

This slice is complete when execution is location-agnostic and reliability-critical state can be reconstructed from Postgres.
