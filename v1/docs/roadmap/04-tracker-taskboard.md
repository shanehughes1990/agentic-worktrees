# Slice 04 — Tracker and Taskboard Model

## Objective

Implement tracker-agnostic planning and execution intake through a canonical internal taskboard model, with Postgres-backed canonical persistence.

## Task Checklist

- [x] Define canonical tracker/taskboard domain model in `internal/domain/tracker`.
- [x] Define tracker provider port contracts in application layer.
- [x] Implement local JSON tracker/taskboard adapter in infrastructure.
- [x] Add board-source selection at project/workflow boundary.
- [x] Define external provider adapter boundary (Jira/Linear-ready contract).
- [x] Add ingestion/sync integration tests against canonical model.
- [x] Persist board snapshots on ingestion sync (`tracker_board_snapshots`).
- [ ] Implement full normalized Postgres tracker model (`tracker_boards`, `tracker_epics`, `tracker_tasks`, `tracker_task_outcomes`).

## Deliverables

- Canonical tracker/taskboard domain model.
- `tracker` provider contract for board/task lifecycle operations.
- Local JSON tracker/taskboard adapter (V1-authored).
- Postgres snapshot persistence for board sync events.
- Planned normalized relational tracker persistence model.

## In Scope

- Domain normalization independent of vendor schemas.
- Adapter contract that keeps orchestration logic unchanged across providers.
- Source selection mechanism per project/workflow.
- Durable Postgres persistence for canonical tracker state.

## Out of Scope

- Full Jira/Linear feature parity in V1 baseline.
- UI-specific board rendering details.

## Acceptance Criteria

- Orchestration consumes canonical model only.
- Local JSON workflows run through canonical contract and persist board snapshots in Postgres.
- Normalized relational tracker persistence path is defined and implementable without orchestration rewrites.

## Dependencies

- Slice 00 complete.
- Slice 01 contracts available for integration.
- Slice 03 supervisor model available for orchestration alignment.

## Exit Check

This slice is complete when task intake and execution planning are provider-agnostic and canonical tracker state is durably persisted in Postgres.
