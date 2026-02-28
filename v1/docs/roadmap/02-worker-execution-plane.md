# Slice 02 — Worker Execution Plane

## Objective

Deliver one execution plane that supports local and remote workers with shared dispatch, lease, and resume semantics.

## Task Checklist

- [x] Introduce queue/task-engine abstraction in application layer.
- [x] Add first concrete queue adapter (`asynq`) in infrastructure layer.
- [x] Wire worker runtime bootstrap to start queue consumer.
- [x] Register at least one worker handler through the task engine path.
- [x] Define worker capability advertisement contract.
- [x] Implement lease ownership/renewal model.
- [x] Implement checkpoint/resume contract across retries and worker boundaries.
- [x] Define and implement remote worker adapter contract.
- [x] Implement SCM-backed remote bootstrap sequence.
- [x] Add end-to-end local+remote parity integration tests.

## Deliverables

- Worker registration and capability advertisement contracts.
- Dispatch + lease model with deterministic ownership and renewal.
- Checkpoint/resume contract across retries and worker boundaries.
- Local worker adapter and remote worker adapter contract.
- SCM-backed remote bootstrap sequence contract.

## In Scope

- Queue/task engine abstraction and first concrete adapter wiring.
- Worker lifecycle states and failure-class handling.
- Correlation IDs across dispatch, execution, and completion.

## Out of Scope

- Provider-specific optimization beyond baseline reliability.
- UI-specific worker fleet controls beyond required control-plane APIs.

## Acceptance Criteria

- Same job contract runs through local and remote execution paths.
- Remote workers bootstrap from SCM origin without pre-existing checkout assumptions.
- Transient failures can resume from checkpoints.
- Execution outputs provide the signal surface needed by supervisor policies.

## Dependencies

- Slice 00 complete.
- Slice 01 complete.
- Container runtime alignment with slice 08.

## Exit Check

This slice is complete when execution is location-agnostic and provides reliable policy inputs for orchestrator decisions.
