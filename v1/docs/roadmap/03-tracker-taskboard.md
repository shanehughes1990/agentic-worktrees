# Slice 03 — Tracker + Taskboard

## Objective

Implement tracker-agnostic work-item and board/task ingestion under the single `tracker` slot.

## Scope

- Canonical internal taskboard domain model.
- `tracker` slot contract for issues + board/task lifecycle.
- Local JSON board adapter as first V1 provider.
- External-provider adapter entrypoint (schema and boundary only).

## Acceptance Criteria

- Orchestration flow consumes canonical tracker model only.
- Local JSON board workflows run behavior-compatible with MVP intent (without code sharing).
- External adapters can be added without changing orchestration logic.

## Dependencies

- Orchestrator supervisor.
- GraphQL query/mutation exposure.
