# Slice 04 — Tracker and Taskboard Model

## Objective

Implement tracker-agnostic planning and execution intake through a canonical internal taskboard model.

## Deliverables

- Canonical tracker/taskboard domain model.
- `tracker` slot contract for work-item and board/task lifecycle operations.
- Local JSON tracker/taskboard adapter (V1-authored).
- External-provider adapter boundary (extension point, not full implementation).

## In Scope

- Domain model normalization independent of vendor schemas.
- Adapter contract that keeps orchestration logic unchanged across providers.
- Selection mechanism for board source per project/workflow.

## Out of Scope

- Full Jira/Linear feature parity in V1 baseline.
- UI-specific board rendering details.

## Acceptance Criteria

- Orchestration consumes canonical model only.
- Local JSON workflows run reliably through canonical contract.
- New tracker providers can be added without orchestration rewrites.

## Dependencies

- Slice 00 complete.
- Slice 01 contracts available for integration.
- Slice 03 supervisor model available for orchestration alignment.

## Exit Check

This slice is complete when task intake and execution planning are provider-agnostic at application boundaries.
